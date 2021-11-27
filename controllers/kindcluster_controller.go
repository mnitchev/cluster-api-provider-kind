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

	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/log"

	"github.com/go-logr/logr"
	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

//counterfeiter:generate . ClusterProvider
//counterfeiter:generate . ClusterClient
//counterfeiter:generate . KindClusterClient

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters/finalizers,verbs=update

type ClusterProvider interface {
	Create(string) error
	Exists(string) (bool, error)
	Delete(string) error
}

type KindClusterClient interface {
	Get(context.Context, types.NamespacedName) (*kclusterv1.KindCluster, error)
	AddFinalizer(context.Context, *kclusterv1.KindCluster) error
	RemoveFinalizer(context.Context, *kclusterv1.KindCluster) error
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
	logger = logger.WithValues("name", req.Name, "namespace", req.Namespace)

	kindCluster, err := r.kindClusters.Get(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "failed to get KindCluster")
		return ctrl.Result{}, err
	}

	cluster, _ := r.clusters.Get(ctx, kindCluster)
	if cluster == nil {
		logger.Info("KindCluster not owned by Cluster yet")
		return ctrl.Result{}, nil
	}

	if !kindCluster.DeletionTimestamp.IsZero() {
		return r.reconcileDeletion(ctx, logger, kindCluster)
	}

	return r.reconcileNormal(ctx, logger, kindCluster)
}

func (r *KindClusterReconciler) reconcileDeletion(ctx context.Context, logger logr.Logger, kindCluster *kclusterv1.KindCluster) (ctrl.Result, error) {
	err := r.clusterProvider.Delete(kindCluster.Spec.Name)
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

func (r *KindClusterReconciler) reconcileNormal(ctx context.Context, logger logr.Logger, kindCluster *kclusterv1.KindCluster) (ctrl.Result, error) {
	exists, err := r.clusterProvider.Exists(kindCluster.Spec.Name)
	if err != nil {
		logger.Error(err, "failed to check if kind cluster exists")
		return ctrl.Result{}, err
	}

	if exists {
		return ctrl.Result{}, nil
	}

	err = r.kindClusters.AddFinalizer(ctx, kindCluster)
	if err != nil {
		logger.Error(err, "failed to add finalizer")
		return ctrl.Result{}, err
	}

	err = r.clusterProvider.Create(kindCluster.Spec.Name)
	if err != nil {
		logger.Error(err, "failed to create cluster")
		return ctrl.Result{}, err
	}

	return r.updateStatus(ctx, logger, kindCluster)
}

func (r *KindClusterReconciler) updateStatus(ctx context.Context, logger logr.Logger, kindCluster *kclusterv1.KindCluster) (ctrl.Result, error) {
	status := kclusterv1.KindClusterStatus{
		Ready: true,
	}

	err := r.kindClusters.UpdateStatus(ctx, status, kindCluster)
	if err != nil {
		logger.Error(err, "failed to update status")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
