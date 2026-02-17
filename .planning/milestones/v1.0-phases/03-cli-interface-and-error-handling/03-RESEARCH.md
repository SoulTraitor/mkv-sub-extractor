# Phase 3: CLI Interface and Error Handling - Research

**Researched:** 2026-02-17
**Domain:** Go CLI interaction — argument parsing, interactive prompts, progress reporting, error formatting
**Confidence:** HIGH

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions

#### Interactive Selection UX
- Arrow-key menu for both file picker and track selector (promptui/survey style, not numbered list)
- Track selector supports multi-select (user can check multiple subtitle tracks for batch extraction)
- Always confirm selection even when only one MKV file or one subtitle track exists (no auto-skip)
- Image-based subtitle tracks (PGS/VobSub) appear in the track selector but are grayed out / disabled with a visual indicator that they are not extractable

#### Command-Line Argument Design
- File path as positional argument: `mkv-sub-extractor video.mkv` (no flag needed for input file)
- No arguments -> scan current directory for MKV files -> interactive file picker
- Track specification: `--track 3,5` (comma-separated track numbers for non-interactive/scriptable mode)
- Output directory: `--output <dir>` overrides default (default = same directory as MKV file)
- No `--list` mode -- interactive mode already shows track listing
- When `--track` is specified, run non-interactively (scriptable for CI/automation)

#### Output Feedback and Progress
- Real-time progress bar during extraction (based on processed packets / total packets or similar metric)
- `--quiet` flag suppresses progress output (only final file paths on stdout)
- `--verbose` flag adds debug-level information (packet counts, timing details, codec decisions)
- Default verbosity: progress bar + completion summary

#### Completion Summary
- Claude's Discretion -- design the most informative yet concise completion output (file paths, track info, subtitle counts are all reasonable)

#### Error Message Style
- Structured errors similar to rustc/elm: title line + context + suggestion, multi-line format
- Each error type has a distinct, human-readable message (not generic "something went wrong")
- Exit codes are categorized by error type (1=general, 2=file error, 3=track error, etc.)
- Multi-track extraction: if one track fails, continue extracting remaining tracks; report all failures at the end
- Partial output files cleaned up on individual track failure (already implemented in Phase 2 pipeline)

### Claude's Discretion
- Exact progress bar library/implementation choice
- Completion summary format and content
- Short flag aliases (e.g., -t for --track, -o for --output, -q for --quiet, -v for --verbose)
- Help text formatting and examples
- Color usage in terminal output (errors, progress, success)

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

<phase_requirements>
## Phase Requirements

| ID | Description | Research Support |
|----|-------------|-----------------|
| CLI-01 | User can specify MKV file path via command-line argument | pflag positional arg handling (remaining args after flag parse) |
| CLI-02 | No file arg: scan current directory, interactive file picker | `filepath.Glob("*.mkv")` + charmbracelet/huh Select |
| CLI-03 | Interactive subtitle track listing with selection | charmbracelet/huh MultiSelect with custom Option rendering |
| CLI-04 | `--track` flag specifies track numbers, skips interactive | pflag IntSliceVarP with comma-separated values |
| CLI-05 | Output to MKV directory by default, auto-generated names | Already implemented in `pkg/output/naming.go` |
| CLI-06 | `--output` flag overrides output directory | pflag StringVarP passed to `ExtractTrackToASS` outputDir param |
| ERRH-01 | Non-MKV file: clear error message | Custom error type with structured format, exit code 2 |
| ERRH-02 | File not found / unreadable: clear error | Custom error type with structured format, exit code 2 |
| ERRH-03 | No subtitle tracks: clear message | Detect via `MKVInfo.Info.SubtitleCount == 0`, exit code 3 |
| ERRH-04 | Image subtitle selected: prompt text-only + suggest tools | Detect via `track.IsExtractable == false`, structured error with suggestion |
</phase_requirements>

## Summary

Phase 3 wraps the completed extraction engine (Phase 2) in a polished CLI experience. The primary technical challenge is the interactive selection UX: the user requires arrow-key navigation with multi-select for tracks, with grayed-out display of image-based subtitle tracks. The secondary challenge is structured rustc-style error messages.

