# Architecture Patterns

**Domain:** MKV subtitle extraction CLI tool (pure Go)
**Researched:** 2026-02-16
**Confidence:** MEDIUM -- based on EBML spec (RFC 8794), Matroska spec, and Go idioms from training data. Web verification was unavailable; element IDs and structural claims should be validated against official specs during implementation.

## Recommended Architecture

A layered architecture with five distinct components, each with a clean Go interface boundary. Data flows downward through parsing layers and upward through extraction, with the CLI orchestrating the process.

```
+------------------------------------------------------------------+
|                         CLI Layer (cmd/)                          |
|  - Argument parsing, interactive menus, file discovery            |
+------------------------------------------------------------------+
            |                                        ^
            v                                        |
+------------------------------------------------------------------+
|                    Orchestrator (extractor/)                      |
|  - Coordinates parse -> identify -> extract -> convert flow      |
+------------------------------------------------------------------+
            |              |              |              |
            v              v              v              v
+----------------+ +----------------+ +----------------+ +----------------+
|  EBML Parser   | |  Matroska      | |  Subtitle      | |  Format        |
|  (ebml/)       | |  Decoder       | |  Codec         | |  Writer        |
|                | |  (matroska/)   | |  Handlers      | |  (format/ass/) |
|  - VINT decode | |  - Element IDs | |  (subtitle/)   | |                |
|  - Element     | |  - Track info  | |  - SRT parse   | |  - ASS header  |
|    scanning    | |  - Cluster     | |  - SSA/ASS     | |  - Dialogue    |
|  - Type read   | |    navigation  | |    parse       | |    lines       |
+----------------+ +----------------+ +----------------+ +----------------+
```

### Component Boundaries

| Component | Package | Responsibility | Communicates With |
|-----------|---------|---------------|-------------------|
| **EBML Parser** | `pkg/ebml` | Reads raw EBML binary: VINT decoding, element ID/size parsing, element type discrimination (master vs data), forward scanning/seeking | Matroska Decoder (provides element stream) |
| **Matroska Decoder** | `pkg/matroska` | Interprets EBML elements as Matroska semantics: identifies Segment children, parses TrackEntry, navigates Clusters, extracts Blocks | Orchestrator (provides track list + block data) |
| **Subtitle Codec Handlers** | `pkg/subtitle` | Decodes raw subtitle data from Blocks into structured subtitle events based on CodecID (S_TEXT/UTF8, S_TEXT/SSA, S_TEXT/ASS) | Orchestrator (receives raw block data, returns subtitle events) |
| **Format Writer** | `pkg/format/ass` | Converts structured subtitle events into valid ASS file output with proper headers, styles, and dialogue lines | Orchestrator (receives subtitle events, writes to io.Writer) |
| **CLI Layer** | `cmd/mkv-sub-extractor` | User interaction: arg parsing, MKV file discovery, interactive track selection menus, progress display, error reporting | Orchestrator (provides file path + track selection, receives results) |
| **Orchestrator** | `pkg/extractor` | Coordinates the full pipeline: open file, parse tracks, present to CLI, extract selected track, convert format, write output | All other components |

## Detailed Component Design

### 1. EBML Parser (`pkg/ebml`)

This is the foundation layer. EBML (Extensible Binary Meta Language, RFC 8794) is a binary format where every element has the structure:

```
[Element ID (VINT)] [Data Size (VINT)] [Data (Size bytes)]
```

**VINT (Variable-Width Integer) Encoding:**
The leading byte determines width by the position of the first set bit:

| Leading bits | Width | Usable bits |
|-------------|-------|-------------|
| `1xxxxxxx` | 1 byte | 7 |
| `01xxxxxx` | 2 bytes | 14 |
| `001xxxxx` | 3 bytes | 21 |
| `0001xxxx` | 4 bytes | 28 |
| ... | up to 8 bytes | up to 56 |

**Key types:**

