package upload

import "context"

type ObjectStorage interface {
	Upload(ctx context.Context, req UploadRequest) (*UploadResponse, error)
}
