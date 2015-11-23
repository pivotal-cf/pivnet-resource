package pivnet

import (
	"fmt"
	"net/http"
)

type Client interface {
	ProductVersions(string) ([]string, error)
}

type client struct {
	url string
}

func NewClient(url string) Client {
	return &client{
		url: url,
	}
}

func (c client) ProductVersions(id string) ([]string, error) {
	productURL := c.url + "/products/" + id
	fmt.Printf("productURL: %s\n", productURL)
	_, err := http.Get(productURL)
	return []string{"v0.0.0"}, err
}
