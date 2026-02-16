package assout

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"mkv-sub-extractor/pkg/subtitle"
)

// WriteSRTAsASS writes subtitle events to w in ASS format using a default
// generated header with Microsoft YaHei font style. Events are sorted by
// start time. SRT HTML tags in event text are converted to ASS override tags,
// and newlines are converted to \N hard line breaks.
func WriteSRTAsASS(w io.Writer, events []subtitle.SubtitleEvent) error {
	// Write the default header (Script Info + V4+ Styles)
	if _, err := io.WriteString(w, defaultASSHeader); err != nil {
		return fmt.Errorf("writing ASS header: %w", err)
	}

	// Write Events section header
	if _, err := io.WriteString(w, "\r\n[Events]\r\n"); err != nil {
		return fmt.Errorf("writing events section: %w", err)
	}

	// Write the events Format line
	if _, err := io.WriteString(w, defaultEventsFormat); err != nil {
		return fmt.Errorf("writing events format: %w", err)
	}

	// Sort events by start time (ascending)
	sorted := make([]subtitle.SubtitleEvent, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		return sorted[i].Start < sorted[j].Start
	})

	// Write each event as a Dialogue line
	for _, ev := range sorted {
		start := FormatASSTimestamp(ev.Start)
		end := FormatASSTimestamp(ev.End)

		// Convert SRT tags to ASS override tags
		text := subtitle.ConvertSRTTagsToASS(ev.Text)

		// Strip \r and convert \n to \N (ASS hard line break)
		text = strings.ReplaceAll(text, "\r", "")
		text = strings.ReplaceAll(text, "\n", `\N`)

		line := fmt.Sprintf("Dialogue: 0,%s,%s,Default,,0,0,0,,%s\r\n", start, end, text)
		if _, err := io.WriteString(w, line); err != nil {
			return fmt.Errorf("writing dialogue line: %w", err)
		}
	}

	return nil
}
