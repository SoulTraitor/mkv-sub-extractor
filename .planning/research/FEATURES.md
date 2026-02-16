# Feature Landscape

**Domain:** MKV Subtitle Extraction CLI Tool (Pure Go)
**Researched:** 2026-02-16
**Research mode:** Ecosystem -- based on training data analysis of mkvextract, ffmpeg, SubtitleEdit, and Matroska specification. Web verification was unavailable; confidence levels adjusted accordingly.

## Table Stakes

Features users expect from any subtitle extraction tool. Missing any of these and users will reach for mkvextract or ffmpeg instead.

| Feature | Why Expected | Complexity | Notes |
|---------|--------------|------------|-------|
| List subtitle tracks with metadata | Every tool does this. Users need to know what's in the file before extracting. | Medium | Must show: track number, codec ID, language tag (BCP 47 / legacy 3-letter), track name. Requires EBML parsing of TrackEntry elements. |
| Extract SRT (S_TEXT/UTF8) subtitles | SRT is the most common text subtitle format in MKV files. | Medium | Stored as BlockGroup/SimpleBlock data. Each cue is a block with timestamp + duration + UTF-8 text payload. |
| Extract SSA/ASS (S_TEXT/SSA, S_TEXT/ASS) subtitles | Second most common text subtitle format. Also the project's output target. | Medium | Header/style info stored in CodecPrivate. Cue data stored in blocks with "ReadOrder, Layer, Style, Name, MarginL, MarginR, MarginV, Effect, Text" fields (comma-separated, no Dialogue: prefix). |
| Convert SRT to ASS output | Core value proposition of this tool -- uniform ASS output. | Medium | Requires generating ASS header with default styles, mapping SRT timestamps to ASS timestamps (h:mm:ss.cc format), wrapping text in Dialogue lines. SRT italic/bold tags (\<i\>, \<b\>) should map to ASS override tags ({\\i1}, {\\b1}). |
| Interactive track selection | Project requirement. Users want to see tracks and pick one. | Low | Display numbered list, read stdin. Must handle single-track auto-selection gracefully. |
| CLI parameter mode | Essential for scripting and automation. mkvextract and ffmpeg both support this. | Low | Accept file path as positional arg. Track selection via flag (e.g., `--track 2`). |
| Automatic output filename | Users expect sensible defaults. mkvextract requires explicit output; ffmpeg requires explicit output. Doing this better is easy. | Low | Pattern: `{video_name}.{language}.ass` or `{video_name}.{track_name}.ass`. Handle collision with numeric suffix. |
| `--output` directory flag | Standard CLI convention. In project requirements. | Low | Create directory if not exists. Validate path is writable. |
| Error handling with clear messages | Bad file, corrupt container, unsupported codec -- all must produce helpful errors, not stack traces. | Low | "Track 3 is PGS (image-based subtitle) -- only text subtitles are supported" is far better than a generic parse error. |
| Cross-platform single binary | Go's strength. Users expect `go install` or download a binary and run it. | Low | No CGO, no external deps. Already a project constraint. |

## Differentiators

Features that set this tool apart from mkvextract/ffmpeg. Not expected, but valued.

| Feature | Value Proposition | Complexity | Notes |
|---------|-------------------|------------|-------|
| Zero external dependencies | mkvextract requires MKVToolNix install. ffmpeg requires ffmpeg install. This tool is a single binary drop. | N/A (architectural) | This is the core differentiator. Every feature should protect this constraint. |
| Interactive file browser | When run with no args, scan current directory for MKV files and let user pick. Neither mkvextract nor ffmpeg does this. | Low | Scan cwd for `*.mkv` files, display numbered list. If only one MKV, auto-select with confirmation. |
| Human-readable track display | mkvextract's `mkvmerge --identify` output is JSON or machine-oriented. ffmpeg's `ffprobe` output is verbose. Show a clean, formatted table of subtitle tracks. | Low | Example: `[1] English (SRT)  [2] Chinese (ASS) "Commentary"  [3] Japanese (ASS)` |
| Smart SRT-to-ASS conversion with style preservation | Basic tools just dump raw text. This tool can map SRT formatting tags to ASS override codes, generating a proper ASS header with sensible defaults. | Medium | Map `<i>` to `{\i1}`, `<b>` to `{\b1}`, `<font color="#RRGGBB">` to `{\c&HBBGGRR&}`. Generate [Script Info], [V4+ Styles] with a readable default style (good font, good size). |
| Passthrough for ASS tracks | If the source is already ASS, write it directly without re-encoding. Preserves all styling, fonts references, animations. | Low | Reconstruct from CodecPrivate (header) + block data (dialogue lines). No conversion needed. |
| Language-aware default naming | Auto-name output files using the track's language metadata. Useful for media players that auto-load subtitles by language suffix. | Low | `Movie.en.ass`, `Movie.zh.ass`, `Movie.ja.ass`. Follows common media player conventions (Plex, Kodi, Jellyfin naming). |
| Subtitle track summary / preview | Show first few lines of a subtitle track before extracting, so user can verify they picked the right one. | Medium | Requires reading initial blocks, parsing timestamps, displaying preview text. Useful when multiple tracks have same language. |
| Encoding detection and normalization | Some older MKV files have subtitle tracks with non-UTF8 encoding (especially CJK content). Detect and convert to UTF-8. | High | Would need encoding detection heuristics (BOM sniffing, chardet-style analysis). Edge case but real pain point for CJK users. |
| Progress indication for large files | MKV files can be 1-50+ GB. Seeking through a large file to find subtitle blocks should show progress. | Low | Simple progress bar or percentage. Only needed for files where seeking takes noticeable time. |
| `--list` mode (no extraction) | Just show tracks and exit. Useful for scripting: pipe track info to other tools. | Low | `mkv-sub-extractor --list movie.mkv` outputs structured track info. |

