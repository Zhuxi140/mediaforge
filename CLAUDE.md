# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Build/Run

```bash
wails dev          # Live dev (hot-reload frontend, Go backend with Vite dev server)
wails build        # Production build → build/bin/mediaforge.exe
```

Frontend-only (inside `frontend/`):
```bash
npm run dev        # Vite dev server
npm run build      # vue-tsc type-check + vite build
```

## Architecture

**Stack:** Wails v2 desktop app — Go backend + Vue 3 frontend (TypeScript, Vite). Single `App.vue` SFC, no component splitting.

**Entry:** `main.go` — creates `RenamerApp` and `MediaApp` controllers, binds both to Wails frontend via `Bind: []interface{}{renamerApp, mediaApp}`.

**Controller layer** (`controller/`): Thin Wails binding. Each controller holds a `context.Context` from `Startup()`. Methods are exposed to the frontend as RPC calls via Go struct method binding. `RenamerApp` handles file selection dialogs and delegates to `pkg/renamer`. `MediaApp` delegates to `pkg/media`.

**Core packages** (no Wails dependency):
- `pkg/renamer/` — Two rename modes: Basic (prefix/suffix/string-replace) and Smart (regex extraction → template `{var}` assembly). Entry: `GeneratePreview()` → `ExecuteRename()`. Also `QuickRenameOnDisk()` for inline single-file edits.
- `pkg/media/engine.go` — Embeds `bin/ffmpeg.exe` via `//go:embed bin/*`, extracts to `%APPDATA%/MediaForge/engine/ffmpeg.exe` on first run (`init()` time). Path stored in `localFFmpegPath`.
- `pkg/media/transcode.go` — `ProcessMediaAsync()` validates, constructs FFmpeg args, launches goroutine with cancellable context. `RunConvert()` builds codec args (hwaccel: NVENC/QSV/AMF/VideoToolbox), monitors stderr for `time=` progress via Wails events (`ffmpeg-progress-{id}`, `ffmpeg-done-{id}`, `ffmpeg-error-{id}`). Active tasks stored in `sync.Map` for `CancelTask()`.
- `pkg/media/subtitle.go` — `CheckSubtitleCompatibility()` (warns on MP4 incompatible subs), `GetSubtitleStreams()` (regex-parses ffprobe output), `ProcessSubtitleAsync()` and `ConvertSubtitle()`, `selectCodec()` maps format→FFmpeg codec.
- `pkg/media/types.go` — `FFmpegTask` struct, `Quality` consts (High/Medium/Low/Custom), format extension maps (`VideoExts`, `ValidAudioFormats`, `ValidSubtitleFormats`).

**Frontend:** Single `App.vue` with three app modes toggled via sidebar nav (`activeApp`): Renamer, FFmpeg factory, Subtitle center. Calls Go controllers via auto-generated Wails bindings (`../wailsjs/go/controller/...`). Custom modal system for confirmations/warnings. File selection shared across all three modes.
