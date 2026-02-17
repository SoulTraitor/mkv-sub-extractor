---
phase: 03-cli-interface-and-error-handling
plan: 01
subsystem: cli
tags: [pflag, huh, lipgloss, progressbar, charmbracelet, error-handling, interactive-tui]

# Dependency graph
requires:
  - phase: 02-subtitle-extraction-and-ass-output
    provides: "ExtractTrackToASS pipeline, SubtitleTrack type"
provides:
  - "CLIError structured error type with rustc-style formatted output"
  - "Config struct and ParseFlags() with pflag-based flag parsing"
  - "ValidateConfig() for pre-parse input validation"
  - "PickMKVFile() interactive file selector using huh"
  - "SelectTracks() interactive multi-select with image-track context"
affects: [03-02-PLAN]

# Tech tracking
tech-stack:
  added: [spf13/pflag v1.0.5, charmbracelet/huh v0.8.0, charmbracelet/lipgloss v1.1.0, schollz/progressbar/v3 v3.19.0]
  patterns: [rustc-style error formatting, factory function error constructors, lipgloss styled terminal output, huh form-based interactive selection]

key-files:
  created:
    - pkg/cli/errors.go
    - pkg/cli/flags.go
    - pkg/cli/interactive.go
  modified:
    - go.mod
    - go.sum

key-decisions:
  - "Exit code scheme: 1=general, 2=file, 3=track, 4=extraction"
  - "Error codes follow category prefixes: E0x=file, E1x=track, E2x=extraction"
  - "CLIError.Error() returns Title only; Format() gives full rustc-style output"
  - "MKV file deduplication on case-insensitive FS using ToLower map key"
  - "Image tracks shown as faint lipgloss text above multi-select, not inline"

patterns-established:
  - "Factory function pattern: ErrXxx() constructors return *CLIError with preset fields"
  - "lipgloss style variables at package level for consistent terminal coloring"
  - "huh form pattern: build options, create form, run, extract value"

requirements-completed: [CLI-01, CLI-02, CLI-03, CLI-04, CLI-06, ERRH-01, ERRH-02, ERRH-03, ERRH-04]

# Metrics
duration: 4min
completed: 2026-02-17
---

# Phase 3 Plan 01: CLI Foundation Summary

**Structured CLI error types with rustc-style lipgloss formatting, pflag-based flag parsing with -t/-o/-q/-v, and charmbracelet/huh interactive file picker and track multi-selector**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-17T03:55:06Z
- **Completed:** 2026-02-17T03:59:00Z
- **Tasks:** 2
- **Files modified:** 5

## Accomplishments
- CLIError type with 10 factory functions covering file, track, and extraction error categories with rustc-style Format() output
- Config struct with ParseFlags() using pflag for --track/-t, --output/-o, --quiet/-q, --verbose/-v and positional MKV path
- ValidateConfig() checks file existence, .mkv extension, readability, directory conflicts, and mutually exclusive flags
- PickMKVFile() scans current directory and presents huh Select menu for MKV file selection
- SelectTracks() presents huh MultiSelect with grayed-out image-track context and minimum-1 validation

## Task Commits

Each task was committed atomically:

1. **Task 1: Create structured error types and flag parsing** - `7f4fb3b` (feat)
2. **Task 2: Create interactive file picker and track multi-selector** - `c3e2203` (feat)

## Files Created/Modified
- `pkg/cli/errors.go` - CLIError type with rustc-style Format(), 10 factory functions, lipgloss-colored output
- `pkg/cli/flags.go` - Config struct, ParseFlags() with pflag, ValidateConfig() for pre-parse validation
- `pkg/cli/interactive.go` - PickMKVFile() with huh Select, SelectTracks() with huh MultiSelect, formatTrackOption()
- `go.mod` - Added pflag, huh, lipgloss, progressbar/v3 dependencies
- `go.sum` - Updated with all transitive dependency hashes

## Decisions Made
- Exit code scheme: 1=general, 2=file error, 3=track error, 4=extraction error (per user decision from CONTEXT.md)
- Error codes use category prefixes (E0x for file errors, E1x for track errors, E2x for extraction errors) for machine-readable categorization
- CLIError.Error() returns just the Title for standard error chaining; Format() provides the full rustc-style multi-line output
- MKV file deduplication uses strings.ToLower map key to handle case-insensitive Windows filesystem overlap between *.mkv and *.MKV globs
- Image-based tracks displayed as faint lipgloss text above the multi-select rather than as disabled options within it (cleaner separation)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
- `go mod tidy` failed due to Go proxy network issues downloading test dependencies for charmbracelet packages. This does not affect the build since all required dependencies are already in go.sum from the initial `go get` commands. The project compiles and vets cleanly.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- All pkg/cli building blocks are ready for Plan 03-02 (orchestration layer)
- Errors, flags, and interactive prompts provide the complete API surface the orchestration wires together
- The progressbar/v3 dependency is installed and ready for progress reporting in Plan 03-02

## Self-Check: PASSED

All files verified present. All commits verified in git log.

---
*Phase: 03-cli-interface-and-error-handling*
*Completed: 2026-02-17*
