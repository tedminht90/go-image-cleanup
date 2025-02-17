package metrics

import "time"

type MetricsCollector interface {
	IncImagesRemoved()
	IncImagesSkipped()
	ObserveCleanupDuration(duration time.Duration)
	SetLastCleanupTime(timestamp time.Time)
	IncCleanupErrors()
}
