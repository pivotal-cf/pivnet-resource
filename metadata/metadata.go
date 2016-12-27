package metadata

import "fmt"

type Metadata struct {
	Release              *Release              `yaml:"release,omitempty"`
	ProductFiles         []ProductFile         `yaml:"product_files,omitempty"`
	Dependencies         []Dependency          `yaml:"dependencies,omitempty"`
	DependencySpecifiers []DependencySpecifier `yaml:"dependency_specifiers,omitempty"`
	UpgradePaths         []UpgradePath         `yaml:"upgrade_paths,omitempty"`
	FileGroups           []FileGroup           `yaml:"file_groups,omitempty"`
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
	MD5                string   `yaml:"md5,omitempty"`
	ID                 int      `yaml:"id,omitempty"`
	Version            string   `yaml:"version,omitempty"`
	DocsURL            string   `yaml:"docs_url,omitempty"`
	SystemRequirements []string `yaml:"system_requirements,omitempty"`
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

	for i, d := range m.Dependencies {
		dependentReleaseID := d.Release.ID
		if dependentReleaseID == 0 {
			if d.Release.Version == "" || d.Release.Product.Slug == "" {
				return nil, fmt.Errorf(
					"Either ReleaseID or release version and product slug must be provided for dependency[%d]",
					i,
				)
			}
		}
	}

	for i, u := range m.UpgradePaths {
		if u.ID == 0 && u.Version == "" {
			return nil, fmt.Errorf(
				"Either id or version must be provided for upgrade_paths[%d]",
				i,
			)
		}
	}

	var deprecations []string
	if len(m.Dependencies) > 0 {
		deprecations = append(
			deprecations,
			"Use of 'dependencies' is deprecated - use 'dependency_specifiers' instead",
		)
	}

	return deprecations, nil
}
