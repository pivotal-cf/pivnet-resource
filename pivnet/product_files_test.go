package pivnet_test

import (
	"errors"
	"fmt"
	"net/http"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
	"github.com/pivotal-cf-experimental/pivnet-resource/logger/loggerfakes"
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

		fakeLogger = &loggerfakes.FakeLogger{}
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

	Describe("Get Product Files", func() {
		var (
			release            pivnet.Release
			response           pivnet.ProductFiles
			responseStatusCode int
		)

		BeforeEach(func() {
			release = pivnet.Release{
				Links: &pivnet.Links{
					ProductFiles: map[string]string{"href": apiAddress + apiPrefix + "/products/banana/releases/666/product_files"},
				},
			}

			response = pivnet.ProductFiles{[]pivnet.ProductFile{
				{ID: 3, AWSObjectKey: "anything", Links: &pivnet.Links{Download: map[string]string{"href": "/products/banana/releases/666/product_files/6/download"}}},
				{ID: 4, AWSObjectKey: "something", Links: &pivnet.Links{Download: map[string]string{"href": "/products/banana/releases/666/product_files/8/download"}}},
			}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases/666/product_files"),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		Context("when links are nil", func() {
			BeforeEach(func() {
				release.Links = nil
			})

			It("returns error", func() {
				_, err := client.GetProductFiles(release)
				Expect(err).To(HaveOccurred())

				Expect(err).To(MatchError(errors.New("No links found")))
			})
		})

		It("returns the product files for the given release", func() {
			product, err := client.GetProductFiles(release)
			Expect(err).NotTo(HaveOccurred())
			Expect(product.ProductFiles).To(HaveLen(2))

			Expect(product.ProductFiles[0].AWSObjectKey).To(Equal("anything"))
			Expect(product.ProductFiles[1].AWSObjectKey).To(Equal("something"))

			Expect(product.ProductFiles[0].Links.Download["href"]).To(Equal("/products/banana/releases/666/product_files/6/download"))
			Expect(product.ProductFiles[1].Links.Download["href"]).To(Equal("/products/banana/releases/666/product_files/8/download"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := client.GetProductFiles(release)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("Get Product File", func() {
		var (
			productSlug string
			productID   int
			releaseID   int

			response           pivnet.ProductFileResponse
			responseStatusCode int
		)

		BeforeEach(func() {
			productSlug = "banana"
			productID = 8
			releaseID = 12

			response = pivnet.ProductFileResponse{pivnet.ProductFile{
				ID:           productID,
				AWSObjectKey: "something",
				Links: &pivnet.Links{Download: map[string]string{
					"href": "/products/banana/releases/666/product_files/8/download"},
				},
			}}

			responseStatusCode = http.StatusOK
		})

		JustBeforeEach(func() {
			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						apiPrefix+"/products/banana/releases/12/product_files/8",
					),
					ghttp.RespondWithJSONEncoded(responseStatusCode, response),
				),
			)
		})

		It("returns the product file without error", func() {
			product, err := client.GetProductFile(
				productSlug,
				releaseID,
				productID,
			)
			Expect(err).NotTo(HaveOccurred())

			Expect(product.ID).To(Equal(productID))
			Expect(product.AWSObjectKey).To(Equal("something"))

			Expect(product.Links.Download["href"]).
				To(Equal("/products/banana/releases/666/product_files/8/download"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			BeforeEach(func() {
				responseStatusCode = http.StatusTeapot
			})

			It("returns an error", func() {
				_, err := client.GetProductFile(
					productSlug,
					releaseID,
					productID,
				)
				Expect(err).To(HaveOccurred())

				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("Create Product File", func() {
		var (
			createProductFileConfig pivnet.CreateProductFileConfig
		)

		BeforeEach(func() {
			createProductFileConfig = pivnet.CreateProductFileConfig{
				ProductSlug:  productSlug,
				Name:         "some-file-name",
				FileVersion:  "some-file-version",
				AWSObjectKey: "some-aws-object-key",
			}
		})

		Context("when the config is valid", func() {
			type requestBody struct {
				ProductFile pivnet.ProductFile `json:"product_file"`
			}

			const (
				expectedFileType = "Software"
			)

			var (
				expectedRequestBody requestBody

				validResponse = `{"product_file":{"id":1234}}`
			)

			BeforeEach(func() {
				expectedRequestBody = requestBody{
					ProductFile: pivnet.ProductFile{
						FileType:     expectedFileType,
						FileVersion:  createProductFileConfig.FileVersion,
						Name:         createProductFileConfig.Name,
						MD5:          createProductFileConfig.MD5,
						AWSObjectKey: createProductFileConfig.AWSObjectKey,
					},
				}
			})

			It("creates the release with the minimum required fields", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/product_files"),
						ghttp.VerifyJSONRepresenting(&expectedRequestBody),
						ghttp.RespondWith(http.StatusCreated, validResponse),
					),
				)

				productFile, err := client.CreateProductFile(createProductFileConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(productFile.ID).To(Equal(1234))
			})

			Context("when the optional description is present", func() {
				var (
					description string

					productFileResponse pivnet.ProductFileResponse
				)

				BeforeEach(func() {
					description = "some\nmulti-line\ndescription"

					expectedRequestBody.ProductFile.Description = description

					productFileResponse = pivnet.ProductFileResponse{pivnet.ProductFile{
						ID:          1234,
						Description: description,
					}}
				})

				It("creates the product file with the description field", func() {
					server.AppendHandlers(
						ghttp.CombineHandlers(
							ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/product_files"),
							ghttp.VerifyJSONRepresenting(&expectedRequestBody),
							ghttp.RespondWithJSONEncoded(http.StatusCreated, productFileResponse),
						),
					)

					createProductFileConfig.Description = description

					productFile, err := client.CreateProductFile(createProductFileConfig)
					Expect(err).NotTo(HaveOccurred())
					Expect(productFile.Description).To(Equal(description))
				})
			})
		})

		Context("when the server responds with a non-201 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", apiPrefix+"/products/"+productSlug+"/product_files"),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.CreateProductFile(createProductFileConfig)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 201")))
			})
		})

		Context("when the aws object key is empty", func() {
			BeforeEach(func() {
				createProductFileConfig = pivnet.CreateProductFileConfig{
					ProductSlug:  productSlug,
					Name:         "some-file-name",
					FileVersion:  "some-file-version",
					AWSObjectKey: "",
				}
			})

			It("returns an error", func() {
				_, err := client.CreateProductFile(createProductFileConfig)
				Expect(err).To(HaveOccurred())

				Expect(err.Error()).To(ContainSubstring("AWS object key"))
			})
		})
	})

	Describe("Delete Product File", func() {
		var (
			id = 1234
		)

		It("deletes the product file", func() {
			response := []byte(`{"product_file":{"id":1234}}`)

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"DELETE",
						fmt.Sprintf("%s/products/%s/product_files/%d", apiPrefix, productSlug, id)),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			productFile, err := client.DeleteProductFile(productSlug, id)
			Expect(err).NotTo(HaveOccurred())

			Expect(productFile.ID).To(Equal(id))
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"DELETE",
							fmt.Sprintf("%s/products/%s/product_files/%d", apiPrefix, productSlug, id)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.DeleteProductFile(productSlug, id)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 200")))
			})
		})
	})

	Describe("Add Product File", func() {
		var (
			productID     = 1234
			releaseID     = 2345
			productFileID = 3456

			expectedRequestBody = `{"product_file":{"id":3456}}`
		)

		Context("when the server responds with a 204 status code", func() {
			It("returns without error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%d/releases/%d/add_product_file",
							apiPrefix,
							productID,
							releaseID,
						)),
						ghttp.VerifyJSON(expectedRequestBody),
						ghttp.RespondWith(http.StatusNoContent, nil),
					),
				)

				err := client.AddProductFile(productID, releaseID, productFileID)
				Expect(err).NotTo(HaveOccurred())
			})
		})

		Context("when the server responds with a non-204 status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("PATCH", fmt.Sprintf(
							"%s/products/%d/releases/%d/add_product_file",
							apiPrefix,
							productID,
							releaseID,
						)),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				err := client.AddProductFile(productID, releaseID, productFileID)
				Expect(err).To(MatchError(errors.New(
					"Pivnet returned status code: 418 for the request - expected 204")))
			})
		})
	})
})
