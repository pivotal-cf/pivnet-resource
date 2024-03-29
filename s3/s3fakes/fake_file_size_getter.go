// Code generated by counterfeiter. DO NOT EDIT.
package s3fakes

import (
	"sync"
)

type FakeFileSizeGetter struct {
	FileSizeStub        func(string) (int64, error)
	fileSizeMutex       sync.RWMutex
	fileSizeArgsForCall []struct {
		arg1 string
	}
	fileSizeReturns struct {
		result1 int64
		result2 error
	}
	fileSizeReturnsOnCall map[int]struct {
		result1 int64
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeFileSizeGetter) FileSize(arg1 string) (int64, error) {
	fake.fileSizeMutex.Lock()
	ret, specificReturn := fake.fileSizeReturnsOnCall[len(fake.fileSizeArgsForCall)]
	fake.fileSizeArgsForCall = append(fake.fileSizeArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.FileSizeStub
	fakeReturns := fake.fileSizeReturns
	fake.recordInvocation("FileSize", []interface{}{arg1})
	fake.fileSizeMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeFileSizeGetter) FileSizeCallCount() int {
	fake.fileSizeMutex.RLock()
	defer fake.fileSizeMutex.RUnlock()
	return len(fake.fileSizeArgsForCall)
}

func (fake *FakeFileSizeGetter) FileSizeCalls(stub func(string) (int64, error)) {
	fake.fileSizeMutex.Lock()
	defer fake.fileSizeMutex.Unlock()
	fake.FileSizeStub = stub
}

func (fake *FakeFileSizeGetter) FileSizeArgsForCall(i int) string {
	fake.fileSizeMutex.RLock()
	defer fake.fileSizeMutex.RUnlock()
	argsForCall := fake.fileSizeArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeFileSizeGetter) FileSizeReturns(result1 int64, result2 error) {
	fake.fileSizeMutex.Lock()
	defer fake.fileSizeMutex.Unlock()
	fake.FileSizeStub = nil
	fake.fileSizeReturns = struct {
		result1 int64
		result2 error
	}{result1, result2}
}

func (fake *FakeFileSizeGetter) FileSizeReturnsOnCall(i int, result1 int64, result2 error) {
	fake.fileSizeMutex.Lock()
	defer fake.fileSizeMutex.Unlock()
	fake.FileSizeStub = nil
	if fake.fileSizeReturnsOnCall == nil {
		fake.fileSizeReturnsOnCall = make(map[int]struct {
			result1 int64
			result2 error
		})
	}
	fake.fileSizeReturnsOnCall[i] = struct {
		result1 int64
		result2 error
	}{result1, result2}
}

func (fake *FakeFileSizeGetter) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.fileSizeMutex.RLock()
	defer fake.fileSizeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeFileSizeGetter) recordInvocation(key string, args []interface{}) {
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
