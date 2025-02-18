package handlers

import (
	"go-image-cleanup/internal/domain/metrics"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsHandler struct {
	handler fiber.Handler
	metrics metrics.MetricsCollector
	logger  *zap.Logger
}

func NewMetricsHandler(metrics metrics.MetricsCollector, logger *zap.Logger) *MetricsHandler {
	if logger == nil {
		panic("logger is required for MetricsHandler")
	}
	if metrics == nil {
		panic("metrics collector is required for MetricsHandler")
	}

	// Configure prometheus handler with custom options
	handlerOpts := promhttp.HandlerOpts{
		ErrorLog:          zap.NewStdLog(logger),
		ErrorHandling:     promhttp.ContinueOnError,
		Registry:          prometheus.DefaultRegisterer.(*prometheus.Registry),
		EnableOpenMetrics: true,
	}

	return &MetricsHandler{
		handler: adaptor.HTTPHandler(promhttp.HandlerFor(prometheus.DefaultGatherer, handlerOpts)),
		metrics: metrics,
		logger:  logger,
	}
}

func (h *MetricsHandler) Handle(c *fiber.Ctx) error {
	h.logger.Debug("Starting metrics request handling",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()))

	// Error handling wrapper for promhttp.Handler
	handler := func(c *fiber.Ctx) error {
		err := h.handler(c)
		if err != nil {
			h.logger.Error("Prometheus handler error",
				zap.Error(err),
				zap.String("path", c.Path()),
				zap.String("method", c.Method()))

			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"error": "Failed to collect metrics",
				"code":  fiber.StatusInternalServerError,
			})
		}
		return err
	}

	// Add response headers
	c.Set("Content-Type", "text/plain; version=0.0.4")

	// Execute handler with error handling
	if err := handler(c); err != nil {
		h.logger.Error("Failed to handle metrics request",
			zap.Error(err),
			zap.String("path", c.Path()),
			zap.String("method", c.Method()))

		return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
			"error": "Internal Server Error",
			"code":  fiber.StatusInternalServerError,
		})
	}

	h.logger.Debug("Metrics request completed successfully",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()))

	return nil
}
