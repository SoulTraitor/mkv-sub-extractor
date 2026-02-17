# Phase 1: MKV Parsing and Track Discovery - Research

**Researched:** 2026-02-16
**Domain:** EBML/Matroska container parsing, track metadata extraction, subtitle track identification
**Confidence:** HIGH

## Summary

Phase 1 requires opening an MKV file, parsing its EBML/Matroska structure, and listing all subtitle tracks with metadata (track number, language as human-readable name, format type, track name, codec ID, default/forced flags). Additionally, basic file metadata must be displayed before the track listing (file name, file size, duration, subtitle track count). Non-text subtitle tracks (PGS/VobSub/DVB) must appear in the listing but be clearly marked as image-based and not extractable.

The `luispater/matroska-go` v1.2.4 library provides a high-level Demuxer API that directly maps to every requirement in this phase. The `NewDemuxer()` -> `GetFileInfo()` -> `GetNumTracks()` + `GetTrackInfo(i)` flow delivers all needed metadata without writing any EBML parsing code. The library handles VINT decoding, unknown-size elements, SeekHead navigation, and element ordering internally. For language display, `golang.org/x/text/language` + `golang.org/x/text/language/display` provides conversion from ISO 639-2 three-letter codes to both English names and native-script names (e.g., "eng" -> "English", "chi" -> "中文").

**Primary recommendation:** Use `matroska-go` v1.2.4 for all MKV parsing. Do NOT write a custom EBML parser for Phase 1. Use `golang.org/x/text` for language name resolution. Implement human-readable file size formatting as a simple utility function (no external dependency needed).

<user_constraints>
## User Constraints (from CONTEXT.md)

### Locked Decisions
- **Track listing format**: Numbered list format (not table), e.g. `[1] English (SRT) "Commentary" [default]`
- Each track shows: track number, language (human-readable name like "English"/"中文"), format type (SRT/ASS/SSA), track name (if present), codec ID (e.g. S_TEXT/UTF8), default/forced flags
- If language tag is absent, display "Unknown"
- **Non-text subtitle handling**: Image subtitle tracks (PGS/VobSub/DVB) are shown in the listing but clearly marked as "not extractable (image subtitle)"
- Image subtitle tracks are not selectable -- their numbers are skipped in selection
- If an MKV has only image subtitles and no text subtitles: display all tracks with their markers, then print a message: "未找到可提取的文本字幕轨道"
- **File metadata display**: Before listing tracks, show basic file information: file name, file size (human-readable), duration (e.g. "1:42:30"), number of subtitle tracks

### Claude's Discretion
- Exact formatting and spacing of the track listing
- How to display codec ID alongside format type (inline vs separate line)
- Duration parsing approach from Matroska Segment/Info elements
- Color or styling of terminal output (if any)

### Deferred Ideas (OUT OF SCOPE)
None -- discussion stayed within phase scope
</user_constraints>

## Standard Stack

### Core

| Library | Version | Purpose | Why Standard |
|---------|---------|---------|--------------|
| `github.com/luispater/matroska-go` | v1.2.4 | MKV container demuxing, EBML parsing, track enumeration | Pull-based Demuxer API with `GetTrackInfo()`, `GetFileInfo()`, `ReadPacket()`. Handles all EBML complexity (VINT, unknown-size, SeekHead) internally. MIT licensed, requires Go 1.24+, last updated Aug 2025. TrackInfo provides Number, Type, CodecID, CodecPrivate, Language, Name, Default, Forced fields -- all needed for Phase 1. |
| `golang.org/x/text` | latest | ISO 639-2 language code to human-readable name conversion | `language.ParseBase("eng")` parses 3-letter ISO 639 codes. `display.Self.Name(tag)` returns native-script name ("中文", "日本語"). `display.English.Languages().Name(tag)` returns English name. Official Go extended library, comprehensive CLDR data. |

