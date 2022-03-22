package consumer

import (
	"io"
	"os"
)

type Uploader struct {
	// Path for file
	Path string
}

func (u *Uploader) NeedsUpload() bool {
	return true
}

func (u *Uploader) UploadData() (string, io.Reader, error) {
	file, err := os.Open(u.Path)
	if err != nil {
		return "", nil, err
	}

	return file.Name(), file, nil
}

func (u *Uploader) SendData() string {
	return ""
}
