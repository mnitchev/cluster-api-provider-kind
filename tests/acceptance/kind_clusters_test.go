package acceptance_test

import (
	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	kindcluster "sigs.k8s.io/kind/pkg/cluster"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
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

		namespace = uuid.New().String()
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

	Describe("Create", func() {
		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, namespaceObj)).To(Succeed())
			Expect(k8sClient.Delete(ctx, kindCluster)).To(Succeed())
			Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
		})

		It("sets the control plane endpoint", func() {
			actualCluster := &kclusterv1.KindCluster{}
			Eventually(func() bool {
				namespacedName := types.NamespacedName{
					Name:      "potato",
					Namespace: namespace,
				}
				err := k8sClient.Get(ctx, namespacedName, actualCluster)
				Expect(err).NotTo(HaveOccurred())
				return actualCluster.Status.Ready
			}).Should(BeTrue())
			endpoint := actualCluster.Spec.ControlPlaneEndpoint
			Expect(endpoint.Host).To(Equal("127.0.0.1"))
			Expect(endpoint.Port).To(BeNumerically(">", 1024))
		})

		It("sets the owner cluster's status to Provisioned", func() {
			actualCluster := &clusterv1.Cluster{}
			Eventually(func() string {
				namespacedName := types.NamespacedName{
					Name:      "carrot",
					Namespace: namespace,
				}
				err := k8sClient.Get(ctx, namespacedName, actualCluster)
				Expect(err).NotTo(HaveOccurred())
				return actualCluster.Status.Phase
			}).Should(Equal(string(clusterv1.ClusterPhaseProvisioned)))

			clusters, err := clusterProvider.List()
			Expect(err).NotTo(HaveOccurred())
			Expect(clusters).To(ContainElement(clusterName))
		})
	})

	Describe("Delete clsuter-api Cluster", func() {
		AfterEach(func() {
			Expect(k8sClient.Delete(ctx, namespaceObj)).To(Succeed())
		})

		BeforeEach(func() {
			actualCluster := &kclusterv1.KindCluster{}
			Eventually(func() bool {
				namespacedName := types.NamespacedName{
					Name:      "potato",
					Namespace: namespace,
				}
				err := k8sClient.Get(ctx, namespacedName, actualCluster)
				Expect(err).NotTo(HaveOccurred())
				return actualCluster.Status.Ready
			}).Should(BeTrue())
			Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
		})

		It("deletes the local kind cluster", func() {
			Eventually(func() []string {
				clusters, err := clusterProvider.List()
				Expect(err).NotTo(HaveOccurred())
				return clusters
			}).ShouldNot(ContainElement(clusterName))
		})

		It("deletes the kind cluster", func() {
			actualCluster := &kclusterv1.KindCluster{}
			Eventually(func() error {
				namespacedName := types.NamespacedName{
					Name:      "potato",
					Namespace: namespace,
				}
				return k8sClient.Get(ctx, namespacedName, actualCluster)
			}).Should(HaveOccurred())
		})
	})
})
