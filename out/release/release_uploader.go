package release

import (
	"fmt"
	"log"
	"path/filepath"

	pivnet "github.com/pivotal-cf/go-pivnet"
	"github.com/pivotal-cf/pivnet-resource/metadata"
)

type ReleaseUploader struct {
	s3          s3Client
	pivnet      uploadClient
	logger      *log.Logger
	md5Summer   md5Summer
	metadata    metadata.Metadata
	sourcesDir  string
	productSlug string
}

//go:generate counterfeiter --fake-name UploadClient . uploadClient
type uploadClient interface {
	FindProductForSlug(slug string) (pivnet.Product, error)
	CreateProductFile(pivnet.CreateProductFileConfig) (pivnet.ProductFile, error)
	AddProductFile(productSlug string, releaseID int, productFileID int) error
	GetProductFiles(productSlug string) ([]pivnet.ProductFile, error)
	DeleteProductFile(productSlug string, releaseID int) (pivnet.ProductFile, error)
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
	sourcesDir,
	productSlug string,
) ReleaseUploader {
	return ReleaseUploader{
		s3:          s3,
		pivnet:      pivnet,
		logger:      logger,
		md5Summer:   md5Summer,
		metadata:    metadata,
		sourcesDir:  sourcesDir,
		productSlug: productSlug,
	}
}

func (u ReleaseUploader) Upload(release pivnet.Release, exactGlobs []string) error {
	for _, exactGlob := range exactGlobs {
		fullFilepath := filepath.Join(u.sourcesDir, exactGlob)
		fileContentsMD5, err := u.md5Summer.SumFile(fullFilepath)
		if err != nil {
			return err
		}

		u.logger.Println(fmt.Sprintf("uploading to s3: '%s'", exactGlob))

		awsObjectKey, err := u.s3.UploadFile(exactGlob)
		if err != nil {
			return err
		}

		filename := filepath.Base(exactGlob)

		var description string
		uploadAs := filename
		fileType := "Software"

		for _, f := range u.metadata.ProductFiles {
			if f.File == exactGlob {
				u.logger.Println(fmt.Sprintf(
					"exact glob '%s' matches metadata file: '%s'",
					exactGlob,
					f.File,
				))

				if f.UploadAs != "" {
					u.logger.Println(fmt.Sprintf(
						"uploading '%s' to remote filename: '%s' instead",
						exactGlob,
						f.UploadAs,
					))
					uploadAs = f.UploadAs
				}

				description = f.Description
				if f.FileType != "" {
					fileType = f.FileType
				}
			} else {
				u.logger.Println(fmt.Sprintf(
					"exact glob '%s' does not match metadata file: '%s'",
					exactGlob,
					f.File,
				))
			}
		}

		u.logger.Println(fmt.Sprintf("Getting all existing product files"))

		productFiles, err := u.pivnet.GetProductFiles(u.productSlug)
		if err != nil {
			return err
		}

		for _, pf := range productFiles {
			if pf.AWSObjectKey == awsObjectKey {
				u.logger.Println(fmt.Sprintf("Deleting existing product file with AWSObjectKey: '%s'", pf.AWSObjectKey))

				_, err = u.pivnet.DeleteProductFile(u.productSlug, pf.ID)
				if err != nil {
					return err
				}

				break
			}
		}

		u.logger.Println(fmt.Sprintf(
			"Creating product file with remote name: '%s'",
			uploadAs,
		))

		productFile, err := u.pivnet.CreateProductFile(pivnet.CreateProductFileConfig{
			ProductSlug:  u.productSlug,
			Name:         uploadAs,
			AWSObjectKey: awsObjectKey,
			FileVersion:  release.Version,
			MD5:          fileContentsMD5,
			Description:  description,
			FileType:     fileType,
		})
		if err != nil {
			return err
		}

		u.logger.Println(fmt.Sprintf(
			"Adding product file: '%s' with ID: %d",
			filename,
			productFile.ID,
		))

		err = u.pivnet.AddProductFile(u.productSlug, release.ID, productFile.ID)
		if err != nil {
			return err
		}
	}

	return nil
}