### Supporting

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `os` (stdlib) | N/A | File opening, `os.Stat` for file size | File metadata (name, size) |
| `fmt` (stdlib) | N/A | Formatted output for track listing | Display formatting |
| `io` (stdlib) | N/A | `io.ReadSeeker` interface, `io.EOF` | Demuxer input, packet reading |
| `time` (stdlib) | N/A | `time.Duration` for duration formatting | Convert nanoseconds to "H:MM:SS" |
| `strings` (stdlib) | N/A | String manipulation | Codec ID parsing, display formatting |
| `math` (stdlib) | N/A | `math.Log`, `math.Pow` for file size formatting | Human-readable file size |

### Alternatives Considered

| Instead of | Could Use | Tradeoff |
|------------|-----------|----------|
| `matroska-go` | `remko/go-mkvparse` | Push-based event handler requires ~3x more code; must build state machine from individual EBML element callbacks; last updated Jan 2022 |
| `matroska-go` | Custom EBML parser | Months of work, EBML spec has 200+ element types; `matroska-go` handles all of this |
| `golang.org/x/text` | `barbashov/iso639-3` | Only provides English names; no native-script display; `x/text` is official Go project with CLDR data |
| `golang.org/x/text` | Hard-coded language map | Incomplete, maintenance burden; 440+ ISO 639-2 codes exist |
| `dustin/go-humanize` | Stdlib utility function | Adding dependency for one function (`Bytes()`) is overkill; 10-line stdlib function covers this |

**Installation:**
```bash
go mod init mkv-sub-extractor
go get github.com/luispater/matroska-go@v1.2.4
go get golang.org/x/text@latest
```

## Architecture Patterns

### Recommended Project Structure (Phase 1 scope)
```
mkv-sub-extractor/
  cmd/
    mkv-sub-extractor/
      main.go              -- Entry point (minimal for Phase 1 -- just file path arg)
  pkg/
    mkvinfo/
      mkvinfo.go           -- MKV file info and track listing logic
      mkvinfo_test.go
      types.go             -- FileInfo, SubtitleTrack types
      language.go           -- ISO 639 language name resolution
      language_test.go
      format.go            -- Human-readable file size, duration formatting
      format_test.go
      display.go           -- Track listing output formatting
      display_test.go
  go.mod
  go.sum
```

### Pattern 1: Wrapper Types Over Library Types
**What:** Define our own `FileInfo` and `SubtitleTrack` types that hold only the data we need, populated from `matroska-go`'s `SegmentInfo` and `TrackInfo`.
**When to use:** Always -- decouple display logic from library internals.
**Why:** If `matroska-go` is swapped for `go-mkvparse`, only the population code changes; display logic stays the same.
**Example:**
```go
// pkg/mkvinfo/types.go

// FileInfo holds basic file metadata for display.
type FileInfo struct {
    FileName       string
    FileSize       int64          // bytes
    Duration       time.Duration  // from SegmentInfo.Duration (nanoseconds)
    SubtitleCount  int            // total subtitle tracks (text + image)
    TextSubCount   int            // extractable text subtitle tracks only
}

// SubtitleTrack holds metadata for a single subtitle track.
type SubtitleTrack struct {
    Number       uint8          // Matroska track number
    Index        int            // Display index (1-based, sequential)
    Language     string         // ISO 639-2 code (e.g., "eng", "jpn")
    LanguageName string         // Human-readable name (e.g., "English", "中文")
    FormatType   string         // "SRT", "ASS", "SSA", "PGS", "VobSub", "DVB", etc.
    CodecID      string         // Raw codec ID (e.g., "S_TEXT/UTF8")
    Name         string         // Track name from metadata (may be empty)
    IsDefault    bool           // FlagDefault
    IsForced     bool           // FlagForced
    IsText       bool           // true for S_TEXT/* codecs, false for image subtitles
    IsExtractable bool          // true if text-based and extractable
}
```

