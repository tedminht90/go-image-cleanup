package cleanup

import (
	"context"
	"time"
)

// CleanupStats chứa thống kê về quá trình cleanup
type CleanupStats struct {
	HostInfo   string        `json:"host_info"`
	StartTime  time.Time     `json:"start_time"`
	EndTime    time.Time     `json:"end_time"`
	Duration   time.Duration `json:"duration"`
	TotalCount int           `json:"total_count"`
	Removed    int           `json:"removed"`
	Skipped    int           `json:"skipped"`
}

type CleanupUseCase interface {
	Cleanup(ctx context.Context) error

	// GetLastCleanupStats trả về thông tin của lần cleanup gần nhất
	GetLastCleanupStats() (*CleanupStats, error)
}
