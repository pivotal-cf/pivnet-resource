// This file was generated by counterfeiter
package downloaderfakes

import (
	"sync"

	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"
)

type FakeDownloader struct {
	DownloadStub        func(downloadDir string, downloadLinks map[string]string, token string) ([]string, error)
	downloadMutex       sync.RWMutex
	downloadArgsForCall []struct {
		downloadDir   string
		downloadLinks map[string]string
		token         string
	}
	downloadReturns struct {
		result1 []string
		result2 error
	}
}

func (fake *FakeDownloader) Download(downloadDir string, downloadLinks map[string]string, token string) ([]string, error) {
	fake.downloadMutex.Lock()
	fake.downloadArgsForCall = append(fake.downloadArgsForCall, struct {
		downloadDir   string
		downloadLinks map[string]string
		token         string
	}{downloadDir, downloadLinks, token})
	fake.downloadMutex.Unlock()
	if fake.DownloadStub != nil {
		return fake.DownloadStub(downloadDir, downloadLinks, token)
	} else {
		return fake.downloadReturns.result1, fake.downloadReturns.result2
	}
}

func (fake *FakeDownloader) DownloadCallCount() int {
	fake.downloadMutex.RLock()
	defer fake.downloadMutex.RUnlock()
	return len(fake.downloadArgsForCall)
}

func (fake *FakeDownloader) DownloadArgsForCall(i int) (string, map[string]string, string) {
	fake.downloadMutex.RLock()
	defer fake.downloadMutex.RUnlock()
	return fake.downloadArgsForCall[i].downloadDir, fake.downloadArgsForCall[i].downloadLinks, fake.downloadArgsForCall[i].token
}

func (fake *FakeDownloader) DownloadReturns(result1 []string, result2 error) {
	fake.DownloadStub = nil
	fake.downloadReturns = struct {
		result1 []string
		result2 error
	}{result1, result2}
}

var _ downloader.Downloader = new(FakeDownloader)
