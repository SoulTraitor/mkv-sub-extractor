---
phase: 02-subtitle-extraction-and-ass-output
verified: 2026-02-16T15:30:00Z
status: human_needed
score: 4/5 must-haves verified
must_haves:
  truths:
    - "Extracting an ASS/SSA track produces a valid ASS file that preserves all original styles from CodecPrivate (passthrough mode)"
    - "Extracting an SRT track produces a valid ASS file with properly generated Script Info and V4+ Styles headers"
    - "Subtitle timestamps in the output file are accurate (Cluster Timestamp + Block offset + TimestampScale correctly computed, millisecond-to-centisecond conversion preserves precision)"
    - "SRT formatting tags (italic, bold) are correctly mapped to ASS override tags in the converted output"
    - "Output ASS files render correctly when loaded in a media player (mpv/VLC)"
  artifacts:
    - path: "pkg/subtitle/types.go"
      provides: "SubtitleEvent struct shared by all extraction paths"
    - path: "pkg/assout/timestamp.go"
      provides: "FormatASSTimestamp function for nanosecond-to-centisecond conversion"
    - path: "pkg/subtitle/ass.go"
      provides: "ParseASSBlockData function for Matroska ASS/SSA block data parsing"
    - path: "pkg/subtitle/srt.go"
      provides: "ConvertSRTTagsToASS function for HTML-to-ASS tag mapping"
    - path: "pkg/extract/extract.go"
      provides: "ExtractSubtitlePackets function with gap-fill strategy"
    - path: "pkg/assout/writer.go"
      provides: "WriteASSPassthrough function for ASS/SSA tracks"
    - path: "pkg/assout/ssa_convert.go"
      provides: "ConvertSSAHeaderToASS for SSA V4 to ASS V4+ conversion"
    - path: "pkg/assout/template.go"
      provides: "Default ASS header template for SRT-to-ASS conversion"
    - path: "pkg/assout/srt_writer.go"
      provides: "WriteSRTAsASS for SRT-to-ASS conversion"
    - path: "pkg/output/naming.go"
      provides: "GenerateOutputPath for output file naming with collision handling"
    - path: "pkg/extract/pipeline.go"
      provides: "ExtractTrackToASS public API orchestrating full pipeline"
  key_links:
    - from: "pkg/extract/pipeline.go"
      to: "pkg/extract/extract.go"
      via: "Calls ExtractSubtitlePackets to get raw packets from MKV"
    - from: "pkg/extract/pipeline.go"
      to: "pkg/assout/writer.go"
      via: "Calls WriteASSPassthrough for ASS/SSA codec tracks"
    - from: "pkg/extract/pipeline.go"
      to: "pkg/assout/srt_writer.go"
      via: "Calls WriteSRTAsASS for SRT codec tracks"
    - from: "pkg/extract/pipeline.go"
      to: "pkg/subtitle/ass.go"
      via: "Calls ParseASSBlockData to convert raw packets to SubtitleEvents"
    - from: "pkg/extract/pipeline.go"
      to: "pkg/output/naming.go"
      via: "Calls GenerateOutputPath for output file naming"
    - from: "pkg/assout/writer.go"
      to: "pkg/assout/timestamp.go"
      via: "Uses FormatASSTimestamp for Dialogue line timestamps"
    - from: "pkg/assout/writer.go"
      to: "pkg/assout/ssa_convert.go"
      via: "Calls ConvertSSAHeaderToASS when codec is S_TEXT/SSA"
    - from: "pkg/assout/srt_writer.go"
      to: "pkg/assout/template.go"
      via: "WriteSRTAsASS writes defaultASSHeader before events"
    - from: "pkg/assout/srt_writer.go"
      to: "pkg/subtitle/srt.go"
      via: "Uses ConvertSRTTagsToASS for tag mapping"
    - from: "pkg/assout/srt_writer.go"
      to: "pkg/assout/timestamp.go"
      via: "Uses FormatASSTimestamp for Dialogue line timestamps"
