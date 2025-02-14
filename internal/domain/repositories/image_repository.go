package repositories

import (
	"context"
	"go-image-cleanup/internal/domain/models"
)

type ImageRepository interface {
    // GetAllImages returns all images from the container runtime
    GetAllImages(ctx context.Context) ([]models.Image, error)
    
    // GetUsedImages returns a map of image IDs that are currently in use
    GetUsedImages(ctx context.Context) (map[string]bool, error)
    
    // RemoveImage removes an image by its ID
    RemoveImage(ctx context.Context,imageID string) error
}