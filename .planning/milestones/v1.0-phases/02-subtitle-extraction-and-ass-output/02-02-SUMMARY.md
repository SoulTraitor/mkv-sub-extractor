---
phase: 02-subtitle-extraction-and-ass-output
plan: 02
subsystem: subtitle
tags: [mkv, extraction, ass, ssa, passthrough, gap-fill, v4-conversion, matroska-go]

# Dependency graph
requires:
  - phase: 01-mkv-parsing
    provides: "MKV track discovery types (SubtitleTrack with CodecID, TrackNumber)"
  - phase: 02-subtitle-extraction-and-ass-output
    plan: 01
    provides: "SubtitleEvent struct, FormatASSTimestamp, ParseASSBlockData"
provides:
  - "ExtractSubtitlePackets: MKV packet extraction filtered by track number with gap-fill"
  - "ApplyGapFill: missing EndTime recovery using next-packet-start or 5s default"
  - "ConvertSSAHeaderToASS: SSA V4 to ASS V4+ header conversion (18->23 style fields)"
  - "WriteASSPassthrough: complete ASS file output from CodecPrivate + SubtitleEvents"
affects: [02-04-extraction-pipeline, 03-cli-interface]

# Tech tracking
tech-stack:
  added: []
  patterns: [gap-fill-strategy, codec-based-dispatch, header-passthrough-with-conversion]

key-files:
  created:
    - "pkg/extract/extract.go"
    - "pkg/extract/extract_test.go"
    - "pkg/assout/ssa_convert.go"
    - "pkg/assout/ssa_convert_test.go"
    - "pkg/assout/writer.go"
    - "pkg/assout/writer_test.go"
  modified: []

key-decisions:
  - "ExtractSubtitlePackets returns a warning string for timestamp sanity check rather than failing"
  - "ApplyGapFill exported separately from ExtractSubtitlePackets for unit testability without real demuxer"
  - "ConvertSSAHeaderToASS returns input unchanged when [V4 Styles] not found (idempotent for V4+)"
  - "WriteASSPassthrough normalizes all line endings to CRLF regardless of input encoding"
  - "Events sorted by StartTime primary, ReadOrder secondary for stable output ordering"

patterns-established:
  - "Codec-based dispatch: codecID string drives SSA-vs-ASS behavior in writer"
  - "Header passthrough: CodecPrivate written verbatim for ASS, converted for SSA, never parsed beyond section detection"
  - "Gap-fill strategy: EndTime <= StartTime treated as missing; filled from next packet or 5s default"
  - "Defensive [Events] handling: case-insensitive section detection, Format line presence check, no duplication"

# Metrics
duration: 5min
completed: 2026-02-16
---

# Phase 2 Plan 2: ASS/SSA Extraction Pipeline Summary

**MKV packet extraction with track filtering and gap-fill, SSA V4-to-V4+ header conversion (18->23 style fields with alignment mapping), and ASS passthrough writer producing valid CRLF ASS files from CodecPrivate + sorted Dialogue lines**

## Performance

- **Duration:** 5 min
- **Started:** 2026-02-16T14:39:58Z
- **Completed:** 2026-02-16T14:45:23Z
- **Tasks:** 2
- **Files created:** 6

## Accomplishments
- ExtractSubtitlePackets reads packets from matroska-go Demuxer filtered by track number, with timestamp sanity checking for IEEE 754 bit-encoding detection
- ApplyGapFill handles missing BlockDuration via gap-fill strategy: EndTime set to next packet's StartTime, or StartTime + 5s for last packet (10 test cases)
- ConvertSSAHeaderToASS performs full SSA V4 to ASS V4+ conversion: ScriptType, section header, Format line (18->23 fields), Style line field remapping (TertiaryColour->OutlineColour, +6 new fields, -AlphaLevel), alignment conversion (SSA->numpad), Marked->Layer in Events (9 alignment + 8 conversion test cases)
- WriteASSPassthrough produces complete ASS files from CodecPrivate header + SubtitleEvent list, with SSA conversion dispatch, [Events] deduplication, CRLF line endings, and timestamp-then-ReadOrder sorting (12 test cases)

## Task Commits

Each task was committed atomically:

1. **Task 1: MKV packet extraction and SSA V4->V4+ conversion** - `9b8e9cb` (feat)
2. **Task 2: ASS passthrough writer** - `6033157` (feat)

## Files Created/Modified
- `pkg/extract/extract.go` - ExtractSubtitlePackets (matroska-go ReadPacket loop with track filter) and ApplyGapFill (missing EndTime recovery)
- `pkg/extract/extract_test.go` - 10 test cases covering gap-fill edge cases (empty, nil, sort-before-fill, multiple missing)
- `pkg/assout/ssa_convert.go` - ConvertSSAHeaderToASS with ScriptType, section header, Format line, Style line remapping, alignment conversion, Marked->Layer
- `pkg/assout/ssa_convert_test.go` - 9 test functions covering full conversion, individual transformations, alignment mapping (9 values), multiple styles, passthrough for V4+
- `pkg/assout/writer.go` - WriteASSPassthrough combining CodecPrivate header with sorted Dialogue lines, CRLF output, [Events] deduplication
- `pkg/assout/writer_test.go` - 12 test cases covering header scenarios, sorting, SSA/ASS dispatch, CRLF, field preservation, case-insensitive detection

## Decisions Made
- ExtractSubtitlePackets returns a warning string (not error) for IEEE 754 timestamp detection, allowing caller to decide severity
- ApplyGapFill exported as standalone function so gap-fill logic can be unit tested without constructing a real matroska-go Demuxer
- ConvertSSAHeaderToASS is idempotent: returns unchanged input if [V4 Styles] not found, safe to call on any header
- WriteASSPassthrough normalizes all line endings (LF, CRLF, CR) to CRLF for consistent ASS output regardless of source encoding
- Secondary sort by ReadOrder ensures stable ordering when multiple events share the same StartTime

## Deviations from Plan

None - plan executed exactly as written.

## Issues Encountered

None.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- ExtractSubtitlePackets ready for integration into extraction pipeline (Plan 04)
- WriteASSPassthrough ready for ASS/SSA track output (Plan 04)
- ConvertSSAHeaderToASS called internally by WriteASSPassthrough; no separate integration needed
- ApplyGapFill handles the STATE.md blocker concern about missing BlockDuration
- No blockers or concerns

## Self-Check: PASSED

All 6 created files verified present on disk. Both commit hashes (9b8e9cb, 6033157) verified in git log.

---
*Phase: 02-subtitle-extraction-and-ass-output*
*Completed: 2026-02-16*
