# MKV Subtitle Extractor

## What This Is

一个纯 Go 实现的命令行工具，从 MKV 视频文件中提取文本字幕轨道（SRT、SSA/ASS），并转换为 ASS 格式保存。支持交互式选择和命令行参数两种模式，ASS/SSA 字幕保留原始样式，SRT 字幕自动转换。

## Core Value

用户能从 MKV 文件中快速、可靠地提取字幕并得到 ASS 格式的输出文件。

## Requirements

### Validated

- ✓ 纯 Go 解析 MKV (EBML/Matroska) 容器格式，读取轨道信息 — v1.0
- ✓ 识别并提取文本字幕轨道（SRT、SSA/ASS） — v1.0
- ✓ 将提取的字幕转换为 ASS 格式输出 — v1.0
- ✓ 交互式文件选择：不传参时扫描当前目录列出 MKV 文件供用户选择 — v1.0
- ✓ 命令行参数直接指定 MKV 文件路径 — v1.0
- ✓ 交互式轨道选择：列出所有字幕轨道（含语言、格式信息），用户输入编号选择 — v1.0
- ✓ 输出文件默认保存到 MKV 同目录，文件名自动生成 — v1.0
- ✓ 支持 `--output` 参数指定输出目录 — v1.0

### Active

(None — v1.0 shipped, next milestone not planned)

### Out of Scope

- 图片字幕（PGS/VobSub）提取 — 需要 OCR，复杂度远超文本字幕
- 批量处理 — v1 聚焦单文件，批量后续考虑
- 调用外部工具（ffmpeg/mkvextract）— 设计目标是纯 Go 零依赖
- 输出其他格式（SRT/VTT 等）— 聚焦 ASS 单一输出格式
- GUI/完整 TUI 界面 — 简单交互式 CLI 足够

## Context

Shipped v1.0 with 4,571 LOC Go.
Tech stack: matroska-go (MKV parsing), charmbracelet/huh (interactive TUI), lipgloss (terminal styling), pflag (flags), progressbar/v3.
Human-tested with real MKV files containing ASS, SRT, and image-based subtitle tracks.
7 bugs found and fixed during human testing (filename collision, V4 Styles conversion, empty field handling, font sizing, etc.).

## Constraints

- **Language**: Go — 项目语言要求
- **Dependencies**: 零外部运行时依赖 — 编译为单个二进制文件分发
- **Subtitle scope**: 仅文本字幕 — 不处理图片字幕格式

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 纯 Go 实现 | 零外部依赖，单文件分发，跨平台 | ✓ Good |
| 仅支持文本字幕 | 图片字幕需 OCR，复杂度过高 | ✓ Good |
| 交互式 + 参数两种模式 | 兼顾易用性和脚本调用场景 | ✓ Good |
| matroska-go 库 | Go 生态中少数可用的 MKV 解析库 | ✓ Good (需 Duration float 变通) |
| ASS passthrough 模式 | 保留原始样式，不破坏字幕效果 | ✓ Good |
| 始终执行 V4→V4+ 转换 | 部分 ASS 轨道 CodecPrivate 含 V4 格式头部 | ✓ Good (人工测试验证) |
| SRT 字号 58pt@1080p | 匹配播放器默认 SRT 渲染大小 | ✓ Good |
| 共享 existingPaths 批量命名 | 防止同语言多轨道文件名冲突 | ✓ Good (人工测试验证) |
| charmbracelet/huh 交互 | 美观的箭头键选择器和多选 | ✓ Good (Validate 需 post-submit) |

---
*Last updated: 2026-02-17 after v1.0 milestone*
