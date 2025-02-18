package metrics

import (
	"os"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

type PrometheusMetrics struct {
	ImagesRemoved   *prometheus.CounterVec
	ImagesSkipped   *prometheus.CounterVec
	CleanupDuration *prometheus.HistogramVec
	LastCleanupTime *prometheus.GaugeVec
	CleanupErrors   *prometheus.CounterVec
	hostname        string
}

func NewPrometheusMetrics() *PrometheusMetrics {
	// Get hostname for labels
	hostname, err := os.Hostname()
	if err != nil {
		hostname = "unknown"
	}

	return &PrometheusMetrics{
		ImagesRemoved: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "image_cleanup_removed_total",
			Help: "The total number of images removed",
		}, []string{"hostname"}),

		ImagesSkipped: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "image_cleanup_skipped_total",
			Help: "The total number of images skipped",
		}, []string{"hostname"}),

		CleanupDuration: promauto.NewHistogramVec(prometheus.HistogramOpts{
			Name:    "image_cleanup_duration_seconds",
			Help:    "Time spent running image cleanup",
			Buckets: prometheus.DefBuckets,
		}, []string{"hostname"}),

		LastCleanupTime: promauto.NewGaugeVec(prometheus.GaugeOpts{
			Name: "image_cleanup_last_run_timestamp",
			Help: "Timestamp of the last cleanup run",
		}, []string{"hostname"}),

		CleanupErrors: promauto.NewCounterVec(prometheus.CounterOpts{
			Name: "image_cleanup_errors_total",
			Help: "The total number of cleanup errors",
		}, []string{"hostname"}),
		hostname: hostname,
	}
}