human_verification:
  - test: "Run tool against an MKV file with ASS subtitles; open output .ass file in mpv/VLC"
    expected: "Subtitles render with original styles (fonts, colors, positioning) at correct times"
    why_human: "Visual rendering correctness and timing accuracy against video cannot be verified programmatically"
  - test: "Run tool against an MKV file with SRT subtitles; open output .ass file in mpv/VLC"
    expected: "Subtitles render with Microsoft YaHei font, white text with black outline, bottom-center, at correct times"
    why_human: "Visual appearance of generated default style and timing sync requires human observation"
  - test: "Compare subtitle timing accuracy -- check if subtitles appear/disappear within ~100ms of expected moments"
    expected: "Timestamps are accurate; no visible drift or offset across the duration of the video"
    why_human: "Timing precision evaluation against video playback requires real-time human observation"
---

# Phase 2: Subtitle Extraction and ASS Output Verification Report

**Phase Goal:** User can select a subtitle track and get a valid, correctly-timed ASS file as output
**Verified:** 2026-02-16T15:30:00Z
**Status:** human_needed
**Re-verification:** No -- initial verification

## Goal Achievement

### Observable Truths

| # | Truth | Status | Evidence |
|---|-------|--------|----------|
| 1 | Extracting an ASS/SSA track produces a valid ASS file that preserves all original styles from CodecPrivate (passthrough mode) | VERIFIED | `WriteASSPassthrough` in `pkg/assout/writer.go` writes CodecPrivate bytes verbatim for ASS tracks (line 28-33); SSA tracks get `ConvertSSAHeaderToASS` conversion before write (line 31-33); 12 writer tests pass including `TestWriteASSPassthrough_ASSCodecPreservesVerbatim` and `TestWriteASSPassthrough_SSACodecTriggersConversion`; pipeline dispatches correctly via codec ID switch in `pipeline.go` (lines 127-134) |
| 2 | Extracting an SRT track produces a valid ASS file with properly generated Script Info and V4+ Styles headers | VERIFIED | `WriteSRTAsASS` in `pkg/assout/srt_writer.go` writes `defaultASSHeader` containing `[Script Info]`, `ScriptType: v4.00+`, `[V4+ Styles]`, and `Style: Default,Microsoft YaHei,22,...`; template in `template.go` has full 23-field V4+ format; 9 tests verify header content including `TestWriteSRTAsASS_StartsWithScriptInfoAndContainsMicrosoftYaHei`; pipeline routes SRT via `track.CodecID == "S_TEXT/UTF8"` (line 131) |
| 3 | Subtitle timestamps in the output file are accurate (Cluster Timestamp + Block offset + TimestampScale correctly computed, millisecond-to-centisecond conversion preserves precision) | VERIFIED | `FormatASSTimestamp` uses integer-only rounding `(ns + 5_000_000) / 10_000_000` (no float64 drift); 11 tests cover zero, boundaries, rounding up/down, multi-digit hours; `ExtractSubtitlePackets` reads `pkt.StartTime`/`pkt.EndTime` from matroska-go demuxer which handles Cluster/Block timestamp computation internally; `ApplyGapFill` handles missing BlockDuration; both writers call `FormatASSTimestamp` for Dialogue line output |
| 4 | SRT formatting tags (italic, bold) are correctly mapped to ASS override tags in the converted output | VERIFIED | `ConvertSRTTagsToASS` in `pkg/subtitle/srt.go` maps `<b>` to `{\b1}`, `<i>` to `{\i1}`, `<u>` to `{\u1}`, `<font color>` to `{\c&HBBGGRR&}` with RGB-to-BGR reversal; 20 test cases cover bold, italic, underline, hex colors, named colors (7 names), nesting, case insensitivity, unknown tag stripping; `WriteSRTAsASS` calls `ConvertSRTTagsToASS(ev.Text)` for each event (srt_writer.go line 45) |
| 5 | Output ASS files render correctly when loaded in a media player (mpv/VLC) | ? UNCERTAIN | Summary claims human verification passed ("output ASS files render correctly in media player, timestamps accurate, styles preserved" per 02-04-SUMMARY.md). However, this is a SUMMARY claim, not code-verifiable. Requires human re-verification to confirm |

