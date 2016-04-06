package metadata

import "fmt"

type Metadata struct {
	Release      *Release      `yaml:"release,omitempty"`
	ProductFiles []ProductFile `yaml:"product_files"`
}

type Release struct {
	Version               string   `yaml:"version"`
	ReleaseType           string   `yaml:"release_type"`
	EULASlug              string   `yaml:"eula_slug"`
	ReleaseDate           string   `yaml:"release_date"`
	Description           string   `yaml:"description"`
	ReleaseNotesURL       string   `yaml:"release_notes_url"`
	Availability          string   `yaml:"availability"`
	UserGroupIDs          []string `yaml:"user_group_ids"`
	Controlled            bool     `yaml:"controlled"`
	ECCN                  string   `yaml:"eccn"`
	LicenseException      string   `yaml:"license_exception"`
	EndOfSupportDate      string   `yaml:"end_of_support_date"`
	EndOfGuidanceDate     string   `yaml:"end_of_guidance_date"`
	EndOfAvailabilityDate string   `yaml:"end_of_availability_date"`
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

	if m.Release != nil {
		if m.Release.Version == "" {
			return fmt.Errorf("missing required value %q", "version")
		}

		if m.Release.ReleaseType == "" {
			return fmt.Errorf("missing required value %q", "release_type")
		}

		if m.Release.EULASlug == "" {
			return fmt.Errorf("missing required value %q", "eula_slug")
		}
	}

	return nil
}
