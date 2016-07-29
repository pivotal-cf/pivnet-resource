package out_test

import (
	"errors"
	"io/ioutil"
	"log"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/pivnet-resource/concourse"
	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/out"
	"github.com/pivotal-cf-experimental/pivnet-resource/out/outfakes"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Out", func() {
	Describe("Run", func() {
		var (
			logger    *log.Logger
			finalizer *outfakes.Finalizer
			creator   *outfakes.Creator
			validator *outfakes.Validation
			uploader  *outfakes.Uploader
			globber   *outfakes.Globber
			cmd       out.OutCommand
		)

		BeforeEach(func() {
			logger = log.New(ioutil.Discard, "doesn't matter", 0)
			finalizer = &outfakes.Finalizer{}
			creator = &outfakes.Creator{}
			validator = &outfakes.Validation{}
			uploader = &outfakes.Uploader{}
			globber = &outfakes.Globber{}

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
				Logger:     logger,
				OutDir:     "some/out/dir",
				SourcesDir: "some/sources/dir",
				GlobClient: globber,
				Validation: validator,
				Creator:    creator,
				Finalizer:  finalizer,
				Uploader:   uploader,
				M:          meta,
			}

			cmd = out.NewOutCommand(config)

			globber.ExactGlobsReturns([]string{"some-glob-1", "some-glob-2"}, nil)
			creator.CreateReturns(pivnet.Release{ID: 1337, Availability: "none"}, nil)
			finalizer.FinalizeReturns(concourse.OutResponse{
				Version: concourse.Version{
					ProductVersion: "some-returned-product-version",
				},
			}, nil)
		})

		It("returns a concourse out response", func() {
			request := concourse.OutRequest{}
			response, err := cmd.Run(request)
			Expect(err).NotTo(HaveOccurred())

			Expect(response).To(Equal(concourse.OutResponse{
				Version: concourse.Version{
					ProductVersion: "some-returned-product-version",
				},
			}))

			Expect(globber.ExactGlobsCallCount()).To(Equal(1))

			Expect(creator.CreateCallCount()).To(Equal(1))

			pivnetRelease, exactGlobs := uploader.UploadArgsForCall(0)
			Expect(pivnetRelease).To(Equal(pivnet.Release{ID: 1337, Availability: "none"}))
			Expect(exactGlobs).To(Equal([]string{"some-glob-1", "some-glob-2"}))

			Expect(finalizer.FinalizeArgsForCall(0)).To(Equal(pivnet.Release{ID: 1337, Availability: "none"}))
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
					Expect(err).To(MatchError(errors.New(`product_files were provided in metadata that match no globs: [some-glob-1]`)))
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
