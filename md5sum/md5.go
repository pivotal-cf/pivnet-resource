package md5sum

import (
	"crypto/md5"
	"fmt"
	"io"
	"os"
)

//go:generate counterfeiter . FileSummer

type FileSummer interface {
	SumFile(filepath string) (string, error)
}

type filesummer struct {
}

func NewFileSummer() FileSummer {
	return &filesummer{}
}

func (f filesummer) SumFile(filepath string) (string, error) {
	fileToSum, err := os.Open(filepath)
	if err != nil {
		return "", err
	}
	defer fileToSum.Close()

	hash := md5.New()
	_, err = io.Copy(hash, fileToSum)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}
