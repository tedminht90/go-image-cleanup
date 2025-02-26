package handlers

import (
	"context"
	"go-image-cleanup/internal/usecases/cleanup"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type CleanupHandler struct {
	cleanupUseCase cleanup.CleanupUseCase
	logger         *zap.Logger
}

func NewCleanupHandler(cleanupUseCase cleanup.CleanupUseCase, logger *zap.Logger) *CleanupHandler {
	return &CleanupHandler{
		cleanupUseCase: cleanupUseCase,
		logger:         logger,
	}
}

// GetCleanupStatus trả về thông tin về lần cleanup gần nhất
func (h *CleanupHandler) GetCleanupStatus(c *fiber.Ctx) error {
	h.logger.Info("Get cleanup status API endpoint called",
		zap.String("ip", c.IP()),
		zap.String("method", c.Method()))

	stats, err := h.cleanupUseCase.GetLastCleanupStats()
	if err != nil {
		h.logger.Error("Failed to get cleanup stats", zap.Error(err))
		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"status":  "error",
			"message": "Failed to retrieve cleanup status",
			"error":   err.Error(),
		})
	}

	return c.Status(fiber.StatusOK).JSON(fiber.Map{
		"status":        "success",
		"host_info":     stats.HostInfo,
		"start_time":    stats.StartTime.Format(time.RFC3339),
		"end_time":      stats.EndTime.Format(time.RFC3339),
		"duration":      stats.Duration.String(),
		"total_count":   stats.TotalCount,
		"removed_count": stats.Removed,
		"skipped_count": stats.Skipped,
	})
}

// TriggerCleanup handles API requests to start the cleanup process
func (h *CleanupHandler) TriggerCleanup(c *fiber.Ctx) error {
	h.logger.Info("Cleanup API endpoint called",
		zap.String("ip", c.IP()),
		zap.String("method", c.Method()))

	// Create a context with timeout
	ctx, cancel := context.WithTimeout(c.Context(), 5*time.Minute)

	// Start cleanup in a goroutine to avoid blocking the API response
	go func() {
		defer cancel()

		if err := h.cleanupUseCase.Cleanup(ctx); err != nil {
			h.logger.Error("API-triggered cleanup failed", zap.Error(err))
		}
	}()

	return c.Status(fiber.StatusAccepted).JSON(fiber.Map{
		"status":  "accepted",
		"message": "Cleanup job has been triggered",
		"time":    time.Now().Format(time.RFC3339),
	})
}
