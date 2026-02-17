# Phase 2: Subtitle Extraction and ASS Output - Research

**Researched:** 2026-02-16
**Domain:** Matroska subtitle block extraction, ASS/SSA format output, SRT-to-ASS conversion
**Confidence:** HIGH

## Summary

Phase 2 builds on Phase 1's track discovery to actually extract subtitle data from MKV files and produce valid ASS files. There are two distinct paths: (1) ASS/SSA passthrough, where the CodecPrivate header is written verbatim and dialogue lines are reconstructed from block data, and (2) SRT-to-ASS conversion, where a default ASS header is generated and plain-text cues are wrapped in ASS Dialogue lines with HTML tag mapping.

The `matroska-go` library's `ReadPacket()` API provides `Packet.StartTime` and `Packet.EndTime` in nanoseconds, `Packet.Data` as raw bytes, and `Packet.Track` for track filtering. This gives us everything needed: loop ReadPacket, filter by track number, collect subtitle data. For ASS/SSA tracks, `TrackInfo.CodecPrivate` contains the full ASS header ([Script Info] + [V4+ Styles] sections). Each block's `Data` contains the event fields in Matroska order: `ReadOrder,Layer,Style,Name,MarginL,MarginR,MarginV,Effect,Text`. For SRT tracks, each block's `Data` is plain UTF-8 text (one cue), and CodecPrivate is empty.

The decision to NOT use `go-astisub` for this phase is a key recommendation. The user's locked decision requires CodecPrivate to be output **verbatim** as the ASS header. go-astisub's `ReadFromSSA` -> `WriteToSSA` pipeline parses and regenerates the ASS file, which can alter formatting, drop unknown sections, reorder fields, or lose custom metadata. For passthrough mode, direct string manipulation is both simpler and more faithful. For SRT-to-ASS conversion, the ASS header generation is simple enough (a template with ~15 lines of config) that a library dependency adds complexity without value. Hand-rolling the ASS output for this specific use case (one default style, simple tag mapping) is the correct approach. go-astisub remains valuable as a future dependency if the project expands to support more subtitle formats.

**Primary recommendation:** Hand-roll ASS output using direct string formatting. Use `matroska-go` ReadPacket for extraction. Do NOT add `go-astisub` as a dependency for Phase 2.

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **ASS passthrough strategy**: CodecPrivate data output verbatim as ASS file header -- preserve all styles, fonts, resolution settings without modification. SSA (old V4) tracks are converted to ASS V4+ format for modern player compatibility. Dialogue lines sorted by start timestamp in output (not MKV storage order). [Events] Format line: extract from CodecPrivate if present; generate standard default if absent.
- **SRT to ASS conversion style**: Default font: Microsoft YaHei. Default style: white text + black outline, font size ~20-24, bottom-center aligned. HTML tag mapping: `<b>` -> `{\b1}`, `<i>` -> `{\i1}`, `<u>` -> `{\u1}`, `<font color>` -> `{\c&H...&}`. Multi-line subtitles: SRT newlines converted to `\N` (hard line break) in ASS.
- **Timestamp precision and alignment**: Millisecond to centisecond conversion: round, not truncate. Negative or zero timestamps: clamp to 0:00:00.00. Missing BlockDuration: end time set to next subtitle's start time (gap-fill strategy).
- **Output file naming and encoding**: Naming format: `{video_basename}.{lang_code}.ass` using ISO 639-2 three-letter codes. Name collision: use track name to differentiate if available; fall back to sequence number. File encoding: UTF-8 without BOM. Line endings: Claude's Discretion.

