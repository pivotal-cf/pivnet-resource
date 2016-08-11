package filegroup

import (
	"fmt"
	"io"
	"strconv"
	"strings"

	"github.com/olekukonko/tablewriter"
	pivnet "github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/errorhandler"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

//go:generate counterfeiter . PivnetClient
type PivnetClient interface {
	FileGroups(productSlug string) ([]pivnet.FileGroup, error)
	FileGroupsForRelease(productSlug string, releaseID int) ([]pivnet.FileGroup, error)
	ReleaseForProductVersion(productSlug string, releaseVersion string) (pivnet.Release, error)
	FileGroup(productSlug string, fileGroupID int) (pivnet.FileGroup, error)
	DeleteFileGroup(productSlug string, fileGroupID int) (pivnet.FileGroup, error)
}

type FileGroupClient struct {
	pivnetClient PivnetClient
	eh           errorhandler.ErrorHandler
	format       string
	outputWriter io.Writer
	printer      printer.Printer
}

func NewFileGroupClient(
	pivnetClient PivnetClient,
	eh errorhandler.ErrorHandler,
	format string,
	outputWriter io.Writer,
	printer printer.Printer,
) *FileGroupClient {
	return &FileGroupClient{
		pivnetClient: pivnetClient,
		eh:           eh,
		format:       format,
		outputWriter: outputWriter,
		printer:      printer,
	}
}

func (c *FileGroupClient) List(productSlug string, releaseVersion string) error {
	if releaseVersion == "" {
		fileGroups, err := c.pivnetClient.FileGroups(productSlug)
		if err != nil {
			return c.eh.HandleError(err)
		}

		return c.printFileGroups(fileGroups)
	}

	release, err := c.pivnetClient.ReleaseForProductVersion(productSlug, releaseVersion)
	if err != nil {
		return c.eh.HandleError(err)
	}

	fileGroups, err := c.pivnetClient.FileGroupsForRelease(
		productSlug,
		release.ID,
	)
	if err != nil {
		return c.eh.HandleError(err)
	}

	return c.printFileGroups(fileGroups)
}

func (c *FileGroupClient) printFileGroups(fileGroups []pivnet.FileGroup) error {
	switch c.format {

	case printer.PrintAsTable:
		table := tablewriter.NewWriter(c.outputWriter)
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
		return c.printer.PrintJSON(fileGroups)
	case printer.PrintAsYAML:
		return c.printer.PrintYAML(fileGroups)
	}

	return nil
}

func (c *FileGroupClient) Get(productSlug string, fileGroupID int) error {
	fileGroup, err := c.pivnetClient.FileGroup(productSlug, fileGroupID)
	if err != nil {
		return c.eh.HandleError(err)
	}

	return c.printFileGroup(fileGroup)
}

func (c *FileGroupClient) printFileGroup(fileGroup pivnet.FileGroup) error {
	switch c.format {

	case printer.PrintAsTable:
		table := tablewriter.NewWriter(c.outputWriter)
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
		return c.printer.PrintJSON(fileGroup)
	case printer.PrintAsYAML:
		return c.printer.PrintYAML(fileGroup)
	}

	return nil
}

func (c *FileGroupClient) Delete(productSlug string, fileGroupID int) error {
	_, err := c.pivnetClient.DeleteFileGroup(productSlug, fileGroupID)
	if err != nil {
		return c.eh.HandleError(err)
	}

	if c.format == printer.PrintAsTable {
		_, err = fmt.Fprintf(
			c.outputWriter,
			"file group %d deleted successfully for %s\n",
			fileGroupID,
			productSlug,
		)
	}

	return nil
}
