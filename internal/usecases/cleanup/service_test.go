package cleanup

import (
	"context"
	"go-image-cleanup/internal/domain/models"
	"testing"
	"time"

	"go.uber.org/zap"
)

// Mock repository
type mockImageRepository struct {
    images     []models.Image
    usedImages map[string]bool
    removeErr  error
}

func (m *mockImageRepository) GetAllImages(ctx context.Context) ([]models.Image, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        return m.images, nil
    }
}

func (m *mockImageRepository) GetUsedImages(ctx context.Context) (map[string]bool, error) {
    select {
    case <-ctx.Done():
        return nil, ctx.Err()
    default:
        return m.usedImages, nil
    }
}

func (m *mockImageRepository) RemoveImage(ctx context.Context, imageID string) error {
    select {
    case <-ctx.Done():
        return ctx.Err()
    default:
        return m.removeErr
    }
}

// Mock notifier
type mockNotifier struct {
    messages []string
}

func (m *mockNotifier) SendNotification(message string) error {
    m.messages = append(m.messages, message)
    return nil
}

func TestCleanupService(t *testing.T) {
    // Setup logger
    logger, _ := zap.NewDevelopment()

    tests := []struct {
        name           string
        images        []models.Image
        usedImages    map[string]bool
        removeErr     error
        wantRemoved   int
        wantNotified  bool
        timeout       time.Duration
        cancelContext bool
        sleepBefore   time.Duration // Add sleep to test timeout scenarios
    }{
        {
            name: "successful cleanup",
            images: []models.Image{
                {ID: "1", Tags: []string{"tag1"}},
                {ID: "2", Tags: []string{"tag2"}},
            },
            usedImages: map[string]bool{"1": true},
            removeErr:  nil,
            wantRemoved: 1,
            wantNotified: true,
            timeout: 5 * time.Second,
            cancelContext: false,
        },
        {
            name: "cleanup with context cancellation",
            images: []models.Image{
                {ID: "1", Tags: []string{"tag1"}},
                {ID: "2", Tags: []string{"tag2"}},
            },
            usedImages: map[string]bool{},
            removeErr:  nil,
            wantRemoved: 0,
            wantNotified: false,
            timeout: 5 * time.Second,
            cancelContext: true,
        },
        {
            name: "cleanup with timeout",
            images: []models.Image{
                {ID: "1", Tags: []string{"tag1"}},
                {ID: "2", Tags: []string{"tag2"}},
            },
            usedImages: map[string]bool{},
            removeErr:  nil,
            wantRemoved: 0,
            wantNotified: false,
            timeout: 100 * time.Millisecond,
            sleepBefore: 200 * time.Millisecond, // Sleep longer than timeout
            cancelContext: false,
        },
    }

    for _, tt := range tests {
        t.Run(tt.name, func(t *testing.T) {
            // Setup context
            ctx, cancel := context.WithTimeout(context.Background(), tt.timeout)
            defer cancel()

            // If test requires context cancellation
            if tt.cancelContext {
                cancel()
            }

            // Setup mocks
            repo := &mockImageRepository{
                images:     tt.images,
                usedImages: tt.usedImages,
                removeErr:  tt.removeErr,
            }
            notifier := &mockNotifier{}

            // Create service
            service := NewCleanupService(repo, notifier, logger)

            // If test requires sleep before cleanup
            if tt.sleepBefore > 0 {
                time.Sleep(tt.sleepBefore)
            }

            // Run cleanup
            err := service.Cleanup(ctx)

            // Check error based on test case
            if tt.cancelContext && err == nil {
                t.Error("expected error due to cancelled context, got nil")
            }
            if tt.sleepBefore > 0 && err == nil {
                t.Error("expected error due to timeout, got nil")
            }
            if !tt.cancelContext && tt.sleepBefore == 0 && tt.removeErr == nil && err != nil {
                t.Errorf("unexpected error: %v", err)
            }

            // Check notifications
            if tt.wantNotified && len(notifier.messages) == 0 {
                t.Error("expected notification, but none was sent")
            }
            if !tt.wantNotified && len(notifier.messages) > 0 {
                t.Error("unexpected notification was sent")
            }
        })
    }
}