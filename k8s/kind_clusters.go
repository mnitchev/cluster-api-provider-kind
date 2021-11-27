package k8s

import (
	"context"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
	"sigs.k8s.io/controller-runtime/pkg/controller/controllerutil"
)

const ClusterFinalizer = "kindcluster.infrastructure.cluster.x-k8s.io"

type KindClusters struct {
	runtimeClient client.Client
}

func NewKindClusters(runtimeClient client.Client) *KindClusters {
	return &KindClusters{
		runtimeClient: runtimeClient,
	}
}

func (c *KindClusters) Get(ctx context.Context, namespacedName types.NamespacedName) (*kclusterv1.KindCluster, error) {
	cluster := &kclusterv1.KindCluster{}
	err := c.runtimeClient.Get(ctx, namespacedName, cluster)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func (c *KindClusters) AddFinalizer(ctx context.Context, cluster *kclusterv1.KindCluster) error {
	originalCluster := cluster.DeepCopy()
	controllerutil.AddFinalizer(cluster, ClusterFinalizer)
	return c.runtimeClient.Patch(ctx, cluster, client.MergeFrom(originalCluster))
}

func (c *KindClusters) RemoveFinalizer(ctx context.Context, cluster *kclusterv1.KindCluster) error {
	originalCluster := cluster.DeepCopy()
	controllerutil.RemoveFinalizer(cluster, ClusterFinalizer)
	return c.runtimeClient.Patch(ctx, cluster, client.MergeFrom(originalCluster))
}

func (c *KindClusters) UpdateStatus(ctx context.Context, status kclusterv1.KindClusterStatus, cluster *kclusterv1.KindCluster) error {
	originalCluster := cluster.DeepCopy()
	cluster.Status = status
	return c.runtimeClient.Status().Patch(ctx, cluster, client.MergeFrom(originalCluster))
}
