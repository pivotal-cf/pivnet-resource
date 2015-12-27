package uploader

type Client interface {
	Upload() error
}

type client struct {
	fileGlob       string
	filepathPrefix string
	sourcesDir     string

	transport Transport
}

type Config struct {
	FileGlob       string
	FilepathPrefix string
	SourcesDir     string

	Transport Transport
}

func NewClient(config Config) Client {
	return &client{
		fileGlob:       config.FileGlob,
		filepathPrefix: config.FilepathPrefix,
		sourcesDir:     config.SourcesDir,

		transport: config.Transport,
	}
}

func (c client) Upload() error {
	err := c.transport.Upload(
		c.fileGlob,
		"product_files/"+c.filepathPrefix+"/",
		c.sourcesDir,
	)

	return err
}
