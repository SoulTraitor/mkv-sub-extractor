# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-16)

**Core value:** User can quickly and reliably extract subtitles from MKV files and get ASS format output
**Current focus:** Phase 2 - Subtitle Extraction and ASS Output

## Current Position

Phase: 2 of 3 (Subtitle Extraction and ASS Output)
Plan: 1 of 4 in current phase
Status: Executing
Last activity: 2026-02-16 -- Completed 02-01 core pure functions (TDD, 39 tests)

Progress: [████░░░░░░] 40%

## Performance Metrics

**Velocity:**
- Total plans completed: 3
- Average duration: 8 min
- Total execution time: 0.38 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-mkv-parsing | 2 | 20 min | 10 min |
| 02-subtitle-extraction-and-ass-output | 1 | 3 min | 3 min |

**Recent Trend:**
- Last 5 plans: 5m, 15m, 3m
- Trend: Accelerating (pure function TDD plans are fast)

*Updated after each plan completion*

## Accumulated Context

### Decisions

Decisions are logged in PROJECT.md Key Decisions table.
Recent decisions affecting current work:

- Roadmap: 3-phase structure (Parse -> Extract -> CLI) derived from requirement categories and dependency chain
- 01-01: Exported ClassifyCodec and ResolveLanguageName as public API for reuse by display layer
- 01-01: Unknown codecs return raw codec ID string with isText=false (never panic)
- ~~01-01: SegmentInfo.Duration uint64 cast directly to time.Duration (nanoseconds match)~~ SUPERSEDED by duration fix below
- 01-02: matroska-go Duration bug — ReadUInt on float element returns raw IEEE 754 bits; must reinterpret via Float32frombits/Float64frombits then multiply by TimecodeScale
- 01-02: ShouldShowDefault suppresses [default] when all tracks have it
- 02-01: Integer-only centisecond rounding (ns + 5_000_000) / 10_000_000 avoids float64 drift
- 02-01: ParseASSBlockData returns (readOrder, layer, remaining, err) for flexible Dialogue construction
- 02-01: Package-level compiled regexes for SRT tag conversion performance
- 02-01: 7 named SRT colors supported; unknown colors cause tag stripping (not error)

### Pending Todos

None.

### Blockers/Concerns

- ~~Research notes matroska-go has 0 known importers; fallback to go-mkvparse if issues arise during Phase 1~~ RESOLVED: matroska-go v1.2.4 working correctly with Go 1.25.6, API matches documentation
- ~~EBML Element IDs from research training data should be verified against current Matroska spec early in Phase 1~~ RESOLVED: TypeSubtitle=17, TrackInfo fields, SegmentInfo.Duration all verified against actual library source
- **matroska-go Duration bug**: Library reads Matroska float Duration element with ReadUInt(), returning IEEE 754 bits. Workaround applied in mkvinfo.go. May affect other float elements in Phase 2 (e.g., packet timestamps — verify during extraction).

## Session Continuity

Last session: 2026-02-16
Stopped at: Completed 02-01-PLAN.md (core pure functions)
Resume file: .planning/phases/02-subtitle-extraction-and-ass-output/02-02-PLAN.md