### Claude's Discretion
- Line ending style (CRLF vs LF) -- choose based on ASS ecosystem conventions
- Exact font size within the 20-24 range for SRT -> ASS default style
- ASS PlayResX/PlayResY values for generated SRT -> ASS headers
- SSA -> ASS V4+ conversion specifics (field mapping, style format differences)
- How to handle edge cases in SRT HTML tag parsing (nested tags, malformed HTML)

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/luispater/matroska-go` | v1.2.4 | Subtitle packet reading via `ReadPacket()` | Already in go.mod from Phase 1. Provides `Packet{Track, StartTime, EndTime, Data}` in nanoseconds. Handles Cluster/Block navigation, timestamp calculation, and BlockDuration internally. |
| Go stdlib (`strings`, `fmt`, `sort`, `bytes`, `os`, `path/filepath`, `math`, `regexp`) | N/A | ASS file generation, timestamp formatting, SRT tag parsing, file I/O | ASS output is text-based. No library needed for generating a well-defined text format with known structure. |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `golang.org/x/text` | v0.34.0 | Already in go.mod; language name resolution from Phase 1 | Used for output file naming (language codes are already resolved in Phase 1 types) |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| Hand-rolled ASS output | `asticode/go-astisub` v0.38.0 | go-astisub parses+regenerates ASS, which **breaks the passthrough requirement**. It may alter CodecPrivate formatting, drop unknown sections, reorder fields. For SRT conversion, the overhead of a full subtitle library for generating a single default style + simple tag mapping is not justified. |
| Hand-rolled SRT tag parser | `golang.org/x/net/html` | HTML tokenizer is overkill for the 4 supported tags. A simple regex or manual parser for `<b>`, `<i>`, `<u>`, `<font color>` is sufficient and avoids a new dependency. |

## Architecture Patterns

### Recommended Project Structure (Phase 2 additions)

```
mkv-sub-extractor/
  pkg/
    mkvinfo/           # Phase 1 (existing) -- track listing
    extract/           # NEW: subtitle block extraction from MKV
      extract.go       # ReadPacket loop, track filtering, raw block collection
      extract_test.go
    subtitle/          # NEW: subtitle data model and codec handlers
      types.go         # SubtitleEvent struct (start, end, text, layer, style, etc.)
      ass.go           # ASS/SSA block data parser (ReadOrder,Layer,Style,...,Text)
      ass_test.go
      srt.go           # SRT block data handler (plain text + HTML tag mapping)
      srt_test.go
    assout/            # NEW: ASS file output writer
      writer.go        # Writes complete ASS file (header + events)
      writer_test.go
      template.go      # Default ASS header template for SRT conversion
      timestamp.go     # ASS timestamp formatting (H:MM:SS.CC)
      timestamp_test.go
      ssa_convert.go   # SSA V4 -> ASS V4+ header conversion
      ssa_convert_test.go
    output/            # NEW: output file path generation
      naming.go        # {basename}.{lang}.ass naming, collision handling
      naming_test.go
