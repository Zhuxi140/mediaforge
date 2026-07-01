# 后端接口合约 (Controller API Contract)

> Wails v2 通过 Go struct 方法绑定将后端方法暴露给前端。前端通过 `wailsjs/go/controller/{AppName}.js` 调用。
> 所有方法均为异步（返回 `Promise`）。

---

## RenamerApp

绑定路径：`controller.RenamerApp`

### SelectFiles

- **签名**: `SelectFiles() => Promise<string[]>`
- **描述**: 打开系统文件选择对话框，返回选中文件的绝对路径数组
- **对话框过滤**: 所有文件 / 视频(`.mp4;.mkv`) / 音频(`.mp3;.wav`) / 图片(`.jpg;.png`) / 字幕(`.srt;.ass`)
- **异常**: 对话框取消或出错时返回 `null`
- **调用方**: 文件选择按钮、拖拽添加

### PreviewRename

- **签名**: `PreviewRename(Paths: string[], rule: RenameRule) => Promise<RenamePreview[]>`
- **描述**: 生成重命名预览列表，不执行实际文件操作
- **参数**:
  - `Paths`: 文件绝对路径数组
  - `rule`: 重命名规则（含模式、前后缀、替换规则、正则规则、模板）
- **返回**: 每条包含 `originalPath`, `originalName`, `newName`, `newPath`, `hasConflict`, `formatError`
- **副作用**: 可能检查目标文件是否存在（冲突检测）
- **调用方**: RenamerPanel 文件列表变化或规则变化时

### ApplyRename

- **签名**: `ApplyRename(Previews: RenamePreview[]) => Promise<string>`
- **描述**: 执行文件重命名
- **参数**: `Previews` — 需包含 `originalPath` 和 `newPath`
- **返回**: `"success"` 或错误描述字符串
- **安全**: 遇同名冲突或格式错误立即中止并返回错误
- **调用方**: "执行重命名"按钮

### QuickRename

- **签名**: `QuickRename(oldPath: string, newName: string) => Promise<string>`
- **描述**: 单文件快速重命名（行内编辑）
- **参数**:
  - `oldPath`: 原文件绝对路径
  - `newName`: 新文件名（含扩展名）
- **返回**: 新文件绝对路径
- **异常**: 目标文件已存在时返回错误
- **调用方**: 行内编辑"保存"按钮

---

## MediaApp

绑定路径：`controller.MediaApp`

### SubmitMediaTask

- **签名**: `SubmitMediaTask(task: FFmpegTask) => Promise<void>`
- **描述**: 提交音视频转码任务（异步执行）
- **参数**:
  - `task.ID`: 唯一标识（前端生成）
  - `task.InputPath`: 输入文件绝对路径
  - `task.OutputPath`: 输出目录（空则输出到原文件同级）
  - `task.Format`: 目标格式（如 `"mp4"`、`"mp3"`）
  - `task.Quality`: `"Low" | "Medium" | "High" | "Custom"`
  - `task.VideoCRF`: 自定义视频 CRF（Quality=Custom 时使用）
  - `task.AudioVBR`: 自定义音频 VBR（Quality=Custom 时使用）
  - `task.HwAccel`: `"cpu" | "nvidia" | "intel" | "amd" | "mac"`
  - `task.ForceDropSubtitle`: 是否强制丢弃字幕
- **返回值/事件**: 无直接返回，通过 Wails Events 接收结果：
  - `ffmpeg-progress-{task.ID}`: 进度时间字符串
  - `ffmpeg-done-{task.ID}`: 成功
  - `ffmpeg-error-{task.ID}`: 错误原因字符串
- **特殊返回值**: 函数返回错误字符串 `"WARN_SUBTITLE"` 表示字幕不兼容需用户确认（通过 `throw` 传递）
- **调用方**: FFmpegPanel 转换按钮

### CancelMediaTask

