package globs

import (
	"fmt"
	"path/filepath"

	"github.com/pivotal-golang/lager"
)

type Globber interface {
	ExactGlobs() ([]string, error)
}

type globber struct {
	fileGlob   string
	sourcesDir string

	logger lager.Logger
}

type GlobberConfig struct {
	FileGlob   string
	SourcesDir string

	Logger lager.Logger
}

func NewGlobber(config GlobberConfig) Globber {
	return &globber{
		fileGlob:   config.FileGlob,
		sourcesDir: config.SourcesDir,

		logger: config.Logger,
	}
}

func (g globber) ExactGlobs() ([]string, error) {
	matches, err := filepath.Glob(filepath.Join(g.sourcesDir, g.fileGlob))
	if err != nil {
		return nil, err
	}

	if len(matches) == 0 {
		return nil, fmt.Errorf("no matches found for pattern: %s", g.fileGlob)
	}

	absPathSourcesDir, err := filepath.Abs(g.sourcesDir)
	if err != nil {
		panic(err)
	}

	g.logger.Debug("Absolute path to sourcesDir", lager.Data{"sources dir": absPathSourcesDir})

	exactGlobs := []string{}
	for _, match := range matches {
		g.logger.Debug("Matched file", lager.Data{"file": match})

		absPath, err := filepath.Abs(match)
		if err != nil {
			panic(err)
		}

		exactGlob, err := filepath.Rel(absPathSourcesDir, absPath)
		if err != nil {
			panic(err)
		}

		g.logger.Debug("Exact globs matched", lager.Data{"glob": exactGlob, "matches": match})

		exactGlobs = append(exactGlobs, exactGlob)
	}

	return exactGlobs, nil
}
