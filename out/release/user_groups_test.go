package release_test

import (
	"errors"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("UserGroupsUpdater", func() {
	Describe("UpdateUserGroups", func() {
		var (
			pivnetClient *releasefakes.UserGroupsUpdaterClient

			mdata metadata.Metadata

			productSlug   string
			pivnetRelease pivnet.Release

			userGroupsUpdater release.UserGroupsUpdater
		)

		BeforeEach(func() {
			pivnetClient = &releasefakes.UserGroupsUpdaterClient{}

			productSlug = "some-product-slug"

			pivnetRelease = pivnet.Release{
				Availability: "some-value",
				ID:           1337,
				Version:      "some-version",
				EULA: &pivnet.EULA{
					Slug: "a_eula_slug",
				},
			}

			mdata = metadata.Metadata{
				Release: &metadata.Release{
					Availability: "some-value",
					Version:      "some-version",
					EULASlug:     "a_eula_slug",
				},
				ProductFiles: []metadata.ProductFile{},
			}

			pivnetClient.UpdateReleaseReturns(pivnet.Release{Version: "a-diff-version", EULA: &pivnet.EULA{Slug: "eula_slug"}}, nil)
		})

		JustBeforeEach(func() {
			userGroupsUpdater = release.NewUserGroupsUpdater(
				pivnetClient,
				mdata,
				productSlug,
			)
		})

		It("updates availability", func() {
			_, err := userGroupsUpdater.UpdateUserGroups(pivnetRelease)
			Expect(err).NotTo(HaveOccurred())

			Expect(pivnetClient.UpdateReleaseCallCount()).To(Equal(1))

			invokedProductSlug, invokedReleaseUpdate := pivnetClient.UpdateReleaseArgsForCall(0)
			Expect(invokedProductSlug).To(Equal(productSlug))
			Expect(invokedReleaseUpdate).To(Equal(pivnet.Release{ID: pivnetRelease.ID, Availability: pivnetRelease.Availability}))
		})

		Context("when the release availability is Admins Only", func() {
			BeforeEach(func() {
				mdata.Release.Availability = "Admins Only"
			})

			It("does not update release or user groups", func() {
				_, err := userGroupsUpdater.UpdateUserGroups(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.UpdateReleaseCallCount()).To(BeZero())
				Expect(pivnetClient.AddUserGroupCallCount()).To(BeZero())
			})
		})

		Context("when the release availability is Selected User Groups Only", func() {
			BeforeEach(func() {
				mdata.Release.Availability = "Selected User Groups Only"
				mdata.Release.UserGroupIDs = []string{"111", "222"}

				pivnetClient.UpdateReleaseReturns(pivnet.Release{ID: 2001, Version: "another-version", EULA: &pivnet.EULA{Slug: "eula_slug"}}, nil)
			})

			It("returns a final concourse out response", func() {
				response, err := userGroupsUpdater.UpdateUserGroups(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddUserGroupCallCount()).To(Equal(2))

				slug, releaseID, userGroupID := pivnetClient.AddUserGroupArgsForCall(0)
				Expect(slug).To(Equal(productSlug))
				Expect(releaseID).To(Equal(2001))
				Expect(userGroupID).To(Equal(111))

				slug, releaseID, userGroupID = pivnetClient.AddUserGroupArgsForCall(1)
				Expect(slug).To(Equal(productSlug))
				Expect(releaseID).To(Equal(2001))
				Expect(userGroupID).To(Equal(222))

				Expect(response.Version).To(Equal("another-version"))
			})

			Context("when an error occurs", func() {
				Context("when a user group ID cannpt be converted to a number", func() {
					BeforeEach(func() {
						mdata.Release.UserGroupIDs = []string{"&&&"}
					})

					It("returns an error", func() {
						_, err := userGroupsUpdater.UpdateUserGroups(pivnetRelease)
						Expect(err).To(MatchError(ContainSubstring(`parsing "&&&": invalid syntax`)))
					})
				})

				Context("when adding a user group to pivnet fails", func() {
					BeforeEach(func() {
						pivnetClient.AddUserGroupReturns(errors.New("failed to add user group"))
					})

					It("returns an error", func() {
						_, err := userGroupsUpdater.UpdateUserGroups(pivnetRelease)
						Expect(err).To(MatchError(errors.New("failed to add user group")))
					})
				})
			})
		})
	})
})
