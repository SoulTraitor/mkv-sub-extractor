package cli

import (
	"fmt"
	"os"
	"path/filepath"
	"strings"

	"github.com/charmbracelet/lipgloss"

	"mkv-sub-extractor/pkg/mkvinfo"
)

// Lipgloss styles for the completion summary.
var (
	successStyle = lipgloss.NewStyle().Foreground(lipgloss.Color("10")) // green
	failStyle    = lipgloss.NewStyle().Foreground(lipgloss.Color("9"))  // red
	dimStyle     = lipgloss.NewStyle().Faint(true)
	boldStyle    = lipgloss.NewStyle().Bold(true)
)

// Run is the main entry point for the CLI. It parses flags, validates input,
// dispatches to the appropriate mode (interactive or scriptable), and returns
// a process exit code.
func Run() int {
	cfg := ParseFlags()

	if err := ValidateConfig(cfg); err != nil {
		fmt.Fprintln(os.Stderr, err.Format())
		return err.ExitCode
	}

	// Dispatch based on mode: --track present means scriptable, otherwise interactive.
	if len(cfg.TrackNumbers) > 0 {
		return runScriptable(cfg)
	}
	return runInteractive(cfg)
}

// runInteractive handles the interactive mode: file picker -> MKV parse ->
// track selection -> extraction -> completion summary.
func runInteractive(cfg Config) int {
	mkvPath := cfg.MKVPath

	// If no MKV path provided, present the interactive file picker.
	if mkvPath == "" {
		picked, err := PickMKVFile()
		if err != nil {
			if cliErr, ok := err.(*CLIError); ok {
				fmt.Fprintln(os.Stderr, cliErr.Format())
				return cliErr.ExitCode
			}
			fmt.Fprintf(os.Stderr, "Error: %v\n", err)
			return ExitGeneral
		}
		mkvPath = picked
	}

	// Parse MKV info.
	result, err := mkvinfo.GetMKVInfo(mkvPath)
	if err != nil {
		cliErr := ErrCannotReadFile(mkvPath, err)
		fmt.Fprintln(os.Stderr, cliErr.Format())
		return cliErr.ExitCode
	}

	// Check for subtitle tracks.
	if result.Info.SubtitleCount == 0 {
		cliErr := ErrNoSubtitleTracks(mkvPath)
		fmt.Fprintln(os.Stderr, cliErr.Format())
		return cliErr.ExitCode
	}

	// Display file info and track listing.
	fmt.Println(mkvinfo.FormatTrackListing(*result))
	fmt.Println()

	// Interactive track selection.
	selectedTracks, err := SelectTracks(result.Tracks)
	if err != nil {
		if cliErr, ok := err.(*CLIError); ok {
			fmt.Fprintln(os.Stderr, cliErr.Format())
			return cliErr.ExitCode
		}
		// User cancelled (huh returns ErrUserAborted or similar).
		if strings.Contains(err.Error(), "user aborted") ||
			strings.Contains(err.Error(), "aborted") {
			fmt.Println("Cancelled.")
			return 0
		}
		fmt.Fprintf(os.Stderr, "Error: %v\n", err)
		return ExitGeneral
	}

	// Extract tracks with progress.
	results := ExtractWithProgress(mkvPath, selectedTracks, cfg.OutputDir, cfg.Quiet)

	// Print completion summary.
	printCompletionSummary(results, cfg.Quiet)

	// Return exit code based on results.
	return exitCodeFromResults(results)
}

