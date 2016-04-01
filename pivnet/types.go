package pivnet

type ReleasesResponse struct {
	Releases []Release `json:"releases,omitempty"`
}

type CreateReleaseResponse struct {
	Release Release `json:"release,omitempty"`
}

type ProductFileResponse struct {
	ProductFile ProductFile `json:"product_file,omitempty"`
}

type Release struct {
	ID              int    `json:"id,omitempty"`
	Availability    string `json:"availability,omitempty"`
	Eula            *Eula  `json:"eula,omitempty"`
	OSSCompliant    string `json:"oss_compliant,omitempty"`
	ReleaseDate     string `json:"release_date,omitempty"`
	ReleaseType     string `json:"release_type,omitempty"`
	Version         string `json:"version,omitempty"`
	Links           *Links `json:"_links,omitempty"`
	Description     string `json:"description,omitempty"`
	ReleaseNotesURL string `json:"release_notes_url,omitempty"`
	ECCN            string `json:"eccn,omitempty"`
}

type Eula struct {
	Slug    string `json:"slug,omitempty"`
	ID      int    `json:"id,omitempty"`
	Version string `json:"version,omitempty"`
	Links   *Links `json:"_links,omitempty"`
}

type EulaResponse struct {
	AcceptedAt string `json:"accepted_at,omitempty"`
	Links      *Links `json:"_links,omitempty"`
}

type ProductFiles struct {
	ProductFiles []ProductFile `json:"product_files,omitempty"`
}

type ProductFile struct {
	ID           int    `json:"id,omitempty"`
	AWSObjectKey string `json:"aws_object_key,omitempty"`
	Links        *Links `json:"_links,omitempty"`
	FileType     string `json:"file_type,omitempty"`
	FileVersion  string `json:"file_version,omitempty"`
	Name         string `json:"name,omitempty"`
	MD5          string `json:"md5,omitempty"`
	Description  string `json:"description,omitempty"`
}

type Links struct {
	Eula           map[string]string `json:"eula,omitempty"`
	Download       map[string]string `json:"download,omitempty"`
	ProductFiles   map[string]string `json:"product_files,omitempty"`
	EULAAcceptance map[string]string `json:"eula_acceptance,omitempty"`
}

type Product struct {
	ID   int    `json:"id,omitempty"`
	Slug string `json:"slug"`
}

type UserGroups struct {
	UserGroups []UserGroup `json:"user_groups,omitempty"`
}

type UserGroup struct {
	ID          int    `json:"id,omitempty"`
	Name        string `json:"name,omitempty"`
	Description string `json:"description,omitempty"`
}