- **签名**: `CancelMediaTask(id: string) => Promise<void>`
- **描述**: 取消指定 ID 的转码任务
- **实现**: 调用 `context.CancelFunc` + 从 `sync.Map` 删除
- **事件**: 被取消任务触发 `ffmpeg-error-{id}` 事件，消息为 `"任务已取消"`
- **调用方**: "停止"按钮

### CheckInputFile

- **签名**: `CheckInputFile(filePath: string) => Promise<string>`
- **描述**: 检查文件是否为支持的媒体格式
- **返回**: `"success"` 或错误描述（如 `"暂不支持的格式: .TXT"`）
- **校验范围**: 视频扩展名、音频扩展名、gif
- **调用方**: 添加媒体文件时

### ScanSubtitles

- **签名**: `ScanSubtitles(inputPath: string) => Promise<SubtitleStream[]>`
- **描述**: 扫描视频文件的内封字幕流
- **参数**: `inputPath` — 视频文件绝对路径
- **返回**: `SubtitleStream[]` 数组，每项含 `Index`(流索引)、`Language`、`Codec`
- **异常**: 非视频文件或 FFmpeg 未初始化时返回错误
- **实现**: 执行 `ffmpeg -i inputFile`，解析输出中的 `Stream #0:N ... Subtitle:` 行
- **调用方**: SubtitlePanel 添加视频文件时

### ExtractSubtitle

- **签名**: `ExtractSubtitle(id, inputPath, streamIndex, outDir, targetFormat: string) => Promise<void>`
- **描述**: 从视频中异步提取指定字幕流
- **参数**:
  - `id`: 任务 ID（用于事件追踪）
  - `inputPath`: 视频路径
  - `streamIndex`: 字幕流索引（如 `"0"`）
  - `outDir`: 输出目录（空则输出到视频同级）
  - `targetFormat`: 目标格式（如 `"srt"`、`"ass"`）
- **事件**: `ffmpeg-done-{id}` / `ffmpeg-error-{id}`
- **调用方**: SubtitlePanel 提取按钮

### ConvertSubtitle

- **签名**: `ConvertSubtitle(inputPath: string, outDir: string, outPutName: string, targetFormat: string) => Promise<void>`
- **描述**: 字幕格式互转（同步执行）
- **参数**:
  - `inputPath`: 源字幕文件路径
  - `outDir`: 输出目录
  - `outPutName`: 输出文件名（不填则保持原名）
  - `targetFormat`: 目标格式（如 `"ass"`、`"vtt"`）
- **异常**: 源文件非字幕格式或转换失败时返回错误
- **调用方**: SubtitlePanel 互转按钮

### SelectMediaFile

- **签名**: `SelectMediaFile() => Promise<string>`
- **描述**: 打开系统文件选择对话框，仅显示媒体文件（视频/音频），返回选中文件的绝对路径
- **对话框过滤**: 视频(`.mp4;.mkv;.avi;.mov;.wmv;.flv;.webm`) / 音频(`.mp3;.wav;.aac;.flac;.ogg;.wma;.m4a`)
- **异常**: 对话框取消或出错时返回空字符串 `""`
- **调用方**: MediaInfoPanel 选择媒体文件按钮

### GetMediaInfo

- **签名**: `GetMediaInfo(inputPath: string) => Promise<media.MediaInfo>`
- **描述**: 获取指定媒体文件的完整元数据信息
- **参数**: `inputPath` — 媒体文件绝对路径
- **返回**: `media.MediaInfo` 结构体，包含通用信息、视频流数组、音频流数组、字幕流数组
- **实现**: 优先调用 `ffprobe -v quiet -print_format json -show_format -show_streams` 解析 JSON；若 ffprobe 不可用则回退到 `ffmpeg -i` 文本解析
- **异常**: 文件不存在、非媒体文件或解析失败时返回错误
- **调用方**: MediaInfoPanel 选择文件后自动调用

---

## M3U8App

### ProbeM3U8URL

