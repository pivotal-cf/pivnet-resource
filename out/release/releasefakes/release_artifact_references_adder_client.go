// Code generated by counterfeiter. DO NOT EDIT.
package releasefakes

import (
	"sync"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
)

type ReleaseArtifactReferencesAdderClient struct {
	AddArtifactReferenceStub        func(string, int, int) error
	addArtifactReferenceMutex       sync.RWMutex
	addArtifactReferenceArgsForCall []struct {
		arg1 string
		arg2 int
		arg3 int
	}
	addArtifactReferenceReturns struct {
		result1 error
	}
	addArtifactReferenceReturnsOnCall map[int]struct {
		result1 error
	}
	ArtifactReferencesStub        func(string) ([]pivnet.ArtifactReference, error)
	artifactReferencesMutex       sync.RWMutex
	artifactReferencesArgsForCall []struct {
		arg1 string
	}
	artifactReferencesReturns struct {
		result1 []pivnet.ArtifactReference
		result2 error
	}
	artifactReferencesReturnsOnCall map[int]struct {
		result1 []pivnet.ArtifactReference
		result2 error
	}
	ArtifactReferencesForDigestStub        func(string, string) ([]pivnet.ArtifactReference, error)
	artifactReferencesForDigestMutex       sync.RWMutex
	artifactReferencesForDigestArgsForCall []struct {
		arg1 string
		arg2 string
	}
	artifactReferencesForDigestReturns struct {
		result1 []pivnet.ArtifactReference
		result2 error
	}
	artifactReferencesForDigestReturnsOnCall map[int]struct {
		result1 []pivnet.ArtifactReference
		result2 error
	}
	CreateArtifactReferenceStub        func(pivnet.CreateArtifactReferenceConfig) (pivnet.ArtifactReference, error)
	createArtifactReferenceMutex       sync.RWMutex
	createArtifactReferenceArgsForCall []struct {
		arg1 pivnet.CreateArtifactReferenceConfig
	}
	createArtifactReferenceReturns struct {
		result1 pivnet.ArtifactReference
		result2 error
	}
	createArtifactReferenceReturnsOnCall map[int]struct {
		result1 pivnet.ArtifactReference
		result2 error
	}
	DeleteArtifactReferenceStub        func(string, int) (pivnet.ArtifactReference, error)
	deleteArtifactReferenceMutex       sync.RWMutex
	deleteArtifactReferenceArgsForCall []struct {
		arg1 string
		arg2 int
	}
	deleteArtifactReferenceReturns struct {
		result1 pivnet.ArtifactReference
		result2 error
	}
	deleteArtifactReferenceReturnsOnCall map[int]struct {
		result1 pivnet.ArtifactReference
		result2 error
	}
	GetArtifactReferenceStub        func(string, int) (pivnet.ArtifactReference, error)
	getArtifactReferenceMutex       sync.RWMutex
	getArtifactReferenceArgsForCall []struct {
		arg1 string
		arg2 int
	}
	getArtifactReferenceReturns struct {
		result1 pivnet.ArtifactReference
		result2 error
	}
	getArtifactReferenceReturnsOnCall map[int]struct {
		result1 pivnet.ArtifactReference
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ReleaseArtifactReferencesAdderClient) AddArtifactReference(arg1 string, arg2 int, arg3 int) error {
	fake.addArtifactReferenceMutex.Lock()
	ret, specificReturn := fake.addArtifactReferenceReturnsOnCall[len(fake.addArtifactReferenceArgsForCall)]
	fake.addArtifactReferenceArgsForCall = append(fake.addArtifactReferenceArgsForCall, struct {
		arg1 string
		arg2 int
		arg3 int
	}{arg1, arg2, arg3})
	stub := fake.AddArtifactReferenceStub
	fakeReturns := fake.addArtifactReferenceReturns
	fake.recordInvocation("AddArtifactReference", []interface{}{arg1, arg2, arg3})
	fake.addArtifactReferenceMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *ReleaseArtifactReferencesAdderClient) AddArtifactReferenceCallCount() int {
	fake.addArtifactReferenceMutex.RLock()
	defer fake.addArtifactReferenceMutex.RUnlock()
	return len(fake.addArtifactReferenceArgsForCall)
}

func (fake *ReleaseArtifactReferencesAdderClient) AddArtifactReferenceCalls(stub func(string, int, int) error) {
	fake.addArtifactReferenceMutex.Lock()
	defer fake.addArtifactReferenceMutex.Unlock()
	fake.AddArtifactReferenceStub = stub
}

func (fake *ReleaseArtifactReferencesAdderClient) AddArtifactReferenceArgsForCall(i int) (string, int, int) {
	fake.addArtifactReferenceMutex.RLock()
	defer fake.addArtifactReferenceMutex.RUnlock()
	argsForCall := fake.addArtifactReferenceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *ReleaseArtifactReferencesAdderClient) AddArtifactReferenceReturns(result1 error) {
	fake.addArtifactReferenceMutex.Lock()
	defer fake.addArtifactReferenceMutex.Unlock()
	fake.AddArtifactReferenceStub = nil
	fake.addArtifactReferenceReturns = struct {
		result1 error
	}{result1}
}

func (fake *ReleaseArtifactReferencesAdderClient) AddArtifactReferenceReturnsOnCall(i int, result1 error) {
	fake.addArtifactReferenceMutex.Lock()
	defer fake.addArtifactReferenceMutex.Unlock()
	fake.AddArtifactReferenceStub = nil
	if fake.addArtifactReferenceReturnsOnCall == nil {
		fake.addArtifactReferenceReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.addArtifactReferenceReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferences(arg1 string) ([]pivnet.ArtifactReference, error) {
	fake.artifactReferencesMutex.Lock()
	ret, specificReturn := fake.artifactReferencesReturnsOnCall[len(fake.artifactReferencesArgsForCall)]
	fake.artifactReferencesArgsForCall = append(fake.artifactReferencesArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.ArtifactReferencesStub
	fakeReturns := fake.artifactReferencesReturns
	fake.recordInvocation("ArtifactReferences", []interface{}{arg1})
	fake.artifactReferencesMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesCallCount() int {
	fake.artifactReferencesMutex.RLock()
	defer fake.artifactReferencesMutex.RUnlock()
	return len(fake.artifactReferencesArgsForCall)
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesCalls(stub func(string) ([]pivnet.ArtifactReference, error)) {
	fake.artifactReferencesMutex.Lock()
	defer fake.artifactReferencesMutex.Unlock()
	fake.ArtifactReferencesStub = stub
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesArgsForCall(i int) string {
	fake.artifactReferencesMutex.RLock()
	defer fake.artifactReferencesMutex.RUnlock()
	argsForCall := fake.artifactReferencesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesReturns(result1 []pivnet.ArtifactReference, result2 error) {
	fake.artifactReferencesMutex.Lock()
	defer fake.artifactReferencesMutex.Unlock()
	fake.ArtifactReferencesStub = nil
	fake.artifactReferencesReturns = struct {
		result1 []pivnet.ArtifactReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesReturnsOnCall(i int, result1 []pivnet.ArtifactReference, result2 error) {
	fake.artifactReferencesMutex.Lock()
	defer fake.artifactReferencesMutex.Unlock()
	fake.ArtifactReferencesStub = nil
	if fake.artifactReferencesReturnsOnCall == nil {
		fake.artifactReferencesReturnsOnCall = make(map[int]struct {
			result1 []pivnet.ArtifactReference
			result2 error
		})
	}
	fake.artifactReferencesReturnsOnCall[i] = struct {
		result1 []pivnet.ArtifactReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesForDigest(arg1 string, arg2 string) ([]pivnet.ArtifactReference, error) {
	fake.artifactReferencesForDigestMutex.Lock()
	ret, specificReturn := fake.artifactReferencesForDigestReturnsOnCall[len(fake.artifactReferencesForDigestArgsForCall)]
	fake.artifactReferencesForDigestArgsForCall = append(fake.artifactReferencesForDigestArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.ArtifactReferencesForDigestStub
	fakeReturns := fake.artifactReferencesForDigestReturns
	fake.recordInvocation("ArtifactReferencesForDigest", []interface{}{arg1, arg2})
	fake.artifactReferencesForDigestMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesForDigestCallCount() int {
	fake.artifactReferencesForDigestMutex.RLock()
	defer fake.artifactReferencesForDigestMutex.RUnlock()
	return len(fake.artifactReferencesForDigestArgsForCall)
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesForDigestCalls(stub func(string, string) ([]pivnet.ArtifactReference, error)) {
	fake.artifactReferencesForDigestMutex.Lock()
	defer fake.artifactReferencesForDigestMutex.Unlock()
	fake.ArtifactReferencesForDigestStub = stub
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesForDigestArgsForCall(i int) (string, string) {
	fake.artifactReferencesForDigestMutex.RLock()
	defer fake.artifactReferencesForDigestMutex.RUnlock()
	argsForCall := fake.artifactReferencesForDigestArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesForDigestReturns(result1 []pivnet.ArtifactReference, result2 error) {
	fake.artifactReferencesForDigestMutex.Lock()
	defer fake.artifactReferencesForDigestMutex.Unlock()
	fake.ArtifactReferencesForDigestStub = nil
	fake.artifactReferencesForDigestReturns = struct {
		result1 []pivnet.ArtifactReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseArtifactReferencesAdderClient) ArtifactReferencesForDigestReturnsOnCall(i int, result1 []pivnet.ArtifactReference, result2 error) {
	fake.artifactReferencesForDigestMutex.Lock()
	defer fake.artifactReferencesForDigestMutex.Unlock()
	fake.ArtifactReferencesForDigestStub = nil
	if fake.artifactReferencesForDigestReturnsOnCall == nil {
		fake.artifactReferencesForDigestReturnsOnCall = make(map[int]struct {
			result1 []pivnet.ArtifactReference
			result2 error
		})
	}
	fake.artifactReferencesForDigestReturnsOnCall[i] = struct {
		result1 []pivnet.ArtifactReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseArtifactReferencesAdderClient) CreateArtifactReference(arg1 pivnet.CreateArtifactReferenceConfig) (pivnet.ArtifactReference, error) {
	fake.createArtifactReferenceMutex.Lock()
	ret, specificReturn := fake.createArtifactReferenceReturnsOnCall[len(fake.createArtifactReferenceArgsForCall)]
	fake.createArtifactReferenceArgsForCall = append(fake.createArtifactReferenceArgsForCall, struct {
		arg1 pivnet.CreateArtifactReferenceConfig
	}{arg1})
	stub := fake.CreateArtifactReferenceStub
	fakeReturns := fake.createArtifactReferenceReturns
	fake.recordInvocation("CreateArtifactReference", []interface{}{arg1})
	fake.createArtifactReferenceMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ReleaseArtifactReferencesAdderClient) CreateArtifactReferenceCallCount() int {
	fake.createArtifactReferenceMutex.RLock()
	defer fake.createArtifactReferenceMutex.RUnlock()
	return len(fake.createArtifactReferenceArgsForCall)
}

func (fake *ReleaseArtifactReferencesAdderClient) CreateArtifactReferenceCalls(stub func(pivnet.CreateArtifactReferenceConfig) (pivnet.ArtifactReference, error)) {
	fake.createArtifactReferenceMutex.Lock()
	defer fake.createArtifactReferenceMutex.Unlock()
	fake.CreateArtifactReferenceStub = stub
}

func (fake *ReleaseArtifactReferencesAdderClient) CreateArtifactReferenceArgsForCall(i int) pivnet.CreateArtifactReferenceConfig {
	fake.createArtifactReferenceMutex.RLock()
	defer fake.createArtifactReferenceMutex.RUnlock()
	argsForCall := fake.createArtifactReferenceArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ReleaseArtifactReferencesAdderClient) CreateArtifactReferenceReturns(result1 pivnet.ArtifactReference, result2 error) {
	fake.createArtifactReferenceMutex.Lock()
	defer fake.createArtifactReferenceMutex.Unlock()
	fake.CreateArtifactReferenceStub = nil
	fake.createArtifactReferenceReturns = struct {
		result1 pivnet.ArtifactReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseArtifactReferencesAdderClient) CreateArtifactReferenceReturnsOnCall(i int, result1 pivnet.ArtifactReference, result2 error) {
	fake.createArtifactReferenceMutex.Lock()
	defer fake.createArtifactReferenceMutex.Unlock()
	fake.CreateArtifactReferenceStub = nil
	if fake.createArtifactReferenceReturnsOnCall == nil {
		fake.createArtifactReferenceReturnsOnCall = make(map[int]struct {
			result1 pivnet.ArtifactReference
			result2 error
		})
	}
	fake.createArtifactReferenceReturnsOnCall[i] = struct {
		result1 pivnet.ArtifactReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseArtifactReferencesAdderClient) DeleteArtifactReference(arg1 string, arg2 int) (pivnet.ArtifactReference, error) {
	fake.deleteArtifactReferenceMutex.Lock()
	ret, specificReturn := fake.deleteArtifactReferenceReturnsOnCall[len(fake.deleteArtifactReferenceArgsForCall)]
	fake.deleteArtifactReferenceArgsForCall = append(fake.deleteArtifactReferenceArgsForCall, struct {
		arg1 string
		arg2 int
	}{arg1, arg2})
	stub := fake.DeleteArtifactReferenceStub
	fakeReturns := fake.deleteArtifactReferenceReturns
	fake.recordInvocation("DeleteArtifactReference", []interface{}{arg1, arg2})
	fake.deleteArtifactReferenceMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ReleaseArtifactReferencesAdderClient) DeleteArtifactReferenceCallCount() int {
	fake.deleteArtifactReferenceMutex.RLock()
	defer fake.deleteArtifactReferenceMutex.RUnlock()
	return len(fake.deleteArtifactReferenceArgsForCall)
}

func (fake *ReleaseArtifactReferencesAdderClient) DeleteArtifactReferenceCalls(stub func(string, int) (pivnet.ArtifactReference, error)) {
	fake.deleteArtifactReferenceMutex.Lock()
	defer fake.deleteArtifactReferenceMutex.Unlock()
	fake.DeleteArtifactReferenceStub = stub
}

func (fake *ReleaseArtifactReferencesAdderClient) DeleteArtifactReferenceArgsForCall(i int) (string, int) {
	fake.deleteArtifactReferenceMutex.RLock()
	defer fake.deleteArtifactReferenceMutex.RUnlock()
	argsForCall := fake.deleteArtifactReferenceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *ReleaseArtifactReferencesAdderClient) DeleteArtifactReferenceReturns(result1 pivnet.ArtifactReference, result2 error) {
	fake.deleteArtifactReferenceMutex.Lock()
	defer fake.deleteArtifactReferenceMutex.Unlock()
	fake.DeleteArtifactReferenceStub = nil
	fake.deleteArtifactReferenceReturns = struct {
		result1 pivnet.ArtifactReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseArtifactReferencesAdderClient) DeleteArtifactReferenceReturnsOnCall(i int, result1 pivnet.ArtifactReference, result2 error) {
	fake.deleteArtifactReferenceMutex.Lock()
	defer fake.deleteArtifactReferenceMutex.Unlock()
	fake.DeleteArtifactReferenceStub = nil
	if fake.deleteArtifactReferenceReturnsOnCall == nil {
		fake.deleteArtifactReferenceReturnsOnCall = make(map[int]struct {
			result1 pivnet.ArtifactReference
			result2 error
		})
	}
	fake.deleteArtifactReferenceReturnsOnCall[i] = struct {
		result1 pivnet.ArtifactReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseArtifactReferencesAdderClient) GetArtifactReference(arg1 string, arg2 int) (pivnet.ArtifactReference, error) {
	fake.getArtifactReferenceMutex.Lock()
	ret, specificReturn := fake.getArtifactReferenceReturnsOnCall[len(fake.getArtifactReferenceArgsForCall)]
	fake.getArtifactReferenceArgsForCall = append(fake.getArtifactReferenceArgsForCall, struct {
		arg1 string
		arg2 int
	}{arg1, arg2})
	stub := fake.GetArtifactReferenceStub
	fakeReturns := fake.getArtifactReferenceReturns
	fake.recordInvocation("GetArtifactReference", []interface{}{arg1, arg2})
	fake.getArtifactReferenceMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ReleaseArtifactReferencesAdderClient) GetArtifactReferenceCallCount() int {
	fake.getArtifactReferenceMutex.RLock()
	defer fake.getArtifactReferenceMutex.RUnlock()
	return len(fake.getArtifactReferenceArgsForCall)
}

func (fake *ReleaseArtifactReferencesAdderClient) GetArtifactReferenceCalls(stub func(string, int) (pivnet.ArtifactReference, error)) {
	fake.getArtifactReferenceMutex.Lock()
	defer fake.getArtifactReferenceMutex.Unlock()
	fake.GetArtifactReferenceStub = stub
}

func (fake *ReleaseArtifactReferencesAdderClient) GetArtifactReferenceArgsForCall(i int) (string, int) {
	fake.getArtifactReferenceMutex.RLock()
	defer fake.getArtifactReferenceMutex.RUnlock()
	argsForCall := fake.getArtifactReferenceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *ReleaseArtifactReferencesAdderClient) GetArtifactReferenceReturns(result1 pivnet.ArtifactReference, result2 error) {
	fake.getArtifactReferenceMutex.Lock()
	defer fake.getArtifactReferenceMutex.Unlock()
	fake.GetArtifactReferenceStub = nil
	fake.getArtifactReferenceReturns = struct {
		result1 pivnet.ArtifactReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseArtifactReferencesAdderClient) GetArtifactReferenceReturnsOnCall(i int, result1 pivnet.ArtifactReference, result2 error) {
	fake.getArtifactReferenceMutex.Lock()
	defer fake.getArtifactReferenceMutex.Unlock()
	fake.GetArtifactReferenceStub = nil
	if fake.getArtifactReferenceReturnsOnCall == nil {
		fake.getArtifactReferenceReturnsOnCall = make(map[int]struct {
			result1 pivnet.ArtifactReference
			result2 error
		})
	}
	fake.getArtifactReferenceReturnsOnCall[i] = struct {
		result1 pivnet.ArtifactReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseArtifactReferencesAdderClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.addArtifactReferenceMutex.RLock()
	defer fake.addArtifactReferenceMutex.RUnlock()
	fake.artifactReferencesMutex.RLock()
	defer fake.artifactReferencesMutex.RUnlock()
	fake.artifactReferencesForDigestMutex.RLock()
	defer fake.artifactReferencesForDigestMutex.RUnlock()
	fake.createArtifactReferenceMutex.RLock()
	defer fake.createArtifactReferenceMutex.RUnlock()
	fake.deleteArtifactReferenceMutex.RLock()
	defer fake.deleteArtifactReferenceMutex.RUnlock()
	fake.getArtifactReferenceMutex.RLock()
	defer fake.getArtifactReferenceMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ReleaseArtifactReferencesAdderClient) recordInvocation(key string, args []interface{}) {
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