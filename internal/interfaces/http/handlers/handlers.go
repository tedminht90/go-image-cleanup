package handlers

import (
	"go-image-cleanup/internal/domain/metrics"
	"go-image-cleanup/internal/usecases/cleanup"

	"go.uber.org/zap"
)

type Handlers struct {
	Health  *HealthHandler
	Version *VersionHandler
	Metrics *MetricsHandler
	Cleanup *CleanupHandler
	logger  *zap.Logger
}

func NewHandlers(
	logger *zap.Logger,
	version,
	buildTime string,
	metrics metrics.MetricsCollector,
	cleanupUseCase cleanup.CleanupUseCase,
) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(logger),
		Version: NewVersionHandler(version, buildTime),
		Metrics: NewMetricsHandler(metrics, logger),
		Cleanup: NewCleanupHandler(cleanupUseCase, logger),
		logger:  logger,
	}
}
