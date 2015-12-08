package pivnet

type Response struct {
	Releases []Release `json:"releases"`
}

type Release struct {
	Version string `json:"version"`
}

type ConcourseRequest struct {
	Source  ConcourseSource `json:"source"`
	Version struct{}        `json:"version"`
}

type ConcourseResponse []Release

type ConcourseSource struct {
	APIToken     string `json:"api_token"`
	ResourceName string `json:"resource_name"`
}
