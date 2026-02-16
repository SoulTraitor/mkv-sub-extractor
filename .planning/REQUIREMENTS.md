# Requirements: MKV Subtitle Extractor

**Defined:** 2026-02-16
**Core Value:** 用户能从 MKV 文件中快速、可靠地提取字幕并得到 ASS 格式的输出文件

## v1 Requirements

Requirements for initial release. Each maps to roadmap phases.

### Parsing

- [ ] **PARS-01**: 工具能解析 MKV (EBML/Matroska) 容器格式，正确处理 VINT 编码
- [ ] **PARS-02**: 工具能读取 Segment、Tracks、Cluster 等 Matroska 核心元素
- [ ] **PARS-03**: 工具能处理未知大小 (unknown-size) 元素
- [ ] **PARS-04**: 工具使用 io.ReadSeeker 流式处理，支持大文件（1-50+ GB）

### Track Discovery

- [ ] **TRAK-01**: 工具能列出 MKV 文件中所有字幕轨道
- [ ] **TRAK-02**: 轨道列表显示轨道编号、编解码器类型、语言标签、轨道名称
- [ ] **TRAK-03**: 遇到非文本字幕轨道（PGS/VobSub）时显示清晰提示而非报错

### Extraction

- [ ] **EXTR-01**: 工具能提取 SRT (S_TEXT/UTF8) 字幕轨道的内容
- [ ] **EXTR-02**: 工具能提取 SSA/ASS (S_TEXT/SSA, S_TEXT/ASS) 字幕轨道的内容
- [ ] **EXTR-03**: ASS 轨道提取时保留 CodecPrivate 中的原始样式信息（直通模式）
- [ ] **EXTR-04**: 正确计算字幕时间戳（Cluster Timestamp + Block Offset + TimestampScale）

### Conversion

- [ ] **CONV-01**: 工具能将 SRT 字幕转换为 ASS 格式输出
- [ ] **CONV-02**: 转换时生成合理的 ASS 头部（[Script Info]、[V4+ Styles]）
- [ ] **CONV-03**: SRT 格式标签（`<i>`、`<b>`）正确映射为 ASS override tags（`{\i1}`、`{\b1}`）
- [ ] **CONV-04**: 时间戳转换保持精度（毫秒→厘秒正确舍入）

### CLI Interface

- [ ] **CLI-01**: 用户可通过命令行参数直接指定 MKV 文件路径
- [ ] **CLI-02**: 不传文件参数时，扫描当前目录列出 MKV 文件供交互式选择
- [ ] **CLI-03**: 运行后交互式列出所有字幕轨道，用户输入编号选择
- [ ] **CLI-04**: 支持 `--track` 参数直接指定轨道编号跳过交互选择
- [ ] **CLI-05**: 输出文件默认保存到 MKV 同目录，文件名自动生成（如 `video.en.ass`）
- [ ] **CLI-06**: 支持 `--output` 参数指定输出目录

### Error Handling

- [ ] **ERRH-01**: 非 MKV 文件输入时给出明确错误提示
- [ ] **ERRH-02**: 文件不存在或无法读取时给出明确错误提示
- [ ] **ERRH-03**: 无字幕轨道时给出明确提示
- [ ] **ERRH-04**: 图片字幕轨道被选择时提示"仅支持文本字幕"并建议替代工具

## v2 Requirements

Deferred to future release. Tracked but not in current roadmap.

### Enhanced Features

- **ENCX-01**: 非 UTF-8 编码检测和转换（CJK 内容）
- **BATC-01**: 批量处理，支持传入目录处理所有 MKV 文件
- **LIST-01**: `--list` 模式，只列出轨道信息不提取
- **PREV-01**: 字幕预览，提取前显示前几行内容
- **FONT-01**: 从 MKV 附件中提取字体文件
- **WEBV-01**: 支持 S_TEXT/WEBVTT 格式提取和转换

## Out of Scope

| Feature | Reason |
|---------|--------|
| 图片字幕提取（PGS/VobSub/DVB） | 需要 OCR，复杂度过高，违反零依赖目标 |
| 字幕编辑/时间轴调整 | SubtitleEdit/Aegisub 专业领域 |
| GUI/完整 TUI 界面 | 过度工程，简单交互式 CLI 足够 |
| 字幕嵌入 MKV（muxing） | 写入 MKV 是完全不同的领域 |
| 音频/视频流提取 | 不同领域，推荐 ffmpeg |
| 网络/URL 输入 | 仅支持本地文件 |
| 输出其他格式（SRT/VTT 等） | 聚焦 ASS 单一输出格式 |
| 调用外部工具（ffmpeg/mkvextract） | 设计目标是纯 Go 零依赖 |

## Traceability

| Requirement | Phase | Status |
|-------------|-------|--------|
| PARS-01 | Phase 1 | Pending |
| PARS-02 | Phase 1 | Pending |
| PARS-03 | Phase 1 | Pending |
| PARS-04 | Phase 1 | Pending |
| TRAK-01 | Phase 1 | Pending |
| TRAK-02 | Phase 1 | Pending |
| TRAK-03 | Phase 1 | Pending |
| EXTR-01 | Phase 2 | Pending |
| EXTR-02 | Phase 2 | Pending |
| EXTR-03 | Phase 2 | Pending |
| EXTR-04 | Phase 2 | Pending |
| CONV-01 | Phase 2 | Pending |
| CONV-02 | Phase 2 | Pending |
| CONV-03 | Phase 2 | Pending |
| CONV-04 | Phase 2 | Pending |
| CLI-01 | Phase 3 | Pending |
| CLI-02 | Phase 3 | Pending |
| CLI-03 | Phase 3 | Pending |
| CLI-04 | Phase 3 | Pending |
| CLI-05 | Phase 3 | Pending |
| CLI-06 | Phase 3 | Pending |
| ERRH-01 | Phase 3 | Pending |
| ERRH-02 | Phase 3 | Pending |
| ERRH-03 | Phase 3 | Pending |
| ERRH-04 | Phase 3 | Pending |

**Coverage:**
- v1 requirements: 25 total
- Mapped to phases: 25
- Unmapped: 0

---
*Requirements defined: 2026-02-16*
*Last updated: 2026-02-16 after roadmap creation*
