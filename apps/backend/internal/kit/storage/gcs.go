package storage

import (
	"context"
	"fmt"
	"io"
	"time"

	"cloud.google.com/go/storage"

	"github.com/everyday-studio/ollm/internal/domain"
)

type gcsStorageService struct {
	client     *storage.Client
	bucketName string
}

// NewGCSStorageService creates a new Google Cloud Storage service.
func NewGCSStorageService(ctx context.Context, bucketName string) (domain.StorageService, error) {
	client, err := storage.NewClient(ctx)
	if err != nil {
		return nil, fmt.Errorf("failed to create gcs client: %w", err)
	}

	return &gcsStorageService{
		client:     client,
		bucketName: bucketName,
	}, nil
}

// UploadImage uploads an image to Google Cloud Storage.
func (s *gcsStorageService) UploadImage(ctx context.Context, file io.Reader, objectName, contentType string) (string, error) {
	ctx, cancel := context.WithTimeout(ctx, time.Second*50)
	defer cancel()

	wc := s.client.Bucket(s.bucketName).Object(objectName).NewWriter(ctx)
	wc.ContentType = contentType
	wc.CacheControl = "public, max-age=31536000" // 1 year cache

	if _, err := io.Copy(wc, file); err != nil {
		return "", fmt.Errorf("io.Copy: %v", err)
	}
	if err := wc.Close(); err != nil {
		return "", fmt.Errorf("Writer.Close: %v", err)
	}

	// Return public URL based on the bucket name
	return fmt.Sprintf("https://storage.googleapis.com/%s/%s", s.bucketName, objectName), nil
}
