// This file was generated by counterfeiter
package releasefakes

import (
	"sync"

	go_pivnet "github.com/pivotal-cf/go-pivnet/v3"
)

type UserGroupsUpdaterClient struct {
	UpdateReleaseStub        func(productSlug string, release go_pivnet.Release) (go_pivnet.Release, error)
	updateReleaseMutex       sync.RWMutex
	updateReleaseArgsForCall []struct {
		productSlug string
		release     go_pivnet.Release
	}
	updateReleaseReturns struct {
		result1 go_pivnet.Release
		result2 error
	}
	AddUserGroupStub        func(productSlug string, releaseID int, userGroupID int) error
	addUserGroupMutex       sync.RWMutex
	addUserGroupArgsForCall []struct {
		productSlug string
		releaseID   int
		userGroupID int
	}
	addUserGroupReturns struct {
		result1 error
	}
	invocations      map[string][][]interface{}
	invocationsMutex sync.RWMutex
}

func (fake *UserGroupsUpdaterClient) UpdateRelease(productSlug string, release go_pivnet.Release) (go_pivnet.Release, error) {
	fake.updateReleaseMutex.Lock()
	fake.updateReleaseArgsForCall = append(fake.updateReleaseArgsForCall, struct {
		productSlug string
		release     go_pivnet.Release
	}{productSlug, release})
	fake.recordInvocation("UpdateRelease", []interface{}{productSlug, release})
	fake.updateReleaseMutex.Unlock()
	if fake.UpdateReleaseStub != nil {
		return fake.UpdateReleaseStub(productSlug, release)
	} else {
		return fake.updateReleaseReturns.result1, fake.updateReleaseReturns.result2
	}
}

func (fake *UserGroupsUpdaterClient) UpdateReleaseCallCount() int {
	fake.updateReleaseMutex.RLock()
	defer fake.updateReleaseMutex.RUnlock()
	return len(fake.updateReleaseArgsForCall)
}

func (fake *UserGroupsUpdaterClient) UpdateReleaseArgsForCall(i int) (string, go_pivnet.Release) {
	fake.updateReleaseMutex.RLock()
	defer fake.updateReleaseMutex.RUnlock()
	return fake.updateReleaseArgsForCall[i].productSlug, fake.updateReleaseArgsForCall[i].release
}

func (fake *UserGroupsUpdaterClient) UpdateReleaseReturns(result1 go_pivnet.Release, result2 error) {
	fake.UpdateReleaseStub = nil
	fake.updateReleaseReturns = struct {
		result1 go_pivnet.Release
		result2 error
	}{result1, result2}
}

func (fake *UserGroupsUpdaterClient) AddUserGroup(productSlug string, releaseID int, userGroupID int) error {
	fake.addUserGroupMutex.Lock()
	fake.addUserGroupArgsForCall = append(fake.addUserGroupArgsForCall, struct {
		productSlug string
		releaseID   int
		userGroupID int
	}{productSlug, releaseID, userGroupID})
	fake.recordInvocation("AddUserGroup", []interface{}{productSlug, releaseID, userGroupID})
	fake.addUserGroupMutex.Unlock()
	if fake.AddUserGroupStub != nil {
		return fake.AddUserGroupStub(productSlug, releaseID, userGroupID)
	} else {
		return fake.addUserGroupReturns.result1
	}
}

func (fake *UserGroupsUpdaterClient) AddUserGroupCallCount() int {
	fake.addUserGroupMutex.RLock()
	defer fake.addUserGroupMutex.RUnlock()
	return len(fake.addUserGroupArgsForCall)
}

func (fake *UserGroupsUpdaterClient) AddUserGroupArgsForCall(i int) (string, int, int) {
	fake.addUserGroupMutex.RLock()
	defer fake.addUserGroupMutex.RUnlock()
	return fake.addUserGroupArgsForCall[i].productSlug, fake.addUserGroupArgsForCall[i].releaseID, fake.addUserGroupArgsForCall[i].userGroupID
}

func (fake *UserGroupsUpdaterClient) AddUserGroupReturns(result1 error) {
	fake.AddUserGroupStub = nil
	fake.addUserGroupReturns = struct {
		result1 error
	}{result1}
}

func (fake *UserGroupsUpdaterClient) Invocations() map[string][][]interface{} {
	fake.invocationsMutex.RLock()
	defer fake.invocationsMutex.RUnlock()
	fake.updateReleaseMutex.RLock()
	defer fake.updateReleaseMutex.RUnlock()
	fake.addUserGroupMutex.RLock()
	defer fake.addUserGroupMutex.RUnlock()
	return fake.invocations
}

func (fake *UserGroupsUpdaterClient) recordInvocation(key string, args []interface{}) {
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
