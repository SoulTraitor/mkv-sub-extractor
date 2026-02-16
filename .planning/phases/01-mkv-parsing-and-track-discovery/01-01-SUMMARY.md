---
phase: 01-mkv-parsing-and-track-discovery
plan: 01
subsystem: parsing
tags: [matroska-go, ebml, mkv, x-text, iso639, subtitle]

# Dependency graph
requires: []
provides:
  - "pkg/mkvinfo package with GetMKVInfo, ClassifyCodec, ResolveLanguageName, FormatFileSize, FormatDuration"
  - "FileInfo, SubtitleTrack, MKVInfo domain types"
  - "Go module with matroska-go and x/text dependencies"
affects: [01-02-PLAN, 02-subtitle-extraction]

# Tech tracking
tech-stack:
  added: [matroska-go v1.2.4, golang.org/x/text v0.34.0]
  patterns: [wrapper-types-over-library, codec-classification-map, language-fallback-chain]

key-files:
  created:
    - go.mod
    - go.sum
    - pkg/mkvinfo/types.go
    - pkg/mkvinfo/codec.go
    - pkg/mkvinfo/codec_test.go
    - pkg/mkvinfo/language.go
    - pkg/mkvinfo/language_test.go
    - pkg/mkvinfo/format.go
    - pkg/mkvinfo/format_test.go
    - pkg/mkvinfo/mkvinfo.go
    - pkg/mkvinfo/mkvinfo_test.go
  modified: []

key-decisions:
  - "Exported ClassifyCodec and ResolveLanguageName as public API for reuse by display layer"
  - "Unknown codecs return raw codec ID string (not panic) with isText=false"
  - "SegmentInfo.Duration uint64 cast directly to time.Duration (nanoseconds match)"

patterns-established:
  - "Wrapper types: FileInfo/SubtitleTrack decouple domain from matroska-go library types"
  - "Codec classification map: static map[string]codecInfo for O(1) lookup"
  - "Language fallback chain: native-script -> English -> raw code -> Unknown"
  - "Error wrapping: fmt.Errorf with %w for all error returns"

# Metrics
duration: 5min
completed: 2026-02-16
---

# Phase 1 Plan 1: MKV Parsing and Track Discovery Summary

**Go mkvinfo package with matroska-go Demuxer integration, 10-codec classification map, native-script language resolution via x/text, and file/duration formatting utilities**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-16T13:24:20Z
- **Completed:** 2026-02-16T13:29:38Z
- **Tasks:** 3
- **Files modified:** 11

## Accomplishments
- Working `pkg/mkvinfo` package that opens MKV files and enumerates all subtitle tracks with full metadata
- Codec classification covering all 10 standard Matroska subtitle codec IDs (SRT, ASS, SSA, WebVTT, USF, PGS, VobSub, DVB, HDMV Text, ARIB) plus graceful fallback for unknowns
- Language resolution converting ISO 639-2 codes to native-script names (eng->English, chi->中文, jpn->日本語, fre->francais, ger->Deutsch)
- File size formatting (e.g. "1.2 GB") and duration formatting (e.g. "1:42:30") with edge case handling
- 18 unit tests all passing, covering happy paths and edge cases (unknown codecs, invalid languages, zero/negative values, non-MKV files)

## Task Commits

Each task was committed atomically:

1. **Task 1: Initialize Go module and create domain types** - `e853a84` (feat)
2. **Task 2: Implement codec classification, language resolution, and formatting utilities** - `1d3172f` (feat)
3. **Task 3: Implement core MKV parsing function using matroska-go** - `2a45f11` (feat)

## Files Created/Modified
- `go.mod` - Go module definition with matroska-go v1.2.4 and x/text v0.34.0 dependencies
- `go.sum` - Dependency checksums
- `pkg/mkvinfo/types.go` - FileInfo, SubtitleTrack, and MKVInfo domain types
- `pkg/mkvinfo/codec.go` - Codec ID classification map and ClassifyCodec function
- `pkg/mkvinfo/codec_test.go` - Tests for all 10 known codecs + unknown + empty
- `pkg/mkvinfo/language.go` - ResolveLanguageName with native-script fallback chain
- `pkg/mkvinfo/language_test.go` - Tests for eng, chi, jpn, fre, ger, und, empty, invalid
- `pkg/mkvinfo/format.go` - FormatFileSize and FormatDuration utilities
- `pkg/mkvinfo/format_test.go` - Tests for various sizes and durations including edge cases
- `pkg/mkvinfo/mkvinfo.go` - GetMKVInfo function integrating all sub-functions via matroska-go Demuxer
- `pkg/mkvinfo/mkvinfo_test.go` - Error path tests for non-existent and non-MKV files

## Decisions Made
- Exported ClassifyCodec and ResolveLanguageName as public API (capitalized names) so the display layer in Plan 02 can reuse them if needed
- Unknown codec IDs return the raw ID string as format type with isText=false -- never panics
- SegmentInfo.Duration (uint64 nanoseconds) is cast directly to time.Duration since Go's time.Duration is also nanoseconds
- FlagDefault is passed through as-is from matroska-go; smart display logic deferred to Plan 02

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered
None

## User Setup Required
None - no external service configuration required.

## Next Phase Readiness
- The `pkg/mkvinfo` package is ready for the display layer (Plan 02) to consume `MKVInfo` structs
- All exported functions are available: `GetMKVInfo`, `ClassifyCodec`, `ResolveLanguageName`, `FormatFileSize`, `FormatDuration`
- matroska-go library confirmed working with Go 1.25.6; API matches research documentation exactly

## Self-Check: PASSED

- All 11 created files verified on disk
- All 3 task commits verified in git history (e853a84, 1d3172f, 2a45f11)

---
*Phase: 01-mkv-parsing-and-track-discovery*
*Completed: 2026-02-16*
