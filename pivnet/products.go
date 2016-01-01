package pivnet

import "net/http"

func (c client) FindProductForSlug(slug string) (Product, error) {
	url := c.url + "/products/" + slug

	var response Product
	err := c.makeRequest(
		"GET",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return Product{}, err
	}

	return response, nil
}
