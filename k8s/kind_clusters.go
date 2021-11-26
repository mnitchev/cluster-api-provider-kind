package k8s

import (
	"context"

	"github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
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

func (c *KindClusters) Get(ctx context.Context, namespacedName types.NamespacedName) (*v1alpha3.KindCluster, error) {
	cluster := &v1alpha3.KindCluster{}
	err := c.runtimeClient.Get(ctx, namespacedName, cluster)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}

func (c *KindClusters) AddFinalizer(ctx context.Context, cluster *v1alpha3.KindCluster) error {
	controllerutil.AddFinalizer(cluster, ClusterFinalizer)
	return c.runtimeClient.Update(ctx, cluster)
}

func (c *KindClusters) RemoveFinalizer(ctx context.Context, cluster *v1alpha3.KindCluster) error {
	controllerutil.RemoveFinalizer(cluster, ClusterFinalizer)
	return c.runtimeClient.Update(ctx, cluster)
}
