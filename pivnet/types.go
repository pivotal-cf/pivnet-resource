package pivnet

type Response struct {
	Releases []Release `json:"releases"`
}

type Release struct {
	Version string `json:"version"`
}
