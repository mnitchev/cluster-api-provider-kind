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

	ctrl "sigs.k8s.io/controller-runtime"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/log"
	"sigs.k8s.io/kind/pkg/cluster"

	infrastructurev1alpha3 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
)

type ClusterProvider interface {
	Create(name string, options ...cluster.CreateOption) error
}

// KindClusterReconciler reconciles a KindCluster object
type KindClusterReconciler struct {
	client          client.Client
	clusterProvider ClusterProvider
}

func NewKindClusterReconciler(client client.Client, clusterProvider ClusterProvider) *KindClusterReconciler {
	return &KindClusterReconciler{
		client:          client,
		clusterProvider: clusterProvider,
	}
}

//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters,verbs=get;list;watch;create;update;patch;delete
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters/status,verbs=get;update;patch
//+kubebuilder:rbac:groups=infrastructure.cluster.x-k8s.io,resources=kindclusters/finalizers,verbs=update

// Reconcile is part of the main kubernetes reconciliation loop which aims to
// move the current state of the cluster closer to the desired state.
// TODO(user): Modify the Reconcile function to compare the state specified by
// the KindCluster object against the actual cluster state, and then
// perform operations to make the cluster state reflect the state specified by
// the user.
//
// For more details, check Reconcile and its Result here:
// - https://pkg.go.dev/sigs.k8s.io/controller-runtime@v0.10.0/pkg/reconcile
func (r *KindClusterReconciler) Reconcile(ctx context.Context, req ctrl.Request) (ctrl.Result, error) {
	log := log.FromContext(ctx)

	kindCluster := &infrastructurev1alpha3.KindCluster{}
	err := r.client.Get(ctx, req.NamespacedName, kindCluster)
	if err != nil {
		log.Error(err, "failed to get kind cluster")
		return ctrl.Result{}, err
	}

	err = r.clusterProvider.Create(kindCluster.Spec.Name, cluster.CreateWithKubeconfigPath("/junk/kubeconfig"))
	if err != nil {
		log.Error(err, "failed to create kind cluster")
		return ctrl.Result{}, err
	}

	return ctrl.Result{}, nil
}

// SetupWithManager sets up the controller with the Manager.
func (r *KindClusterReconciler) SetupWithManager(mgr ctrl.Manager) error {
	return ctrl.NewControllerManagedBy(mgr).
		For(&infrastructurev1alpha3.KindCluster{}).
		Complete(r)
}