- **签名**: `ProbeM3U8URL(url: string) => Promise<media.MediaInfo>`
- **描述**: 验证 M3U8 链接有效性，获取视频流的元数据信息
- **参数**: `url` — M3U8 完整链接地址
- **返回**: `media.MediaInfo` 结构体（不含文件名和文件大小）
- **实现**: 调用 `ffprobe -v quiet -print_format json -show_format -show_streams`，30 秒超时
- **异常**: 链接无效、超时、非视频流时返回错误
- **调用方**: M3U8Panel 验证按钮

### DownloadM3U8

- **签名**: `DownloadM3U8(url: string, outputPath: string) => Promise<string>`
- **描述**: 启动 M3U8 流媒体 FFmpeg 下载任务
- **参数**:
  - `url` — M3U8 链接
  - `outputPath` — 保存文件完整路径
- **返回**: 任务 ID（字符串）
- **实现**: 启动 FFmpeg 子进程下载，实时计算速度/进度，通过事件推送到前端
- **事件**: `m3u8-progress-{id}`, `m3u8-done-{id}`, `m3u8-error-{id}`

### CancelM3U8Task

- **签名**: `CancelM3U8Task(taskId: string) => void`
- **描述**: 取消指定 M3U8 下载任务
- **参数**: `taskId` — 任务 ID
- **实现**: 通过 `m3u8Processes` map 直接调用 `cmd.Process.Kill()` 终止 FFmpeg 进程 + 取消 context

### CancelDownload

- **签名**: `CancelDownload() => void`
- **描述**: 取消当前正在运行的 M3U8 下载任务
- **备注**: 等效于 `CancelM3U8Task`，用于旧版前端兼容

### SelectOutputDir

- **签名**: `SelectOutputDir() => Promise<string>`
- **描述**: 弹出系统文件选择对话框，选择 M3U8 下载保存目录
- **返回**: 选择的目录路径，取消返回空字符串

---

## 类型定义对照

### Go → TypeScript (Wails 自动生成)

| Go struct | TypeScript class | 说明 |
|-----------|-----------------|------|
| `media.FFmpegTask` | `media.FFmpegTask` | 转码任务参数 |
| `media.SubtitleStream` | `media.SubtitleStream` | 字幕流信息 |
| `media.MediaInfo` | `media.MediaInfo` | 媒体元数据（通用信息+各流数组） |
| `media.VideoStream` | `media.VideoStream` | 视频流信息（编码、分辨率、帧率等） |
| `media.AudioStream` | `media.AudioStream` | 音频流信息（编码、采样率、声道等） |
| `renamer.RenameRule` | `renamer.RenameRule` | 重命名规则 |
| `renamer.RenamePreview` | `renamer.RenamePreview` | 重命名预览项 |
| `renamer.ExtractRule` | `renamer.ExtractRule` | 正则提取规则 |

### 前端自定义类型（`frontend/src/types.ts`）

| 接口 | 对应 Go struct | 用途 |
|------|---------------|------|
| `MediaTask` | — | UI 层转码任务包装（含状态） |
| `SubtitleTask` | — | UI 层字幕提取任务包装 |
| `SubConvertTask` | — | UI 层字幕互转任务包装 |
| `M3U8Task` | — | UI 层 M3U8 下载任务包装（含 url/output/进度） |
| `MediaSettings` | — | 转码设置持久化 |
| `ModalState` | — | 弹窗状态 |

---

## 事件约定

| 事件模式 | 载荷类型 | 触发时机 |
|---------|---------|---------|
| `ffmpeg-progress-{id}` | `string` | 转码进度更新（`time=` 行） |
| `ffmpeg-done-{id}` | `string` | 转码/提取成功 |
| `ffmpeg-error-{id}` | `string` | 转码/提取失败（含 `"任务已取消"`） |
| `m3u8-progress-{id}` | `string` | M3U8 下载进度更新 |
| `m3u8-done-{id}` | `string` | M3U8 下载完成 |
| `m3u8-error-{id}` | `string` | M3U8 下载失败或取消 |
