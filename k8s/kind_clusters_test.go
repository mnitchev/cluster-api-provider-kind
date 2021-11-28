package k8s_test

import (
	"context"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"github.com/mnitchev/cluster-api-provider-kind/k8s"
	"k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
)

var _ = Describe("KindClusters", func() {
	var (
		kindClusters   *k8s.KindClusters
		kindCluster    *kclusterv1.KindCluster
		ctx            context.Context
		namespacedName types.NamespacedName
	)

	BeforeEach(func() {
		ctx = context.Background()
		kindClusters = k8s.NewKindClusters(k8sClient)

		namespacedName = types.NamespacedName{
			Name:      "potato",
			Namespace: "default",
		}
		kindCluster = &kclusterv1.KindCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      namespacedName.Name,
				Namespace: namespacedName.Namespace,
			},
			Spec: kclusterv1.KindClusterSpec{
				Name: "the-kind-cluster-name",
			},
		}
	})

	JustBeforeEach(func() {
		Expect(k8sClient.Create(ctx, kindCluster)).To(Succeed())
	})

	AfterEach(func() {
		Expect(k8sClient.Delete(ctx, kindCluster)).To(Succeed())
	})

	Describe("Get", func() {
		It("gets the existing kind cluster", func() {
			actualCluster, err := kindClusters.Get(ctx, namespacedName)
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

			err = k8sClient.Get(ctx, namespacedName, kindCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(kindCluster.Finalizers).To(ContainElement(k8s.ClusterFinalizer))

			err = kindClusters.RemoveFinalizer(ctx, kindCluster)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, namespacedName, kindCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(kindCluster.Finalizers).NotTo(ContainElement(k8s.ClusterFinalizer))
		})

		When("the finalizer is not set", func() {
			When("removing the finalizer", func() {
				It("does not fail", func() {
					err := kindClusters.RemoveFinalizer(ctx, kindCluster)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})

		When("the finalizer is already set", func() {
			When("adding the finalizer", func() {
				It("does not fail", func() {
					err := kindClusters.AddFinalizer(ctx, kindCluster)
					Expect(err).NotTo(HaveOccurred())

					err = kindClusters.AddFinalizer(ctx, kindCluster)
					Expect(err).NotTo(HaveOccurred())

					// remove finalizer so the resource can be deleted
					err = kindClusters.RemoveFinalizer(ctx, kindCluster)
					Expect(err).NotTo(HaveOccurred())
				})
			})
		})
	})

	Describe("AddControlPlaneEndpoint", func() {
		var endpoint kclusterv1.APIEndpoint
		BeforeEach(func() {
			endpoint = kclusterv1.APIEndpoint{
				Host: "127.0.0.1",
				Port: 1337,
			}
		})

		JustBeforeEach(func() {
			err := kindClusters.SetControlPlaneEndpoint(ctx, endpoint, kindCluster)
			Expect(err).NotTo(HaveOccurred())
		})

		It("adds the control plane endpoint", func() {
			err := k8sClient.Get(ctx, namespacedName, kindCluster)
			Expect(err).NotTo(HaveOccurred())
			actualEndpoint := kindCluster.Spec.ControlPlaneEndpoint
			Expect(actualEndpoint.Host).To(Equal("127.0.0.1"))
			Expect(actualEndpoint.Port).To(Equal(1337))
		})

		When("the endpoint is already set", func() {
			BeforeEach(func() {
				existingEndpoint := kclusterv1.APIEndpoint{
					Host: "127.0.0.1",
					Port: 1337,
				}
				kindCluster.Spec.ControlPlaneEndpoint = existingEndpoint

				endpoint.Host = "172.0.0.1"
				endpoint.Port = 8080
			})

			It("overrides it", func() {
				err := k8sClient.Get(ctx, namespacedName, kindCluster)
				Expect(err).NotTo(HaveOccurred())
				actualEndpoint := kindCluster.Spec.ControlPlaneEndpoint
				Expect(actualEndpoint.Host).To(Equal("172.0.0.1"))
				Expect(actualEndpoint.Port).To(Equal(8080))
			})
		})
	})

	Describe("UpdateStatus", func() {
		It("updates the status", func() {
			status := kclusterv1.KindClusterStatus{
				Ready: true,
			}
			err := kindClusters.UpdateStatus(ctx, status, kindCluster)
			Expect(err).NotTo(HaveOccurred())

			err = k8sClient.Get(ctx, namespacedName, kindCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(kindCluster.Status.Ready).To(BeTrue())
		})

		When("the cluster does not exist", func() {
			It("returns an error", func() {
				status := kclusterv1.KindClusterStatus{
					Ready: true,
				}
				missingCluster := &kclusterv1.KindCluster{
					ObjectMeta: metav1.ObjectMeta{
						Name:      "carrot",
						Namespace: namespacedName.Namespace,
					},
					Spec: kclusterv1.KindClusterSpec{
						Name: "the-kind-cluster-name",
					},
				}
				err := kindClusters.UpdateStatus(ctx, status, missingCluster)
				Expect(err).To(HaveOccurred())
				Expect(errors.IsNotFound(err)).To(BeTrue())
			})
		})
	})
})
