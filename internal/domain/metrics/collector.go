package metrics

import "time"

type MetricsCollector interface {
	IncImagesRemoved()
	IncImagesSkipped()
	ObserveCleanupDuration(duration time.Duration)
	SetLastCleanupTime(timestamp time.Time)
	IncCleanupErrors()

	// HTTP metrics
	IncHttpRequests(path, method string, status int)
	IncHttpTimeout(path, method string)
	IncHttpError(path, method string, status int, errorType string)
}
