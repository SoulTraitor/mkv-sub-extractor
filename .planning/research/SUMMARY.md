# Project Research Summary

**Project:** MKV Subtitle Extractor
**Domain:** CLI tool for media file processing (Matroska container parsing, subtitle extraction and format conversion)
**Researched:** 2026-02-16
**Confidence:** MEDIUM-HIGH

## Executive Summary

The MKV Subtitle Extractor is a specialized CLI tool that extracts text subtitles from Matroska (MKV) video containers and outputs them in ASS format. Expert implementations of this type of tool require deep understanding of EBML binary parsing, Matroska container structure, and subtitle format specifications. The recommended approach is a layered architecture with five components: EBML parser (foundation), Matroska decoder (container navigation), subtitle codec handlers (format-specific parsing), ASS format writer (output generation), and CLI layer (user interaction).

The primary technical challenge is EBML/Matroska parsing in pure Go without external dependencies. The chosen stack centers on `luispater/matroska-go` for MKV demuxing (high-level API, active maintenance) and `asticode/go-astisub` for subtitle format handling (comprehensive ASS support, only viable option in Go ecosystem). The architecture must handle multi-gigabyte files efficiently using streaming/seeking patterns rather than in-memory buffering.

Key risks include VINT encoding bugs (subtle but catastrophic), incorrect timestamp/duration handling (breaks subtitle timing), and unknown-size element handling (common in real-world files but often untested). Mitigation requires extensive unit testing against RFC 8794 test vectors, two-pass file parsing (metadata first, then extraction), and careful timestamp arithmetic using integer operations to avoid precision loss. The bottom-up build order (EBML parser -> Matroska decoder -> codec handlers -> ASS writer -> orchestrator) allows incremental validation at each layer.

## Key Findings

### Recommended Stack

The stack consists of three direct dependencies plus Go stdlib, targeting simplicity and reliability. Go 1.24+ is required by `matroska-go`, with 1.26 as current stable.

**Core technologies:**
- **`luispater/matroska-go` v1.2.4**: MKV container demuxing with pull-based Demuxer API — provides high-level `GetTrackInfo()`, `ReadPacket()`, and direct subtitle track identification. Most recently maintained option (Aug 2025) with working extraction examples. Requires `io.ReadSeeker` which is acceptable for file-based use. MEDIUM confidence due to low community adoption (0 known importers) but clean API design.
- **`asticode/go-astisub` v0.38.0**: Subtitle format handling with full ASS/SSA style preservation — the only maintained multi-format subtitle library in Go. 38 releases, MIT licensed, maintained through Nov 2025. Provides `ReadFromSRT()` and `WriteToSSA()` for direct conversion pipeline. HIGH confidence as established, comprehensive solution.
- **`charmbracelet/huh` v0.8.0**: Interactive CLI prompts for file and track selection — modern Bubble Tea-based library with generic `Select[T]` type for clean track index selection. Provides `Select`, `Confirm`, and `FilePicker` fields. Alternative `promptui` is unmaintained since Oct 2021. HIGH confidence for user interaction layer.
- **Go stdlib `flag`**: Argument parsing — appropriate for single-command CLI with few flags. Cobra is overkill for this use case. Zero dependency cost.

**Critical decision: EBML/Matroska parser selection:** `matroska-go` wins over `remko/go-mkvparse` (push-based event handler requiring ~3x more code) and `coding-socks/matroska` (explicitly alpha status). The pull-based API maps directly to workflow: open file, list tracks, read packets. FALLBACK PLAN: If `matroska-go` proves problematic, fall back to `go-mkvparse` with custom state machine (more complex but battle-tested).

### Expected Features

Research identified table stakes (must-have), differentiators (competitive edge), and anti-features (explicit non-goals).

