package downloader_test

import (
	"fmt"
	"io/ioutil"
	"net/http"
	"os"

	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Downloader", func() {
	var (
		server     *ghttp.Server
		apiAddress string
		dir        string
	)

	BeforeEach(func() {
		var err error
		server = ghttp.NewServer()
		apiAddress = server.URL()
		dir, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())
	})

	AfterEach(func() {
		err := os.RemoveAll(dir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Download", func() {
		var token string

		Context("Success", func() {
			BeforeEach(func() {
				header := http.Header{}
				header.Add("Location", apiAddress+"/some-redirect-link")
				token = "1234"

				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/the-first-post", ""),
						ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", "1234")),
						ghttp.RespondWith(http.StatusFound, nil, header),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/some-redirect-link"),
						ghttp.RespondWith(http.StatusOK, make([]byte, 10, 14)),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/the-second-post", ""),
						ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", "1234")),
						ghttp.RespondWith(http.StatusFound, nil, header),
					),
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("GET", "/some-redirect-link"),
						ghttp.RespondWith(http.StatusOK, make([]byte, 10, 14)),
					),
				)
			})

			It("Downloads the files into the directory provided", func() {
				fileNames := map[string]string{
					"the-first-post":  apiAddress + "/the-first-post",
					"the-second-post": apiAddress + "/the-second-post",
				}

				err := downloader.Download(dir, fileNames, token)
				Expect(err).NotTo(HaveOccurred())

				dataDir, err := os.Open(dir)
				Expect(err).ShouldNot(HaveOccurred())

				files, err := dataDir.Readdir(2)
				Expect(err).ShouldNot(HaveOccurred())
				Expect(files).To(HaveLen(2))

				for _, f := range files {
					Expect(f.Size()).ToNot(BeZero())
				}
			})
		})

		Context("when it fails to make a request", func() {
			It("raises an error", func() {
				Expect(downloader.Download(dir, map[string]string{"^731drop": "&h%%%%"}, token)).NotTo(Succeed())
			})
		})
	})
})
