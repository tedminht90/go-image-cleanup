package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

type MetricsHandler struct {
	handler fiber.Handler
}

func NewMetricsHandler() *MetricsHandler {
	return &MetricsHandler{
		handler: adaptor.HTTPHandler(promhttp.Handler()),
	}
}

func (h *MetricsHandler) Handle(c *fiber.Ctx) error {
	return h.handler(c)
}
