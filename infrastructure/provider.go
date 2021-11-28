package infrastructure

import (
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"k8s.io/client-go/tools/clientcmd"
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

func (p *KindProvider) GetControlPlaneEndpoint(name string) (host string, port int, err error) {
	kubeconfigFile, err := ioutil.TempFile("", "")
	if err != nil {
		return "", 0, err
	}
	defer kubeconfigFile.Close()
	defer os.RemoveAll(kubeconfigFile.Name())

	err = p.clusterProvider.ExportKubeConfig(name, kubeconfigFile.Name())
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
