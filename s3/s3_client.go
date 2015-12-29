package s3

import (
	"encoding/json"
	"io"
	"os/exec"

	"github.com/pivotal-cf-experimental/pivnet-resource/logger"
)

type Client interface {
	Upload(fileGlob string, to string, sourcesDir string) error
}

type client struct {
	accessKeyID     string
	secretAccessKey string
	regionName      string
	bucket          string

	logger logger.Logger

	stdout io.Writer
	stderr io.Writer

	outBinaryPath string
}

type NewClientConfig struct {
	AccessKeyID     string
	SecretAccessKey string
	RegionName      string
	Bucket          string

	Logger logger.Logger

	Stdout io.Writer
	Stderr io.Writer

	OutBinaryPath string
}

func NewClient(config NewClientConfig) Client {
	return &client{
		accessKeyID:     config.AccessKeyID,
		secretAccessKey: config.SecretAccessKey,
		regionName:      config.RegionName,
		bucket:          config.Bucket,
		stdout:          config.Stdout,
		stderr:          config.Stderr,
		outBinaryPath:   config.OutBinaryPath,
		logger:          config.Logger,
	}
}

func (c client) Upload(fileGlob string, to string, sourcesDir string) error {
	s3Input := Request{
		Source: Source{
			AccessKeyID:     c.accessKeyID,
			SecretAccessKey: c.secretAccessKey,
			Bucket:          c.bucket,
			RegionName:      c.regionName,
		},
		Params: Params{
			File: fileGlob,
			To:   to,
		},
	}

	c.logger.Debugf("input to s3out: %+v, sourcesDir: %s\n", s3Input, sourcesDir)

	cmd := exec.Command(c.outBinaryPath, sourcesDir)

	cmdIn, err := cmd.StdinPipe()
	if err != nil {
		return err
	}

	cmd.Stdout = c.stderr
	cmd.Stderr = c.stderr

	err = cmd.Start()
	if err != nil {
		return err
	}

	err = json.NewEncoder(cmdIn).Encode(s3Input)
	if err != nil {
		return err
	}

	err = cmd.Wait()
	if err != nil {
		return err
	}

	return nil
}
