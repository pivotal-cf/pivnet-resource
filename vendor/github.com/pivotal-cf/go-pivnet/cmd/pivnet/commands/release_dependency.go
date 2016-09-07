package commands

import "github.com/pivotal-cf/go-pivnet/cmd/pivnet/commands/releasedependency"

type ReleaseDependenciesCommand struct {
	ProductSlug    string `long:"product-slug" short:"p" description:"Product slug e.g. p-mysql" required:"true"`
	ReleaseVersion string `long:"release-version" short:"r" description:"Release version e.g. 0.1.2-rc1" required:"true"`
}

//go:generate counterfeiter . ReleaseDependencyClient
type ReleaseDependencyClient interface {
	List(productSlug string, releaseVersion string) error
}

var NewReleaseDependencyClient = func() ReleaseDependencyClient {
	return releasedependency.NewReleaseDependencyClient(
		NewPivnetClient(),
		ErrorHandler,
		Pivnet.Format,
		OutputWriter,
		Printer,
	)
}

func (command *ReleaseDependenciesCommand) Execute([]string) error {
	Init()

	return NewReleaseDependencyClient().List(command.ProductSlug, command.ReleaseVersion)
}
