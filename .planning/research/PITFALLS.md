# Domain Pitfalls

**Domain:** Pure Go MKV subtitle extraction (EBML/Matroska parsing, SRT/ASS conversion)
**Researched:** 2026-02-16
**Overall confidence:** MEDIUM (based on training data knowledge of EBML RFC 8794, Matroska spec, and Go binary parsing patterns; web verification was unavailable)

---

## Critical Pitfalls

Mistakes that cause incorrect output, data corruption, or panics on real-world files.

---

### Pitfall 1: EBML VINT (Variable-Size Integer) Decoding Errors

**What goes wrong:** EBML uses a variable-length integer encoding (VINT) where the number of leading zero bits in the first byte determines the total byte width (1-8 bytes). The leading 1 bit is a marker, NOT part of the value. Implementations commonly make one or more of these mistakes:
- Failing to mask out the VINT_MARKER bit from the first byte before assembling the value
- Treating `0xFF` (all-ones-value) as a regular integer instead of the reserved "unknown size" sentinel
- Miscounting leading zeros (off-by-one: the marker bit IS a zero-count boundary, not part of the count)
- Not handling the full 1-8 byte range (only implementing up to 4 bytes)

**Why it happens:** The encoding is deceptively simple-looking but has subtle rules. Developers familiar with UTF-8 variable-length encoding assume EBML VINT works the same way -- it does not. In EBML, a VINT_WIDTH of N means (N-1) leading zero bits followed by a 1-bit marker, then (7*N - N) = (6*N + N - N) ... more precisely, N bytes total with (7*N) data bits minus the width descriptor bits.

**Consequences:** Every single element in the file is framed by VINTs (Element ID + Data Size). A single VINT decoding bug means the parser reads garbage for the element length, seeks to a wrong offset, and either panics on out-of-bounds access or silently skips/corrupts all subsequent data.

**Prevention:**
- Implement VINT decoding as a single, heavily-tested function with explicit test vectors from RFC 8794
- Test vectors MUST include: 1-byte values, 8-byte values, the "unknown size" sentinel for each width, zero values, maximum values at each width boundary
- Specifically test: `0x81` = value 1 (1-byte VINT), `0x40 0x01` = value 1 (2-byte VINT), `0xFF` = unknown (1-byte), `0x01 0xFF 0xFF 0xFF 0xFF 0xFF 0xFF 0xFF` = unknown (8-byte)
- Use `uint64` for all VINT values in Go -- never `int` or `int32`

**Detection:** Parser fails on seemingly valid MKV files; element sizes are implausibly large (billions of bytes); seeking past element boundaries lands in the middle of another element.

**Phase relevance:** Phase 1 (EBML parser foundation). Must be rock-solid before anything else.

---

### Pitfall 2: "Unknown Size" Elements and Streaming/Damaged Files

**What goes wrong:** In EBML, an element can have a data size encoded as "all VINT_DATA bits set to 1" which means "unknown size." This is common for the Segment element (and sometimes Cluster elements) in MKV files, especially those from streaming sources, live recordings, or files written by muxers that don't do a second pass. The parser must handle these by scanning forward for the next sibling element at the same or higher level, rather than trying to seek past a known number of bytes.

**Why it happens:** Developers test with well-formed files from Handbrake or MKVToolNix which always write known sizes. They never encounter unknown-size elements until a user provides a file from OBS, FFmpeg pipe output, or a partially-downloaded torrent.

**Consequences:** The parser either panics (trying to allocate a buffer sized `0x00FFFFFFFFFFFFFF` bytes), hangs (seeking past end of file), or skips the entire Segment element (which contains ALL the actual data).

**Prevention:**
- When reading element size, check if the decoded VINT value equals the "unknown size" sentinel for its width
- For unknown-size Master elements (Segment, Cluster): parse children until hitting an element ID that belongs to a parent/sibling level, or EOF
- This requires knowing the Matroska schema (which element IDs can appear at which level) -- build or embed a schema table
- Specifically test with files where Segment has unknown size (very common in practice)

**Detection:** Parser chokes on files from FFmpeg (`ffmpeg -i input -c copy output.mkv` often produces unknown-size Segments); memory usage spikes indicate buffer over-allocation from misinterpreted unknown sizes.