```go
// Element represents a parsed EBML element header.
type Element struct {
    ID       ElementID  // Parsed element ID
    DataSize int64      // Size of the data payload (-1 for unknown size)
    Offset   int64      // Byte offset of this element in the file
    DataPos  int64      // Byte offset where data begins (after ID + size)
}

// ElementID is a raw EBML element identifier.
type ElementID uint32

// Reader wraps an io.ReadSeeker and provides EBML-level operations.
type Reader struct {
    r      io.ReadSeeker
    offset int64
}

// Key methods:
// ReadVINT() (value uint64, width int, err error)
// ReadElementHeader() (Element, error)
// ReadUint(size int64) (uint64, error)
// ReadInt(size int64) (int64, error)
// ReadFloat(size int64) (float64, error)
// ReadString(size int64) (string, error)
// ReadBytes(size int64) ([]byte, error)
// Skip(size int64) error
```

**Design decisions:**
- Use `io.ReadSeeker` (not `io.Reader`) because MKV files can be gigabytes and we must skip video/audio clusters. Seeking is non-negotiable.
- Element IDs are stored as `uint32` since Matroska uses at most 4-byte IDs.
- Unknown size (-1) must be handled because Matroska Segments and some Clusters use "unknown size" encoding (all VINT data bits set to 1).
- Do NOT load entire file into memory. Stream through, seeking past irrelevant elements.

### 2. Matroska Decoder (`pkg/matroska`)

Interprets the EBML element stream using Matroska-specific element IDs. The Matroska container has this hierarchy:

```
EBML Header (ID: 0x1A45DFA3)
  EBMLVersion
  EBMLReadVersion
  DocType = "matroska" or "webm"

Segment (ID: 0x18538067) -- usually unknown size
  SeekHead (ID: 0x114D9B74) -- optional index
  Info (ID: 0x1549A966)
    TimestampScale (ID: 0x2AD7B1) -- nanoseconds per tick, default 1000000
    Duration (ID: 0x4489)
  Tracks (ID: 0x1654AE6B)
    TrackEntry (ID: 0xAE) -- one per track
      TrackNumber (ID: 0xD7)
      TrackUID (ID: 0x73C5)
      TrackType (ID: 0x83) -- 0x11 = subtitle
      CodecID (ID: 0x86) -- e.g. "S_TEXT/UTF8"
      CodecPrivate (ID: 0x63A2) -- SSA/ASS header for those codecs
      Language (ID: 0x22B59C) -- ISO 639-2 three-letter code
      Name (ID: 0x536E) -- human-readable track name
  Cluster (ID: 0x1F43B675) -- many of these
    Timestamp (ID: 0xE7) -- cluster-level timestamp
    SimpleBlock (ID: 0xA3)
    BlockGroup (ID: 0xA0)
      Block (ID: 0xA1)
      BlockDuration (ID: 0x9B)
  Cues (ID: 0x1C53BB6B) -- seek index, optional
```

**Key types:**

```go
// TrackInfo holds parsed metadata for a single track.
type TrackInfo struct {
    Number       uint64
    UID          uint64
    Type         TrackType // Video=1, Audio=2, Subtitle=17
    CodecID      string    // "S_TEXT/UTF8", "S_TEXT/SSA", "S_TEXT/ASS"
    CodecPrivate []byte    // SSA/ASS header data
    Language     string    // "eng", "jpn", "und", etc.
    Name         string    // User-facing track name
}

// SubtitleBlock holds a raw subtitle data block with timing.
type SubtitleBlock struct {
    TrackNumber uint64
    Timestamp   int64    // Absolute timestamp in milliseconds
    Duration    int64    // Duration in milliseconds (-1 if absent)
    Data        []byte   // Raw subtitle content
}

// Decoder navigates an MKV file and extracts structural info.
type Decoder struct {
    reader         *ebml.Reader
    timestampScale uint64 // nanoseconds per timestamp tick
}

// Key methods:
// ParseHeader() (DocType string, err error)
// ParseTracks() ([]TrackInfo, error)
// ExtractSubtitleBlocks(trackNumber uint64) ([]SubtitleBlock, error)
```

