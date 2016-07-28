package commands

import (
	"fmt"
	"io"
	"os"

	pb "gopkg.in/cheggaaa/pb.v1"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/extension"
	"github.com/pivotal-cf-experimental/go-pivnet/logger"
)

type DownloadProductFileCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1" required:"true"`
	ProductFileID  int    `long:"product-file-id" description:"Product file ID e.g. 1234" required:"true"`
	Filepath       string `long:"filepath" description:"Local filepath to download file to e.g. /tmp/my-file" required:"true"`
	AcceptEULA     bool   `long:"accept-eula" description:"Automatically accept EULA if necessary"`
}

func (command *DownloadProductFileCommand) Execute([]string) error {
	client := NewClient()

	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	var release pivnet.Release
	for _, r := range releases {
		if r.Version == command.ReleaseVersion {
			release = r
			break
		}
	}

	if release.Version != command.ReleaseVersion {
		return fmt.Errorf("release not found")
	}

	extendedClient := extension.NewExtendedClient(client, Pivnet.Logger)

	downloadLink := fmt.Sprintf(
		"/products/%s/releases/%d/product_files/%d/download",
		command.ProductSlug,
		release.ID,
		command.ProductFileID,
	)

	Pivnet.Logger.Debug("Creating local file", logger.Data{"downloadLink": downloadLink, "localFilepath": command.Filepath})
	file, err := os.Create(command.Filepath)
	if err != nil {
		return err // not tested
	}

	Pivnet.Logger.Debug("Determining file size", logger.Data{"downloadLink": downloadLink})
	productFile, err := client.ProductFiles.GetForRelease(
		command.ProductSlug,
		release.ID,
		command.ProductFileID,
	)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	if command.AcceptEULA {
		Pivnet.Logger.Debug("Accepting EULA")
		err = client.EULA.Accept(command.ProductSlug, release.ID)
		if err != nil {
			return ErrorHandler.HandleError(err)
		}
	}

	progress := newProgressBar(productFile.Size, os.Stderr)
	onDemandProgress := &startOnDemandProgressBar{progress, false}

	multiWriter := io.MultiWriter(file, onDemandProgress)

	Pivnet.Logger.Debug(
		"Downloading link to local file",
		logger.Data{
			"downloadLink":  downloadLink,
			"localFilepath": command.Filepath,
		},
	)
	err = extendedClient.DownloadFile(multiWriter, downloadLink)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	progress.Finish()
	return nil
}

type startOnDemandProgressBar struct {
	progressbar *pb.ProgressBar
	started     bool
}

func (w *startOnDemandProgressBar) Write(b []byte) (int, error) {
	if !w.started {
		w.progressbar.Start()
		w.started = true
	}
	return w.progressbar.Write(b)
}

func newProgressBar(total int, output io.Writer) *pb.ProgressBar {
	progress := pb.New(total)

	progress.Output = output
	progress.ShowSpeed = true
	progress.Units = pb.U_BYTES
	progress.NotPrint = true

	return progress.SetWidth(80)
}