**Phase relevance:** Phase 1 (EBML parser). Must be addressed alongside VINT decoding.

---

### Pitfall 3: Confusing Element ID Encoding with Data Size Encoding

**What goes wrong:** Both Element IDs and Data Sizes use VINT encoding, but with a critical difference: for Element IDs, the VINT_MARKER bit is PART of the ID value (it is NOT masked out). For Data Sizes, the marker bit IS masked out. Implementations that use the same VINT decoding path for both, masking the marker in both cases, will decode Element IDs incorrectly.

**Why it happens:** RFC 8794 is clear about this, but many tutorials and blog posts gloss over the distinction. The instinct is to write one `readVINT()` function and use it everywhere.

**Consequences:** Element IDs are misidentified. The parser matches the wrong elements, producing garbage track metadata or silently skipping subtitle tracks entirely.

**Prevention:**
- Implement two separate functions: `readVINTValue()` (masks marker, for data sizes) and `readVINTRaw()` (preserves marker, for element IDs)
- Or: one function that returns both the raw bytes and the decoded value, with callers choosing which to use
- Test against known Element IDs: EBML header = `0x1A45DFA3`, Segment = `0x18538067`, Tracks = `0x1654AE6B`, Cluster = `0x1F43B675`

**Detection:** No tracks found in a file that clearly has tracks; wrong CodecID values; parser navigates into wrong elements.

**Phase relevance:** Phase 1. Must be correct from day one.

---

### Pitfall 4: Incorrect Block/SimpleBlock Timestamp and Duration Handling

**What goes wrong:** Subtitle cue timing in Matroska comes from combining the Cluster Timestamp (absolute, in TimestampScale units) with the Block's relative timestamp (a signed 16-bit integer offset from the Cluster Timestamp). Common mistakes:
- Treating the Block timestamp as unsigned (it is signed `int16`, can be negative)
- Forgetting to apply TimestampScale (defaults to 1,000,000 nanoseconds = 1ms, but can be different)
- Missing that subtitle Blocks often have a BlockDuration element that specifies how long the cue displays -- without it, the cue has no end time
- Using SimpleBlock for subtitles when Block+BlockDuration is required (SimpleBlock has no duration)

**Why it happens:** Audio/video blocks don't need explicit durations (they're derived from codec frame sizes or next block timestamps). Subtitle blocks absolutely require explicit durations because subtitle cues are sparse -- there can be long gaps between cues. Developers who learn the Block structure from audio/video examples miss this.

**Consequences:** Subtitles display with duration=0 (flash on screen and vanish), or display forever until the next cue, or have timestamps shifted by minutes due to signed/unsigned confusion.

**Prevention:**
- Parse the 16-bit block timestamp as `int16` explicitly: `int16(buf[0])<<8 | int16(buf[1])`
- Always read BlockDuration for subtitle tracks; if missing, log a warning (some poorly-muxed files omit it)
- Convert all timestamps through TimestampScale: `absoluteTimestampNs = (clusterTimestamp + blockRelativeTimestamp) * timestampScale`
- Then convert to the output format's time units (ASS uses centiseconds: `h:mm:ss.cc`)

**Detection:** Subtitles appear at wrong times; subtitle cues display for 0 seconds or forever; timestamps go negative or wrap around.

**Phase relevance:** Phase 2 (subtitle extraction). Core to correct subtitle output.

---

### Pitfall 5: SRT-to-ASS Timestamp Precision Loss

**What goes wrong:** SRT uses millisecond precision (`hh:mm:ss,mmm`), ASS uses centisecond precision (`h:mm:ss.cc`). Naive conversion truncates instead of rounding, or worse, uses the wrong separator (SRT uses comma, ASS uses dot). Additionally, the Matroska internal timestamps are in nanoseconds (scaled), which must be converted to centiseconds for ASS without accumulating floating-point drift.

**Why it happens:** Developers use floating-point division (`ms / 10.0`) and format with `%02d`, which truncates. Or they parse the SRT timestamp string and just reformat it without accounting for the precision difference.

