package s3

import (
	"fmt"
	"io"
	"path/filepath"

	"github.com/concourse/s3-resource"
)

type Client interface {
	Upload(fileGlob string, to string, sourcesDir string) error
}

type client struct {
	accessKeyID     string
	secretAccessKey string
	regionName      string
	bucket          string

	stderr io.Writer

	s3client s3resource.S3Client
}

type NewClientConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	RegionName      string
	Bucket          string

	Stderr io.Writer
}

func NewClient(config NewClientConfig) Client {
	endpoint := ""
	disableSSL := false

	awsConfig := s3resource.NewAwsConfig(
		config.AccessKeyID,
		config.SecretAccessKey,
		config.RegionName,
		endpoint,
		disableSSL,
	)

	s3client := s3resource.NewS3Client(
		config.Stderr,
		awsConfig,
	)

	return &client{
		accessKeyID:     config.AccessKeyID,
		secretAccessKey: config.SecretAccessKey,
		regionName:      config.RegionName,
		bucket:          config.Bucket,
		stderr:          config.Stderr,
		s3client:        s3client,
	}
}

func (c client) Upload(fileGlob string, to string, sourcesDir string) error {
	matches, err := filepath.Glob(filepath.Join(sourcesDir, fileGlob))

	if err != nil {
		return err
	}

	if len(matches) == 0 {
		return fmt.Errorf("no matches found for pattern: %s", fileGlob)
	}

	if len(matches) > 1 {
		return fmt.Errorf("more than one match found for pattern: %s\n%v", fileGlob, matches)
	}

	localPath := matches[0]
	remotePath := filepath.Join(to, filepath.Base(localPath))

	acl := "private"

	fmt.Fprintf(c.stderr, "Uploading %s to s3://%s/%s\n", localPath, c.bucket, remotePath)
	_, err = c.s3client.UploadFile(
		c.bucket,
		remotePath,
		localPath,
		acl,
	)
	if err != nil {
		return err
	}
	fmt.Fprintf(c.stderr, "Successfully uploaded %s to s3://%s/%s\n", localPath, c.bucket, remotePath)

	return nil
}
