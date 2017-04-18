package release

import (
	"fmt"
	"path/filepath"
	"time"

	"github.com/pivotal-cf/go-pivnet/logger"
	"github.com/pivotal-cf/pivnet-resource/metadata"
	pivnet "github.com/pivotal-cf/go-pivnet"
)

type ReleaseUploader struct {
	s3            s3Client
	pivnet        uploadClient
	logger        logger.Logger
	sha256Summer  sha256Summer
	md5Summer     md5Summer
	metadata      metadata.Metadata
	sourcesDir    string
	productSlug   string
	asyncTimeout  time.Duration
	pollFrequency time.Duration
}

//go:generate counterfeiter --fake-name UploadClient . uploadClient
type uploadClient interface {
	FindProductForSlug(slug string) (pivnet.Product, error)
	CreateProductFile(pivnet.CreateProductFileConfig) (pivnet.ProductFile, error)
	AddProductFile(productSlug string, releaseID int, productFileID int) error
	ProductFiles(productSlug string) ([]pivnet.ProductFile, error)
	ProductFile(productSlug string, productFileID int) (pivnet.ProductFile, error)
	DeleteProductFile(productSlug string, releaseID int) (pivnet.ProductFile, error)
}

//go:generate counterfeiter --fake-name S3Client . s3Client
type s3Client interface {
	UploadFile(string) (string, error)
}

//go:generate counterfeiter --fake-name Sha256Summer . sha256Summer
type sha256Summer interface {
	SumFile(filepath string) (string, error)
}

//go:generate counterfeiter --fake-name Md5Summer . md5Summer
type md5Summer interface {
	SumFile(filepath string) (string, error)
}

func NewReleaseUploader(
	s3 s3Client,
	pivnet uploadClient,
	logger logger.Logger,
	sha256Summer sha256Summer,
	md5Summer md5Summer,
	metadata metadata.Metadata,
	sourcesDir,
	productSlug string,
	asyncTimeout time.Duration,
	pollFrequency time.Duration,
) ReleaseUploader {
	return ReleaseUploader{
		s3:            s3,
		pivnet:        pivnet,
		logger:        logger,
		sha256Summer:  sha256Summer,
		md5Summer:     md5Summer,
		metadata:      metadata,
		sourcesDir:    sourcesDir,
		productSlug:   productSlug,
		asyncTimeout:  asyncTimeout,
		pollFrequency: pollFrequency,
	}
}

func (u ReleaseUploader) Upload(release pivnet.Release, exactGlobs []string) error {
	for _, exactGlob := range exactGlobs {
		fullFilepath := filepath.Join(u.sourcesDir, exactGlob)
		fileContentsSHA256, err := u.sha256Summer.SumFile(fullFilepath)
		if err != nil {
			return err
		}

		fileContentsMD5, err := u.md5Summer.SumFile(fullFilepath)
		if err != nil {
			return err
		}

		u.logger.Info(fmt.Sprintf("uploading to s3: '%s'", exactGlob))

		awsObjectKey, err := u.s3.UploadFile(exactGlob)
		if err != nil {
			return err
		}

		filename := filepath.Base(exactGlob)

		var description string
		var docsURL string
		var systemRequirements []string
		var platforms []string
		var includedFiles []string

		uploadAs := filename
		fileType := "Software"

		for _, f := range u.metadata.ProductFiles {
			if f.File == exactGlob {
				u.logger.Info(fmt.Sprintf(
					"exact glob '%s' matches metadata file: '%s'",
					exactGlob,
					f.File,
				))

				if f.UploadAs != "" {
					u.logger.Info(fmt.Sprintf(
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

				if f.DocsURL != "" {
					docsURL = f.DocsURL
				}

				if len(f.SystemRequirements) > 0 {
					systemRequirements = f.SystemRequirements
				}

				if len(f.Platforms) > 0 {
					platforms = f.Platforms
				}

				if len(f.IncludedFiles) > 0 {
					includedFiles = f.IncludedFiles
				}
			} else {
				u.logger.Info(fmt.Sprintf(
					"exact glob '%s' does not match metadata file: '%s'",
					exactGlob,
					f.File,
				))
			}
		}

		productFiles, err := u.pivnet.ProductFiles(u.productSlug)
		if err != nil {
			return err
		}

		for _, pf := range productFiles {
			if pf.AWSObjectKey == awsObjectKey {
				u.logger.Info(fmt.Sprintf("Deleting existing product file with AWSObjectKey: '%s'", pf.AWSObjectKey))

				_, err = u.pivnet.DeleteProductFile(u.productSlug, pf.ID)
				if err != nil {
					return err
				}

				break
			}
		}

		u.logger.Info(fmt.Sprintf(
			"Creating product file with remote name: '%s'",
			uploadAs,
		))

		productFile, err := u.pivnet.CreateProductFile(pivnet.CreateProductFileConfig{
			ProductSlug:        u.productSlug,
			Name:               uploadAs,
			AWSObjectKey:       awsObjectKey,
			FileVersion:        release.Version,
			SHA256:             fileContentsSHA256,
			MD5:                fileContentsMD5,
			Description:        description,
			FileType:           fileType,
			DocsURL:            docsURL,
			SystemRequirements: systemRequirements,
			Platforms:          platforms,
			IncludedFiles:      includedFiles,
		})
		if err != nil {
			return err
		}

		u.logger.Info(fmt.Sprintf(
			"Adding product file: '%s' with ID: %d",
			uploadAs,
			productFile.ID,
		))

		err = u.pivnet.AddProductFile(u.productSlug, release.ID, productFile.ID)
		if err != nil {
			return err
		}

		err = u.pollForProductFile(productFile)
		if err != nil {
			return err
		}
	}

	return nil
}

func (u ReleaseUploader) pollForProductFile(productFile pivnet.ProductFile) error {
	u.logger.Info(fmt.Sprintf(
		"Polling product file: '%s' for async transfer - will wait up to %v",
		productFile.Name,
		u.asyncTimeout,
	))

	timeoutTimer := time.NewTimer(u.asyncTimeout)
	pollTicker := time.NewTicker(u.pollFrequency)

	for {
		select {
		case <-timeoutTimer.C:
			return fmt.Errorf("timed out")
		case <-pollTicker.C:
			pf, err := u.pivnet.ProductFile(u.productSlug, productFile.ID)
			if err != nil {
				return err
			}

			if pf.FileTransferStatus == "complete" {
				u.logger.Info(fmt.Sprintf(
					"Product file: '%s' async transfer complete",
					productFile.Name,
				))

				timeoutTimer.Stop()
				pollTicker.Stop()

				return nil
			}

			u.logger.Info(fmt.Sprintf(
				"Product file: '%s' async transfer incomplete",
				productFile.Name,
			))
		}
	}
}
