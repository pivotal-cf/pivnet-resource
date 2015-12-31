package pivnet

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
)

type CreateProductFileConfig struct {
	ProductName  string
	FileVersion  string
	AWSObjectKey string
	Name         string
}

func (c client) GetProductFiles(release Release) (ProductFiles, error) {
	productFiles := ProductFiles{}

	err := c.makeRequest(
		"GET",
		release.Links.ProductFiles["href"],
		http.StatusOK,
		nil,
		&productFiles,
	)
	if err != nil {
		return ProductFiles{}, err
	}

	return productFiles, nil
}

func (c client) CreateProductFile(config CreateProductFileConfig) (ProductFile, error) {
	url := c.url + "/products/" + config.ProductName + "/product_files"

	body := createProductFileBody{
		ProductFile: ProductFile{
			MD5:          "not-supported-yet",
			FileType:     "Software",
			FileVersion:  config.FileVersion,
			AWSObjectKey: config.AWSObjectKey,
			Name:         config.Name,
		},
	}

	b, err := json.Marshal(body)
	if err != nil {
		panic(err)
	}

	var response ProductFileResponse
	err = c.makeRequest(
		"POST",
		url,
		http.StatusCreated,
		bytes.NewReader(b),
		&response,
	)
	if err != nil {
		return ProductFile{}, err
	}

	return response.ProductFile, nil
}

type createProductFileBody struct {
	ProductFile ProductFile `json:"product_file"`
}

func (c client) DeleteProductFile(productName string, id int) (ProductFile, error) {
	url := fmt.Sprintf(
		"%s/products/%s/product_files/%d",
		c.url,
		productName,
		id,
	)

	var response ProductFileResponse
	err := c.makeRequest(
		"DELETE",
		url,
		http.StatusOK,
		nil,
		&response,
	)
	if err != nil {
		return ProductFile{}, err
	}

	return response.ProductFile, nil
}
