# MediaForge

基于 Wails v2 的桌面工具集：批量文件重命名、FFmpeg 影音转换、字幕提取与格式互转。

## 开发

```bash
wails dev          # 热重载开发
```

前端单独开发（`frontend/` 目录下）：
```bash
npm run dev        # Vite 开发服务器
npm run build      # 类型检查 + 生产构建
```

## 构建

```bash
wails build        # → build/bin/mediaforge.exe
```

FFmpeg 引擎通过 `//go:embed bin/*` 内嵌，首次运行自动释放到 `%APPDATA%/MediaForge/engine/`。

## 架构

详见 `CLAUDE.md` 和 `AGENTS.md`。
