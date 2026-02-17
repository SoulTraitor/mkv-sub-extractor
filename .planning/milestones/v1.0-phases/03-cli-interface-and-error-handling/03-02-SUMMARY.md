---
phase: 03-cli-interface-and-error-handling
plan: 02
subsystem: cli
tags: [orchestration, progress-bar, run-dispatch, main-entry, human-verification]

# Dependency graph
requires:
  - phase: 03-cli-interface-and-error-handling
    plan: 01
    provides: "CLIError types, Config/ParseFlags, PickMKVFile, SelectTracks"
provides:
  - "Run() two-mode CLI dispatch (interactive vs scriptable)"
  - "ExtractWithProgress() continue-on-failure extraction with shared collision map"
  - "printCompletionSummary() lipgloss-styled extraction results"
  - "Minimal main.go entry point"
affects: []

# Tech tracking
tech-stack:
  added: []
  patterns: [two-mode dispatch, continue-on-failure extraction, shared existingPaths for batch collision, post-submit validation]

key-files:
  created:
    - pkg/cli/run.go
    - pkg/cli/progress.go
  modified:
    - cmd/mkv-sub-extractor/main.go
    - pkg/cli/interactive.go
    - pkg/extract/pipeline.go
    - pkg/subtitle/ass.go
    - pkg/assout/writer.go
    - pkg/assout/template.go
    - pkg/assout/writer_test.go

key-decisions:
  - "ExtractTrackToASSShared shares existingPaths across batch to prevent same-language filename collisions"
  - "Empty Layer/ReadOrder in ASS block data defaults to 0 (not error)"
  - "Always run ConvertSSAHeaderToASS regardless of codec ID (handles ASS tracks with V4 headers)"
  - "SRT default font size 58pt at 1080p PlayRes (matches player default SRT rendering scale)"
  - "Esc key added as quit binding; huh Quit is form-level so Description used for Esc hint"
  - "huh Validate moved to post-submit to avoid showing error on initial render"
  - "progressbar ClearOnFinish removed due to Windows rendering artifact"

patterns-established:
  - "Shared existingPaths map pattern for batch collision-aware output naming"
  - "Post-submit validation pattern for huh forms (check after Run, not via Validate callback)"

requirements-completed: [CLI-05]

# Metrics
duration: 8min
completed: 2026-02-17
---

# Phase 3 Plan 02: CLI Orchestration Summary

**Two-mode CLI dispatch, continue-on-failure extraction with progress bar, completion summary, and main entry point. Includes 6 bug fixes from human verification testing.**

## Performance

- **Duration:** ~8 min (code) + human testing & bug fixes
- **Completed:** 2026-02-17
- **Tasks:** 2 (1 auto + 1 human verification)
- **Files modified:** 9

## Accomplishments
- Run() with two-mode dispatch: runInteractive (file picker -> parse -> select -> extract) and runScriptable (--track direct extraction)
- ExtractWithProgress() with shared existingPaths for collision-aware batch naming and simplified progress bar
- printCompletionSummary() with lipgloss green/red styled per-track results and count summary
- Minimal main.go entry point delegating to cli.Run()
- 6 bug fixes from human verification: same-language collision, Esc key support, keyboard hints, empty Layer parsing, V4 Styles conversion, SRT font size

## Bug Fixes from Human Testing

1. **Same-language filename collision**: Two "chi" tracks overwrote each other. Fix: Added ExtractTrackToASSShared with shared existingPaths map across batch.
2. **No Esc exit**: huh default KeyMap only binds Ctrl+C. Fix: Added quitKeyMap() with Esc binding.
3. **Missing keyboard hints**: No Esc hint visible. Fix: Description shows "Esc cancel" (huh Quit is form-level, not in field KeyBinds).
4. **Empty Layer field**: ASS block data with empty Layer caused strconv.Atoi("") error. Fix: Default empty ReadOrder/Layer to 0.
5. **V4 Styles not converted for ASS codec**: WriteASSPassthrough only converted for S_TEXT/SSA. Fix: Always call ConvertSSAHeaderToASS (idempotent for V4+).
6. **SRT font too small**: Fontsize 22 at 1080p PlayRes (~2% screen height). Fix: Changed to 58 (~5.4%, matching player defaults). Margins increased.
7. **Progress bar ClearOnFinish artifact**: Windows terminal showed 50% instead of 100%. Fix: Removed ClearOnFinish option.

## Task Commits

1. **Task 1: Create orchestration layer** - `4259985` (feat)
2. **Bug fixes from human testing** - `e758fdd` (fix) + `9e16cb1` (fix)

## Files Created/Modified
- `pkg/cli/run.go` — Run(), runInteractive(), runScriptable(), printCompletionSummary(), exitCodeFromResults()
- `pkg/cli/progress.go` — TrackResult, ExtractWithProgress() with shared existingPaths and simplified progress bar
- `cmd/mkv-sub-extractor/main.go` — Minimal entry point: os.Exit(cli.Run())
- `pkg/cli/interactive.go` — quitKeyMap() with Esc, Description hints, post-submit validation
- `pkg/extract/pipeline.go` — ExtractTrackToASSShared() accepting shared existingPaths
- `pkg/subtitle/ass.go` — Empty Layer/ReadOrder handling in ParseASSBlockData
- `pkg/assout/writer.go` — Unconditional ConvertSSAHeaderToASS call
- `pkg/assout/template.go` — Fontsize 22→58, margins 10→20/30
- `pkg/assout/writer_test.go` — New test for ASS codec with V4 Styles

## Deviations from Plan
- Plan noted that per-track existingPaths maps were "acceptable" for simplicity. Human testing proved this wrong (same-language collision). Added ExtractTrackToASSShared with shared map.
- 6 additional bug fixes not in original plan, discovered during human verification.

## Self-Check: PASSED

All files verified present. Tool builds, all tests pass, human verification complete.

---
*Phase: 03-cli-interface-and-error-handling*
*Completed: 2026-02-17*