### Pattern 2: Codec ID Classification
**What:** Map Matroska codec ID strings to human-readable format type and text/image classification.
**When to use:** When populating SubtitleTrack from TrackInfo.
**Example:**
```go
// Source: Matroska subtitle codec specification
// https://www.matroska.org/technical/subtitles.html

type codecInfo struct {
    FormatType  string // Human-readable name
    IsText      bool   // true = text, false = image/bitmap
}

var codecMap = map[string]codecInfo{
    "S_TEXT/UTF8":    {FormatType: "SRT", IsText: true},
    "S_TEXT/SSA":     {FormatType: "SSA", IsText: true},
    "S_TEXT/ASS":     {FormatType: "ASS", IsText: true},
    "S_TEXT/WEBVTT":  {FormatType: "WebVTT", IsText: true},
    "S_TEXT/USF":     {FormatType: "USF", IsText: true},
    "S_HDMV/PGS":    {FormatType: "PGS", IsText: false},
    "S_VOBSUB":      {FormatType: "VobSub", IsText: false},
    "S_DVBSUB":      {FormatType: "DVB", IsText: false},
    "S_HDMV/TEXTST": {FormatType: "HDMV Text", IsText: true},
    "S_ARIBSUB":     {FormatType: "ARIB", IsText: true},
}
```

### Pattern 3: Language Name Resolution with Fallback Chain
**What:** Convert ISO 639-2 language codes to human-readable names, with graceful fallback.
**When to use:** Every track display.
**Why:** Matroska stores language as ISO 639-2 three-letter codes (e.g., "eng", "jpn", "chi"). Users want to see "English", "中文", etc. The `matroska-go` library may also provide LanguageBCP47 in future versions.
**Example:**
```go
// Source: golang.org/x/text/language and golang.org/x/text/language/display
import (
    "golang.org/x/text/language"
    "golang.org/x/text/language/display"
)

// resolveLanguageName converts an ISO 639-2 code to a display name.
// Returns the native-script name (e.g., "中文" for Chinese).
// Falls back to English name, then raw code, then "Unknown".
func resolveLanguageName(code string) string {
    if code == "" || code == "und" {
        return "Unknown"
    }
    base, err := language.ParseBase(code)
    if err != nil {
        return code // Return raw code if unparseable
    }
    tag := language.Make(base.String())

    // Try native-script name first (e.g., "中文", "日本語")
    name := display.Self.Name(tag)
    if name != "" {
        return name
    }
    // Fallback to English name
    name = display.English.Languages().Name(tag)
    if name != "" {
        return name
    }
    return code
}
```

### Pattern 4: Track Listing Display Format
**What:** Format track listing according to user's locked decision.
**When to use:** Output display.
**Example:**
```go
// User decision: numbered list format
// [1] English (SRT) "Commentary" S_TEXT/UTF8 [default]
// [2] 中文 (ASS) S_TEXT/ASS
// [3] PGS -- not extractable (image subtitle)

func formatTrackLine(t SubtitleTrack) string {
    var parts []string

    // Language and format
    if t.IsText {
        parts = append(parts, fmt.Sprintf("[%d] %s (%s)", t.Index, t.LanguageName, t.FormatType))
    } else {
        parts = append(parts, fmt.Sprintf("[%d] %s (%s)", t.Index, t.LanguageName, t.FormatType))
    }

    // Track name (if present)
    if t.Name != "" {
        parts = append(parts, fmt.Sprintf("%q", t.Name))
    }

    // Codec ID
    parts = append(parts, t.CodecID)

    // Flags
    if t.IsDefault {
        parts = append(parts, "[default]")
    }
    if t.IsForced {
        parts = append(parts, "[forced]")
    }

    // Image subtitle marker
    if !t.IsText {
        parts = append(parts, "-- not extractable (image subtitle)")
    }

    return strings.Join(parts, " ")
}
```

### Anti-Patterns to Avoid

- **Writing a custom EBML parser:** `matroska-go` handles all EBML complexity. Writing our own parser is months of work for zero benefit in Phase 1. The library provides `GetTrackInfo()` which gives us everything we need.
- **Loading entire MKV into memory:** `matroska-go` uses `io.ReadSeeker` internally and streams through the file. We just open an `os.File` and pass it in.
- **Parsing tracks manually from EBML elements:** The library's `GetNumTracks()` + `GetTrackInfo(i)` API does this. Do NOT use the lower-level `EBMLReader` or `MatroskaParser` unless the `Demuxer` API proves insufficient.
- **Adding `go-humanize` just for file size formatting:** A 10-line function using `math.Log` handles this. Do not add a dependency for one utility.

