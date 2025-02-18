package handlers

import (
	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type VersionHandler struct {
	version   string
	buildTime string
	logger    *zap.Logger
}

func NewVersionHandler(version, buildTime string) *VersionHandler {
	return &VersionHandler{
		version:   version,
		buildTime: buildTime,
	}
}

func (h *VersionHandler) GetVersion(c *fiber.Ctx) error {
	return c.JSON(fiber.Map{
		"version":   h.version,
		"buildTime": h.buildTime,
		"status":    "ok",
	})
}
