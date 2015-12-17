package concourse

type Request struct {
	Source  Source            `json:"source"`
	Version map[string]string `json:"version"`
}

type Response []Release

type Release struct {
	ProductVersion string `json:"product_version"`
}

type Source struct {
	APIToken        string `json:"api_token"`
	ProductName     string `json:"product_name"`
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
}

type OutRequest struct {
	Params OutParams `json:"params"`
	Source Source    `json:"source"`
}

type OutParams struct {
	File           string `json:"file"`
	FilepathPrefix string `json:"s3_filepath_prefix"`
}