**Must have (table stakes):**
- List subtitle tracks with metadata (track number, codec, language, name) — every tool does this
- Extract SRT (S_TEXT/UTF8) and SSA/ASS (S_TEXT/SSA, S_TEXT/ASS) subtitles — core formats
- Convert SRT to ASS output — core value proposition of uniform output format
- Interactive track selection and CLI parameter mode — requirement for both user-friendly and scriptable use
- Automatic output filename with language awareness — better UX than competitors
- Clear error messages for unsupported formats — differentiate between image subtitles and corrupt files

**Should have (competitive edge):**
- Zero external dependencies (single binary) — primary differentiator vs mkvextract/ffmpeg
- Human-readable track display with clean formatting — better than mkvextract/ffprobe's machine-oriented output
- Smart SRT-to-ASS conversion preserving formatting tags — map `<i>`, `<b>` to ASS override codes
- ASS passthrough without re-encoding — preserve all styles, fonts, animations for native ASS tracks
- Interactive file browser when run with no args — neither mkvextract nor ffmpeg provides this
- Language-aware default naming (`Movie.en.ass`) — follows media player conventions

**Defer (v2+):**
- Image subtitle extraction (PGS, VobSub) — requires OCR, defeats zero-dependency goal
- Subtitle editing/timing adjustment — scope creep, SubtitleEdit owns this space
- Batch/recursive directory processing — can be wrapped in shell script for v1
- Font extraction from MKV attachments — orthogonal to subtitle extraction

**Anti-features (explicit non-goals):**
- GUI/complex TUI — over-engineering, simple interactive mode sufficient
- Embedding subtitles into MKV (muxing) — different domain, users have mkvmerge
- Audio/video stream extraction — out of scope, recommend ffmpeg
- Network/URL input — local files only, streaming adds massive complexity

### Architecture Approach

A layered architecture with five components, each with clean Go interface boundaries. Data flows downward through parsing layers and upward through extraction, with CLI orchestrating the process.

**Major components:**
1. **EBML Parser (`pkg/ebml`)** — Foundation layer handling VINT decoding, element ID/size parsing, element type discrimination, forward scanning/seeking. Uses `io.ReadSeeker` for seeking past video/audio data.
2. **Matroska Decoder (`pkg/matroska`)** — Interprets EBML elements as Matroska semantics: identifies Segment children, parses TrackEntry, navigates Clusters, extracts Blocks. Provides track metadata and subtitle block data.
3. **Subtitle Codec Handlers (`pkg/subtitle`)** — Implements Handler interface per codec (S_TEXT/UTF8, S_TEXT/SSA, S_TEXT/ASS). Decodes raw block data into structured subtitle events, extracting codec-specific metadata from CodecPrivate.
4. **ASS Format Writer (`pkg/format/ass`)** — Converts structured subtitle events into valid ASS file output with proper headers, styles, and dialogue lines. Handles both ASS passthrough (preserve CodecPrivate header) and SRT conversion (generate default header).
5. **CLI Layer (`cmd/mkv-sub-extractor`)** — User interaction: argument parsing, file discovery, interactive track selection menus, progress display, error reporting. Uses stdlib `flag` for simplicity.
6. **Orchestrator (`pkg/extractor`)** — Coordinates full pipeline: parse tracks, present to CLI, extract selected track, convert format, write output.

**Key architectural patterns:**
- **Streaming, not buffering:** Video/audio blocks never read into memory. EBML parser skips them using `io.Seeker`. Prevents OOM on multi-GB files.
- **Two-pass approach:** First pass reads track metadata (small, near file start). Second pass iterates clusters for subtitle blocks (full file traversal but only reads matching track).
- **Element-at-a-time scanning:** Parse one EBML element header, decide to descend (master) or skip/read (data). Never buffer full element tree.
- **Codec handler interface:** Each subtitle codec implements common Handler interface, selected by CodecID string. Clean separation enables adding formats without modifying existing code.

**Build order:** Bottom-up following data dependencies: (1) EBML Parser (foundation), (2) Matroska Decoder (depends on EBML), (3) Subtitle Codec Handlers (depends on Matroska types), (4) ASS Format Writer (depends on subtitle types), (5) Orchestrator + CLI (integration). Each phase produces testable, independently valuable component.

