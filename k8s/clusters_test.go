package k8s_test

import (
	"context"

	"github.com/google/uuid"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"github.com/mnitchev/cluster-api-provider-kind/k8s"
	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
)

var _ = Describe("Clusters", func() {
	var (
		clusters     *k8s.Clusters
		cluster      *clusterv1.Cluster
		kindCluster  *kclusterv1.KindCluster
		ctx          context.Context
		namespace    string
		namespaceObj *corev1.Namespace
	)

	BeforeEach(func() {
		ctx = context.Background()
		clusters = k8s.NewClusters(k8sClient)

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
				OwnerReferences: []metav1.OwnerReference{
					{
						APIVersion: clusterv1.GroupVersion.String(),
						Kind:       "Cluster",
						Name:       cluster.Name,
						UID:        cluster.UID,
					},
				},
			},
			Spec: kclusterv1.KindClusterSpec{
				Name: "the-kind-cluster-name",
			},
		}
		Expect(k8sClient.Create(ctx, kindCluster)).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, namespaceObj)).To(Succeed())
		Expect(k8sClient.Delete(ctx, cluster)).To(Succeed())
		Expect(k8sClient.Delete(ctx, kindCluster)).To(Succeed())
	})

	Describe("Get", func() {
		It("gets the existing kind cluster", func() {
			var actualCluster *clusterv1.Cluster
			Eventually(func() *clusterv1.Cluster {
				var err error
				Expect(k8sClient.Get(ctx, types.NamespacedName{Name: "potato", Namespace: namespace}, kindCluster)).To(Succeed())
				actualCluster, err = clusters.Get(ctx, kindCluster)
				Expect(err).NotTo(HaveOccurred())
				return actualCluster
			}).ShouldNot(BeNil())

			Expect(actualCluster.Name).To(Equal("carrot"))
			Expect(actualCluster.Namespace).To(Equal(namespace))
		})

		When("the kind cluster is not owned by a cluster", func() {
			var missingCluster *kclusterv1.KindCluster

			BeforeEach(func() {
				missingCluster = &kclusterv1.KindCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "tomato",
						Namespace: namespace,
					},
					Spec: kclusterv1.KindClusterSpec{
						Name: "the-kind-cluster-name",
					},
				}
				Expect(k8sClient.Create(ctx, missingCluster)).To(Succeed())
			})

			AfterEach(func() {
				Expect(k8sClient.Delete(ctx, missingCluster)).To(Succeed())
			})

			It("returns an error", func() {
				actualCluster, err := clusters.Get(ctx, missingCluster)
				Expect(err).NotTo(HaveOccurred())
				Expect(actualCluster).To(BeNil())
			})
		})
	})
})
