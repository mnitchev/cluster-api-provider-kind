/*
Copyright 2021.

Licensed under the Apache License, Version 2.0 (the "License");
you may not use this file except in compliance with the License.
You may obtain a copy of the License at

http://www.apache.org/licenses/LICENSE-2.0

Unless required by applicable law or agreed to in writing, software
distributed under the License is distributed on an "AS IS" BASIS,
WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
See the License for the specific language governing permissions and
limitations under the License.
*/

package controllers

import (
	"context"
	"errors"
	"fmt"

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"github.com/mnitchev/cluster-api-provider-kind/k8s"
)

//counterfeiter:generate . ClusterProvider
//counterfeiter:generate . ClusterClient
//counterfeiter:generate . KindClusterClient

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters/finalizers,verbs=update
//+kubebuilder:rbac:groups=cluster.x-k8s.io,resources=clusters,verbs=get;list;watch

type ClusterProvider interface {
	Create(*kclusterv1.KindCluster) error
	Exists(*kclusterv1.KindCluster) (bool, error)
	Delete(*kclusterv1.KindCluster) error
	GetControlPlaneEndpoint(*kclusterv1.KindCluster) (string, int, error)
}

type KindClusterClient interface {
	Get(context.Context, types.NamespacedName) (*kclusterv1.KindCluster, error)
	AddFinalizer(context.Context, *kclusterv1.KindCluster) error
	RemoveFinalizer(context.Context, *kclusterv1.KindCluster) error
	SetControlPlaneEndpoint(context.Context, kclusterv1.APIEndpoint, *kclusterv1.KindCluster) error
	UpdateStatus(context.Context, kclusterv1.KindClusterStatus, *kclusterv1.KindCluster) error
}

type ClusterClient interface {
	Get(context.Context, *kclusterv1.KindCluster) (*clusterv1.Cluster, error)
}

// KindClusterReconciler reconciles a KindCluster object
type KindClusterReconciler struct {
	clusters        ClusterClient
	kindClusters    KindClusterClient
	clusterProvider ClusterProvider
}

func NewKindClusterReconciler(clusters ClusterClient, kindClusters KindClusterClient, clusterProvider ClusterProvider) *KindClusterReconciler {
	return &KindClusterReconciler{
		clusters:        clusters,
		kindClusters:    kindClusters,
		clusterProvider: clusterProvider,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KindClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kclusterv1.KindCluster{}).
		Complete(r)
}

func (r *KindClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	kindCluster, err := r.kindClusters.Get(ctx, req.NamespacedName)
	if k8serrors.IsNotFound(err) {
		logger.Info("KindCluster no longer exists")
		return ctrl.Result{}, nil
	}
	if err != nil {
		logger.Error(err, "failed to get KindCluster")
		return ctrl.Result{}, err
	}

	cluster, err := r.clusters.Get(ctx, kindCluster)
	if err != nil {
		logger.Error(err, "failed to get owner cluster")
		return ctrl.Result{}, err
	}

	if cluster == nil {
		logger.Info("KindCluster not owned by Cluster yet")
		return ctrl.Result{}, nil
	}

	logger = logger.WithValues("cluster-name", kindCluster.Spec.Name)
	ctx = log.IntoContext(ctx, logger)

	if !kindCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDeletion(ctx, kindCluster)
	}

	if kindCluster.Status.Phase == kclusterv1.ClusterPhaseProvisioning {
		logger.Info("cluster still creating - skipping event")
		return ctrl.Result{Requeue: true}, nil
	}
	return r.reconcileNormal(ctx, kindCluster)
}

