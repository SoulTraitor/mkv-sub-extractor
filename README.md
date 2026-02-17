# mkv-sub-extractor

从 MKV 文件中提取文本字幕轨道并输出为 ASS 格式的命令行工具。

支持交互式选择和命令行参数两种模式，ASS/SSA 字幕保留原始样式，SRT 字幕自动转换为 ASS 格式。

## 功能

- 解析 MKV 容器，列出所有字幕轨道及元数据（编号、编码格式、语言、名称）
- ASS/SSA 字幕原样提取，保留 CodecPrivate 中的所有样式定义
- SRT 字幕自动转换为 ASS 格式（Microsoft YaHei 字体，1080p 分辨率）
- SSA V4 格式头部自动转换为 ASS V4+ 格式
- 交互式文件选择和字幕轨多选
- 非交互式批量提取（`--track` 参数）
- 智能输出文件命名，自动处理同语言轨道的文件名冲突
- 图片类字幕轨（PGS/VobSub）在界面中标注但不可提取

## 安装

需要 Go 1.25+：

```bash
go install github.com/SoulTraitor/mkv-sub-extractor/cmd/mkv-sub-extractor@latest
```

或从源码构建：

```bash
git clone https://github.com/SoulTraitor/mkv-sub-extractor.git
cd mkv-sub-extractor
go build -o mkv-sub-extractor ./cmd/mkv-sub-extractor/
```

## 使用

### 交互模式

```bash
# 扫描当前目录的 MKV 文件，交互选择
mkv-sub-extractor

# 指定文件，交互选择字幕轨
mkv-sub-extractor video.mkv
```

### 命令行模式

```bash
# 提取指定轨道（非交互）
mkv-sub-extractor video.mkv --track 3,5

# 指定输出目录
mkv-sub-extractor video.mkv --track 3 --output ./subs/

# 静默模式，只输出文件路径
mkv-sub-extractor video.mkv --track 3 --quiet
```

### 参数

| 参数 | 缩写 | 说明 |
|------|------|------|
| `--track` | `-t` | 指定提取的轨道编号，逗号分隔 |
| `--output` | `-o` | 输出目录（默认与 MKV 文件同目录） |
| `--quiet` | `-q` | 静默模式，仅输出文件路径 |
| `--verbose` | `-v` | 详细输出 |

### 输出文件命名

输出文件名格式为 `{视频名}.{语言代码}.ass`，例如：

```
video.eng.ass       # 英文字幕
video.chi.ass       # 中文字幕
video.chi.繁體.ass  # 同语言多轨时使用轨道名称区分
video.chi.2.ass     # 无轨道名称时使用序号区分
```

## 支持的字幕格式

### 可提取（文本类）

| 编码格式 | 说明 |
|----------|------|
| S_TEXT/ASS | ASS 字幕（原样提取） |
| S_TEXT/SSA | SSA 字幕（自动转换为 ASS） |
| S_TEXT/UTF8 | SRT 字幕（转换为 ASS） |

### 仅显示（图片类）

| 编码格式 | 说明 |
|----------|------|
| S_HDMV/PGS | Blu-ray PGS 字幕 |
| S_VOBSUB | DVD VobSub 字幕 |

图片类字幕需要 OCR 工具（如 SubtitleEdit）处理，本工具不支持提取。

## 退出码

| 退出码 | 含义 |
|--------|------|
| 0 | 成功 |
| 1 | 通用错误 |
| 2 | 文件错误（未找到、非 MKV、无法读取） |
| 3 | 轨道错误（无字幕、图片类轨道、轨道不存在） |
| 4 | 提取错误 |

## 依赖

- [matroska-go](https://github.com/luispater/matroska-go) — MKV 容器解析
- [huh](https://github.com/charmbracelet/huh) — 交互式 TUI 表单
- [lipgloss](https://github.com/charmbracelet/lipgloss) — 终端样式
- [pflag](https://github.com/spf13/pflag) — POSIX 风格命令行参数解析
- [progressbar](https://github.com/schollz/progressbar) — 进度条

## License

[MIT](LICENSE)
