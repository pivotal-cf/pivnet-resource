package commands

import (
	"fmt"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
)

type ReleaseDependenciesCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

func (command *ReleaseDependenciesCommand) Execute([]string) error {
	client := NewClient()

	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return err
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

	releaseDependencies, err := client.ReleaseDependencies.List(command.ProductSlug, release.ID)
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
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
	case PrintAsJSON:
		return printJSON(releaseDependencies)
	case PrintAsYAML:
		return printYAML(releaseDependencies)
	}

	return nil
}