## Anti-Features

Features to explicitly NOT build. These add complexity without matching the tool's value proposition.

| Anti-Feature | Why Avoid | What to Do Instead |
|--------------|-----------|-------------------|
| Image subtitle extraction (PGS, VobSub, DVB) | Requires OCR (Tesseract or similar), massive complexity, defeats zero-dependency goal. PGS is bitmap data, not text. | Print clear message: "Track N is image-based (PGS). This tool extracts text subtitles only. Use mkvextract + other tools for image subtitles." |
| Subtitle editing / timing adjustment | Scope creep. SubtitleEdit and Aegisub own this space. | Extract faithfully, let users edit in dedicated tools. |
| Batch/recursive directory processing | v1 scope is single file. Batch adds complexity around error handling, partial failures, output naming conflicts. | Users can wrap in shell: `for f in *.mkv; do mkv-sub-extractor "$f"; done`. Consider for v2. |
| GUI / TUI with full curses interface | Over-engineering. Interactive mode with simple stdin/stdout is sufficient and more scriptable. | Keep it simple: numbered lists + user input. No dependency on terminal UI libraries. |
| Embedding subtitles into MKV (muxing) | Writing MKV files is an entirely different beast from reading them. EBML writing is complex. | Out of scope. Users have mkvmerge for muxing. |
| Audio/video stream extraction | Different domain entirely. This tool is subtitle-focused. | Recommend ffmpeg or mkvextract for A/V extraction. |
| Network/URL input (HTTP, streaming) | MKV files are local. Supporting streaming protocols adds massive complexity. | Accept local file paths only. |
| Conversion to formats other than ASS | Project spec says ASS output. Supporting SRT/VTT/etc. output fragments the value proposition. | If SRT source, convert to ASS. If ASS source, pass through. One output format = simpler code, simpler testing. |
| Font extraction from MKV attachments | MKV can embed fonts as attachments. Extracting them is useful but orthogonal to subtitle extraction. | Could be a v2 feature if demanded. Mention in docs that fonts are not extracted. |
| Forced/default track flag handling | Matroska has FlagForced and FlagDefault on tracks. Automatically selecting based on these flags adds complexity. | Show the flags in track listing so users can decide. Don't auto-select. |

## Feature Dependencies

```
EBML Parser ──> Track Enumeration ──> Track Display (interactive list)
                     |                       |
                     v                       v
              Block Reading ──> SRT Parser ──> ASS Converter ──> File Writer
                     |
                     v
              ASS Passthrough ──────────────────────────────> File Writer
```

Detailed dependency chain:

1. **EBML/Matroska parser** -- foundation. Everything depends on reading the container.
   - Parse EBML header, Segment, Tracks, Clusters
   - Required by: every other feature

2. **Track enumeration** -- depends on EBML parser
   - Read TrackEntry elements, extract CodecID, Language, Name, TrackNumber, TrackUID
   - Required by: track display, track selection, extraction

3. **Track display** -- depends on track enumeration
   - Format track info for human consumption
   - Required by: interactive mode, `--list` mode

4. **Block reading / cue extraction** -- depends on EBML parser + track selection
   - Navigate Clusters, read BlockGroup/SimpleBlock for selected track
   - Extract timestamp, duration, payload data
   - Required by: SRT parser, ASS passthrough

5. **SRT parser** -- depends on block reading
   - Parse S_TEXT/UTF8 block payloads (sequential cue text)
   - Required by: ASS converter

6. **ASS converter** -- depends on SRT parser
   - Generate ASS header, convert SRT cues to Dialogue lines
   - Required by: file writer (SRT source path)

7. **ASS passthrough** -- depends on block reading
   - Reconstruct ASS file from CodecPrivate + block payloads
   - Required by: file writer (ASS source path)

