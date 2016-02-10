package md5

import (
	"crypto/md5"
	"fmt"
	"io/ioutil"
)

//go:generate counterfeiter . Summer

type Summer interface {
	Sum() (string, error)
}

type fileContentsSummer struct {
	filepath string
}

func NewFileContentsSummer(filepath string) Summer {
	return &fileContentsSummer{
		filepath: filepath,
	}
}

func (f fileContentsSummer) Sum() (string, error) {
	fileContents, err := ioutil.ReadFile(f.filepath)
	if err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", md5.Sum(fileContents)), nil
}
