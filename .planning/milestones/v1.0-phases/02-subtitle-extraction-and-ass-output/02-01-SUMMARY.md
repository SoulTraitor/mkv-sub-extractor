---
phase: 02-subtitle-extraction-and-ass-output
plan: 01
subsystem: subtitle
tags: [ass, srt, timestamp, parsing, tdd, regex, pure-functions]

# Dependency graph
requires:
  - phase: 01-mkv-parsing
    provides: "MKV track discovery types (SubtitleTrack with CodecID, TrackNumber)"
provides:
  - "SubtitleEvent struct shared by ASS passthrough and SRT conversion paths"
  - "FormatASSTimestamp: nanosecond-to-centisecond conversion with rounding"
  - "ParseASSBlockData: Matroska ASS block data field extraction (ReadOrder, Layer, remaining)"
  - "ConvertSRTTagsToASS: HTML b/i/u/font-color to ASS override tags with BGR color reversal"
affects: [02-02-ass-passthrough-writer, 02-03-srt-to-ass-conversion, 02-04-extraction-pipeline]

# Tech tracking
tech-stack:
  added: []
  patterns: [tdd-red-green-refactor, package-level-compiled-regex, integer-only-timestamp-arithmetic]

key-files:
  created:
    - "pkg/subtitle/types.go"
    - "pkg/subtitle/ass.go"
    - "pkg/subtitle/ass_test.go"
    - "pkg/subtitle/srt.go"
    - "pkg/subtitle/srt_test.go"
    - "pkg/assout/timestamp.go"
    - "pkg/assout/timestamp_test.go"
  modified: []

key-decisions:
  - "Integer-only centisecond rounding: (ns + 5_000_000) / 10_000_000 avoids float64 drift"
  - "ParseASSBlockData returns ReadOrder, Layer, and remaining string separately for flexible Dialogue line construction"
  - "SRT tag conversion uses package-level compiled regexes for performance"
  - "7 named colors supported (red, blue, green, yellow, white, cyan, magenta) matching common SRT usage"

patterns-established:
  - "TDD workflow: RED (failing test commit) -> GREEN (implementation commit) per task"
  - "Package-level regex compilation via var = regexp.MustCompile for all SRT patterns"
  - "Pure function design: all functions are stateless string transformations, no I/O"

# Metrics
duration: 3min
completed: 2026-02-16
---

# Phase 2 Plan 1: Core Pure Functions Summary

**ASS timestamp formatting, Matroska block data parsing, and SRT-to-ASS tag conversion with 39 test cases covering rounding boundaries, comma-in-text, BGR color reversal, and named colors**

## Performance

- **Duration:** 3 min
- **Started:** 2026-02-16T14:33:41Z
- **Completed:** 2026-02-16T14:36:30Z
- **Tasks:** 2
- **Files created:** 7

## Accomplishments
- FormatASSTimestamp converts nanoseconds to H:MM:SS.CC with correct centisecond rounding using integer-only arithmetic (11 test cases)
- ParseASSBlockData safely splits Matroska ASS block data into ReadOrder, Layer, and remaining fields using SplitN(9) to preserve commas in Text (7 test cases)
- ConvertSRTTagsToASS maps b/i/u/font-color HTML tags to ASS override tags with RGB-to-BGR hex reversal and 7 named color support (21 test cases)
- SubtitleEvent struct defined as the shared data model for both ASS passthrough and SRT conversion paths

## Task Commits

Each task was committed atomically (TDD RED/GREEN phases):

1. **Task 1: SubtitleEvent types + ASS timestamp formatting**
   - RED: `61ebe57` (test) - SubtitleEvent struct + 11 failing timestamp tests
   - GREEN: `eb77669` (feat) - FormatASSTimestamp implementation, all tests pass

2. **Task 2: ASS block data parsing + SRT tag conversion**
   - RED: `be0e3e8` (test) - 7 ParseASSBlockData + 21 ConvertSRTTagsToASS failing tests
   - GREEN: `d8d11d6` (feat) - Both functions implemented, all 28 tests pass

## Files Created/Modified
- `pkg/subtitle/types.go` - SubtitleEvent struct with Start/End nanoseconds, Layer, Style, Name, Margins, Effect, Text, ReadOrder
- `pkg/assout/timestamp.go` - FormatASSTimestamp: integer arithmetic ns-to-centisecond conversion
- `pkg/assout/timestamp_test.go` - 11 test cases: zero, boundaries, rounding up/down, multi-digit hours
- `pkg/subtitle/ass.go` - ParseASSBlockData: SplitN(9) field extraction with ReadOrder/Layer parsing
- `pkg/subtitle/ass_test.go` - 7 test cases: basic, commas in text, empty text, error cases
- `pkg/subtitle/srt.go` - ConvertSRTTagsToASS: regex-based tag mapping with BGR color reversal
- `pkg/subtitle/srt_test.go` - 21 test cases: b/i/u, hex colors, named colors, case insensitive, nesting, unknown tags

## Decisions Made
- Integer-only centisecond rounding via `(ns + 5_000_000) / 10_000_000` to avoid float64 drift (per research anti-pattern guidance)
- ParseASSBlockData returns `(readOrder, layer, remaining, err)` with layer extracted separately (plan specified 3 return values plus error, but layer is needed independently by the writer for the Dialogue line)
- All regexes compiled at package level (`var = regexp.MustCompile(...)`) for zero per-call allocation
- Named color map covers 7 most common SRT colors; unknown color names cause tag stripping (not error)

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- SubtitleEvent struct ready for use by ASS passthrough writer (Plan 02) and SRT-to-ASS converter (Plan 03)
- FormatASSTimestamp ready for use by ASS output writer (Plan 02, 03)
- ParseASSBlockData ready for ASS extraction pipeline (Plan 02, 04)
- ConvertSRTTagsToASS ready for SRT-to-ASS conversion pipeline (Plan 03)
- No blockers or concerns

## Self-Check: PASSED

All 7 created files verified present on disk. All 4 commit hashes (61ebe57, eb77669, be0e3e8, d8d11d6) verified in git log.

---
*Phase: 02-subtitle-extraction-and-ass-output*
*Completed: 2026-02-16*
