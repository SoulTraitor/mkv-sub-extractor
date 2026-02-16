# Roadmap: MKV Subtitle Extractor

## Overview

Build a pure Go CLI tool that extracts text subtitles from MKV files and outputs them as ASS format. The project progresses from container parsing (can the tool read MKV files and find subtitle tracks?) through extraction and conversion (can it pull subtitle data out and produce valid ASS files?) to polished CLI experience (can a user interactively select files and tracks with clear error messages?). Each phase delivers a testable, standalone capability.

## Phases

**Phase Numbering:**
- Integer phases (1, 2, 3): Planned milestone work
- Decimal phases (2.1, 2.2): Urgent insertions (marked with INSERTED)

Decimal phases appear between their surrounding integers in numeric order.

- [x] **Phase 1: MKV Parsing and Track Discovery** - Parse EBML/Matroska containers and list subtitle tracks with metadata
- [ ] **Phase 2: Subtitle Extraction and ASS Output** - Extract subtitle block data and produce valid ASS files (passthrough and SRT conversion)
- [ ] **Phase 3: CLI Interface and Error Handling** - Interactive file/track selection, argument parsing, output control, and robust error messages

## Phase Details

### Phase 1: MKV Parsing and Track Discovery
**Goal**: User can point the tool at any MKV file and see a list of all subtitle tracks with their metadata
**Depends on**: Nothing (first phase)
**Requirements**: PARS-01, PARS-02, PARS-03, PARS-04, TRAK-01, TRAK-02, TRAK-03
**Success Criteria** (what must be TRUE):
  1. Tool can open a valid MKV file and parse its EBML header without error
  2. Tool correctly lists all subtitle tracks in an MKV file, showing track number, codec type (SRT/SSA/ASS), language, and name
  3. Tool handles large MKV files (1+ GB) without excessive memory usage (streaming/seeking, not buffering)
  4. Non-text subtitle tracks (PGS/VobSub) appear in the listing with a clear indicator that they are image-based and not extractable
**Plans:** 2 plans

Plans:
- [x] 01-01-PLAN.md -- Project setup, domain types, codec classification, language resolution, formatting utilities, and core MKV parsing function
- [x] 01-02-PLAN.md -- Display formatting for track listings and file metadata, main.go entry point, human verification with real MKV file

### Phase 2: Subtitle Extraction and ASS Output
**Goal**: User can select a subtitle track and get a valid, correctly-timed ASS file as output
**Depends on**: Phase 1
**Requirements**: EXTR-01, EXTR-02, EXTR-03, EXTR-04, CONV-01, CONV-02, CONV-03, CONV-04
**Success Criteria** (what must be TRUE):
  1. Extracting an ASS/SSA track produces a valid ASS file that preserves all original styles from CodecPrivate (passthrough mode)
  2. Extracting an SRT track produces a valid ASS file with properly generated Script Info and V4+ Styles headers
  3. Subtitle timestamps in the output file are accurate (Cluster Timestamp + Block offset + TimestampScale correctly computed, millisecond-to-centisecond conversion preserves precision)
  4. SRT formatting tags (italic, bold) are correctly mapped to ASS override tags in the converted output
  5. Output ASS files render correctly when loaded in a media player (mpv/VLC)
**Plans**: TBD

Plans:
- [ ] 02-01: TBD
- [ ] 02-02: TBD

### Phase 3: CLI Interface and Error Handling
**Goal**: User has a polished, self-explanatory command-line tool that works both interactively and via arguments, with clear guidance on errors
**Depends on**: Phase 2
**Requirements**: CLI-01, CLI-02, CLI-03, CLI-04, CLI-05, CLI-06, ERRH-01, ERRH-02, ERRH-03, ERRH-04
**Success Criteria** (what must be TRUE):
  1. Running the tool with no arguments in a directory containing MKV files presents an interactive file picker, then an interactive track selector, and produces an ASS file
  2. Running the tool with a file path argument and --track flag extracts the specified track non-interactively (scriptable mode)
  3. Output files are saved to the MKV's directory by default with auto-generated names (e.g., video.en.ass), and --output overrides the directory
  4. Passing a non-MKV file, a nonexistent path, a file with no subtitles, or selecting an image subtitle track each produces a distinct, human-readable error message
**Plans**: TBD

Plans:
- [ ] 03-01: TBD
- [ ] 03-02: TBD

## Progress

**Execution Order:**
Phases execute in numeric order: 1 -> 2 -> 3

| Phase | Plans Complete | Status | Completed |
|-------|----------------|--------|-----------|
| 1. MKV Parsing and Track Discovery | 2/2 | Complete | 2026-02-16 |
| 2. Subtitle Extraction and ASS Output | 0/? | Not started | - |
| 3. CLI Interface and Error Handling | 0/? | Not started | - |
