package constants

import "time"

const (
	ServiceName     = "image-cleanup"
	ConfigPath      = "/etc/image-cleanup/.env"
	DefaultLogPath  = "/var/log/image-cleanup"
	APIVersion      = "v1"
	ShutdownTimeout = 5 * time.Second
	CleanupTimeout  = 30 * time.Minute
)
