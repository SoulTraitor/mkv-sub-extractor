# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-16)

**Core value:** User can quickly and reliably extract subtitles from MKV files and get ASS format output
**Current focus:** Phase 3 - CLI Interface and Error Handling

## Current Position

Phase: 3 of 3 (CLI Interface and Error Handling) -- IN PROGRESS
Plan: 1 of 2 in current phase (plan 01 complete)
Status: Executing Phase 3
Last activity: 2026-02-17 -- Plan 03-01 complete

Progress: [████████████░░░] 83%

## Performance Metrics

**Velocity:**
- Total plans completed: 7
- Average duration: 6 min
- Total execution time: 0.73 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-mkv-parsing | 2 | 20 min | 10 min |
| 02-subtitle-extraction-and-ass-output | 4 | 20 min | 5 min |

| 03-cli-interface-and-error-handling | 1 | 4 min | 4 min |

**Recent Trend:**
- Last 5 plans: 3m, 4m, 5m, 8m, 4m
- Trend: Consistent (averaging 5 min/plan)

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
- 02-02: ApplyGapFill exported separately for unit testability; EndTime <= StartTime treated as missing
- 02-02: WriteASSPassthrough normalizes all line endings to CRLF; sorts by StartTime then ReadOrder
- 02-02: ConvertSSAHeaderToASS is idempotent (returns unchanged for V4+ input)
- 02-03: Default ASS style: Microsoft YaHei 22pt, 1920x1080 PlayRes, bottom-center, white text with black outline
- 02-03: Collision resolution chain: base path -> track name -> sequence numbers (.2, .3)
- 02-03: WriteSRTAsASS copies events slice before sorting (defensive copy pattern)
- 02-04: ExtractTrackToASS is the single public entry point; extractTrackToASS (internal) accepts shared existingPaths for batch collision tracking
- 02-04: packetsToEvents splits ASS remaining field into exactly 7 parts (Style through Text) with SplitN
- 02-04: Partial output files cleaned up on write error (os.Remove)
- 02-04: ExtractSubtitlePackets warning silently ignored at pipeline level
- 03-01: Exit code scheme: 1=general, 2=file, 3=track, 4=extraction
- 03-01: Error codes use category prefixes (E0x=file, E1x=track, E2x=extraction)
- 03-01: CLIError.Error() returns Title only; Format() gives full rustc-style output
- 03-01: MKV file deduplication on case-insensitive FS via ToLower map key
- 03-01: Image tracks displayed as faint lipgloss above multi-select, not inline

### Pending Todos

None.

### Blockers/Concerns

- ~~Research notes matroska-go has 0 known importers; fallback to go-mkvparse if issues arise during Phase 1~~ RESOLVED: matroska-go v1.2.4 working correctly with Go 1.25.6, API matches documentation
- ~~EBML Element IDs from research training data should be verified against current Matroska spec early in Phase 1~~ RESOLVED: TypeSubtitle=17, TrackInfo fields, SegmentInfo.Duration all verified against actual library source
- **matroska-go Duration bug**: Library reads Matroska float Duration element with ReadUInt(), returning IEEE 754 bits. Workaround applied in mkvinfo.go. May affect other float elements in Phase 2 (e.g., packet timestamps — verify during extraction).

## Session Continuity

Last session: 2026-02-17
Stopped at: Completed 03-01-PLAN.md
Resume file: .planning/phases/03-cli-interface-and-error-handling/03-02-PLAN.md
