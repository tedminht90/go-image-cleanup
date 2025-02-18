// internal/interfaces/http/handlers/metrics.go
package handlers

import (
	"go-image-cleanup/internal/domain/metrics"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type RegistryGetter interface {
	GetRegistry() *prometheus.Registry
}

type MetricsHandler struct {
	handler fiber.Handler
	metrics metrics.MetricsCollector
	logger  *zap.Logger
}

func NewMetricsHandler(metricsCollector metrics.MetricsCollector, logger *zap.Logger) *MetricsHandler {
	if logger == nil {
		panic("logger is required for MetricsHandler")
	}
	if metricsCollector == nil {
		panic("metrics collector is required for MetricsHandler")
	}

	// Try to get registry if available
	var registry *prometheus.Registry
	if getter, ok := metricsCollector.(RegistryGetter); ok {
		registry = getter.GetRegistry()
	} else {
		registry = prometheus.DefaultRegisterer.(*prometheus.Registry)
	}

	// Configure prometheus handler
	handlerOpts := promhttp.HandlerOpts{
		ErrorLog:          zap.NewStdLog(logger),
		ErrorHandling:     promhttp.ContinueOnError,
		Registry:          registry,
		EnableOpenMetrics: true,
	}

	return &MetricsHandler{
		handler: adaptor.HTTPHandler(promhttp.HandlerFor(registry, handlerOpts)),
		metrics: metricsCollector,
		logger:  logger,
	}
}

func (h *MetricsHandler) Handle(c *fiber.Ctx) error {
	h.logger.Debug("Starting metrics request handling",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()))

	// Add response headers
	c.Set("Content-Type", "text/plain; version=0.0.4")

	// Execute handler
	if err := h.handler(c); err != nil {
		h.logger.Error("Failed to handle metrics request",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Failed to collect metrics",
			"code":  fiber.StatusInternalServerError,
		})
	}

	h.logger.Debug("Metrics request completed successfully",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()))

	return nil
}
