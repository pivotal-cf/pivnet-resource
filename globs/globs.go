package globs

import (
	"fmt"
	"path/filepath"

	"github.com/pivotal-cf/go-pivnet/logger"
)

type Globber struct {
	fileGlob   string
	sourcesDir string

	logger logger.Logger
}

type GlobberConfig struct {
	FileGlob   string
	SourcesDir string

	Logger logger.Logger
}

func NewGlobber(config GlobberConfig) *Globber {
	return &Globber{
		fileGlob:   config.FileGlob,
		sourcesDir: config.SourcesDir,

		logger: config.Logger,
	}
}

func (g Globber) ExactGlobs() ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(g.sourcesDir, g.fileGlob))
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for pattern: '%s'", g.fileGlob)
	}

	absPathSourcesDir, err := filepath.Abs(g.sourcesDir)
	if err != nil {
		panic(err)
	}

	exactGlobs := []string{}
	for _, match := range matches {
		absPath, err := filepath.Abs(match)
		if err != nil {
			panic(err)
		}

		exactGlob, err := filepath.Rel(absPathSourcesDir, absPath)
		if err != nil {
			panic(err)
		}

		exactGlobs = append(exactGlobs, exactGlob)
	}

	return exactGlobs, nil
}
