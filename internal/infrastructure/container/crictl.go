// internal/infrastructure/container/crictl.go
package container

import (
	"context"
	"encoding/json"
	"fmt"
	"go-image-cleanup/internal/domain/models"
	"os/exec"

	"go.uber.org/zap"
)

type CrictlRepository struct {
    logger *zap.Logger
}

func NewCrictlRepository(logger *zap.Logger) *CrictlRepository {
    return &CrictlRepository{
        logger: logger,
    }
}

func (r *CrictlRepository) executeCommand(ctx context.Context, args ...string) ([]byte, error) {
    cmd := exec.CommandContext(ctx, "crictl", args...)
    output, err := cmd.CombinedOutput()
    if err != nil {
        return nil, fmt.Errorf("command failed: %w, output: %s", err, string(output))
    }
    return output, nil
}

func (r *CrictlRepository) GetAllImages(ctx context.Context) ([]models.Image, error) {
    output, err := r.executeCommand(ctx, "images", "--output=json")
    if err != nil {
        return nil, fmt.Errorf("failed to execute crictl images: %w", err)
    }

    var response struct {
        Images []struct {
            ID       string   `json:"id"`
            RepoTags []string `json:"repoTags"`
        } `json:"images"`
    }

    if err := json.Unmarshal(output, &response); err != nil {
        return nil, fmt.Errorf("failed to parse images output: %w", err)
    }

    var images []models.Image
    for _, img := range response.Images {
        images = append(images, models.Image{
            ID:    img.ID,
            Tags:  img.RepoTags,
        })
    }

    r.logger.Debug("Retrieved all images", zap.Int("count", len(images)))
    return images, nil
}

func (r *CrictlRepository) GetUsedImages(ctx context.Context) (map[string]bool, error) {
    output, err := r.executeCommand(ctx, "ps", "-a", "--output=json")
    if err != nil {
        return nil, fmt.Errorf("failed to execute crictl ps: %w", err)
    }

    var response struct {
        Containers []struct {
            ImageID string `json:"imageRef"`
        } `json:"containers"`
    }

    if err := json.Unmarshal(output, &response); err != nil {
        return nil, fmt.Errorf("failed to parse containers output: %w", err)
    }

    usedImages := make(map[string]bool)
    for _, container := range response.Containers {
        if container.ImageID != "" {
            usedImages[container.ImageID] = true
        }
    }

    r.logger.Debug("Retrieved used images", zap.Int("count", len(usedImages)))
    return usedImages, nil
}

func (r *CrictlRepository) RemoveImage(ctx context.Context, imageID string) error {
    _, err := r.executeCommand(ctx, "rmi", imageID)
    if err != nil {
        return fmt.Errorf("failed to remove image %s: %w", imageID, err)
    }

    r.logger.Debug("Removed image", zap.String("imageID", imageID))
    return nil
}