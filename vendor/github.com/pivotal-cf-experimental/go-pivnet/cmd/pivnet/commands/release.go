package commands

import (
	"fmt"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
)

type ReleasesCommand struct {
	ProductSlug string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
}

type ReleaseCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

type DeleteReleaseCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

func (command *ReleasesCommand) Execute([]string) error {
	client := NewClient()
	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Version", "Description"})

		for _, r := range releases {
			table.Append([]string{
				strconv.Itoa(r.ID), r.Version, r.Description,
			})
		}
		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(releases)
	case PrintAsYAML:
		return printYAML(releases)
	}

	return nil
}

func (command *ReleaseCommand) Execute([]string) error {
	client := NewClient()
	releases, err := client.Releases.List(command.ProductSlug)
	if err != nil {
		return err
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
		return err
	}

	etag, err := client.ReleaseETag(command.ProductSlug, foundRelease.ID)
	if err != nil {
		return err
	}

	r := CLIRelease{
		release,
		etag,
	}

	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Version", "Description", "ETag"})

		table.Append([]string{
			strconv.Itoa(release.ID), release.Version, release.Description, etag,
		})
		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(r)
	case PrintAsYAML:
		return printYAML(r)
	}

	return nil
}

func (command *DeleteReleaseCommand) Execute([]string) error {
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

	err = client.Releases.Delete(release, command.ProductSlug)
	if err != nil {
		return err
	}

	if Pivnet.Format == PrintAsTable {
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
