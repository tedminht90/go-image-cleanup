// internal/interfaces/http/handlers/health.go
package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type HealthHandler struct {
	logger *zap.Logger
}

func NewHealthHandler(logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		logger: logger,
	}
}

func (h *HealthHandler) Status(c *fiber.Ctx) error {
	h.logger.Debug("Health check requested",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()))

	return c.JSON(fiber.Map{
		"status": "ok",
	})
}
