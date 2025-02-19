package middleware

import (
	"go-image-cleanup/internal/domain/metrics"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

func MetricsMiddleware(metricsCollector metrics.MetricsCollector, logger *zap.Logger) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Store start time
		start := time.Now()

		// Process request
		err := c.Next()

		// Get response status
		status := c.Response().StatusCode()
		path := getPathPattern(c.Path())

		// Record request metrics
		metricsCollector.IncHttpRequests(path, c.Method(), status)

		// Track timeouts
		if status == fiber.StatusRequestTimeout {
			metricsCollector.IncHttpTimeout(path, c.Method())
		}

		// Track errors
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
			metricsCollector.IncHttpError(path, c.Method(), status, errorType)
		}

		logger.Debug("Request processed",
			zap.String("path", path),
			zap.String("method", c.Method()),
			zap.Int("status", status),
			zap.Duration("duration", time.Since(start)))

		return err
	}
}

// getPathPattern returns the pattern of the path for consistent metrics
func getPathPattern(path string) string {
	// Strip trailing slash
	if len(path) > 1 && path[len(path)-1] == '/' {
		path = path[:len(path)-1]
	}

	switch {
	case path == "/":
		return "/"
	case path == "/health":
		return "/health"
	case path == "/metrics":
		return "/metrics"
	case path == "/version":
		return "/version"
	case path == "/favicon.ico":
		return "/favicon.ico"
	case strings.HasPrefix(path, "/api/v1/"):
		// For API endpoints, keep only first part after /api/v1/
		parts := strings.Split(path[7:], "/")
		if len(parts) > 0 {
			return "/api/v1/" + parts[0]
		}
	}

	return "/other"
}
