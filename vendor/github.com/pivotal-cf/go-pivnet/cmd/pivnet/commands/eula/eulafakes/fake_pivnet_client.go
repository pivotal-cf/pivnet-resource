// This file was generated by counterfeiter
package eulafakes

import (
	"sync"

	go_pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands/eula"
)

type FakePivnetClient struct {
	AcceptEULAStub        func(productSlug string, releaseID int) error
	acceptEULAMutex       sync.RWMutex
	acceptEULAArgsForCall []struct {
		productSlug string
		releaseID   int
	}
	acceptEULAReturns struct {
		result1 error
	}
	EULAsStub        func() ([]go_pivnet.EULA, error)
	eULAsMutex       sync.RWMutex
	eULAsArgsForCall []struct{}
	eULAsReturns     struct {
		result1 []go_pivnet.EULA
		result2 error
	}
	EULAStub        func(eulaSlug string) (go_pivnet.EULA, error)
	eULAMutex       sync.RWMutex
	eULAArgsForCall []struct {
		eulaSlug string
	}
	eULAReturns struct {
		result1 go_pivnet.EULA
		result2 error
	}
	ReleaseForProductVersionStub        func(productSlug string, releaseVersion string) (go_pivnet.Release, error)
	releaseForProductVersionMutex       sync.RWMutex
	releaseForProductVersionArgsForCall []struct {
		productSlug    string
		releaseVersion string
	}
	releaseForProductVersionReturns struct {
		result1 go_pivnet.Release
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakePivnetClient) AcceptEULA(productSlug string, releaseID int) error {
	fake.acceptEULAMutex.Lock()
	fake.acceptEULAArgsForCall = append(fake.acceptEULAArgsForCall, struct {
		productSlug string
		releaseID   int
	}{productSlug, releaseID})
	fake.recordInvocation("AcceptEULA", []interface{}{productSlug, releaseID})
	fake.acceptEULAMutex.Unlock()
	if fake.AcceptEULAStub != nil {
		return fake.AcceptEULAStub(productSlug, releaseID)
	} else {
		return fake.acceptEULAReturns.result1
	}
}

func (fake *FakePivnetClient) AcceptEULACallCount() int {
	fake.acceptEULAMutex.RLock()
	defer fake.acceptEULAMutex.RUnlock()
	return len(fake.acceptEULAArgsForCall)
}

func (fake *FakePivnetClient) AcceptEULAArgsForCall(i int) (string, int) {
	fake.acceptEULAMutex.RLock()
	defer fake.acceptEULAMutex.RUnlock()
	return fake.acceptEULAArgsForCall[i].productSlug, fake.acceptEULAArgsForCall[i].releaseID
}

func (fake *FakePivnetClient) AcceptEULAReturns(result1 error) {
	fake.AcceptEULAStub = nil
	fake.acceptEULAReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakePivnetClient) EULAs() ([]go_pivnet.EULA, error) {
	fake.eULAsMutex.Lock()
	fake.eULAsArgsForCall = append(fake.eULAsArgsForCall, struct{}{})
	fake.recordInvocation("EULAs", []interface{}{})
	fake.eULAsMutex.Unlock()
	if fake.EULAsStub != nil {
		return fake.EULAsStub()
	} else {
		return fake.eULAsReturns.result1, fake.eULAsReturns.result2
	}
}

func (fake *FakePivnetClient) EULAsCallCount() int {
	fake.eULAsMutex.RLock()
	defer fake.eULAsMutex.RUnlock()
	return len(fake.eULAsArgsForCall)
}

func (fake *FakePivnetClient) EULAsReturns(result1 []go_pivnet.EULA, result2 error) {
	fake.EULAsStub = nil
	fake.eULAsReturns = struct {
		result1 []go_pivnet.EULA
		result2 error
	}{result1, result2}
}

func (fake *FakePivnetClient) EULA(eulaSlug string) (go_pivnet.EULA, error) {
	fake.eULAMutex.Lock()
	fake.eULAArgsForCall = append(fake.eULAArgsForCall, struct {
		eulaSlug string
	}{eulaSlug})
	fake.recordInvocation("EULA", []interface{}{eulaSlug})
	fake.eULAMutex.Unlock()
	if fake.EULAStub != nil {
		return fake.EULAStub(eulaSlug)
	} else {
		return fake.eULAReturns.result1, fake.eULAReturns.result2
	}
}

func (fake *FakePivnetClient) EULACallCount() int {
	fake.eULAMutex.RLock()
	defer fake.eULAMutex.RUnlock()
	return len(fake.eULAArgsForCall)
}

func (fake *FakePivnetClient) EULAArgsForCall(i int) string {
	fake.eULAMutex.RLock()
	defer fake.eULAMutex.RUnlock()
	return fake.eULAArgsForCall[i].eulaSlug
}

func (fake *FakePivnetClient) EULAReturns(result1 go_pivnet.EULA, result2 error) {
	fake.EULAStub = nil
	fake.eULAReturns = struct {
		result1 go_pivnet.EULA
		result2 error
	}{result1, result2}
}

func (fake *FakePivnetClient) ReleaseForProductVersion(productSlug string, releaseVersion string) (go_pivnet.Release, error) {
	fake.releaseForProductVersionMutex.Lock()
	fake.releaseForProductVersionArgsForCall = append(fake.releaseForProductVersionArgsForCall, struct {
		productSlug    string
		releaseVersion string
	}{productSlug, releaseVersion})
	fake.recordInvocation("ReleaseForProductVersion", []interface{}{productSlug, releaseVersion})
	fake.releaseForProductVersionMutex.Unlock()
	if fake.ReleaseForProductVersionStub != nil {
		return fake.ReleaseForProductVersionStub(productSlug, releaseVersion)
	} else {
		return fake.releaseForProductVersionReturns.result1, fake.releaseForProductVersionReturns.result2
	}
}

func (fake *FakePivnetClient) ReleaseForProductVersionCallCount() int {
	fake.releaseForProductVersionMutex.RLock()
	defer fake.releaseForProductVersionMutex.RUnlock()
	return len(fake.releaseForProductVersionArgsForCall)
}

func (fake *FakePivnetClient) ReleaseForProductVersionArgsForCall(i int) (string, string) {
	fake.releaseForProductVersionMutex.RLock()
	defer fake.releaseForProductVersionMutex.RUnlock()
	return fake.releaseForProductVersionArgsForCall[i].productSlug, fake.releaseForProductVersionArgsForCall[i].releaseVersion
}

func (fake *FakePivnetClient) ReleaseForProductVersionReturns(result1 go_pivnet.Release, result2 error) {
	fake.ReleaseForProductVersionStub = nil
	fake.releaseForProductVersionReturns = struct {
		result1 go_pivnet.Release
		result2 error
	}{result1, result2}
}

func (fake *FakePivnetClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.acceptEULAMutex.RLock()
	defer fake.acceptEULAMutex.RUnlock()
	fake.eULAsMutex.RLock()
	defer fake.eULAsMutex.RUnlock()
	fake.eULAMutex.RLock()
	defer fake.eULAMutex.RUnlock()
	fake.releaseForProductVersionMutex.RLock()
	defer fake.releaseForProductVersionMutex.RUnlock()
	return fake.invocations
}

func (fake *FakePivnetClient) recordInvocation(key string, args []interface{}) {
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

var _ eula.PivnetClient = new(FakePivnetClient)