## Don't Hand-Roll

| Problem | Don't Build | Use Instead | Why |
|---------|-------------|-------------|-----|
| EBML/VINT decoding | Custom VINT parser | `matroska-go` Demuxer | VINT has subtle rules (marker bit handling, unknown-size sentinel, ID vs size distinction); library is tested |
| Matroska element navigation | Custom element scanner | `matroska-go` Demuxer | SeekHead, unknown-size Segments, element ordering -- all handled by library |
| Track metadata extraction | Parse TrackEntry elements | `matroska-go` GetTrackInfo() | Library returns populated TrackInfo with all fields |
| ISO 639-2 to language name | Hard-coded map of ~440 codes | `golang.org/x/text/language/display` | CLDR data, native-script names, maintained by Go team |
| Duration from SegmentInfo | Manual TimestampScale calculation | `matroska-go` GetFileInfo().Duration | Library returns Duration in nanoseconds, already scaled |

**Key insight:** Phase 1 is about displaying track information, not about parsing EBML. The library does the hard work; our code formats the output.

## Common Pitfalls

### Pitfall 1: FlagDefault Default Value is 1 (true)
**What goes wrong:** In Matroska, if the FlagDefault element is absent from a TrackEntry, the default value is 1 (true/enabled). This means ALL tracks without an explicit FlagDefault=0 are considered "default". Most MKV files have multiple tracks all marked as default. Displaying `[default]` for every track is noisy and unhelpful.
**Why it happens:** The Matroska spec (RFC 9559, Element ID 0x88) defines FlagDefault with default=1.
**How to avoid:** The `matroska-go` library's `TrackInfo.Default` field reflects this. Consider only showing `[default]` when exactly one subtitle track has it set, or when it provides meaningful differentiation (not all tracks are default).
**Warning signs:** Every subtitle track shows `[default]` in the listing.

### Pitfall 2: Language Field Default is "eng"
**What goes wrong:** If the Language element (0x22B59C) is absent from a TrackEntry, the default value per the Matroska spec is "eng" (English). This means a track without any language tag will appear as "English" even if the content is Chinese or Japanese. The `matroska-go` library will return "eng" for tracks with no language tag.
**Why it happens:** The Matroska spec (RFC 9559) defines Language with default="eng".
**How to avoid:** Check if the language matches the spec default "eng" -- but we cannot reliably distinguish "explicitly set to eng" from "absent, defaulted to eng." Best approach: trust the value from the library. If the track has no language and the library returns "eng", display "English" -- this matches what every other tool does (mkvinfo, ffprobe, etc.).
**Warning signs:** Non-English tracks displaying as "English" with no track name to differentiate.

### Pitfall 3: SegmentInfo Duration May Be Zero or Absent
**What goes wrong:** Not all MKV files have a Duration element in their SegmentInfo. Files from streaming sources, OBS recordings, or truncated downloads may have Duration=0 or the element may be absent entirely.
**Why it happens:** Duration is an optional element. Muxers that don't do a second pass (e.g., FFmpeg pipe output) may not write it.
**How to avoid:** Check if `SegmentInfo.Duration == 0` and display "Unknown" instead of "0:00:00". The `matroska-go` library returns Duration as `uint64` in nanoseconds; 0 means unknown.
**Warning signs:** Duration displays as "0:00:00" for streaming-source files.

### Pitfall 4: Unknown Codec IDs
**What goes wrong:** An MKV file contains a subtitle track with a codec ID not in our `codecMap` (e.g., a future codec, or a non-standard codec ID from a niche muxer). The code panics or silently skips the track.
**Why it happens:** The codec map is finite but Matroska allows arbitrary codec IDs.
**How to avoid:** Use a fallback in the codec classification: unknown codec IDs should be displayed as-is with "Unknown" format type, classified as non-extractable by default. Never panic on an unknown codec ID.
**Warning signs:** Missing tracks in the listing; panics on unusual MKV files.

