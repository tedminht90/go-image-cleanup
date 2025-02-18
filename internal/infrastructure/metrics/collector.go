package metrics

import (
	"time"

	"go.uber.org/zap"
)

func (p *PrometheusMetrics) IncImagesRemoved() {
	p.ImagesRemoved.WithLabelValues(p.hostname).Inc()
	p.logger.Debug("Images removed metric incremented",
		zap.String("metric", "image_cleanup_removed_total"),
		zap.String("hostname", p.hostname))
}

func (p *PrometheusMetrics) IncImagesSkipped() {
	p.ImagesSkipped.WithLabelValues(p.hostname).Inc()
	p.logger.Debug("Images skipped metric incremented",
		zap.String("metric", "image_cleanup_skipped_total"),
		zap.String("hostname", p.hostname))
}

func (p *PrometheusMetrics) ObserveCleanupDuration(duration time.Duration) {
	p.CleanupDuration.WithLabelValues(p.hostname).Observe(duration.Seconds())
	p.logger.Debug("Cleanup duration observed",
		zap.String("metric", "image_cleanup_duration_seconds"),
		zap.String("hostname", p.hostname),
		zap.Float64("duration_seconds", duration.Seconds()))
}

func (p *PrometheusMetrics) SetLastCleanupTime(timestamp time.Time) {
	p.LastCleanupTime.WithLabelValues(p.hostname).Set(float64(timestamp.Unix()))
	p.logger.Debug("Last cleanup time set",
		zap.String("metric", "image_cleanup_last_run_timestamp"),
		zap.String("hostname", p.hostname),
		zap.Time("timestamp", timestamp))
}

func (p *PrometheusMetrics) IncCleanupErrors() {
	p.CleanupErrors.WithLabelValues(p.hostname).Inc()
	p.logger.Debug("Cleanup errors metric incremented",
		zap.String("metric", "image_cleanup_errors_total"),
		zap.String("hostname", p.hostname))
}
