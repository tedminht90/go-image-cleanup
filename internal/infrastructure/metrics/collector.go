package metrics

import "time"

// Implement interface for Prometheus
func (p *PrometheusMetrics) IncImagesRemoved() {
	p.ImagesRemoved.Inc()
}

func (p *PrometheusMetrics) IncImagesSkipped() {
	p.ImagesSkipped.Inc()
}

func (p *PrometheusMetrics) ObserveCleanupDuration(duration time.Duration) {
	p.CleanupDuration.Observe(duration.Seconds())
}

func (p *PrometheusMetrics) SetLastCleanupTime(timestamp time.Time) {
	p.LastCleanupTime.Set(float64(timestamp.Unix()))
}

func (p *PrometheusMetrics) IncCleanupErrors() {
	p.CleanupErrors.Inc()
}
