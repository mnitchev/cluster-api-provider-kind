package kind_test

import (
	"os"

	"github.com/google/uuid"
	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"github.com/mnitchev/cluster-api-provider-kind/infrastructure"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"sigs.k8s.io/kind/pkg/cluster"
	"sigs.k8s.io/kind/pkg/cluster/constants"
)

var _ = Describe("KindProvider", func() {
	var (
		kindProvider    *infrastructure.KindProvider
		clusterProvider *cluster.Provider
		name            string
		kindCluster     *kclusterv1.KindCluster
	)

	BeforeEach(func() {
		name = uuid.New().String()
		kindCluster = &kclusterv1.KindCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
			Spec: kclusterv1.KindClusterSpec{
				Name: name,
			},
		}
		clusterProvider = cluster.NewProvider()
		kindProvider = infrastructure.NewKindProvider(kubeconfig, clusterProvider)
	})

	Describe("Create", func() {
		JustBeforeEach(func() {
			err := kindProvider.Create(kindCluster)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(clusterProvider.Delete(name, kubeconfig)).To(Succeed())
		})

		It("creates the kind cluster", func() {
			clusters, err := clusterProvider.List()
			Expect(err).NotTo(HaveOccurred())
			Expect(clusters).To(ContainElement(name))
		})

		When("the cluster already exists", func() {
			It("returns an error", func() {
				err := kindProvider.Create(kindCluster)
				Expect(err).To(HaveOccurred())
			})
		})

		When("the cluster has specified a number of control plane and worker nodes", func() {
			BeforeEach(func() {
				kindCluster.Spec.ControlPlaneNodes = 2
				kindCluster.Spec.WorkerNodes = 2
			})

			It("creates the cluster with the specified number of nodes", func() {
				nodes, err := clusterProvider.ListNodes(name)
				Expect(err).NotTo(HaveOccurred())
				controlPlaneNodes := 0
				workerNodes := 0
				for _, n := range nodes {
					role, err := n.Role()
					Expect(err).NotTo(HaveOccurred())

					if role == constants.ControlPlaneNodeRoleValue {
						controlPlaneNodes++
					}
					if role == constants.WorkerNodeRoleValue {
						workerNodes++
					}
				}

				Expect(controlPlaneNodes).To(Equal(2))
				Expect(workerNodes).To(Equal(2))
			})
		})
	})

	Describe("Exists", func() {
		When("the cluster exists", func() {
			BeforeEach(func() {
				err := clusterProvider.Create(name, cluster.CreateWithKubeconfigPath(kubeconfig))
				Expect(err).NotTo(HaveOccurred())
			})

			AfterEach(func() {
				Expect(clusterProvider.Delete(name, kubeconfig)).To(Succeed())
			})

			It("returns true", func() {
				exists, err := kindProvider.Exists(kindCluster)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})
		})

		When("the cluster does not exist", func() {
			It("returns false", func() {
				exists, err := kindProvider.Exists(kindCluster)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})
		})
	})

	Describe("Delete", func() {
		When("the cluster exists", func() {
			BeforeEach(func() {
				err := clusterProvider.Create(name, cluster.CreateWithKubeconfigPath(kubeconfig))
				Expect(err).NotTo(HaveOccurred())
			})

			It("succeeds", func() {
				err := kindProvider.Delete(kindCluster)
				Expect(err).NotTo(HaveOccurred())

				clusters, err := clusterProvider.List()
				Expect(err).NotTo(HaveOccurred())
				Expect(clusters).NotTo(ContainElement(name))
			})
		})

		When("the cluster does not exist", func() {
			It("returns an error", func() {
				err := kindProvider.Delete(kindCluster)
				Expect(err).NotTo(HaveOccurred())
			})
		})
	})

	Describe("GetControlPlaneEndpoint", func() {
		BeforeEach(func() {
			err := kindProvider.Create(kindCluster)
			Expect(err).NotTo(HaveOccurred())
		})

		AfterEach(func() {
			Expect(clusterProvider.Delete(name, kubeconfig)).To(Succeed())
		})

		It("gets the endpoint", func() {
			host, port, err := kindProvider.GetControlPlaneEndpoint(kindCluster)
			Expect(err).NotTo(HaveOccurred())
			Expect(host).To(Equal("127.0.0.1"))
			Expect(port).To(BeNumerically(">", 1024))
		})
	})

	When("the docker binary is missing from the PATH", func() {
		DescribeTable("operations return an error",
			func(operation func() error) {
				// save PATH so we can restore it after the test
				pathEnv := os.Getenv("PATH")
				// set PATH to empty, making the docker binary inaccessible
				Expect(os.Setenv("PATH", "")).To(Succeed())

				Expect(operation()).NotTo(Succeed())

				// restore the PATH so it doesn't interfere with other tests on
				// the node
				Expect(os.Setenv("PATH", pathEnv)).To(Succeed())
			},

			Entry("create", func() error {
				return kindProvider.Create(kindCluster)
			}),
			Entry("exists", func() error {
				_, err := kindProvider.Exists(kindCluster)
				return err
			}),
			Entry("delete", func() error {
				return kindProvider.Delete(kindCluster)
			}),
			Entry("get control plane endpoint", func() error {
				_, err := kindProvider.Exists(kindCluster)
				return err
			}),
		)
	})
})
