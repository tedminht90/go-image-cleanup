package handlers

import (
	"go-image-cleanup/internal/domain/metrics"

	"go.uber.org/zap"
)

type Handlers struct {
	Health  *HealthHandler
	Version *VersionHandler
	Metrics *MetricsHandler
	logger  *zap.Logger
}

func NewHandlers(logger *zap.Logger, version, buildTime string, metrics metrics.MetricsCollector) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(logger),
		Version: NewVersionHandler(version, buildTime),
		Metrics: NewMetricsHandler(metrics, logger),
		logger:  logger,
	}
}
