package in

import (
	"bufio"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"

	"github.com/h2non/filetype"
)

var archiveMimetypes = []string{
	"application/x-gzip",
	"application/gzip",
	"application/x-tar",
	"application/zip",
}

func mimetype(r *bufio.Reader) (string, error) {
	bs, err := r.Peek(512)
	if err != nil && err != io.EOF {
		return "", err
	}

	kind, err := filetype.Match(bs)
	if err != nil {
		return "", err
	}

	return kind.MIME.Value, nil
}

func archiveMimetype(filename string) string {
	f, err := os.Open(filename)
	if err != nil {
		return ""
	}
	defer f.Close()

	mime, err := mimetype(bufio.NewReader(f))
	if err != nil {
		return ""
	}

	for i := range archiveMimetypes {
		if strings.HasPrefix(mime, archiveMimetypes[i]) {
			return archiveMimetypes[i]
		}
	}

	return ""
}

func inflate(mime, path, destination string) error {
	var cmd *exec.Cmd

	switch mime {
	case "application/zip":
		cmd = exec.Command("unzip", "-P", "", "-d", destination, path)
		defer os.Remove(path)

	case "application/x-tar":
		cmd = exec.Command("tar", "xf", path, "-C", destination)
		defer os.Remove(path)

	case "application/gzip", "application/x-gzip":
		cmd = exec.Command("gunzip", path)

	default:
		return fmt.Errorf("don't know how to extract %s", mime)
	}

	return cmd.Run()
}

func extractArchive(mime, filename string) error {
	destDir := filepath.Dir(filename)

	err := inflate(mime, filename, destDir)
	if err != nil {
		return fmt.Errorf("failed to extract archive: %s", err)
	}

	if mime == "application/gzip" || mime == "application/x-gzip" {
		fileInfos, err := ioutil.ReadDir(destDir)
		if err != nil {
			return fmt.Errorf("failed to read dir: %s", err)
		}

		if len(fileInfos) != 1 {
			return fmt.Errorf("%d files found after gunzip; expected 1", len(fileInfos))
		}

		filename = filepath.Join(destDir, fileInfos[0].Name())
		mime = archiveMimetype(filename)
		if mime == "application/x-tar" {
			err = inflate(mime, filename, destDir)
			if err != nil {
				return fmt.Errorf("failed to extract archive: %s", err)
			}
		}
	}

	return nil
}
