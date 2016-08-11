package releaseupgradepath

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
	ReleaseUpgradePaths(productSlug string, releaseID int) ([]pivnet.ReleaseUpgradePath, error)
}

type ReleaseUpgradePathClient struct {
	pivnetClient PivnetClient
	eh           errorhandler.ErrorHandler
	format       string
	outputWriter io.Writer
	printer      printer.Printer
}

func NewReleaseUpgradePathClient(
	pivnetClient PivnetClient,
	eh errorhandler.ErrorHandler,
	format string,
	outputWriter io.Writer,
	printer printer.Printer,
) *ReleaseUpgradePathClient {
	return &ReleaseUpgradePathClient{
		pivnetClient: pivnetClient,
		eh:           eh,
		format:       format,
		outputWriter: outputWriter,
		printer:      printer,
	}
}

func (c *ReleaseUpgradePathClient) List(productSlug string, releaseVersion string) error {
	release, err := c.pivnetClient.ReleaseForProductVersion(productSlug, releaseVersion)
	if err != nil {
		return c.eh.HandleError(err)
	}

	releaseUpgradePaths, err := c.pivnetClient.ReleaseUpgradePaths(
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
		})

		for _, r := range releaseUpgradePaths {
			table.Append([]string{
				strconv.Itoa(r.Release.ID),
				r.Release.Version,
			})
		}
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return c.printer.PrintJSON(releaseUpgradePaths)
	case printer.PrintAsYAML:
		return c.printer.PrintYAML(releaseUpgradePaths)
	}

	return nil
}
