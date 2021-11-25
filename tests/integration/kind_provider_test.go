package integration_test

import (
	"os"

	"github.com/google/uuid"
	"github.com/mnitchev/cluster-api-provider-kind/infrastructure"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/ginkgo/extensions/table"
	. "github.com/onsi/gomega"
	"sigs.k8s.io/kind/pkg/cluster"
)

var _ = Describe("KindProvider", func() {
	var (
		kindProvider    *infrastructure.KindProvider
		clusterProvider *cluster.Provider
		name            string
	)

	BeforeEach(func() {
		name = uuid.New().String()
		clusterProvider = cluster.NewProvider()
		kindProvider = infrastructure.NewKindProvider(kubeconfig, clusterProvider)
	})

	Describe("Create", func() {
		BeforeEach(func() {
			err := kindProvider.Create(name)
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
				err := kindProvider.Create(name)
				Expect(err).To(HaveOccurred())
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

			It("returns the true", func() {
				exists, err := kindProvider.Exists(name)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeTrue())
			})
		})

		When("the cluster does not exist", func() {
			It("returns false", func() {
				exists, err := kindProvider.Exists(name)
				Expect(err).NotTo(HaveOccurred())
				Expect(exists).To(BeFalse())
			})
		})
	})

	When("the docker binary is missing from the PATH", func() {
		DescribeTable("operations return an error",
			func(operation func() error) {
				// save PATH so we can restore it after the test
				pathEnv := os.Getenv("PATH")
				Expect(os.Setenv("PATH", "")).To(Succeed())

				Expect(operation()).NotTo(Succeed())

				// restore the PATH so it doesn't interfere with other tests on
				// the node
				Expect(os.Setenv("PATH", pathEnv)).To(Succeed())
			},

			Entry("create", func() error {
				return kindProvider.Create(name)
			}),
			Entry("exists", func() error {
				_, err := kindProvider.Exists(name)
				return err
			}),
		)
	})
})
