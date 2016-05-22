package commands

import (
	"strconv"

	"github.com/olekukonko/tablewriter"
	"github.com/pivotal-cf-experimental/go-pivnet"
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
		return err
	}

	return printProduct(product)
}

func printProducts(products []pivnet.Product) error {
	switch Pivnet.Format {
	case PrintAsTable:
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
	case PrintAsJSON:
		return printJSON(products)
	case PrintAsYAML:
		return printYAML(products)
	}

	return nil
}

func printProduct(product pivnet.Product) error {
	switch Pivnet.Format {
	case PrintAsTable:
		table := tablewriter.NewWriter(OutputWriter)
		table.SetHeader([]string{"ID", "Slug", "Name"})

		productAsString := []string{
			strconv.Itoa(product.ID), product.Slug, product.Name,
		}
		table.Append(productAsString)
		table.Render()
		return nil
	case PrintAsJSON:
		return printJSON(product)
	case PrintAsYAML:
		return printYAML(product)
	}

	return nil

}

func (command *ProductsCommand) Execute([]string) error {
	client := NewClient()
	products, err := client.Products.List()
	if err != nil {
		return err
	}

	return printProducts(products)
}