### Pitfall 5: TrackNumber vs. Array Index
**What goes wrong:** Matroska TrackNumber (element 0xD7) is a 1-based identifier that can be non-sequential (e.g., track numbers 1, 3, 5 with no 2 or 4). The `matroska-go` `GetTrackInfo(i)` uses a 0-based array index `i` (0 to `GetNumTracks()-1`), NOT the TrackNumber. Confusing the two leads to wrong track references.
**Why it happens:** Track numbers in Matroska are arbitrary identifiers, not indices. They identify which blocks belong to which track in Clusters.
**How to avoid:** Use the array index for iteration, but display `TrackInfo.Number` as the track identifier. Our `SubtitleTrack.Index` (display index, 1-based sequential) may differ from `TrackInfo.Number` (Matroska track number).
**Warning signs:** User selects track "3" but gets data from track "1".

### Pitfall 6: x/text/language ParseBase and Bibliographic vs. Terminology Codes
**What goes wrong:** ISO 639-2 has two variants for some languages: bibliographic (e.g., "chi" for Chinese) and terminology (e.g., "zho" for Chinese). Matroska uses the bibliographic form ("chi", "fre", "ger") while Go's `language.ParseBase()` may normalize them differently.
**Why it happens:** Historical dual-code system in ISO 639-2.
**How to avoid:** `language.ParseBase()` handles both "chi" and "zho" correctly, normalizing to the BCP 47 base tag "zh". Test with the common bibliographic codes: "chi" -> "zh" (Chinese), "fre" -> "fr" (French), "ger" -> "de" (German), "dut" -> "nl" (Dutch), "cze" -> "cs" (Czech).
**Warning signs:** Language names not resolving for French/German/Chinese tracks.

## Code Examples

Verified patterns from official sources:

### Opening MKV and Getting File Info
```go
// Source: matroska-go pkg.go.dev documentation + example/extracter
import (
    "os"
    matroska "github.com/luispater/matroska-go"
)

func getFileInfo(path string) (*FileInfo, error) {
    file, err := os.Open(path)
    if err != nil {
        return nil, fmt.Errorf("cannot open file: %w", err)
    }
    defer file.Close()

    // Get file size from os.Stat
    stat, err := file.Stat()
    if err != nil {
        return nil, fmt.Errorf("cannot stat file: %w", err)
    }

    // Create demuxer (parses EBML header and Segment metadata)
    demuxer, err := matroska.NewDemuxer(file)
    if err != nil {
        return nil, fmt.Errorf("cannot parse MKV: %w", err)
    }
    defer demuxer.Close()

    // Get segment info (contains duration)
    segInfo, err := demuxer.GetFileInfo()
    if err != nil {
        return nil, fmt.Errorf("cannot read file info: %w", err)
    }

    info := &FileInfo{
        FileName: filepath.Base(path),
        FileSize: stat.Size(),
        Duration: time.Duration(segInfo.Duration), // Already in nanoseconds
    }

    return info, nil
}
```

### Enumerating Subtitle Tracks
```go
// Source: matroska-go pkg.go.dev documentation
func listSubtitleTracks(demuxer *matroska.Demuxer) ([]SubtitleTrack, error) {
    numTracks, err := demuxer.GetNumTracks()
    if err != nil {
        return nil, err
    }

    var tracks []SubtitleTrack
    displayIndex := 1

    for i := uint(0); i < numTracks; i++ {
        info, err := demuxer.GetTrackInfo(i)
        if err != nil {
            continue // Skip tracks that can't be read
        }

        // Filter to subtitle tracks only (Type == 17)
        if info.Type != matroska.TypeSubtitle {
            continue
        }

        codec := classifyCodec(info.CodecID)

        track := SubtitleTrack{
            Number:        info.Number,
            Index:         displayIndex,
            Language:      info.Language,
            LanguageName:  resolveLanguageName(info.Language),
            FormatType:    codec.FormatType,
            CodecID:       info.CodecID,
            Name:          info.Name,
            IsDefault:     info.Default,
            IsForced:      info.Forced,
            IsText:        codec.IsText,
            IsExtractable: codec.IsText, // Only text subtitles are extractable
        }

        tracks = append(tracks, track)
        displayIndex++
    }

    return tracks, nil
}
```

