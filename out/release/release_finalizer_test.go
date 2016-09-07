package release_test

import (
	"errors"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out/release"
	"github.com/pivotal-cf/pivnet-resource/out/release/releasefakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseFinalizer", func() {
	Describe("Finalize", func() {
		var (
			pivnetClient  *releasefakes.UpdateClient
			params        concourse.OutParams
			pivnetRelease pivnet.Release
			finalizer     release.ReleaseFinalizer
		)

		BeforeEach(func() {
			pivnetClient = &releasefakes.UpdateClient{}

			params = concourse.OutParams{}

			pivnetRelease = pivnet.Release{
				Availability: "Admins Only",
				ID:           1337,
				Version:      "some-version",
				EULA: &pivnet.EULA{
					Slug: "a_eula_slug",
				},
			}

			meta := metadata.Metadata{
				Release: &metadata.Release{
					Availability: "Admins Only",
					Version:      "some-version",
					EULASlug:     "a_eula_slug",
				},
				ProductFiles: []metadata.ProductFile{},
			}

			finalizer = release.NewFinalizer(pivnetClient, params, meta, "/some/sources/dir", "a-product-slug")
		})

		Context("when the release availability is Admins Only", func() {
			BeforeEach(func() {
				pivnetClient.ReleaseETagReturns("some-etag", nil)
			})

			It("returns a final concourse out response", func() {
				response, err := finalizer.Finalize(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.UpdateReleaseCallCount()).To(BeZero())
				Expect(pivnetClient.AddUserGroupCallCount()).To(BeZero())

				Expect(response).To(Equal(concourse.OutResponse{
					Version: concourse.Version{
						ProductVersion: "some-version#some-etag",
					},
					Metadata: []concourse.Metadata{
						{Name: "version", Value: "some-version"},
						{Name: "release_type", Value: ""},
						{Name: "release_date", Value: ""},
						{Name: "description", Value: ""},
						{Name: "release_notes_url", Value: ""},
						{Name: "availability", Value: "Admins Only"},
						{Name: "controlled", Value: "false"},
						{Name: "eccn", Value: ""},
						{Name: "license_exception", Value: ""},
						{Name: "end_of_support_date", Value: ""},
						{Name: "end_of_guidance_date", Value: ""},
						{Name: "end_of_availability_date", Value: ""},
						{Name: "eula_slug", Value: "a_eula_slug"},
					},
				}))
			})

			Context("when an error occurs", func() {
				Context("when the release ETag cannot be created", func() {
					BeforeEach(func() {
						pivnetClient.ReleaseETagReturns("", errors.New("some etag error"))
					})

					It("returns an error", func() {
						_, err := finalizer.Finalize(pivnetRelease)
						Expect(err).To(MatchError(errors.New("some etag error")))
					})
				})
			})
		})

		Context("when the release availability is Selected User Groups Only", func() {
			BeforeEach(func() {
				meta := metadata.Metadata{
					Release: &metadata.Release{
						Availability: "Selected User Groups Only",
						EULASlug:     "eula_slug",
						UserGroupIDs: []string{"111", "222"},
					},
				}

				pivnetClient.UpdateReleaseReturns(pivnet.Release{ID: 2001, Version: "another-version", EULA: &pivnet.EULA{Slug: "eula_slug"}}, nil)
				pivnetClient.ReleaseETagReturns("a-sep-etag", nil)

				finalizer = release.NewFinalizer(pivnetClient, params, meta, "/some/sources/dir", "a-product-slug")
			})

			It("returns a final concourse out response", func() {
				response, err := finalizer.Finalize(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				slug, releaseID, userGroupID := pivnetClient.AddUserGroupArgsForCall(0)
				Expect(slug).To(Equal("a-product-slug"))
				Expect(releaseID).To(Equal(2001))
				Expect(userGroupID).To(Equal(111))

				slug, releaseID, userGroupID = pivnetClient.AddUserGroupArgsForCall(1)
				Expect(slug).To(Equal("a-product-slug"))
				Expect(releaseID).To(Equal(2001))
				Expect(userGroupID).To(Equal(222))

				Expect(response.Version).To(Equal(concourse.Version{
					ProductVersion: "another-version#a-sep-etag",
				}))
			})

			Context("when an error occurs", func() {
				Context("when a user group ID cannpt be converted to a number", func() {
					BeforeEach(func() {
						meta := metadata.Metadata{
							Release: &metadata.Release{
								Availability: "Selected User Groups Only",
								EULASlug:     "eula_slug",
								UserGroupIDs: []string{"&&&"},
							},
						}

						finalizer = release.NewFinalizer(pivnetClient, params, meta, "/some/sources/dir", "a-product-slug")
					})

					It("returns an error", func() {
						_, err := finalizer.Finalize(pivnetRelease)
						Expect(err).To(MatchError(ContainSubstring(`parsing "&&&": invalid syntax`)))
					})
				})

				Context("when adding a user group to pivnet fails", func() {
					BeforeEach(func() {
						pivnetClient.AddUserGroupReturns(errors.New("failed to add user group"))
					})

					It("returns an error", func() {
						_, err := finalizer.Finalize(pivnetRelease)
						Expect(err).To(MatchError(errors.New("failed to add user group")))
					})
				})
			})
		})

		Context("when the release availability is any other value", func() {
			BeforeEach(func() {
				meta := metadata.Metadata{
					Release: &metadata.Release{
						Availability: "some other group",
						EULASlug:     "eula_slug",
						UserGroupIDs: []string{"111", "222"},
					},
				}

				pivnetClient.UpdateReleaseReturns(pivnet.Release{Version: "a-diff-version", EULA: &pivnet.EULA{Slug: "eula_slug"}}, nil)
				pivnetClient.ReleaseETagReturns("a-diff-etag", nil)

				finalizer = release.NewFinalizer(pivnetClient, params, meta, "/some/sources/dir", "a-product-slug")
			})

			It("returns a final concourse out response", func() {
				response, err := finalizer.Finalize(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.AddUserGroupCallCount()).To(BeZero())

				productSlug, releaseUpdate := pivnetClient.UpdateReleaseArgsForCall(0)
				Expect(productSlug).To(Equal("a-product-slug"))
				Expect(releaseUpdate).To(Equal(pivnet.Release{ID: 1337, Availability: "some other group"}))

				Expect(response.Version).To(Equal(concourse.Version{
					ProductVersion: "a-diff-version#a-diff-etag",
				}))

				Expect(response.Metadata).To(ContainElement(concourse.Metadata{Name: "version", Value: "a-diff-version"}))
				Expect(response.Metadata).To(ContainElement(concourse.Metadata{Name: "controlled", Value: "false"}))
				Expect(response.Metadata).To(ContainElement(concourse.Metadata{Name: "eula_slug", Value: "eula_slug"}))
			})

			Context("when an errors occurs", func() {
				Context("updating the release fails", func() {
					BeforeEach(func() {
						meta := metadata.Metadata{
							Release: &metadata.Release{
								Availability: "some other group",
								EULASlug:     "eula_slug",
								UserGroupIDs: []string{"111", "222"},
							},
						}

						pivnetClient.UpdateReleaseReturns(pivnet.Release{}, errors.New("there was a problem updating the release"))

						finalizer = release.NewFinalizer(pivnetClient, params, meta, "/some/sources/dir", "a-product-slug")
					})

					It("returns an error", func() {
						_, err := finalizer.Finalize(pivnetRelease)
						Expect(err).To(MatchError(errors.New("there was a problem updating the release")))
					})
				})
			})
		})
	})
})
