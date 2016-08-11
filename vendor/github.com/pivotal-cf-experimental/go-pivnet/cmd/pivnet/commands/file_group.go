package commands

import "github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/commands/filegroup"

type FileGroupsCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"r" description:"Release version e.g. 0.1.2-rc1"`
}

type FileGroupCommand struct {
	ProductSlug string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	FileGroupID int    `long:"file-group-id" description:"Filegroup ID e.g. 1234" required:"true"`
}

type DeleteFileGroupCommand struct {
	ProductSlug string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	FileGroupID int    `long:"file-group-id" description:"File group ID e.g. 1234" required:"true"`
}

//go:generate counterfeiter . FileGroupClient
type FileGroupClient interface {
	List(productSlug string, releaseVersion string) error
	Get(productSlug string, productFileID int) error
	Delete(productSlug string, productFileID int) error
}

var NewFileGroupClient = func() FileGroupClient {
	return filegroup.NewFileGroupClient(
		NewPivnetClient(),
		ErrorHandler,
		Pivnet.Format,
		OutputWriter,
		Printer,
	)
}

func (command *FileGroupsCommand) Execute([]string) error {
	Init()
	return NewFileGroupClient().List(command.ProductSlug, command.ReleaseVersion)
}

func (command *FileGroupCommand) Execute([]string) error {
	Init()
	return NewFileGroupClient().Get(command.ProductSlug, command.FileGroupID)
}

func (command *DeleteFileGroupCommand) Execute([]string) error {
	Init()
	return NewFileGroupClient().Delete(command.ProductSlug, command.FileGroupID)
}
