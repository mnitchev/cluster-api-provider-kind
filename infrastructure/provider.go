package infrastructure

import (
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/cluster"
)

const defaultWaitTime = 10 * time.Minute

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

func (p *KindProvider) Create(kindCluster *kclusterv1.KindCluster) error {
	return p.clusterProvider.Create(
		kindCluster.Spec.Name,
		cluster.CreateWithKubeconfigPath(p.kubeconfigPath),
		cluster.CreateWithWaitForReady(defaultWaitTime))
}

func (p *KindProvider) Exists(kindCluster *kclusterv1.KindCluster) (bool, error) {
	clusters, err := p.clusterProvider.List()
	if err != nil {
		return false, err
	}

	for _, cluster := range clusters {
		if cluster == kindCluster.Spec.Name {
			return true, nil
		}
	}

	return false, nil
}

func (p *KindProvider) Delete(kindCluster *kclusterv1.KindCluster) error {
	return p.clusterProvider.Delete(kindCluster.Spec.Name, p.kubeconfigPath)
}

func (p *KindProvider) GetControlPlaneEndpoint(kindCluster *kclusterv1.KindCluster) (host string, port int, err error) {
	kubeconfigFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", 0, err
	}
	defer kubeconfigFile.Close()
	defer os.RemoveAll(kubeconfigFile.Name())

	err = p.clusterProvider.ExportKubeConfig(kindCluster.Spec.Name, kubeconfigFile.Name())
	if err != nil {
		return "", 0, err
	}

	kubeConfig, err := clientcmd.BuildConfigFromFlags("", kubeconfigFile.Name())
	if err != nil {
		return "", 0, err
	}

	apiURL, err := url.Parse(kubeConfig.Host)
	if err != nil {
		return "", 0, err
	}

	host, portStr, err := net.SplitHostPort(apiURL.Host)
	if err != nil {
		return "", 0, err
	}

	port, err = strconv.Atoi(portStr)
	if err != nil {
		return "", 0, err
	}

	return host, port, nil
}
