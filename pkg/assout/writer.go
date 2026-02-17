package assout

import (
	"fmt"
	"io"
	"sort"
	"strings"

	"mkv-sub-extractor/pkg/subtitle"
)

// WriteASSPassthrough writes a complete ASS file from a CodecPrivate header and
// subtitle events. The CodecPrivate bytes are written as the ASS header, with the
// [Events] section and Dialogue lines appended.
//
// If the CodecPrivate contains [V4 Styles] (SSA format), the header is automatically
// converted to [V4+ Styles] (ASS format) using ConvertSSAHeaderToASS. This handles
// both S_TEXT/SSA tracks and S_TEXT/ASS tracks with legacy V4 headers.
//
// The function handles three CodecPrivate scenarios:
//   - No [Events] section: appends [Events] + default Format line
//   - [Events] section without Format line: appends default Format line
//   - [Events] section with Format line: uses existing Format line (no duplication)
//
// Events are sorted by StartTime (primary) and ReadOrder (secondary) before writing.
// Output uses CRLF line endings per ASS convention.
func WriteASSPassthrough(w io.Writer, codecPrivate []byte, codecID string, events []subtitle.SubtitleEvent) error {
	header := string(codecPrivate)

	// Always attempt SSA→ASS header conversion. Some MKV files have CodecID
	// "S_TEXT/ASS" but CodecPrivate with [V4 Styles] (SSA format). The conversion
	// function already guards on [V4 Styles] presence, so calling it unconditionally
	// is safe — it's a no-op for proper V4+ headers.
	header = ConvertSSAHeaderToASS(header)

	// Normalize line endings: convert any existing CRLF to LF first for consistent processing,
	// then we'll write with CRLF
	header = strings.ReplaceAll(header, "\r\n", "\n")
	header = strings.ReplaceAll(header, "\r", "\n")

	// Determine [Events] section handling
	headerLower := strings.ToLower(header)
	eventsIdx := strings.Index(headerLower, "[events]")
	hasEventsSection := eventsIdx >= 0
	hasFormatLine := false

	if hasEventsSection {
		// Check if there's a Format: line after [Events]
		afterEvents := header[eventsIdx:]
		lines := strings.Split(afterEvents, "\n")
		for _, line := range lines[1:] { // skip the [Events] line itself
			trimmed := strings.TrimSpace(line)
			if trimmed == "" {
				continue
			}
			if strings.HasPrefix(trimmed, "Format:") {
				hasFormatLine = true
				break
			}
			// If we hit a non-empty, non-Format line or a section header, stop looking
			if strings.HasPrefix(trimmed, "[") {
				break
			}
		}
	}

	// Write header with CRLF line endings
	headerLines := strings.Split(header, "\n")
	for i, line := range headerLines {
		// Don't add trailing CRLF if this is the last line and it's empty
		// (which would have resulted from a trailing newline in the original)
		if i == len(headerLines)-1 && line == "" {
			break
		}
		if _, err := io.WriteString(w, line+"\r\n"); err != nil {
			return fmt.Errorf("writing header: %w", err)
		}
	}

	// Append [Events] section if not present in header
	if !hasEventsSection {
		if _, err := io.WriteString(w, "\r\n[Events]\r\n"); err != nil {
			return fmt.Errorf("writing events section: %w", err)
		}
		if _, err := io.WriteString(w, defaultEventsFormat); err != nil {
			return fmt.Errorf("writing events format: %w", err)
		}
	} else if !hasFormatLine {
		// [Events] exists but no Format line -- add it
		if _, err := io.WriteString(w, defaultEventsFormat); err != nil {
			return fmt.Errorf("writing events format: %w", err)
		}
	}

	// Sort events: primary by StartTime, secondary by ReadOrder
	sorted := make([]subtitle.SubtitleEvent, len(events))
	copy(sorted, events)
	sort.Slice(sorted, func(i, j int) bool {
		if sorted[i].Start != sorted[j].Start {
			return sorted[i].Start < sorted[j].Start
		}
		return sorted[i].ReadOrder < sorted[j].ReadOrder
	})

	// Write Dialogue lines
	for _, ev := range sorted {
		start := FormatASSTimestamp(ev.Start)
		end := FormatASSTimestamp(ev.End)

		line := fmt.Sprintf("Dialogue: %d,%s,%s,%s,%s,%s,%s,%s,%s,%s\r\n",
			ev.Layer,
			start, end,
			ev.Style, ev.Name,
			ev.MarginL, ev.MarginR, ev.MarginV,
			ev.Effect, ev.Text,
		)
		if _, err := io.WriteString(w, line); err != nil {
			return fmt.Errorf("writing dialogue line: %w", err)
		}
	}

	return nil
}
