package repositories

import (
	"context"
	"time"
)

// CleanupResult định nghĩa kết quả của một lần cleanup
type CleanupResult struct {
	ID         string        `json:"id"`
	HostInfo   string        `json:"host_info"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	TotalCount int           `json:"total_count"`
	Removed    int           `json:"removed"`
	Skipped    int           `json:"skipped"`
	CreatedAt  time.Time     `json:"created_at"`
}

// CleanupResultRepository định nghĩa interface cho việc lưu trữ kết quả cleanup
type CleanupResultRepository interface {
	// SaveResult lưu kết quả của một lần cleanup
	SaveResult(ctx context.Context, result CleanupResult) error

	// GetLatestResult lấy kết quả của lần cleanup gần nhất
	GetLatestResult(ctx context.Context) (*CleanupResult, error)

	// GetResultByID lấy kết quả cleanup theo ID
	GetResultByID(ctx context.Context, id string) (*CleanupResult, error)

	// GetResults lấy danh sách kết quả cleanup, có phân trang
	GetResults(ctx context.Context, limit, offset int) ([]CleanupResult, error)
}
