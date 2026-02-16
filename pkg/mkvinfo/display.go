package mkvinfo

import (
	"fmt"
	"strings"
)

// ShouldShowDefault returns true only if NOT all tracks have IsDefault=true.
// Per Matroska spec, if all tracks have FlagDefault=true (common default behavior),
// showing [default] on every line is noisy and provides no useful differentiation.
func ShouldShowDefault(tracks []SubtitleTrack) bool {
	if len(tracks) == 0 {
		return false
	}
	for _, t := range tracks {
		if !t.IsDefault {
			return true // At least one is not default, so the flag is meaningful
		}
	}
	return false // All are default — flag adds no information
}

// FormatFileInfoHeader returns a multi-line string with file metadata.
// Format:
//
//	File: {FileName}
//	Size: {size} | Duration: {duration} | Subtitle tracks: {count}
func FormatFileInfoHeader(info FileInfo) string {
	size := FormatFileSize(info.FileSize)
	dur := FormatDuration(info.Duration)
	return fmt.Sprintf("File: %s\nSize: %s | Duration: %s | Subtitle tracks: %d",
		info.FileName, size, dur, info.SubtitleCount)
}

// FormatTrackLine returns a single formatted line for a subtitle track.
//
// For text tracks:
//
//	[{Index}] {LanguageName} ({FormatType}) "Name" CodecID [default] [forced]
//
// For image tracks, appends: -- not extractable (image subtitle)
//
// Parts are space-separated. Optional parts (name, flags) are omitted when empty/false.
func FormatTrackLine(t SubtitleTrack, showDefault bool) string {
	var parts []string

	// Index and language
	parts = append(parts, fmt.Sprintf("[%d]", t.Index))
	parts = append(parts, t.LanguageName)
	parts = append(parts, fmt.Sprintf("(%s)", t.FormatType))

	// Track name (only if non-empty, wrapped in double quotes)
	if t.Name != "" {
		parts = append(parts, fmt.Sprintf("%q", t.Name))
	}

	// Codec ID (always shown)
	parts = append(parts, t.CodecID)

	// Flags
	if showDefault && t.IsDefault {
		parts = append(parts, "[default]")
	}
	if t.IsForced {
		parts = append(parts, "[forced]")
	}

	// Image subtitle marker
	if !t.IsText {
		parts = append(parts, "-- not extractable (image subtitle)")
	}

	return strings.Join(parts, " ")
}

// FormatTrackListing returns the complete formatted output for an MKV file:
// file info header, blank line, then each track on its own line.
// Includes appropriate messages for edge cases (no text subs, no subs at all).
func FormatTrackListing(info MKVInfo) string {
	var sb strings.Builder

	// File info header
	sb.WriteString(FormatFileInfoHeader(info.Info))
	sb.WriteString("\n")

	if info.Info.SubtitleCount == 0 {
		sb.WriteString("\nNo subtitle tracks found in this file.")
		return sb.String()
	}

	// Determine whether to show [default] flags
	showDefault := ShouldShowDefault(info.Tracks)

	// Track lines
	sb.WriteString("\n")
	for i, track := range info.Tracks {
		sb.WriteString(FormatTrackLine(track, showDefault))
		if i < len(info.Tracks)-1 {
			sb.WriteString("\n")
		}
	}

	// No extractable text subtitles message
	if info.Info.TextSubCount == 0 && info.Info.SubtitleCount > 0 {
		sb.WriteString("\n\n")
		sb.WriteString("未找到可提取的文本字幕轨道")
	}

	return sb.String()
}
