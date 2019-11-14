// Code generated by counterfeiter. DO NOT EDIT.
package releasefakes

import (
	"sync"

	pivnet "github.com/pivotal-cf/go-pivnet/v3"
)

type ReleaseImageReferencesAdderClient struct {
	AddImageReferenceStub        func(string, int, int) error
	addImageReferenceMutex       sync.RWMutex
	addImageReferenceArgsForCall []struct {
		arg1 string
		arg2 int
		arg3 int
	}
	addImageReferenceReturns struct {
		result1 error
	}
	addImageReferenceReturnsOnCall map[int]struct {
		result1 error
	}
	CreateImageReferenceStub        func(pivnet.CreateImageReferenceConfig) (pivnet.ImageReference, error)
	createImageReferenceMutex       sync.RWMutex
	createImageReferenceArgsForCall []struct {
		arg1 pivnet.CreateImageReferenceConfig
	}
	createImageReferenceReturns struct {
		result1 pivnet.ImageReference
		result2 error
	}
	createImageReferenceReturnsOnCall map[int]struct {
		result1 pivnet.ImageReference
		result2 error
	}
	ImageReferencesStub        func(string) ([]pivnet.ImageReference, error)
	imageReferencesMutex       sync.RWMutex
	imageReferencesArgsForCall []struct {
		arg1 string
	}
	imageReferencesReturns struct {
		result1 []pivnet.ImageReference
		result2 error
	}
	imageReferencesReturnsOnCall map[int]struct {
		result1 []pivnet.ImageReference
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *ReleaseImageReferencesAdderClient) AddImageReference(arg1 string, arg2 int, arg3 int) error {
	fake.addImageReferenceMutex.Lock()
	ret, specificReturn := fake.addImageReferenceReturnsOnCall[len(fake.addImageReferenceArgsForCall)]
	fake.addImageReferenceArgsForCall = append(fake.addImageReferenceArgsForCall, struct {
		arg1 string
		arg2 int
		arg3 int
	}{arg1, arg2, arg3})
	fake.recordInvocation("AddImageReference", []interface{}{arg1, arg2, arg3})
	fake.addImageReferenceMutex.Unlock()
	if fake.AddImageReferenceStub != nil {
		return fake.AddImageReferenceStub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1
	}
	fakeReturns := fake.addImageReferenceReturns
	return fakeReturns.result1
}

func (fake *ReleaseImageReferencesAdderClient) AddImageReferenceCallCount() int {
	fake.addImageReferenceMutex.RLock()
	defer fake.addImageReferenceMutex.RUnlock()
	return len(fake.addImageReferenceArgsForCall)
}

func (fake *ReleaseImageReferencesAdderClient) AddImageReferenceCalls(stub func(string, int, int) error) {
	fake.addImageReferenceMutex.Lock()
	defer fake.addImageReferenceMutex.Unlock()
	fake.AddImageReferenceStub = stub
}

func (fake *ReleaseImageReferencesAdderClient) AddImageReferenceArgsForCall(i int) (string, int, int) {
	fake.addImageReferenceMutex.RLock()
	defer fake.addImageReferenceMutex.RUnlock()
	argsForCall := fake.addImageReferenceArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *ReleaseImageReferencesAdderClient) AddImageReferenceReturns(result1 error) {
	fake.addImageReferenceMutex.Lock()
	defer fake.addImageReferenceMutex.Unlock()
	fake.AddImageReferenceStub = nil
	fake.addImageReferenceReturns = struct {
		result1 error
	}{result1}
}

func (fake *ReleaseImageReferencesAdderClient) AddImageReferenceReturnsOnCall(i int, result1 error) {
	fake.addImageReferenceMutex.Lock()
	defer fake.addImageReferenceMutex.Unlock()
	fake.AddImageReferenceStub = nil
	if fake.addImageReferenceReturnsOnCall == nil {
		fake.addImageReferenceReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.addImageReferenceReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *ReleaseImageReferencesAdderClient) CreateImageReference(arg1 pivnet.CreateImageReferenceConfig) (pivnet.ImageReference, error) {
	fake.createImageReferenceMutex.Lock()
	ret, specificReturn := fake.createImageReferenceReturnsOnCall[len(fake.createImageReferenceArgsForCall)]
	fake.createImageReferenceArgsForCall = append(fake.createImageReferenceArgsForCall, struct {
		arg1 pivnet.CreateImageReferenceConfig
	}{arg1})
	fake.recordInvocation("CreateImageReference", []interface{}{arg1})
	fake.createImageReferenceMutex.Unlock()
	if fake.CreateImageReferenceStub != nil {
		return fake.CreateImageReferenceStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.createImageReferenceReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ReleaseImageReferencesAdderClient) CreateImageReferenceCallCount() int {
	fake.createImageReferenceMutex.RLock()
	defer fake.createImageReferenceMutex.RUnlock()
	return len(fake.createImageReferenceArgsForCall)
}

func (fake *ReleaseImageReferencesAdderClient) CreateImageReferenceCalls(stub func(pivnet.CreateImageReferenceConfig) (pivnet.ImageReference, error)) {
	fake.createImageReferenceMutex.Lock()
	defer fake.createImageReferenceMutex.Unlock()
	fake.CreateImageReferenceStub = stub
}

func (fake *ReleaseImageReferencesAdderClient) CreateImageReferenceArgsForCall(i int) pivnet.CreateImageReferenceConfig {
	fake.createImageReferenceMutex.RLock()
	defer fake.createImageReferenceMutex.RUnlock()
	argsForCall := fake.createImageReferenceArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ReleaseImageReferencesAdderClient) CreateImageReferenceReturns(result1 pivnet.ImageReference, result2 error) {
	fake.createImageReferenceMutex.Lock()
	defer fake.createImageReferenceMutex.Unlock()
	fake.CreateImageReferenceStub = nil
	fake.createImageReferenceReturns = struct {
		result1 pivnet.ImageReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseImageReferencesAdderClient) CreateImageReferenceReturnsOnCall(i int, result1 pivnet.ImageReference, result2 error) {
	fake.createImageReferenceMutex.Lock()
	defer fake.createImageReferenceMutex.Unlock()
	fake.CreateImageReferenceStub = nil
	if fake.createImageReferenceReturnsOnCall == nil {
		fake.createImageReferenceReturnsOnCall = make(map[int]struct {
			result1 pivnet.ImageReference
			result2 error
		})
	}
	fake.createImageReferenceReturnsOnCall[i] = struct {
		result1 pivnet.ImageReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseImageReferencesAdderClient) ImageReferences(arg1 string) ([]pivnet.ImageReference, error) {
	fake.imageReferencesMutex.Lock()
	ret, specificReturn := fake.imageReferencesReturnsOnCall[len(fake.imageReferencesArgsForCall)]
	fake.imageReferencesArgsForCall = append(fake.imageReferencesArgsForCall, struct {
		arg1 string
	}{arg1})
	fake.recordInvocation("ImageReferences", []interface{}{arg1})
	fake.imageReferencesMutex.Unlock()
	if fake.ImageReferencesStub != nil {
		return fake.ImageReferencesStub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	fakeReturns := fake.imageReferencesReturns
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *ReleaseImageReferencesAdderClient) ImageReferencesCallCount() int {
	fake.imageReferencesMutex.RLock()
	defer fake.imageReferencesMutex.RUnlock()
	return len(fake.imageReferencesArgsForCall)
}

func (fake *ReleaseImageReferencesAdderClient) ImageReferencesCalls(stub func(string) ([]pivnet.ImageReference, error)) {
	fake.imageReferencesMutex.Lock()
	defer fake.imageReferencesMutex.Unlock()
	fake.ImageReferencesStub = stub
}

func (fake *ReleaseImageReferencesAdderClient) ImageReferencesArgsForCall(i int) string {
	fake.imageReferencesMutex.RLock()
	defer fake.imageReferencesMutex.RUnlock()
	argsForCall := fake.imageReferencesArgsForCall[i]
	return argsForCall.arg1
}

func (fake *ReleaseImageReferencesAdderClient) ImageReferencesReturns(result1 []pivnet.ImageReference, result2 error) {
	fake.imageReferencesMutex.Lock()
	defer fake.imageReferencesMutex.Unlock()
	fake.ImageReferencesStub = nil
	fake.imageReferencesReturns = struct {
		result1 []pivnet.ImageReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseImageReferencesAdderClient) ImageReferencesReturnsOnCall(i int, result1 []pivnet.ImageReference, result2 error) {
	fake.imageReferencesMutex.Lock()
	defer fake.imageReferencesMutex.Unlock()
	fake.ImageReferencesStub = nil
	if fake.imageReferencesReturnsOnCall == nil {
		fake.imageReferencesReturnsOnCall = make(map[int]struct {
			result1 []pivnet.ImageReference
			result2 error
		})
	}
	fake.imageReferencesReturnsOnCall[i] = struct {
		result1 []pivnet.ImageReference
		result2 error
	}{result1, result2}
}

func (fake *ReleaseImageReferencesAdderClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.addImageReferenceMutex.RLock()
	defer fake.addImageReferenceMutex.RUnlock()
	fake.createImageReferenceMutex.RLock()
	defer fake.createImageReferenceMutex.RUnlock()
	fake.imageReferencesMutex.RLock()
	defer fake.imageReferencesMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *ReleaseImageReferencesAdderClient) recordInvocation(key string, args []interface{}) {
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