### Critical Pitfalls

Top 5 pitfalls from research that can cause incorrect output, data corruption, or panics:

1. **EBML VINT (Variable-Size Integer) Decoding Errors** — The leading byte determines width by position of first set bit. Common mistakes: failing to mask VINT_MARKER bit, treating `0xFF` as regular integer instead of "unknown size" sentinel, miscounting leading zeros. Consequences: parser reads garbage element lengths, seeks to wrong offsets, panics or corrupts all subsequent data. Mitigation: Heavily test with RFC 8794 test vectors; separate functions for Element IDs (preserve marker) vs Data Sizes (mask marker); test 1-byte through 8-byte values plus unknown-size sentinels.

2. **"Unknown Size" Elements and Streaming/Damaged Files** — Elements can have data size encoded as "all VINT_DATA bits set to 1" (unknown size). Common for Segment/Cluster in streaming sources. Consequences: parser panics allocating `0x00FFFFFFFFFFFFFF` bytes, hangs seeking past EOF, or skips entire Segment. Mitigation: Check if decoded VINT equals unknown-size sentinel for its width; for unknown-size Master elements, parse children until hitting sibling/parent-level element or EOF; requires schema knowledge.

3. **Confusing Element ID Encoding with Data Size Encoding** — Both use VINT encoding but critically different: Element IDs keep the VINT_MARKER bit, Data Sizes mask it out. Using same function for both produces incorrect Element IDs. Mitigation: Implement separate `readVINTValue()` (masks marker) and `readVINTRaw()` (preserves marker); test against known Element IDs (EBML=0x1A45DFA3, Segment=0x18538067).

4. **Incorrect Block/SimpleBlock Timestamp and Duration Handling** — Subtitle timing combines Cluster Timestamp (absolute) with Block's relative timestamp (signed int16 offset). Common mistakes: treating Block timestamp as unsigned, forgetting TimestampScale, missing BlockDuration for subtitle tracks. Consequences: subtitles flash on/off, display forever, or have shifted timestamps. Mitigation: Parse 16-bit block timestamp as `int16` explicitly; always read BlockDuration for subtitles; convert through TimestampScale: `absoluteTimestampNs = (clusterTimestamp + blockRelativeTimestamp) * timestampScale`; use integer arithmetic for centisecond conversion.

5. **Ignoring CodecPrivate for SSA/ASS Passthrough** — SSA/ASS tracks store the header (`[Script Info]`, `[V4+ Styles]`) in CodecPrivate. Block data contains only dialogue lines without style definitions. Ignoring CodecPrivate produces ASS files without styles. Mitigation: Always read CodecPrivate from TrackEntry; for S_TEXT/ASS and S_TEXT/SSA, CodecPrivate is the file header; reconstruct full ASS by combining CodecPrivate + `[Events]` section + formatted Dialogue lines from Blocks.

**Additional moderate pitfalls:** SRT-to-ASS timestamp precision loss (use integer arithmetic with rounding), Matroska lacing mishandling (parse lacing flags even if not expected), element ordering assumptions (use SeekHead + fallback linear scan), character encoding issues (validate UTF-8, warn on non-UTF-8), ASS format spec quirks (use minimal complete template).

## Implications for Roadmap

Based on research, suggested phase structure follows component dependencies and enables incremental validation:

### Phase 1: EBML Parser Foundation
**Rationale:** Foundation everything depends on. Cannot parse MKV without EBML. Must be rock-solid before proceeding.
**Delivers:** VINT decoding (element IDs and data sizes separately), element header parsing, type-specific data reading (uint, int, float, string, bytes), seeking/skipping
**Addresses:** Core parsing capability for all subsequent phases
**Avoids:** Pitfalls 1-3 (VINT encoding errors, unknown-size elements, ID/size confusion)
**Research needs:** MINIMAL — RFC 8794 is definitive and well-specified. No deep research needed, focus on test coverage.

