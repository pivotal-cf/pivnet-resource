package commands

import (
	"encoding/json"
	"fmt"
	"io"
	"os"

	"gopkg.in/yaml.v2"

	"github.com/pivotal-cf-experimental/go-pivnet"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/lagershim"
	"github.com/pivotal-cf-experimental/go-pivnet/cmd/pivnet/version"
	"github.com/pivotal-cf-experimental/go-pivnet/extension"
	"github.com/pivotal-golang/lager"
	"github.com/robdimsdale/sanitizer"
)

const (
	PrintAsTable = "table"
	PrintAsJSON  = "json"
	PrintAsYAML  = "yaml"

	DefaultHost = "https://network.pivotal.io"
)

var (
	OutputWriter io.Writer
	LogWriter    io.Writer
)

type PivnetCommand struct {
	Version func() `short:"v" long:"version" description:"Print the version of Pivnet and exit"`

	Help HelpCommand `command:"help" description:"Print this help message"`

	Format  string `long:"format" description:"Format to print as" default:"table" choice:"table" choice:"json" choice:"yaml"`
	Verbose bool   `long:"verbose" description:"Display verbose output"`

	APIToken string `long:"api-token" description:"Pivnet API token"`
	Host     string `long:"host" description:"Pivnet API Host"`

	ReleaseTypes ReleaseTypesCommand `command:"release-types" description:"List release types"`

	EULAs      EULAsCommand      `command:"eulas" description:"List EULAs"`
	EULA       EULACommand       `command:"eula" description:"Show EULA"`
	AcceptEULA AcceptEULACommand `command:"accept-eula" description:"Accept EULA"`

	Products ProductsCommand `command:"products" description:"List products"`
	Product  ProductCommand  `command:"product" description:"Show product"`

	ProductFiles      ProductFilesCommand      `command:"product-files" description:"List product files"`
	ProductFile       ProductFileCommand       `command:"product-file" description:"Show product file"`
	AddProductFile    AddProductFileCommand    `command:"add-product-file" description:"Add product file to release"`
	RemoveProductFile RemoveProductFileCommand `command:"remove-product-file" description:"Remove product file from release"`
	DeleteProductFile DeleteProductFileCommand `command:"delete-product-file" description:"Delete product file"`

	FileGroups      FileGroupsCommand      `command:"file-groups" description:"List file groups"`
	FileGroup       FileGroupCommand       `command:"file-group" description:"Show file group"`
	DeleteFileGroup DeleteFileGroupCommand `command:"delete-file-group" description:"Delete file group"`

	Releases      ReleasesCommand      `command:"releases" description:"List releases"`
	Release       ReleaseCommand       `command:"release" description:"Show release"`
	DeleteRelease DeleteReleaseCommand `command:"delete-release" description:"Delete release"`

	UserGroups      UserGroupsCommand      `command:"user-groups" description:"List user groups"`
	UserGroup       UserGroupCommand       `command:"user-group" description:"Show user group"`
	AddUserGroup    AddUserGroupCommand    `command:"add-user-group" description:"Add user group to release"`
	CreateUserGroup CreateUserGroupCommand `command:"create-user-group" description:"Create user group"`
	UpdateUserGroup UpdateUserGroupCommand `command:"update-user-group" description:"Update user group"`
	DeleteUserGroup DeleteUserGroupCommand `command:"delete-user-group" description:"Delete user group"`

	ReleaseDependencies ReleaseDependenciesCommand `command:"release-dependencies" description:"List user groups"`

	ReleaseUpgradePaths ReleaseUpgradePathsCommand `command:"release-upgrade-paths" description:"List release upgrade paths"`
}

var Pivnet PivnetCommand

func init() {
	Pivnet.Version = func() {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	if Pivnet.Host == "" {
		Pivnet.Host = DefaultHost
	}

}

func NewClient() extension.ExtendedClient {
	if OutputWriter == nil {
		OutputWriter = os.Stdout
	}

	if LogWriter == nil {
		switch Pivnet.Format {
		case PrintAsJSON, PrintAsYAML:
			LogWriter = os.Stderr
			break
		default:
			LogWriter = os.Stdout
		}
	}

	l := lager.NewLogger("pivnet CLI")

	sanitized := map[string]string{
		Pivnet.APIToken: "*** redacted api token ***",
	}

	OutputWriter = sanitizer.NewSanitizer(sanitized, OutputWriter)
	LogWriter = sanitizer.NewSanitizer(sanitized, LogWriter)

	if Pivnet.Verbose {
		l.RegisterSink(lager.NewWriterSink(LogWriter, lager.DEBUG))
	} else {
		l.RegisterSink(lager.NewWriterSink(LogWriter, lager.INFO))
	}

	useragent := fmt.Sprintf(
		"go-pivnet/%s",
		version.Version,
	)

	ls := lagershim.NewLagerShim(l)

	pivnetClient := extension.NewExtendedClient(
		pivnet.ClientConfig{
			Token:     Pivnet.APIToken,
			Host:      Pivnet.Host,
			UserAgent: useragent,
		},
		ls,
	)

	return pivnetClient
}

func printYAML(object interface{}) error {
	b, err := yaml.Marshal(object)
	if err != nil {
		return err
	}

	output := fmt.Sprintf("---\n%s\n", string(b))
	OutputWriter.Write([]byte(output))
	return nil
}

func printJSON(object interface{}) error {
	b, err := json.Marshal(object)
	if err != nil {
		return err
	}

	OutputWriter.Write(b)
	return nil
}