```

### Pattern 1: Packet Collection with Track Filtering

**What:** Read all packets from the demuxer, filter by target track number, collect subtitle packets.
**When to use:** Always, for extraction.
**Example:**
```go
// Source: matroska-go pkg.go.dev API + example/extracter
func ExtractSubtitlePackets(demuxer *matroska.Demuxer, trackNumber uint8) ([]*matroska.Packet, error) {
    var packets []*matroska.Packet
    for {
        pkt, err := demuxer.ReadPacket()
        if err == io.EOF {
            break
        }
        if err != nil {
            return nil, fmt.Errorf("read packet error: %w", err)
        }
        if pkt.Track == trackNumber {
            packets = append(packets, pkt)
        }
    }
    return packets, nil
}
```

### Pattern 2: ASS Passthrough Reconstruction

**What:** Combine CodecPrivate header + generated [Events] section from block data.
**When to use:** For S_TEXT/ASS and S_TEXT/SSA tracks.
**Why:** Matroska stores the ASS header in CodecPrivate and dialogue fields in block data (without "Dialogue:" prefix, without timestamps). We reconstruct the full ASS file.
**Example:**
```go
// Source: Matroska subtitle spec (matroska.org/technical/subtitles.html)
func WriteASSPassthrough(w io.Writer, codecPrivate []byte, events []SubtitleEvent) error {
    // 1. Write CodecPrivate verbatim as header
    w.Write(codecPrivate)

    // 2. Check if CodecPrivate already contains [Events] section
    //    If it does, extract Format line from it
    //    If not, generate default Format line
    header := string(codecPrivate)
    if !containsEventsSection(header) {
        fmt.Fprintf(w, "\n[Events]\n")
        fmt.Fprintf(w, "Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")
    }

    // 3. Sort events by start timestamp (user decision: sorted output)
    sort.Slice(events, func(i, j int) bool {
        return events[i].Start < events[j].Start
    })

    // 4. Write Dialogue lines with timestamps from Matroska Block timing
    for _, e := range events {
        fmt.Fprintf(w, "Dialogue: %s,%s,%s,%s\n",
            e.Layer,
            formatASSTimestamp(e.Start),
            formatASSTimestamp(e.End),
            e.RemainingFields, // Style,Name,MarginL,MarginR,MarginV,Effect,Text
        )
    }
    return nil
}
```

### Pattern 3: Matroska ASS/SSA Block Data Parsing

**What:** Parse the comma-separated fields in each block's Data for ASS/SSA tracks.
**When to use:** For every subtitle block from an ASS/SSA track.
**Critical detail:** Block data format is `ReadOrder,Layer,Style,Name,MarginL,MarginR,MarginV,Effect,Text`. The Text field may contain commas, so split into at most 9 fields (first 8 by comma, 9th is everything remaining).
**Example:**
```go
// Source: Matroska spec - matroska.org/technical/subtitles.html
// Block data: "0,0,Default,,0,0,0,,Hello, world!"
// Fields:     ReadOrder=0, Layer=0, Style=Default, Name="", MarginL=0,
//             MarginR=0, MarginV=0, Effect="", Text="Hello, world!"
func ParseASSBlockData(data []byte) (readOrder int, fields string, err error) {
    s := string(data)
    // Split into at most 9 parts (ReadOrder + 8 remaining)
    parts := strings.SplitN(s, ",", 9)
    if len(parts) < 9 {
        return 0, "", fmt.Errorf("expected 9 fields, got %d", len(parts))
    }
    readOrder, err = strconv.Atoi(strings.TrimSpace(parts[0]))
    if err != nil {
        return 0, "", fmt.Errorf("invalid ReadOrder: %w", err)
    }
    // Rejoin fields 1-8 (Layer through Text) for Dialogue line
    remaining := strings.Join(parts[1:], ",")
    return readOrder, remaining, nil
}
```

### Pattern 4: SRT-to-ASS Conversion Pipeline

**What:** Generate default ASS header, then wrap each SRT cue in an ASS Dialogue line with tag mapping.
**When to use:** For S_TEXT/UTF8 (SRT) tracks.
**Example:**
```go
func WriteSRTAsASS(w io.Writer, events []SubtitleEvent) error {
    // 1. Write generated [Script Info]
    writeDefaultScriptInfo(w)

    // 2. Write [V4+ Styles] with default style
    writeDefaultStyles(w)

    // 3. Write [Events]
    fmt.Fprintf(w, "\n[Events]\n")
    fmt.Fprintf(w, "Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text\n")

    sort.Slice(events, func(i, j int) bool {
        return events[i].Start < events[j].Start
    })

    for _, e := range events {
        text := convertSRTTagsToASS(e.Text)
        text = strings.ReplaceAll(text, "\n", "\\N")
        text = strings.ReplaceAll(text, "\r", "")
        fmt.Fprintf(w, "Dialogue: 0,%s,%s,Default,,0,0,0,,%s\n",
            formatASSTimestamp(e.Start),
            formatASSTimestamp(e.End),
            text,
        )
    }
    return nil
}
```

### Pattern 5: ASS Timestamp Formatting

**What:** Convert nanoseconds to ASS timestamp format `H:MM:SS.CC` with proper rounding.
**When to use:** Every Dialogue line output.
**Critical:** Centisecond rounding, NOT truncation. Single-digit hour. Two-digit centiseconds.
**Example:**
```go
// Source: ASS format guide (github.com/libass/libass/wiki/ASS-File-Format-Guide)
// Format: h:mm:ss.dd where h=0-595, mm/ss=00-60, dd=00-99
func formatASSTimestamp(ns uint64) string {
    // Clamp negative/zero to 0
    // Round nanoseconds to centiseconds: (ns + 5_000_000) / 10_000_000
    cs := (ns + 5_000_000) / 10_000_000

    hours := cs / 360000
    cs %= 360000
    minutes := cs / 6000
    cs %= 6000
    seconds := cs / 100
    centiseconds := cs % 100

    return fmt.Sprintf("%d:%02d:%02d.%02d", hours, minutes, seconds, centiseconds)
}
```

### Anti-Patterns to Avoid

- **Using go-astisub for ASS passthrough:** The library parses and regenerates the file, potentially altering CodecPrivate content. The user decision explicitly requires verbatim output of CodecPrivate. Writing it as raw bytes is both simpler and more correct.
- **Using float64 for timestamp arithmetic:** Floating-point accumulates drift. Use integer arithmetic throughout: nanoseconds -> centiseconds via `(ns + 5_000_000) / 10_000_000`.
- **Parsing SRT text format from block data:** In Matroska, SRT blocks are already individual cues (no sequence numbers, no timestamp lines). Do NOT build an SRT parser that handles `-->` arrows or blank-line delimiters -- that is for standalone `.srt` files, not MKV-extracted data.
- **Assuming fixed block order:** Blocks may arrive out of MKV storage order. Always sort by start timestamp before output (user decision: sorted output).
- **Splitting ASS block data naively on all commas:** The Text field (last field) can contain commas. Always use `strings.SplitN(s, ",", 9)` to limit splitting.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| MKV packet reading | Custom Cluster/Block parser | `matroska-go` ReadPacket() | Library handles Cluster navigation, Block parsing, lacing, timestamp calculation, BlockDuration. We already depend on it. |
| EBML/VINT decoding | Custom VINT parser | `matroska-go` internals | Already solved by Phase 1 dependency. |
| Language code resolution | Custom ISO 639-2 map | Phase 1's `ResolveLanguageName()` + `golang.org/x/text` | Already implemented and tested. |

**Key insight:** Phase 2 SHOULD hand-roll ASS output (it's a well-defined text format), but should NOT hand-roll MKV packet extraction (matroska-go handles all the binary complexity). The split is: library for binary parsing, custom code for text generation.

## Common Pitfalls

### Pitfall 1: matroska-go Duration Bug May Affect Packet Timestamps
**What goes wrong:** Phase 1 discovered that `matroska-go` reads Matroska float elements (like Duration) using `ReadUInt()`, returning raw IEEE 754 bits as uint64 instead of the actual float value. The workaround was `math.Float32frombits()`/`math.Float64frombits()`. If the library has a similar issue with `Packet.StartTime` or `Packet.EndTime`, timestamps could be raw float bits instead of nanoseconds.
**Why it happens:** Internal library implementation detail -- it may use `ReadUInt` for all numeric fields.
**How to avoid:** Test early with a real MKV file. Extract a few subtitle packets and verify that `StartTime`/`EndTime` produce sensible nanosecond values (e.g., a subtitle at 1 minute should have StartTime ~60,000,000,000 ns). If values look like IEEE 754 bit patterns (very large integers like 4.6e18), the same Float-from-bits workaround is needed.
**Warning signs:** Timestamps that are astronomical numbers (>10^18) instead of reasonable nanosecond values.
**Confidence:** MEDIUM -- The example/extracter code uses `packet.StartTime` directly in `formatSRTTime(ns uint64)` and divides by 10^9 to get seconds. This suggests StartTime/EndTime ARE in nanoseconds (not raw float bits). But verify early.

### Pitfall 2: Missing BlockDuration (EndTime = 0)
**What goes wrong:** Some poorly-muxed MKV files have subtitle blocks without BlockDuration. The `Packet.EndTime` would be 0 or equal to `StartTime`, making the subtitle display for 0 duration.
**Why it happens:** SimpleBlock does not carry duration; only BlockGroup with BlockDuration does. Subtitle tracks should always use BlockGroup, but some muxers get this wrong.
**How to avoid:** User decision specifies gap-fill strategy: when EndTime is missing or zero, set end time to the next subtitle's start time. For the last subtitle, use start time + a reasonable default (e.g., 5 seconds). Check: if `pkt.EndTime <= pkt.StartTime`, treat as missing duration.
**Warning signs:** Subtitles that flash and disappear instantly, or display for exactly 0 seconds.

### Pitfall 3: ASS/SSA Block Data Field Count
**What goes wrong:** Splitting block data on commas produces wrong field boundaries because the Text field contains commas.
**Why it happens:** Naive `strings.Split(s, ",")` splits the Text field too.
**How to avoid:** Always use `strings.SplitN(s, ",", 9)` -- first 8 fields are delimited, 9th field (Text) is everything remaining. The field order is: ReadOrder, Layer, Style, Name, MarginL, MarginR, MarginV, Effect, Text.
**Warning signs:** Dialogue text is truncated at the first comma; extra empty Dialogue lines appear.

### Pitfall 4: CodecPrivate Trailing Content
**What goes wrong:** Some CodecPrivate data includes a partial `[Events]` section or even a Format line. Writing it verbatim and then appending our own `[Events]` section creates a duplicate.
**Why it happens:** Different muxers store different amounts of the original ASS file in CodecPrivate. Most store only [Script Info] + [V4+ Styles], but some include [Events] header.
**How to avoid:** After writing CodecPrivate verbatim, check if it already contains `[Events]` and/or a `Format:` line. If it does, do not add another. If it contains a Format line, use that format for our Dialogue lines.
**Warning signs:** Duplicate `[Events]` sections; players display an error or skip events.

### Pitfall 5: SSA V4 to ASS V4+ Conversion Complexity
**What goes wrong:** SSA uses `[V4 Styles]` with 18 style fields; ASS uses `[V4+ Styles]` with 23 fields. Simply renaming the section header without adjusting field count/names produces invalid ASS.
**Why it happens:** SSA v4 styles have: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding. ASS v4+ styles have: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR, MarginV, Encoding.
**How to avoid:** Parse the SSA Format line to identify fields, map TertiaryColour -> OutlineColour, add missing fields with defaults (Underline=0, StrikeOut=0, ScaleX=100, ScaleY=100, Spacing=0, Angle=0), remove AlphaLevel, change section header. Also update `ScriptType: v4.00+` in Script Info. Also change event format from `Marked` to `Layer`.
**Warning signs:** Players report "invalid style format" or apply no styles.

### Pitfall 6: SRT HTML Tag Nesting and Malformation
**What goes wrong:** SRT HTML tags can be nested (`<b><i>text</i></b>`), overlapping (`<b><i>text</b></i>`), or malformed (`<b>text` without closing tag). Naive regex replacement can produce broken ASS override tags.
**How to avoid:** Use a simple state-machine or stack-based parser for SRT tags. For the 4 supported tags (`<b>`, `<i>`, `<u>`, `<font color>`), process opening tags to insert `{\b1}`, `{\i1}`, etc. at the position, and closing tags to insert `{\b0}`, `{\i0}`, etc. For unclosed tags: auto-close at end of cue (ASS override tags reset per line anyway). For malformed tags: strip them (do not crash).
**Warning signs:** ASS output has unmatched `{\b1}` without `{\b0}`; font color tags produce malformed `{\c}` tags.

### Pitfall 7: Font Color HTML-to-ASS Conversion
**What goes wrong:** SRT uses `<font color="#RRGGBB">` (RGB order), ASS uses `{\c&HBBGGRR&}` (BGR order, reversed). Failing to reverse the byte order produces wrong colors.
**Why it happens:** ASS color format is BGR, not RGB. The `&H` prefix and `&` suffix are also easy to get wrong.
**How to avoid:** Parse the hex color, extract R/G/B components, reassemble as `&HBBGGRR&`. Handle both `#RRGGBB` and named color formats (common: `red`, `blue`, `green`, `white`, `yellow`). For unknown color names, strip the tag rather than produce invalid ASS.
**Warning signs:** Red text appears blue; blue text appears red; named colors not recognized.