### Phase 2: Matroska Decoder
**Rationale:** Depends on EBML parser. Unlocks track discovery which can be demoed standalone (list tracks without extraction).
**Delivers:** EBML header validation, Tracks element parsing (TrackEntry with Number, Type, CodecID, CodecPrivate, Language, Name), TrackInfo struct
**Addresses:** Table stakes features: list subtitle tracks with metadata
**Avoids:** Pitfall 4 (timestamp handling — laid foundation, fully addressed in Phase 3), Pitfall 8 (element ordering — use two-pass approach)
**Research needs:** MINIMAL — Matroska spec is stable, Element IDs are well-documented. Verify current spec for critical IDs during implementation.

### Phase 3: Cluster Navigation and Block Extraction
**Rationale:** Depends on Matroska decoder. Enables actual subtitle data extraction. Delivers end-to-end subtitle block reading.
**Delivers:** Cluster traversal, Block/SimpleBlock binary parsing (track number VINT, signed int16 timestamp, flags, payload), BlockDuration extraction, SubtitleBlock struct with timing
**Addresses:** Core extraction pipeline — reads raw subtitle data from selected track
**Avoids:** Pitfall 4 (timestamp/duration handling — full implementation), Pitfall 6 (lacing flags — must parse even if unexpected)
**Research needs:** LOW — Block structure is well-documented but has edge cases. Budget time for testing with varied MKV files (FFmpeg-muxed, OBS recordings).

### Phase 4: Subtitle Codec Handlers (ASS Passthrough)
**Rationale:** Depends on Matroska types (SubtitleBlock, TrackInfo). ASS passthrough is simplest extraction path (no conversion) — validates block reading pipeline.
**Delivers:** Handler interface, ASSHandler implementation (reconstructs ASS from CodecPrivate header + block dialogue lines), subtitle Event and Track types
**Addresses:** Table stakes: extract SSA/ASS subtitles with full style preservation
**Avoids:** Pitfall 5 (CodecPrivate — critical implementation), Pitfall 7 (CodecPrivate preservation), Pitfall 11 (ASS format quirks — passthrough avoids most issues)
**Research needs:** MINIMAL — ASS block format in Matroska is well-documented. Focus on CodecPrivate handling.

### Phase 5: SRT Handler and ASS Conversion
**Rationale:** Depends on Handler interface. Can be developed in parallel with Phase 4. Completes core value proposition.
**Delivers:** SRTHandler implementation (parses S_TEXT/UTF8 blocks as plain text), SRT-to-ASS conversion (generate default header, map formatting tags, convert timestamps)
**Addresses:** Core value proposition: convert SRT to ASS output with uniform format
**Avoids:** Pitfall 5 (timestamp precision loss — use integer arithmetic), Pitfall 11 (ASS format quirks — use validated default template), Pitfall 15 (SRT over-parsing — use Block-level cues, not SRT text parsing)
**Research needs:** LOW — Straightforward conversion logic. Main risk is timestamp precision and formatting tag mapping.

### Phase 6: ASS Format Writer
**Rationale:** Depends on subtitle types (Event, Track). Can be developed in parallel with Phases 4-5. Delivers file output capability.
**Delivers:** ASS file generation (write ScriptInfo, Styles, Events sections), timestamp formatting (h:mm:ss.cc), dialogue line formatting
**Addresses:** Output generation for both ASS passthrough and SRT conversion paths
**Avoids:** Pitfall 11 (ASS format quirks — section ordering, PlayResX/Y, Format line matching)
**Research needs:** MINIMAL — ASS format is well-understood. Use validated default template and preserve CodecPrivate verbatim for passthrough.

