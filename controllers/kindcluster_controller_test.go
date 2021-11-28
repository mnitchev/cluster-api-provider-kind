package controllers_test

import (
	"context"
	"errors"
	"time"

	kclusterv1 "github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"github.com/mnitchev/cluster-api-provider-kind/controllers"
	"github.com/mnitchev/cluster-api-provider-kind/controllers/controllersfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	k8serrors "k8s.io/apimachinery/pkg/api/errors"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/runtime/schema"
	"k8s.io/apimachinery/pkg/types"
	clusterv1 "sigs.k8s.io/cluster-api/api/v1beta1"
	ctrl "sigs.k8s.io/controller-runtime"
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

	Describe("Create", func() {
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

		It("updates the status to not ready before creating the cluster", func() {
			Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(2))
			actualContext, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(0)
			Expect(actualContext).To(Equal(ctx))
			Expect(actualStatus.Ready).To(BeFalse())
			Expect(actualCluster).To(Equal(kindCluster))
		})

		It("creates a cluster using the cluster provider", func() {
			Expect(clusterProvider.CreateCallCount()).To(Equal(1))
			name := clusterProvider.CreateArgsForCall(0)
			Expect(name).To(Equal("the-kind-cluster-name"))
		})

		It("registers the finalizer", func() {
			Expect(kindClusterClient.AddFinalizerCallCount()).To(Equal(1))
			_, actualCluster := kindClusterClient.AddFinalizerArgsForCall(0)
			Expect(actualCluster).To(Equal(kindCluster))
		})

		It("updates the status to ready", func() {
			Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(2))
			actualContext, actualStatus, actualCluster := kindClusterClient.UpdateStatusArgsForCall(1)
			Expect(actualContext).To(Equal(ctx))
			Expect(actualStatus.Ready).To(BeTrue())
			Expect(actualCluster).To(Equal(kindCluster))
		})

		It("gets the control plane endpoint", func() {
			Expect(clusterProvider.GetControlPlaneEndpointCallCount()).To(Equal(1))
			name := clusterProvider.GetControlPlaneEndpointArgsForCall(0)
			Expect(name).To(Equal("the-kind-cluster-name"))
		})

		It("sets the control plane endpoint", func() {
			Expect(kindClusterClient.SetControlPlaneEndpointCallCount()).To(Equal(1))
			actualContext, actualEndpoint, actualCluster := kindClusterClient.SetControlPlaneEndpointArgsForCall(0)
			Expect(actualContext).To(Equal(ctx))
			Expect(actualEndpoint.Host).To(Equal("127.0.0.1"))
			Expect(actualEndpoint.Port).To(Equal(1337))
			Expect(actualCluster).To(Equal(kindCluster))
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

		When("adding the finalizer fails", func() {
			BeforeEach(func() {
				kindClusterClient.AddFinalizerReturns(errors.New("boom"))
			})

			It("returns an error", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})

			It("should not try to create the cluster", func() {
				Expect(clusterProvider.CreateCallCount()).To(Equal(0))
			})
		})

		When("the real kind cluster already exists", func() {
			BeforeEach(func() {
				clusterProvider.ExistsReturns(true, nil)
			})

			It("reconciles successfully", func() {
				Expect(result.Requeue).To(BeFalse())
				Expect(reconcileErr).NotTo(HaveOccurred())
			})

			It("should not try to create the cluster", func() {
				Expect(clusterProvider.CreateCallCount()).To(Equal(0))
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

		When("creating the cluster fails", func() {
			BeforeEach(func() {
				clusterProvider.CreateReturns(errors.New("boom"))
			})

			It("requeues the event", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})
		})

		When("setting the not ready status fails", func() {
			BeforeEach(func() {
				kindClusterClient.UpdateStatusReturns(errors.New("boom"))
			})

			It("requeues the event", func() {
				Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(1))
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})
		})

		When("getting the control plane endpoint fails", func() {
			BeforeEach(func() {
				clusterProvider.GetControlPlaneEndpointReturns("", 0, errors.New("boom"))
			})

			It("requeues the event", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})
		})

		When("setting the control plane endpoint fails", func() {
			BeforeEach(func() {
				kindClusterClient.SetControlPlaneEndpointReturns(errors.New("boom"))
			})

			It("requeues the event", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})
		})

		When("setting the ready status fails", func() {
			BeforeEach(func() {
				kindClusterClient.UpdateStatusReturnsOnCall(1, errors.New("boom"))
			})

			It("requeues the event", func() {
				Expect(kindClusterClient.UpdateStatusCallCount()).To(Equal(2))
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
			})
		})
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			now := metav1.NewTime(time.Now())
			kindCluster.DeletionTimestamp = &now
			kindClusterClient.GetReturns(kindCluster, nil)
		})

		It("deletes the cluster", func() {
			Expect(clusterProvider.DeleteCallCount()).To(Equal(1))
		})

		It("removes the finalizer", func() {
			Expect(kindClusterClient.RemoveFinalizerCallCount()).To(Equal(1))
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
