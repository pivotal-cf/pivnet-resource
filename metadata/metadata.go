package metadata

import "fmt"

type Metadata struct {
	Release               *Release               `yaml:"release,omitempty"`
	ProductFiles          []ProductFile          `yaml:"product_files,omitempty"`
	DependencySpecifiers  []DependencySpecifier  `yaml:"dependency_specifiers,omitempty"`
	UpgradePathSpecifiers []UpgradePathSpecifier `yaml:"upgrade_path_specifiers,omitempty"`
	FileGroups            []FileGroup            `yaml:"file_groups,omitempty"`

	// Deprecated
	Dependencies []Dependency  `yaml:"dependencies,omitempty"`
	UpgradePaths []UpgradePath `yaml:"upgrade_paths,omitempty"`
}

type Release struct {
	ID                    int                  `yaml:"id,omitempty"`
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
	File               string   `yaml:"file,omitempty"`
	Description        string   `yaml:"description,omitempty"`
	UploadAs           string   `yaml:"upload_as,omitempty"`
	AWSObjectKey       string   `yaml:"aws_object_key,omitempty"`
	FileType           string   `yaml:"file_type,omitempty"`
	FileVersion        string   `yaml:"file_version,omitempty"`
	SHA256             string   `yaml:"sha256,omitempty"`
	MD5                string   `yaml:"md5,omitempty"`
	ID                 int      `yaml:"id,omitempty"`
	Version            string   `yaml:"version,omitempty"`
	DocsURL            string   `yaml:"docs_url,omitempty"`
	SystemRequirements []string `yaml:"system_requirements,omitempty"`
	Platforms          []string `yaml:"platforms,omitempty"`
	IncludedFiles      []string `yaml:"included_files,omitempty"`
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

type DependencySpecifier struct {
	ID          int    `yaml:"id,omitempty"`
	Specifier   string `yaml:"specifier,omitempty"`
	ProductSlug string `yaml:"product_slug,omitempty"`
}

type UpgradePathSpecifier struct {
	ID        int    `yaml:"id,omitempty"`
	Specifier string `yaml:"specifier,omitempty"`
}

func (m Metadata) Validate() ([]string, error) {
	for _, productFile := range m.ProductFiles {
		if productFile.File == "" {
			return nil, fmt.Errorf("empty value for file")
		}
	}

	if m.Release == nil {
		return nil, fmt.Errorf("missing required value %q", "release")
	}

	if m.Release.Version == "" {
		return nil, fmt.Errorf("missing required value %q", "version")
	}

	if m.Release.ReleaseType == "" {
		return nil, fmt.Errorf("missing required value %q", "release_type")
	}

	if m.Release.EULASlug == "" {
		return nil, fmt.Errorf("missing required value %q", "eula_slug")
	}

	for i, d := range m.DependencySpecifiers {
		if d.ProductSlug == "" {
			return nil, fmt.Errorf(
				"Dependent product slug must be provided for dependency_specifiers[%d]",
				i,
			)
		}
		if d.Specifier == "" {
			return nil, fmt.Errorf(
				"Specifier must be provided for dependency_specifiers[%d]",
				i,
			)
		}
	}

	for i, d := range m.UpgradePathSpecifiers {
		if d.Specifier == "" {
			return nil, fmt.Errorf(
				"Specifier must be provided for upgrade_path_specifiers[%d]",
				i,
			)
		}
	}

	if len(m.Dependencies) > 0 {
		return nil, fmt.Errorf(
			"'dependencies' is deprecated. Please use 'dependency_specifiers' to add all dependency metadata.",
		)
	}

	if len(m.UpgradePaths) > 0 {
		return nil, fmt.Errorf(
			"'upgrade_paths' is deprecated. Please use 'upgrade_path_specifiers' to add all upgrade path metadata.",
		)
	}

	var deprecations []string
	return deprecations, nil
}
