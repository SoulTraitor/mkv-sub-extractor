---
phase: 02-subtitle-extraction-and-ass-output
plan: 04
subsystem: subtitle
tags: [pipeline, integration, extraction, ass, srt, passthrough, demuxer, matroska-go]

# Dependency graph
requires:
  - phase: 02-subtitle-extraction-and-ass-output
    plan: 02
    provides: "ExtractSubtitlePackets, ApplyGapFill, WriteASSPassthrough, ConvertSSAHeaderToASS"
  - phase: 02-subtitle-extraction-and-ass-output
    plan: 03
    provides: "WriteSRTAsASS, GenerateOutputPath"
provides:
  - "ExtractTrackToASS: single-track extraction from MKV to ASS file (primary public API)"
  - "ExtractTracksToASS: batch extraction with shared collision-aware naming"
  - "packetsToEvents: codec-specific raw packet to SubtitleEvent conversion"
affects: [03-cli-interface]

# Tech tracking
tech-stack:
  added: []
  patterns: [integration-pipeline, shared-collision-map, codec-switch-dispatch]

key-files:
  created:
    - "pkg/extract/pipeline.go"
    - "pkg/extract/pipeline_test.go"
  modified:
    - "cmd/mkv-sub-extractor/main.go"

key-decisions:
  - "ExtractTrackToASS is the single public entry point; internal extractTrackToASS accepts shared existingPaths for batch collision tracking"
  - "packetsToEvents splits ASS remaining field into exactly 7 parts (Style through Text) with SplitN"
  - "SRT packets get zero margins, empty effect, Default style, and Layer 0"
  - "Partial output files are cleaned up on write error (os.Remove after failed write)"
  - "ExtractSubtitlePackets warning is silently ignored at pipeline level (caller can log if needed)"

patterns-established:
  - "Integration pipeline pattern: single function orchestrates open -> demux -> extract -> convert -> name -> write"
  - "Shared map for batch operations: existingPaths map passed through for cross-track collision awareness"
  - "Codec dispatch via switch: codecID string determines passthrough vs conversion path"

# Metrics
duration: 8min
completed: 2026-02-16
---

# Phase 2 Plan 4: Extraction Pipeline Integration Summary

**ExtractTrackToASS and ExtractTracksToASS public API orchestrating full MKV-to-ASS pipeline with codec-aware dispatch (ASS passthrough / SRT conversion), verified with real MKV files in media player**

## Performance

- **Duration:** ~8 min (execution) + human verification
- **Started:** 2026-02-16T14:50:23Z
- **Completed:** 2026-02-16 (human verification approved)
- **Tasks:** 2 (1 auto + 1 checkpoint)
- **Files created:** 2 (pipeline.go, pipeline_test.go)
- **Files modified:** 1 (main.go for verification)

## Accomplishments
- ExtractTrackToASS provides the single-function public API that Phase 3 CLI will call: takes MKV path, track info, and output dir, produces a valid ASS file
- ExtractTracksToASS enables batch extraction from the same MKV with shared collision-aware naming map
- ASS/SSA tracks use passthrough mode preserving original styles from CodecPrivate
- SRT tracks are automatically converted to ASS format with generated Microsoft YaHei headers
- Real MKV file verification confirmed: output ASS files render correctly in media player, timestamps accurate, styles preserved
- Pipeline correctly orchestrates all Phase 2 packages: mkvinfo, extract, subtitle, assout, output

## Task Commits

Each task was committed atomically:

1. **Task 1: Create extraction pipeline integrating all Phase 2 packages** - `d878f3a` (feat)
2. **Task 2 (prep): Temporary main.go extraction for pipeline verification** - `362328d` (chore)

## Files Created/Modified
- `pkg/extract/pipeline.go` - Public API: ExtractTrackToASS (single), ExtractTracksToASS (batch), packetsToEvents (codec dispatch)
- `pkg/extract/pipeline_test.go` - Compilation checks, signature verification, error handling tests for non-existent and invalid MKV files
- `cmd/mkv-sub-extractor/main.go` - Temporary extraction call added for human verification (Phase 3 will replace with proper CLI)

## Decisions Made
- ExtractTrackToASS delegates to internal extractTrackToASS with a fresh existingPaths map; ExtractTracksToASS shares the map across tracks for correct collision handling
- packetsToEvents uses strings.SplitN with limit 7 to parse ASS remaining field, allowing commas in the Text portion
- SRT packets mapped to SubtitleEvent with Layer=0, Style="Default", zero margins, empty Name/Effect
- Partial output files cleaned up on write error to avoid leaving corrupt files on disk
- Warning from ExtractSubtitlePackets (gap-fill applied count) silently ignored at pipeline level -- Phase 3 can surface via logging

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- Phase 2 is COMPLETE: all 4 plans executed successfully
- ExtractTrackToASS and ExtractTracksToASS are the public API surface that Phase 3 (CLI) will call
- All packages compile clean (go build, go vet, go test all pass)
- Real MKV verification confirmed correct ASS output for both ASS passthrough and SRT conversion paths
- Ready for Phase 3: CLI interface with track selection, batch extraction, and user interaction

## Self-Check: PASSED

All 2 created files verified present on disk. Both commit hashes (d878f3a, 362328d) verified in git log.

---
*Phase: 02-subtitle-extraction-and-ass-output*
*Completed: 2026-02-16*
