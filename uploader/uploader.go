package uploader

import (
	"fmt"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
)

type Client interface {
	Upload() error
}

type client struct {
	fileGlob       string
	filepathPrefix string
	sourcesDir     string

	transport Transport
	logger    logger.Logger
}

type Config struct {
	FileGlob       string
	FilepathPrefix string
	SourcesDir     string

	Transport Transport
	Logger    logger.Logger
}

func NewClient(config Config) Client {
	return &client{
		fileGlob:       config.FileGlob,
		filepathPrefix: config.FilepathPrefix,
		sourcesDir:     config.SourcesDir,

		transport: config.Transport,
		logger:    config.Logger,
	}
}

func (c client) Upload() error {
	matches, err := filepath.Glob(filepath.Join(c.sourcesDir, c.fileGlob))
	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return fmt.Errorf("no matches found for pattern: %s", c.fileGlob)
	}

	absPathSourcesDir, err := filepath.Abs(c.sourcesDir)
	if err != nil {
		panic(err)
	}
	c.logger.Debugf("abs path to sourcesDir: %s\n", absPathSourcesDir)

	for _, match := range matches {
		c.logger.Debugf("matched file: %v\n", match)

		absPath, err := filepath.Abs(match)
		if err != nil {
			panic(err)
		}

		exactGlob, err := filepath.Rel(absPathSourcesDir, absPath)
		if err != nil {
			panic(err)
		}
		c.logger.Debugf("exact glob: %s\n", exactGlob)

		err = c.transport.Upload(
			exactGlob,
			"product_files/"+c.filepathPrefix+"/",
			c.sourcesDir,
		)
		if err != nil {
			return err
		}
	}

	return nil
}
