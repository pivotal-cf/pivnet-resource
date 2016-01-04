package concourse

type Source struct {
	APIToken        string `json:"api_token"`
	ProductSlug     string `json:"product_slug"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

type CheckRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type Version struct {
	ProductVersion string `json:"product_version"`
}

type CheckResponse []Version

type InRequest struct {
	Source  Source  `json:"source"`
	Version Version `json:"version"`
}

type InResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata,omitempty"`
}

type Metadata struct {
	Name  string `json:"name,omitempty"`
	Value string `json:"value,omitempty"`
}

type OutRequest struct {
	Params OutParams `json:"params"`
	Source Source    `json:"source"`
}

type OutParams struct {
	FileGlob        string `json:"file_glob"`
	FilepathPrefix  string `json:"s3_filepath_prefix"`
	VersionFile     string `json:"version_file"`
	ReleaseTypeFile string `json:"release_type_file"`
	ReleaseDateFile string `json:"release_date_file"`
	EulaSlugFile    string `json:"eula_slug_file"`
	DescriptionFile string `json:"description_file"`
}

type OutResponse struct {
	Version  Version  `json:"version"`
	Metadata []string `json:"metadata,omitempty"`
}
