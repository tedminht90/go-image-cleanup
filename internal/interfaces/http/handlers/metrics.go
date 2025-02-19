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

func NewMetricsHandler(metricsCollector metrics.MetricsCollector, logger *zap.Logger) *MetricsHandler {
	if logger == nil {
		panic("logger is required for MetricsHandler")
	}
	if metricsCollector == nil {
		panic("metrics collector is required for MetricsHandler")
	}

	// Create handler with instrumentation
	handlerOpts := promhttp.HandlerOpts{
		ErrorLog:          zap.NewStdLog(logger),
		ErrorHandling:     promhttp.ContinueOnError,
		EnableOpenMetrics: true,
	}

	// Create an instrumented handler that will automatically track metrics
	instrumentedHandler := promhttp.InstrumentMetricHandler(
		prometheus.DefaultRegisterer,
		promhttp.HandlerFor(prometheus.DefaultGatherer, handlerOpts),
	)

	return &MetricsHandler{
		handler: adaptor.HTTPHandler(instrumentedHandler),
		metrics: metricsCollector,
		logger:  logger,
	}
}

func (h *MetricsHandler) Handle(c *fiber.Ctx) error {
	h.logger.Debug("Handling metrics request",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()))

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