**Consequences:** Subtitles are off by up to 9ms per cue. Over a 2-hour movie with hundreds of cues, the cumulative visual effect is noticeable -- subtitles feel "laggy" or "early" in sections.

**Prevention:**
- Use integer arithmetic throughout: `centiseconds = (milliseconds + 5) / 10` for proper rounding
- For Matroska timestamps: `centiseconds = (nanoseconds + 5_000_000) / 10_000_000`
- Never use `float64` for timestamp arithmetic in the conversion pipeline
- Verify with a test: `12:34:56,789` (SRT) should become `1:23:45.68` ... wait, let me be precise: SRT `01:23:45,789` = 5025789ms = 502578.9cs, rounds to `1:23:45.79` in ASS
- Use Go's `time.Duration` for intermediate representation if helpful, but be aware it's nanosecond-based

**Detection:** A/B comparison of extracted subtitles against mkvextract/ffmpeg output shows timing differences; subtitles feel slightly mistimed when played back.

**Phase relevance:** Phase 3 (format conversion). Affects output quality directly.

---

### Pitfall 6: Matroska Lacing Mishandling

**What goes wrong:** Matroska supports "lacing" -- packing multiple frames into a single Block. There are four lacing modes: no lacing, Xiph lacing, fixed-size lacing, and EBML lacing. Each has different rules for encoding frame sizes within the block. Subtitle tracks almost never use lacing (and shouldn't), but the parser must still handle the lacing flags in the Block header correctly, or it will misinterpret the subtitle data.

**Why it happens:** The lacing flags are bits 1-2 of the Block header's flag byte (after the track number VINT and the 16-bit timestamp). Even if subtitles don't use lacing, the parser must read these flags to correctly locate the start of the actual frame data. Ignoring lacing flags means the parser might try to interpret flag bytes as subtitle text.

**Consequences:** Subtitle text starts with garbage bytes (the misinterpreted lacing header); for files that DO use lacing on subtitle tracks (rare but legal), the parser produces concatenated or corrupted cues.

**Prevention:**
- Always parse the Block header completely: track number (VINT), timestamp (int16), flags (1 byte, with lacing bits)
- If lacing is indicated (flags bits 1-2 are non-zero), parse the lacing header to find individual frame boundaries
- For subtitle extraction, it's acceptable to support only "no lacing" and error on laced subtitle blocks with a clear message, since lacing on subtitles is extremely rare
- But DO NOT skip parsing the flags byte -- it's always present

**Detection:** Subtitle text starts with non-printable characters or byte sequences that look like binary data; first character of each subtitle is garbled.

**Phase relevance:** Phase 2 (Block parsing). Must handle even if lacing is not expected.

---

## Moderate Pitfalls

---

### Pitfall 7: Ignoring CodecPrivate for SSA/ASS Passthrough

**What goes wrong:** For SSA/ASS subtitle tracks (`S_TEXT/SSA`, `S_TEXT/ASS`), the Matroska container stores the ASS header (including `[Script Info]`, `[V4+ Styles]`, and other metadata sections) in the track's `CodecPrivate` element. The actual Block data contains only the dialogue lines, formatted as ASS "Dialogue" events WITHOUT the `Dialogue:` prefix and layer/timing fields (those come from the Block's timestamp and BlockDuration). Implementations that ignore CodecPrivate will produce ASS files without style definitions.

**Why it happens:** For SRT tracks (`S_TEXT/UTF8`), CodecPrivate is typically empty or absent, so developers don't look for it. When they move to ASS support, they assume subtitle data is self-contained in each Block.

**Consequences:** Extracted ASS files have no style information, so all text renders in the player's default style (usually white Arial, no positioning). The entire value of ASS over SRT -- styled text, karaoke effects, screen positioning -- is lost.

**Prevention:**
- Always read `CodecPrivate` from `TrackEntry` for subtitle tracks
- For `S_TEXT/ASS` and `S_TEXT/SSA`: CodecPrivate contains the file header; each Block contains a single event line in the format `ReadOrder,Layer,Style,Name,MarginL,MarginR,MarginV,Effect,Text`
- Reconstruct the full ASS file by combining CodecPrivate header + `[Events]` section header + formatted Dialogue lines from Blocks
- For `S_TEXT/UTF8` (SRT): CodecPrivate is usually absent; each Block is a plain-text subtitle cue (without SRT sequence numbers or timestamps)

