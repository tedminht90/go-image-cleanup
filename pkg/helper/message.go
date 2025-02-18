package helper

import (
	"fmt"
	"time"
)

// FormatCleanupMessage formats the cleanup notification message with emojis
func FormatCleanupMessage(
	hostInfo string,
	startTime time.Time,
	endTime time.Time,
	duration time.Duration,
	total int,
	removed int,
	skipped int,
) string {
	return fmt.Sprintf(`🔄 Image cleanup completed on:
%s

⏱ Time Information:
Started: %s
Finished: %s
Duration: %s

📊 Results:
🔹 Total: %d
✅ Removed: %d
⏭ Skipped: %d`,
		hostInfo,
		FormatICT(startTime),
		FormatICT(endTime),
		duration.Round(time.Second),
		total,
		removed,
		skipped)
}
