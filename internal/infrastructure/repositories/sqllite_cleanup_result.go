package repositories

import (
	"context"
	"database/sql"
	"fmt"
	"go-image-cleanup/internal/domain/repositories"
	"os"
	"path/filepath"
	"time"

	"github.com/google/uuid"
	_ "github.com/mattn/go-sqlite3" // SQLite driver
	"go.uber.org/zap"
)

// Đảm bảo SQLiteCleanupResultRepository implement CleanupResultRepository
var _ repositories.CleanupResultRepository = (*SQLiteCleanupResultRepository)(nil)

type SQLiteCleanupResultRepository struct {
	db     *sql.DB
	logger *zap.Logger
}

// NewSQLiteCleanupResultRepository tạo instance mới của repository
func NewSQLiteCleanupResultRepository(dbPath string, logger *zap.Logger) (*SQLiteCleanupResultRepository, error) {
	// Đảm bảo thư mục tồn tại
	dir := filepath.Dir(dbPath)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return nil, fmt.Errorf("failed to create data directory: %w", err)
	}

	// Mở kết nối đến database
	db, err := sql.Open("sqlite3", dbPath)
	if err != nil {
		return nil, fmt.Errorf("failed to open database: %w", err)
	}

	// Kiểm tra kết nối
	if err := db.Ping(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to ping database: %w", err)
	}

	repo := &SQLiteCleanupResultRepository{
		db:     db,
		logger: logger,
	}

	// Khởi tạo schema
	if err := repo.initSchema(); err != nil {
		db.Close()
		return nil, fmt.Errorf("failed to initialize schema: %w", err)
	}

	return repo, nil
}

// Khởi tạo schema database
func (r *SQLiteCleanupResultRepository) initSchema() error {
	_, err := r.db.Exec(`
		CREATE TABLE IF NOT EXISTS cleanup_results (
			id TEXT PRIMARY KEY,
			host_info TEXT NOT NULL,
			start_time TIMESTAMP NOT NULL,
			end_time TIMESTAMP NOT NULL,
			duration_ms INTEGER NOT NULL,
			total_count INTEGER NOT NULL,
			removed INTEGER NOT NULL,
			skipped INTEGER NOT NULL,
			created_at TIMESTAMP NOT NULL
		);
		CREATE INDEX IF NOT EXISTS idx_cleanup_results_start_time ON cleanup_results(start_time);
		CREATE INDEX IF NOT EXISTS idx_cleanup_results_created_at ON cleanup_results(created_at);
	`)
	return err
}

// Close đóng kết nối database
func (r *SQLiteCleanupResultRepository) Close() error {
	return r.db.Close()
}

// SaveResult lưu kết quả cleanup vào database
func (r *SQLiteCleanupResultRepository) SaveResult(ctx context.Context, result repositories.CleanupResult) error {
	// Tạo ID nếu chưa có
	if result.ID == "" {
		result.ID = uuid.New().String()
	}

	// Đặt thời gian tạo nếu chưa có
	if result.CreatedAt.IsZero() {
		result.CreatedAt = time.Now()
	}

	_, err := r.db.ExecContext(ctx, `
		INSERT INTO cleanup_results
		(id, host_info, start_time, end_time, duration_ms, total_count, removed, skipped, created_at)
		VALUES (?, ?, ?, ?, ?, ?, ?, ?, ?)
	`,
		result.ID,
		result.HostInfo,
		result.StartTime,
		result.EndTime,
		result.Duration.Milliseconds(),
		result.TotalCount,
		result.Removed,
		result.Skipped,
		result.CreatedAt,
	)

	if err != nil {
		return fmt.Errorf("failed to save cleanup result: %w", err)
	}

	r.logger.Info("Cleanup result saved to SQLite",
		zap.String("id", result.ID))

	return nil
}

// GetLatestResult lấy kết quả cleanup gần nhất
func (r *SQLiteCleanupResultRepository) GetLatestResult(ctx context.Context) (*repositories.CleanupResult, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, host_info, start_time, end_time, duration_ms, total_count, removed, skipped, created_at
		FROM cleanup_results
		ORDER BY start_time DESC
		LIMIT 1
	`)

	return r.scanResult(row)
}

// GetResultByID lấy kết quả cleanup theo ID
func (r *SQLiteCleanupResultRepository) GetResultByID(ctx context.Context, id string) (*repositories.CleanupResult, error) {
	row := r.db.QueryRowContext(ctx, `
		SELECT id, host_info, start_time, end_time, duration_ms, total_count, removed, skipped, created_at
		FROM cleanup_results
		WHERE id = ?
	`, id)

	return r.scanResult(row)
}

// GetResults lấy danh sách kết quả cleanup có phân trang
func (r *SQLiteCleanupResultRepository) GetResults(ctx context.Context, limit, offset int) ([]repositories.CleanupResult, error) {
	rows, err := r.db.QueryContext(ctx, `
		SELECT id, host_info, start_time, end_time, duration_ms, total_count, removed, skipped, created_at
		FROM cleanup_results
		ORDER BY start_time DESC
		LIMIT ? OFFSET ?
	`, limit, offset)

	if err != nil {
		return nil, fmt.Errorf("failed to query results: %w", err)
	}
	defer rows.Close()

	var results []repositories.CleanupResult
	for rows.Next() {
		var id, hostInfo string
		var startTime, endTime, createdAt time.Time
		var durationMs, totalCount, removed, skipped int64

		err := rows.Scan(&id, &hostInfo, &startTime, &endTime, &durationMs, &totalCount, &removed, &skipped, &createdAt)
		if err != nil {
			return nil, fmt.Errorf("failed to scan row: %w", err)
		}

		results = append(results, repositories.CleanupResult{
			ID:         id,
			HostInfo:   hostInfo,
			StartTime:  startTime,
			EndTime:    endTime,
			Duration:   time.Duration(durationMs) * time.Millisecond,
			TotalCount: int(totalCount),
			Removed:    int(removed),
			Skipped:    int(skipped),
			CreatedAt:  createdAt,
		})
	}

	if err := rows.Err(); err != nil {
		return nil, fmt.Errorf("error iterating rows: %w", err)
	}

	return results, nil
}

// scanResult đọc một kết quả từ sql.Row
func (r *SQLiteCleanupResultRepository) scanResult(row *sql.Row) (*repositories.CleanupResult, error) {
	var id, hostInfo string
	var startTime, endTime, createdAt time.Time
	var durationMs, totalCount, removed, skipped int64

	err := row.Scan(&id, &hostInfo, &startTime, &endTime, &durationMs, &totalCount, &removed, &skipped, &createdAt)
	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("no cleanup results found")
	}
	if err != nil {
		return nil, fmt.Errorf("failed to scan result: %w", err)
	}

	return &repositories.CleanupResult{
		ID:         id,
		HostInfo:   hostInfo,
		StartTime:  startTime,
		EndTime:    endTime,
		Duration:   time.Duration(durationMs) * time.Millisecond,
		TotalCount: int(totalCount),
		Removed:    int(removed),
		Skipped:    int(skipped),
		CreatedAt:  createdAt,
	}, nil
}
