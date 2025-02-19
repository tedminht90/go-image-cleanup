package metrics

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

type PrometheusMetrics struct {
	registry           *prometheus.Registry
	ImagesRemoved      *prometheus.CounterVec
	ImagesSkipped      *prometheus.CounterVec
	CleanupDuration    *prometheus.HistogramVec
	LastCleanupTime    *prometheus.GaugeVec
	CleanupErrors      *prometheus.CounterVec
	HttpRequestTotal   *prometheus.CounterVec
	HttpRequestTimeout *prometheus.CounterVec
	HttpRequestErrors  *prometheus.CounterVec
	hostname           string
	logger             *zap.Logger
}

func (p *PrometheusMetrics) GetRegistry() *prometheus.Registry {
	return p.registry
}

func NewPrometheusMetrics(logger *zap.Logger) *PrometheusMetrics {
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
		logger.Error("Failed to get hostname", zap.Error(err))
	}

	// Use default registry
	registry := prometheus.DefaultRegisterer.(*prometheus.Registry)

	// Create metrics instance
	metrics := &PrometheusMetrics{
		registry: registry,
		ImagesRemoved: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "image_cleanup",
			Name:      "removed_total",
			Help:      "The total number of images removed",
		}, []string{"hostname"}),

		ImagesSkipped: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "image_cleanup",
			Name:      "skipped_total",
			Help:      "The total number of images skipped",
		}, []string{"hostname"}),

		CleanupDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Namespace: "image_cleanup",
			Name:      "duration_seconds",
			Help:      "Time spent running image cleanup",
			Buckets:   prometheus.DefBuckets,
		}, []string{"hostname"}),

		LastCleanupTime: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Namespace: "image_cleanup",
			Name:      "last_run_timestamp",
			Help:      "Timestamp of the last cleanup run",
		}, []string{"hostname"}),

		CleanupErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "image_cleanup",
			Name:      "errors_total",
			Help:      "The total number of cleanup errors",
		}, []string{"hostname"}),

		HttpRequestTotal: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "image_cleanup",
			Name:      "http_requests_total",
			Help:      "Total number of HTTP requests",
		}, []string{"hostname", "code", "method", "path"}),

		HttpRequestTimeout: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "image_cleanup",
			Name:      "http_request_timeouts_total",
			Help:      "Total number of HTTP request timeouts",
		}, []string{"hostname", "path", "method"}),

		HttpRequestErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Namespace: "image_cleanup",
			Name:      "http_request_errors_total",
			Help:      "Total number of HTTP request errors",
		}, []string{"hostname", "path", "method", "status", "error_type"}),

		hostname: hostname,
		logger:   logger,
	}

	logger.Info("Prometheus metrics initialized",
		zap.String("hostname", hostname))

	return metrics
}
