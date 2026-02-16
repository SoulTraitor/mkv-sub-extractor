---
phase: 02-subtitle-extraction-and-ass-output
plan: 03
subsystem: subtitle
tags: [ass, srt, conversion, template, file-naming, collision-handling]

# Dependency graph
requires:
  - phase: 02-subtitle-extraction-and-ass-output
    plan: 01
    provides: "SubtitleEvent struct, FormatASSTimestamp, ConvertSRTTagsToASS"
provides:
  - "WriteSRTAsASS: complete SRT-to-ASS conversion with default Microsoft YaHei template"
  - "defaultASSHeader: generated ASS header constant for SRT tracks"
  - "GenerateOutputPath: output file naming with language codes and collision handling"
  - "sanitizeFileName: filesystem-safe track name sanitization"
affects: [02-04-extraction-pipeline]

# Tech tracking
tech-stack:
  added: []
  patterns: [sort-copy-pattern, collision-resolution-chain, regex-filename-sanitization]

key-files:
  created:
    - "pkg/assout/template.go"
    - "pkg/assout/srt_writer.go"
    - "pkg/assout/srt_writer_test.go"
    - "pkg/output/naming.go"
    - "pkg/output/naming_test.go"
  modified: []

key-decisions:
  - "Default ASS style: Microsoft YaHei 22pt, 1920x1080 PlayRes, bottom-center alignment, white text with black outline"
  - "Collision resolution chain: base path -> track name -> sequence numbers (.2, .3, etc.)"
  - "WriteSRTAsASS copies events slice before sorting to avoid mutating caller's data"
  - "sanitizeFileName limits to 50 chars and falls back to 'track' if empty after sanitization"

patterns-established:
  - "Copy-then-sort pattern: copy input slice before sorting to preserve caller's order"
  - "Collision resolution chain: try unique path -> named variant -> numbered fallback"
  - "CRLF line endings throughout all ASS output (string constants + Dialogue lines)"

# Metrics
duration: 4min
completed: 2026-02-16
---

# Phase 2 Plan 3: SRT-to-ASS Conversion and Output Naming Summary

**WriteSRTAsASS converter with default Microsoft YaHei 22pt template producing valid ASS files, plus GenerateOutputPath with language-code naming and three-tier collision resolution**

## Performance

- **Duration:** 4 min
- **Started:** 2026-02-16T14:39:59Z
- **Completed:** 2026-02-16T14:44:00Z
- **Tasks:** 2
- **Files created:** 5

## Accomplishments
- WriteSRTAsASS produces complete valid ASS files from SRT subtitle events with default header, sorted Dialogue lines, SRT tag conversion, and \N newline handling
- Default ASS template uses Microsoft YaHei 22pt, 1920x1080 PlayRes, white text with black outline, bottom-center alignment, CRLF line endings
- GenerateOutputPath generates {basename}.{lang}.ass paths with ISO 639-2 codes, collision handling via track name then sequence numbers
- sanitizeFileName strips unsafe characters, collapses underscores, limits to 50 chars, and falls back to "track" for empty results
- 24 total tests across both packages (9 WriteSRTAsASS + 15 naming/sanitization)

## Task Commits

Each task was committed atomically:

1. **Task 1: Default ASS template and SRT-to-ASS writer** - `d20147d` (feat)
2. **Task 2: Output file naming with collision handling** - `b9f0378` (feat)

## Files Created/Modified
- `pkg/assout/template.go` - Default ASS header constant (Script Info + V4+ Styles with Microsoft YaHei) and events Format line
- `pkg/assout/srt_writer.go` - WriteSRTAsASS: writes header, sorts events, converts tags, outputs Dialogue lines with CRLF
- `pkg/assout/srt_writer_test.go` - 9 tests: two events, sorting, tag conversion, newlines, empty list, CRLF, style/layer, non-mutation
- `pkg/output/naming.go` - GenerateOutputPath with language code naming and collision handling; sanitizeFileName helper
- `pkg/output/naming_test.go` - 15 tests: basic naming, empty/und language, directory paths, collisions, triple collision, sanitization edge cases

## Decisions Made
- Default ASS style: Microsoft YaHei 22pt (user decision for font, middle of 20-24 range for size), PlayRes 1920x1080, bottom-center alignment (Alignment=2), white text (&H00FFFFFF) with black outline (&H00000000) and semi-transparent shadow (&H80000000)
- Collision resolution uses three-tier chain: base path first, then sanitized track name, then sequence numbers starting at 2
- WriteSRTAsASS copies the events slice before sorting to avoid mutating the caller's data (defensive copy pattern)
- sanitizeFileName limits to 50 characters and returns "track" as fallback when all characters are stripped

## Deviations from Plan

### Auto-fixed Issues

**1. [Rule 1 - Bug] Fixed sort test matching "Second" in header's "SecondaryColour"**
- **Found during:** Task 1 (WriteSRTAsASS tests)
- **Issue:** Test used `strings.Index(output, "Second")` which matched "SecondaryColour" in the ASS header before the Dialogue line, causing false sort failure
- **Fix:** Changed test to search for specific Dialogue line prefixes (`Dialogue: 0,0:00:01.00,` and `Dialogue: 0,0:00:10.00,`) instead of bare text strings
- **Files modified:** pkg/assout/srt_writer_test.go
- **Verification:** All 9 tests pass after fix
- **Committed in:** d20147d (Task 1 commit)

---

**Total deviations:** 1 auto-fixed (1 bug in test)
**Impact on plan:** Test assertion corrected for accuracy. No scope creep.

## Issues Encountered

None beyond the test assertion fix documented above.

## User Setup Required

None - no external service configuration required.

## Next Phase Readiness
- WriteSRTAsASS ready for use by the extraction pipeline (Plan 04) to handle S_TEXT/UTF8 (SRT) tracks
- GenerateOutputPath ready for both ASS passthrough (Plan 02) and SRT conversion (Plan 03) extraction paths
- defaultASSHeader and defaultEventsFormat are package-private constants consumed only by WriteSRTAsASS
- No blockers or concerns

## Self-Check: PASSED

All 5 created files verified present on disk. Both commit hashes (d20147d, b9f0378) verified in git log.

---
*Phase: 02-subtitle-extraction-and-ass-output*
*Completed: 2026-02-16*
