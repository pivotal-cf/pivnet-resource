package metadata

import "fmt"

type Metadata struct {
	Release      *Release      `yaml:"release,omitempty"`
	ProductFiles []ProductFile `yaml:"product_files,omitempty"`
	Dependencies []Dependency  `yaml:"dependencies,omitempty"`
	UpgradePaths []UpgradePath `yaml:"upgrade_paths,omitempty"`
	FileGroups   []FileGroup   `yaml:"file_groups,omitempty"`
}

type Release struct {
	Version               string               `yaml:"version"`
	ReleaseType           string               `yaml:"release_type"`
	EULASlug              string               `yaml:"eula_slug"`
	ReleaseDate           string               `yaml:"release_date"`
	Description           string               `yaml:"description"`
	ReleaseNotesURL       string               `yaml:"release_notes_url"`
	Availability          string               `yaml:"availability"`
	UserGroupIDs          []string             `yaml:"user_group_ids,omitempty"`
	Controlled            bool                 `yaml:"controlled"`
	ECCN                  string               `yaml:"eccn"`
	LicenseException      string               `yaml:"license_exception"`
	EndOfSupportDate      string               `yaml:"end_of_support_date"`
	EndOfGuidanceDate     string               `yaml:"end_of_guidance_date"`
	EndOfAvailabilityDate string               `yaml:"end_of_availability_date"`
	ProductFiles          []ReleaseProductFile `yaml:"product_files,omitempty"`
}

type ReleaseProductFile struct {
	ID int `yaml:"id,omitempty"`
}

type ProductFile struct {
	File         string `yaml:"file,omitempty"`
	Description  string `yaml:"description,omitempty"`
	UploadAs     string `yaml:"upload_as,omitempty"`
	AWSObjectKey string `yaml:"aws_object_key,omitempty"`
	FileType     string `yaml:"file_type,omitempty"`
	FileVersion  string `yaml:"file_version,omitempty"`
	MD5          string `yaml:"md5,omitempty"`
	ID           int    `yaml:"id,omitempty"`
}

type FileGroup struct {
	ID           int                    `yaml:"id,omitempty"`
	Name         string                 `yaml:"name,omitempty"`
	ProductFiles []FileGroupProductFile `yaml:"product_files,omitempty"`
}

type FileGroupProductFile struct {
	ID int `yaml:"id,omitempty"`
}

type Dependency struct {
	Release DependentRelease `yaml:"release,omitempty"`
}

type UpgradePath struct {
	ID      int    `yaml:"id,omitempty"`
	Version string `yaml:"version,omitempty"`
}

type DependentRelease struct {
	ID      int     `yaml:"id,omitempty"`
	Version string  `yaml:"version,omitempty"`
	Product Product `yaml:"product,omitempty"`
}

type Product struct {
	ID   int    `yaml:"id,omitempty"`
	Slug string `yaml:"slug,omitempty"`
	Name string `yaml:"name,omitempty"`
}

func (m Metadata) Validate() error {
	for _, productFile := range m.ProductFiles {
		if productFile.File == "" {
			return fmt.Errorf("empty value for file")
		}
	}

	if m.Release == nil {
		return fmt.Errorf("missing required value %q", "release")
	}

	if m.Release.Version == "" {
		return fmt.Errorf("missing required value %q", "version")
	}

	if m.Release.ReleaseType == "" {
		return fmt.Errorf("missing required value %q", "release_type")
	}

	if m.Release.EULASlug == "" {
		return fmt.Errorf("missing required value %q", "eula_slug")
	}

	for i, d := range m.Dependencies {
		dependentReleaseID := d.Release.ID
		if dependentReleaseID == 0 {
			if d.Release.Version == "" || d.Release.Product.Slug == "" {
				return fmt.Errorf(
					"Either ReleaseID or release version and product slug must be provided for dependency[%d]",
					i,
				)
			}
		}
	}

	for i, u := range m.UpgradePaths {
		if u.ID == 0 && u.Version == "" {
			return fmt.Errorf(
				"Either id or version must be provided for upgrade_paths[%d]",
				i,
			)
		}
	}

	return nil
}
