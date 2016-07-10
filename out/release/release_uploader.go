package release

import (
	"log"
	"path/filepath"

	"github.com/pivotal-cf-experimental/pivnet-resource/metadata"
	"github.com/pivotal-cf-experimental/pivnet-resource/pivnet"
)

type ReleaseUploader struct {
	s3          s3Client
	pivnet      uploadClient
	logger      *log.Logger
	md5Summer   md5Summer
	metadata    metadata.Metadata
	skipUpload  bool
	sourcesDir  string
	productSlug string
}

//go:generate counterfeiter --fake-name UploadClient . uploadClient
type uploadClient interface {
	FindProductForSlug(slug string) (pivnet.Product, error)
	CreateProductFile(pivnet.CreateProductFileConfig) (pivnet.ProductFile, error)
	AddProductFile(productID int, releaseID int, productFileID int) error
}

//go:generate counterfeiter --fake-name S3Client . s3Client
type s3Client interface {
	UploadFile(string) (string, error)
}

//go:generate counterfeiter --fake-name Md5Summer . md5Summer
type md5Summer interface {
	SumFile(filepath string) (string, error)
}

func NewReleaseUploader(
	s3 s3Client,
	pivnet uploadClient,
	logger *log.Logger,
	md5Summer md5Summer,
	metadata metadata.Metadata,
	skip bool,
	sourcesDir,
	productSlug string,
) ReleaseUploader {
	return ReleaseUploader{
		s3:          s3,
		pivnet:      pivnet,
		logger:      logger,
		md5Summer:   md5Summer,
		metadata:    metadata,
		skipUpload:  skip,
		sourcesDir:  sourcesDir,
		productSlug: productSlug,
	}
}

func (u ReleaseUploader) Upload(release pivnet.Release, exactGlobs []string) error {
	if u.skipUpload {
		u.logger.Println(
			"file glob and s3_filepath_prefix not provided - skipping upload to s3")
		return nil
	}

	for _, exactGlob := range exactGlobs {
		fullFilepath := filepath.Join(u.sourcesDir, exactGlob)
		fileContentsMD5, err := u.md5Summer.SumFile(fullFilepath)
		if err != nil {
			return err
		}

		u.logger.Printf("uploading file to s3: %s\n", exactGlob)

		remotePath, err := u.s3.UploadFile(exactGlob)
		if err != nil {
			return err
		}

		product, err := u.pivnet.FindProductForSlug(u.productSlug)
		if err != nil {
			return err
		}

		filename := filepath.Base(exactGlob)

		var description string
		uploadAs := filename
		fileType := "Software"

		for _, f := range u.metadata.ProductFiles {
			if f.File == exactGlob {
				u.logger.Printf(
					"exact glob '%s' matches metadata file: '%s'\n",
					exactGlob,
					f.File,
				)

				description = f.Description

				if f.UploadAs != "" {
					u.logger.Printf(
						"upload_as provided for exact glob: '%s' - uploading to remote filename: '%s' instead\n",
						exactGlob,
						f.UploadAs,
					)
					uploadAs = f.UploadAs
				}
				if f.FileType != "" {
					fileType = f.FileType
				}
			} else {
				u.logger.Printf(
					"exact glob %s does not match metadata file: %s\n",
					exactGlob,
					f.File,
				)
			}
		}

		u.logger.Printf(
			"Creating product file with product_slug: %s and filename: %s",
			u.productSlug,
			uploadAs,
		)

		productFile, err := u.pivnet.CreateProductFile(pivnet.CreateProductFileConfig{
			ProductSlug:  u.productSlug,
			Name:         uploadAs,
			AWSObjectKey: remotePath,
			FileVersion:  release.Version,
			MD5:          fileContentsMD5,
			Description:  description,
			FileType:     fileType,
		})
		if err != nil {
			return err
		}

		u.logger.Printf(
			"Adding product file with product_slug: %s and filename: %s",
			u.productSlug,
			filename,
		)

		err = u.pivnet.AddProductFile(product.ID, release.ID, productFile.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
