package services

import (
	"context"
	"fmt"
	"io"
	"net/url"
	"time"

	"github.com/minio/minio-go/v7"
	"github.com/minio/minio-go/v7/pkg/credentials"
	"github.com/dansonserge/DaFileService/config"
)

type MinioService struct {
	client *minio.Client
}

func NewMinioService(cfg *config.Config) (*MinioService, error) {
	client, err := minio.New(cfg.MinioEndpoint, &minio.Options{
		Creds:  credentials.NewStaticV4(cfg.MinioAccessKey, cfg.MinioSecretKey, ""),
		Secure: cfg.MinioUseSSL,
	})
	if err != nil {
		return nil, fmt.Errorf("failed to create storage client: %w", err)
	}

	return &MinioService{client: client}, nil
}

func (s *MinioService) UploadFile(ctx context.Context, bucketName, objectName string, reader io.Reader, size int64, contentType string) error {
	// Ensure bucket exists before jurisdiction synchronization
	exists, err := s.client.BucketExists(ctx, bucketName)
	if err != nil {
		return err
	}
	if !exists {
		err = s.client.MakeBucket(ctx, bucketName, minio.MakeBucketOptions{})
		if err != nil {
			return err
		}
	}

	_, err = s.client.PutObject(ctx, bucketName, objectName, reader, size, minio.PutObjectOptions{
		ContentType: contentType,
	})
	return err
}

func (s *MinioService) DownloadFile(ctx context.Context, bucketName, objectName string) (io.ReadCloser, minio.ObjectInfo, error) {
	obj, err := s.client.GetObject(ctx, bucketName, objectName, minio.GetObjectOptions{})
	if err != nil {
		return nil, minio.ObjectInfo{}, err
	}

	info, err := obj.Stat()
	if err != nil {
		return nil, minio.ObjectInfo{}, err
	}

	return obj, info, nil
}

func (s *MinioService) GetPresignedURL(ctx context.Context, bucketName, objectName string, expires time.Duration) (*url.URL, error) {
	return s.client.PresignedGetObject(ctx, bucketName, objectName, expires, nil)
}

func (s *MinioService) DeleteFile(ctx context.Context, bucketName, objectName string) error {
	return s.client.RemoveObject(ctx, bucketName, objectName, minio.RemoveObjectOptions{})
}

func (s *MinioService) ListBuckets(ctx context.Context) ([]minio.BucketInfo, error) {
	return s.client.ListBuckets(ctx)
}
