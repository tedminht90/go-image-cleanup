// internal/infrastructure/metrics/prometheus.go
package metrics

import (
    "github.com/prometheus/client_golang/prometheus"
    "github.com/prometheus/client_golang/prometheus/promauto"
)

var (
    ImagesRemoved = promauto.NewCounter(prometheus.CounterOpts{
        Name: "image_cleanup_removed_total",
        Help: "The total number of images removed",
    })

    ImagesSkipped = promauto.NewCounter(prometheus.CounterOpts{
        Name: "image_cleanup_skipped_total",
        Help: "The total number of images skipped",
    })

    CleanupDuration = promauto.NewHistogram(prometheus.HistogramOpts{
        Name:    "image_cleanup_duration_seconds",
        Help:    "Time spent running image cleanup",
        Buckets: prometheus.DefBuckets,
    })

    LastCleanupTime = promauto.NewGauge(prometheus.GaugeOpts{
        Name: "image_cleanup_last_run_timestamp",
        Help: "Timestamp of the last cleanup run",
    })

    CleanupErrors = promauto.NewCounter(prometheus.CounterOpts{
        Name: "image_cleanup_errors_total",
        Help: "The total number of cleanup errors",
    })
)

