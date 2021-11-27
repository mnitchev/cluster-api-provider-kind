package infrastructure

import (
	"time"

	"sigs.k8s.io/kind/pkg/cluster"
)

const defaultWaitTime = 5 * time.Minute

type KindProvider struct {
	kubeconfigPath  string
	clusterProvider *cluster.Provider
}

func NewKindProvider(kubeconfigPath string, clusterProvider *cluster.Provider) *KindProvider {
	return &KindProvider{
		kubeconfigPath:  kubeconfigPath,
		clusterProvider: clusterProvider,
	}
}

func (p *KindProvider) Create(name string) error {
	return p.clusterProvider.Create(
		name,
		cluster.CreateWithKubeconfigPath(p.kubeconfigPath),
		cluster.CreateWithWaitForReady(defaultWaitTime))
}

func (p *KindProvider) Exists(name string) (bool, error) {
	clusters, err := p.clusterProvider.List()
	if err != nil {
		return false, err
	}

	for _, cluster := range clusters {
		if cluster == name {
			return true, nil
		}
	}

	return false, nil
}

func (p *KindProvider) Delete(name string) error {
	return p.clusterProvider.Delete(name, p.kubeconfigPath)
}
