package s3

type Request struct {
	Params Params `json:"params"`
	Source Source `json:"source"`
}

type Params struct {
	File string `json:"file"`
	To   string `json:"to"`
}

type Source struct {
	AccessKeyID     string `json:"access_key_id"`
	SecretAccessKey string `json:"secret_access_key"`
	Bucket          string `json:"bucket"`
	RegionName      string `json:"region_name"`
	Regexp          string `json:"regexp"`
}
