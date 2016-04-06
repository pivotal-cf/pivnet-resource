package concourse

type Source struct {
	APIToken        string `json:"api_token"`
	ProductSlug     string `json:"product_slug"`
	AccessKeyID     string `json:"access_key_id"`
	ProductVersion  string `json:"product_version"`
	SecretAccessKey string `json:"secret_access_key"`
	Bucket          string `json:"bucket"`
	Endpoint        string `json:"endpoint"`
	Region          string `json:"region"`
	ReleaseType     string `json:"release_type"`
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
	Source  Source   `json:"source"`
	Version Version  `json:"version"`
	Params  InParams `json:"params"`
}

type InParams struct {
	Globs []string `json:"globs"`
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
	FileGlob            string `json:"file_glob"`
	FilepathPrefix      string `json:"s3_filepath_prefix"`
	VersionFile         string `json:"version_file"`
	ReleaseTypeFile     string `json:"release_type_file"`
	ReleaseDateFile     string `json:"release_date_file"`
	EULASlugFile        string `json:"eula_slug_file"`
	DescriptionFile     string `json:"description_file"`
	ReleaseNotesURLFile string `json:"release_notes_url_file"`
	AvailabilityFile    string `json:"availability_file"`
	UserGroupIDsFile    string `json:"user_group_ids_file"`
	MetadataFile        string `json:"metadata_file"`
}

type OutResponse struct {
	Version  Version    `json:"version"`
	Metadata []Metadata `json:"metadata,omitempty"`
}
