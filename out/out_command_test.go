package out_test

import (
	"errors"
	"io/ioutil"
	"log"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/concourse"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	"github.com/pivotal-cf/pivnet-resource/out"
	"github.com/pivotal-cf/pivnet-resource/out/outfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Out", func() {
	Describe("Run", func() {
		var (
			logger                   *log.Logger
			finalizer                *outfakes.Finalizer
			userGroupsUpdater        *outfakes.UserGroupsUpdater
			releaseDependenciesAdder *outfakes.ReleaseDependenciesAdder
			creator                  *outfakes.Creator
			validator                *outfakes.Validation
			uploader                 *outfakes.Uploader
			globber                  *outfakes.Globber
			cmd                      out.OutCommand

			productSlug string
		)

		BeforeEach(func() {
			logger = log.New(ioutil.Discard, "doesn't matter", 0)
			finalizer = &outfakes.Finalizer{}
			userGroupsUpdater = &outfakes.UserGroupsUpdater{}
			releaseDependenciesAdder = &outfakes.ReleaseDependenciesAdder{}
			creator = &outfakes.Creator{}
			validator = &outfakes.Validation{}
			uploader = &outfakes.Uploader{}
			globber = &outfakes.Globber{}

			productSlug = "some-product-slug"

			meta := metadata.Metadata{
				Release: &metadata.Release{
					Version: "release-version",
				},
				ProductFiles: []metadata.ProductFile{
					{
						File: "some-glob-1",
					},
					{
						File: "some-glob-2",
					},
				},
			}

			config := out.OutCommandConfig{
				Logger:                   logger,
				OutDir:                   "some/out/dir",
				SourcesDir:               "some/sources/dir",
				GlobClient:               globber,
				Validation:               validator,
				Creator:                  creator,
				Finalizer:                finalizer,
				UserGroupsUpdater:        userGroupsUpdater,
				ReleaseDependenciesAdder: releaseDependenciesAdder,
				Uploader:                 uploader,
				M:                        meta,
			}

			cmd = out.NewOutCommand(config)

			creator.CreateReturns(pivnet.Release{ID: 1337, Availability: "none"}, nil)

			globber.ExactGlobsReturns([]string{"some-glob-1", "some-glob-2"}, nil)

			userGroupsUpdater.UpdateUserGroupsReturns(pivnet.Release{ID: 1337, Availability: "none"}, nil)

			releaseDependenciesAdder.AddReleaseDependenciesReturns(nil)

			finalizer.FinalizeReturns(concourse.OutResponse{
				Version: concourse.Version{
					ProductVersion: "some-returned-product-version",
				},
			}, nil)
		})

		It("returns a concourse out response", func() {
			request := concourse.OutRequest{
				Source: concourse.Source{
					ProductSlug: productSlug,
				},
			}

			response, err := cmd.Run(request)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(Equal(concourse.OutResponse{
				Version: concourse.Version{
					ProductVersion: "some-returned-product-version",
				},
			}))

			Expect(creator.CreateCallCount()).To(Equal(1))

			Expect(globber.ExactGlobsCallCount()).To(Equal(1))

			Expect(userGroupsUpdater.UpdateUserGroupsCallCount()).To(Equal(1))

			Expect(releaseDependenciesAdder.AddReleaseDependenciesCallCount()).To(Equal(1))

			Expect(finalizer.FinalizeCallCount()).To(Equal(1))

			pivnetRelease, exactGlobs := uploader.UploadArgsForCall(0)
			Expect(pivnetRelease).To(Equal(pivnet.Release{ID: 1337, Availability: "none"}))
			Expect(exactGlobs).To(Equal([]string{"some-glob-1", "some-glob-2"}))

			pivnetRelease = userGroupsUpdater.UpdateUserGroupsArgsForCall(0)
			Expect(pivnetRelease).To(Equal(pivnet.Release{ID: 1337, Availability: "none"}))

			invokedRelease := finalizer.FinalizeArgsForCall(0)
			Expect(invokedRelease).To(Equal(pivnet.Release{ID: 1337, Availability: "none"}))
		})

		Context("when an error occurs", func() {
			Context("when outdir is not provided", func() {
				It("returns an error", func() {
					cmd := out.NewOutCommand(out.OutCommandConfig{SourcesDir: ""})
					request := concourse.OutRequest{}

					_, err := cmd.Run(request)
					Expect(err).To(MatchError(errors.New("out dir must be provided")))
				})
			})

			Context("when the validation fails", func() {
				BeforeEach(func() {
					validator.ValidateReturns(errors.New("some validation error"))
				})

				It("returns an error", func() {
					request := concourse.OutRequest{}

					_, err := cmd.Run(request)
					Expect(err).To(MatchError(errors.New("some validation error")))
				})
			})

			Context("when gathering the exact globs fails", func() {
				BeforeEach(func() {
					globber.ExactGlobsReturns([]string{}, errors.New("some exact globs error"))
				})

				It("returns an error", func() {
					request := concourse.OutRequest{}

					_, err := cmd.Run(request)
					Expect(err).To(MatchError(errors.New("some exact globs error")))
				})
			})

			Context("when product files were provided that match no globs", func() {
				BeforeEach(func() {
					meta := metadata.Metadata{
						Release: &metadata.Release{
							Version: "release-version",
						},
						ProductFiles: []metadata.ProductFile{
							{
								File: "some-glob-1",
							},
						},
					}

					config := out.OutCommandConfig{
						Logger:     logger,
						OutDir:     "some/out/dir",
						SourcesDir: "some/sources/dir",
						GlobClient: globber,
						Validation: validator,
						Creator:    nil,
						Finalizer:  nil,
						Uploader:   nil,
						M:          meta,
					}

					cmd = out.NewOutCommand(config)

					globber.ExactGlobsReturns([]string{"this-is-missing"}, nil)
				})

				It("returns an error", func() {
					request := concourse.OutRequest{}

					_, err := cmd.Run(request)
					Expect(err).To(MatchError(errors.New(`product files were provided in metadata that match no globs: [some-glob-1]`)))
				})
			})

			Context("when a release cannot be created", func() {
				BeforeEach(func() {
					creator.CreateReturns(pivnet.Release{}, errors.New("some create error"))
				})

				It("returns an error", func() {
					request := concourse.OutRequest{}

					_, err := cmd.Run(request)
					Expect(err).To(MatchError(errors.New("some create error")))
				})
			})

			Context("when a release cannot be uploaded", func() {
				BeforeEach(func() {
					uploader.UploadReturns(errors.New("some upload error"))
				})

				It("returns an error", func() {
					request := concourse.OutRequest{}

					_, err := cmd.Run(request)
					Expect(err).To(MatchError(errors.New("some upload error")))
				})
			})

			Context("when user groups cannot be updated", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = errors.New("some user group error")
					userGroupsUpdater.UpdateUserGroupsReturns(pivnet.Release{}, expectedErr)
				})

				It("returns an error", func() {
					request := concourse.OutRequest{}

					_, err := cmd.Run(request)
					Expect(err).To(Equal(expectedErr))
				})
			})

			Context("when dependencies cannot be added", func() {
				var (
					expectedErr error
				)

				BeforeEach(func() {
					expectedErr = errors.New("some release dependency error")
					releaseDependenciesAdder.AddReleaseDependenciesReturns(expectedErr)
				})

				It("returns an error", func() {
					request := concourse.OutRequest{}

					_, err := cmd.Run(request)
					Expect(err).To(Equal(expectedErr))
				})
			})

			Context("when a release cannot be finalized", func() {
				BeforeEach(func() {
					finalizer.FinalizeReturns(concourse.OutResponse{}, errors.New("some finalize error"))
				})

				It("returns an error", func() {
					request := concourse.OutRequest{}

					_, err := cmd.Run(request)
					Expect(err).To(MatchError(errors.New("some finalize error")))
				})
			})
		})
	})
})
