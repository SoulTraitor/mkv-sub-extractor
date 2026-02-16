# Phase 2: Subtitle Extraction and ASS Output - Context

**Gathered:** 2026-02-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Extract subtitle track data from MKV files and produce valid ASS format output files. ASS/SSA tracks use passthrough mode (preserving original styles from CodecPrivate). SRT tracks are converted to ASS with generated headers and tag mapping. CLI interaction and file selection are Phase 3.

</domain>

<decisions>
## Implementation Decisions

### ASS passthrough strategy
- CodecPrivate data output verbatim as ASS file header — preserve all styles, fonts, resolution settings without modification
- SSA (old V4) tracks are converted to ASS V4+ format for modern player compatibility
- Dialogue lines sorted by start timestamp in output (not MKV storage order)
- [Events] Format line: extract from CodecPrivate if present; generate standard default if absent

### SRT to ASS conversion style
- Default font: Microsoft YaHei (微软雅黑) for all languages
- Default style: white text + black outline, font size ~20-24, bottom-center aligned — similar to mpv/VLC default subtitle rendering
- HTML tag mapping: `<b>` → `{\b1}`, `<i>` → `{\i1}`, `<u>` → `{\u1}`, `<font color>` → `{\c&H...&}`
- Multi-line subtitles: SRT newlines converted to `\N` (hard line break) in ASS

### Timestamp precision and alignment
- Millisecond to centisecond conversion: round (四舍五入), not truncate
- Negative or zero timestamps: clamp to 0:00:00.00 (ensure output always valid)
- Missing BlockDuration: end time set to next subtitle's start time (gap-fill strategy)

### Output file naming and encoding
- Naming format: `{video_basename}.{lang_code}.ass` (e.g., movie.eng.ass, movie.chi.ass) using ISO 639-2 three-letter codes
- Name collision (same language, multiple tracks): use track name to differentiate if available; fall back to sequence number if no track name
- File encoding: UTF-8 without BOM
- Line endings: Claude's Discretion

### Claude's Discretion
- Line ending style (CRLF vs LF) — choose based on ASS ecosystem conventions
- Exact font size within the 20-24 range for SRT → ASS default style
- ASS PlayResX/PlayResY values for generated SRT → ASS headers
- SSA → ASS V4+ conversion specifics (field mapping, style format differences)
- How to handle edge cases in SRT HTML tag parsing (nested tags, malformed HTML)

</decisions>

<specifics>
## Specific Ideas

No specific requirements — open to standard approaches

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 02-subtitle-extraction-and-ass-output*
*Context gathered: 2026-02-16*
