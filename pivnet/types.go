package pivnet

type Response struct {
	Releases []Release `json:"releases,omitempty"`
}

type Release struct {
	ID           int    `json:"id,omitempty"`
	Availability string `json:"availability,omitempty"`
	Eula         Eula   `json:"eula,omitempty"`
	OSSCompliant string `json:"oss_compliant,omitempty"`
	ReleaseDate  string `json:"release_date,omitempty"`
	ReleaseType  string `json:"release_type,omitempty"`
	Version      string `json:"version,omitempty"`
	Links        Links  `json:"_links,omitempty"`
}

type Eula struct {
	Slug    string `json:"slug,omitempty"`
	ID      int    `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
	Links   Links  `json:"_links,omitempty"`
}

type ProductFiles struct {
	ProductFiles []ProductFile `json:"product_files,omitempty"`
}

type ProductFile struct {
	ID           int    `json:"id,omitempty"`
	AWSObjectKey string `json:"aws_object_key,omitempty"`
	Links        Links  `json:"_links,omitempty"`
}

type Links struct {
	Download     map[string]string `json:"download,omitempty"`
	ProductFiles map[string]string `json:"product_files,omitempty"`
}
