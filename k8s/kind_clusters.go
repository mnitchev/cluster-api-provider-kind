package k8s

import (
	"context"

	"github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"k8s.io/apimachinery/pkg/types"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

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
