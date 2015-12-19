package pivnet

type Response struct {
	Releases []Release `json:"releases"`
}

type Release struct {
	ID           int    `json:"id"`
	Availability string `json:"availability"`
	Eula         Eula   `json:"eula"`
	OSSCompliant string `json:"oss_compliant"`
	ReleaseDate  string `json:"release_date"`
	ReleaseType  string `json:"release_type"`
	Version      string `json:"version"`
	Links        Links  `json:"_links"`
}

type Eula struct {
	Slug    string `json:"slug"`
	ID      int    `json:"id"`
	Version string `json:"version"`
	Links   Links  `json:"_links"`
}

type ProductFiles struct {
	ProductFiles []ProductFile `json:"product_files"`
}

type ProductFile struct {
	ID           int
	AWSObjectKey string `json:"aws_object_key"`
	Links        Links  `json:"_links"`
}

type Links struct {
	Download     map[string]string `json:"download"`
	ProductFiles map[string]string `json:"product_files"`
}
