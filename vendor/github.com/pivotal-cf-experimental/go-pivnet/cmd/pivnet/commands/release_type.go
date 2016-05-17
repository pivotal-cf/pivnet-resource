package commands

import "github.com/olekukonko/tablewriter"

type ReleaseTypesCommand struct {
}

func (command *ReleaseTypesCommand) Execute([]string) error {
	client := NewClient()
	releaseTypes, err := client.ReleaseTypes.Get()
	if err != nil {
		return err
	}

	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ReleaseTypes"})

		for _, r := range releaseTypes {
			table.Append([]string{r})
		}
		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(releaseTypes)
	case PrintAsYAML:
		return printYAML(releaseTypes)
	}

	return nil
}
