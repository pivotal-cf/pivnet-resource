package globs

import (
	"fmt"
	"log"
	"path/filepath"
)

type Globber interface {
	ExactGlobs() ([]string, error)
}

type globber struct {
	fileGlob   string
	sourcesDir string

	logger *log.Logger
}

type GlobberConfig struct {
	FileGlob   string
	SourcesDir string

	Logger *log.Logger
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

	g.logger.Printf("Absolute path to sourcesDir: %s", absPathSourcesDir)

	exactGlobs := []string{}
	for _, match := range matches {
		g.logger.Printf("Matched file: %s", match)

		absPath, err := filepath.Abs(match)
		if err != nil {
			panic(err)
		}

		exactGlob, err := filepath.Rel(absPathSourcesDir, absPath)
		if err != nil {
			panic(err)
		}

		g.logger.Printf("Exact glob %s matched %s", exactGlob, "matches")

		exactGlobs = append(exactGlobs, exactGlob)
	}

	return exactGlobs, nil
}
