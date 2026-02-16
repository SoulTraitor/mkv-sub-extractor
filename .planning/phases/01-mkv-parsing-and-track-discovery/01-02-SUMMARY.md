---
phase: 01-mkv-parsing-and-track-discovery
plan: 02
subsystem: display
tags: [display, cli, formatting, track-listing]

# Dependency graph
requires: [01-01-PLAN]
provides:
  - "Display formatting: FormatTrackLine, FormatFileInfoHeader, FormatTrackListing, ShouldShowDefault"
  - "CLI entry point: cmd/mkv-sub-extractor/main.go"
affects: [02-subtitle-extraction, 03-cli-interface]

# Tech tracking
tech-stack:
  added: []
  patterns: [smart-default-display, numbered-track-listing, chinese-messages]

key-files:
  created:
    - pkg/mkvinfo/display.go
    - pkg/mkvinfo/display_test.go
    - cmd/mkv-sub-extractor/main.go
    - .gitignore
  modified:
    - pkg/mkvinfo/mkvinfo.go

key-decisions:
  - "ShouldShowDefault suppresses [default] flag when ALL tracks have it (Matroska spec quirk)"
  - "matroska-go Duration bug: ReadUInt on float element returns raw IEEE 754 bits; reinterpret via Float32frombits/Float64frombits"
  - "Image subtitle tracks shown but marked 'not extractable (image subtitle)'"

patterns-established:
  - "Smart flag display: ShouldShowDefault prevents noisy flags when all tracks default"
  - "Library bug workaround: reinterpret uint64 bits as float for Matroska Duration element"
  - "Chinese user-facing messages for no-text-subtitles case"

# Metrics
duration: 15min
completed: 2026-02-16
---

# Phase 1 Plan 2: Display Formatting and CLI Entry Point Summary

**Track listing display, file metadata header, smart default flag logic, CLI entry point, and matroska-go Duration float bug fix**

## Performance

- **Duration:** ~15 min (including human verification and bug fix iteration)
- **Completed:** 2026-02-16
- **Tasks:** 2 (1 auto + 1 human-verify checkpoint)
- **Files created:** 4
- **Files modified:** 1

## Accomplishments
- Display formatting matching user's locked decisions: `[N] Language (Format) "Name" CodecID [flags]`
- File metadata header: file name, human-readable size, duration (H:MM:SS), subtitle track count
- ShouldShowDefault smart logic suppresses noisy [default] flags when all tracks have FlagDefault=true
- Image subtitle tracks clearly marked with `-- not extractable (image subtitle)` suffix
- Chinese message "未找到可提取的文本字幕轨道" for MKVs with only image subtitles
- Minimal CLI entry point accepting file path argument
- Discovered and fixed matroska-go Duration bug (ReadUInt on float element)
- Human verification passed with real MKV file

## Task Commits

1. **Task 1: Display formatting and main entry point** - `87dc39d` (feat)
2. **Duration bug fix** - `4f1fd75` (fix) - discovered during human verification

## Files Created/Modified
- `pkg/mkvinfo/display.go` - FormatTrackLine, FormatFileInfoHeader, FormatTrackListing, ShouldShowDefault
- `pkg/mkvinfo/display_test.go` - 8 tests covering text/image tracks, flags, headers, edge cases
- `cmd/mkv-sub-extractor/main.go` - Minimal CLI: accepts file path, prints track listing
- `.gitignore` - Binary and IDE exclusions
- `pkg/mkvinfo/mkvinfo.go` - **Modified**: Fixed Duration parsing (IEEE 754 float reinterpretation)

## Decisions Made
- ShouldShowDefault returns false when all tracks have IsDefault=true (prevents noisy output)
- matroska-go Duration bug workaround: check if uint64 fits in 32 bits → Float32frombits, else Float64frombits, then multiply by TimecodeScale

## Deviations from Plan
- **Duration bug**: Plan assumed `time.Duration(segInfo.Duration)` was correct (based on library comment "duration in nanoseconds"). Human verification revealed the library reads a float EBML element with ReadUInt(), returning raw IEEE 754 bits. Required investigation of matroska-go source and a two-iteration fix.

## Issues Encountered
- **matroska-go Duration bug**: The library's `SegmentInfo.Duration` comment says "duration in nanoseconds" but `parser.go` reads the Matroska Duration element (which is a float per spec) using `element.ReadUInt()` instead of `element.ReadFloat()`. This returns the raw IEEE 754 bit pattern as a uint64. For 4-byte floats (common in MKV), this gives ~1.2 billion nanoseconds ≈ 1 second regardless of actual duration. Fixed by reinterpreting bits as float before scaling.

## User Setup Required
None

## Next Phase Readiness
- Phase 1 complete: users can run `mkv-sub-extractor.exe <file.mkv>` and see all subtitle tracks
- `FormatTrackListing` output verified against real MKV files
- Duration bug workaround is stable and handles both float32 and float64 Duration elements
- Ready for Phase 2 (Subtitle Extraction and ASS Output)

## Self-Check: PASSED

- All files verified on disk
- All tests pass (`go test ./...`)
- `go vet ./...` clean
- Human verification passed with real MKV file
- Binary builds and runs correctly

---
*Phase: 01-mkv-parsing-and-track-discovery*
*Completed: 2026-02-16*
