package releasedependency

import (
	"io"
	"strconv"

	"github.com/olekukonko/tablewriter"
	pivnet "github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

//go:generate counterfeiter . PivnetClient
type PivnetClient interface {
	ReleaseForProductVersion(productSlug string, releaseVersion string) (pivnet.Release, error)
	ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error)
}

type ReleaseDependencyClient struct {
	pivnetClient PivnetClient
	eh           errorhandler.ErrorHandler
	format       string
	outputWriter io.Writer
	printer      printer.Printer
}

func NewReleaseDependencyClient(
	pivnetClient PivnetClient,
	eh errorhandler.ErrorHandler,
	format string,
	outputWriter io.Writer,
	printer printer.Printer,
) *ReleaseDependencyClient {
	return &ReleaseDependencyClient{
		pivnetClient: pivnetClient,
		eh:           eh,
		format:       format,
		outputWriter: outputWriter,
		printer:      printer,
	}
}

func (c *ReleaseDependencyClient) List(productSlug string, releaseVersion string) error {
	release, err := c.pivnetClient.ReleaseForProductVersion(productSlug, releaseVersion)
	if err != nil {
		return c.eh.HandleError(err)
	}

	releaseDependencies, err := c.pivnetClient.ReleaseDependencies(
		productSlug,
		release.ID,
	)
	if err != nil {
		return c.eh.HandleError(err)
	}

	switch c.format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(c.outputWriter)
		table.SetHeader([]string{
			"ID",
			"Version",
			"Product ID",
			"Product Slug",
		})

		for _, r := range releaseDependencies {
			table.Append([]string{
				strconv.Itoa(r.Release.ID),
				r.Release.Version,
				strconv.Itoa(r.Release.Product.ID),
				r.Release.Product.Slug,
			})
		}
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return c.printer.PrintJSON(releaseDependencies)
	case printer.PrintAsYAML:
		return c.printer.PrintYAML(releaseDependencies)
	}

	return nil
}