**charmbracelet/huh** (v0.8.0) is the clear choice for interactive prompts. It is the modern successor to the now-archived AlecAivazis/survey library, provides both `Select` and `MultiSelect` fields with arrow-key navigation out of the box, and is built on top of Bubble Tea. One gap: huh does not natively support "disabled" options, so image-based subtitle tracks require a workaround (display them in the list with a visual indicator but use validation to reject selection, or filter them from MultiSelect and display them separately as context).

For argument parsing, **spf13/pflag** (the POSIX flag library) is sufficient -- no need for cobra since this is a single-command tool. pflag gives `--flag` style, short aliases (`-t`), and `IntSlice` type for comma-separated `--track 3,5`.

For progress reporting, **schollz/progressbar/v3** is a lightweight, thread-safe progress bar that works standalone (no Bubble Tea integration needed). It is the most popular Go progress bar library and supports customization, colors, and silent mode.

For terminal colors in error messages and output, **charmbracelet/lipgloss** integrates naturally with huh (same ecosystem) and provides an expressive styling API for colored, formatted terminal output.

**Primary recommendation:** Use pflag for argument parsing, charmbracelet/huh for interactive prompts, schollz/progressbar for extraction progress, and lipgloss for colored output/error formatting. No cobra needed.

## Standard Stack

### Core
| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/spf13/pflag` | v1.0.5 | POSIX/GNU flag parsing with short aliases | De facto standard for Go CLI flags; drop-in replacement for stdlib flag; used by cobra, kubectl, docker CLI |
| `github.com/charmbracelet/huh` | v0.8.0 | Interactive Select and MultiSelect prompts | Modern replacement for archived survey/promptui; built on Bubble Tea; official Charmbracelet library |
| `github.com/charmbracelet/lipgloss` | v1.x | Terminal text styling and colors | Pairs with huh (same ecosystem); expressive API; auto-detects color profile |
| `github.com/schollz/progressbar/v3` | v3.17+ | Real-time progress bar during extraction | Most popular Go progress bar; thread-safe; supports silent mode; simple API |

### Supporting
| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `github.com/charmbracelet/bubbletea` | (transitive) | TUI framework | Not used directly -- huh depends on it internally |

### Alternatives Considered
| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| pflag | stdlib `flag` | stdlib lacks `--long-flag` POSIX style, no short aliases, no IntSlice |
| pflag | cobra | Cobra adds subcommand routing not needed for single-command tool; pulls pflag anyway |
| huh | manifoldco/promptui | promptui lacks multi-select; less actively maintained |
| huh | AlecAivazis/survey | Archived April 2024; maintainer recommends bubbletea ecosystem |
| huh | raw bubbletea | Requires writing Model/Update/View from scratch; huh abstracts this |
| schollz/progressbar | charmbracelet/bubbles progress | Bubbles progress requires Bubble Tea program; cannot easily mix with huh forms |
| lipgloss | fatih/color | fatih/color is simpler but lacks layout primitives; lipgloss matches huh ecosystem |

**Installation:**
```bash
go get github.com/spf13/pflag@v1.0.5
go get github.com/charmbracelet/huh@v0.8.0
go get github.com/charmbracelet/lipgloss
go get github.com/schollz/progressbar/v3
```

## Architecture Patterns

### Recommended Project Structure
```
cmd/mkv-sub-extractor/
    main.go              # Entry point: parse flags, dispatch to interactive or scriptable mode
pkg/
    cli/
        flags.go         # Flag definitions, parsing, validation
        interactive.go   # Interactive file picker + track selector (huh forms)
        progress.go      # Progress bar wrapper for extraction
        errors.go        # Structured error types and formatting
        run.go           # Main orchestration: run(config) -> exit code
    extract/             # (existing) extraction pipeline
    mkvinfo/             # (existing) MKV parsing
    output/              # (existing) output naming
    assout/              # (existing) ASS file writing
    subtitle/            # (existing) subtitle types
