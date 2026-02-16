package assout

import "fmt"

// FormatASSTimestamp converts a duration in nanoseconds to an ASS timestamp
// string in the format "H:MM:SS.CC" (single-digit hour, centisecond precision).
//
// Centiseconds are rounded (not truncated): (ns + 5_000_000) / 10_000_000.
// Uses integer arithmetic only to avoid floating-point drift.
func FormatASSTimestamp(ns uint64) string {
	// Round nanoseconds to centiseconds
	cs := (ns + 5_000_000) / 10_000_000

	hours := cs / 360000
	cs %= 360000
	minutes := cs / 6000
	cs %= 6000
	seconds := cs / 100
	centiseconds := cs % 100

	return fmt.Sprintf("%d:%02d:%02d.%02d", hours, minutes, seconds, centiseconds)
}
