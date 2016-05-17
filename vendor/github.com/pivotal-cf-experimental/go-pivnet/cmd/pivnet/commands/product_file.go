package commands

import (
	"fmt"
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
)

type ProductFilesCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1"`
}

type ProductFileCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1"`
	ProductFileID  int    `long:"product-file-id" description:"Product file ID e.g. 1234" required:"true"`
}

type AddProductFileCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1" required:"true"`
	ProductFileID  int    `long:"product-file-id" description:"Product file ID e.g. 1234" required:"true"`
}

type RemoveProductFileCommand struct {
	ProductSlug    string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" description:"Release version e.g. 0.1.2-rc1" required:"true"`
	ProductFileID  int    `long:"product-file-id" description:"Product file ID e.g. 1234" required:"true"`
}

type DeleteProductFileCommand struct {
	ProductSlug   string `long:"product-slug" description:"Product slug e.g. p-mysql" required:"true"`
	ProductFileID int    `long:"product-file-id" description:"Product file ID e.g. 1234" required:"true"`
}

func (command *ProductFilesCommand) Execute([]string) error {
	client := NewClient()

	if command.ReleaseVersion == "" {
		productFiles, err := client.ProductFiles.List(
			command.ProductSlug,
		)
		if err != nil {
			return err
		}

		return printProductFiles(productFiles)
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

	productFiles, err := client.ProductFiles.ListForRelease(
		command.ProductSlug,
		release.ID,
	)
	if err != nil {
		return err
	}

	return printProductFiles(productFiles)
}

func printProductFiles(productFiles []pivnet.ProductFile) error {
	switch Pivnet.Format {

	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{
			"ID",
			"Name",
			"File Version",
			"AWS Object Key",
		})

		for _, productFile := range productFiles {
			productFileAsString := []string{
				strconv.Itoa(productFile.ID),
				productFile.Name,
				productFile.FileVersion,
				productFile.AWSObjectKey,
			}
			table.Append(productFileAsString)
		}
		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(productFiles)
	case PrintAsYAML:
		return printYAML(productFiles)
	}

	return nil
}

func printProductFile(productFile pivnet.ProductFile) error {
	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{
			"ID",
			"Name",
			"File Version",
			"File Type",
			"Description",
			"MD5",
			"AWS Object Key",
		})

		productFileAsString := []string{
			strconv.Itoa(productFile.ID),
			productFile.Name,
			productFile.FileVersion,
			productFile.FileType,
			productFile.Description,
			productFile.MD5,
			productFile.AWSObjectKey,
		}
		table.Append(productFileAsString)
		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(productFile)
	case PrintAsYAML:
		return printYAML(productFile)
	}

	return nil
}

func (command *ProductFileCommand) Execute([]string) error {
	client := NewClient()

	if command.ReleaseVersion == "" {
		productFile, err := client.ProductFiles.Get(
			command.ProductSlug,
			command.ProductFileID,
		)
		if err != nil {
			return err
		}
		return printProductFile(productFile)
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

	productFile, err := client.ProductFiles.GetForRelease(
		command.ProductSlug,
		release.ID,
		command.ProductFileID,
	)
	if err != nil {
		return err
	}

	return printProductFile(productFile)
}

func (command *AddProductFileCommand) Execute([]string) error {
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

	err = client.ProductFiles.AddToRelease(
		command.ProductSlug,
		release.ID,
		command.ProductFileID,
	)
	if err != nil {
		return err
	}

	if Pivnet.Format == PrintAsTable {
		_, err = fmt.Fprintf(
			OutputWriter,
			"product file %d added successfully to %s/%s\n",
			command.ProductFileID,
			command.ProductSlug,
			command.ReleaseVersion,
		)
	}

	return nil
}

func (command *RemoveProductFileCommand) Execute([]string) error {
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

	err = client.ProductFiles.RemoveFromRelease(
		command.ProductSlug,
		release.ID,
		command.ProductFileID,
	)
	if err != nil {
		return err
	}

	if Pivnet.Format == PrintAsTable {
		_, err = fmt.Fprintf(
			OutputWriter,
			"product file %d removed successfully from %s/%s\n",
			command.ProductFileID,
			command.ProductSlug,
			command.ReleaseVersion,
		)
	}

	return nil
}

func (command *DeleteProductFileCommand) Execute([]string) error {
	client := NewClient()

	productFile, err := client.ProductFiles.Delete(
		command.ProductSlug,
		command.ProductFileID,
	)
	if err != nil {
		return err
	}

	if Pivnet.Format == PrintAsTable {
		_, err = fmt.Fprintf(
			OutputWriter,
			"product file %d deleted successfully for %s\n",
			command.ProductFileID,
			command.ProductSlug,
		)
	}

	return printProductFile(productFile)
}
