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

	kclustersv1alpha3 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
)

//counterfeiter:generate . ClusterProvider
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
	Get(context.Context, types.NamespacedName) (*kclustersv1alpha3.KindCluster, error)
	AddFinalizer(context.Context, *kclustersv1alpha3.KindCluster) error
}

// KindClusterReconciler reconciles a KindCluster object
type KindClusterReconciler struct {
	runtimeClient   KindClusterClient
	clusterProvider ClusterProvider
}

func NewKindClusterReconciler(client KindClusterClient, clusterProvider ClusterProvider) *KindClusterReconciler {
	return &KindClusterReconciler{
		runtimeClient:   client,
		clusterProvider: clusterProvider,
	}
}

// SetupWithManager sets up the controller with the Manager.
func (r *KindClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&kclustersv1alpha3.KindCluster{}).
		Complete(r)
}

func (r *KindClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	logger := log.FromContext(ctx)

	kindCluster, err := r.runtimeClient.Get(ctx, req.NamespacedName)
	if err != nil {
		logger.Error(err, "failed to get KindCluster")
		return ctrl.Result{}, err
	}

	if !kindCluster.DeletionTimestamp.IsZero() {
		err = r.clusterProvider.Delete(kindCluster.Spec.Name)
		return ctrl.Result{}, err
	}

	exists, _ := r.clusterProvider.Exists(kindCluster.Spec.Name)
	if exists {
		return ctrl.Result{}, nil
	}

	err = r.runtimeClient.AddFinalizer(ctx, kindCluster)
	if err != nil {
		return ctrl.Result{}, err
	}

	err = r.clusterProvider.Create(kindCluster.Spec.Name)
	if err != nil {
		logger.Error(err, "failed to create cluster")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}
