// Package cli provides the command-line interface for mkv-sub-extractor,
// including structured error types, flag parsing, and interactive prompts.
package cli

import (
	"fmt"

	"github.com/charmbracelet/lipgloss"
)

// Exit code categories.
const (
	ExitGeneral    = 1 // general / unexpected error
	ExitFileError  = 2 // file-related error (not found, not MKV, unreadable)
	ExitTrackError = 3 // track-related error (no subtitles, image-only, not found)
	ExitExtraction = 4 // extraction error (write failure, partial failure)
)

// CLIError is a structured error type with rustc-style formatting.
// Each error carries a machine-readable code, human-readable context,
// and an actionable suggestion for the user.
type CLIError struct {
	Code       string // error code, e.g. "E01", "E02"
	Title      string // short title, e.g. "File Not Found"
	Context    string // relevant path or identifier
	Detail     string // what happened (1 sentence)
	Suggestion string // how to fix it (optional, may be empty)
	ExitCode   int    // process exit code
}

// Error satisfies the error interface. Returns the title for simple error chains.
func (e *CLIError) Error() string {
	return e.Title
}

// lipgloss styles for rustc-style error formatting.
var (
	errorTitleStyle = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))
	arrowPipeStyle  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))
	suggestionStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))
)

// Format renders the error in rustc-style multi-line format:
//
//	error[E02]: File Not Found
//	  --> video.mkv
//	   |
//	   = The file "video.mkv" does not exist or cannot be read.
//	   = Try: Check the file path and ensure the file exists.
func (e *CLIError) Format() string {
	title := errorTitleStyle.Render(fmt.Sprintf("error[%s]: %s", e.Code, e.Title))
	arrow := arrowPipeStyle.Render("  -->")
	pipe := arrowPipeStyle.Render("   |")
	eq := arrowPipeStyle.Render("   =")

	result := fmt.Sprintf("%s\n%s %s\n%s\n%s %s", title, arrow, e.Context, pipe, eq, e.Detail)

	if e.Suggestion != "" {
		suggestion := suggestionStyle.Render(fmt.Sprintf("Try: %s", e.Suggestion))
		result += fmt.Sprintf("\n%s %s", eq, suggestion)
	}

	return result
}

// --- Factory functions ---

// ErrFileNotFound creates a CLIError for when the specified file does not exist.
func ErrFileNotFound(path string) *CLIError {
	return &CLIError{
		Code:       "E01",
		Title:      "File Not Found",
		Context:    path,
		Detail:     fmt.Sprintf("The file %q does not exist or cannot be read.", path),
		Suggestion: "Check the file path and ensure the file exists.",
		ExitCode:   ExitFileError,
	}
}

// ErrNotMKVFile creates a CLIError for when the file is not an MKV file.
func ErrNotMKVFile(path string) *CLIError {
	return &CLIError{
		Code:       "E02",
		Title:      "Not an MKV File",
		Context:    path,
		Detail:     fmt.Sprintf("The file %q does not have a .mkv extension.", path),
		Suggestion: "Provide a Matroska (.mkv) file.",
		ExitCode:   ExitFileError,
	}
}

// ErrCannotReadFile creates a CLIError for when the file exists but cannot be read.
func ErrCannotReadFile(path string, reason error) *CLIError {
	return &CLIError{
		Code:       "E03",
		Title:      "Cannot Read File",
		Context:    path,
		Detail:     fmt.Sprintf("The file %q exists but cannot be read: %v", path, reason),
		Suggestion: "Check file permissions and ensure it is not a directory.",
		ExitCode:   ExitFileError,
	}
}

// ErrNoMKVFilesInDir creates a CLIError for when no MKV files are found in the directory.
func ErrNoMKVFilesInDir(dir string) *CLIError {
	return &CLIError{
		Code:       "E04",
		Title:      "No MKV Files Found",
		Context:    dir,
		Detail:     fmt.Sprintf("No .mkv files were found in the directory %q.", dir),
		Suggestion: "Navigate to a directory containing MKV files, or specify a file path directly.",
		ExitCode:   ExitFileError,
	}
}

// ErrNoSubtitleTracks creates a CLIError for when the MKV has no subtitle tracks at all.
func ErrNoSubtitleTracks(path string) *CLIError {
	return &CLIError{
		Code:       "E10",
		Title:      "No Subtitle Tracks",
		Context:    path,
		Detail:     fmt.Sprintf("The file %q contains no subtitle tracks.", path),
		Suggestion: "",
		ExitCode:   ExitTrackError,
	}
}

// ErrNoExtractableTracks creates a CLIError for when only image-based subtitles exist.
func ErrNoExtractableTracks(path string) *CLIError {
	return &CLIError{
		Code:       "E11",
		Title:      "No Extractable Subtitle Tracks",
		Context:    path,
		Detail:     fmt.Sprintf("The file %q contains only image-based subtitle tracks (e.g., PGS, VobSub) which cannot be extracted as text.", path),
		Suggestion: "Use OCR tools like SubtitleEdit or PGS2SRT to convert image subtitles to text.",
		ExitCode:   ExitTrackError,
	}
}

// ErrImageTrackSelected creates a CLIError for when the user selects an image-based track.
func ErrImageTrackSelected(trackIndex int, formatType string) *CLIError {
	return &CLIError{
		Code:       "E12",
		Title:      "Image-Based Track Selected",
		Context:    fmt.Sprintf("Track %d (%s)", trackIndex, formatType),
		Detail:     fmt.Sprintf("Track %d is an image-based subtitle (%s) and cannot be extracted as text.", trackIndex, formatType),
		Suggestion: "Use OCR tools like SubtitleEdit or PGS2SRT to convert image subtitles to text.",
		ExitCode:   ExitTrackError,
	}
}

// ErrTrackNotFound creates a CLIError for when a specified track number does not exist.
func ErrTrackNotFound(trackNum int, path string) *CLIError {
	return &CLIError{
		Code:       "E13",
		Title:      "Track Not Found",
		Context:    path,
		Detail:     fmt.Sprintf("Track number %d does not exist in %q.", trackNum, path),
		Suggestion: "Run without --track to see available tracks in interactive mode.",
		ExitCode:   ExitTrackError,
	}
}

// ErrExtractionFailed creates a CLIError for when extraction of a track fails.
func ErrExtractionFailed(trackIndex int, reason error) *CLIError {
	return &CLIError{
		Code:       "E20",
		Title:      "Extraction Failed",
		Context:    fmt.Sprintf("Track %d", trackIndex),
		Detail:     fmt.Sprintf("Failed to extract track %d: %v", trackIndex, reason),
		Suggestion: "",
		ExitCode:   ExitExtraction,
	}
}

// ErrPartialFailure creates a CLIError for when some tracks were extracted but others failed.
func ErrPartialFailure(succeeded int, failed int) *CLIError {
	return &CLIError{
		Code:       "E21",
		Title:      "Partial Extraction Failure",
		Context:    fmt.Sprintf("%d succeeded, %d failed", succeeded, failed),
		Detail:     fmt.Sprintf("Extracted %d track(s) successfully, but %d track(s) failed.", succeeded, failed),
		Suggestion: "Check the error details above for each failed track.",
		ExitCode:   ExitExtraction,
	}
}
