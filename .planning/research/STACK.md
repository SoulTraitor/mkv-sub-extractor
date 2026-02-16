# Technology Stack

**Project:** MKV Subtitle Extractor
**Researched:** 2026-02-16
**Overall Confidence:** MEDIUM-HIGH

---

## Recommended Stack

### Language & Runtime

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go | 1.24+ | Language runtime | Project requirement. Go 1.24 is the minimum needed by matroska-go. Go 1.26 is the current latest stable release. Target 1.24 in go.mod for broader compatibility. |

### MKV/EBML Parsing

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `github.com/luispater/matroska-go` | v1.2.4 | MKV container demuxing, track enumeration, packet reading | **Primary recommendation.** High-level Demuxer API with `GetTrackInfo()`, `ReadPacket()`, and direct subtitle track identification (`TypeSubtitle = 17`). Provides `CodecID`, `CodecPrivate`, `Name`, `Language` fields on TrackInfo. Most recently maintained (Aug 2025), actively versioned (v1.2.4), includes a working subtitle extraction example. Requires `io.ReadSeeker` which is fine for file-based use. |

### Subtitle Format Handling

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `github.com/asticode/go-astisub` | v0.38.0 | SRT/SSA/ASS parsing and writing, format conversion | **The** subtitle library in the Go ecosystem. 38 releases, MIT licensed, maintained through Nov 2025. Full ASS/SSA style preservation: `StyleAttributes` with 25+ SSA-specific fields (font, colors, alignment, margins, outline, shadow). Preserves `Metadata` including `SSAPlayResX/Y`, `SSAScriptType`, `SSAWrapStyle`. Has `ReadFromSRT()` and `WriteToSSA()` which is exactly the SRT-to-ASS conversion pipeline we need. Handles edge cases via `SSAOptions` callbacks for unknown sections and invalid lines. |

### CLI Interaction

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `github.com/charmbracelet/huh` | v0.8.0 | Interactive file and track selection prompts | Modern, well-maintained (Oct 2024, 980 importers). Generic `Select[T]` type means we can use `Select[int]` for track indices directly. Built on Bubble Tea for robust terminal handling. Supports accessible mode. Provides `Select`, `Confirm`, and `FilePicker` fields -- all three are needed for this project. Alternative was `promptui` (3470 importers, more battle-tested) but it has not been updated since Oct 2021 and `huh` has a cleaner API with generics support. |

### CLI Argument Parsing

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| Go `flag` (stdlib) | N/A | Command-line argument parsing | This is a simple single-command CLI with a few flags (`--output`, `--file`, `--track`). Cobra (v1.10.2, 184K importers) is massively overkill -- it is designed for subcommand-based CLIs like `git` or `kubectl`. The stdlib `flag` package handles positional args and flags cleanly. Zero dependency cost. If requirements grow to subcommands later, Cobra can be added, but YAGNI applies here. |

### Terminal Output

| Technology | Version | Purpose | Why |
|------------|---------|---------|-----|
| `github.com/charmbracelet/lipgloss` | v1.1.0 | Styled terminal output (track listings, progress, errors) | Already a transitive dependency via `huh`. Provides clean table-like formatting for track listings without pulling in a separate table library. Stable at v1.1.0 (Mar 2025). |

---

## Stack Decision: MKV Parser Deep Dive

This was the hardest decision in the stack. Here is the full analysis.

### Option A: `luispater/matroska-go` v1.2.4 -- RECOMMENDED

**API style:** Pull-based demuxer. Call `GetNumTracks()`, `GetTrackInfo(i)`, `ReadPacket()`.

```go
demuxer, err := matroska.NewDemuxer(file)
numTracks, _ := demuxer.GetNumTracks()
for i := uint(0); i < numTracks; i++ {
    info, _ := demuxer.GetTrackInfo(i)
    if info.Type == matroska.TypeSubtitle {
        fmt.Printf("Track %d: %s (%s)\n", info.Number, info.Name, info.CodecID)
    }
}
// Then ReadPacket() in a loop to get subtitle data
```

**Strengths:**
- High-level API -- enumerate tracks, read packets, done
- `TrackInfo` has `Name`, `Language`, `CodecID`, `CodecPrivate` -- everything needed
- `Packet` has `StartTime`, `EndTime` in nanoseconds, `Data []byte` -- clean
- Includes `GetAttachments()` for font extraction (future feature)
- Tagged v1.2.4 with active maintenance (Aug 2025)
- Includes working subtitle extraction example code

**Weaknesses:**
- Requires Go 1.24+ (reasonable given current Go 1.26)
- 0 known importers on pkg.go.dev (relatively new)
- `io.ReadSeeker` requirement means no stdin streaming (not a problem for files)

**Confidence:** MEDIUM -- fewer community users, but API is well-designed and the library is actively maintained.

### Option B: `remko/go-mkvparse` v0.14.0

**API style:** Push-based event handler. Implement the `Handler` interface (7 methods).

