package pivnet_test

import (
	"errors"
	"fmt"
	"net/http"
	"time"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	logger_fakes "github.com/pivotal-cf-experimental/pivnet-resource/logger/fakes"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

var _ = Describe("PivnetClient - product files", func() {
	var (
		server     *ghttp.Server
		client     pivnet.Client
		token      string
		apiAddress string
		userAgent  string

		newClientConfig pivnet.NewClientConfig
		fakeLogger      logger.Logger
	)

	BeforeEach(func() {
		server = ghttp.NewServer()
		apiAddress = server.URL()
		token = "my-auth-token"
		userAgent = "pivnet-resource/0.1.0 (some-url)"

		fakeLogger = &logger_fakes.FakeLogger{}
		newClientConfig = pivnet.NewClientConfig{
			Endpoint:  apiAddress,
			Token:     token,
			UserAgent: userAgent,
		}
		client = pivnet.NewClient(newClientConfig, fakeLogger)
	})

	AfterEach(func() {
		server.Close()
	})

	Describe("ReleasesForProductSlug", func() {
		It("returns the releases for the product slug", func() {
			response := `{"releases": [{"id":2,"version":"1.2.3"},{"id": 3, "version": "3.2.1", "_links": {"product_files": {"href":"https://banana.org/cookies/download"}}}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases"),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			releases, err := client.ReleasesForProductSlug("banana")
			Expect(err).NotTo(HaveOccurred())
			Expect(releases).To(HaveLen(2))
			Expect(releases[0].ID).To(Equal(2))
			Expect(releases[1].ID).To(Equal(3))
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases"),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.ReleasesForProductSlug("banana")
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("GetRelease", func() {
		It("returns the release based on the name and version", func() {
			response := `{"releases": [{"id": 3, "version": "3.2.1", "_links": {"product_files": {"href":"https://banana.org/cookies/download"}}}]}`

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases"),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			release, err := client.GetRelease("banana", "3.2.1")
			Expect(err).NotTo(HaveOccurred())
			Expect(release.Links.ProductFiles["href"]).To(Equal("https://banana.org/cookies/download"))
		})

		Context("when the requested version is not available but the request is successful", func() {
			It("returns an error", func() {
				response := `{"releases": [{"id": 3, "version": "3.2.1"}]}`

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases"),
						ghttp.RespondWith(http.StatusOK, response),
					),
				)

				_, err := client.GetRelease("banana", "1.0.0")
				Expect(err).To(MatchError(errors.New("The requested version: 1.0.0 - could not be found")))
			})
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases"),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.GetRelease("banana", "1.0.0")
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("Create Release", func() {
		var (
			productVersion      string
			createReleaseConfig pivnet.CreateReleaseConfig
		)

		BeforeEach(func() {
			productVersion = "1.2.3.4"

			createReleaseConfig = pivnet.CreateReleaseConfig{
				EulaSlug:       "some_eula",
				ReleaseType:    "Not a real release",
				ProductVersion: productVersion,
				ProductSlug:    productSlug,
			}
		})

		Context("when the config is valid", func() {
			type requestBody struct {
				Release pivnet.Release `json:"release"`
			}

			var (
				expectedReleaseDate string
				expectedRequestBody requestBody

				validResponse string
			)

			BeforeEach(func() {
				expectedReleaseDate = time.Now().Format("2006-01-02")

				expectedRequestBody = requestBody{
					Release: pivnet.Release{
						Availability: "Admins Only",
						OSSCompliant: "confirm",
						ReleaseDate:  expectedReleaseDate,
						ReleaseType:  createReleaseConfig.ReleaseType,
						Eula: &pivnet.Eula{
							Slug: createReleaseConfig.EulaSlug,
						},
						Version: createReleaseConfig.ProductVersion,
					},
				}

				validResponse = `{"release": {"id": 3, "version": "1.2.3.4"}}`
			})

			It("creates the release with the minimum required fields", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
						ghttp.VerifyJSONRepresenting(&expectedRequestBody),
						ghttp.RespondWith(http.StatusCreated, validResponse),
					),
				)

				release, err := client.CreateRelease(createReleaseConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(release.Version).To(Equal(productVersion))
			})

			Context("when the optional release date is present", func() {
				var (
					releaseDate string
				)

				BeforeEach(func() {
					releaseDate = "2015-12-24"

					createReleaseConfig.ReleaseDate = releaseDate
					expectedRequestBody.Release.ReleaseDate = releaseDate
				})

				It("creates the release with the release date field", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
							ghttp.VerifyJSONRepresenting(&expectedRequestBody),
							ghttp.RespondWith(http.StatusCreated, validResponse),
						),
					)

					release, err := client.CreateRelease(createReleaseConfig)
					Expect(err).NotTo(HaveOccurred())
					Expect(release.Version).To(Equal(productVersion))
				})
			})

			Describe("optional description field", func() {
				var (
					description string
				)

				Context("when the optional description field is present", func() {
					BeforeEach(func() {
						description = "some description"

						createReleaseConfig.Description = description
						expectedRequestBody.Release.Description = description
					})

					It("creates the release with the description field", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
								ghttp.VerifyJSONRepresenting(&expectedRequestBody),
								ghttp.RespondWith(http.StatusCreated, validResponse),
							),
						)

						release, err := client.CreateRelease(createReleaseConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(release.Version).To(Equal(productVersion))
					})
				})

				Context("when the optional description field is not present", func() {
					BeforeEach(func() {
						description = ""

						createReleaseConfig.Description = description
						expectedRequestBody.Release.Description = description
					})

					It("creates the release with an empty description field", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
								ghttp.VerifyJSONRepresenting(&expectedRequestBody),
								ghttp.RespondWith(http.StatusCreated, validResponse),
							),
						)

						release, err := client.CreateRelease(createReleaseConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(release.Version).To(Equal(productVersion))
					})
				})
			})

			Describe("optional release notes URL field", func() {
				var (
					releaseNotesURL string
				)

				Context("when the optional release notes URL field is present", func() {
					BeforeEach(func() {
						releaseNotesURL = "some releaseNotesURL"

						createReleaseConfig.ReleaseNotesURL = releaseNotesURL
						expectedRequestBody.Release.ReleaseNotesURL = releaseNotesURL
					})

					It("creates the release with the release notes URL field", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
								ghttp.VerifyJSONRepresenting(&expectedRequestBody),
								ghttp.RespondWith(http.StatusCreated, validResponse),
							),
						)

						release, err := client.CreateRelease(createReleaseConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(release.Version).To(Equal(productVersion))
					})
				})

				Context("when the optional release notes URL field is not present", func() {
					BeforeEach(func() {
						releaseNotesURL = ""

						createReleaseConfig.ReleaseNotesURL = releaseNotesURL
						expectedRequestBody.Release.ReleaseNotesURL = releaseNotesURL
					})

					It("creates the release with an empty release notes URL field", func() {
						server.AppendHandlers(
							ghttp.CombineHandlers(
								ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
								ghttp.VerifyJSONRepresenting(&expectedRequestBody),
								ghttp.RespondWith(http.StatusCreated, validResponse),
							),
						)

						release, err := client.CreateRelease(createReleaseConfig)
						Expect(err).NotTo(HaveOccurred())
						Expect(release.Version).To(Equal(productVersion))
					})
				})
			})
		})

		Context("when the server responds with a non-201 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/releases"),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.CreateRelease(createReleaseConfig)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 201")))
			})
		})
	})

	Describe("UpdateRelease", func() {
		It("submits the updated values for a release with OSS compliance", func() {
			release := pivnet.Release{
				ID:      42,
				Version: "1.2.3.4",
				Eula: &pivnet.Eula{
					Slug: "some-eula",
					ID:   15,
				},
			}

			patchURL := fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, "banana-slug", release.ID)

			response := `{"release": {"id": 42, "version": "1.2.3.4"}}`
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("PATCH", patchURL),
					ghttp.VerifyJSON(`{"release":{"id": 42, "version": "1.2.3.4", "eula":{"slug":"some-eula","id":15}, "oss_compliant":"confirm"}}`),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			release, err := client.UpdateRelease("banana-slug", release)
			Expect(err).NotTo(HaveOccurred())
			Expect(release.Version).To(Equal("1.2.3.4"))
		})

		Context("when the server responds with a non-200 status code", func() {
			It("returns the error", func() {
				release := pivnet.Release{ID: 111}
				patchURL := fmt.Sprintf("%s/products/%s/releases/%d", apiPrefix, "banana-slug", release.ID)

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", patchURL),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.UpdateRelease("banana-slug", release)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("DeleteRelease", func() {
		var (
			release pivnet.Release
		)

		BeforeEach(func() {
			release = pivnet.Release{
				ID: 1234,
			}
		})

		It("deletes the release", func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/products/banana/releases/%d", apiPrefix, release.ID)),
					ghttp.RespondWith(http.StatusNoContent, nil),
				),
			)

			err := client.DeleteRelease(release, "banana")
			Expect(err).NotTo(HaveOccurred())
		})

		Context("when the server responds with a non-204 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("DELETE", fmt.Sprintf("%s/products/banana/releases/%d", apiPrefix, release.ID)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				err := client.DeleteRelease(release, "banana")
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 204")))
			})
		})
	})
})
