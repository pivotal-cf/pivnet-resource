package concourse

import "github.com/pivotal-cf-experimental/pivnet-resource/pivnet"

type Request struct {
	Source  Source            `json:"source"`
	Version map[string]string `json:"version"`
}

type Response []pivnet.Release

type Source struct {
	APIToken    string `json:"api_token"`
	ProductName string `json:"product_name"`
}
