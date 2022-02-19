package uploader

import (
	"io"
	"io/ioutil"
	"os"
)

type Uploader interface {
	Upload(name string, path string, reader io.Reader) (string, error)
}

func NewLocalUploader() *LocalUploader {
	return &LocalUploader{}
}

type LocalUploader struct{}

func (*LocalUploader) Upload(name string, path string, reader io.Reader) (string, error) {
	fullPath := path + name

	fileContent, err := ioutil.ReadAll(reader)
	if err != nil {
		return "", err
	}

	file, err := os.Create(fullPath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	_, err = file.Write(fileContent)
	if err != nil {
		return "", err
	}

	return fullPath, nil
}
