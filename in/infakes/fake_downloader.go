// Code generated by counterfeiter. DO NOT EDIT.
package infakes

import (
	"sync"

	pivnet "github.com/pivotal-cf/go-pivnet/v7"
)

type FakeDownloader struct {
	DownloadStub        func([]pivnet.ProductFile, string, int) ([]string, error)
	downloadMutex       sync.RWMutex
	downloadArgsForCall []struct {
		arg1 []pivnet.ProductFile
		arg2 string
		arg3 int
	}
	downloadReturns struct {
		result1 []string
		result2 error
	}
	downloadReturnsOnCall map[int]struct {
		result1 []string
		result2 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *FakeDownloader) Download(arg1 []pivnet.ProductFile, arg2 string, arg3 int) ([]string, error) {
	var arg1Copy []pivnet.ProductFile
	if arg1 != nil {
		arg1Copy = make([]pivnet.ProductFile, len(arg1))
		copy(arg1Copy, arg1)
	}
	fake.downloadMutex.Lock()
	ret, specificReturn := fake.downloadReturnsOnCall[len(fake.downloadArgsForCall)]
	fake.downloadArgsForCall = append(fake.downloadArgsForCall, struct {
		arg1 []pivnet.ProductFile
		arg2 string
		arg3 int
	}{arg1Copy, arg2, arg3})
	stub := fake.DownloadStub
	fakeReturns := fake.downloadReturns
	fake.recordInvocation("Download", []interface{}{arg1Copy, arg2, arg3})
	fake.downloadMutex.Unlock()
	if stub != nil {
		return stub(arg1, arg2, arg3)
	}
	if specificReturn {
		return ret.result1, ret.result2
	}
	return fakeReturns.result1, fakeReturns.result2
}

func (fake *FakeDownloader) DownloadCallCount() int {
	fake.downloadMutex.RLock()
	defer fake.downloadMutex.RUnlock()
	return len(fake.downloadArgsForCall)
}

func (fake *FakeDownloader) DownloadCalls(stub func([]pivnet.ProductFile, string, int) ([]string, error)) {
	fake.downloadMutex.Lock()
	defer fake.downloadMutex.Unlock()
	fake.DownloadStub = stub
}

func (fake *FakeDownloader) DownloadArgsForCall(i int) ([]pivnet.ProductFile, string, int) {
	fake.downloadMutex.RLock()
	defer fake.downloadMutex.RUnlock()
	argsForCall := fake.downloadArgsForCall[i]
	return argsForCall.arg1, argsForCall.arg2, argsForCall.arg3
}

func (fake *FakeDownloader) DownloadReturns(result1 []string, result2 error) {
	fake.downloadMutex.Lock()
	defer fake.downloadMutex.Unlock()
	fake.DownloadStub = nil
	fake.downloadReturns = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeDownloader) DownloadReturnsOnCall(i int, result1 []string, result2 error) {
	fake.downloadMutex.Lock()
	defer fake.downloadMutex.Unlock()
	fake.DownloadStub = nil
	if fake.downloadReturnsOnCall == nil {
		fake.downloadReturnsOnCall = make(map[int]struct {
			result1 []string
			result2 error
		})
	}
	fake.downloadReturnsOnCall[i] = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

func (fake *FakeDownloader) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.downloadMutex.RLock()
	defer fake.downloadMutex.RUnlock()
	copiedInvocations := map[string][][]interface{}{}
	for key, value := range fake.invocations {
		copiedInvocations[key] = value
	}
	return copiedInvocations
}

func (fake *FakeDownloader) recordInvocation(key string, args []interface{}) {
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
