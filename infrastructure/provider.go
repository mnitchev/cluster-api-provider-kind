package infrastructure

import (
	"io/ioutil"
	"net"
	"net/url"
	"os"
	"strconv"
	"time"

	"k8s.io/client-go/tools/clientcmd"
	"sigs.k8s.io/kind/pkg/apis/config/v1alpha4"
	"sigs.k8s.io/kind/pkg/cluster"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
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
		cluster.CreateWithV1Alpha4Config(toConfig(kindCluster)),
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

	err = p.clusterProvider.ExportKubeConfig(kindCluster.Spec.Name, kubeconfigFile.Name(), false)
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

func toConfig(kindCluster *kclusterv1.KindCluster) *v1alpha4.Cluster {
	nodes := []v1alpha4.Node{}
	for i := 0; i < kindCluster.Spec.ControlPlaneNodes; i++ {
		nodes = append(nodes, v1alpha4.Node{Role: v1alpha4.ControlPlaneRole})
	}
	for i := 0; i < kindCluster.Spec.WorkerNodes; i++ {
		nodes = append(nodes, v1alpha4.Node{Role: v1alpha4.WorkerRole})
	}
	return &v1alpha4.Cluster{
		Nodes: nodes,
	}
}
