package cli

import (
	"fmt"
	"os"
	"strings"

	"github.com/spf13/pflag"
)

// Config holds parsed command-line arguments.
type Config struct {
	MKVPath      string // positional argument: path to MKV file
	TrackNumbers []int  // --track / -t: specific track numbers to extract
	OutputDir    string // --output / -o: output directory (default: same as MKV)
	Quiet        bool   // --quiet / -q: suppress progress output
	Verbose      bool   // --verbose / -v: enable debug-level output
}

// ParseFlags parses command-line arguments using pflag and returns a Config.
// The MKV file path is taken from the first positional argument.
func ParseFlags() Config {
	var cfg Config

	pflag.IntSliceVarP(&cfg.TrackNumbers, "track", "t", nil, "track numbers to extract (comma-separated, e.g., -t 1,3)")
	pflag.StringVarP(&cfg.OutputDir, "output", "o", "", "output directory for extracted files")
	pflag.BoolVarP(&cfg.Quiet, "quiet", "q", false, "suppress progress output (only print file paths)")
	pflag.BoolVarP(&cfg.Verbose, "verbose", "v", false, "enable verbose/debug output")

	pflag.Usage = func() {
		fmt.Fprintf(os.Stderr, "Usage: mkv-sub-extractor [OPTIONS] [FILE.mkv]\n\n")
		fmt.Fprintf(os.Stderr, "Extract subtitle tracks from MKV files to ASS format.\n\n")
		fmt.Fprintf(os.Stderr, "Arguments:\n")
		fmt.Fprintf(os.Stderr, "  FILE.mkv    MKV file to extract subtitles from\n")
		fmt.Fprintf(os.Stderr, "              If omitted, scans current directory for MKV files\n\n")
		fmt.Fprintf(os.Stderr, "Options:\n")
		pflag.PrintDefaults()
		fmt.Fprintf(os.Stderr, "\nExamples:\n")
		fmt.Fprintf(os.Stderr, "  mkv-sub-extractor video.mkv           Extract interactively\n")
		fmt.Fprintf(os.Stderr, "  mkv-sub-extractor video.mkv -t 1,3    Extract tracks 1 and 3\n")
		fmt.Fprintf(os.Stderr, "  mkv-sub-extractor -o subs/ video.mkv  Output to subs/ directory\n")
		fmt.Fprintf(os.Stderr, "  mkv-sub-extractor                     Scan directory for MKV files\n")
	}

	pflag.Parse()

	// First positional argument is the MKV file path.
	if args := pflag.Args(); len(args) > 0 {
		cfg.MKVPath = args[0]
	}

	return cfg
}

// ValidateConfig checks the Config for errors that can be detected before
// opening the MKV file. Track number validation happens later after parsing
// the MKV, since we need the actual track list to verify.
//
// Returns nil if validation passes, or a *CLIError describing the problem.
func ValidateConfig(cfg Config) *CLIError {
	// Check mutually exclusive flags.
	if cfg.Quiet && cfg.Verbose {
		return &CLIError{
			Code:       "E00",
			Title:      "Conflicting Flags",
			Context:    "--quiet and --verbose",
			Detail:     "The --quiet and --verbose flags cannot be used together.",
			Suggestion: "Use only one of --quiet (-q) or --verbose (-v).",
			ExitCode:   ExitGeneral,
		}
	}

	// If MKV path is provided, validate it.
	if cfg.MKVPath != "" {
		// Check file exists.
		info, err := os.Stat(cfg.MKVPath)
		if os.IsNotExist(err) {
			return ErrFileNotFound(cfg.MKVPath)
		}
		if err != nil {
			return ErrCannotReadFile(cfg.MKVPath, err)
		}

		// Check it's not a directory.
		if info.IsDir() {
			return ErrCannotReadFile(cfg.MKVPath, fmt.Errorf("path is a directory, not a file"))
		}

		// Check .mkv extension (case-insensitive).
		if !strings.HasSuffix(strings.ToLower(cfg.MKVPath), ".mkv") {
			return ErrNotMKVFile(cfg.MKVPath)
		}

		// Check file is readable by attempting to open it.
		f, err := os.Open(cfg.MKVPath)
		if err != nil {
			return ErrCannotReadFile(cfg.MKVPath, err)
		}
		f.Close()
	}

	// If output dir is provided, ensure it exists (create if needed).
	if cfg.OutputDir != "" {
		if err := os.MkdirAll(cfg.OutputDir, 0o755); err != nil {
			return &CLIError{
				Code:       "E05",
				Title:      "Cannot Create Output Directory",
				Context:    cfg.OutputDir,
				Detail:     fmt.Sprintf("Failed to create output directory %q: %v", cfg.OutputDir, err),
				Suggestion: "Check the path and ensure you have write permissions.",
				ExitCode:   ExitFileError,
			}
		}
	}

	return nil
}
