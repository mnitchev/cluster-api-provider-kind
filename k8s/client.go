package k8s

import (
	"context"

	"github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"k8s.io/apimachinery/pkg/types"
)

type Client struct{}

func (c *Client) GetKindCluster(ctx context.Context, namespacedName types.NamespacedName) (*v1alpha3.KindCluster, error) {
	return nil, nil
}
