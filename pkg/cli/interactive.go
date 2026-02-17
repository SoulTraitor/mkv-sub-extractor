package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/bubbles/key"
	"github.com/charmbracelet/huh"
	"github.com/charmbracelet/lipgloss"

	"mkv-sub-extractor/pkg/mkvinfo"
)

// quitKeyMap returns a KeyMap with Esc added as an alternative quit binding
// alongside Ctrl+C, so users can exit interactive forms without killing the process.
// The help text ensures "esc" appears in the bottom key hints bar.
func quitKeyMap() *huh.KeyMap {
	km := huh.NewDefaultKeyMap()
	km.Quit = key.NewBinding(
		key.WithKeys("ctrl+c", "esc"),
		key.WithHelp("esc", "cancel"),
	)
	return km
}

// faintStyle is used to display non-extractable (image-based) tracks
// as grayed-out context above the track multi-selector.
var faintStyle = lipgloss.NewStyle().Faint(true)

// PickMKVFile scans the current working directory for MKV files and presents
// an interactive arrow-key selector. Returns the selected file path.
//
// Per user decision: always presents the selector even if only 1 file exists.
func PickMKVFile() (string, error) {
	dir, err := os.Getwd()
	if err != nil {
		return "", fmt.Errorf("get working directory: %w", err)
	}

	// Scan for MKV files using both cases; deduplicate via map (Windows
	// is case-insensitive so *.mkv and *.MKV may overlap).
	seen := make(map[string]bool)
	var files []string

	for _, pattern := range []string{"*.mkv", "*.MKV"} {
		matches, err := filepath.Glob(filepath.Join(dir, pattern))
		if err != nil {
			continue
		}
		for _, m := range matches {
			base := filepath.Base(m)
			key := strings.ToLower(base)
			if !seen[key] {
				seen[key] = true
				files = append(files, base)
			}
		}
	}

	if len(files) == 0 {
		return "", ErrNoMKVFilesInDir(dir)
	}

	// Build huh options.
	options := make([]huh.Option[string], len(files))
	for i, f := range files {
		options[i] = huh.NewOption(f, f)
	}

	var selected string
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewSelect[string]().
				Title("Select an MKV file").
				Description("Esc cancel").
				Options(options...).
				Value(&selected),
		),
	).WithKeyMap(quitKeyMap())

	if err := form.Run(); err != nil {
		return "", fmt.Errorf("file picker: %w", err)
	}

	return filepath.Join(dir, selected), nil
}

// SelectTracks presents an interactive multi-select for extractable subtitle tracks.
// Image-based tracks are displayed as grayed-out context above the selector.
//
// Per user decision: always presents the selector even if only 1 extractable track.
// Validation rejects empty selection.
func SelectTracks(tracks []mkvinfo.SubtitleTrack) ([]mkvinfo.SubtitleTrack, error) {
	// Separate extractable and image-based tracks.
	var extractable []mkvinfo.SubtitleTrack
	var imageBased []mkvinfo.SubtitleTrack

	for _, t := range tracks {
		if t.IsExtractable {
			extractable = append(extractable, t)
		} else {
			imageBased = append(imageBased, t)
		}
	}

	if len(extractable) == 0 {
		if len(tracks) > 0 {
			// Tracks exist but none are extractable (all image-based).
			return nil, ErrNoExtractableTracks(fmt.Sprintf("%d image-based track(s)", len(imageBased)))
		}
		return nil, ErrNoSubtitleTracks("(unknown)")
	}

	// Display image-based tracks as grayed-out context above the selector.
	if len(imageBased) > 0 {
		fmt.Println(faintStyle.Render("  Image-based tracks (not extractable):"))
		for _, t := range imageBased {
			line := fmt.Sprintf("    %s -- use OCR tools like SubtitleEdit", formatTrackOption(t))
			fmt.Println(faintStyle.Render(line))
		}
		fmt.Println() // blank line before the selector
	}

	// Build huh options from extractable tracks.
	// Map by index for lookup after selection.
	trackByIndex := make(map[int]mkvinfo.SubtitleTrack)
	options := make([]huh.Option[int], len(extractable))
	for i, t := range extractable {
		options[i] = huh.NewOption(formatTrackOption(t), t.Index)
		trackByIndex[t.Index] = t
	}

	var selectedIndices []int
	form := huh.NewForm(
		huh.NewGroup(
			huh.NewMultiSelect[int]().
				Title("Select subtitle tracks to extract").
				Description("Space/x toggle · Esc cancel").
				Options(options...).
				Value(&selectedIndices),
		),
	).WithKeyMap(quitKeyMap())

	if err := form.Run(); err != nil {
		return nil, fmt.Errorf("track selector: %w", err)
	}

	// Reject empty selection (checked after submit instead of via Validate
	// to avoid showing an error message on initial render).
	if len(selectedIndices) == 0 {
		return nil, fmt.Errorf("track selector: no tracks selected")
	}

	// Map selected indices back to SubtitleTrack objects.
	result := make([]mkvinfo.SubtitleTrack, 0, len(selectedIndices))
	for _, idx := range selectedIndices {
		if t, ok := trackByIndex[idx]; ok {
			result = append(result, t)
		}
	}

	return result, nil
}

// formatTrackOption formats a SubtitleTrack as a concise menu label.
// Format: [{Index}] {LanguageName} ({FormatType}) "Name"
// The optional Name suffix is included only when non-empty.
func formatTrackOption(t mkvinfo.SubtitleTrack) string {
	label := fmt.Sprintf("[%d] %s (%s)", t.Index, t.LanguageName, t.FormatType)
	if t.Name != "" {
		label += fmt.Sprintf(" %q", t.Name)
	}
	return label
}
