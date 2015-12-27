package uploader

//go:generate counterfeiter . Transport

type Transport interface {
	Upload(fileGlob string, filepathPrefix string, sourcesDir string) error
}