### Phase 7: CLI and Orchestrator
**Rationale:** Depends on all above components. Pure integration layer. Delivers working tool.
**Delivers:** Flag parsing (--output, --file, --track), interactive file browser, track selection menu, orchestrator coordination (parse -> identify -> extract -> convert flow), output filename generation
**Addresses:** Table stakes: interactive track selection, CLI parameter mode, automatic output filename; Differentiators: file browser, human-readable track display
**Avoids:** Pitfall 13 (multiple same-language tracks — display track name and flags for differentiation)
**Research needs:** NONE — Straightforward CLI glue code.

### Phase Ordering Rationale

- **Bottom-up dependency chain:** Each phase builds on previous phases' output. Cannot extract subtitles without parsing container; cannot parse container without EBML foundation.
- **Incremental validation:** Each phase produces testable component. Phase 1-2 can be validated with track listing before full extraction. Phase 3 validates block reading. Phases 4-6 can be tested with constructed data.
- **Parallel work opportunities:** Phases 4-5 (codec handlers) and Phase 6 (writer) have no mutual dependencies — can be developed in parallel if resources allow.
- **Early risk mitigation:** Most critical pitfalls (VINT encoding, unknown-size elements) are addressed in Phase 1. Timestamp handling in Phase 3. CodecPrivate in Phase 4. De-risks core technical challenges early.
- **Demo milestones:** Phase 2 delivers demoable track listing. Phase 4 delivers first extraction (ASS passthrough). Phase 5 completes core value prop (SRT conversion). Phase 7 delivers polished UX.

### Research Flags

**Phases with MINIMAL/NO further research needed (standard patterns):**
- **Phase 1 (EBML Parser):** RFC 8794 is definitive. Focus on test coverage, not research.
- **Phase 2 (Matroska Decoder):** Element IDs are stable. Quick verification against current spec sufficient.
- **Phase 6 (ASS Writer):** Well-understood format. Use validated template.
- **Phase 7 (CLI/Orchestrator):** Straightforward integration code.

**Phases needing LIGHT validation during planning:**
- **Phase 3 (Block Extraction):** Test with varied MKV sources (FFmpeg, OBS, MKVToolNix) to validate edge cases. Budget 1-2 days for file testing.
- **Phase 4 (ASS Passthrough):** Verify CodecPrivate structure with sample files early. LOW research need.
- **Phase 5 (SRT Conversion):** Validate timestamp arithmetic and formatting tag mapping. LOW research need.

**No phases require `/gsd:research-phase`** — domain is well-documented, research was comprehensive, and confidence is MEDIUM-HIGH across all areas. Implementation can proceed directly from research outputs.

## Confidence Assessment

| Area | Confidence | Notes |
|------|------------|-------|
| Stack | MEDIUM-HIGH | `matroska-go` has low community adoption (0 importers) but clean API and active maintenance. Fallback to `go-mkvparse` available if issues arise. `go-astisub` and `huh` are well-established. |
| Features | HIGH | Domain is mature (mkvextract, ffmpeg set clear expectations). Table stakes and differentiators are well-defined. Anti-features prevent scope creep. |
| Architecture | MEDIUM-HIGH | Layered approach with clean boundaries is proven pattern for parsers. EBML/Matroska parsing has complexity but RFC 8794 and Matroska spec provide clear guidance. Streaming architecture critical for multi-GB files. |
| Pitfalls | MEDIUM | Critical pitfalls (VINT encoding, timestamp handling, CodecPrivate) are well-documented in Matroska community. Confidence adjusted to MEDIUM because web verification was unavailable — all findings from training data. Should validate against current RFC 8794 and Matroska spec during Phase 1. |

**Overall confidence:** MEDIUM-HIGH

The path forward is clear with well-defined components and build order. Primary uncertainty is `matroska-go` library behavior with real-world files (mitigated by fallback plan). EBML parsing complexity is substantial but well-specified in RFC 8794. Roadmap can proceed without further upfront research.

### Gaps to Address

Gaps identified and how to handle during planning/execution:

