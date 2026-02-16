# Phase 1: MKV Parsing and Track Discovery - Context

**Gathered:** 2026-02-16
**Status:** Ready for planning

<domain>
## Phase Boundary

Parse MKV (EBML/Matroska) containers and list all subtitle tracks with metadata. Users can point the tool at an MKV file and see what subtitle tracks are available. Extraction and conversion are Phase 2; CLI interaction modes are Phase 3.

</domain>

<decisions>
## Implementation Decisions

### Track listing format
- Numbered list format (not table), e.g. `[1] English (SRT) "Commentary" [default]`
- Each track shows: track number, language (human-readable name like "English"/"中文"), format type (SRT/ASS/SSA), track name (if present), codec ID (e.g. S_TEXT/UTF8), default/forced flags
- If language tag is absent, display "Unknown"

### Non-text subtitle handling
- Image subtitle tracks (PGS/VobSub/DVB) are shown in the listing but clearly marked as "not extractable (image subtitle)"
- Image subtitle tracks are not selectable — their numbers are skipped in selection
- If an MKV has only image subtitles and no text subtitles: display all tracks with their markers, then print a message: "未找到可提取的文本字幕轨道"

### File metadata display
- Before listing tracks, show basic file information:
  - File name
  - File size (human-readable, e.g. "1.2 GB")
  - Duration (e.g. "1:42:30")
  - Number of subtitle tracks
- This gives the user context about the file before making a selection

### Claude's Discretion
- Exact formatting and spacing of the track listing
- How to display codec ID alongside format type (inline vs separate line)
- Duration parsing approach from Matroska Segment/Info elements
- Color or styling of terminal output (if any)

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

*Phase: 01-mkv-parsing-and-track-discovery*
*Context gathered: 2026-02-16*