## Code Examples

Verified patterns from official sources:

### Default ASS Header Template for SRT Conversion

```go
// Source: ASS format guide (github.com/libass/libass/wiki/ASS-File-Format-Guide)
// Discretion choices: PlayResX=1920, PlayResY=1080, FontSize=22, CRLF line endings
const defaultASSHeader = "[Script Info]\r\n" +
    "; Script generated by mkv-sub-extractor\r\n" +
    "ScriptType: v4.00+\r\n" +
    "PlayResX: 1920\r\n" +
    "PlayResY: 1080\r\n" +
    "WrapStyle: 0\r\n" +
    "ScaledBorderAndShadow: yes\r\n" +
    "\r\n" +
    "[V4+ Styles]\r\n" +
    "Format: Name, Fontname, Fontsize, PrimaryColour, SecondaryColour, " +
    "OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, " +
    "ScaleX, ScaleY, Spacing, Angle, BorderStyle, Outline, Shadow, " +
    "Alignment, MarginL, MarginR, MarginV, Encoding\r\n" +
    "Style: Default,Microsoft YaHei,22," +
    "&H00FFFFFF,&H000000FF,&H00000000,&H80000000," + // White primary, blue secondary, black outline, semi-transparent back
    "0,0,0,0," +   // Not bold, not italic, not underline, not strikeout
    "100,100,0,0," + // ScaleX 100%, ScaleY 100%, Spacing 0, Angle 0
    "1,2,1," +       // BorderStyle=outline, Outline=2px, Shadow=1px
    "2," +           // Alignment=2 (bottom-center, numpad style)
    "10,10,10," +    // MarginL=10, MarginR=10, MarginV=10
    "1\r\n"          // Encoding=1 (UTF-8)
```

