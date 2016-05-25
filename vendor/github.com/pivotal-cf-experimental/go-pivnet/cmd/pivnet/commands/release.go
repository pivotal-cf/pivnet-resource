package commands

import (
	"fmt"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
	"github.com/pivotal-cf-experimental/go-pivnet/extension"
)

type ReleasesCommand struct {
	ProductSlug string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
}

type ReleaseCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

type DeleteReleaseCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

func (command *ReleasesCommand) Execute([]string) error {
	client := NewClient()
	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	switch Pivnet.Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Version", "Description"})

		for _, r := range releases {
			table.Append([]string{
				strconv.Itoa(r.ID), r.Version, r.Description,
			})
		}
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return Printer.PrintJSON(releases)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(releases)
	}

	return nil
}

func (command *ReleaseCommand) Execute([]string) error {
	client := NewClient()
	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	var foundRelease pivnet.Release
	for _, r := range releases {
		if r.Version == command.ReleaseVersion {
			foundRelease = r
			break
		}
	}

	if foundRelease.Version != command.ReleaseVersion {
		return fmt.Errorf("release not found")
	}

	release, err := client.Releases.Get(command.ProductSlug, foundRelease.ID)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	extendedClient := extension.NewExtendedClient(client, Pivnet.Logger)
	etag, err := extendedClient.ReleaseETag(command.ProductSlug, foundRelease.ID)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	r := CLIRelease{
		release,
		etag,
	}

	switch Pivnet.Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Version", "Description", "ETag"})

		table.Append([]string{
			strconv.Itoa(release.ID), release.Version, release.Description, etag,
		})
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return Printer.PrintJSON(r)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(r)
	}

	return nil
}

func (command *DeleteReleaseCommand) Execute([]string) error {
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

	err = client.Releases.Delete(release, command.ProductSlug)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	if Pivnet.Format == printer.PrintAsTable {
		_, err = fmt.Fprintf(
			OutputWriter,
			"release %s deleted successfully for %s\n",
			command.ReleaseVersion,
			command.ProductSlug,
		)
	}

	return nil
}

type CLIRelease struct {
	pivnet.Release `yaml:",inline"`
	ETag           string `json:"etag,omitempty"`
}
