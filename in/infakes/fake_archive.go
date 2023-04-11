// Code generated by counterfeiter. DO NOT EDIT.
package infakes

import (
	"sync"
)

type FakeArchive struct {
	ExtractStub        func(string, string) error
	extractMutex       sync.RWMutex
	extractArgsForCall []struct {
		arg1 string
		arg2 string
	}
	extractReturns struct {
		result1 error
	}
	extractReturnsOnCall map[int]struct {
		result1 error
	}
	MimetypeStub        func(string) string
	mimetypeMutex       sync.RWMutex
	mimetypeArgsForCall []struct {
		arg1 string
	}
	mimetypeReturns struct {
		result1 string
	}
	mimetypeReturnsOnCall map[int]struct {
		result1 string
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeArchive) Extract(arg1 string, arg2 string) error {
	fake.extractMutex.Lock()
	ret, specificReturn := fake.extractReturnsOnCall[len(fake.extractArgsForCall)]
	fake.extractArgsForCall = append(fake.extractArgsForCall, struct {
		arg1 string
		arg2 string
	}{arg1, arg2})
	stub := fake.ExtractStub
	fakeReturns := fake.extractReturns
	fake.recordInvocation("Extract", []interface{}{arg1, arg2})
	fake.extractMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeArchive) ExtractCallCount() int {
	fake.extractMutex.RLock()
	defer fake.extractMutex.RUnlock()
	return len(fake.extractArgsForCall)
}

func (fake *FakeArchive) ExtractCalls(stub func(string, string) error) {
	fake.extractMutex.Lock()
	defer fake.extractMutex.Unlock()
	fake.ExtractStub = stub
}

func (fake *FakeArchive) ExtractArgsForCall(i int) (string, string) {
	fake.extractMutex.RLock()
	defer fake.extractMutex.RUnlock()
	argsForCall := fake.extractArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2
}

func (fake *FakeArchive) ExtractReturns(result1 error) {
	fake.extractMutex.Lock()
	defer fake.extractMutex.Unlock()
	fake.ExtractStub = nil
	fake.extractReturns = struct {
		result1 error
	}{result1}
}

func (fake *FakeArchive) ExtractReturnsOnCall(i int, result1 error) {
	fake.extractMutex.Lock()
	defer fake.extractMutex.Unlock()
	fake.ExtractStub = nil
	if fake.extractReturnsOnCall == nil {
		fake.extractReturnsOnCall = make(map[int]struct {
			result1 error
		})
	}
	fake.extractReturnsOnCall[i] = struct {
		result1 error
	}{result1}
}

func (fake *FakeArchive) Mimetype(arg1 string) string {
	fake.mimetypeMutex.Lock()
	ret, specificReturn := fake.mimetypeReturnsOnCall[len(fake.mimetypeArgsForCall)]
	fake.mimetypeArgsForCall = append(fake.mimetypeArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.MimetypeStub
	fakeReturns := fake.mimetypeReturns
	fake.recordInvocation("Mimetype", []interface{}{arg1})
	fake.mimetypeMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1
	}
	return fakeReturns.result1
}

func (fake *FakeArchive) MimetypeCallCount() int {
	fake.mimetypeMutex.RLock()
	defer fake.mimetypeMutex.RUnlock()
	return len(fake.mimetypeArgsForCall)
}

func (fake *FakeArchive) MimetypeCalls(stub func(string) string) {
	fake.mimetypeMutex.Lock()
	defer fake.mimetypeMutex.Unlock()
	fake.MimetypeStub = stub
}

func (fake *FakeArchive) MimetypeArgsForCall(i int) string {
	fake.mimetypeMutex.RLock()
	defer fake.mimetypeMutex.RUnlock()
	argsForCall := fake.mimetypeArgsForCall[i]
	return argsForCall.arg1
}

func (fake *FakeArchive) MimetypeReturns(result1 string) {
	fake.mimetypeMutex.Lock()
	defer fake.mimetypeMutex.Unlock()
	fake.MimetypeStub = nil
	fake.mimetypeReturns = struct {
		result1 string
	}{result1}
}

func (fake *FakeArchive) MimetypeReturnsOnCall(i int, result1 string) {
	fake.mimetypeMutex.Lock()
	defer fake.mimetypeMutex.Unlock()
	fake.MimetypeStub = nil
	if fake.mimetypeReturnsOnCall == nil {
		fake.mimetypeReturnsOnCall = make(map[int]struct {
			result1 string
		})
	}
	fake.mimetypeReturnsOnCall[i] = struct {
		result1 string
	}{result1}
}

func (fake *FakeArchive) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.extractMutex.RLock()
	defer fake.extractMutex.RUnlock()
	fake.mimetypeMutex.RLock()
	defer fake.mimetypeMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeArchive) recordInvocation(key string, args []interface{}) {
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
