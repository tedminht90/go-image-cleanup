package middleware

import (
	"context"
	"time"

	"github.com/gofiber/fiber/v2"
)

func TimeoutMiddleware(timeout time.Duration) fiber.Handler {
	return func(c *fiber.Ctx) error {
		// Create context with timeout
		ctx, cancel := context.WithTimeout(c.Context(), timeout)
		defer cancel()

		// Replace context
		c.SetUserContext(ctx)

		// Handle the request
		err := c.Next()

		// Check if context was canceled
		if ctx.Err() == context.DeadlineExceeded {
			return c.Status(fiber.StatusRequestTimeout).JSON(fiber.Map{
				"error": "Request timeout",
			})
		}

		return err
	}
}