**Score:** 4/5 truths verified (1 needs human verification)

### Required Artifacts

| Artifact | Expected | Status | Details |
|----------|----------|--------|---------|
| `pkg/subtitle/types.go` | SubtitleEvent struct | VERIFIED | 17 lines, 11 fields: Start, End, Layer, Style, Name, MarginL/R/V, Effect, Text, ReadOrder |
| `pkg/assout/timestamp.go` | FormatASSTimestamp function | VERIFIED | 22 lines, integer-only arithmetic, correct H:MM:SS.CC format |
| `pkg/subtitle/ass.go` | ParseASSBlockData function | VERIFIED | 43 lines, SplitN(9) for comma handling, returns readOrder/layer/remaining |
| `pkg/subtitle/srt.go` | ConvertSRTTagsToASS function | VERIFIED | 104 lines, 10 compiled regexes, 7 named colors, RGB-to-BGR reversal |
| `pkg/extract/extract.go` | ExtractSubtitlePackets + ApplyGapFill | VERIFIED | 101 lines, demuxer.ReadPacket loop, track filter, gap-fill with sort, sanity check |
| `pkg/assout/ssa_convert.go` | ConvertSSAHeaderToASS function | VERIFIED | 249 lines, full V4-to-V4+ conversion: ScriptType, section header, 18->23 field remap, alignment map, Marked->Layer |
| `pkg/assout/writer.go` | WriteASSPassthrough function | VERIFIED | 122 lines, codec dispatch, CRLF normalization, [Events] deduplication, sorted Dialogue output |
| `pkg/assout/template.go` | defaultASSHeader constant | VERIFIED | 29 lines, Microsoft YaHei 22pt, 1920x1080, V4+ styles, CRLF endings |
| `pkg/assout/srt_writer.go` | WriteSRTAsASS function | VERIFIED | 58 lines, writes default header, sorts events, converts tags, \n-to-\N conversion |
| `pkg/output/naming.go` | GenerateOutputPath function | VERIFIED | 92 lines, {basename}.{lang}.ass format, 3-tier collision resolution, sanitizeFileName |
| `pkg/extract/pipeline.go` | ExtractTrackToASS public API | VERIFIED | 200 lines, full pipeline orchestration, codec dispatch, packetsToEvents, error cleanup |

### Key Link Verification

| From | To | Via | Status | Details |
|------|----|-----|--------|---------|
| pipeline.go | extract.go | ExtractSubtitlePackets | WIRED | Line 99: `packets, warning, err := ExtractSubtitlePackets(demuxer, track.Number)` |
| pipeline.go | writer.go | WriteASSPassthrough | WIRED | Line 129: `err = assout.WriteASSPassthrough(outFile, codecPrivate, track.CodecID, events)` |
| pipeline.go | srt_writer.go | WriteSRTAsASS | WIRED | Line 131: `err = assout.WriteSRTAsASS(outFile, events)` |
| pipeline.go | ass.go | ParseASSBlockData | WIRED | Line 152: `readOrder, layer, remaining, err := subtitle.ParseASSBlockData(pkt.Data)` |
| pipeline.go | naming.go | GenerateOutputPath | WIRED | Line 118: `outputPath := output.GenerateOutputPath(videoForNaming, track, existingPaths)` |
| writer.go | timestamp.go | FormatASSTimestamp | WIRED | Lines 106-107: `start := FormatASSTimestamp(ev.Start)` / `end := FormatASSTimestamp(ev.End)` |
| writer.go | ssa_convert.go | ConvertSSAHeaderToASS | WIRED | Line 32: `header = ConvertSSAHeaderToASS(header)` |
| srt_writer.go | template.go | defaultASSHeader | WIRED | Line 18: `io.WriteString(w, defaultASSHeader)` |
| srt_writer.go | srt.go | ConvertSRTTagsToASS | WIRED | Line 45: `text := subtitle.ConvertSRTTagsToASS(ev.Text)` |
| srt_writer.go | timestamp.go | FormatASSTimestamp | WIRED | Lines 41-42: `start := FormatASSTimestamp(ev.Start)` / `end := FormatASSTimestamp(ev.End)` |
| main.go | pipeline.go | ExtractTrackToASS | WIRED | Line 33: `outputPath, err := extract.ExtractTrackToASS(path, track, "")` |

