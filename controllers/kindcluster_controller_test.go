package controllers_test

import (
	"context"
	"errors"
	"time"

	. "github.com/onsi/ginkgo/v2"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"github.com/mnitchev/cluster-api-provider-kind/controllers"
	"github.com/mnitchev/cluster-api-provider-kind/controllers/controllersfakes"
	"github.com/mnitchev/cluster-api-provider-kind/k8s"
)

var _ = Describe("KindclusterController", func() {
	var (
		reconciler        *controllers.KindClusterReconciler
		clusterProvider   *controllersfakes.FakeClusterProvider
		kindClusterClient *controllersfakes.FakeKindClusterClient
		clusterClient     *controllersfakes.FakeClusterClient
		ctx               context.Context
		result            ctrl.Result
		reconcileErr      error
		kindCluster       *kclusterv1.KindCluster
		cluster           *clusterv1.Cluster
	)

	BeforeEach(func() {
		ctx = context.Background()
		clusterProvider = new(controllersfakes.FakeClusterProvider)
		clusterClient = new(controllersfakes.FakeClusterClient)
		kindClusterClient = new(controllersfakes.FakeKindClusterClient)
		reconciler = controllers.NewKindClusterReconciler(clusterClient, kindClusterClient, clusterProvider)

		kindCluster = &kclusterv1.KindCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
			Spec: kclusterv1.KindClusterSpec{
				Name: "the-kind-cluster-name",
			},
		}
		kindClusterClient.GetReturns(kindCluster, nil)

		cluster = &clusterv1.Cluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
		}
		clusterClient.GetReturns(cluster, nil)
		clusterProvider.GetControlPlaneEndpointReturns("127.0.0.1", 1337, nil)
	})

	JustBeforeEach(func() {
		request := ctrl.Request{
			NamespacedName: types.NamespacedName{
				Name:      "foo",
				Namespace: "bar",
			},
		}
		result, reconcileErr = reconciler.Reconcile(ctx, request)
	})

	It("does not return an error", func() {
		Expect(reconcileErr).NotTo(HaveOccurred())
	})

	It("reconciles successfully", func() {
		Expect(result.Requeue).To(BeFalse())
	})

	It("gets the kind cluster using the client", func() {
		Expect(kindClusterClient.GetCallCount()).To(Equal(1))
		actualCtx, namespacedName := kindClusterClient.GetArgsForCall(0)
		Expect(actualCtx).To(Equal(ctx))
		Expect(namespacedName.Name).To(Equal("foo"))
		Expect(namespacedName.Namespace).To(Equal("bar"))
	})

	It("gets the cluster-api Cluster using the client", func() {
		Expect(clusterClient.GetCallCount()).To(Equal(1))
		actualCtx, actualKindCluster := clusterClient.GetArgsForCall(0)
		Expect(actualCtx).To(Equal(ctx))
		Expect(actualKindCluster).To(Equal(kindCluster))
	})

	It("updates the status to not ready and phase pending", func() {
		Expect(kindClusterClient.UpdateStatusCallCount()).To(BeNumerically(">=", 1))
		_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
		Expect(actualStatus.Ready).To(BeFalse())
		Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhasePending))
		Expect(actualCluster).To(Equal(kindCluster))
	})

	When("getting the kind cluster fails", func() {
		BeforeEach(func() {
			kindClusterClient.GetReturns(nil, errors.New("boom"))
		})

		It("requeues the event", func() {
			Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
		})

		When("the kind cluster does not exist", func() {
			BeforeEach(func() {
				kindClusterClient.GetReturns(nil, k8serrors.NewNotFound(schema.GroupResource{}, "foo"))
			})
			It("does not requeue the event", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
			})
		})
	})

	When("the KindCluster is not owned by a Cluster", func() {
		BeforeEach(func() {
			clusterClient.GetReturns(nil, nil)
		})

		It("does not return an error", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
			Expect(result.Requeue).NotTo(BeTrue())
		})

		It("does not create the cluster", func() {
			Expect(clusterProvider.CreateCallCount()).To(Equal(0))
			Expect(kindClusterClient.AddFinalizerCallCount()).To(Equal(0))
		})
	})

	When("getting the owner cluster fails", func() {
		BeforeEach(func() {
			clusterClient.GetReturns(nil, errors.New("boom"))
		})

		It("reqturns an error", func() {
			Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
		})
	})

	Describe("Phase Pending", func() {
		BeforeEach(func() {
			kindCluster.Status.Ready = false
			kindCluster.Status.Phase = kclusterv1.ClusterPhasePending
			kindClusterClient.GetReturns(kindCluster, nil)
		})

		It("does not return an error", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
		})

		It("reques the event", func() {
			Expect(result.Requeue).To(BeTrue())
		})

		It("updates the status to not ready and phase provisioning", func() {
			Expect(kindClusterClient.UpdateStatusCallCount()).To(BeNumerically(">=", 1))
			_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
			Expect(actualStatus.Ready).To(BeFalse())
			Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhaseProvisioning))
			Expect(actualCluster).To(Equal(kindCluster))
		})

		It("registers the finalizer", func() {
			Expect(kindClusterClient.AddFinalizerCallCount()).To(Equal(1))
			_, actualCluster := kindClusterClient.AddFinalizerArgsForCall(0)
			Expect(actualCluster).To(Equal(kindCluster))
		})

		It("creates a cluster using the cluster provider", func() {
			// use eventually as the implementation starts a go routine
			Eventually(clusterProvider.CreateCallCount).Should(Equal(1))
			actualCluster := clusterProvider.CreateArgsForCall(0)
			Expect(actualCluster).To(Equal(kindCluster))
		})

		It("updates the status to provisioned after create finishes", func() {
			Eventually(kindClusterClient.UpdateStatusCallCount).Should(Equal(2))
			_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(1)
			Expect(actualStatus.Ready).To(BeFalse())
			Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhaseProvisioned))
			Expect(actualCluster).To(Equal(kindCluster))
		})

		When("the real kind cluster already exists", func() {
			BeforeEach(func() {
				clusterProvider.ExistsReturns(true, nil)
			})

			It("reqturns an error", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("cluster already exists")))
			})

			It(" the cluster by setting the status to provisioned", func() {
				Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(1))
				_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
				Expect(actualStatus.Ready).To(BeFalse())
				Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhasePending))
				Expect(actualCluster).To(Equal(kindCluster))
			})
		})

		When("adding the finalizer fails", func() {
			BeforeEach(func() {
				kindClusterClient.AddFinalizerReturns(errors.New("boom"))
			})

			It("returns an error", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})

			It("does not try to create the cluster", func() {
				Expect(clusterProvider.CreateCallCount()).To(Equal(0))
			})

			It("does not update the status to provisioning", func() {
				Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(1))
				_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
				Expect(actualStatus.Ready).To(BeFalse())
				Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhasePending))
				Expect(actualCluster).To(Equal(kindCluster))
			})
		})

		When("creating the cluster fails", func() {
			BeforeEach(func() {
				clusterProvider.CreateReturns(errors.New("boom"))
			})

			It("updates the status to not ready and phase pending", func() {
				Eventually(kindClusterClient.UpdateStatusCallCount).Should(Equal(2))
				_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(1)
				Expect(actualStatus.Ready).To(BeFalse())
				Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhasePending))
				Expect(actualCluster).To(Equal(kindCluster))
			})
		})
	})

	Describe("Phase Provisioning", func() {
		BeforeEach(func() {
			kindCluster.Status.Ready = false
			kindCluster.Status.Phase = kclusterv1.ClusterPhaseProvisioning
			kindClusterClient.GetReturns(kindCluster, nil)
		})

		It("does not return an error", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
		})

		It("reques the event", func() {
			Expect(result.Requeue).To(BeTrue())
		})

		It("does not create a cluster", func() {
			Expect(clusterProvider.CreateCallCount()).To(Equal(0))
		})

		It("does not get the control plane endpoint", func() {
			Expect(clusterProvider.GetControlPlaneEndpointCallCount()).To(Equal(0))
		})

		It("does not update the status", func() {
			Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(0))
		})
	})

	Describe("Phase Provisioned", func() {
		BeforeEach(func() {
			kindCluster.Status.Ready = false
			kindCluster.Status.Phase = kclusterv1.ClusterPhaseProvisioned
			kindClusterClient.GetReturns(kindCluster, nil)
			clusterProvider.ExistsReturns(true, nil)
		})

		It("gets the control plane endpoint", func() {
			Expect(clusterProvider.GetControlPlaneEndpointCallCount()).To(Equal(1))
			actualCluster := clusterProvider.GetControlPlaneEndpointArgsForCall(0)
			Expect(actualCluster).To(Equal(kindCluster))
		})

		It("sets the control plane endpoint", func() {
			Expect(kindClusterClient.SetControlPlaneEndpointCallCount()).To(Equal(1))
			_, actualEndpoint, actualCluster := kindClusterClient.SetControlPlaneEndpointArgsForCall(0)
			Expect(actualEndpoint.Host).To(Equal("127.0.0.1"))
			Expect(actualEndpoint.Port).To(Equal(1337))
			Expect(actualCluster).To(Equal(kindCluster))
		})

		It("updates the status to ready and phase ready", func() {
			Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(1))
			_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
			Expect(actualStatus.Ready).To(BeTrue())
			Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhaseReady))
			Expect(actualCluster).To(Equal(kindCluster))
		})

		When("getting the control plane endpoint fails", func() {
			BeforeEach(func() {
				clusterProvider.GetControlPlaneEndpointReturns("", 0, errors.New("boom"))
			})

			It("requeues the event", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})

			It("does not update the status to ready", func() {
				Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(1))
				_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
				Expect(actualStatus.Ready).To(BeFalse())
				Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhaseProvisioned))
				Expect(actualCluster).To(Equal(kindCluster))
			})
		})

		When("setting the control plane endpoint fails", func() {
			BeforeEach(func() {
				kindClusterClient.SetControlPlaneEndpointReturns(errors.New("boom"))
			})

			It("requeues the event", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})

			It("does not update the status to ready", func() {
				Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(1))
				_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
				Expect(actualStatus.Ready).To(BeFalse())
				Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhaseProvisioned))
				Expect(actualCluster).To(Equal(kindCluster))
			})
		})
	})

	Describe("Phase Ready", func() {
		BeforeEach(func() {
			kindCluster.Status.Ready = true
			kindCluster.Status.Phase = kclusterv1.ClusterPhaseReady
			kindClusterClient.GetReturns(kindCluster, nil)
			clusterProvider.ExistsReturns(true, nil)
		})

		It("reconciles successfully", func() {
			Expect(reconcileErr).NotTo(HaveOccurred())
			Expect(result.Requeue).To(BeFalse())
		})

		It("should not try to create the cluster", func() {
			Expect(clusterProvider.CreateCallCount()).To(Equal(0))
		})

		It("does not update the status", func() {
			Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(1))
			_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
			Expect(actualStatus.Ready).To(BeTrue())
			Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhaseReady))
			Expect(actualCluster).To(Equal(kindCluster))
		})

		When("the cluster doesn't exist", func() {
			BeforeEach(func() {
				clusterProvider.ExistsReturns(false, nil)
			})

			It("updates the status to pending", func() {
				Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(1))
				_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
				Expect(actualStatus.Ready).To(BeFalse())
				Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhasePending))
				Expect(actualCluster).To(Equal(kindCluster))
			})
		})

		When("checking if the cluster exists fails", func() {
			BeforeEach(func() {
				clusterProvider.ExistsReturns(false, errors.New("boom"))
			})

			It("requeues the event", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})
		})
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			now := metav1.NewTime(time.Now())
			kindCluster.DeletionTimestamp = &now
			kindCluster.Finalizers = []string{k8s.ClusterFinalizer}
			kindClusterClient.GetReturns(kindCluster, nil)
		})

		It("deletes the cluster", func() {
			Expect(clusterProvider.DeleteCallCount()).To(Equal(1))
			actualCluster := clusterProvider.DeleteArgsForCall(0)
			Expect(actualCluster).To(Equal(kindCluster))
		})

		It("removes the finalizer", func() {
			Expect(kindClusterClient.RemoveFinalizerCallCount()).To(Equal(1))
		})

		It("updates the status to deleted", func() {
			Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(1))
			_, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
			Expect(actualStatus.Ready).To(BeFalse())
			Expect(actualStatus.Phase).To(Equal(kclusterv1.ClusterPhaseDeleting))
			Expect(actualCluster).To(Equal(kindCluster))
		})

		When("updating the status fails", func() {
			BeforeEach(func() {
				kindClusterClient.UpdateStatusReturns(errors.New("boom"))
			})

			It("reconciles successfully", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
				Expect(result.Requeue).To(BeFalse())
				Expect(result.RequeueAfter).To(BeZero())
			})
		})

		When("the KindCluster is not owned by a Cluster", func() {
			BeforeEach(func() {
				clusterClient.GetReturns(nil, nil)
			})

			It("does not return an error", func() {
				Expect(reconcileErr).NotTo(HaveOccurred())
				Expect(result.Requeue).NotTo(BeTrue())
			})

			It("does not create the cluster", func() {
				Expect(clusterProvider.DeleteCallCount()).To(Equal(0))
				Expect(kindClusterClient.RemoveFinalizerCallCount()).To(Equal(0))
			})
		})

		When("removing the finalizer fails", func() {
			BeforeEach(func() {
				kindClusterClient.RemoveFinalizerReturns(errors.New("boom"))
			})

			It("returns an error", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})
		})

		When("the finalizer has already been removed", func() {
			BeforeEach(func() {
				kindCluster.Finalizers = []string{}
				kindClusterClient.GetReturns(kindCluster, nil)
			})

			It("does nothing", func() {
				Expect(kindClusterClient.RemoveFinalizerCallCount()).To(Equal(0))
			})
		})

		When("deleting the cluster fails", func() {
			BeforeEach(func() {
				clusterProvider.DeleteReturns(errors.New("boom"))
			})

			It("returns an error", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})

			It("does not remove the finalizer", func() {
				Expect(kindClusterClient.RemoveFinalizerCallCount()).To(Equal(0))
			})
		})
	})
})
