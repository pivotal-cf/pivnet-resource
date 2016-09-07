package commands

import "github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands/release"

type ReleasesCommand struct {
	ProductSlug string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
}

type ReleaseCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"r" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

type DeleteReleaseCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"r" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

//go:generate counterfeiter . ReleaseClient
type ReleaseClient interface {
	List(productSlug string) error
	Get(productSlug string, releaseVersion string) error
	Delete(productSlug string, releaseVersion string) error
}

var NewReleaseClient = func() ReleaseClient {
	return release.NewReleaseClient(
		NewPivnetClient(),
		ErrorHandler,
		Pivnet.Format,
		OutputWriter,
		Printer,
	)
}

func (command *ReleasesCommand) Execute([]string) error {
	Init()

	return NewReleaseClient().List(command.ProductSlug)
}

func (command *ReleaseCommand) Execute([]string) error {
	Init()

	return NewReleaseClient().Get(command.ProductSlug, command.ReleaseVersion)
}

func (command *DeleteReleaseCommand) Execute([]string) error {
	Init()

	return NewReleaseClient().Delete(command.ProductSlug, command.ReleaseVersion)
}
