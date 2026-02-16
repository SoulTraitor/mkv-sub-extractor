package mkvinfo

import (
	"fmt"
	"time"
)

// FormatFileSize converts a byte count to a human-readable string.
// Uses 1024-based units (KiB-style values with KB/MB/GB/TB labels).
// Returns "Unknown" for negative values.
func FormatFileSize(bytes int64) string {
	if bytes < 0 {
		return "Unknown"
	}
	const unit = 1024
	if bytes < unit {
		return fmt.Sprintf("%d B", bytes)
	}
	div, exp := int64(unit), 0
	for n := bytes / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}

// FormatDuration converts a time.Duration to "H:MM:SS" format.
// Returns "Unknown" for zero or negative durations.
func FormatDuration(d time.Duration) string {
	if d <= 0 {
		return "Unknown"
	}
	h := int(d.Hours())
	m := int(d.Minutes()) % 60
	s := int(d.Seconds()) % 60
	return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}
