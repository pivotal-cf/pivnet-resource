package commands

import (
	"fmt"
	"io"
	"log"
	"net/http"
	"os"

	"github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/errorhandler"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/gp"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/logshim"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/printer"
	"github.com/pivotal-cf/go-pivnet/cmd/pivnet/version"
	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/robdimsdale/sanitizer"
)

type PivnetClient interface {
	Products() ([]pivnet.Product, error)
	FindProductForSlug(slug string) (pivnet.Product, error)

	ReleaseTypes() ([]pivnet.ReleaseType, error)

	ReleasesForProductSlug(productSlug string) ([]pivnet.Release, error)
	Release(productSlug string, releaseID int) (pivnet.Release, error)
	DeleteRelease(productSlug string, release pivnet.Release) error

	ReleaseDependencies(productSlug string, releaseID int) ([]pivnet.ReleaseDependency, error)

	ReleaseUpgradePaths(productSlug string, releaseID int) ([]pivnet.ReleaseUpgradePath, error)

	AcceptEULA(productSlug string, releaseID int) error
	EULAs() ([]pivnet.EULA, error)
	EULA(eulaSlug string) (pivnet.EULA, error)

	GetProductFiles(productSlug string) ([]pivnet.ProductFile, error)
	GetProductFile(productSlug string, productFileID int) (pivnet.ProductFile, error)
	GetProductFilesForRelease(productSlug string, releaseID int) ([]pivnet.ProductFile, error)
	GetProductFileForRelease(productSlug string, releaseID int, productFileID int) (pivnet.ProductFile, error)
	DeleteProductFile(productSlug string, releaseID int) (pivnet.ProductFile, error)
	AddProductFile(productSlug string, releaseID int, productFileID int) error
	RemoveProductFile(productSlug string, releaseID int, productFileID int) error

	FileGroups(productSlug string) ([]pivnet.FileGroup, error)
	FileGroupsForRelease(productSlug string, releaseID int) ([]pivnet.FileGroup, error)
	FileGroup(productSlug string, fileGroupID int) (pivnet.FileGroup, error)
	CreateFileGroup(productSlug string, name string) (pivnet.FileGroup, error)
	DeleteFileGroup(productSlug string, fileGroupID int) (pivnet.FileGroup, error)

	UserGroups() ([]pivnet.UserGroup, error)
	UserGroupsForRelease(productSlug string, releaseID int) ([]pivnet.UserGroup, error)
	UserGroup(userGroupID int) (pivnet.UserGroup, error)
	CreateUserGroup(name string, description string, members []string) (pivnet.UserGroup, error)
	UpdateUserGroup(userGroup pivnet.UserGroup) (pivnet.UserGroup, error)
	DeleteUserGroup(userGroupID int) error
	AddUserGroup(productSlug string, releaseID int, userGroupID int) error
	RemoveUserGroup(productSlug string, releaseID int, userGroupID int) error
	AddMemberToGroup(userGroupID int, emailAddress string, admin bool) (pivnet.UserGroup, error)
	RemoveMemberFromGroup(userGroupID int, emailAddress string) (pivnet.UserGroup, error)

	CreateRequest(method string, url string, body io.Reader) (*http.Request, error)
	MakeRequest(method string, url string, expectedResponseCode int, body io.Reader, data interface{}) (*http.Response, error)
}

const (
	DefaultHost = "https://network.pivotal.io"
)

var (
	OutputWriter io.Writer
	LogWriter    io.Writer

	ErrorHandler errorhandler.ErrorHandler
	Printer      printer.Printer
)