### Human-Readable File Size (No External Dependency)
```go
// No external dependency needed -- simple stdlib function.
func formatFileSize(bytes int64) string {
    if bytes < 0 {
        return "Unknown"
    }
    const unit = 1024
    if bytes < unit {
        return fmt.Sprintf("%d B", bytes)
    }
    div, exp := int64(unit), 0
    for n := bytes / unit; n >= unit; n /= unit {
        div *= unit
        exp++
    }
    return fmt.Sprintf("%.1f %cB", float64(bytes)/float64(div), "KMGTPE"[exp])
}
// 1073741824 -> "1.0 GB"
// 1288490189 -> "1.2 GB"
```

### Duration Formatting
```go
// Convert nanosecond duration to "H:MM:SS" format.
func formatDuration(d time.Duration) string {
    if d <= 0 {
        return "Unknown"
    }
    h := int(d.Hours())
    m := int(d.Minutes()) % 60
    s := int(d.Seconds()) % 60
    return fmt.Sprintf("%d:%02d:%02d", h, m, s)
}
// 6150000000000 ns (1:42:30) -> "1:42:30"
```

### Language Name Resolution
```go
// Source: golang.org/x/text/language and golang.org/x/text/language/display
import (
    "golang.org/x/text/language"
    "golang.org/x/text/language/display"
)

func resolveLanguageName(code string) string {
    if code == "" || code == "und" {
        return "Unknown"
    }

    base, err := language.ParseBase(code)
    if err != nil {
        return code
    }

    tag := language.Make(base.String())

    // Native-script name: "English", "中文", "日本語"
    if name := display.Self.Name(tag); name != "" {
        return name
    }

    // Fallback to English name
    namer := display.English.Languages()
    if name := namer.Name(tag); name != "" {
        return name
    }

    return code
}
// "eng" -> "English"
// "chi" -> "中文"
// "jpn" -> "日本語"
// "und" -> "Unknown"
// "fre" -> "francais"  (via bibliographic -> BCP 47 "fr")
```

## State of the Art

| Old Approach | Current Approach | When Changed | Impact |
|--------------|------------------|--------------|--------|
| Matroska draft spec | RFC 9559 (final standard) | October 2024 | Matroska is now an IETF standard, not a draft. Element IDs are frozen. |
| EBML RFC 8794 only | RFC 8794 updated by RFC 9559 | October 2024 | RFC 9559 permits previously reserved Element IDs |
| Language element only (ISO 639-2) | LanguageBCP47 element added (0x22B59D) | RFC 9559 | Modern files may include BCP 47 language tags alongside legacy ISO 639-2. `matroska-go` v1.2.4 does NOT expose LanguageBCP47 -- use Language field. |
| No accessibility flags | FlagHearingImpaired (0x55AB), FlagVisualImpaired (0x55AC), FlagTextDescriptions (0x55AD) | RFC 9559 | New elements in Matroska v4+. `matroska-go` likely does not expose these yet. Not needed for Phase 1 but good to know. |

**Deprecated/outdated:**
- `Timecode` element name: renamed to `Timestamp` in RFC 9559 (same element ID 0xE7, same semantics)
- `TimecodeScale` element name: renamed to `TimestampScale` in RFC 9559 (same element ID 0x2AD7B1). `matroska-go` uses `TimecodeScale` naming internally.

## Matroska Element ID Reference (Phase 1 Relevant)

Verified against RFC 9559 and the official Matroska Element Specification:

