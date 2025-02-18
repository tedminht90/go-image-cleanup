package router

import (
	"errors"
	"go-image-cleanup/internal/domain/metrics"
	"go-image-cleanup/internal/interfaces/http/handlers"
	"go-image-cleanup/internal/interfaces/http/middleware"
	"go-image-cleanup/pkg/constants"
	"net/http"
	"strings"
	"time"

	"github.com/gofiber/fiber/v2"
	"go.uber.org/zap"
)

// ErrorResponse represents a structured error response
type ErrorResponse struct {
	Status  int    `json:"status"`
	Message string `json:"message"`
	Path    string `json:"path"`
}

type FiberApp struct {
	*fiber.App
}

func NewFiberApp(logger *zap.Logger) *FiberApp {
	app := fiber.New(fiber.Config{
		AppName:               constants.ServiceName,
		DisableStartupMessage: true,
		IdleTimeout:           60 * time.Second, // Tăng idle timeout
		ReadTimeout:           30 * time.Second, // Tăng read timeout
		WriteTimeout:          30 * time.Second, // Tăng write timeout
		DisableKeepalive:      false,            // Enable keepalive
		ServerHeader:          constants.ServiceName,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			// Default status code and error message
			status := fiber.StatusInternalServerError
			message := "Internal Server Error"

			// Handle specific error types
			var fiberError *fiber.Error
			if errors.As(err, &fiberError) {
				status = fiberError.Code
				message = fiberError.Message
			} else {
				// Handle other common error types
				switch {
				case errors.Is(err, fiber.ErrRequestTimeout):
					status = fiber.StatusRequestTimeout
					message = "Request Timeout"
				case errors.Is(err, fiber.ErrTooManyRequests):
					status = fiber.StatusTooManyRequests
					message = "Too Many Requests"
				case strings.Contains(err.Error(), "deadline exceeded"):
					status = fiber.StatusGatewayTimeout
					message = "Gateway Timeout"
				case strings.Contains(err.Error(), "broken pipe"):
					status = fiber.StatusBadRequest
					message = "Client Disconnected"
				}
			}

			// Ignore favicon.ico errors
			if c.Path() == "/favicon.ico" {
				return nil
			}

			// Log error with appropriate level based on status code
			logErr := logger.Error
			if status < http.StatusInternalServerError {
				logErr = logger.Warn
			}

			logErr("Request failed",
				zap.Error(err),
				zap.String("url", c.Path()),
				zap.String("method", c.Method()),
				zap.Int("status", status),
				zap.String("ip", c.IP()),
				zap.String("user_agent", string(c.Request().Header.UserAgent())))

			// Return structured error response
			return c.Status(status).JSON(ErrorResponse{
				Status:  status,
				Message: message,
				Path:    c.Path(),
			})
		},
	})

	return &FiberApp{app}
}

func SetupRoutes(app *FiberApp, handlers *handlers.Handlers, metricsCollector metrics.MetricsCollector, logger *zap.Logger) {
	// Add middleware
	app.Use(middleware.Recovery(logger))
	app.Use(middleware.Logger(logger))
	app.Use(middleware.MetricsMiddleware(metricsCollector, logger))

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

	// Add 404 handler
	app.Use(func(c *fiber.Ctx) error {
		return c.Status(fiber.StatusNotFound).JSON(ErrorResponse{
			Status:  fiber.StatusNotFound,
			Message: "Route not found",
			Path:    c.Path(),
		})
	})
}

func setupAPIRoutes(router fiber.Router, handlers *handlers.Handlers) {
	// Future API endpoints will go here
}
