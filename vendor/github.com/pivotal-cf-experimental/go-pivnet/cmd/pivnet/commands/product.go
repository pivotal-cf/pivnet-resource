package commands

import (
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/printer"
)

type ProductCommand struct {
	ProductSlug string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
}

type ProductsCommand struct {
}

func (command *ProductCommand) Execute([]string) error {
	client := NewClient()
	product, err := client.Products.Get(command.ProductSlug)
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	return printProduct(product)
}

func printProducts(products []pivnet.Product) error {
	switch Pivnet.Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		for _, product := range products {
			productAsString := []string{
				strconv.Itoa(product.ID), product.Slug, product.Name,
			}
			table.Append(productAsString)
		}
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return Printer.PrintJSON(products)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(products)
	}

	return nil
}

func printProduct(product pivnet.Product) error {
	switch Pivnet.Format {
	case printer.PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		productAsString := []string{
			strconv.Itoa(product.ID), product.Slug, product.Name,
		}
		table.Append(productAsString)
		table.Render()
		return nil
	case printer.PrintAsJSON:
		return Printer.PrintJSON(product)
	case printer.PrintAsYAML:
		return Printer.PrintYAML(product)
	}

	return nil

}

func (command *ProductsCommand) Execute([]string) error {
	client := NewClient()
	products, err := client.Products.List()
	if err != nil {
		return ErrorHandler.HandleError(err)
	}

	return printProducts(products)
}