```go
type SubHandler struct {
    mkvparse.DefaultHandler
}
func (h *SubHandler) HandleString(id mkvparse.ElementID, val string, info mkvparse.ElementInfo) error {
    // manually check element IDs, build state machine
}
func (h *SubHandler) HandleBinary(id mkvparse.ElementID, data []byte, info mkvparse.ElementInfo) error {
    // handle codec private, subtitle data blocks
}
```

**Strengths:**
- Most established (5 importers, v0.14.0, MIT)
- Zero dependencies
- `ParseSections()` for selective parsing (fast)
- Well-documented with examples

**Weaknesses:**
- Push/event-based API is significantly more complex for our use case
- Must manually build track state from individual element callbacks
- Must manually correlate Cluster/Block data with track numbers
- Must maintain state machine for nested EBML elements
- Last updated Jan 2022 (4 years ago)
- Subtitle extraction requires ~3x more code than matroska-go

**Confidence:** HIGH for stability, but LOW for developer experience on this use case.

### Option C: `coding-socks/matroska`

**API style:** Scanner-based (iterator pattern).

**Strengths:**
- Has `ExtractTract()` function for direct track extraction
- Clean Scanner API with `Init()`, `Next()`, `Cluster()`
- Subtitle codec constants defined (`SubtitleCodecTEXTASS`, `SubtitleCodecTEXTUTF8`, etc.)

**Weaknesses:**
- Explicitly alpha status: "Public API can change between days"
- No tagged versions (pseudo-version only)
- Function name typo (`ExtractTract` not `ExtractTrack`) suggests rough edges
- Last update Apr 2025

**Verdict:** Too immature for production use. Pass.

### Option D: `5rahim/gomkv` v0.2.1

Fork of `remko/go-mkvparse` with identical API. Updated Jun 2025 but no meaningful divergence from upstream documented. Same push-based complexity applies. No advantage over using the original.

### Decision Rationale

`matroska-go` wins because the pull-based Demuxer API directly maps to our workflow:
1. Open file -> `NewDemuxer(file)`
2. List subtitle tracks -> `GetNumTracks()` + `GetTrackInfo(i)` with `Type == 17`
3. Read subtitle data -> `ReadPacket()` loop filtering by track number
4. Get codec private data -> `TrackInfo.CodecPrivate` (needed for ASS header)

With `go-mkvparse`, steps 2-4 would each require building custom state machines from individual EBML element callbacks. The complexity difference is substantial.

---

## Stack Decision: Subtitle Library

### `asticode/go-astisub` v0.38.0 -- RECOMMENDED (only real option)

The Go ecosystem has no other subtitle library worth considering. Everything else on pkg.go.dev is either a CLI tool (not a library) or abandoned. `go-astisub` is:

- **The only maintained Go library** for multi-format subtitle reading/writing
- **38 releases** over active development
- **Full ASS style preservation** including all SSA fields
- **Direct conversion pipeline:** `ReadFromSRT(reader)` -> modify -> `WriteToSSA(writer)`

For SRT-to-ASS conversion specifically:
```go
// Read SRT data extracted from MKV
subs, err := astisub.ReadFromSRT(srtReader)

// Write as ASS
err = subs.WriteToSSA(outputFile)
```

For ASS tracks, the raw data from MKV can be written directly (codec private = ASS header, block data = dialogue lines) or parsed through astisub for validation/normalization.

**Confidence:** HIGH -- well-established, actively maintained, comprehensive API.

### What about writing a custom ASS/SRT parser?

Not recommended. ASS format has significant complexity:
- `[Script Info]` section with 15+ metadata fields
- `[V4+ Styles]` section with 23 style attributes per line
- `[Events]` section with inline override tags (`{\pos(x,y)}`, `{\fad(in,out)}`, etc.)
- Multiple SSA versions (v4, v4+) with different field counts

SRT is simpler but still has edge cases (HTML tags, ASS-style formatting, BOM handling). `go-astisub` handles all of these.

---

## Supporting Libraries

| Library | Version | Purpose | When to Use |
|---------|---------|---------|-------------|
| `io` (stdlib) | N/A | `io.ReadSeeker`, `io.Reader`, `io.Writer` interfaces | Always -- core I/O abstraction |
| `os` (stdlib) | N/A | File operations, directory scanning | File listing, output file creation |
| `path/filepath` (stdlib) | N/A | Cross-platform path manipulation | Output filename generation, directory walking |
| `fmt` (stdlib) | N/A | Formatted output | Track info display, error messages |
| `strings` (stdlib) | N/A | String manipulation | Codec ID parsing, subtitle text processing |
| `time` (stdlib) | N/A | Duration handling | Subtitle timestamp conversion (nanoseconds to duration) |
| `bytes` (stdlib) | N/A | Buffer operations | Building subtitle data from packets |
| `encoding/binary` (stdlib) | N/A | Binary data parsing | Only if manual EBML parsing is needed for edge cases |

---

## Alternatives Considered

