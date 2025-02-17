package middleware

import (
	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"go.uber.org/zap"
)

func Recovery(log *zap.Logger) fiber.Handler {
	return recover.New(recover.Config{
		EnableStackTrace: true,
		StackTraceHandler: func(c *fiber.Ctx, e interface{}) {
			log.Error("Panic recovered",
				zap.Any("error", e),
				zap.String("url", c.Path()),
				zap.String("method", c.Method()))
			c.Status(500).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
	})
}
