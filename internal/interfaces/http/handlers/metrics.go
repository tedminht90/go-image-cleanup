package handlers

import (
	"go-image-cleanup/internal/domain/metrics"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
)

type MetricsHandler struct {
	handler fiber.Handler
	metrics metrics.MetricsCollector
	logger  *zap.Logger
}

func NewMetricsHandler(metrics metrics.MetricsCollector, logger *zap.Logger) *MetricsHandler {
	return &MetricsHandler{
		handler: adaptor.HTTPHandler(promhttp.Handler()),
		metrics: metrics,
		logger:  logger,
	}
}

func (h *MetricsHandler) Handle(c *fiber.Ctx) error {
	h.logger.Debug("Handling metrics request",
		zap.String("path", c.Path()),
		zap.String("method", c.Method()))

	return h.handler(c)
}