func (r *KindClusterReconciler) reconcileDeletion(ctx context.Context, kindCluster *kclusterv1.KindCluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("reconciling delete")
	defer logger.Info("done reconciling delete")

	if !controllerutil.ContainsFinalizer(kindCluster, k8s.ClusterFinalizer) {
		logger.Info("cluster does not have finalizer")
		return ctrl.Result{}, nil
	}

	status := &kclusterv1.KindClusterStatus{
		Ready: false,
		Phase: kclusterv1.ClusterPhaseDeleting,
	}
	r.updateStatus(logger, status, kindCluster)

	err := r.clusterProvider.Delete(kindCluster)
	if err != nil {
		logger.Error(err, "failed to delete kind cluster")
		return ctrl.Result{}, err
	}

	err = r.kindClusters.RemoveFinalizer(ctx, kindCluster)
	if err != nil {
		logger.Error(err, "failed to remove finalizer")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

func (r *KindClusterReconciler) reconcileNormal(ctx context.Context, kindCluster *kclusterv1.KindCluster) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	logger.Info("reconciling create")
	defer logger.Info("done reconciling create. Cluster will be created in background")

	// Always update the status.
	// By default do not change the status - this is so we don't change the
	// status in the event of an error. In this case we should requeue the
	// event and try again.
	status := &kclusterv1.KindClusterStatus{
		Ready:          kindCluster.Status.Ready,
		Phase:          kindCluster.Status.Phase,
		FailureMessage: kindCluster.Status.FailureMessage,
	}
	defer r.updateStatus(logger, status, kindCluster)

	if kindCluster.Status.Phase == "" {
		status.Ready = false
		status.Phase = kclusterv1.ClusterPhasePending

		return ctrl.Result{}, nil
	}

	exists, err := r.clusterProvider.Exists(kindCluster)
	if err != nil {
		logger.Error(err, "failed to check if kind cluster exists")
		return ctrl.Result{}, err
	}

	if exists && kindCluster.Status.Phase == kclusterv1.ClusterPhasePending {
		existsErr := errors.New("cluster already exists")
		logger.Error(existsErr, "failed to reconcile")

		status.Ready = false
		status.Phase = kclusterv1.ClusterPhasePending

		// Sometimes kind will fail to create a cluster, but still report it is
		// there. This is because of a docker bug.
		// See https://github.com/kubernetes-sigs/kind/issues/2530 and
		// https://github.com/moby/moby/issues/40835
		if status.FailureMessage == "" {
			status.FailureMessage = existsErr.Error()
		}

		return ctrl.Result{}, existsErr
	}

	if !exists && createdCluster(kindCluster.Status.Phase) {
		logger.Info("cluster does not exist")
		status.Ready = false
		status.Phase = kclusterv1.ClusterPhasePending
		return ctrl.Result{}, nil
	}

	if kindCluster.Status.Phase == kclusterv1.ClusterPhaseProvisioned {
		logger.Info("setting control plane endpoint")
		err = r.setControlPlaneEndpoint(ctx, logger, kindCluster)
		if err != nil {
			logger.Error(err, "failed to set control plane endpoint")
			return ctrl.Result{}, err
		}

		status.Ready = true
		status.Phase = kclusterv1.ClusterPhaseReady

		return ctrl.Result{}, nil
	}

	if kindCluster.Status.Phase == kclusterv1.ClusterPhasePending {
		err := r.kindClusters.AddFinalizer(ctx, kindCluster)
		if err != nil {
			logger.Error(err, "failed to add finalizer")
			return ctrl.Result{}, err
		}

		status.Ready = false
		status.Phase = kclusterv1.ClusterPhaseProvisioning

		go r.createCluster(logger, kindCluster)
		return ctrl.Result{Requeue: true}, nil
	}

	return ctrl.Result{}, nil
}

func (r *KindClusterReconciler) createCluster(logger logr.Logger, kindCluster *kclusterv1.KindCluster) {
	logger.Info("starting cluster creation")
	defer logger.Info("cluster created")

	status := &kclusterv1.KindClusterStatus{
		Ready: false,
		Phase: kclusterv1.ClusterPhaseProvisioned,
	}
	defer r.updateStatus(logger, status, kindCluster)

	err := r.clusterProvider.Create(kindCluster)
	if err != nil {
		status.Phase = kclusterv1.ClusterPhasePending
		status.FailureMessage = fmt.Sprintf("failed to create cluster: %v", err)
		logger.Error(err, "failed to create cluster")
		return
	}
}

func (r *KindClusterReconciler) updateStatus(logger logr.Logger, status *kclusterv1.KindClusterStatus, kindCluster *kclusterv1.KindCluster) {
	err := r.kindClusters.UpdateStatus(context.Background(), *status, kindCluster)
	if err != nil {
		logger.Error(err, "failed to update status")
	}
}

func (r *KindClusterReconciler) setControlPlaneEndpoint(ctx context.Context, logger logr.Logger, kindCluster *kclusterv1.KindCluster) error {
	host, port, err := r.clusterProvider.GetControlPlaneEndpoint(kindCluster)
	if err != nil {
		return err
	}

	endpoint := kclusterv1.APIEndpoint{
		Host: host,
		Port: port,
	}
	return r.kindClusters.SetControlPlaneEndpoint(ctx, endpoint, kindCluster)
}

func createdCluster(phase kclusterv1.ClusterPhase) bool {
	return phase == kclusterv1.ClusterPhaseProvisioned || phase == kclusterv1.ClusterPhaseReady
}