type PivnetCommand struct {
	VersionFunc func() `short:"v" long:"version" description:"Print the version of this CLI and exit"`

	Format  string `long:"format" description:"Format to print as" default:"table" choice:"table" choice:"json" choice:"yaml"`
	Verbose bool   `long:"verbose" description:"Display verbose output"`

	APIToken string `long:"api-token" description:"Pivnet API token"`
	Host     string `long:"host" description:"Pivnet API Host"`

	Help    HelpCommand    `command:"help" alias:"h" description:"Print this help message"`
	Version VersionCommand `command:"version" alias:"v" description:"Print the version of this CLI and exit"`

	ReleaseTypes ReleaseTypesCommand `command:"release-types" alias:"rts" description:"List release types"`

	EULAs      EULAsCommand      `command:"eulas" alias:"es" description:"List EULAs"`
	EULA       EULACommand       `command:"eula" alias:"e" description:"Show EULA"`
	AcceptEULA AcceptEULACommand `command:"accept-eula" alias:"ae" description:"Accept EULA"`

	Products ProductsCommand `command:"products" alias:"ps" description:"List products"`
	Product  ProductCommand  `command:"product" alias:"p" description:"Show product"`

	ProductFiles      ProductFilesCommand      `command:"product-files" alias:"pfs" description:"List product files"`
	ProductFile       ProductFileCommand       `command:"product-file" alias:"pf" description:"Show product file"`
	AddProductFile    AddProductFileCommand    `command:"add-product-file" alias:"apf" description:"Add product file to release"`
	RemoveProductFile RemoveProductFileCommand `command:"remove-product-file" alias:"rpf" description:"Remove product file from release"`
	DeleteProductFile DeleteProductFileCommand `command:"delete-product-file" alias:"dpf" description:"Delete product file"`

	DownloadProductFile DownloadProductFileCommand `command:"download-product-file" alias:"dlpf" description:"Download product file"`

	FileGroups      FileGroupsCommand      `command:"file-groups" alias:"fgs" description:"List file groups"`
	FileGroup       FileGroupCommand       `command:"file-group" alias:"fg" description:"Show file group"`
	CreateFileGroup CreateFileGroupCommand `command:"create-file-group" alias:"cfg" description:"Create file group"`
	UpdateFileGroup UpdateFileGroupCommand `command:"update-file-group" alias:"ufg" description:"Update file group"`
	DeleteFileGroup DeleteFileGroupCommand `command:"delete-file-group" alias:"dfg" description:"Delete file group"`

	Releases      ReleasesCommand      `command:"releases" alias:"rs" description:"List releases"`
	Release       ReleaseCommand       `command:"release" alias:"r" description:"Show release"`
	DeleteRelease DeleteReleaseCommand `command:"delete-release" alias:"dr" description:"Delete release"`

	UserGroups      UserGroupsCommand      `command:"user-groups" alias:"ugs" description:"List user groups"`
	UserGroup       UserGroupCommand       `command:"user-group" alias:"ug" description:"Show user group"`
	AddUserGroup    AddUserGroupCommand    `command:"add-user-group" alias:"aug" description:"Add user group to release"`
	RemoveUserGroup RemoveUserGroupCommand `command:"remove-user-group" alias:"rug" description:"Remove user group from release"`
	CreateUserGroup CreateUserGroupCommand `command:"create-user-group" alias:"cug" description:"Create user group"`
	UpdateUserGroup UpdateUserGroupCommand `command:"update-user-group" alias:"uug" description:"Update user group"`
	DeleteUserGroup DeleteUserGroupCommand `command:"delete-user-group" alias:"dug" description:"Delete user group"`

	AddUserGroupMember    AddUserGroupMemberCommand    `command:"add-user-group-member" alias:"augm" description:"Add user group member to group"`
	RemoveUserGroupMember RemoveUserGroupMemberCommand `command:"remove-user-group-member" alias:"rugm" description:"Remove user group member from group"`

	ReleaseDependencies ReleaseDependenciesCommand `command:"release-dependencies" alias:"rds" description:"List release dependencies"`

	ReleaseUpgradePaths ReleaseUpgradePathsCommand `command:"release-upgrade-paths" alias:"rups" description:"List release upgrade paths"`

	Logger    logger.Logger
	userAgent string
}

var Pivnet PivnetCommand

func init() {
	Pivnet.VersionFunc = func() {
		fmt.Println(version.Version)
		os.Exit(0)
	}

	if Pivnet.Host == "" {
		Pivnet.Host = DefaultHost
	}
}

func NewPivnetClient() *gp.CompositeClient {
	return gp.NewCompositeClient(
		pivnet.ClientConfig{
			Token:     Pivnet.APIToken,
			Host:      Pivnet.Host,
			UserAgent: Pivnet.userAgent,
		},
		Pivnet.Logger,
	)
}

func Init() {
	if OutputWriter == nil {
		OutputWriter = os.Stdout
	}

	if LogWriter == nil {
		switch Pivnet.Format {
		case printer.PrintAsJSON, printer.PrintAsYAML:
			LogWriter = os.Stderr
			break
		default:
			LogWriter = os.Stdout
		}
	}

	if ErrorHandler == nil {
		ErrorHandler = errorhandler.NewErrorHandler(Pivnet.Format, OutputWriter, LogWriter)
	}

	if Printer == nil {
		Printer = printer.NewPrinter(OutputWriter)
	}

	sanitized := map[string]string{
		Pivnet.APIToken: "*** redacted api token ***",
	}

	OutputWriter = sanitizer.NewSanitizer(sanitized, OutputWriter)
	LogWriter = sanitizer.NewSanitizer(sanitized, LogWriter)

	infoLogger := log.New(LogWriter, "", log.LstdFlags)
	debugLogger := log.New(LogWriter, "", log.LstdFlags)

	Pivnet.userAgent = fmt.Sprintf(
		"go-pivnet/%s",
		version.Version,
	)

	Pivnet.Logger = logshim.NewLogShim(infoLogger, debugLogger, Pivnet.Verbose)
}
