package handlers

import "go.uber.org/zap"

type Handlers struct {
	Health  *HealthHandler
	Version *VersionHandler
	Metrics *MetricsHandler
	logger  *zap.Logger
}

func NewHandlers(logger *zap.Logger, version, buildTime string) *Handlers {
	return &Handlers{
		Health:  NewHealthHandler(logger),
		Version: NewVersionHandler(version, buildTime),
		Metrics: NewMetricsHandler(),
		logger:  logger,
	}
}
