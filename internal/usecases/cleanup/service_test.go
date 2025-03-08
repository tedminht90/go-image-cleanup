package cleanup

import (
	"context"
	"fmt"
	"go-image-cleanup/internal/domain/models"
	"go-image-cleanup/internal/domain/repositories"
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

// Mock cleanup result repository
type mockCleanupResultRepository struct {
	savedResults []repositories.CleanupResult
}

func (m *mockCleanupResultRepository) SaveResult(ctx context.Context, result repositories.CleanupResult) error {
	m.savedResults = append(m.savedResults, result)
	return nil
}

func (m *mockCleanupResultRepository) GetLatestResult(ctx context.Context) (*repositories.CleanupResult, error) {
	if len(m.savedResults) == 0 {
		return nil, fmt.Errorf("no cleanup results found")
	}
	// Return the most recent result
	result := m.savedResults[len(m.savedResults)-1]
	return &result, nil
}

func (m *mockCleanupResultRepository) GetResultByID(ctx context.Context, id string) (*repositories.CleanupResult, error) {
	for _, result := range m.savedResults {
		if result.ID == id {
			return &result, nil
		}
	}
	return nil, fmt.Errorf("result with ID %s not found", id)
}

func (m *mockCleanupResultRepository) GetResults(ctx context.Context, limit, offset int) ([]repositories.CleanupResult, error) {
	total := len(m.savedResults)
	if offset >= total {
		return []repositories.CleanupResult{}, nil
	}

	end := offset + limit
	if end > total {
		end = total
	}

	return m.savedResults[offset:end], nil
}

// Mock notifier
type mockNotifier struct {
	messages []string
}

func (m *mockNotifier) SendNotification(message string) error {
	m.messages = append(m.messages, message)
	return nil
}

// Mock metrics collector
type mockMetricsCollector struct {
	imagesRemoved   int
	imagesSkipped   int
	cleanupErrors   int
	lastCleanupTime time.Time
	cleanupDuration time.Duration
	httpRequests    map[string]int // track requests by path
	httpTimeouts    map[string]int // track timeouts by path
	httpErrors      map[string]int // track errors by path
}

func (m *mockMetricsCollector) IncImagesRemoved() {
	m.imagesRemoved++
}

func (m *mockMetricsCollector) IncImagesSkipped() {
	m.imagesSkipped++
}

func (m *mockMetricsCollector) ObserveCleanupDuration(duration time.Duration) {
	m.cleanupDuration = duration
}

func (m *mockMetricsCollector) SetLastCleanupTime(timestamp time.Time) {
	m.lastCleanupTime = timestamp
}

func (m *mockMetricsCollector) IncCleanupErrors() {
	m.cleanupErrors++
}

// New methods to implement HTTP metrics
func (m *mockMetricsCollector) IncHttpRequests(path, method string, status int) {
	if m.httpRequests == nil {
		m.httpRequests = make(map[string]int)
	}
	key := fmt.Sprintf("%s-%s-%d", path, method, status)
	m.httpRequests[key]++
}

func (m *mockMetricsCollector) IncHttpTimeout(path, method string) {
	if m.httpTimeouts == nil {
		m.httpTimeouts = make(map[string]int)
	}
	key := fmt.Sprintf("%s-%s", path, method)
	m.httpTimeouts[key]++
}

func (m *mockMetricsCollector) IncHttpError(path, method string, status int, errorType string) {
	if m.httpErrors == nil {
		m.httpErrors = make(map[string]int)
	}
	key := fmt.Sprintf("%s-%s-%d-%s", path, method, status, errorType)
	m.httpErrors[key]++
}

func TestCleanupService(t *testing.T) {
	// Setup logger
	logger, _ := zap.NewDevelopment()

	tests := []struct {
		name          string
		images        []models.Image
		usedImages    map[string]bool
		removeErr     error
		wantRemoved   int
		wantSkipped   int
		wantErrors    int
		wantNotified  bool
		wantSaved     bool // Đã lưu kết quả vào repository chưa
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
			usedImages:    map[string]bool{"1": true},
			removeErr:     nil,
			wantRemoved:   1,
			wantSkipped:   1,
			wantErrors:    0,
			wantNotified:  true,
			wantSaved:     true,
			timeout:       5 * time.Second,
			cancelContext: false,
		},
		{
			name: "cleanup with context cancellation",
			images: []models.Image{
				{ID: "1", Tags: []string{"tag1"}},
				{ID: "2", Tags: []string{"tag2"}},
			},
			usedImages:    map[string]bool{},
			removeErr:     nil,
			wantRemoved:   0,
			wantNotified:  false,
			wantSaved:     false,
			timeout:       5 * time.Second,
			cancelContext: true,
		},
		{
			name: "cleanup with timeout",
			images: []models.Image{
				{ID: "1", Tags: []string{"tag1"}},
				{ID: "2", Tags: []string{"tag2"}},
			},
			usedImages:    map[string]bool{},
			removeErr:     nil,
			wantRemoved:   0,
			wantNotified:  false,
			wantSaved:     false,
			timeout:       100 * time.Millisecond,
			sleepBefore:   200 * time.Millisecond, // Sleep longer than timeout
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
			metrics := &mockMetricsCollector{}
			resultRepo := &mockCleanupResultRepository{} // Thêm mock repository mới

			// Create service
			service := NewCleanupService(repo, resultRepo, notifier, metrics, logger)

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

			// Check metrics
			if metrics.imagesRemoved != tt.wantRemoved {
				t.Errorf("expected %d removed images, got %d", tt.wantRemoved, metrics.imagesRemoved)
			}
			if metrics.imagesSkipped != tt.wantSkipped {
				t.Errorf("expected %d skipped images, got %d", tt.wantSkipped, metrics.imagesSkipped)
			}
			if metrics.cleanupErrors != tt.wantErrors {
				t.Errorf("expected %d errors, got %d", tt.wantErrors, metrics.cleanupErrors)
			}
			if !metrics.lastCleanupTime.IsZero() && metrics.cleanupDuration == 0 {
				t.Error("cleanup duration not set but last cleanup time was set")
			}

			// Check notifications
			if tt.wantNotified && len(notifier.messages) == 0 {
				t.Error("expected notification, but none was sent")
			}
			if !tt.wantNotified && len(notifier.messages) > 0 {
				t.Error("unexpected notification was sent")
			}

			// Check if results were saved to repository
			if tt.wantSaved && len(resultRepo.savedResults) == 0 {
				t.Error("expected cleanup result to be saved to repository, but none was found")
			}
			if !tt.wantSaved && len(resultRepo.savedResults) > 0 {
				t.Error("unexpected cleanup result was saved to repository")
			}
		})
	}
}
