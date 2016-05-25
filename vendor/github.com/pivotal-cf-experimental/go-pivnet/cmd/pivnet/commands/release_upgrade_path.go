package commands

import (
	"fmt"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

type ReleaseUpgradePathsCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

func (command *ReleaseUpgradePathsCommand) Execute([]string) error {
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

	releaseUpgradePaths, err := client.ReleaseUpgradePaths.Get(command.ProductSlug, release.ID)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	switch Pivnet.Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
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
		return Printer.PrintJSON(releaseUpgradePaths)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(releaseUpgradePaths)
	}

	return nil
}
