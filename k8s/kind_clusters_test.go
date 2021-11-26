package k8s_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	"github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"github.com/mnitchev/cluster-api-provider-kind/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("KindClusters", func() {
	var (
		kindClusters *k8s.KindClusters
		kindCluster  *v1alpha3.KindCluster
		ctx          context.Context
	)

	BeforeEach(func() {
		ctx = context.Background()
		kindClusters = k8s.NewKindClusters(k8sClient)

		kindCluster = &v1alpha3.KindCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "potato",
				Namespace: "default",
			},
			Spec: v1alpha3.KindClusterSpec{
				Name: "the-kind-cluster-name",
			},
		}
		Expect(k8sClient.Create(ctx, kindCluster)).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, kindCluster)).To(Succeed())
	})

	Describe("Get", func() {
		It("gets the existing kind cluster", func() {
			actualCluster, err := kindClusters.Get(ctx, types.NamespacedName{Name: "potato", Namespace: "default"})
			Expect(err).NotTo(HaveOccurred())
			Expect(actualCluster).To(Equal(kindCluster))
		})

		When("the cluster does not exist", func() {
			It("returns a not found error", func() {
				actualCluster, err := kindClusters.Get(ctx, types.NamespacedName{Name: "carrot", Namespace: "default"})
				Expect(errors.IsNotFound(err)).To(BeTrue())
				Expect(actualCluster).To(BeNil())
			})
		})
	})

	Describe("Finalizers", func() {
		It("adds and removes the finalizers", func() {
			err := kindClusters.AddFinalizer(ctx, kindCluster)
			Expect(err).NotTo(HaveOccurred())

			namespacedName := types.NamespacedName{
				Name:      "potato",
				Namespace: "default",
			}
			err = k8sClient.Get(ctx, namespacedName, kindCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(kindCluster.Finalizers).To(ContainElement(k8s.ClusterFinalizer))

			err = kindClusters.RemoveFinalizer(ctx, kindCluster)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, namespacedName, kindCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(kindCluster.Finalizers).NotTo(ContainElement(k8s.ClusterFinalizer))
		})
	})
})
