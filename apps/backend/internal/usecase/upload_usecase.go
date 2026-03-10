package usecase

import (
	"context"
	"fmt"
	"path/filepath"
	"strings"
	"time"

	"github.com/everyday-studio/ollm/internal/domain"
)

type uploadUseCase struct {
	storageService domain.StorageService
	userRepo       domain.UserRepository
	gameRepo       domain.GameRepository
}

// NewUploadUseCase creates a new upload use case.
func NewUploadUseCase(
	storageService domain.StorageService,
	userRepo domain.UserRepository,
	gameRepo domain.GameRepository,
) domain.UploadUseCase {
	return &uploadUseCase{
		storageService: storageService,
		userRepo:       userRepo,
		gameRepo:       gameRepo,
	}
}

// UploadImage handles the business logic for verifying permissions and uploading an image
func (u *uploadUseCase) UploadImage(ctx context.Context, req *domain.UploadImageRequest) (*domain.UploadResponse, error) {
	if req.FileSize > 5*1024*1024 { // 5MB limit
		return nil, fmt.Errorf("%w: file too large (max 5MB)", domain.ErrInvalidInput)
	}

	updaterID := req.UpdaterID

	// Ownership and permission validation
	switch req.Type {
	case domain.UploadTypeUserProfile:
		if req.RefID != updaterID {
			updater, err := u.userRepo.GetByID(ctx, updaterID)
			if err != nil {
				return nil, domain.ErrUnauthorized
			}
			if updater.Role != domain.RoleAdmin {
				return nil, domain.ErrForbidden
			}
		}
	case domain.UploadTypeGameThumbnail:
		updater, err := u.userRepo.GetByID(ctx, updaterID)
		if err != nil {
			return nil, domain.ErrUnauthorized
		}

		if updater.Role != domain.RoleAdmin {
			// If not an admin, they must be the author of the game
			game, err := u.gameRepo.GetByID(ctx, req.RefID)
			if err != nil {
				return nil, fmt.Errorf("%w: game not found, save the game first", domain.ErrNotFound)
			}
			if game.AuthorID != updaterID {
				return nil, domain.ErrForbidden
			}
		}
	default:
		return nil, fmt.Errorf("%w: unsupported upload type '%s'", domain.ErrInvalidInput, req.Type)
	}

	// Extension validation
	ext := strings.ToLower(filepath.Ext(req.Filename))
	if ext != ".png" && ext != ".jpg" && ext != ".jpeg" && ext != ".webp" && ext != ".gif" {
		return nil, fmt.Errorf("%w: unsupported file extension %s", domain.ErrInvalidInput, ext)
	}

	// Build exact path requested by user: e.g. "game/{id}.png"
	// Ensure we fix the extension to .png to match original plan (or keep original ext if desired, plan says {id}.png)
	objectName := fmt.Sprintf("%s/%s.png", string(req.Type), req.RefID)

	url, err := u.storageService.UploadImage(ctx, req.File, objectName, req.ContentType)
	if err != nil {
		return nil, fmt.Errorf("%w: failed to upload image: %v", domain.ErrInternal, err)
	}

	return &domain.UploadResponse{
		URL:       url,
		UpdatedAt: time.Now(),
	}, nil
}
