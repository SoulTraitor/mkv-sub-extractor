# Milestones

## v1.0 MVP (Shipped: 2026-02-17)

**Phases completed:** 3 phases, 8 plans
**Lines of code:** 4,571 Go
**Timeline:** 2 days (2026-02-16 → 2026-02-17)

**Delivered:** A pure Go CLI tool that extracts text subtitles from MKV files and outputs ASS format, with both interactive and scriptable modes.

**Key accomplishments:**
- Pure Go MKV container parsing with matroska-go, supporting streaming/seeking for large files
- Complete subtitle track discovery with codec classification, language resolution, and metadata display
- ASS/SSA subtitle extraction with original style preservation (passthrough mode)
- SRT-to-ASS conversion with HTML tag mapping, default template, and proper timestamp handling
- Interactive CLI with file picker and track multi-selector (charmbracelet/huh)
- Scriptable mode with --track/--output/--quiet/--verbose flags and structured error messages

---