### Requirements Coverage

| Requirement | Status | Blocking Issue |
|-------------|--------|----------------|
| EXTR-01: Extract SRT (S_TEXT/UTF8) tracks | SATISFIED | Pipeline handles S_TEXT/UTF8 via WriteSRTAsASS path |
| EXTR-02: Extract SSA/ASS (S_TEXT/SSA, S_TEXT/ASS) tracks | SATISFIED | Pipeline handles both via WriteASSPassthrough with codec dispatch |
| EXTR-03: Preserve CodecPrivate styles in ASS extraction | SATISFIED | WriteASSPassthrough writes CodecPrivate verbatim for ASS, converts for SSA |
| EXTR-04: Correct timestamp calculation | SATISFIED | matroska-go handles Cluster+Block timestamps; FormatASSTimestamp rounds correctly |
| CONV-01: Convert SRT to ASS format | SATISFIED | WriteSRTAsASS produces complete ASS files from SRT events |
| CONV-02: Generate proper ASS headers | SATISFIED | defaultASSHeader has [Script Info], [V4+ Styles] with Microsoft YaHei |
| CONV-03: SRT tags mapped to ASS override tags | SATISFIED | ConvertSRTTagsToASS maps b/i/u/font-color with 20 test cases |
| CONV-04: Timestamp precision preserved (ms to cs) | SATISFIED | Integer rounding `(ns + 5_000_000) / 10_000_000` with 11 boundary tests |

### Anti-Patterns Found

| File | Line | Pattern | Severity | Impact |
|------|------|---------|----------|--------|
| (none) | - | - | - | No TODO/FIXME/PLACEHOLDER/stub patterns found in any Phase 2 files |

### Human Verification Required

### 1. ASS/SSA Track Rendering in Media Player

**Test:** Build the tool with `go build -o mkv-sub-extractor.exe ./cmd/mkv-sub-extractor/` and run against an MKV file containing ASS subtitles. Open the output .ass file in mpv or VLC.
**Expected:** Subtitles render with original styles (fonts, colors, positioning) preserved from the source file. Timing matches video playback accurately.
**Why human:** Visual rendering correctness and style fidelity cannot be verified programmatically -- requires observing the rendered output against the video.

### 2. SRT Track Rendering in Media Player

**Test:** Run the tool against an MKV file containing SRT subtitles. Open the output .ass file in mpv or VLC.
**Expected:** Subtitles render with Microsoft YaHei font, white text with black outline, bottom-center alignment. Bold/italic formatting from SRT source is preserved. Timing matches video playback.
**Why human:** Visual appearance of the generated default style and formatting tag conversion requires real-time human observation against video.

### 3. Timestamp Accuracy Across Video Duration

**Test:** Play a video with extracted subtitles and compare subtitle appearance/disappearance to expected dialogue timing, especially at the beginning, middle, and end of the video.
**Expected:** Subtitles appear and disappear within ~100ms of the correct moments. No visible drift or cumulative offset.
**Why human:** Timing precision evaluation against running video playback requires real-time observation. The integer arithmetic is proven correct by unit tests, but the upstream timestamp values from matroska-go can only be validated against real media.

### Gaps Summary

No automated gaps found. All 11 artifacts exist, are substantive (no stubs or placeholders), and are fully wired together. All 79 tests pass. All 10 key links are verified at the code level. All 8 requirements (EXTR-01 through EXTR-04, CONV-01 through CONV-04) are satisfied by the implementation.

The only item requiring attention is human verification of media player rendering (Truth 5). The 02-04-SUMMARY.md claims this was verified during plan execution ("output ASS files render correctly in media player"), but SUMMARY claims are not treated as verified evidence by this report. A human should confirm rendering correctness with real MKV files.

---

_Verified: 2026-02-16T15:30:00Z_
_Verifier: Claude (gsd-verifier)_
