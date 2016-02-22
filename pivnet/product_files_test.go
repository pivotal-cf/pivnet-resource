package pivnet_test

import (
	"encoding/json"
	"errors"
	"fmt"
	"net/http"

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

	Describe("Get Product Files", func() {
		It("returns the product files for the given release", func() {
			response, err := json.Marshal(pivnet.ProductFiles{[]pivnet.ProductFile{
				{ID: 3, AWSObjectKey: "anything", Links: &pivnet.Links{Download: map[string]string{"href": "/products/banana/releases/666/product_files/6/download"}}},
				{ID: 4, AWSObjectKey: "something", Links: &pivnet.Links{Download: map[string]string{"href": "/products/banana/releases/666/product_files/8/download"}}},
			},
			})
			Expect(err).NotTo(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases/666/product_files"),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

			release := pivnet.Release{
				Links: &pivnet.Links{
					ProductFiles: map[string]string{"href": apiAddress + apiPrefix + "/products/banana/releases/666/product_files"},
				},
			}

			product, err := client.GetProductFiles(release)
			Expect(err).NotTo(HaveOccurred())
			Expect(product.ProductFiles).To(HaveLen(2))

			Expect(product.ProductFiles[0].AWSObjectKey).To(Equal("anything"))
			Expect(product.ProductFiles[1].AWSObjectKey).To(Equal("something"))

			Expect(product.ProductFiles[0].Links.Download["href"]).To(Equal("/products/banana/releases/666/product_files/6/download"))
			Expect(product.ProductFiles[1].Links.Download["href"]).To(Equal("/products/banana/releases/666/product_files/8/download"))
		})

		Context("when the server responds with a non-2XX status code", func() {
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", apiPrefix+"/products/banana/releases/666/product_files"),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)
				release := pivnet.Release{
					Links: &pivnet.Links{
						ProductFiles: map[string]string{"href": apiAddress + apiPrefix + "/products/banana/releases/666/product_files"},
					},
				}

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
		)

		BeforeEach(func() {
			productSlug = "banana"
			productID = 8
			releaseID = 12
		})

		It("returns the product file for the given productSlug, productID and release ID", func() {
			response, err := json.Marshal(pivnet.ProductFileResponse{pivnet.ProductFile{
				ID:           productID,
				AWSObjectKey: "something",
				Links: &pivnet.Links{Download: map[string]string{
					"href": "/products/banana/releases/666/product_files/8/download"},
				},
			}})
			Expect(err).NotTo(HaveOccurred())

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest(
						"GET",
						apiPrefix+"/products/banana/releases/12/product_files/8",
					),
					ghttp.RespondWith(http.StatusOK, response),
				),
			)

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
			It("returns an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest(
							"GET",
							apiPrefix+"/products/banana/releases/12/product_files/8",
						),
						ghttp.RespondWith(http.StatusTeapot, nil),
					),
				)

				_, err := client.GetProductFile(
					productSlug,
					releaseID,
					productID,
				)
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

				release, err := client.CreateProductFile(createProductFileConfig)
				Expect(err).NotTo(HaveOccurred())
				Expect(release.ID).To(Equal(1234))
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