```

### Pattern 1: Two-Mode Dispatch (Interactive vs Scriptable)

**What:** Based on whether `--track` is provided, dispatch to either interactive mode (huh prompts) or scriptable mode (direct extraction with no prompts).

**When to use:** Any CLI tool that supports both human and CI/automation usage.

**Example:**
```go
func run() int {
    cfg := parseFlags()

    // Determine mode
    if cfg.TrackNumbers != nil && len(cfg.TrackNumbers) > 0 {
        return runScriptable(cfg)
    }
    return runInteractive(cfg)
}

func runScriptable(cfg Config) int {
    // No prompts, direct extraction, structured output
    // Exit codes based on error type
}

func runInteractive(cfg Config) int {
    // Step 1: Resolve MKV file (arg or picker)
    // Step 2: Parse MKV info
    // Step 3: Track multi-select
    // Step 4: Extract with progress
    // Step 5: Completion summary
}
```

### Pattern 2: Structured Error Types with Formatting

**What:** Define error types that carry category, title, context, and suggestion. A single formatting function renders them in rustc style.

**When to use:** When the user requires distinct, helpful error messages per error category.

**Example:**
```go
type CLIError struct {
    Category    ErrorCategory  // FileError, TrackError, GeneralError
    Title       string         // "File Not Found"
    Detail      string         // what happened
    Suggestion  string         // how to fix it
    ExitCode    int            // 1, 2, 3, etc.
}

func (e *CLIError) Format() string {
    // error[E02]: File Not Found
    //   --> video.mkv
    //   |
    //   = The file "video.mkv" does not exist or cannot be read.
    //   = Try: Check the file path and ensure the file exists.
}
```

### Pattern 3: Grayed-Out Image Tracks in MultiSelect

**What:** Since huh v0.8.0 does not support disabled options, show image tracks in the list with a visual suffix like " (not extractable - image subtitle)" using dimmed/faint styling, and use validation to reject them if selected.

**When to use:** Implementing the locked decision that image tracks appear but are grayed out.

**Implementation approach:**
```go
// Build options for track multi-select
var options []huh.Option[int]
for _, track := range tracks {
    label := formatTrackLabel(track)
    if !track.IsExtractable {
        // Append visual indicator using lipgloss faint style
        label = lipgloss.NewStyle().Faint(true).Render(label) + " (image - not extractable)"
    }
    options = append(options, huh.NewOption(label, track.Index))
}

// Validate: reject selection of non-extractable tracks
multiSelect.Validate(func(selected []int) error {
    for _, idx := range selected {
        track := trackByIndex(idx)
        if !track.IsExtractable {
            return fmt.Errorf("Track %d (%s) is an image subtitle and cannot be extracted", track.Index, track.FormatType)
        }
    }
    return nil
})
```

**Alternative approach (recommended):** Only include extractable tracks in the MultiSelect options, but display a separate informational block above the selector showing image-based tracks that are not extractable. This avoids the need for disabled options entirely while still meeting the "visual indicator" requirement.

```go
// Show image tracks as context above the selector
if len(imageTracks) > 0 {
    fmt.Println(dimStyle.Render("  Image-based tracks (not extractable):"))
    for _, t := range imageTracks {
        fmt.Println(dimStyle.Render(fmt.Sprintf("    [%d] %s (%s) -- image subtitle", t.Index, t.LanguageName, t.FormatType)))
    }
    fmt.Println()
}

// MultiSelect only has extractable tracks
huh.NewMultiSelect[int]().
    Title("Select subtitle tracks to extract").
    Options(extractableOptions...).
    Value(&selectedTracks)
```

### Pattern 4: Continue-on-Failure Multi-Track Extraction

**What:** The current `ExtractTracksToASS` stops on first error. Phase 3 needs to continue extracting remaining tracks when one fails, collecting all results and errors.

**When to use:** The user explicitly decided this behavior.

**Example:**
```go
type TrackResult struct {
    Track      mkvinfo.SubtitleTrack
    OutputPath string
    Error      error
}

