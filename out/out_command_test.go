package out_test

import (
	"errors"
	"github.com/pivotal-cf/go-pivnet/v7"
	"github.com/pivotal-cf/go-pivnet/v7/logger"
	"github.com/pivotal-cf/go-pivnet/v7/logshim"
	"github.com/pivotal-cf/pivnet-resource/v3/concourse"
	"github.com/pivotal-cf/pivnet-resource/v3/metadata"
	"github.com/pivotal-cf/pivnet-resource/v3/out"
	"github.com/pivotal-cf/pivnet-resource/v3/out/outfakes"
	"log"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
)

var _ = Describe("Out", func() {
	Describe("Run", func() {
		var (
			fakeLogger logger.Logger

			finalizer                      *outfakes.Finalizer
			userGroupsUpdater              *outfakes.UserGroupsUpdater
			releaseFileGroupsAdder         *outfakes.ReleaseFileGroupsAdder
			releaseArtifactReferencesAdder *outfakes.ReleaseArtifactReferencesAdder
			releaseDependenciesAdder       *outfakes.ReleaseDependenciesAdder
			dependencySpecifiersCreator    *outfakes.DependencySpecifiersCreator
			releaseUpgradePathsAdder       *outfakes.ReleaseUpgradePathsAdder
			upgradePathSpecifiersCreator   *outfakes.UpgradePathSpecifiersCreator
			creator                        *outfakes.Creator
			finder                         *outfakes.Finder
			validator                      *outfakes.Validation
			uploader                       *outfakes.Uploader
			globber                        *outfakes.Globber
			cmd                            out.OutCommand

			skipUpload bool
			request    concourse.OutRequest

			productSlug string

			returnedExactGlobs []string

			pivnetRelease pivnet.Release

			validateErr                     error
			createErr                       error
			findErr                         error
			exactGlobsErr                   error
			uploadErr                       error
			updateUserGroupErr              error
			addReleaseFileGroupsErr         error
			addReleaseArtifactReferencesErr error
			addReleaseDependenciesErr       error
			createDependencySpecifiersErr   error
			addReleaseUpgradePathsErr       error
			createUpgradePathSpecifiersErr  error
			finalizeErr                     error
		)

		BeforeEach(func() {
			logger := log.New(GinkgoWriter, "", log.LstdFlags)
			fakeLogger = logshim.NewLogShim(logger, logger, true)

			finalizer = &outfakes.Finalizer{}
			userGroupsUpdater = &outfakes.UserGroupsUpdater{}
			releaseFileGroupsAdder = &outfakes.ReleaseFileGroupsAdder{}
			releaseArtifactReferencesAdder = &outfakes.ReleaseArtifactReferencesAdder{}
			releaseDependenciesAdder = &outfakes.ReleaseDependenciesAdder{}
			dependencySpecifiersCreator = &outfakes.DependencySpecifiersCreator{}
			releaseUpgradePathsAdder = &outfakes.ReleaseUpgradePathsAdder{}
			upgradePathSpecifiersCreator = &outfakes.UpgradePathSpecifiersCreator{}
			creator = &outfakes.Creator{}
			finder = &outfakes.Finder{}
			validator = &outfakes.Validation{}
			uploader = &outfakes.Uploader{}
			globber = &outfakes.Globber{}

			skipUpload = false

			productSlug = "some-product-slug"

			returnedExactGlobs = []string{"some-glob-1", "some-glob-2"}

			validateErr = nil
			createErr = nil
			findErr = nil
			exactGlobsErr = nil
			uploadErr = nil
			updateUserGroupErr = nil
			addReleaseFileGroupsErr = nil
			addReleaseArtifactReferencesErr = nil
			addReleaseDependenciesErr = nil
			createDependencySpecifiersErr = nil
			addReleaseUpgradePathsErr = nil
			createUpgradePathSpecifiersErr = nil
			finalizeErr = nil
			pivnetRelease = pivnet.Release{}
		})

		Describe("Release creation", func() {
			JustBeforeEach(func() {
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
					Logger:                         fakeLogger,
					OutDir:                         "some/out/dir",
					SourcesDir:                     "some/sources/dir",
					GlobClient:                     globber,
					Validation:                     validator,
					Creator:                        creator,
					Finalizer:                      finalizer,
					UserGroupsUpdater:              userGroupsUpdater,
					ReleaseFileGroupsAdder:         releaseFileGroupsAdder,
					ReleaseArtifactReferencesAdder: releaseArtifactReferencesAdder,
					ReleaseDependenciesAdder:       releaseDependenciesAdder,
					DependencySpecifiersCreator:    dependencySpecifiersCreator,
					ReleaseUpgradePathsAdder:       releaseUpgradePathsAdder,
					UpgradePathSpecifiersCreator:   upgradePathSpecifiersCreator,
					Uploader:                       uploader,
					M:                              meta,
					SkipUpload:                     skipUpload,
					FilesOnly:                      false,
				}

				cmd = out.NewOutCommand(config)

				validator.ValidateReturns(validateErr)
				creator.CreateReturns(pivnet.Release{ID: 1337, Availability: "none", Version: "some-version"}, createErr)

				globber.ExactGlobsReturns(returnedExactGlobs, exactGlobsErr)

				userGroupsUpdater.UpdateUserGroupsReturns(pivnet.Release{ID: 1337, Availability: "none", Version: "some-version"}, updateUserGroupErr)

				uploader.UploadReturns(uploadErr)
				releaseFileGroupsAdder.AddReleaseFileGroupsReturns(addReleaseFileGroupsErr)
				releaseArtifactReferencesAdder.AddReleaseArtifactReferencesReturns(addReleaseArtifactReferencesErr)
				releaseDependenciesAdder.AddReleaseDependenciesReturns(addReleaseDependenciesErr)
				dependencySpecifiersCreator.CreateDependencySpecifiersReturns(createDependencySpecifiersErr)
				releaseUpgradePathsAdder.AddReleaseUpgradePathsReturns(addReleaseUpgradePathsErr)
				upgradePathSpecifiersCreator.CreateUpgradePathSpecifiersReturns(createUpgradePathSpecifiersErr)

				finalizer.FinalizeReturns(concourse.OutResponse{
					Version: concourse.Version{
						ProductVersion: "some-new-version",
					},
				}, finalizeErr)

				request = concourse.OutRequest{
					Source: concourse.Source{
						ProductSlug: productSlug,
					},
				}
			})

			It("returns a concourse out response", func() {
				response, err := cmd.Run(request)
				Expect(err).NotTo(HaveOccurred())

				Expect(response).To(Equal(concourse.OutResponse{
					Version: concourse.Version{
						ProductVersion: "some-new-version",
					},
				}))

				Expect(creator.CreateCallCount()).To(Equal(1))

				Expect(globber.ExactGlobsCallCount()).To(Equal(1))

				Expect(releaseFileGroupsAdder.AddReleaseFileGroupsCallCount()).To(Equal(1))
				Expect(releaseArtifactReferencesAdder.AddReleaseArtifactReferencesCallCount()).To(Equal(1))
				Expect(releaseDependenciesAdder.AddReleaseDependenciesCallCount()).To(Equal(1))
				Expect(dependencySpecifiersCreator.CreateDependencySpecifiersCallCount()).To(Equal(1))
				Expect(releaseUpgradePathsAdder.AddReleaseUpgradePathsCallCount()).To(Equal(1))
				Expect(upgradePathSpecifiersCreator.CreateUpgradePathSpecifiersCallCount()).To(Equal(1))

				Expect(uploader.UploadCallCount()).To(Equal(1))
				invokedPivnetRelease, invokedExactGlobs := uploader.UploadArgsForCall(0)
				Expect(invokedPivnetRelease).To(Equal(pivnet.Release{ID: 1337, Availability: "none", Version: "some-version"}))
				Expect(invokedExactGlobs).To(Equal([]string{"some-glob-1", "some-glob-2"}))

				Expect(userGroupsUpdater.UpdateUserGroupsCallCount()).To(Equal(1))
				invokedPivnetRelease = userGroupsUpdater.UpdateUserGroupsArgsForCall(0)
				Expect(invokedPivnetRelease).To(Equal(pivnet.Release{ID: 1337, Availability: "none", Version: "some-version"}))

				Expect(finalizer.FinalizeCallCount()).To(Equal(1))
				invokedProductSlug, invokedReleaseVersion := finalizer.FinalizeArgsForCall(0)
				Expect(invokedProductSlug).To(Equal(productSlug))
				Expect(invokedReleaseVersion).To(Equal("some-version"))
			})

			Context("when skipUpload is true", func() {
				BeforeEach(func() {
					skipUpload = true
				})

				It("does not invoke the uploader", func() {
					_, err := cmd.Run(request)
					Expect(err).NotTo(HaveOccurred())

					Expect(uploader.UploadCallCount()).To(Equal(0))
				})
			})

			Context("when outdir is not provided", func() {
				It("returns an error", func() {
					cmd := out.NewOutCommand(out.OutCommandConfig{SourcesDir: ""})

					_, err := cmd.Run(request)
					Expect(err).To(MatchError(errors.New("out dir must be provided")))
				})
			})

			Context("when the validation fails", func() {
				BeforeEach(func() {
					validateErr = errors.New("some validation error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err).To(Equal(validateErr))
				})
			})

			Context("when gathering the exact globs fails", func() {
				BeforeEach(func() {
					exactGlobsErr = errors.New("some exact globs error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err).To(Equal(exactGlobsErr))
				})
			})

			Context("when product files were provided that match no globs", func() {
				BeforeEach(func() {
					returnedExactGlobs = []string{"this-is-missing"}
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err.Error()).To(MatchRegexp(
						`product files .* match no globs: \[some-glob-1 some-glob-2\]`))
				})
			})

			Context("when a release cannot be created", func() {
				BeforeEach(func() {
					createErr = errors.New("some create error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err).To(Equal(createErr))
				})
			})

			Context("when a release cannot be uploaded", func() {
				BeforeEach(func() {
					uploadErr = errors.New("upload error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err).To(Equal(uploadErr))
				})
			})

			Context("when user groups cannot be updated", func() {
				BeforeEach(func() {
					updateUserGroupErr = errors.New("some user group error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err).To(Equal(updateUserGroupErr))
				})
			})

			Context("when dependencies cannot be added", func() {
				BeforeEach(func() {
					addReleaseDependenciesErr = errors.New("some release dependencies error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err).To(Equal(addReleaseDependenciesErr))
				})
			})

			Context("when creating dependency specifiers returns an error", func() {
				BeforeEach(func() {
					createDependencySpecifiersErr = errors.New("some dependency specifiers error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err).To(Equal(createDependencySpecifiersErr))
				})
			})

			Context("when upgrade paths cannot be added", func() {
				BeforeEach(func() {
					addReleaseUpgradePathsErr = errors.New("some release upgrade error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err).To(Equal(addReleaseUpgradePathsErr))
				})
			})

			Context("when creating upgrade path specifiers returns an error", func() {
				BeforeEach(func() {
					createUpgradePathSpecifiersErr = errors.New("some upgrade path specifiers error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err).To(Equal(createUpgradePathSpecifiersErr))
				})
			})

			Context("when a release cannot be finalized", func() {
				BeforeEach(func() {
					finalizeErr = errors.New("some finalize error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)
					Expect(err).To(Equal(finalizeErr))
				})
			})
		})

		Describe("Release files updating", func() {
			JustBeforeEach(func() {
				meta := metadata.Metadata{
					ExistingRelease: &metadata.ExistingRelease{
						ID: 123,
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
					Logger:                         fakeLogger,
					OutDir:                         "some/out/dir",
					SourcesDir:                     "some/sources/dir",
					GlobClient:                     globber,
					Validation:                     validator,
					Creator:                        creator,
					Finder:                         finder,
					Finalizer:                      finalizer,
					UserGroupsUpdater:              userGroupsUpdater,
					ReleaseFileGroupsAdder:         releaseFileGroupsAdder,
					ReleaseArtifactReferencesAdder: releaseArtifactReferencesAdder,
					ReleaseDependenciesAdder:       releaseDependenciesAdder,
					DependencySpecifiersCreator:    dependencySpecifiersCreator,
					ReleaseUpgradePathsAdder:       releaseUpgradePathsAdder,
					UpgradePathSpecifiersCreator:   upgradePathSpecifiersCreator,
					Uploader:                       uploader,
					M:                              meta,
					SkipUpload:                     skipUpload,
					FilesOnly:                      true,
				}

				cmd = out.NewOutCommand(config)

				validator.ValidateReturns(validateErr)
				finder.FindReturns(pivnetRelease, findErr)

				globber.ExactGlobsReturns(returnedExactGlobs, exactGlobsErr)

				userGroupsUpdater.UpdateUserGroupsReturns(pivnet.Release{ID: 1337, Availability: "none", Version: "some-version"}, updateUserGroupErr)

				uploader.UploadReturns(uploadErr)
				releaseFileGroupsAdder.AddReleaseFileGroupsReturns(addReleaseFileGroupsErr)
				releaseArtifactReferencesAdder.AddReleaseArtifactReferencesReturns(addReleaseArtifactReferencesErr)
				releaseDependenciesAdder.AddReleaseDependenciesReturns(addReleaseDependenciesErr)
				dependencySpecifiersCreator.CreateDependencySpecifiersReturns(createDependencySpecifiersErr)
				releaseUpgradePathsAdder.AddReleaseUpgradePathsReturns(addReleaseUpgradePathsErr)
				upgradePathSpecifiersCreator.CreateUpgradePathSpecifiersReturns(createUpgradePathSpecifiersErr)

				finalizer.FinalizeReturns(concourse.OutResponse{
					Version: concourse.Version{
						ProductVersion: "some-new-version",
					},
				}, finalizeErr)

				request = concourse.OutRequest{
					Source: concourse.Source{
						ProductSlug: productSlug,
					},
				}
			})

			Context("finder gets a release", func() {
				BeforeEach(func() {
					pivnetRelease = pivnet.Release{ID: 123, Version: "existing-product-version", SoftwareFilesUpdatedAt: "2021-01-01"}
				})

				It("uploads files to a release", func() {
					response, err := cmd.Run(request)
					Expect(err).NotTo(HaveOccurred())

					Expect(response).To(Equal(concourse.OutResponse{
						Version: concourse.Version{
							ProductVersion: "some-new-version",
						},
					}))

					Expect(uploader.UploadCallCount()).To(Equal(1))
					release, globs := uploader.UploadArgsForCall(0)
					Expect(release).To(Equal(pivnet.Release{ID: 123, Version: "existing-product-version", SoftwareFilesUpdatedAt: "2021-01-01"}))
					Expect(globs).To(Equal([]string{"some-glob-1", "some-glob-2"}))

					Expect(creator.CreateCallCount()).To(Equal(0))
					Expect(userGroupsUpdater.UpdateUserGroupsCallCount()).To(Equal(0))
					Expect(releaseFileGroupsAdder.AddReleaseFileGroupsCallCount()).To(Equal(0))
					Expect(releaseArtifactReferencesAdder.AddReleaseArtifactReferencesCallCount()).To(Equal(0))
					Expect(releaseDependenciesAdder.AddReleaseDependenciesCallCount()).To(Equal(0))
					Expect(dependencySpecifiersCreator.CreateDependencySpecifiersCallCount()).To(Equal(0))
					Expect(releaseUpgradePathsAdder.AddReleaseUpgradePathsCallCount()).To(Equal(0))
					Expect(upgradePathSpecifiersCreator.CreateUpgradePathSpecifiersCallCount()).To(Equal(0))
					Expect(finalizer.FinalizeCallCount()).To(Equal(1))
				})
			})

			Context("finder cannot find release", func() {
				BeforeEach(func() {
					findErr = errors.New("some find error")
				})

				It("returns an error", func() {
					_, err := cmd.Run(request)

					Expect(err).To(HaveOccurred())
				})
			})
		})
	})
})
