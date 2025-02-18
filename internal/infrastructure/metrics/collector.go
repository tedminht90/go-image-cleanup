package metrics

import (
	"time"
)

func (p *PrometheusMetrics) IncImagesRemoved() {
	p.ImagesRemoved.WithLabelValues(p.hostname).Inc()
}

func (p *PrometheusMetrics) IncImagesSkipped() {
	p.ImagesSkipped.WithLabelValues(p.hostname).Inc()
}

func (p *PrometheusMetrics) ObserveCleanupDuration(duration time.Duration) {
	p.CleanupDuration.WithLabelValues(p.hostname).Observe(duration.Seconds())
}

func (p *PrometheusMetrics) SetLastCleanupTime(timestamp time.Time) {
	p.LastCleanupTime.WithLabelValues(p.hostname).Set(float64(timestamp.Unix()))
}

func (p *PrometheusMetrics) IncCleanupErrors() {
	p.CleanupErrors.WithLabelValues(p.hostname).Inc()
}
