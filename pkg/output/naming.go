package output

import (
	"fmt"
	"path/filepath"
	"regexp"
	"strings"

	"mkv-sub-extractor/pkg/mkvinfo"
)

// reUnsafeChars matches characters that are not safe for filenames.
var reUnsafeChars = regexp.MustCompile(`[^a-zA-Z0-9_\-. ]`)

// reMultiUnderscore collapses runs of underscores.
var reMultiUnderscore = regexp.MustCompile(`_+`)

// GenerateOutputPath generates the output ASS file path for a subtitle track.
//
// Format: {video_basename}.{lang_code}.ass with ISO 639-2 three-letter codes.
//
// Collision handling:
//  1. If path is unique, return it
//  2. If track has a non-empty Name, try {basename}.{lang}.{sanitized_name}.ass
//  3. Otherwise, try sequence numbers: {basename}.{lang}.2.ass, .3.ass, etc.
//
// The returned path is automatically added to existingPaths.
func GenerateOutputPath(videoPath string, track mkvinfo.SubtitleTrack, existingPaths map[string]bool) string {
	dir := filepath.Dir(videoPath)
	base := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))

	lang := track.Language
	if lang == "" || lang == "und" {
		lang = "und"
	}

	// Try basic path first
	candidate := filepath.Join(dir, fmt.Sprintf("%s.%s.ass", base, lang))
	if !existingPaths[candidate] {
		existingPaths[candidate] = true
		return candidate
	}

	// Collision: try with sanitized track name
	if track.Name != "" {
		sanitized := sanitizeFileName(track.Name)
		candidate = filepath.Join(dir, fmt.Sprintf("%s.%s.%s.ass", base, lang, sanitized))
		if !existingPaths[candidate] {
			existingPaths[candidate] = true
			return candidate
		}
	}

	// Fallback: sequence numbers starting at 2
	for n := 2; ; n++ {
		candidate = filepath.Join(dir, fmt.Sprintf("%s.%s.%d.ass", base, lang, n))
		if !existingPaths[candidate] {
			existingPaths[candidate] = true
			return candidate
		}
	}
}

// sanitizeFileName makes a track name safe for use in filenames.
//
// Replaces characters not in [a-zA-Z0-9_\-. ] with underscore,
// collapses multiple underscores, trims leading/trailing underscores and spaces,
// limits to 50 characters, and returns "track" if the result is empty.
func sanitizeFileName(name string) string {
	// Replace unsafe characters with underscore
	result := reUnsafeChars.ReplaceAllString(name, "_")

	// Collapse multiple underscores
	result = reMultiUnderscore.ReplaceAllString(result, "_")

	// Trim leading/trailing underscores and spaces
	result = strings.Trim(result, "_ ")

	// Limit to 50 characters
	if len(result) > 50 {
		result = result[:50]
		// Re-trim in case we cut in the middle of trailing underscores/spaces
		result = strings.TrimRight(result, "_ ")
	}

	// If empty after sanitization, use "track"
	if result == "" {
		return "track"
	}

	return result
}