| Element | EBML ID | Type | Default | Notes |
|---------|---------|------|---------|-------|
| Segment | 0x18538067 | master | -- | Top-level container, often unknown-size |
| SeekHead | 0x114D9B74 | master | -- | Optional index for locating other elements |
| Info | 0x1549A966 | master | -- | Segment-level metadata |
| TimestampScale | 0x2AD7B1 | uint | 1000000 | Nanoseconds per tick; default = 1ms/tick |
| Duration | 0x4489 | float | -- | In TimestampScale units; optional |
| Tracks | 0x1654AE6B | master | -- | Contains all TrackEntry elements |
| TrackEntry | 0xAE | master | -- | One per track |
| TrackNumber | 0xD7 | uint | -- | 1-based track identifier |
| TrackUID | 0x73C5 | uint | -- | Globally unique track ID |
| TrackType | 0x83 | uint | -- | 1=video, 2=audio, 17=subtitle |
| CodecID | 0x86 | string | -- | e.g. "S_TEXT/UTF8" |
| CodecPrivate | 0x63A2 | binary | -- | Codec-specific data (ASS header) |
| Language | 0x22B59C | string | "eng" | ISO 639-2 three-letter code |
| LanguageBCP47 | 0x22B59D | string | -- | BCP 47 tag (newer, optional) |
| Name | 0x536E | UTF-8 | -- | Human-readable track name |
| FlagDefault | 0x88 | uint | 1 | Track eligible for auto-selection |
| FlagForced | 0x55AA | uint | 0 | Track for forced display |

**Subtitle Codec IDs (from matroska.org/technical/subtitles.html):**

| Codec ID | Format Type | Text/Image | Extractable |
|----------|-------------|------------|-------------|
| S_TEXT/UTF8 | SRT | Text | Yes |
| S_TEXT/SSA | SSA | Text | Yes |
| S_TEXT/ASS | ASS | Text | Yes |
| S_TEXT/WEBVTT | WebVTT | Text | Yes (future) |
| S_TEXT/USF | USF | Text | Yes (future) |
| S_HDMV/PGS | PGS | Image | No |
| S_VOBSUB | VobSub | Image | No |
| S_DVBSUB | DVB | Image | No |
| S_HDMV/TEXTST | HDMV Text | Text | Yes (future) |
| S_ARIBSUB | ARIB | Text | Yes (future) |

## matroska-go API Reference (Phase 1 Relevant)

Key API surface verified from pkg.go.dev documentation:

```go
// Demuxer creation
func NewDemuxer(r io.ReadSeeker) (*Demuxer, error)
func (d *Demuxer) Close()

// File metadata
func (d *Demuxer) GetFileInfo() (*SegmentInfo, error)

// Track enumeration
func (d *Demuxer) GetNumTracks() (uint, error)
func (d *Demuxer) GetTrackInfo(track uint) (*TrackInfo, error)

// SegmentInfo fields (Phase 1 relevant)
type SegmentInfo struct {
    Duration       uint64    // nanoseconds (0 if unknown)
    TimecodeScale  uint64    // nanoseconds per tick, default 1000000
    Title          string
    // ... other fields
}

// TrackInfo fields (Phase 1 relevant)
type TrackInfo struct {
    Number       uint8     // Track number (1-based)
    Type         uint8     // 1=video, 2=audio, 17=subtitle
    CodecID      string    // e.g. "S_TEXT/UTF8"
    CodecPrivate []byte    // Codec-specific data
    Language     string    // ISO 639-2 code
    Name         string    // Human-readable track name
    Default      bool      // FlagDefault
    Forced       bool      // FlagForced
    // ... other fields (video, audio specific)
}

// Constants
const TypeSubtitle = 17
```

## Discretion Recommendations

For areas marked as Claude's Discretion in CONTEXT.md:

### Formatting and Spacing
**Recommendation:** Single-line per track, compact format. Use fixed-width bracket for index alignment.

```
File: Movie.2024.BluRay.1080p.mkv
Size: 8.2 GB | Duration: 2:15:30 | Subtitle tracks: 4

[1] English (SRT) S_TEXT/UTF8 [default]
[2] 中文 (ASS) "简体中文字幕" S_TEXT/ASS
[3] 日本語 (ASS) S_TEXT/ASS [forced]
[4] English (PGS) S_HDMV/PGS -- not extractable (image subtitle)
```

### Codec ID Display
**Recommendation:** Inline on the same line, after the format type in parentheses. Codec ID is a technical detail but useful for disambiguation. Example: `English (SRT) S_TEXT/UTF8` rather than a separate line.

### Duration Parsing
**Recommendation:** Use `SegmentInfo.Duration` from `matroska-go` which is already in nanoseconds. Convert with `time.Duration(segInfo.Duration)` then format as `H:MM:SS`. Handle Duration=0 as "Unknown".

