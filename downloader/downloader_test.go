package downloader_test

import (
	"fmt"
	"io/ioutil"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/pivotal-cf-experimental/pivnet-resource/downloader"

	. "github.com/onsi/ginkgo"
	. "github.com/onsi/gomega"
	"github.com/onsi/gomega/ghttp"
)

var _ = Describe("Downloader", func() {
	var (
		d          *downloader.Downloader
		server     *ghttp.Server
		apiAddress string
		dir        string
		logger     *log.Logger

		apiToken string
	)

	BeforeEach(func() {
		apiToken = "1234-abcd"
		server = ghttp.NewServer()
		apiAddress = server.URL()
		logger = log.New(ioutil.Discard, "doesn't matter", 0)

		var err error
		dir, err = ioutil.TempDir("", "pivnet-resource")
		Expect(err).NotTo(HaveOccurred())
	})

	JustBeforeEach(func() {
		d = downloader.NewDownloader(apiToken, dir, logger)
	})

	AfterEach(func() {
		err := os.RemoveAll(dir)
		Expect(err).NotTo(HaveOccurred())
	})

	Describe("Download", func() {
		var (
			fileNames map[string]string
		)

		It("follows redirects", func() {
			header := http.Header{}
			header.Add("Location", apiAddress+"/some-redirect-link")

			server.AppendHandlers(
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", "/the-first-post", ""),
					ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", apiToken)),
					ghttp.RespondWith(http.StatusFound, nil, header),
				),
				ghttp.CombineHandlers(
					ghttp.VerifyRequest("GET", "/some-redirect-link"),
					ghttp.RespondWith(http.StatusOK, make([]byte, 10, 14)),
				),
			)

			fileNames = map[string]string{
				"the-first-post": apiAddress + "/the-first-post",
			}

			_, err := d.Download(fileNames)
			Expect(err).NotTo(HaveOccurred())
		})

		It("Downloads the files into the directory provided", func() {
			fileNames = map[string]string{
				"file-0": apiAddress + "/post-0",
				"file-1": apiAddress + "/post-1",
				"file-2": apiAddress + "/post-2",
			}

			for i := 0; i < len(fileNames); i++ {
				url := fmt.Sprintf("/post-%d", i)
				server.RouteToHandler("POST", url, ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", url, ""),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf("contents-%d", i)),
				))
			}

			_, err := d.Download(fileNames)
			Expect(err).NotTo(HaveOccurred())
			Expect(len(server.ReceivedRequests())).To(Equal(3))

			dataDir, err := os.Open(dir)
			Expect(err).ShouldNot(HaveOccurred())

			files, err := dataDir.Readdir(3)
			Expect(err).ShouldNot(HaveOccurred())
			Expect(files).To(HaveLen(3))

			for _, f := range files {
				fullPath := filepath.Join(dir, f.Name())
				contents, err := ioutil.ReadFile(fullPath)
				Expect(err).ShouldNot(HaveOccurred())

				split := strings.Split(f.Name(), "-")
				index := split[1]

				expectedContents := []byte(fmt.Sprintf("contents-%s", index))
				Expect(contents).To(Equal(expectedContents))
			}
		})

		It("returns a list of (full) filepaths", func() {
			fileNames := map[string]string{
				"file-0": apiAddress + "/post-0",
				"file-1": apiAddress + "/post-1",
				"file-2": apiAddress + "/post-2",
			}

			for i := 0; i < len(fileNames); i++ {
				url := fmt.Sprintf("/post-%d", i)
				server.RouteToHandler("POST", url, ghttp.CombineHandlers(
					ghttp.VerifyRequest("POST", url, ""),
					ghttp.RespondWith(http.StatusOK, fmt.Sprintf("contents-%d", i)),
				))
			}

			filepaths, err := d.Download(fileNames)
			Expect(err).NotTo(HaveOccurred())

			Expect(len(filepaths)).To(Equal(3))

			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-0")))
			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-1")))
			Expect(filepaths).Should(ContainElement(filepath.Join(dir, "file-2")))
		})

		Context("when the user has not accepted the EULA", func() {
			It("raises an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/the-first-post", ""),
						ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", apiToken)),
						ghttp.RespondWith(http.StatusUnavailableForLegalReasons, nil, nil),
					),
				)

				fileNames := map[string]string{
					"the-first-post": apiAddress + "/the-first-post",
				}

				_, err := d.Download(fileNames)
				Expect(err).To(MatchError("the EULA has not been accepted for the file: the-first-post"))
			})
		})

		Context("when Pivnet returns any other non 302", func() {
			It("raises an error", func() {
				server.AppendHandlers(
					ghttp.CombineHandlers(
						ghttp.VerifyRequest("POST", "/the-first-post", ""),
						ghttp.VerifyHeaderKV("Authorization", fmt.Sprintf("Token %s", apiToken)),
						ghttp.RespondWith(http.StatusUnauthorized, nil, nil),
					),
				)

				fileNames := map[string]string{
					"the-first-post": apiAddress + "/the-first-post",
				}

				_, err := d.Download(fileNames)
				Expect(err).To(MatchError("pivnet returned an error code of 401 for the file: the-first-post"))
			})
		})

		Context("when it fails to make a request", func() {
			It("raises an error", func() {
				_, err := d.Download(map[string]string{"^731drop": "&h%%%%"})

				Expect(err).Should(HaveOccurred())
			})
		})

		Context("when the directory does not already exist", func() {
			BeforeEach(func() {
				dir = filepath.Join(dir, "sub_directory")
			})

			It("creates the directory", func() {
				_, err := d.Download(nil)
				Expect(err).NotTo(HaveOccurred())

				_, err = os.Open(dir)
				Expect(err).ShouldNot(HaveOccurred())
			})
		})
	})
})
