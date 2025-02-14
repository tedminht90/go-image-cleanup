package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"strconv"
	"syscall"
	"time"

	"go-image-cleanup/config"
	"go-image-cleanup/internal/infrastructure/container"
	loggerPkg "go-image-cleanup/internal/infrastructure/logger"
	"go-image-cleanup/internal/infrastructure/notification"
	"go-image-cleanup/internal/interfaces/http/handlers"
	"go-image-cleanup/internal/usecases/cleanup"

	"github.com/gofiber/fiber/v2"
	"github.com/gofiber/fiber/v2/middleware/adaptor"
	fiberLogger "github.com/gofiber/fiber/v2/middleware/logger"
	"github.com/gofiber/fiber/v2/middleware/recover"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Version and BuildTime are set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func validateConfig(cfg *config.Config) error {
	if cfg.TelegramBotToken == "" {
		return fmt.Errorf("TELEGRAM_BOT_TOKEN is required")
	}
	if cfg.TelegramChatID == "" {
		return fmt.Errorf("TELEGRAM_CHAT_ID is required")
	}
	if cfg.CleanupSchedule == "" {
		return fmt.Errorf("CLEANUP_SCHEDULE is required")
	}
	// Validate cron expression
	if _, err := cron.ParseStandard(cfg.CleanupSchedule); err != nil {
		return fmt.Errorf("invalid CLEANUP_SCHEDULE: %v", err)
	}
	// Validate port
	if port, err := strconv.Atoi(cfg.HTTPPort); err != nil || port < 1 || port > 65535 {
		return fmt.Errorf("invalid HTTP_PORT: must be between 1 and 65535")
	}
	return nil
}

func main() {
	// Print version info
	fmt.Printf("Image Cleanup Service %s (built at %s)\n", Version, BuildTime)

	// Load configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Validate configuration
	if err := validateConfig(cfg); err != nil {
		fmt.Printf("Invalid configuration: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := loggerPkg.NewLogger(cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Log startup information
	log.Info("Starting Image Cleanup Service",
		zap.String("version", Version),
		zap.String("buildTime", BuildTime))

	// Initialize dependencies
	repo := container.NewCrictlRepository(log)
	notifier := notification.NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatID, log)
	cleanupService := cleanup.NewCleanupService(repo, notifier, log)
	healthHandler := handlers.NewHealthHandler(log)

	// Context for cleanup jobs
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()

	// Initialize cron job with error handling
	c := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))

	_, err = c.AddFunc(cfg.CleanupSchedule, func() {
		jobCtx, cancel := context.WithTimeout(cleanupCtx, 30*time.Minute)
		defer cancel()

		if err := cleanupService.Cleanup(jobCtx); err != nil {
			log.Error("Cleanup job failed",
				zap.Error(err),
				zap.String("schedule", cfg.CleanupSchedule))
		}
	})
	if err != nil {
		log.Fatal("Failed to schedule cleanup job", zap.Error(err))
	}

	// Start cron scheduler
	c.Start()
	log.Info("Cron scheduler started", zap.String("schedule", cfg.CleanupSchedule))

	// Initialize Fiber app with custom config
	app := fiber.New(fiber.Config{
		AppName:               "Image Cleanup Service",
		DisableStartupMessage: true,
		IdleTimeout:           5 * time.Second,
		ReadTimeout:           10 * time.Second,
		WriteTimeout:          10 * time.Second,
		ErrorHandler: func(c *fiber.Ctx, err error) error {
			log.Error("HTTP error", zap.Error(err))
			return c.Status(500).JSON(fiber.Map{
				"error": "Internal Server Error",
			})
		},
	})

	// Add middleware
	app.Use(recover.New(recover.Config{
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
	}))
	app.Use(fiberLogger.New(fiberLogger.Config{
		Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
	}))

	// Register routes
	app.Get("/health", healthHandler.Status)
	app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))
	app.Get("/version", func(c *fiber.Ctx) error {
		return c.JSON(fiber.Map{
			"version":   Version,
			"buildTime": BuildTime,
		})
	})

	// Graceful shutdown setup
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	// Start server in a goroutine
	serverErr := make(chan error, 1)
	go func() {
		log.Info("Starting HTTP server", zap.String("port", cfg.HTTPPort))
		if err := app.Listen(":" + cfg.HTTPPort); err != nil {
			log.Error("Server error", zap.Error(err))
			serverErr <- err
		}
	}()

	// Wait for interrupt signal or server error
	select {
	case <-quit:
		log.Info("Shutting down server...")
	case err := <-serverErr:
		log.Error("Server error occurred", zap.Error(err))
	}

	// Cancel any running cleanup jobs
	log.Info("Stopping cleanup jobs...")
	cleanupCancel()
	c.Stop()

	// Gracefully shutdown server
	log.Info("Shutting down HTTP server...")
	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Service stopped successfully")
}
