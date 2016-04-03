package metadata

import "fmt"

type Metadata struct {
	Release      Release       `yaml:"release"`
	ProductFiles []ProductFile `yaml:"product_files"`
}

type Release struct {
	Version         string `yaml:"version"`
	ReleaseType     string `yaml:"release_type"`
	EulaSlug        string `yaml:"eula_slug"`
	ReleaseDate     string `yaml:"release_date"`
	Description     string `yaml:"description"`
	ReleaseNotesURL string `yaml:"release_notes_url"`
	Availability    string `yaml:"availability"`
	UserGroupIDs    []int  `yaml:"user_group_ids"`
}

type ProductFile struct {
	File        string `yaml:"file"`
	Description string `yaml:"description"`
	UploadAs    string `yaml:"upload_as"`
}

func (m Metadata) Validate() error {
	for _, productFile := range m.ProductFiles {
		if productFile.File == "" {
			return fmt.Errorf("empty value for file")
		}
	}

	if m.Release.Version == "" {
		return fmt.Errorf("missing required value %q", "version")
	}

	if m.Release.ReleaseType == "" {
		return fmt.Errorf("missing required value %q", "release_type")
	}

	if m.Release.EulaSlug == "" {
		return fmt.Errorf("missing required value %q", "eula_slug")
	}

	return nil
}
