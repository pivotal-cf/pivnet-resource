package uploader

import (
	"fmt"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
)

type Client interface {
	Upload() (map[string]string, error)
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

func (c client) Upload() (map[string]string, error) {
	matches, err := filepath.Glob(filepath.Join(c.sourcesDir, c.fileGlob))
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for pattern: %s", c.fileGlob)
	}

	absPathSourcesDir, err := filepath.Abs(c.sourcesDir)
	if err != nil {
		panic(err)
	}
	c.logger.Debugf("Absolute path to sourcesDir: %s\n", absPathSourcesDir)

	filenamePaths := make(map[string]string)
	for _, match := range matches {
		c.logger.Debugf("Matched file: %s\n", match)

		absPath, err := filepath.Abs(match)
		if err != nil {
			panic(err)
		}

		exactGlob, err := filepath.Rel(absPathSourcesDir, absPath)
		if err != nil {
			panic(err)
		}

		c.logger.Debugf(
			"Exact glob: %s for file %s\n",
			exactGlob,
			match,
		)

		filename := filepath.Base(match)
		remoteDir := "product_files/" + c.filepathPrefix + "/"
		remotePath := fmt.Sprintf("%s%s", remoteDir, filename)

		err = c.transport.Upload(
			exactGlob,
			remoteDir,
			c.sourcesDir,
		)
		if err != nil {
			return nil, err
		}

		filenamePaths[filename] = remotePath
	}

	return filenamePaths, nil
}
