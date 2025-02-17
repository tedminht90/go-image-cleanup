package main

import (
	"context"
	"fmt"
	"os"
	"os/signal"
	"syscall"

	"go-image-cleanup/config"
	"go-image-cleanup/internal/infrastructure/container"
	loggerPkg "go-image-cleanup/internal/infrastructure/logger"
	"go-image-cleanup/internal/infrastructure/metrics"
	"go-image-cleanup/internal/infrastructure/notification"
	"go-image-cleanup/internal/interfaces/http/handlers"
	"go-image-cleanup/internal/interfaces/http/router"
	"go-image-cleanup/internal/usecases/cleanup"
	"go-image-cleanup/pkg/constants"
	"go-image-cleanup/pkg/helper"

	"github.com/robfig/cron/v3"
	"go.uber.org/zap"
)

// Version and BuildTime are set during build
var (
	Version   = "dev"
	BuildTime = "unknown"
)

func main() {
	// Print version info
	fmt.Printf("Image Cleanup Service %s (built at %s)\n", Version, BuildTime)

	// Load and validate configuration
	cfg, err := config.LoadConfig()
	if err != nil {
		fmt.Printf("Failed to load config: %v\n", err)
		os.Exit(1)
	}

	// Initialize logger
	log, err := loggerPkg.NewLogger(cfg.Logger)
	if err != nil {
		fmt.Printf("Failed to initialize logger: %v\n", err)
		os.Exit(1)
	}
	defer log.Sync()

	// Log startup information and configuration
	logStartupInfo(log, cfg, Version, BuildTime)

	// Initialize infrastructure dependencies
	repo := container.NewCrictlRepository(log)
	notifier := notification.NewTelegramNotifier(cfg.TelegramBotToken, cfg.TelegramChatID, log)
	metricsCollector := metrics.NewPrometheusMetrics()

	// Initialize services
	cleanupService := cleanup.NewCleanupService(repo, notifier, metricsCollector, log)

	// Initialize handlers
	handlers := initializeHandlers(log, Version, BuildTime)

	// Setup router and HTTP server
	app := router.NewFiberApp(log)
	router.SetupRoutes(app, handlers, log)

	// Initialize cleanup job context
	cleanupCtx, cleanupCancel := context.WithCancel(context.Background())
	defer cleanupCancel()

	// Setup and start cron jobs
	cronScheduler := setupCronJobs(cleanupCtx, cleanupService, cfg.CleanupSchedule, log)
	cronScheduler.Start()
	defer cronScheduler.Stop()

	// Start server and handle shutdown
	serverErrChan := startServer(app, cfg.HTTPPort, log)
	handleGracefulShutdown(app, serverErrChan, cleanupCancel, log)
}

func logStartupInfo(log *zap.Logger, cfg *config.Config, version, buildTime string) {
	log.Info("Starting Image Cleanup Service",
		zap.String("version", version),
		zap.String("buildTime", buildTime))

	log.Info("Configuration loaded",
		zap.String("telegram_bot_token", helper.MaskValue(cfg.TelegramBotToken)),
		zap.String("telegram_chat_id", helper.MaskValue(cfg.TelegramChatID)),
		zap.String("cleanup_schedule", cfg.CleanupSchedule),
		zap.String("http_port", cfg.HTTPPort))

	log.Info("Logger configuration",
		zap.String("log_level", cfg.Logger.Level),
		zap.String("log_dir", cfg.Logger.LogDir),
		zap.Int("log_max_size", cfg.Logger.MaxSize),
		zap.Int("log_max_backups", cfg.Logger.MaxBackups),
		zap.Int("log_max_age", cfg.Logger.MaxAge),
		zap.Bool("log_compress", cfg.Logger.Compress))
}

func initializeHandlers(log *zap.Logger, version, buildTime string) *handlers.Handlers {
	return handlers.NewHandlers(log, version, buildTime)
}

func setupCronJobs(ctx context.Context, cleanupService *cleanup.CleanupService, schedule string, log *zap.Logger) *cron.Cron {
	c := cron.New(cron.WithChain(
		cron.SkipIfStillRunning(cron.DefaultLogger),
		cron.Recover(cron.DefaultLogger),
	))

	_, err := c.AddFunc(schedule, func() {
		jobCtx, cancel := context.WithTimeout(ctx, constants.CleanupTimeout)
		defer cancel()

		if err := cleanupService.Cleanup(jobCtx); err != nil {
			log.Error("Cleanup job failed",
				zap.Error(err),
				zap.String("schedule", schedule))
		}
	})
	if err != nil {
		log.Fatal("Failed to schedule cleanup job", zap.Error(err))
	}

	log.Info("Cron scheduler started", zap.String("schedule", schedule))
	return c
}

func startServer(app *router.FiberApp, port string, log *zap.Logger) chan error {
	serverErr := make(chan error, 1)
	go func() {
		log.Info("Starting HTTP server", zap.String("port", port))
		if err := app.Listen(":" + port); err != nil {
			log.Error("Server error", zap.Error(err))
			serverErr <- err
		}
	}()
	return serverErr
}

func handleGracefulShutdown(app *router.FiberApp, serverErr chan error, cleanupCancel context.CancelFunc, log *zap.Logger) {
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, os.Interrupt, syscall.SIGTERM)

	select {
	case <-quit:
		log.Info("Shutting down server...")
	case err := <-serverErr:
		log.Error("Server error occurred", zap.Error(err))
	}

	log.Info("Stopping cleanup jobs...")
	cleanupCancel()

	shutdownCtx, shutdownCancel := context.WithTimeout(context.Background(), constants.ShutdownTimeout)
	defer shutdownCancel()

	if err := app.ShutdownWithContext(shutdownCtx); err != nil {
		log.Error("Server forced to shutdown", zap.Error(err))
	}

	log.Info("Service stopped successfully")
}
