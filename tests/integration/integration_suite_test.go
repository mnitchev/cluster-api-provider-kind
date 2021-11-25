package integration_test

import (
	"io/ioutil"
	"os"
	"testing"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var kubeconfig string

func TestIntegration(t *testing.T) {
	RegisterFailHandler(Fail)
	RunSpecs(t, "Integration Suite")
}

var _ = BeforeSuite(func() {
	tempFile, err := ioutil.TempFile("", "kubeconfig")
	Expect(err).NotTo(HaveOccurred())
	kubeconfig = tempFile.Name()
	os.Setenv("KUBECONFIG", kubeconfig)
})

var _ = AfterSuite(func() {
	os.RemoveAll(kubeconfig)
})