8. **File writer** -- depends on ASS converter or ASS passthrough
   - Write final .ass file to disk with proper encoding (UTF-8 BOM for ASS compatibility)

## MVP Recommendation

Prioritize in this order:

1. **EBML/Matroska container parsing** -- without this, nothing works. Parse enough of the spec to read track metadata and subtitle blocks. Do NOT implement full spec -- only elements needed for subtitle extraction.

2. **Track enumeration and display** -- list subtitle tracks with language, format, and name. This is the first user-visible feature and validates the parser.

3. **ASS passthrough extraction** -- extract S_TEXT/ASS tracks by reconstructing from CodecPrivate + block data. This is the simplest extraction path (no conversion needed) and validates the block reading pipeline.

4. **SRT extraction with ASS conversion** -- parse S_TEXT/UTF8 blocks, convert to ASS format. This completes the core value proposition.

5. **Interactive mode** -- file browser + track selection via stdin. Low complexity, high usability impact.

6. **CLI parameter mode** -- `--track`, `--output` flags. Enables scripting.

Defer to v2:
- **Encoding detection** (High complexity, edge case): Only matters for non-UTF8 tracks, which are rare in modern MKV files.
- **Subtitle preview** (Medium complexity): Nice-to-have but not blocking any workflow.
- **Batch processing**: Shell scripting covers this adequately for v1.
- **Font attachment extraction**: Orthogonal to subtitle extraction.

## Format-Specific Notes

### SRT in MKV (S_TEXT/UTF8)
- Codec ID: `S_TEXT/UTF8`
- Each block contains one subtitle cue as plain UTF-8 text
- Timestamps and duration come from the Block/SimpleBlock header and BlockDuration element
- No header data in CodecPrivate (usually empty or absent)
- May contain HTML-like tags: `<i>`, `<b>`, `<u>`, `<font>`
- Confidence: HIGH (well-documented in Matroska spec)

### SSA/ASS in MKV (S_TEXT/SSA, S_TEXT/ASS)
- Codec IDs: `S_TEXT/SSA` (v4) and `S_TEXT/ASS` (v4+)
- CodecPrivate contains the ASS header: [Script Info], [V4+ Styles], optionally [Fonts], [Graphics]
- Each block contains dialogue fields WITHOUT the "Dialogue:" prefix: `ReadOrder,Layer,Style,Name,MarginL,MarginR,MarginV,Effect,Text`
- ReadOrder field (first field) is used to maintain original ordering since blocks may arrive out of order
- Timestamps come from Block header, not from the text payload
- Confidence: HIGH (well-documented in Matroska spec)

### Formats NOT in Scope
- `S_HDMV/PGS` -- Blu-ray bitmap subtitles
- `S_VOBSUB` -- DVD bitmap subtitles
- `S_DVBSUB` -- DVB bitmap subtitles
- `S_TEXT/WEBVTT` -- WebVTT (rare in MKV, could be added in v2 if demand exists)
- `S_TEXT/USF` -- Universal Subtitle Format (extremely rare)

## Competitive Analysis Summary

| Capability | mkvextract | ffmpeg | This Tool (Target) |
|------------|-----------|--------|-------------------|
| Install complexity | MKVToolNix suite required | ffmpeg binary required | Single Go binary, zero deps |
| List tracks | Via `mkvmerge -i` (separate tool) | Via `ffprobe` (separate tool) | Built-in, human-readable |
| Interactive mode | No | No | Yes -- file + track selection |
| SRT extraction | Yes (raw) | Yes (raw or convert) | Yes, with ASS conversion |
| ASS extraction | Yes (raw) | Yes (raw or convert) | Yes, passthrough |
| Output naming | Manual | Manual | Auto-generated with language |
| Image subtitles | Yes | Yes | No (explicit non-goal) |
| Format conversion | No (separate step) | Yes (inline) | SRT-to-ASS built in |
| Learning curve | Moderate (CLI flags) | High (complex syntax) | Low (interactive + simple CLI) |

## Sources

- Matroska specification (matroska.org/technical/subtitles.html) -- subtitle codec ID definitions and storage format. HIGH confidence from training data.
- MKVToolNix documentation (mkvtoolnix.download) -- mkvextract and mkvmerge capabilities. HIGH confidence from training data.
- FFmpeg documentation (ffmpeg.org) -- subtitle stream handling. HIGH confidence from training data.
- ASS/SSA format specification -- dialogue line format, override tags, script info sections. HIGH confidence from training data.
- Note: Web verification was unavailable during this research session. All findings are based on training data with strong familiarity with these tools and formats. Matroska spec details (codec IDs, block structure, CodecPrivate usage) are stable and well-established -- unlikely to have changed.
