package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
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

func main() {
    // Load configuration
    cfg, err := config.LoadConfig()
    if err != nil {
        panic(fmt.Sprintf("Failed to load config: %v", err))
    }

    // Initialize logger with rotation config
    log, err := loggerPkg.NewLogger(cfg.Logger)
    if err != nil {
        panic(fmt.Sprintf("Failed to initialize logger: %v", err))
    }
    defer log.Sync()

    // Initialize dependencies
    repo := container.NewCrictlRepository(log)
    notifier := notification.NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatID, log)
    cleanupService := cleanup.NewCleanupService(repo, notifier, log)
    healthHandler := handlers.NewHealthHandler(log)

    // Context for cleanup jobs
    cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
    defer cleanupCancel()

    // Initialize cron job
    c := cron.New(cron.WithChain(
        cron.SkipIfStillRunning(cron.DefaultLogger),
    ))
    
    _, err = c.AddFunc(cfg.CleanupSchedule, func() {
        // Create a new context with timeout for each job run
        jobCtx, cancel := context.WithTimeout(cleanupCtx, 30*time.Minute)
        defer cancel()

        if err := cleanupService.Cleanup(jobCtx); err != nil {
            log.Error("Cleanup job failed", zap.Error(err))
        }
    })
    if err != nil {
        log.Fatal("Failed to schedule cleanup job", zap.Error(err))
    }
    c.Start()

    // Initialize Fiber app
    app := fiber.New(fiber.Config{
        AppName:               "Image Cleanup Service",
        DisableStartupMessage: true,
    })

    // Add middleware
    app.Use(recover.New())
    app.Use(fiberLogger.New(fiberLogger.Config{
        Format: "[${time}] ${status} - ${latency} ${method} ${path}\n",
    }))

    // Register routes
    app.Get("/health", healthHandler.Status)
    app.Get("/metrics", adaptor.HTTPHandler(promhttp.Handler()))

    // Graceful shutdown setup
    quit := make(chan os.Signal, 1)
    signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

    // Start server in a goroutine
    serverErr := make(chan error, 1)
    go func() {
        if err := app.Listen(":" + cfg.HTTPPort); err != nil {
            log.Error("Server error", zap.Error(err))
            serverErr <- err
        }
    }()

    log.Info("Server started",
        zap.String("port", cfg.HTTPPort),
        zap.String("cleanup_schedule", cfg.CleanupSchedule),
        zap.String("log_level", cfg.Logger.Level),
        zap.String("log_dir", cfg.Logger.LogDir))

    // Wait for interrupt signal or server error
    select {
    case <-quit:
        log.Info("Shutting down server...")
    case err := <-serverErr:
        log.Error("Server error occurred", zap.Error(err))
    }

    // Cancel any running cleanup jobs
    cleanupCancel()

    // Gracefully shutdown server
    shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer shutdownCancel()

    if err := app.ShutdownWithContext(shutdownCtx); err != nil {
        log.Error("Server forced to shutdown", zap.Error(err))
    }

    // Stop cron jobs
    c.Stop()

    log.Info("Server exited",
        zap.String("log_dir", cfg.Logger.LogDir))
}