package metrics

import (
	"strconv"

	"go.uber.org/zap"
)

func (p *PrometheusMetrics) IncHttpRequests(path, method string, status int) {
	code := strconv.Itoa(status)

	p.HttpRequestTotal.WithLabelValues(
		p.hostname,
		code,   // path first
		method, // then method
		path,   // then status code
	).Inc()

	p.logger.Debug("HTTP request metric incremented",
		zap.String("metric", "image_cleanup_http_requests_total"),
		zap.String("hostname", p.hostname),
		zap.String("code", code),
		zap.String("method", method),
		zap.String("path", path))
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
	statusStr := strconv.Itoa(status)
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
