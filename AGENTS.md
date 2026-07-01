# AGENTS.md — MediaForge

See `CLAUDE.md` for full architecture (stack, entrypoints, controller/pkg boundary, FFmpeg embedding, frontend layout).

More detailed documentation in `docs/`:
- `docs/adr/` — 架构决策记录（为什么选 Wails、Pinia 等）
- `docs/api/controller-contract.md` — 后端接口合约（Go→前端的方法签名）
- `docs/功能需求文档.md` — 功能需求规格
- `docs/功能使用说明.md` — 功能使用文档
- `docs/开发环境搭建指南.md` — 环境搭建步骤

## Build/run

```bash
wails dev          # hot-reload Go+Vite
wails build        # production build → build/bin/mediaforge.exe
```

Frontend-only (in `frontend/`):
```bash
npm run dev        # Vite dev server
npm run build      # vue-tsc --noEmit + vite build
```

## File conventions

- Go: Chinese comments in controller layer, mixed EN/ZH comments in pkg/
- Vue: `<script setup lang="ts">` SFC with Pinia stores, component splitting in `frontend/src/components/`
- CSS: `frontend/src/styles.css` (extracted from App.vue)
- Types: Shared interfaces in `frontend/src/types.ts` (mirrors Go structs)
- State: Pinia stores in `frontend/src/stores/` (renamer, media, subtitle)
- Modal: Composable in `frontend/src/composables/useModal.ts`

## Key gotchas

- **No tests** exist anywhere in the repo — do not assume any test framework.
- **No linter/formatter** config found — skip lint/typecheck commands unless explicitly asked. `.golangci.yml` exists for Go but no CI pipeline runs it.
- **`pkg/converter/` is empty** — do not create files there.
- **`frontend/src/components/` contains the extracted panels** (RenamerPanel, FFmpegPanel, SubtitlePanel, MediaInfoPanel, ModalOverlay).
- **FFmpeg & FFprobe** are embedded via `//go:embed bin/*` and extracted to `%APPDATA%/MediaForge/engine/ffmpeg.exe` and `%APPDATA%/MediaForge/engine/ffprobe.exe` at `init()` time. Both binaries must exist in `pkg/media/bin/` before build.
- **`WARN_SUBTITLE`** is a special error string from Go controllers that the frontend catches to show a modal about incompatible subtitles. `ForceDropSubtitle: true` skips the check.
- **Event convention**: `ffmpeg-progress-{id}`, `ffmpeg-done-{id}`, `ffmpeg-error-{id}` — used for progress reporting from Go to frontend.
- **Go 1.23** is required (see `go.mod`).
