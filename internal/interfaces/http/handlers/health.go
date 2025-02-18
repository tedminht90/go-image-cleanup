package handlers

import (
	"runtime"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type HealthHandler struct {
	logger    *zap.Logger
	startTime time.Time
}

func NewHealthHandler(logger *zap.Logger) *HealthHandler {
	return &HealthHandler{
		logger:    logger,
		startTime: time.Now(),
	}
}

func (h *HealthHandler) Status(c *fiber.Ctx) error {
	// Log request details
	h.logger.Debug("Health check requested",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()),
		zap.String("ip", c.IP()),
		zap.String("user_agent", c.Get("User-Agent")))

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	// Collect health information
	healthInfo := fiber.Map{
		"status":    "ok",
		"uptime":    time.Since(h.startTime).String(),
		"timestamp": time.Now().Format(time.RFC3339),
		"system": fiber.Map{
			"goroutines": runtime.NumGoroutine(),
			"memory": fiber.Map{
				"alloc":       m.Alloc,
				"total_alloc": m.TotalAlloc,
				"sys":         m.Sys,
				"num_gc":      m.NumGC,
			},
		},
	}

	// Log successful health check
	h.logger.Info("Health check successful",
		zap.String("ip", c.IP()),
		zap.String("uptime", healthInfo["uptime"].(string)),
		zap.Int("goroutines", healthInfo["system"].(fiber.Map)["goroutines"].(int)))

	return c.JSON(healthInfo)
}
