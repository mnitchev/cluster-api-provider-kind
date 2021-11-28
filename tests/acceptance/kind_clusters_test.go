package acceptance_test

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

var _ = Describe("KindClusters", func() {
	var (
		cluster         *clusterv1.Cluster
		kindCluster     *kclusterv1.KindCluster
		ctx             context.Context
		namespace       string
		clusterName     string
		namespaceObj    *corev1.Namespace
		clusterProvider *kindcluster.Provider
	)

	BeforeEach(func() {
		ctx = context.Background()
		clusterName = uuid.New().String()
		clusterProvider = kindcluster.NewProvider()

		namespace = fmt.Sprintf("test-%d", GinkgoParallelNode())
		namespaceObj = &corev1.Namespace{}
		namespaceObj.Name = namespace
		Expect(k8sClient.Create(ctx, namespaceObj)).To(Succeed())

		cluster = &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "carrot",
				Namespace: namespace,
			},
			Spec: clusterv1.ClusterSpec{
				InfrastructureRef: &corev1.ObjectReference{
					APIVersion: kclusterv1.GroupVersion.String(),
					Kind:       "KindCluster",
					Name:       "potato",
					Namespace:  namespace,
				},
			},
		}
		Expect(k8sClient.Create(ctx, cluster)).To(Succeed())

		kindCluster = &kclusterv1.KindCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "potato",
				Namespace: namespace,
			},
			Spec: kclusterv1.KindClusterSpec{
				Name: clusterName,
			},
		}
		Expect(k8sClient.Create(ctx, kindCluster)).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, namespaceObj)).To(Succeed())
		Expect(k8sClient.Delete(ctx, kindCluster)).To(Succeed())
		Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
	})

	Describe("Create", func() {
		It("ceates a local kind cluster", func() {
			Eventually(func() []string {
				clusters, err := clusterProvider.List()
				Expect(err).NotTo(HaveOccurred())
				return clusters
			}, "1m").Should(ContainElement(clusterName))
		})
	})
})
