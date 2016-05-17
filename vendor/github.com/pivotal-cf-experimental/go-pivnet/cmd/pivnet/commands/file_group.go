package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
)

type FileGroupsCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1"`
}

type FileGroupCommand struct {
	ProductSlug string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	FileGroupID int    `long:"file-group-id" description:"Filegroup ID e.g. 1234" required:"true"`
}

type DeleteFileGroupCommand struct {
	ProductSlug string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	FileGroupID int    `long:"file-group-id" description:"File group ID e.g. 1234" required:"true"`
}

func (command *FileGroupsCommand) Execute([]string) error {
	client := NewClient()

	if command.ReleaseVersion == "" {
		fileGroups, err := client.FileGroups.List(
			command.ProductSlug,
		)
		if err != nil {
			return err
		}
		return printFileGroups(fileGroups)
	}

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

	fileGroups, err := client.FileGroups.ListForRelease(
		command.ProductSlug,
		release.ID,
	)
	if err != nil {
		return err
	}

	return printFileGroups(fileGroups)
}

func printFileGroups(fileGroups []pivnet.FileGroup) error {
	switch Pivnet.Format {

	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{
			"ID",
			"Name",
			"Product File Names",
		})

		for _, fileGroup := range fileGroups {
			var productFileNames []string

			for _, productFile := range fileGroup.ProductFiles {
				productFileNames = append(productFileNames, productFile.Name)
			}

			fileGroupAsString := []string{
				strconv.Itoa(fileGroup.ID),
				fileGroup.Name,
				strings.Join(productFileNames, ", "),
			}
			table.Append(fileGroupAsString)
		}
		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(fileGroups)
	case PrintAsYAML:
		return printYAML(fileGroups)
	}

	return nil
}

func (command *FileGroupCommand) Execute([]string) error {
	client := NewClient()

	fileGroup, err := client.FileGroups.Get(
		command.ProductSlug,
		command.FileGroupID,
	)
	if err != nil {
		return err
	}

	return printFileGroup(fileGroup)
}

func printFileGroup(fileGroup pivnet.FileGroup) error {
	switch Pivnet.Format {

	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{
			"ID",
			"Name",
			"Product File Names",
		})

		var productFileNames []string

		for _, productFile := range fileGroup.ProductFiles {
			productFileNames = append(productFileNames, productFile.Name)
		}

		fileGroupAsString := []string{
			strconv.Itoa(fileGroup.ID),
			fileGroup.Name,
			strings.Join(productFileNames, ", "),
		}
		table.Append(fileGroupAsString)
		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(fileGroup)
	case PrintAsYAML:
		return printYAML(fileGroup)
	}

	return nil
}

func (command *DeleteFileGroupCommand) Execute([]string) error {
	client := NewClient()

	_, err := client.FileGroups.Delete(
		command.ProductSlug,
		command.FileGroupID,
	)
	if err != nil {
		return err
	}

	if Pivnet.Format == PrintAsTable {
		_, err = fmt.Fprintf(
			OutputWriter,
			"file group %d deleted successfully for %s\n",
			command.FileGroupID,
			command.ProductSlug,
		)
	}

	return err
}
