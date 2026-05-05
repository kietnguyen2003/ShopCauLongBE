package storage

import (
	"context"
	"fmt"
	"path"
	"path/filepath"
	"strings"

	"kafka-order-demo/backend/internal/application/upload"
	"kafka-order-demo/backend/internal/infrastructure/config"

	"github.com/google/uuid"
	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
)

type MinIOStorage struct {
	client    *minio.Client
	bucket    string
	publicURL string
}

func NewMinIOStorage(cfg config.MinIOConfig) (*MinIOStorage, error) {
	client, err := minio.New(cfg.Endpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.AccessKey, cfg.SecretKey, ""),
		Secure: cfg.UseSSL,
	})
	if err != nil {
		return nil, err
	}

	return &MinIOStorage{
		client:    client,
		bucket:    cfg.Bucket,
		publicURL: strings.TrimRight(cfg.PublicURL, "/"),
	}, nil
}

func (s *MinIOStorage) EnsureBucket(ctx context.Context) error {
	exists, err := s.client.BucketExists(ctx, s.bucket)
	if err != nil {
		return err
	}
	if exists {
		return s.setPublicReadPolicy(ctx)
	}

	if err := s.client.MakeBucket(ctx, s.bucket, minio.MakeBucketOptions{}); err != nil {
		return err
	}

	return s.setPublicReadPolicy(ctx)
}

func (s *MinIOStorage) Upload(ctx context.Context, req upload.UploadRequest) (*upload.UploadResponse, error) {
	if err := s.EnsureBucket(ctx); err != nil {
		return nil, err
	}

	objectKey := buildObjectKey(req.Folder, req.FileName)
	_, err := s.client.PutObject(ctx, s.bucket, objectKey, req.File, req.Size, minio.PutObjectOptions{
		ContentType: req.ContentType,
	})
	if err != nil {
		return nil, err
	}

	return &upload.UploadResponse{
		Bucket:      s.bucket,
		ObjectKey:   objectKey,
		URL:         fmt.Sprintf("%s/%s/%s", s.publicURL, s.bucket, objectKey),
		ContentType: req.ContentType,
		Size:        req.Size,
	}, nil
}

func buildObjectKey(folder, fileName string) string {
	cleanFolder := strings.Trim(path.Clean("/"+folder), "/")
	ext := strings.ToLower(filepath.Ext(fileName))
	if ext == "" {
		ext = ".bin"
	}

	return path.Join(cleanFolder, uuid.NewString()+ext)
}

func (s *MinIOStorage) setPublicReadPolicy(ctx context.Context) error {
	policy := fmt.Sprintf(`{
  "Version": "2012-10-17",
  "Statement": [
    {
      "Effect": "Allow",
      "Principal": {"AWS": ["*"]},
      "Action": ["s3:GetObject"],
      "Resource": ["arn:aws:s3:::%s/*"]
    }
  ]
}`, s.bucket)

	return s.client.SetBucketPolicy(ctx, s.bucket, policy)
}
