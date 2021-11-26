package controllers_test

import (
	"context"
	"errors"
	"time"

	"github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"github.com/mnitchev/cluster-api-provider-kind/controllers"
	"github.com/mnitchev/cluster-api-provider-kind/controllers/controllersfakes"
	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/apimachinery/pkg/types"
	ctrl "sigs.k8s.io/controller-runtime"
)

var _ = Describe("KindclusterController", func() {
	var (
		reconciler        *controllers.KindClusterReconciler
		clusterProvider   *controllersfakes.FakeClusterProvider
		kindClusterClient *controllersfakes.FakeKindClusterClient
		ctx               context.Context
		result            ctrl.Result
		reconcileErr      error
		cluster           *v1alpha3.KindCluster
	)

	BeforeEach(func() {
		ctx = context.Background()
		clusterProvider = new(controllersfakes.FakeClusterProvider)
		kindClusterClient = new(controllersfakes.FakeKindClusterClient)
		reconciler = controllers.NewKindClusterReconciler(kindClusterClient, clusterProvider)

		cluster = &v1alpha3.KindCluster{
			ObjectMeta: metav1.ObjectMeta{
				Name:      "foo",
				Namespace: "bar",
			},
			Spec: v1alpha3.KindClusterSpec{
				Name: "the-kind-cluster-name",
			},
		}
		kindClusterClient.GetReturns(cluster, nil)
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

		It("gets the kind cluster using the api client", func() {
			Expect(kindClusterClient.GetCallCount()).To(Equal(1))
			actualCtx, namespacedName := kindClusterClient.GetArgsForCall(0)
			Expect(actualCtx).To(Equal(ctx))
			Expect(namespacedName.Name).To(Equal("foo"))
			Expect(namespacedName.Namespace).To(Equal("bar"))
		})

		It("creates a cluster using the cluster provider", func() {
			Expect(clusterProvider.CreateCallCount()).To(Equal(1))
			name := clusterProvider.CreateArgsForCall(0)
			Expect(name).To(Equal("the-kind-cluster-name"))
		})

		It("registers the finalizer", func() {
			Expect(kindClusterClient.AddFinalizerCallCount()).To(Equal(1))
			_, actualCluster := kindClusterClient.AddFinalizerArgsForCall(0)
			Expect(actualCluster).To(Equal(cluster))
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

		When("getting the kind cluster fails", func() {
			BeforeEach(func() {
				kindClusterClient.GetReturns(nil, errors.New("boom"))
			})

			It("requeues the event", func() {
				Expect(reconcileErr).To(MatchError(ContainSubstring("boom")))
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
	})

	Describe("Delete", func() {
		BeforeEach(func() {
			now := metav1.NewTime(time.Now())
			cluster.DeletionTimestamp = &now
			kindClusterClient.GetReturns(cluster, nil)
		})

		It("deletes the cluster", func() {
			Expect(clusterProvider.DeleteCallCount()).To(Equal(1))
		})

		It("removes the finalizer", func() {
			Expect(kindClusterClient.RemoveFinalizerCallCount()).To(Equal(1))
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
