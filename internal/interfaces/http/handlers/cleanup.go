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
