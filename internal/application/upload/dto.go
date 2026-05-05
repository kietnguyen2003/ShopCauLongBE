package upload

import "io"

type UploadRequest struct {
	File        io.Reader
	FileName    string
	Size        int64
	ContentType string
	Folder      string
}

type UploadResponse struct {
	Bucket      string
	ObjectKey   string
	URL         string
	ContentType string
	Size        int64
}
