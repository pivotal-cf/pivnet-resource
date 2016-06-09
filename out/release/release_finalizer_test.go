package release_test

import (
	"errors"

	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/release/releasefakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("ReleaseFinalizer", func() {
	Describe("Finalize", func() {
		var (
			pivnetClient  *releasefakes.UpdateClient
			fetcherClient *releasefakes.Fetcher
			params        concourse.OutParams
			pivnetRelease pivnet.Release
			finalizer     release.ReleaseFinalizer
		)

		BeforeEach(func() {
			pivnetClient = &releasefakes.UpdateClient{}
			fetcherClient = &releasefakes.Fetcher{}

			params = concourse.OutParams{
				AvailabilityFile: "/some/availability/file",
				UserGroupIDsFile: "/some/usergroupids/file",
			}

			pivnetRelease = pivnet.Release{
				ID:      1337,
				Version: "some-version",
				EULA: &pivnet.EULA{
					Slug: "a_eula_slug",
				},
			}

			finalizer = release.NewFinalizer(pivnetClient, fetcherClient, params, "/some/sources/dir", "a-product-slug")
		})

		Context("when the release availability is Admins Only", func() {
			BeforeEach(func() {
				fetcherClient.FetchReturns("Admins Only")
				pivnetClient.ReleaseETagReturns("some-etag", nil)
			})

			It("returns a final concourse out response", func() {
				response, err := finalizer.Finalize(pivnetRelease)
				Expect(err).NotTo(HaveOccurred())

				Expect(pivnetClient.UpdateReleaseCallCount()).To(BeZero())
				Expect(pivnetClient.AddUserGroupCallCount()).To(BeZero())

				valueToFetch, sourcesDir, availabilityFile := fetcherClient.FetchArgsForCall(0)
				Expect(valueToFetch).To(Equal("Availability"))
				Expect(sourcesDir).To(Equal("/some/sources/dir"))
				Expect(availabilityFile).To(Equal("/some/availability/file"))

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
						{Name: "availability", Value: ""},
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
				fetcherClient.FetchReturns("Selected User Groups Only")
				pivnetClient.UpdateReleaseReturns(pivnet.Release{Version: "another-version", EULA: &pivnet.EULA{Slug: "eula_slug"}}, nil)
				pivnetClient.ReleaseETagReturns("a-sep-etag", nil)
			})

			It("returns a final concourse out response", func() {
			})
		})

		Context("when the release availability is any other value", func() {
			BeforeEach(func() {
				fetcherClient.FetchReturns("some other group")
				pivnetClient.UpdateReleaseReturns(pivnet.Release{Version: "a-diff-version", EULA: &pivnet.EULA{Slug: "eula_slug"}}, nil)
				pivnetClient.ReleaseETagReturns("a-diff-etag", nil)
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
						fetcherClient.FetchReturns("some other group")
						pivnetClient.UpdateReleaseReturns(pivnet.Release{}, errors.New("there was a problem updating the release"))
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
