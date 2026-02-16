# MKV Subtitle Extractor

## What This Is

一个纯 Go 实现的命令行工具，从 MKV 视频文件中提取文本字幕轨道（SRT、SSA/ASS），并转换为 ASS 格式保存。零外部依赖，编译为单个二进制文件即可使用。

## Core Value

用户能从 MKV 文件中快速、可靠地提取字幕并得到 ASS 格式的输出文件。

## Requirements

### Validated

(None yet — ship to validate)

### Active

- [ ] 纯 Go 解析 MKV (EBML/Matroska) 容器格式，读取轨道信息
- [ ] 识别并提取文本字幕轨道（SRT、SSA/ASS）
- [ ] 将提取的字幕转换为 ASS 格式输出
- [ ] 交互式文件选择：不传参时扫描当前目录列出 MKV 文件供用户选择
- [ ] 命令行参数直接指定 MKV 文件路径
- [ ] 交互式轨道选择：列出所有字幕轨道（含语言、格式信息），用户输入编号选择
- [ ] 输出文件默认保存到 MKV 同目录，文件名自动生成
- [ ] 支持 `--output` 参数指定输出目录

### Out of Scope

- 图片字幕（PGS/VobSub）提取 — 需要 OCR，复杂度远超文本字幕
- 批量处理 — v1 聚焦单文件，批量后续考虑
- 调用外部工具（ffmpeg/mkvextract）— 设计目标是纯 Go 零依赖

## Context

- MKV 容器基于 EBML（Extensible Binary Meta Language）格式
- Matroska 规范定义了轨道类型、编解码器 ID 等元数据
- 常见文本字幕编解码器 ID：S_TEXT/UTF8 (SRT)、S_TEXT/SSA、S_TEXT/ASS
- Go 生态中 MKV/EBML 解析库相对较少，可能需要评估可用库或部分自行实现
- ASS (Advanced SubStation Alpha) 是功能丰富的字幕格式，支持样式、定位等

## Constraints

- **Language**: Go — 项目语言要求
- **Dependencies**: 零外部运行时依赖 — 编译为单个二进制文件分发
- **Subtitle scope**: 仅文本字幕 — 不处理图片字幕格式

## Key Decisions

| Decision | Rationale | Outcome |
|----------|-----------|---------|
| 纯 Go 实现 | 零外部依赖，单文件分发，跨平台 | — Pending |
| 仅支持文本字幕 | 图片字幕需 OCR，复杂度过高 | — Pending |
| 交互式 + 参数两种模式 | 兼顾易用性和脚本调用场景 | — Pending |

---
*Last updated: 2026-02-16 after initialization*