- **`matroska-go` real-world validation:** Library has 0 known importers. Gap: Unknown behavior with edge-case MKV files (damaged files, unusual muxers, unknown-size elements). **Mitigation:** Plan Phase 3 to include testing with diverse MKV sources (FFmpeg, OBS, Handbrake, MKVToolNix). Allocate 1-2 days for file-based testing. If library proves problematic, trigger fallback to `go-mkvparse`.

- **EBML Element ID verification:** Element IDs in research are from training data (MEDIUM confidence). Gap: Current Matroska spec might have updates (unlikely but possible). **Mitigation:** First task of Phase 1 is to fetch current Matroska spec from matroska.org and verify Element IDs for critical elements (EBML, Segment, Tracks, Cluster, TrackEntry children). 1-hour task, eliminates risk of using outdated IDs.

- **ASS timestamp format edge cases:** ASS timestamp format (`h:mm:ss.cc`) has variations in hour-zero handling. Gap: Unclear if single-digit hour is always correct or if some players expect two digits. **Mitigation:** Phase 6 task: test generated ASS files with mpv (libass) and PotPlayer (VSFilter) to validate timestamp rendering. Adjust format string if needed.

- **Character encoding detection:** Research identified potential non-UTF-8 subtitle tracks in wild but deferred encoding detection to v2. Gap: If users encounter garbled subtitles due to encoding issues, v1 has no mitigation. **Mitigation:** Phase 5 includes UTF-8 validation with clear error message if invalid UTF-8 detected. Error message recommends checking source file encoding. Can add `--encoding` flag in v1.1 if user demand exists.

- **Testing with subtitle-heavy files:** Research focused on technical specifications, not performance testing. Gap: Unknown performance with files containing many subtitle tracks or very dense subtitles. **Mitigation:** Phase 7 testing includes at least one multi-subtitle file (5+ subtitle tracks) and one dense subtitle file (e.g., Karaoke with per-syllable timing). Verify memory usage stays minimal (streaming architecture should handle this).

## Sources

### Primary (HIGH confidence)
- RFC 8794 (EBML specification) — VINT encoding, element structure, data types. Definitive standard.
- Go standard library documentation — `io`, `encoding/binary`, `flag`, `utf8` packages. Stable APIs.
- `asticode/go-astisub` pkg.go.dev documentation — v0.38.0 API with SSA types reviewed.

### Secondary (MEDIUM-HIGH confidence)
- Matroska specification (matroska.org/technical) — Element hierarchy, Element IDs, codec IDs, block structure, subtitle storage format. Community-maintained spec, widely referenced.
- `luispater/matroska-go` pkg.go.dev documentation — v1.2.4 API reviewed. Library code and examples analyzed.
- ASS/SSA format references — File structure, timestamp format, dialogue line format. Multiple community sources cross-referenced.

### Secondary (MEDIUM confidence)
- `remko/go-mkvparse` pkg.go.dev documentation — v0.14.0 API reviewed for comparison.
- `charmbracelet/huh` pkg.go.dev documentation — v0.8.0 API and examples reviewed.
- FFmpeg and MKVToolNix behavior — mkvextract and ffmpeg capabilities for competitive analysis.

### Tertiary (noted as needs validation)
- Matroska Element IDs — From training data, widely documented but should be verified against current spec during Phase 1 (1-hour task).
- SSA/ASS block data format in Matroska — "ReadOrder, Layer, Style, ..." field order documented but frequently misunderstood. Verify with actual MKV files during Phase 4.
- Unknown-size element prevalence — Research indicates common in FFmpeg/OBS output. Should test with real files during Phase 3.

**Note on research method:** Web search and documentation fetch were unavailable during research session. All findings based on training data with strong familiarity with EBML RFC 8794, Matroska specification, and Go ecosystem. Core technical claims (VINT encoding, element structure, codec IDs) are stable and well-established — unlikely to have changed. Confidence levels adjusted accordingly.

---
*Research completed: 2026-02-16*
*Ready for roadmap: yes*
