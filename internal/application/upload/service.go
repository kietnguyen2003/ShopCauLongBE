package upload

import (
	"context"
	"errors"
	"path/filepath"
	"strings"
)

var (
	ErrFileRequired       = errors.New("file is required")
	ErrUnsupportedFile    = errors.New("unsupported file type")
	ErrFileTooLarge       = errors.New("file size exceeds limit")
	ErrStorageUnavailable = errors.New("storage unavailable")
)

const maxUploadSize = 5 * 1024 * 1024

type Service struct {
	storage ObjectStorage
}

func NewService(storage ObjectStorage) *Service {
	return &Service{storage: storage}
}

func (s *Service) UploadImage(ctx context.Context, req UploadRequest) (*UploadResponse, error) {
	if req.File == nil || req.FileName == "" {
		return nil, ErrFileRequired
	}
	if req.Size > maxUploadSize {
		return nil, ErrFileTooLarge
	}
	if !isAllowedImage(req.FileName, req.ContentType) {
		return nil, ErrUnsupportedFile
	}
	if req.Folder == "" {
		req.Folder = "uploads"
	}

	response, err := s.storage.Upload(ctx, req)
	if err != nil {
		return nil, err
	}

	return response, nil
}

func isAllowedImage(fileName, contentType string) bool {
	switch strings.ToLower(contentType) {
	case "image/jpeg", "image/png", "image/webp", "image/gif":
		return true
	}

	switch strings.ToLower(filepath.Ext(fileName)) {
	case ".jpg", ".jpeg", ".png", ".webp", ".gif":
		return true
	default:
		return false
	}
}