// runScriptable handles the non-interactive scriptable mode with --track flag.
func runScriptable(cfg Config) int {
	// MKV path is required in scriptable mode.
	if cfg.MKVPath == "" {
		cliErr := ErrFileNotFound("(no file specified)")
		cliErr.Detail = "A file path is required when using --track for non-interactive extraction."
		cliErr.Suggestion = "Provide the MKV file path as a positional argument: mkv-sub-extractor video.mkv --track 1,2"
		fmt.Fprintln(os.Stderr, cliErr.Format())
		return cliErr.ExitCode
	}

	// Parse MKV info.
	result, err := mkvinfo.GetMKVInfo(cfg.MKVPath)
	if err != nil {
		cliErr := ErrCannotReadFile(cfg.MKVPath, err)
		fmt.Fprintln(os.Stderr, cliErr.Format())
		return cliErr.ExitCode
	}

	// Build a lookup of subtitle tracks by display index.
	trackByIndex := make(map[int]mkvinfo.SubtitleTrack)
	for _, t := range result.Tracks {
		trackByIndex[t.Index] = t
	}

	// Resolve requested track numbers and validate.
	var resolvedTracks []mkvinfo.SubtitleTrack
	var validationErrors []*CLIError

	for _, num := range cfg.TrackNumbers {
		track, found := trackByIndex[num]
		if !found {
			validationErrors = append(validationErrors, ErrTrackNotFound(num, cfg.MKVPath))
			continue
		}
		if !track.IsExtractable {
			validationErrors = append(validationErrors, ErrImageTrackSelected(track.Index, track.FormatType))
			continue
		}
		resolvedTracks = append(resolvedTracks, track)
	}

	// If any track validation errors, print all and exit.
	if len(validationErrors) > 0 {
		for _, e := range validationErrors {
			fmt.Fprintln(os.Stderr, e.Format())
			fmt.Fprintln(os.Stderr)
		}
		return ExitTrackError
	}

	// Verbose: show file info and selected track details before extraction.
	if cfg.Verbose {
		fmt.Println(mkvinfo.FormatTrackListing(*result))
		fmt.Println()
		fmt.Printf("Selected tracks for extraction: ")
		for i, t := range resolvedTracks {
			if i > 0 {
				fmt.Print(", ")
			}
			fmt.Printf("[%d] %s (%s)", t.Index, t.LanguageName, t.FormatType)
			if t.CodecID != "" {
				fmt.Printf(" codec=%s", t.CodecID)
			}
			if t.Language != "" {
				fmt.Printf(" lang=%s", t.Language)
			}
		}
		fmt.Println()
		fmt.Println()
	}

	// Extract tracks with progress.
	results := ExtractWithProgress(cfg.MKVPath, resolvedTracks, cfg.OutputDir, cfg.Quiet)

	// Quiet mode: only print output file paths on stdout.
	if cfg.Quiet {
		for _, r := range results {
			if r.Error == nil {
				fmt.Println(r.OutputPath)
			}
		}
		return exitCodeFromResults(results)
	}

	// Normal output: print completion summary.
	printCompletionSummary(results, false)

	return exitCodeFromResults(results)
}

// printCompletionSummary prints a styled summary of extraction results.
func printCompletionSummary(results []TrackResult, quiet bool) {
	if quiet {
		return
	}

	succeeded := 0
	failed := 0
	for _, r := range results {
		if r.Error == nil {
			succeeded++
		} else {
			failed++
		}
	}

	fmt.Println()
	if failed == 0 {
		fmt.Println(boldStyle.Render("Extraction complete!"))
	} else {
		fmt.Println(boldStyle.Render("Extraction complete with errors."))
	}
	fmt.Println()

	for _, r := range results {
		track := r.Track
		prefix := fmt.Sprintf("  [%d] %s (%s)", track.Index, track.LanguageName, track.FormatType)

		if r.Error == nil {
			// Show just the filename, not the full path.
			outName := filepath.Base(r.OutputPath)
			line := fmt.Sprintf("%s -> %s", prefix, outName)
			fmt.Println(successStyle.Render(line))
		} else {
			line := fmt.Sprintf("%s -> FAILED: %v", prefix, r.Error)
			fmt.Println(failStyle.Render(line))
		}
	}

	fmt.Println()
	if failed == 0 {
		summary := fmt.Sprintf("  %d track(s) extracted successfully", succeeded)
		fmt.Println(successStyle.Render(summary))
	} else {
		summary := fmt.Sprintf("  %d of %d track(s) extracted successfully", succeeded, len(results))
		fmt.Println(dimStyle.Render(summary))
	}
}

// exitCodeFromResults returns 0 if all extractions succeeded,
// ExitExtraction (4) if any failed.
func exitCodeFromResults(results []TrackResult) int {
	for _, r := range results {
		if r.Error != nil {
			return ExitExtraction
		}
	}
	return 0
}
