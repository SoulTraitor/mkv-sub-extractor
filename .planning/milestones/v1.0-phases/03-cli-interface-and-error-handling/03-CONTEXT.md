# Phase 3: CLI Interface and Error Handling - Context

**Gathered:** 2026-02-17
**Status:** Ready for planning

<domain>
## Phase Boundary

Polished command-line experience for the MKV subtitle extractor. Covers interactive file/track selection, argument parsing, output directory control, progress reporting, and structured error messages. The extraction engine (Phase 2) is complete — this phase wraps it in a user-friendly CLI.

</domain>

<decisions>
## Implementation Decisions

### Interactive Selection UX
- Arrow-key menu for both file picker and track selector (promptui/survey style, not numbered list)
- Track selector supports multi-select (user can check multiple subtitle tracks for batch extraction)
- Always confirm selection even when only one MKV file or one subtitle track exists (no auto-skip)
- Image-based subtitle tracks (PGS/VobSub) appear in the track selector but are grayed out / disabled with a visual indicator that they are not extractable

### Command-Line Argument Design
- File path as positional argument: `mkv-sub-extractor video.mkv` (no flag needed for input file)
- No arguments → scan current directory for MKV files → interactive file picker
- Track specification: `--track 3,5` (comma-separated track numbers for non-interactive/scriptable mode)
- Output directory: `--output <dir>` overrides default (default = same directory as MKV file)
- No `--list` mode — interactive mode already shows track listing
- When `--track` is specified, run non-interactively (scriptable for CI/automation)

### Output Feedback and Progress
- Real-time progress bar during extraction (based on processed packets / total packets or similar metric)
- `--quiet` flag suppresses progress output (only final file paths on stdout)
- `--verbose` flag adds debug-level information (packet counts, timing details, codec decisions)
- Default verbosity: progress bar + completion summary

### Completion Summary
- Claude's Discretion — design the most informative yet concise completion output (file paths, track info, subtitle counts are all reasonable)

### Error Message Style
- Structured errors similar to rustc/elm: title line + context + suggestion, multi-line format
- Each error type has a distinct, human-readable message (not generic "something went wrong")
- Exit codes are categorized by error type (1=general, 2=file error, 3=track error, etc.)
- Multi-track extraction: if one track fails, continue extracting remaining tracks; report all failures at the end
- Partial output files cleaned up on individual track failure (already implemented in Phase 2 pipeline)

### Claude's Discretion
- Exact progress bar library/implementation choice
- Completion summary format and content
- Short flag aliases (e.g., -t for --track, -o for --output, -q for --quiet, -v for --verbose)
- Help text formatting and examples
- Color usage in terminal output (errors, progress, success)

</decisions>

<specifics>
## Specific Ideas

- Interactive selection should feel responsive and polished — arrow keys, visual highlight on current selection
- Error messages should help the user fix the problem, not just report it (like Rust compiler errors)
- Scriptable mode (with --track) should be pipeline-friendly: predictable output, meaningful exit codes, no interactive prompts

</specifics>

<deferred>
## Deferred Ideas

None — discussion stayed within phase scope

</deferred>

---

*Phase: 03-cli-interface-and-error-handling*
*Context gathered: 2026-02-17*
