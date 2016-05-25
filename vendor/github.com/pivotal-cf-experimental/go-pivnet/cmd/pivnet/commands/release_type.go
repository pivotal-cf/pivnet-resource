package commands

import (
	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

type ReleaseTypesCommand struct {
}

func (command *ReleaseTypesCommand) Execute([]string) error {
	client := NewClient()
	releaseTypes, err := client.ReleaseTypes.Get()
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	switch Pivnet.Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ReleaseTypes"})

		for _, r := range releaseTypes {
			table.Append([]string{r})
		}
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return Printer.PrintJSON(releaseTypes)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(releaseTypes)
	}

	return nil
}
