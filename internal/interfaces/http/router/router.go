package router

import (
	"go-image-cleanup/internal/interfaces/http/handlers"
	"go-image-cleanup/internal/interfaces/http/middleware"
	"go-image-cleanup/pkg/constants"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

type FiberApp struct {
	*fiber.App
}

func NewFiberApp(logger *zap.Logger) *FiberApp {
	app := fiber.New(fiber.Config{
		AppName:               constants.ServiceName,
		DisableStartupMessage: true,
		IdleTimeout:           5 * time.Second,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Ignore favicon.ico errors
			if c.Path() == "/favicon.ico" {
				return nil
			}
			logger.Error("Request failed",
				zap.Error(err),
				zap.String("url", c.Path()),
				zap.String("method", c.Method()))
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
	})

	return &FiberApp{app}
}

func SetupRoutes(app *FiberApp, handlers *handlers.Handlers, logger *zap.Logger) {
	// Add middleware
	app.Use(middleware.Recovery(logger))
	app.Use(middleware.Logger(logger))

	// Handle favicon.ico
	app.Get("/favicon.ico", func(c *fiber.Ctx) error {
		return c.SendStatus(204) // No Content
	})

	// Health routes
	app.Get("/health", handlers.Health.Status)

	// Metrics routes
	app.Get("/metrics", handlers.Metrics.Handle)

	// Version routes
	app.Get("/version", handlers.Version.GetVersion)

	// Add API prefix for future endpoints
	api := app.Group("/api/v1")
	setupAPIRoutes(api, handlers)
}

func setupAPIRoutes(router fiber.Router, handlers *handlers.Handlers) {
	// Future API endpoints will go here
}