**Critical implementation details:**

1. **Timestamp calculation:** Block timestamps are relative to their Cluster timestamp. Absolute time = `(ClusterTimestamp + BlockTimestamp) * TimestampScale / 1_000_000` to get milliseconds.

2. **Block binary structure:** A Block/SimpleBlock starts with:
   - Track number (VINT, but NOT using EBML VINT -- it uses Matroska's "lacing" VINT which strips the width marker)
   - Relative timestamp (signed 16-bit big-endian, relative to cluster)
   - Flags byte (lacing type, keyframe, etc.)
   - Payload data

3. **Seeking strategy:** The decoder should:
   - First pass: Read EBML header, then scan Segment children for SeekHead + Tracks.
   - Use SeekHead if present to jump directly to Tracks/Clusters/Cues.
   - If no SeekHead, scan sequentially through Segment children.
   - When extracting: iterate Clusters, within each Cluster check SimpleBlock/BlockGroup, filter by track number.

4. **Duration for subtitles:** For SimpleBlock, duration is NOT included (it's inferred or uses the track's DefaultDuration). For BlockGroup, duration is in the BlockDuration child element. Subtitle tracks almost always use BlockGroup with explicit BlockDuration because every subtitle event has a specific display duration.

### 3. Subtitle Codec Handlers (`pkg/subtitle`)

Each codec handler parses raw block data + CodecPrivate into a common subtitle event model.

```go
// Event represents a single subtitle display event.
type Event struct {
    Start    time.Duration
    End      time.Duration
    Layer    int       // Z-order for overlapping subtitles
    Style    string    // Style name reference (ASS)
    Name     string    // Character name (ASS)
    MarginL  int       // Override margins (ASS)
    MarginR  int
    MarginV  int
    Effect   string    // Transition effect (ASS)
    Text     string    // Subtitle text with inline formatting
}

// Track holds all parsed subtitle events and metadata.
type Track struct {
    Header   string    // ASS header block (styles, script info) if applicable
    Events   []Event
    CodecID  string
    Language string
    Name     string
}

// Handler is the interface each codec implements.
type Handler interface {
    // ParseHeader processes CodecPrivate data (may be nil for SRT).
    ParseHeader(codecPrivate []byte) error
    // ParseBlock processes a single subtitle block into events.
    ParseBlock(block matroska.SubtitleBlock) ([]Event, error)
    // GetHeader returns the ASS-format header (styles, metadata).
    GetHeader() string
}
```

**Codec-specific behavior:**

| CodecID | CodecPrivate | Block Data Format | Conversion Notes |
|---------|-------------|-------------------|------------------|
| `S_TEXT/UTF8` | None | Plain text (SRT text without sequence numbers or timestamps) | Must wrap in ASS dialogue lines, assign default style |
| `S_TEXT/SSA` | SSA header (everything before `[Events]`) | SSA dialogue fields: `Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text` but **without** `Dialogue:` prefix and without `Start`/`End` (timestamps come from Block timing) | Parse comma-separated fields, map to ASS (SSA is a subset of ASS) |
| `S_TEXT/ASS` | ASS header (everything before `[Events]`) | Same as SSA but with ASS extensions | Mostly passthrough -- already ASS format |

**Critical nuance for SSA/ASS block data:** In Matroska, the block data for SSA/ASS tracks contains the dialogue line fields in order: `ReadOrder, Layer, Style, Name, MarginL, MarginR, MarginV, Effect, Text`. The `ReadOrder` is a Matroska-specific addition. Start and End times are NOT in the block data; they come from the Block's timestamp and BlockDuration. This is a common source of bugs.

### 4. Format Writer (`pkg/format/ass`)

Generates valid ASS files.

```go
// Writer outputs a complete ASS file.
type Writer struct {
    w io.Writer
}

// Key methods:
// WriteFile(track subtitle.Track) error
// writeScriptInfo(header string) error
// writeStyles(header string) error
// writeEvents(events []subtitle.Event) error
// formatTimestamp(d time.Duration) string  // -> "H:MM:SS.CC"
```

**ASS file structure:**
```
[Script Info]
; Script generated by mkv-sub-extractor
ScriptType: v4.00+
PlayResX: 384
PlayResY: 288
...

[V4+ Styles]
Format: Name, Fontname, Fontsize, PrimaryColour, ...
Style: Default,Arial,20,&H00FFFFFF,...

[Events]
Format: Layer, Start, End, Style, Name, MarginL, MarginR, MarginV, Effect, Text
Dialogue: 0,0:00:01.00,0:00:04.00,Default,,0,0,0,,Hello world
```

**When the source is SSA/ASS:** The CodecPrivate already contains `[Script Info]` and `[V4+ Styles]` (or `[V4 Styles]` for SSA). Write it verbatim, then generate the `[Events]` section from the extracted blocks.

**When the source is SRT (S_TEXT/UTF8):** Generate a default ASS header with sensible defaults, then convert each plain-text subtitle into an ASS Dialogue line with the Default style. SRT formatting tags (`<b>`, `<i>`, `<u>`, `<font color="...">`) should be converted to ASS override tags (`{\b1}`, `{\i1}`, `{\u1}`, `{\c&H...&}`).

**ASS timestamp format:** `H:MM:SS.CC` (hours:minutes:seconds.centiseconds). Note: single digit hour, two-digit centiseconds. This differs from SRT's `HH:MM:SS,mmm`.

### 5. CLI Layer (`cmd/mkv-sub-extractor`)

```go
// main.go - Entry point
// Responsibilities:
// 1. Parse flags: --output dir, positional MKV path
// 2. If no path: scan CWD for *.mkv, present numbered list
// 3. Open file, call extractor to list tracks
// 4. Present subtitle tracks as numbered list with metadata
// 5. User selects track number
// 6. Call extractor to extract + convert
// 7. Report output file path

// No framework needed. Use standard library flag package.
// Interactive selection: fmt.Print menu, bufio.Scanner for input.
```

### 6. Orchestrator (`pkg/extractor`)

```go
// Extractor coordinates the full extraction pipeline.
type Extractor struct {
    decoder *matroska.Decoder
}

// ListSubtitleTracks returns metadata for all subtitle tracks in the file.
func (e *Extractor) ListSubtitleTracks(r io.ReadSeeker) ([]matroska.TrackInfo, error)

// ExtractTrack extracts a subtitle track and writes ASS to the given writer.
func (e *Extractor) ExtractTrack(r io.ReadSeeker, trackNumber uint64, w io.Writer) error
```

The orchestrator's `ExtractTrack` flow:
1. Parse EBML header, validate DocType is "matroska" or "webm"
2. Parse Tracks element to get TrackInfo for the selected track
3. Instantiate the appropriate subtitle Handler based on CodecID
4. Feed CodecPrivate to handler's ParseHeader
5. Iterate all Clusters, extract Blocks for the target track number
6. Feed each block to handler's ParseBlock, collect Events
7. Sort events by Start time (blocks may not be in order)
8. Pass collected Track to ASS Writer

### Data Flow

```
MKV File (io.ReadSeeker)
    |
    v
[EBML Parser] -- reads raw bytes, yields Element headers
    |
    v
[Matroska Decoder] -- interprets elements, yields TrackInfo + SubtitleBlocks
    |
    +-- TrackInfo[] --> [CLI Layer] -- user sees track list, selects one
    |
    +-- SubtitleBlock[] (for selected track) --> [Subtitle Codec Handler]
    |                                                |
    |                                                v
    |                                          subtitle.Track (Events + Header)
    |                                                |
    |                                                v
    |                                          [ASS Format Writer]
    |                                                |
    v                                                v
(skips video/audio clusters)                   Output .ass File
```

**Key data flow decisions:**

1. **Streaming, not buffering:** Video and audio blocks (which are the vast majority of data in an MKV file) are never read into memory. The EBML parser skips them using `io.Seeker`.

2. **Two-pass approach:** First pass reads tracks metadata (small, near start of file). Second pass iterates clusters for subtitle blocks (requires traversing the full file but only reading blocks matching the target track number).

3. **Track filtering happens at the Matroska Decoder level:** The decoder checks each Block's track number and only returns blocks matching the requested track. This keeps memory usage minimal.

## Patterns to Follow

### Pattern 1: io.ReadSeeker for File Access
**What:** All parsing operates on `io.ReadSeeker`, not `*os.File` or `[]byte`.
**When:** Always, for all file access in the parsing layers.
**Why:** Enables testing with `bytes.Reader`, enables future support for HTTP range requests, keeps components decoupled from filesystem.
```go
func NewDecoder(r io.ReadSeeker) *Decoder {
    return &Decoder{reader: ebml.NewReader(r)}
}
```

### Pattern 2: Element-at-a-time Scanning
**What:** Parse one EBML element header, decide to descend (master element) or skip/read (data element), never buffer the full element tree.
**When:** Parsing any part of the MKV structure.
**Why:** MKV files can be multi-gigabyte. Building a full element tree would OOM.
```go
for {
    elem, err := reader.ReadElementHeader()
    if err == io.EOF { break }
    switch elem.ID {
    case matroska.IDTracks:
        // Descend into master element (read children)
        tracks, err := parseTracks(reader, elem.DataSize)
    case matroska.IDCluster:
        // Process or skip depending on whether we're extracting
        if extracting {
            processCluster(reader, elem)
        } else {
            reader.Skip(elem.DataSize)
        }
    default:
        reader.Skip(elem.DataSize)
    }
}
```

### Pattern 3: Codec Handler Interface
**What:** Each subtitle codec implements a common `Handler` interface, selected by CodecID string.
**When:** Processing subtitle data from blocks.
**Why:** Clean separation of format-specific logic. Adding a new codec (e.g., S_TEXT/WEBVTT) means adding one file, not modifying existing code.
```go
func NewHandler(codecID string) (Handler, error) {
    switch codecID {
    case "S_TEXT/UTF8":
        return &SRTHandler{}, nil
    case "S_TEXT/SSA":
        return &SSAHandler{}, nil
    case "S_TEXT/ASS":
        return &ASSHandler{}, nil
    default:
        return nil, fmt.Errorf("unsupported subtitle codec: %s", codecID)
    }
}
```

### Pattern 4: Table-Driven Element ID Constants
**What:** Define all known Matroska element IDs as typed constants in a single file.
**When:** Defining the Matroska schema.
**Why:** Single source of truth, easy to audit against the spec.
```go
package matroska

import "mkv-sub-extractor/pkg/ebml"

const (
    // Level 0
    IDEBMLHeader   ebml.ElementID = 0x1A45DFA3
    IDSegment      ebml.ElementID = 0x18538067

    // Segment children
    IDSeekHead     ebml.ElementID = 0x114D9B74
    IDInfo         ebml.ElementID = 0x1549A966
    IDTracks       ebml.ElementID = 0x1654AE6B
    IDCluster      ebml.ElementID = 0x1F43B675
    IDCues         ebml.ElementID = 0x1C53BB6B

    // TrackEntry children
    IDTrackEntry   ebml.ElementID = 0xAE
    IDTrackNumber  ebml.ElementID = 0xD7
    IDTrackType    ebml.ElementID = 0x83
    IDCodecID      ebml.ElementID = 0x86
    IDCodecPrivate ebml.ElementID = 0x63A2
    IDLanguage     ebml.ElementID = 0x22B59C
    IDTrackName    ebml.ElementID = 0x536E

    // Cluster children
    IDTimestamp    ebml.ElementID = 0xE7
    IDSimpleBlock  ebml.ElementID = 0xA3
    IDBlockGroup   ebml.ElementID = 0xA0
    IDBlock        ebml.ElementID = 0xA1
    IDBlockDuration ebml.ElementID = 0x9B

    // Info children
    IDTimestampScale ebml.ElementID = 0x2AD7B1
    IDDuration       ebml.ElementID = 0x4489
)
```

### Pattern 5: Errors as Meaningful Types
**What:** Define domain-specific error types for common failure modes.
**When:** Errors that callers need to handle differently.
**Why:** Enables the CLI to give meaningful user-facing error messages.
```go
var (
    ErrNotMatroska      = errors.New("file is not a Matroska container")
    ErrNoSubtitleTracks = errors.New("no subtitle tracks found")
    ErrUnsupportedCodec = errors.New("unsupported subtitle codec")
    ErrCorruptElement   = errors.New("corrupt EBML element")
    ErrTruncatedFile    = errors.New("file appears truncated")
)
```

## Anti-Patterns to Avoid

### Anti-Pattern 1: Loading Entire File Into Memory
**What:** Reading the full MKV file into a `[]byte` before parsing.
**Why bad:** MKV files are commonly 1-20 GB. Immediate OOM on any real content.
**Instead:** Use `io.ReadSeeker` and process elements one at a time, seeking past video/audio data.

### Anti-Pattern 2: Building a Full Element Tree
**What:** Parsing the entire EBML structure into a tree of nodes in memory, then querying it.
**Why bad:** Wastes memory and time. We only need tracks metadata and subtitle blocks.
**Instead:** Scan elements sequentially, descend into elements we care about, skip everything else.

### Anti-Pattern 3: Shared Mutable State Between Components
**What:** Having the Matroska decoder modify subtitle handler state directly, or passing pointers to shared structs across layers.
**Why bad:** Makes components untestable in isolation, creates hidden coupling.
**Instead:** Each component receives input, produces output. The orchestrator passes data between them.

### Anti-Pattern 4: Parsing Track and Cluster in One Pass
**What:** Trying to discover tracks and extract subtitle blocks simultaneously in a single sequential scan.
**Why bad:** You need to know the CodecID/CodecPrivate before you can interpret blocks. Tracks typically appears before Clusters, but the spec does not guarantee ordering of Segment children. SeekHead exists precisely because ordering is not guaranteed.
**Instead:** Two-pass: First pass locates and parses Tracks. Second pass iterates Clusters with a known target track number.

### Anti-Pattern 5: Hardcoding ASS Field Positions
**What:** Assuming SSA/ASS block data always has exactly 9 fields in a fixed order.
**Why bad:** The `Format:` line in the `[Events]` section defines the field order. Some files may have custom field orderings. The block data in Matroska, however, always uses the standard Matroska order (ReadOrder, Layer, Style, Name, MarginL, MarginR, MarginV, Effect, Text). But the Text field may contain commas, so you must split on comma only up to the 9th field, with the remainder being the Text.
**Instead:** Split into at most N fields (where N is the expected count), treating the last field as everything remaining.

## Suggested Build Order

Components should be built bottom-up, following data dependencies:

```
Phase 1: EBML Parser (pkg/ebml)
    |-- No dependencies on other project code
    |-- Can be fully tested with hand-crafted binary fixtures
    |-- Foundation everything else depends on
    |
Phase 2: Matroska Decoder (pkg/matroska)
    |-- Depends on: EBML Parser
    |-- Can be tested with small .mkv fixtures
    |-- Delivers: track listing capability
    |
Phase 3: Subtitle Codec Handlers (pkg/subtitle)
    |-- Depends on: Matroska types (SubtitleBlock, TrackInfo)
    |-- Can be tested with raw block data, no MKV parsing needed
    |-- Delivers: subtitle event parsing
    |
Phase 4: ASS Format Writer (pkg/format/ass)
    |-- Depends on: Subtitle types (Event, Track)
    |-- Can be tested independently with constructed events
    |-- Delivers: file output capability
    |
Phase 5: Orchestrator + CLI (pkg/extractor, cmd/)
    |-- Depends on: All above components
    |-- Integration point
    |-- Delivers: working tool
```

**Rationale for this order:**
- Each phase produces a testable, independently valuable component.
- Phase 1 is prerequisite for everything. Cannot parse MKV without EBML.
- Phase 2 unlocks track discovery, which can be demoed standalone.
- Phases 3 and 4 can actually be developed in parallel since they have no mutual dependency.
- Phase 5 is pure glue code once the components exist.

## Scalability Considerations

| Concern | Typical MKV (2-4 GB) | Large MKV (20+ GB) | Batch (future) |
|---------|----------------------|---------------------|-----------------|
| Memory | Minimal -- only subtitle blocks buffered (KB-MB range) | Same -- seeking skips video/audio data | Process files sequentially, no extra memory |
| Parse time | Subtitle extraction: < 2 seconds (mostly seeking past video clusters) | May take 5-10 seconds due to cluster iteration | Linear scaling per file |
| File handles | 1 input + 1 output | Same | Close between files |
| Cluster traversal | ~500-2000 clusters, subtitle blocks sparse | ~5000-20000 clusters | N/A |

**Key insight:** Subtitle data is tiny compared to the file size. A 2-hour movie with dense subtitles might have 2000 subtitle events totaling ~200 KB of text, inside a 4 GB file. The architecture must handle seeking through gigabytes of video data efficiently to find those few kilobytes of subtitle data.

## Package Layout

```
mkv-sub-extractor/
  cmd/
    mkv-sub-extractor/
      main.go              -- Entry point, flag parsing, interactive UI
  pkg/
    ebml/
      reader.go            -- VINT decoding, element header parsing
      reader_test.go
      types.go             -- ElementID type, Element struct
    matroska/
      ids.go               -- Element ID constants
      decoder.go           -- MKV structure navigation, track parsing
      decoder_test.go
      block.go             -- Block/SimpleBlock binary parsing
      block_test.go
      types.go             -- TrackInfo, SubtitleBlock, TrackType
    subtitle/
      handler.go           -- Handler interface, factory function
      srt.go               -- S_TEXT/UTF8 handler
      srt_test.go
      ssa.go               -- S_TEXT/SSA handler (shared with ASS)
      ssa_test.go
      ass.go               -- S_TEXT/ASS handler
      ass_test.go
      types.go             -- Event, Track structs
    format/
      ass/
        writer.go          -- ASS file generation
        writer_test.go
        defaults.go        -- Default ASS header template
    extractor/
      extractor.go         -- Orchestration logic
      extractor_test.go
  go.mod
  go.sum                   -- Will be empty (zero dependencies)
```

## Sources

- EBML specification: RFC 8794 (https://www.rfc-editor.org/rfc/rfc8794) -- VINT encoding, element structure, data types
- Matroska specification: IETF draft / matroska.org technical docs (https://www.matroska.org/technical/elements.html) -- element hierarchy, element IDs, semantics
- Matroska subtitle specification: (https://www.matroska.org/technical/subtitles.html) -- codec IDs, block data format for subtitle tracks
- Matroska Block structure: (https://www.matroska.org/technical/basics.html) -- Block header parsing, lacing, timestamp encoding
- ASS format reference: (http://www.tcax.org/docs/ass-specs.htm) -- ASS file structure, timestamp format, dialogue line format

**Confidence notes:**
- Element IDs: MEDIUM -- These are from training data and widely documented, but should be verified against the current Matroska spec during implementation. The core IDs (Segment, Tracks, Cluster, etc.) are stable and have not changed.
- VINT encoding: HIGH -- RFC 8794 is a stable IETF standard. The encoding scheme is well-defined.
- SSA/ASS block data format in Matroska: MEDIUM -- The "ReadOrder, Layer, Style, ..." field order for SSA/ASS blocks in Matroska is documented but frequently misunderstood. Verify against actual MKV files during development.
- ASS timestamp format: HIGH -- `H:MM:SS.CC` is a long-standing format detail.
