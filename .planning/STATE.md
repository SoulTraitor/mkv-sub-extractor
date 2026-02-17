# Project State

## Project Reference

See: .planning/PROJECT.md (updated 2026-02-16)

**Core value:** User can quickly and reliably extract subtitles from MKV files and get ASS format output
**Current focus:** All phases complete

## Current Position

Phase: 3 of 3 (CLI Interface and Error Handling) -- COMPLETE
Plan: 2 of 2 in current phase (all complete)
Status: All 3 phases complete
Last activity: 2026-02-17 -- Phase 3 human testing passed, bug fixes committed

Progress: [████████████████] 100%

## Performance Metrics

**Velocity:**
- Total plans completed: 8
- Average duration: 6 min
- Total execution time: ~0.8 hours

**By Phase:**

| Phase | Plans | Total | Avg/Plan |
|-------|-------|-------|----------|
| 01-mkv-parsing | 2 | 20 min | 10 min |
| 02-subtitle-extraction-and-ass-output | 4 | 20 min | 5 min |
| 03-cli-interface-and-error-handling | 2 | 8 min | 4 min |

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
- 02-03: Default ASS style: Microsoft YaHei 58pt, 1920x1080 PlayRes, bottom-center, white text with black outline
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
- 03-02: ExtractTrackToASSShared shares existingPaths across batch to prevent same-language collision
- 03-02: Empty Layer/ReadOrder in ASS block data defaults to 0 (not error)
- 03-02: Always run ConvertSSAHeaderToASS regardless of codec ID (handles ASS tracks with V4 headers)
- 03-02: SRT default font size 58pt at 1080p (matches player default SRT rendering scale)
- 03-02: Esc key added as quit binding; validation moved to post-submit

### Pending Todos

None.

### Blockers/Concerns

- ~~Research notes matroska-go has 0 known importers; fallback to go-mkvparse if issues arise during Phase 1~~ RESOLVED: matroska-go v1.2.4 working correctly with Go 1.25.6, API matches documentation
- ~~EBML Element IDs from research training data should be verified against current Matroska spec early in Phase 1~~ RESOLVED: TypeSubtitle=17, TrackInfo fields, SegmentInfo.Duration all verified against actual library source
- ~~**matroska-go Duration bug**: Library reads Matroska float Duration element with ReadUInt(), returning IEEE 754 bits. Workaround applied in mkvinfo.go.~~ RESOLVED: Workaround applied and verified across all phases

## Session Continuity

Last session: 2026-02-17
Stopped at: Project complete — all 3 phases done
Resume file: N/A
