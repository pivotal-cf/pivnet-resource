package pivnet

type Response struct {
	Releases []Release `json:"releases"`
}

type Release struct {
	Availability string `json:"availability"`
	Eula         Eula   `json:"eula"`
	OSSCompliant string `json:"oss_compliant"`
	ReleaseDate  string `json:"release_date"`
	ReleaseType  string `json:"release_type"`
	Version      string `json:"version"`
}

type Eula struct {
	Slug string `json:"slug"`
}
