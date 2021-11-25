// Code generated by counterfeiter. DO NOT EDIT.
package controllersfakes

import (
	"context"
	"sync"

	"github.com/mnitchev/cluster-api-provider-kind/api/v1alpha3"
	"github.com/mnitchev/cluster-api-provider-kind/controllers"
	"k8s.io/apimachinery/pkg/types"
)

type FakeAPIClient struct {
	GetKindClusterStub        func(context.Context, types.NamespacedName) (*v1alpha3.KindCluster, error)
	getKindClusterMutex       sync.RWMutex
	getKindClusterArgsForCall []struct {
		arg1 context.Context
		arg2 types.NamespacedName
	}
	getKindClusterReturns struct {
		result1 *v1alpha3.KindCluster
		result2 error
	}
	getKindClusterReturnsOnCall map[int]struct {
		result1 *v1alpha3.KindCluster
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeAPIClient) GetKindCluster(arg1 context.Context, arg2 types.NamespacedName) (*v1alpha3.KindCluster, error) {
	fake.getKindClusterMutex.Lock()
	ret, specificReturn := fake.getKindClusterReturnsOnCall[len(fake.getKindClusterArgsForCall)]
	fake.getKindClusterArgsForCall = append(fake.getKindClusterArgsForCall, struct {
		arg1 context.Context
		arg2 types.NamespacedName
	}{arg1, arg2})
	stub := fake.GetKindClusterStub
	fakeReturns := fake.getKindClusterReturns
	fake.recordInvocation("GetKindCluster", []interface{}{arg1, arg2})
	fake.getKindClusterMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeAPIClient) GetKindClusterCallCount() int {
	fake.getKindClusterMutex.RLock()
	defer fake.getKindClusterMutex.RUnlock()
	return len(fake.getKindClusterArgsForCall)
}

func (fake *FakeAPIClient) GetKindClusterCalls(stub func(context.Context, types.NamespacedName) (*v1alpha3.KindCluster, error)) {
	fake.getKindClusterMutex.Lock()
	defer fake.getKindClusterMutex.Unlock()
	fake.GetKindClusterStub = stub
}

func (fake *FakeAPIClient) GetKindClusterArgsForCall(i int) (context.Context, types.NamespacedName) {
	fake.getKindClusterMutex.RLock()
	defer fake.getKindClusterMutex.RUnlock()
	argsForCall := fake.getKindClusterArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeAPIClient) GetKindClusterReturns(result1 *v1alpha3.KindCluster, result2 error) {
	fake.getKindClusterMutex.Lock()
	defer fake.getKindClusterMutex.Unlock()
	fake.GetKindClusterStub = nil
	fake.getKindClusterReturns = struct {
		result1 *v1alpha3.KindCluster
		result2 error
	}{result1, result2}
}

func (fake *FakeAPIClient) GetKindClusterReturnsOnCall(i int, result1 *v1alpha3.KindCluster, result2 error) {
	fake.getKindClusterMutex.Lock()
	defer fake.getKindClusterMutex.Unlock()
	fake.GetKindClusterStub = nil
	if fake.getKindClusterReturnsOnCall == nil {
		fake.getKindClusterReturnsOnCall = make(map[int]struct {
			result1 *v1alpha3.KindCluster
			result2 error
		})
	}
	fake.getKindClusterReturnsOnCall[i] = struct {
		result1 *v1alpha3.KindCluster
		result2 error
	}{result1, result2}
}

func (fake *FakeAPIClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.getKindClusterMutex.RLock()
	defer fake.getKindClusterMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeAPIClient) recordInvocation(key string, args []interface{}) {
	fake.invocationsMutex.Lock()
	defer fake.invocationsMutex.Unlock()
	if fake.invocations == nil {
		fake.invocations = map[string][][]interface{}{}
	}
	if fake.invocations[key] == nil {
		fake.invocations[key] = [][]interface{}{}
	}
	fake.invocations[key] = append(fake.invocations[key], args)
}

var _ controllers.APIClient = new(FakeAPIClient)