func ExtractTracksWithContinuation(mkvPath string, tracks []mkvinfo.SubtitleTrack, outputDir string) []TrackResult {
    existingPaths := make(map[string]bool)
    results := make([]TrackResult, len(tracks))
    for i, track := range tracks {
        outPath, err := extractTrackToASS(mkvPath, track, outputDir, existingPaths)
        results[i] = TrackResult{Track: track, OutputPath: outPath, Error: err}
    }
    return results
}
```

### Anti-Patterns to Avoid
- **Mixing interactive and scriptable output:** When `--track` is specified, never show prompts or progress bars. When `--quiet` is specified, only emit file paths.
- **Swallowing errors silently:** Each error must be visible. The "continue on failure" pattern collects errors and reports them all at the end.
- **Using cobra for a single command:** Cobra adds subcommand routing, help generation, and other features designed for multi-command CLIs. For a single command, pflag is sufficient and avoids unnecessary complexity.
- **Hardcoding ANSI codes:** Use lipgloss for all styling. It detects terminal capabilities and degrades gracefully.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| Arrow-key selection menus | Custom terminal input loop | charmbracelet/huh Select/MultiSelect | Terminal input handling has dozens of edge cases (Windows, macOS, Linux; different terminal emulators; escape sequences; resize) |
| POSIX-style flag parsing | Manual os.Args parsing | spf13/pflag | Short/long flags, type coercion, IntSlice, help text generation |
| Progress bar rendering | Manual `\r` overwrite | schollz/progressbar | Thread safety, terminal width detection, ETA calculation, rate limiting |
| Terminal color support | Raw ANSI escape codes | charmbracelet/lipgloss | Color profile detection (true color, 256, 16, none), graceful degradation |
| MKV file glob scanning | Manual directory walk | `filepath.Glob("*.mkv")` + `filepath.Glob("*.MKV")` | Built-in, handles OS differences |

**Key insight:** Terminal UX has enormous hidden complexity (platform differences, terminal capabilities, resize handling, Unicode width). Libraries like huh and lipgloss encode years of cross-platform fixes.

## Common Pitfalls

### Pitfall 1: Progress Bar Conflicts with huh Rendering
**What goes wrong:** If a progress bar writes to stderr while a huh form is running, the terminal output becomes garbled.
**Why it happens:** Both systems manipulate terminal cursor position and ANSI sequences.
**How to avoid:** Never run a progress bar and a huh form simultaneously. The flow should be: huh form completes -> progress bar starts. Huh uses bubbletea's alt screen; progress bar writes to stderr normally.
**Warning signs:** Garbled output, progress bar overwriting menu items.

### Pitfall 2: huh Does Not Support Disabled Options
**What goes wrong:** You try to create "grayed out" options in a MultiSelect and discover huh's Option type only has `Selected()`, not `Disabled()`.
**Why it happens:** huh v0.8.0 does not implement disabled options. The Option type is `NewOption(key, value).Selected(bool)` -- no other modifiers.
**How to avoid:** Either (a) display image tracks as context above the selector and only include extractable tracks as selectable options, or (b) include all tracks but validate selection and reject non-extractable picks. Approach (a) is cleaner.
**Warning signs:** Attempting to use a nonexistent `.Disabled()` method.

### Pitfall 3: pflag IntSlice Parses Comma-Separated AND Repeated Flags
**What goes wrong:** `--track 3 --track 5` and `--track 3,5` both work, but `--track 3, 5` (with space after comma) may behave unexpectedly.
**Why it happens:** pflag's IntSlice splits on commas within a single value and also appends repeated flags.
**How to avoid:** Document the expected format clearly in help text: `--track 3,5`. Validate that all track numbers exist in the parsed MKV.
**Warning signs:** Users get "invalid syntax" errors with spaces around commas.

### Pitfall 4: Exit Code 0 on Partial Failure
**What goes wrong:** If 3 tracks are requested and 2 succeed but 1 fails, returning exit code 0 misleads scripts.
**Why it happens:** The "continue on failure" pattern might return success if any track succeeded.
**How to avoid:** Use a specific exit code (e.g., 4) for partial failure. Scripts can distinguish "all succeeded" (0) from "some failed" (4) from "all failed" (3).
**Warning signs:** CI pipelines not detecting extraction failures.

### Pitfall 5: ExtractTracksToASS Stops on First Error
**What goes wrong:** The current `ExtractTracksToASS` in pipeline.go returns immediately on the first track error, not continuing to remaining tracks.
**Why it happens:** Phase 2 implementation returns early on error: `return outputPaths, fmt.Errorf(...)`.
**How to avoid:** Phase 3 must either (a) modify `ExtractTracksToASS` to continue on error, or (b) call `ExtractTrackToASS` individually per track in a loop at the CLI layer. Option (b) is simpler and doesn't require changing the Phase 2 API.
**Warning signs:** User selects 5 tracks, track 2 fails, tracks 3-5 never extracted.

### Pitfall 6: Windows Terminal Color Support
**What goes wrong:** ANSI escape sequences appear as raw text on older Windows terminals (pre-Windows 10 1511).
**Why it happens:** Windows console does not natively understand ANSI sequences without enabling virtual terminal processing.
**How to avoid:** lipgloss detects terminal capabilities and degrades gracefully. Ensure the tool does not assume color support -- use lipgloss for all styling and it handles detection.
**Warning signs:** Raw `\033[31m` text appearing in output on Windows.

## Code Examples

### Flag Parsing with pflag (positional arg + flags)
```go
// Source: spf13/pflag documentation + standard Go patterns
package cli

import (
    "os"
    flag "github.com/spf13/pflag"
)

type Config struct {
    MKVPath      string
    TrackNumbers []int
    OutputDir    string
    Quiet        bool
    Verbose      bool
}

func ParseFlags() Config {
    var cfg Config
    flag.IntSliceVarP(&cfg.TrackNumbers, "track", "t", nil, "Track numbers to extract (comma-separated, e.g., 3,5)")
    flag.StringVarP(&cfg.OutputDir, "output", "o", "", "Output directory (default: same as MKV file)")
    flag.BoolVarP(&cfg.Quiet, "quiet", "q", false, "Suppress progress output; only print output file paths")
    flag.BoolVarP(&cfg.Verbose, "verbose", "v", false, "Print debug-level details (packet counts, timing, codec decisions)")
    flag.Parse()

    // Positional argument: MKV file path
    args := flag.Args()
    if len(args) > 0 {
        cfg.MKVPath = args[0]
    }
    return cfg
}
```

### Interactive File Picker with huh
```go
// Source: charmbracelet/huh documentation
package cli

import (
    "fmt"
    "path/filepath"

    "github.com/charmbracelet/huh"
)

func PickMKVFile() (string, error) {
    matches, _ := filepath.Glob("*.mkv")
    upper, _ := filepath.Glob("*.MKV")
    matches = append(matches, upper...)

    if len(matches) == 0 {
        return "", fmt.Errorf("no MKV files found in current directory")
    }

    var selected string
    options := make([]huh.Option[string], len(matches))
    for i, m := range matches {
        options[i] = huh.NewOption(m, m)
    }

    err := huh.NewSelect[string]().
        Title("Select an MKV file").
        Options(options...).
        Value(&selected).
        Run()

    return selected, err
}
```

### Track Multi-Select with huh
```go
// Source: charmbracelet/huh documentation
package cli

import (
    "fmt"
    "github.com/charmbracelet/huh"
    "github.com/charmbracelet/lipgloss"
    "mkv-sub-extractor/pkg/mkvinfo"
)

var dimStyle = lipgloss.NewStyle().Faint(true)

func SelectTracks(tracks []mkvinfo.SubtitleTrack) ([]mkvinfo.SubtitleTrack, error) {
    var extractable []mkvinfo.SubtitleTrack
    var imageBased []mkvinfo.SubtitleTrack

    for _, t := range tracks {
        if t.IsExtractable {
            extractable = append(extractable, t)
        } else {
            imageBased = append(imageBased, t)
        }
    }

    // Display image-based tracks as non-selectable context
    if len(imageBased) > 0 {
        fmt.Println(dimStyle.Render("  Image-based tracks (not extractable):"))
        for _, t := range imageBased {
            fmt.Println(dimStyle.Render(fmt.Sprintf("    [%d] %s (%s)", t.Index, t.LanguageName, t.FormatType)))
        }
        fmt.Println()
    }

    // Build selectable options from extractable tracks
    options := make([]huh.Option[int], len(extractable))
    for i, t := range extractable {
        label := fmt.Sprintf("[%d] %s (%s)", t.Index, t.LanguageName, t.FormatType)
        if t.Name != "" {
            label += fmt.Sprintf(" %q", t.Name)
        }
        options[i] = huh.NewOption(label, t.Index)
    }

    var selectedIndices []int
    err := huh.NewMultiSelect[int]().
        Title("Select subtitle tracks to extract").
        Options(options...).
        Value(&selectedIndices).
        Run()
    if err != nil {
        return nil, err
    }

    // Map indices back to tracks
    var selected []mkvinfo.SubtitleTrack
    indexMap := make(map[int]mkvinfo.SubtitleTrack)
    for _, t := range extractable {
        indexMap[t.Index] = t
    }
    for _, idx := range selectedIndices {
        selected = append(selected, indexMap[idx])
    }
    return selected, nil
}
```

### Progress Bar During Extraction
```go
// Source: schollz/progressbar documentation
package cli

import (
    "github.com/schollz/progressbar/v3"
    "mkv-sub-extractor/pkg/extract"
    "mkv-sub-extractor/pkg/mkvinfo"
)

func ExtractWithProgress(mkvPath string, tracks []mkvinfo.SubtitleTrack, outputDir string, quiet bool) []TrackResult {
    results := make([]TrackResult, len(tracks))

    var bar *progressbar.ProgressBar
    if !quiet {
        bar = progressbar.Default(int64(len(tracks)), "Extracting")
    }

    for i, track := range tracks {
        outPath, err := extract.ExtractTrackToASS(mkvPath, track, outputDir)
        results[i] = TrackResult{Track: track, OutputPath: outPath, Error: err}
        if bar != nil {
            bar.Add(1)
        }
    }

    if bar != nil {
        bar.Finish()
    }
    return results
}
```

### Structured Error Formatting
```go
// Source: Custom implementation inspired by rustc error format
package cli

import (
    "fmt"
    "github.com/charmbracelet/lipgloss"
)

var (
    errorTitle  = lipgloss.NewStyle().Bold(true).Foreground(lipgloss.Color("9"))  // red
    errorArrow  = lipgloss.NewStyle().Foreground(lipgloss.Color("12"))            // blue
    errorSugg   = lipgloss.NewStyle().Foreground(lipgloss.Color("10"))            // green
)

type CLIError struct {
    Code       string // e.g. "E02"
    Title      string // e.g. "File Not Found"
    Context    string // e.g. path or description
    Detail     string // what happened
    Suggestion string // how to fix
    ExitCode   int
}

func (e *CLIError) Format() string {
    s := errorTitle.Render(fmt.Sprintf("error[%s]: %s", e.Code, e.Title)) + "\n"
    s += errorArrow.Render("  --> ") + e.Context + "\n"
    s += errorArrow.Render("   |") + "\n"
    s += errorArrow.Render("   = ") + e.Detail + "\n"
    if e.Suggestion != "" {
        s += errorSugg.Render("   = Try: ") + e.Suggestion + "\n"
    }
    return s
}
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| AlecAivazis/survey | charmbracelet/huh | 2024 (survey archived April 2024) | survey is read-only; huh is the community-recommended replacement |
| manifoldco/promptui | charmbracelet/huh | 2023-2024 | promptui lacks multi-select; last release 2020 |
| fatih/color for terminal styling | charmbracelet/lipgloss | 2022-2023 | lipgloss provides layout + styling; ecosystem match with huh |
| cobra for all CLIs | pflag for single-command tools | Always (but often overlooked) | cobra adds unnecessary subcommand infrastructure for simple tools |

**Deprecated/outdated:**
- AlecAivazis/survey: Archived April 2024. Maintainer recommends bubbletea ecosystem.
- manifoldco/promptui: Not officially deprecated but effectively unmaintained (last release ~2020). Lacks multi-select.

## Open Questions

1. **Progress granularity within a single track**
   - What we know: The progress bar will track track-by-track progress (1 of N tracks). This is coarse if the user extracts only 1 track.
   - What's unclear: Can we get packet-level progress? The current `ExtractSubtitlePackets` reads all packets in a single loop with no callback. Adding a callback would require modifying the Phase 2 API.
   - Recommendation: For Phase 3, use track-level progress (simple, no Phase 2 changes). If a single-track progress bar is desired, the simplest approach is a spinner (indeterminate progress) rather than a percentage bar. This is adequate because subtitle extraction is fast (typically < 1 second per track). If needed later, packet-level callbacks can be added as a follow-up.

2. **huh v0.8.0 stability on Windows**
   - What we know: huh is built on bubbletea which has broad Windows support. The project is pre-v1 (v0.8.0).
   - What's unclear: Whether edge cases exist with Windows Terminal vs legacy conhost.
   - Recommendation: Test interactive mode on Windows early. Lipgloss handles color degradation; huh handles input. Both are Charmbracelet libraries with Windows CI.

3. **MKV file scanning case sensitivity**
   - What we know: Windows filesystem is case-insensitive but `filepath.Glob` pattern matching depends on OS.
   - What's unclear: Whether `*.mkv` catches `*.MKV` on Windows.
   - Recommendation: On Windows, `filepath.Glob("*.mkv")` will match case-insensitively. On Linux/macOS, scan for both `*.mkv` and `*.MKV` patterns and deduplicate.

## Sources

### Primary (HIGH confidence)
- charmbracelet/huh Context7 library ID: `/charmbracelet/huh` -- Select, MultiSelect, Form API, Option type, validation, accessible mode
- charmbracelet/lipgloss Context7 library ID: `/charmbracelet/lipgloss` -- Style creation, color rendering, text formatting
- charmbracelet/bubbletea Context7 library ID: `/charmbracelet/bubbletea` -- Model/Update/View pattern, key handling
- charmbracelet/bubbles Context7 library ID: `/charmbracelet/bubbles` -- List component, progress component
- schollz/progressbar Context7 library ID: `/schollz/progressbar` -- Default, Set, Finish, DefaultSilent, OptionSetWriter, OptionSetTheme
- spf13/cobra Context7 library ID: `/websites/pkg_go_dev_github_com_spf13_cobra` -- PositionalArgs, flag parsing (verified cobra not needed)
- [huh v2 pkg.go.dev](https://pkg.go.dev/github.com/charmbracelet/huh/v2) -- Confirmed Option type has no Disabled() method
- [huh releases](https://github.com/charmbracelet/huh/releases) -- Latest stable: v0.8.0 (Oct 2024)

### Secondary (MEDIUM confidence)
- [AlecAivazis/survey](https://github.com/AlecAivazis/survey) -- Confirmed archived April 2024, recommends bubbletea
- [pflag pkg.go.dev](https://pkg.go.dev/github.com/spf13/pflag) -- IntSlice, StringSlice, shorthand flags
- [schollz/progressbar GitHub](https://github.com/schollz/progressbar) -- Thread-safe, OS-agnostic progress bar
- [Go forum: CLI flag libraries](https://forum.golangbridge.org/t/which-go-library-is-the-best-for-parsing-command-line-arguments/27680) -- Community consensus on pflag vs cobra vs stdlib
- GitHub code search: huh.NewMultiSelect usage in cli/cli, hatchet-dev/hatchet, BishopFox/sliver

### Tertiary (LOW confidence)
- None -- all findings verified through primary or secondary sources

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH -- All libraries verified via Context7, official docs, and GitHub. Versions confirmed via release pages. Community adoption verified via GitHub code search.
- Architecture: HIGH -- Two-mode dispatch is a standard CLI pattern. The existing codebase API (ExtractTrackToASS, GetMKVInfo, SubtitleTrack) maps cleanly to the required CLI flows.
- Pitfalls: HIGH -- The huh disabled-options limitation was verified directly against pkg.go.dev source. The ExtractTracksToASS early-return behavior was verified by reading pipeline.go source code. Windows color support is well-documented in lipgloss.

**Research date:** 2026-02-17
**Valid until:** 2026-03-17 (30 days -- libraries are stable/mature)
