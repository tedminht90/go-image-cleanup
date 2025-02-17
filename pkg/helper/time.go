// pkg/helper/time.go
package helper

import "time"

// TimeInICT converts a time to ICT timezone
func TimeInICT(t time.Time) time.Time {
	loc, err := time.LoadLocation("Asia/Bangkok")
	if err != nil {
		// Fallback to UTC+7 if timezone data is not available
		return t.UTC().Add(7 * time.Hour)
	}
	return t.In(loc)
}

// FormatICT formats a time in ICT timezone
func FormatICT(t time.Time) string {
	return TimeInICT(t).Format("2006-01-02 15:04:05 ICT")
}