**Detection:** Extracted ASS files are valid but all text is unstyled; diff against mkvextract output shows missing header sections.

**Phase relevance:** Phase 2-3 (extraction + conversion). Critical for ASS passthrough quality.

---

### Pitfall 8: Matroska Element Ordering Assumptions

**What goes wrong:** The parser assumes elements appear in a specific order (e.g., Tracks before Clusters, or SeekHead at the start of the Segment). While most well-formed MKV files follow this convention, the Matroska spec does NOT mandate element ordering within a Master element. Some muxers place Tracks AFTER the first Cluster, or omit SeekHead entirely.

**Why it happens:** Testing with files from a single muxer (usually MKVToolNix, which produces nicely-ordered files). Other muxers (FFmpeg, older versions of VLC) may order elements differently.

**Consequences:** Parser can't find track information because it stops scanning after hitting the first Cluster; or it crashes because it tries to interpret Cluster data using Tracks parsing logic.

**Prevention:**
- Use SeekHead (if present) to find element positions, but fall back to linear scanning if SeekHead is absent
- Before processing Clusters, ensure Tracks have been parsed; if not, scan forward for Tracks first
- Design the parser to handle a two-pass approach: first pass locates and parses metadata (SegmentInfo, Tracks), second pass processes Clusters
- Or: buffer cluster data and process after metadata is fully known

**Detection:** Parser reports "no tracks found" on files that mkvinfo/ffprobe clearly shows have tracks; parser works for MKVToolNix files but fails on FFmpeg-muxed files.

**Phase relevance:** Phase 1-2. Architecture decision that affects parser design.

---

### Pitfall 9: Go io.Reader vs io.ReadSeeker Design Choice

**What goes wrong:** The parser is built around `io.Reader` (streaming, forward-only), but MKV parsing often requires seeking -- following SeekHead pointers, skipping large video/audio elements, or doing the two-pass approach mentioned above. Retrofitting seeking into a streaming parser is a painful rewrite.

Alternatively: building entirely around `io.ReadSeeker` (requires the whole file on disk) prevents future use with streaming inputs (HTTP range requests, pipes).

**Why it happens:** Go developers default to `io.Reader` because it's the most general interface. But EBML is not a streaming-friendly format for metadata -- you often need random access to find Tracks before you can interpret Clusters.

**Consequences:** Either the parser loads the entire file into memory (unacceptable for multi-GB MKV files), or it can't handle SeekHead and breaks on files where Tracks come after Clusters.

**Prevention:**
- Use `io.ReadSeeker` as the primary interface (covers `os.File`, `bytes.Reader`, etc.)
- For skipping large elements (video/audio data): use `io.Seeker.Seek()` rather than reading and discarding bytes
- For future streaming support: document this as a known limitation and design the element-reading layer with a potential `io.Reader` fallback that buffers only metadata elements
- Use `io.SectionReader` for parsing nested elements within known bounds

**Detection:** Excessive memory usage on large files (reading video data into memory just to skip it); inability to process files where metadata is at the end.

**Phase relevance:** Phase 1 (core parser architecture). Fundamental design decision.

---

### Pitfall 10: Character Encoding Assumptions

**What goes wrong:** Matroska specifies UTF-8 for all string elements. SRT files in the wild, however, frequently use Windows-1252, Shift-JIS, GB2312, or other encodings -- despite the spec saying UTF-8. When a user has an MKV where the muxer embedded a non-UTF-8 SRT track (preserving the original encoding), the extracted text will contain mojibake if blindly treated as UTF-8.

**Why it happens:** Matroska says "UTF-8" so developers trust it. In practice, many MKV files are muxed from subtitle files with incorrect encoding declarations.

**Consequences:** Chinese/Japanese/Korean subtitles display as garbage characters; accented European characters are corrupted.