### SRT HTML Tag to ASS Override Tag Conversion

```go
// Source: SRT de facto standard + ASS override tag syntax
// Handles: <b>, </b>, <i>, </i>, <u>, </u>, <font color="#RRGGBB">, </font>
func convertSRTTagsToASS(text string) string {
    // Simple replacements for basic tags
    text = strings.ReplaceAll(text, "<b>", "{\\b1}")
    text = strings.ReplaceAll(text, "</b>", "{\\b0}")
    text = strings.ReplaceAll(text, "<i>", "{\\i1}")
    text = strings.ReplaceAll(text, "</i>", "{\\i0}")
    text = strings.ReplaceAll(text, "<u>", "{\\u1}")
    text = strings.ReplaceAll(text, "</u>", "{\\u0}")

    // Font color: <font color="#RRGGBB"> -> {\c&HBBGGRR&}
    // Use regex for font tags
    fontColorRe := regexp.MustCompile(`<font\s+color=["']#([0-9a-fA-F]{6})["']>`)
    text = fontColorRe.ReplaceAllStringFunc(text, func(match string) string {
        submatch := fontColorRe.FindStringSubmatch(match)
        if len(submatch) < 2 {
            return ""
        }
        hex := submatch[1]
        // Reverse RGB to BGR: RRGGBB -> BBGGRR
        return fmt.Sprintf("{\\c&H%s%s%s&}", hex[4:6], hex[2:4], hex[0:2])
    })

    // Close font tag -> reset color
    text = strings.ReplaceAll(text, "</font>", "{\\c}")

    // Strip any remaining unrecognized HTML tags
    stripRe := regexp.MustCompile(`<[^>]*>`)
    text = stripRe.ReplaceAllString(text, "")

    return text
}
```

### Output File Naming

```go
// Source: User decision - {video_basename}.{lang_code}.ass
// Collision: track name first, sequence number fallback
func GenerateOutputPath(videoPath string, track SubtitleTrack, existingPaths map[string]bool) string {
    dir := filepath.Dir(videoPath)
    base := strings.TrimSuffix(filepath.Base(videoPath), filepath.Ext(videoPath))
    lang := track.Language
    if lang == "" || lang == "und" {
        lang = "und"
    }

    candidate := filepath.Join(dir, fmt.Sprintf("%s.%s.ass", base, lang))

    if !existingPaths[candidate] {
        existingPaths[candidate] = true
        return candidate
    }

    // Collision: try track name
    if track.Name != "" {
        safeName := sanitizeFileName(track.Name)
        candidate = filepath.Join(dir, fmt.Sprintf("%s.%s.%s.ass", base, lang, safeName))
        if !existingPaths[candidate] {
            existingPaths[candidate] = true
            return candidate
        }
    }

    // Fallback: sequence number
    for i := 2; ; i++ {
        candidate = filepath.Join(dir, fmt.Sprintf("%s.%s.%d.ass", base, lang, i))
        if !existingPaths[candidate] {
            existingPaths[candidate] = true
            return candidate
        }
    }
}
```

### SSA V4 -> ASS V4+ Header Conversion

```go
// Source: SubStation Alpha format wiki (fileformats.fandom.com/wiki/SubStation_Alpha)
// SSA V4 Style fields (18): Name, Fontname, Fontsize, PrimaryColour, SecondaryColour,
//   TertiaryColour, BackColour, Bold, Italic, BorderStyle, Outline, Shadow,
//   Alignment, MarginL, MarginR, MarginV, AlphaLevel, Encoding
// ASS V4+ Style fields (23): Name, Fontname, Fontsize, PrimaryColour, SecondaryColour,
//   OutlineColour, BackColour, Bold, Italic, Underline, StrikeOut, ScaleX, ScaleY,
//   Spacing, Angle, BorderStyle, Outline, Shadow, Alignment, MarginL, MarginR,
//   MarginV, Encoding

