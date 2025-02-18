package metrics

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	"go.uber.org/zap"
)

type PrometheusMetrics struct {
	ImagesRemoved   *prometheus.CounterVec
	ImagesSkipped   *prometheus.CounterVec
	CleanupDuration *prometheus.HistogramVec
	LastCleanupTime *prometheus.GaugeVec
	CleanupErrors   *prometheus.CounterVec
	hostname        string
	logger          *zap.Logger
}

func NewPrometheusMetrics(logger *zap.Logger) *PrometheusMetrics {
	// Get hostname for labels
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
		logger.Error("Failed to get hostname", zap.Error(err))
	}

	metrics := &PrometheusMetrics{
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
		hostname: hostname,
		logger:   logger,
	}

	logger.Info("Prometheus metrics initialized",
		zap.String("hostname", hostname))

	return metrics
}
