package metrics

import (
	"fmt"

	"go.uber.org/zap"
)

// HTTP request metrics
func (p *PrometheusMetrics) IncHttpRequests(path, method string, status int) {
	statusStr := fmt.Sprintf("%d", status)
	p.HttpRequestTotal.WithLabelValues(
		p.hostname,
		path,
		method,
		statusStr,
	).Inc()
	p.logger.Debug("HTTP request metric incremented",
		zap.String("metric", "image_cleanup_http_requests_total"),
		zap.String("hostname", p.hostname),
		zap.String("path", path),
		zap.String("method", method),
		zap.Int("status", status))
}

func (p *PrometheusMetrics) IncHttpTimeout(path, method string) {
	p.HttpRequestTimeout.WithLabelValues(
		p.hostname,
		path,
		method,
	).Inc()
	p.logger.Debug("HTTP timeout metric incremented",
		zap.String("metric", "image_cleanup_http_request_timeouts_total"),
		zap.String("hostname", p.hostname),
		zap.String("path", path),
		zap.String("method", method))
}

func (p *PrometheusMetrics) IncHttpError(path, method string, status int, errorType string) {
	statusStr := fmt.Sprintf("%d", status)
	p.HttpRequestErrors.WithLabelValues(
		p.hostname,
		path,
		method,
		statusStr,
		errorType,
	).Inc()
	p.logger.Debug("HTTP error metric incremented",
		zap.String("metric", "image_cleanup_http_request_errors_total"),
		zap.String("hostname", p.hostname),
		zap.String("path", path),
		zap.String("method", method),
		zap.Int("status", status),
		zap.String("error_type", errorType))
}
