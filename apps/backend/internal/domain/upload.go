package domain

import (
	"context"
	"io"
	"time"
)

// UploadType determines the category of the uploaded image
type UploadType string

const (
	UploadTypeGameThumbnail UploadType = "game"
	UploadTypeUserProfile   UploadType = "user"
)

// UploadImageRequest is the payload from the handler
type UploadImageRequest struct {
	Type        UploadType
	RefID       string
	File        io.Reader
	Filename    string
	ContentType string
	FileSize    int64
	UpdaterID   string // Represents the user_id from the context attempting the upload
}

// UploadResponse represents the response containing the URL of the uploaded image
type UploadResponse struct {
	URL       string    `json:"url"`
	UpdatedAt time.Time `json:"updated_at"`
}

// StorageService defines the contract for interacting with remote storage like GCS
type StorageService interface {
	UploadImage(ctx context.Context, file io.Reader, objectName, contentType string) (string, error)
}

// UploadUseCase defines the business logic for image uploads
type UploadUseCase interface {
	UploadImage(ctx context.Context, req *UploadImageRequest) (*UploadResponse, error)
}
