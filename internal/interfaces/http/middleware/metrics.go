// internal/interfaces/http/middleware/metrics.go
package middleware

import (
	"go-image-cleanup/internal/domain/metrics"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func MetricsMiddleware(metricsCollector metrics.MetricsCollector, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		startTime := time.Now()
		path := c.Path()
		method := c.Method()

		// Process request
		err := c.Next()

		// Get response status
		status := c.Response().StatusCode()
		duration := time.Since(startTime)

		// Record metrics
		metricsCollector.IncHttpRequests(path, method, status)

		// Track timeout requests
		if status == fiber.StatusRequestTimeout {
			metricsCollector.IncHttpTimeout(path, method)
		}

		// Track error requests
		if status >= 500 {
			errorType := "server_error"
			switch status {
			case fiber.StatusServiceUnavailable:
				errorType = "service_unavailable"
			case fiber.StatusGatewayTimeout:
				errorType = "gateway_timeout"
			case fiber.StatusInternalServerError:
				errorType = "internal_server_error"
			}
			metricsCollector.IncHttpError(path, method, status, errorType)
		}

		// Log request details
		logger.Debug("Request processed",
			zap.String("path", path),
			zap.String("method", method),
			zap.Int("status", status),
			zap.Duration("duration", duration))

		return err
	}
}