| Category | Recommended | Alternative | Why Not |
|----------|-------------|-------------|---------|
| MKV parsing | `luispater/matroska-go` | `remko/go-mkvparse` | Push-based API adds ~3x code complexity for subtitle extraction; last updated 2022 |
| MKV parsing | `luispater/matroska-go` | `coding-socks/matroska` | Alpha status, no tagged versions, API instability warning |
| MKV parsing | `luispater/matroska-go` | Custom EBML parser | EBML spec is large (200+ element types); reinventing this is months of work with no upside |
| Subtitle format | `asticode/go-astisub` | Custom SRT/ASS parser | ASS format is deceptively complex (styles, override tags, version differences); no benefit |
| CLI interaction | `charmbracelet/huh` | `manifoldco/promptui` | Unmaintained since Oct 2021; no generics; dated API |
| CLI interaction | `charmbracelet/huh` | Raw `fmt.Scan` / `bufio.Scanner` | Poor UX -- no arrow-key selection, no search, no visual feedback |
| CLI framework | stdlib `flag` | `spf13/cobra` | Massive overkill for single-command CLI; adds 8+ transitive dependencies |
| CLI framework | stdlib `flag` | `urfave/cli` | Same reasoning -- subcommand frameworks for a tool with no subcommands |
| Terminal styling | `charmbracelet/lipgloss` | `fatih/color` | lipgloss is already a transitive dep via huh; more capable for layout |

---

## What NOT to Use

| Technology | Why Not |
|------------|---------|
| FFmpeg/mkvextract via `os/exec` | Violates core project constraint (pure Go, zero external deps) |
| `spf13/cobra` | Single-command CLI does not need subcommand framework |
| `promptui` | Abandoned since 2021; use `huh` instead |
| `coding-socks/matroska` | Alpha, unstable API, typos in public functions |
| Custom EBML parser | Months of work, EBML spec has 200+ element types, well-tested libs exist |
| `encoding/xml` for TTML | Not needed -- project scope is SRT/ASS only |
| CGo bindings to libmatroska | Defeats "pure Go" constraint, complicates cross-compilation |

---

## Installation

```bash
# Initialize module
go mod init mkv-sub-extractor

# Core dependencies
go get github.com/luispater/matroska-go@v1.2.4
go get github.com/asticode/go-astisub@v0.38.0

# CLI interaction
go get github.com/charmbracelet/huh@v0.8.0

# lipgloss comes as transitive dep via huh, but pin explicitly if used directly
go get github.com/charmbracelet/lipgloss@v1.1.0
```

**Total direct dependencies: 3** (matroska-go, go-astisub, huh)
**Stdlib-only alternative: 1** (flag -- no external dep needed for arg parsing)

---

## Fallback Plan

If `luispater/matroska-go` proves problematic at implementation time (e.g., cannot handle certain MKV files, codec private data is incomplete), the fallback is:

1. **First fallback:** Use `remko/go-mkvparse` with a custom state-machine handler. More code but battle-tested.
2. **Second fallback:** Use `coding-socks/matroska` with its `ExtractTract()` function, accepting API instability risk.
3. **Nuclear option:** Write a minimal EBML parser that only handles the ~20 element types needed for subtitle extraction (not the full 200+ spec). This is feasible but adds 1-2 weeks of work.

---

## Confidence Assessment

| Component | Confidence | Rationale |
|-----------|------------|-----------|
| Go language choice | HIGH | Project requirement, excellent fit for CLI tools |
| `matroska-go` for MKV parsing | MEDIUM | Clean API, active maintenance, but low community adoption (0 importers). Verify with real MKV files early. |
| `go-astisub` for subtitles | HIGH | 38 releases, comprehensive ASS support, only real option in Go ecosystem |
| `huh` for CLI interaction | HIGH | 980 importers, active Charm ecosystem, well-documented |
| stdlib `flag` for args | HIGH | Proven, zero-dep, appropriate for single-command CLI |
| `lipgloss` for output styling | HIGH | Stable v1, transitive dep anyway, 7600+ importers |

---

## Sources

- pkg.go.dev search: `ebml matroska` -- 6 results surveyed
- pkg.go.dev search: `mkv parser` -- 5 results surveyed
- pkg.go.dev search: `subtitle srt ass` -- 6 results surveyed (all CLI tools, no libraries except go-astisub found separately)
- pkg.go.dev/github.com/remko/go-mkvparse -- full API documentation reviewed
- pkg.go.dev/github.com/luispater/matroska-go@v1.2.4 -- full API documentation reviewed
- pkg.go.dev/github.com/coding-socks/matroska -- full API documentation reviewed
- pkg.go.dev/github.com/5rahim/gomkv -- fork analysis
- pkg.go.dev/github.com/asticode/go-astisub@v0.38.0 -- full API with SSA types reviewed
- pkg.go.dev/github.com/charmbracelet/huh -- v0.8.0 API and examples reviewed
- pkg.go.dev/github.com/manifoldco/promptui -- v0.9.0 compared (last update Oct 2021)
- pkg.go.dev/github.com/spf13/cobra -- v1.10.2 evaluated and rejected
- pkg.go.dev/github.com/charmbracelet/lipgloss -- v1.1.0 overview
- go.dev/dl -- Go 1.26 confirmed as current stable
