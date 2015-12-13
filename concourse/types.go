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
	APIToken    string `json:"api_token"`
	ProductName string `json:"product_name"`
}
