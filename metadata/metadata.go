package metadata

import "fmt"

type Metadata struct {
	ProductFiles []ProductFile `yaml:"product_files"`
}

type ProductFile struct {
	File        string `yaml:"file"`
	Description string `yaml:"description"`
	UploadAs    string `yaml:"upload_as"`
}

func (m Metadata) Validate() error {
	for _, productFile := range m.ProductFiles {
		if productFile.File == "" {
			return fmt.Errorf("empty value for file")
		}
	}
	return nil
}
