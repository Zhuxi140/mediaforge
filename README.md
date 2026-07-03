# MediaForge

基于 Wails v2 (Go + Vue3 + Pinia) 与 FFmpeg 的桌面影音工具集。

## 功能模块

| 模块 | 说明 |
|------|------|
| 批量重命名 | 基础替换（前缀/后缀/查找替换）+ 正则智能解析，实时预览，冲突检测 |
| 影音工厂 | 视频/音频格式互转（MP4/MKV/AVI/GIF/MP3/WAV/AAC/FLAC），硬件加速编码，GIF 输出，内置元数据查看器 |
| 字幕处理 | 从视频中提取内封字幕流，SRT/ASS/VTT/SSA 等常见格式互转，特殊格式预警 |
| M3U8 流媒体 | 粘贴 M3U8 链接自动验证，hls.js 内置播放器预览，FFmpeg 实时下载 |

## 开发

```bash
wails dev          # 热重载开发（Go + Vite）
```

前端单独开发（`frontend/` 目录下）：
```bash
npm run dev        # Vite 开发服务器
npm run build      # vue-tsc 类型检查 + vite 构建
```

## 构建

```bash
wails build        # → build/bin/mediaforge.exe
```

FFmpeg 和 FFprobe 引擎通过 `//go:embed bin/*` 内嵌，首次运行自动释放到 `%APPDATA%/MediaForge/engine/`。

## 架构

详见 `CLAUDE.md` 和 `AGENTS.md`，详细文档在 `docs/`：
- `docs/adr/` — 架构决策记录
- `docs/api/controller-contract.md` — 后端接口合约
- `docs/功能使用说明.md` — 用户手册
- `docs/开发环境搭建指南.md` — 环境配置步骤
