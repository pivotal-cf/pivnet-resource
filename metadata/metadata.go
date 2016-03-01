package metadata

type Metadata struct {
	ProductFiles []ProductFile `yaml:"product_files"`
}

type ProductFile struct {
	File        string `yaml:"file"`
	Description string `yaml:"description"`
}
