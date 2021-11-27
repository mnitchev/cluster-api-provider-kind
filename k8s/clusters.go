package k8s

import (
	"context"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	"sigs.k8s.io/cluster-api/util"
	"sigs.k8s.io/controller-runtime/pkg/client"
)

type Clusters struct {
	runtimeClient client.Client
}

func NewClusters(runtimeClient client.Client) *Clusters {
	return &Clusters{
		runtimeClient: runtimeClient,
	}
}

func (c *Clusters) Get(ctx context.Context, kindCluster *kclusterv1.KindCluster) (*clusterv1.Cluster, error) {
	cluster, err := util.GetOwnerCluster(ctx, c.runtimeClient, kindCluster.ObjectMeta)
	if err != nil {
		return nil, err
	}

	return cluster, nil
}
