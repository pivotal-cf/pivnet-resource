// Code generated by counterfeiter. DO NOT EDIT.
package outfakes

import (
	"sync"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
)

type ReleaseUpgradePathsAdder struct {
	AddReleaseUpgradePathsStub        func(pivnet.Release) error
	addReleaseUpgradePathsMutex       sync.RWMutex
	addReleaseUpgradePathsArgsForCall []struct {
		arg1 pivnet.Release
	}
	addReleaseUpgradePathsReturns struct {
		result1 error
	}
	addReleaseUpgradePathsReturnsOnCall map[int]struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ReleaseUpgradePathsAdder) AddReleaseUpgradePaths(arg1 pivnet.Release) error {
	fake.addReleaseUpgradePathsMutex.Lock()
	ret, specificReturn := fake.addReleaseUpgradePathsReturnsOnCall[len(fake.addReleaseUpgradePathsArgsForCall)]
	fake.addReleaseUpgradePathsArgsForCall = append(fake.addReleaseUpgradePathsArgsForCall, struct {
		arg1 pivnet.Release
	}{arg1})
	stub := fake.AddReleaseUpgradePathsStub
	fakeReturns := fake.addReleaseUpgradePathsReturns
	fake.recordInvocation("AddReleaseUpgradePaths", []interface{}{arg1})
	fake.addReleaseUpgradePathsMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *ReleaseUpgradePathsAdder) AddReleaseUpgradePathsCallCount() int {
	fake.addReleaseUpgradePathsMutex.RLock()
	defer fake.addReleaseUpgradePathsMutex.RUnlock()
	return len(fake.addReleaseUpgradePathsArgsForCall)
}

func (fake *ReleaseUpgradePathsAdder) AddReleaseUpgradePathsCalls(stub func(pivnet.Release) error) {
	fake.addReleaseUpgradePathsMutex.Lock()
	defer fake.addReleaseUpgradePathsMutex.Unlock()
	fake.AddReleaseUpgradePathsStub = stub
}

func (fake *ReleaseUpgradePathsAdder) AddReleaseUpgradePathsArgsForCall(i int) pivnet.Release {
	fake.addReleaseUpgradePathsMutex.RLock()
	defer fake.addReleaseUpgradePathsMutex.RUnlock()
	argsForCall := fake.addReleaseUpgradePathsArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ReleaseUpgradePathsAdder) AddReleaseUpgradePathsReturns(result1 error) {
	fake.addReleaseUpgradePathsMutex.Lock()
	defer fake.addReleaseUpgradePathsMutex.Unlock()
	fake.AddReleaseUpgradePathsStub = nil
	fake.addReleaseUpgradePathsReturns = struct {
		result1 error
	}{result1}
}

func (fake *ReleaseUpgradePathsAdder) AddReleaseUpgradePathsReturnsOnCall(i int, result1 error) {
	fake.addReleaseUpgradePathsMutex.Lock()
	defer fake.addReleaseUpgradePathsMutex.Unlock()
	fake.AddReleaseUpgradePathsStub = nil
	if fake.addReleaseUpgradePathsReturnsOnCall == nil {
		fake.addReleaseUpgradePathsReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.addReleaseUpgradePathsReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ReleaseUpgradePathsAdder) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.addReleaseUpgradePathsMutex.RLock()
	defer fake.addReleaseUpgradePathsMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ReleaseUpgradePathsAdder) recordInvocation(key string, args []interface{}) {
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