func convertSSAHeaderToASS(header string) string {
    // 1. Replace ScriptType
    header = strings.Replace(header, "ScriptType: v4.00", "ScriptType: v4.00+", 1)

    // 2. Replace [V4 Styles] -> [V4+ Styles]
    header = strings.Replace(header, "[V4 Styles]", "[V4+ Styles]", 1)

    // 3. Update Format line and Style lines
    //    Parse existing styles, remap fields, add missing V4+ fields with defaults
    //    TertiaryColour -> OutlineColour
    //    Remove AlphaLevel
    //    Add: Underline(0), StrikeOut(0), ScaleX(100), ScaleY(100),
    //         Spacing(0), Angle(0) -- inserted between Italic and BorderStyle

    // 4. In [Events], replace "Marked" field with "Layer" field
    //    Marked=0 -> Layer=0

    return convertedHeader
}
```

## Discretion Recommendations

### Line Endings: CRLF
**Recommendation:** Use CRLF (`\r\n`) for ASS output files.
**Rationale:** The ASS format originated on Windows (Sub Station Alpha was a Windows application). The libass ASS File Format Guide states files support "either Unix-style or DOS-style line endings." However, the ASS ecosystem convention is CRLF: Aegisub (the primary ASS editor) writes CRLF, most ASS files in the wild use CRLF, and the original SSA spec was Windows-native. CRLF is also safe cross-platform -- all major subtitle renderers (libass/mpv, VSFilter) handle both, but CRLF is the traditional convention.
**Confidence:** MEDIUM -- libass documentation confirms both are supported. CRLF is the safer/more conventional choice.

### Font Size: 22
**Recommendation:** Use font size 22 for the default SRT-to-ASS style.
**Rationale:** The user specified 20-24 range. At PlayResY=1080, font size 22 produces readable subtitles at bottom-center that match the visual appearance of mpv's default subtitle rendering (~3% of screen height). Size 20 can feel small on 1080p, size 24 can feel large. 22 is the middle ground.
**Confidence:** MEDIUM -- visual preference, testable.

### PlayResX/PlayResY: 1920x1080
**Recommendation:** Use PlayResX=1920, PlayResY=1080 for SRT-to-ASS generated headers.
**Rationale:** This matches the most common video resolution (1080p). The PlayRes values define the coordinate system for positioning and font size scaling. Using 1920x1080 means font sizes and margins map 1:1 to pixels at 1080p, and scale proportionally at other resolutions. The alternative 384x288 is legacy (from the original SSA era) and makes font sizes harder to reason about. Using video resolution as PlayRes is the modern convention (Aegisub defaults to video resolution).
**Confidence:** HIGH -- 1920x1080 is the modern standard for ASS authoring at 1080p resolution.

### SSA -> ASS V4+ Conversion Specifics
**Recommendation:** Perform the following field mapping:
1. `[V4 Styles]` -> `[V4+ Styles]`
2. `ScriptType: v4.00` -> `ScriptType: v4.00+`
3. Style Format field mapping:
   - `TertiaryColour` -> `OutlineColour` (same value)
   - Remove `AlphaLevel` field
   - Add after `Italic`: `Underline` (default: 0), `StrikeOut` (default: 0)
   - Add after `StrikeOut`: `ScaleX` (default: 100), `ScaleY` (default: 100), `Spacing` (default: 0), `Angle` (default: 0)
4. Events: `Marked` field -> `Layer` field (value: `Marked=0` becomes `0`, `Marked=1` stays `1` but mapped to Layer)
5. SSA Alignment to ASS Alignment: SSA uses a different numbering scheme. SSA values 1-3 = bottom-left/center/right, 5-7 = top-left/center/right, 9-11 = middle-left/center/right. ASS uses numpad: 1-3 = bottom, 4-6 = middle, 7-9 = top. Conversion: SSA 1->1, 2->2, 3->3, 5->7, 6->8, 7->9, 9->4, 10->5, 11->6.
**Confidence:** HIGH -- well-documented in format specifications.

### SRT HTML Tag Edge Cases
**Recommendation:** Use a tolerant approach:
1. **Nested tags** (`<b><i>text</i></b>`): Process each tag independently. Both work fine in ASS override tags: `{\b1}{\i1}text{\i0}{\b0}`.
2. **Overlapping tags** (`<b><i>text</b></i>`): Same approach -- just emit open/close tags in order encountered. ASS override tags are stateful and will work correctly.
3. **Unclosed tags** (`<b>text` with no `</b>`): Do not add auto-closing. ASS override tags persist only within a Dialogue line, so unclosed tags naturally reset at line boundaries.
4. **Malformed tags** (`<font color=red>` without quotes, `<font size=5>`): For `<font color=NAME>`, support a small set of named colors (red, blue, green, yellow, white, cyan, magenta). For unrecognized font attributes, strip the tag entirely.
5. **Case insensitivity**: Handle `<B>`, `<I>`, `<FONT>` as equivalent to lowercase.
**Confidence:** MEDIUM -- de facto SRT behavior varies across players, but this approach is safe and compatible.

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| SSA v4 (Sub Station Alpha) | ASS v4+ (Advanced Sub Station Alpha) | ~2004 | SSA is legacy. All modern subtitle work uses ASS. SSA tracks in MKV should be converted to ASS for compatibility. |
| PlayRes 384x288 | PlayRes matching video resolution (1920x1080, 1280x720) | ~2010s | Legacy PlayRes values from 640x480 era. Modern ASS files use video-native resolution. |
| `Timecode` element name | `Timestamp` element name | RFC 9559 (Oct 2024) | Matroska spec standardized naming. `matroska-go` uses legacy `TimecodeScale` naming internally. |
| UTF-8 BOM for ASS files | UTF-8 without BOM | Modern convention | Libass guide says BOM "SHOULD be added" for editor compatibility, but the user decision is UTF-8 without BOM. Modern editors handle both. |

**Deprecated/outdated:**
- SSA v4 `[V4 Styles]` section: Should be converted to ASS `[V4+ Styles]` when encountered
- `AlphaLevel` style field: Removed in ASS v4+, replaced by alpha channel in color values
- `Marked` event field: Replaced by `Layer` in ASS v4+

## Open Questions

1. **matroska-go Packet.StartTime/EndTime reliability**
   - What we know: The example/extracter uses these directly as nanoseconds and divides by 10^9. This strongly suggests they ARE proper nanosecond values, not raw IEEE 754 bits (unlike Duration).
   - What's unclear: Whether the Duration bug (ReadUInt on float) affects any other Packet fields. The Timestamp element in Clusters is uint (not float), so it should be fine. But verify early with a real file.
   - Recommendation: Write a small test that extracts 2-3 subtitle packets from a known MKV file and prints their StartTime/EndTime. If values are sensible (within file duration range), proceed. If they're astronomical, apply the Float-from-bits workaround.
   - **Confidence:** MEDIUM-HIGH -- likely fine, but verify.

2. **matroska-go behavior with missing BlockDuration**
   - What we know: The library fills `Packet.EndTime`. If BlockDuration is missing, EndTime might be 0 or equal to StartTime.
   - What's unclear: What value matroska-go sets for EndTime when BlockDuration is absent.
   - Recommendation: The user decision specifies gap-fill (use next subtitle's start time). Implement this regardless: if EndTime <= StartTime, apply gap-fill. This handles both library behavior and poorly-muxed files.
   - **Confidence:** MEDIUM -- need to test with a file that has missing BlockDuration.

3. **CodecPrivate [Events] section inclusion**
   - What we know: The Matroska spec says CodecPrivate stores "Script Info and the Styles list." But some muxers may include more.
   - What's unclear: Whether any real-world files have partial [Events] sections in CodecPrivate.
   - Recommendation: Defensively check for `[Events]` in CodecPrivate content. If found, do not duplicate the section header.
   - **Confidence:** MEDIUM -- edge case, but cheap to handle.

## Sources

### Primary (HIGH confidence)
- [Matroska Subtitle Specification](https://www.matroska.org/technical/subtitles.html) - Block data field order (ReadOrder, Layer, Style, ..., Text), CodecPrivate contents, S_TEXT/UTF8 storage format
- [Matroska Codec Specifications](https://www.matroska.org/technical/codec_specs.html) - Codec ID definitions, CodecPrivate structure per codec
- [libass ASS File Format Guide](https://github.com/libass/libass/wiki/ASS-File-Format-Guide) - Authoritative modern ASS format specification: timestamp format, section structure, V4+ style fields, line endings, alignment numbering, Encoding field
- [matroska-go pkg.go.dev](https://pkg.go.dev/github.com/luispater/matroska-go) - Packet struct (Track, StartTime, EndTime, Data, Flags), Demuxer.ReadPacket(), TrackInfo.CodecPrivate
- [matroska-go example/extracter](https://github.com/luispater/matroska-go/blob/v1.2.4/example/extracter/main.go) - Working subtitle extraction: formatSRTTime treats StartTime as nanoseconds, ReadPacket loop pattern

### Secondary (MEDIUM confidence)
- [SubStation Alpha - File Formats Wiki](https://fileformats.fandom.com/wiki/SubStation_Alpha) - SSA v4 vs ASS v4+ field differences, style format mapping, Marked-to-Layer conversion
- [SubStation Alpha - MultimediaWiki](https://wiki.multimedia.cx/index.php/SubStation_Alpha) - SSA/ASS format comparison
- [ASS File Format Specification (tcax.org)](http://www.tcax.org/docs/ass-specs.htm) - Older but comprehensive ASS spec reference
- [go-astisub GitHub](https://github.com/asticode/go-astisub) - WriteToSSA implementation, StyleAttributes struct, Metadata SSA fields

### Tertiary (LOW confidence)
- matroska-go internal behavior with missing BlockDuration - Unverified, requires testing
- CodecPrivate partial [Events] section inclusion - Theoretically possible per spec, unconfirmed in practice
- SSA alignment number conversion accuracy - Based on format wiki, may need validation with real SSA files

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - matroska-go ReadPacket API verified from example code + pkg.go.dev; ASS format well-documented
- Architecture: HIGH - Straightforward extraction pipeline; two clear paths (passthrough vs. conversion)
- Pitfalls: HIGH - Timestamp handling, block data parsing, SSA conversion all well-documented with known edge cases
- Code examples: MEDIUM-HIGH - Based on verified API patterns; exact matroska-go behavior with subtitle timestamps needs testing with real files
- SSA-to-ASS conversion: MEDIUM - Field mapping documented across multiple sources but some edge cases may exist in real files

**Research date:** 2026-02-16
**Valid until:** 2026-03-16 (30 days - ASS/SSA format is stable and unchanged for decades; Matroska spec is finalized as RFC 9559)