### Terminal Styling
**Recommendation:** No color/styling in Phase 1. Keep it plain text. This avoids adding `lipgloss` as a dependency until Phase 3 when the full CLI is built. Plain text also works in all terminals and is pipe-friendly.

## Open Questions

1. **matroska-go LanguageBCP47 support**
   - What we know: The library's TrackInfo struct has a `Language` field (ISO 639-2). Web search suggests there may be a `LanguageBCP47` field but this is unconfirmed.
   - What's unclear: Whether `matroska-go` v1.2.4 parses the LanguageBCP47 element (0x22B59D).
   - Recommendation: Use the `Language` field. If LanguageBCP47 is available in a future version, prefer it over Language for more accurate language identification. This is not blocking for Phase 1.

2. **matroska-go behavior with edge-case files**
   - What we know: Library has 0 known importers on pkg.go.dev. API looks clean and well-designed. Has test data.
   - What's unclear: How it handles damaged files, unusual muxers, very large files (50+ GB), files with unknown-size Segments.
   - Recommendation: Test early with diverse MKV files (FFmpeg-muxed, OBS, MKVToolNix, HandBrake). If issues arise, fallback to `remko/go-mkvparse` with custom state machine.

3. **display.Self.Name() for rare/ancient languages**
   - What we know: Works well for major languages (eng, chi, jpn, fre, ger, etc.)
   - What's unclear: Behavior for very rare ISO 639-2 codes or extinct languages.
   - Recommendation: The fallback chain (Self.Name -> English.Name -> raw code -> "Unknown") handles this gracefully. Not a concern for practical use.

## Sources

### Primary (HIGH confidence)
- [RFC 9559 - Matroska Media Container Format Specification](https://datatracker.ietf.org/doc/rfc9559/) - Official IETF standard, October 2024. Element IDs, default values, TrackEntry structure.
- [Matroska Element Specification](https://www.matroska.org/technical/elements.html) - Generated from ebml_matroska.xml. Complete element ID reference.
- [Matroska Subtitle Specification](https://www.matroska.org/technical/subtitles.html) - Subtitle codec IDs and storage format.
- [matroska-go pkg.go.dev](https://pkg.go.dev/github.com/luispater/matroska-go) - v1.2.4 API documentation. Demuxer, TrackInfo, SegmentInfo, Packet types.
- [golang.org/x/text/language](https://pkg.go.dev/golang.org/x/text/language) - ISO 639 parsing functions.
- [golang.org/x/text/language/display](https://pkg.go.dev/golang.org/x/text/language/display) - SelfNamer, English namer for language display names.

### Secondary (MEDIUM confidence)
- [matroska-go example/extracter](https://pkg.go.dev/github.com/luispater/matroska-go/example/extracter) - Working subtitle extraction example demonstrating API usage.
- [barbashov/iso639-3](https://pkg.go.dev/github.com/barbashov/iso639-3) - Alternative ISO 639 library (not recommended; `x/text` is more comprehensive).
- [remko/go-mkvparse](https://github.com/remko/go-mkvparse) - Fallback MKV parser if matroska-go proves problematic.

### Tertiary (LOW confidence)
- matroska-go internal behavior with edge cases (unknown-size elements, damaged files) - Unverified, requires testing with real MKV files.
- LanguageBCP47 field availability in matroska-go v1.2.4 - Web search suggests it may exist but unconfirmed in pkg.go.dev docs.

## Metadata

**Confidence breakdown:**
- Standard stack: HIGH - `matroska-go` API verified from pkg.go.dev, `x/text` is official Go project
- Architecture: HIGH - Straightforward wrapper pattern over library types; no novel design needed
- Pitfalls: HIGH - FlagDefault/Language defaults verified from RFC 9559; codec IDs verified from matroska.org
- Code examples: MEDIUM-HIGH - Based on verified API signatures; exact behavior needs testing with real files

**Research date:** 2026-02-16
**Valid until:** 2026-03-16 (30 days - stack is stable, Matroska spec is finalized as RFC)
