package commands

import (
	"fmt"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

type FileGroupsCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"v" description:"Release version e.g. 0.1.2-rc1"`
}

type FileGroupCommand struct {
	ProductSlug string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	FileGroupID int    `long:"file-group-id" description:"Filegroup ID e.g. 1234" required:"true"`
}

type DeleteFileGroupCommand struct {
	ProductSlug string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	FileGroupID int    `long:"file-group-id" description:"File group ID e.g. 1234" required:"true"`
}

func (command *FileGroupsCommand) Execute([]string) error {
	client := NewClient()

	if command.ReleaseVersion == "" {
		fileGroups, err := client.FileGroups.List(
			command.ProductSlug,
		)
		if err != nil {
			return ErrorHandler.HandleError(err)
		}
		return printFileGroups(fileGroups)
	}

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

	fileGroups, err := client.FileGroups.ListForRelease(
		command.ProductSlug,
		release.ID,
	)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	return printFileGroups(fileGroups)
}

func printFileGroups(fileGroups []pivnet.FileGroup) error {
	switch Pivnet.Format {

	case printer.PrintAsTable:
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
	case printer.PrintAsJSON:
		return Printer.PrintJSON(fileGroups)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(fileGroups)
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
		return ErrorHandler.HandleError(err)
	}

	return printFileGroup(fileGroup)
}

func printFileGroup(fileGroup pivnet.FileGroup) error {
	switch Pivnet.Format {

	case printer.PrintAsTable:
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
	case printer.PrintAsJSON:
		return Printer.PrintJSON(fileGroup)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(fileGroup)
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
		return ErrorHandler.HandleError(err)
	}

	if Pivnet.Format == printer.PrintAsTable {
		_, err = fmt.Fprintf(
			OutputWriter,
			"file group %d deleted successfully for %s\n",
			command.FileGroupID,
			command.ProductSlug,
		)
	}

	return ErrorHandler.HandleError(err)
}
