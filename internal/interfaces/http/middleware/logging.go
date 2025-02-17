package middleware

import (
	"github.com/gofiber/fiber/v2"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"go.uber.org/zap"
)

func Logger(log *zap.Logger) fiber.Handler {
	return fiberLogger.New(fiberLogger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
		Done: func(c *fiber.Ctx, logString []byte) {
			if c.Response().StatusCode() >= 400 {
				log.Warn("HTTP request failed",
					zap.Int("status", c.Response().StatusCode()),
					zap.String("method", c.Method()),
					zap.String("path", c.Path()),
					zap.String("ip", c.IP()))
			}
		},
	})
}