**Prevention:**
- After extracting subtitle text, validate it as UTF-8 using `utf8.Valid()`
- If invalid UTF-8 is detected, attempt common encoding detection (BOM check, statistical analysis, or use the track's Language element as a hint)
- For v1, it's acceptable to: (a) warn the user about non-UTF-8 data, (b) output the raw bytes as-is, (c) optionally offer `--encoding` flag for manual override
- Do NOT silently replace invalid bytes with `U+FFFD` without informing the user

**Detection:** Output file contains `?` or `U+FFFD` replacement characters where readable text is expected; user reports garbled subtitles for non-English content.

**Phase relevance:** Phase 3 (output quality). Can be addressed incrementally -- warn first, auto-detect later.

---

### Pitfall 11: ASS Format Spec Quirks

**What goes wrong:** The ASS format has several non-obvious requirements that trip up generators:
- The `[Script Info]` section MUST come first, before `[V4+ Styles]` and `[Events]`
- Style names are case-sensitive in some players, case-insensitive in others
- The `PlayResX`/`PlayResY` values in `[Script Info]` define the coordinate system for positioning -- if omitted, players use wildly different defaults (384x288 in some, video resolution in others)
- Line breaks within a dialogue line use `\N` (literal backslash-N), not an actual newline character
- The `Format:` line in `[Events]` must exactly match the fields used in `Dialogue:` lines, and field count must match
- SSA (v4) and ASS (v4+) have different style field counts and names

**Why it happens:** There is no single authoritative ASS spec. The format originated from Sub Station Alpha, was extended to Advanced Sub Station Alpha, and different players (VSFilter, libass, xy-VSFilter) interpret edge cases differently.

**Consequences:** Generated ASS files render correctly in some players but not others; styles don't apply; positioning is wrong; dialogue lines are parsed incorrectly by the player.

**Prevention:**
- When generating ASS from SRT: use a minimal but complete template with explicit `PlayResX: 1920`, `PlayResY: 1080`, a default style, and the full `Format:` line
- When passing through ASS: preserve the CodecPrivate header verbatim, do not reformat or "clean up" the style definitions
- Always use `\N` for hard line breaks, `\n` for soft line breaks (word-wrap points)
- Test output with at least two renderers: mpv (libass) and a VSFilter-based player

**Detection:** Subtitles look correct in mpv but wrong in PotPlayer, or vice versa; positioning is off; styles are ignored.

**Phase relevance:** Phase 3 (ASS output generation). Affects correctness of the final deliverable.

---

## Minor Pitfalls

---

### Pitfall 12: Large Element Size Allocation (DoS Vector)

**What goes wrong:** A malformed or malicious MKV file declares a Data Size of 2^56 bytes for a string element. The parser allocates a `[]byte` slice of that size and the process is killed by the OS.

**Prevention:**
- Set reasonable maximum sizes for non-Master elements: string elements (track name, codec ID) should be capped at ~1MB, subtitle block data at ~10MB
- For Master elements, don't allocate buffers -- parse children in-place
- Validate element size against remaining file size before allocation

**Phase relevance:** Phase 1. Simple guard clauses during element reading.

---

### Pitfall 13: Multiple Subtitle Tracks with Same Language

**What goes wrong:** MKV files frequently contain multiple subtitle tracks with the same language tag (e.g., two English tracks -- one for full dialogue, one for signs/songs only; or one normal, one SDH). The user-facing track selection must show enough information to distinguish them.

**Prevention:**
- Display: track number, language, codec, track name (from `Name` element in `TrackEntry`), and any flags (default, forced)
- The `Name` element is the primary differentiator and must be extracted and displayed
- Also show `FlagDefault` and `FlagForced` as these indicate purpose

**Phase relevance:** Phase 2 (track enumeration/selection UI).

---

### Pitfall 14: Void and CRC-32 Elements

**What goes wrong:** MKV files contain `Void` elements (padding, used by muxers to reserve space for future edits) and `CRC-32` elements (checksums for the preceding element). Parsers that don't handle these will misidentify them as unknown elements and potentially error out or corrupt the parse state.

**Prevention:**
- Recognize `Void` (ID `0xEC`) and `CRC-32` (ID `0xBF`) elements and skip them (read their size, seek past their data)
- They can appear as children of almost any Master element
- Do NOT treat unknown element IDs as errors -- skip them by their declared size

**Phase relevance:** Phase 1 (EBML parser). Simple to handle but must not be forgotten.

---

### Pitfall 15: SRT Sequence Number and Blank Line Parsing

**What goes wrong:** When converting extracted SRT data to ASS, the parser must handle SRT's quirks: sequence numbers can be non-sequential or missing; cues can have blank lines within the text (which normally delimit cues); timestamps can use either `,` or `.` as the millisecond separator depending on locale.

**Why it happens:** SRT has no formal specification. Every SRT parser must handle the "informal spec" which tolerates many variations.

**Prevention:**
- In Matroska, SRT tracks (`S_TEXT/UTF8`) store each cue as a separate Block, so the parser does NOT need to split by blank lines or parse SRT sequence numbers -- each Block is already one cue, and timing comes from the Block/Cluster timestamps
- This is a major simplification: do NOT reimplement a full SRT parser for extraction from MKV
- Only implement full SRT parsing if you add standalone SRT file conversion (outside MKV)

**Detection:** Over-engineered SRT parsing code that reimplements sequence number handling unnecessarily.

**Phase relevance:** Phase 2-3. Understanding this avoids wasted work.

---

## Phase-Specific Warnings

| Phase Topic | Likely Pitfall | Mitigation |
|-------------|---------------|------------|
| EBML Parser Core | VINT encoding bugs (Pitfalls 1, 3) | Extensive test vectors from RFC 8794; separate ID vs. size decoding paths |
| EBML Parser Core | Unknown-size elements (Pitfall 2) | Schema-aware parsing; test with FFmpeg-muxed streaming files |
| EBML Parser Core | io.Reader vs io.ReadSeeker (Pitfall 9) | Commit to `io.ReadSeeker` early; use `Seek()` to skip large elements |
| EBML Parser Core | Void/CRC-32 handling (Pitfall 14) | Skip-on-sight; don't treat unknown IDs as fatal |
| Track/Block Parsing | Block timestamp signed int16 (Pitfall 4) | Explicit `int16` cast; test with negative relative timestamps |
| Track/Block Parsing | Lacing flags (Pitfall 6) | Always parse the flags byte; error on unexpected lacing for subtitles |
| Track/Block Parsing | CodecPrivate for ASS (Pitfall 7) | Read CodecPrivate before processing Blocks; reconstruct ASS header |
| Track/Block Parsing | Element ordering (Pitfall 8) | SeekHead + fallback linear scan; don't assume Tracks-before-Clusters |
| Track/Block Parsing | Multiple same-language tracks (Pitfall 13) | Show track name, flags, and codec in selection UI |
| Format Conversion | Timestamp precision loss (Pitfall 5) | Integer arithmetic with rounding; never use float64 for timestamps |
| Format Conversion | ASS format quirks (Pitfall 11) | Minimal complete template; test with multiple players |
| Format Conversion | Encoding issues (Pitfall 10) | UTF-8 validation; warn user; optional encoding flag |
| Format Conversion | SRT over-parsing (Pitfall 15) | Use Block-level cues from MKV, not SRT text parsing |
| General Robustness | Memory exhaustion (Pitfall 12) | Size caps on non-Master elements; validate against file size |

---

## Sources

- RFC 8794: Extensible Binary Meta Language (EBML) specification -- definitive reference for VINT encoding, element structure [MEDIUM confidence: from training data, not live-verified]
- Matroska specification (matroska.org/technical/) -- element hierarchy, codec IDs, Block structure, subtitle storage format [MEDIUM confidence: from training data]
- Sub Station Alpha / Advanced Sub Station Alpha format documentation -- ASS format quirks, section ordering, field definitions [MEDIUM confidence: community documentation, no single authoritative source]
- Go standard library (`encoding/binary`, `io`, `utf8` packages) -- binary parsing patterns [HIGH confidence: well-known stable APIs]
- Community experience with MKV parsing implementations (mkvtoolnix issues, ffmpeg source, various Go EBML libraries) -- real-world edge cases [LOW confidence: aggregated from training data without specific source verification]

**Note:** Web search and documentation fetch were unavailable during this research session. All findings are based on training data knowledge. Confidence levels are adjusted accordingly. Critical pitfalls (1-6) should be verified against the current EBML RFC 8794 and Matroska specification before implementation begins.
