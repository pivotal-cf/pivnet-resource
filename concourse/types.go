package concourse

type Request struct {
	Source  Source            `json:"source"`
	Version map[string]string `json:"version"`
}

type Response []Release

// TODO: Rename to Version
type Release struct {
	ProductVersion string `json:"product_version"`
}

type Source struct {
	APIToken        string `json:"api_token"`
	ProductName     string `json:"product_name"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

type InResponse struct {
	Version  Release    `json:"version"`
	Metadata []Metadata `json:"metadata"`
}

type Metadata struct {
	Name  string `json:"name"`
	Value string `json:"value"`
}

type OutRequest struct {
	Params OutParams `json:"params"`
	Source Source    `json:"source"`
}

type OutParams struct {
	File            string `json:"file"`
	FilepathPrefix  string `json:"s3_filepath_prefix"`
	VersionFile     string `json:"version_file"`
	ReleaseTypeFile string `json:"release_type_file"`
	ReleaseDateFile string `json:"release_date_file"`
	EulaSlugFile    string `json:"eula_slug_file"`
}

type OutResponse struct {
	Version  Release  `json:"version"`
	Metadata []string `json:"metadata"`
}
