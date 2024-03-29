// Code generated by counterfeiter. DO NOT EDIT.
package releasefakes

import (
	"sync"
)

type Md5Summer struct {
	SumFileStub        func(string) (string, error)
	sumFileMutex       sync.RWMutex
	sumFileArgsForCall []struct {
		arg1 string
	}
	sumFileReturns struct {
		result1 string
		result2 error
	}
	sumFileReturnsOnCall map[int]struct {
		result1 string
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *Md5Summer) SumFile(arg1 string) (string, error) {
	fake.sumFileMutex.Lock()
	ret, specificReturn := fake.sumFileReturnsOnCall[len(fake.sumFileArgsForCall)]
	fake.sumFileArgsForCall = append(fake.sumFileArgsForCall, struct {
		arg1 string
	}{arg1})
	stub := fake.SumFileStub
	fakeReturns := fake.sumFileReturns
	fake.recordInvocation("SumFile", []interface{}{arg1})
	fake.sumFileMutex.Unlock()
	if stub != nil {
		return stub(arg1)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *Md5Summer) SumFileCallCount() int {
	fake.sumFileMutex.RLock()
	defer fake.sumFileMutex.RUnlock()
	return len(fake.sumFileArgsForCall)
}

func (fake *Md5Summer) SumFileCalls(stub func(string) (string, error)) {
	fake.sumFileMutex.Lock()
	defer fake.sumFileMutex.Unlock()
	fake.SumFileStub = stub
}

func (fake *Md5Summer) SumFileArgsForCall(i int) string {
	fake.sumFileMutex.RLock()
	defer fake.sumFileMutex.RUnlock()
	argsForCall := fake.sumFileArgsForCall[i]
	return argsForCall.arg1
}

func (fake *Md5Summer) SumFileReturns(result1 string, result2 error) {
	fake.sumFileMutex.Lock()
	defer fake.sumFileMutex.Unlock()
	fake.SumFileStub = nil
	fake.sumFileReturns = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *Md5Summer) SumFileReturnsOnCall(i int, result1 string, result2 error) {
	fake.sumFileMutex.Lock()
	defer fake.sumFileMutex.Unlock()
	fake.SumFileStub = nil
	if fake.sumFileReturnsOnCall == nil {
		fake.sumFileReturnsOnCall = make(map[int]struct {
			result1 string
			result2 error
		})
	}
	fake.sumFileReturnsOnCall[i] = struct {
		result1 string
		result2 error
	}{result1, result2}
}

func (fake *Md5Summer) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.sumFileMutex.RLock()
	defer fake.sumFileMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *Md5Summer) recordInvocation(key string, args []interface{}) {
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
