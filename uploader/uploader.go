package uploader

import (
	"fmt"
	"io"
	"os"
	"path/filepath"
)

type Client interface {
	Upload() error
}

type client struct {
	fileGlob       string
	filepathPrefix string
	sourcesDir     string

	transport   Transport
	debugWriter io.Writer
}

type Config struct {
	FileGlob       string
	FilepathPrefix string
	SourcesDir     string

	Transport   Transport
	DebugWriter io.Writer
}

func NewClient(config Config) Client {
	c := client{
		fileGlob:       config.FileGlob,
		filepathPrefix: config.FilepathPrefix,
		sourcesDir:     config.SourcesDir,

		transport:   config.Transport,
		debugWriter: config.DebugWriter,
	}

	if c.debugWriter == nil {
		c.debugWriter = os.Stderr
	}
	return c
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
	fmt.Fprintf(c.debugWriter, "abs path to sourcesDir: %s\n", absPathSourcesDir)

	for _, match := range matches {
		fmt.Fprintf(c.debugWriter, "matched file: %v\n", match)

		absPath, err := filepath.Abs(match)
		if err != nil {
			panic(err)
		}

		exactGlob, err := filepath.Rel(absPathSourcesDir, absPath)
		if err != nil {
			panic(err)
		}
		fmt.Fprintf(c.debugWriter, "exact glob: %s\n", exactGlob)

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
