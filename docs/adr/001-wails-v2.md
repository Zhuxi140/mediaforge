# ADR-001: Wails v2 作为桌面框架

**状态**: ✅ 已采纳  
**日期**: 2026-06-29  

## 背景

需要构建一个跨平台桌面工具集，涉及 Go 后端（调用 FFmpeg、文件系统操作）和前端 UI。需选择桌面应用框架。

## 可选方案

1. **Electron** — 最成熟，但捆绑 Chromium 导致包体 ~150MB，内存占用高
2. **Tauri** — Rust 后端，需要学习 Rust，与现有 Go 生态不兼容
3. **Wails v2** — Go 后端 + Web 前端，包体小（~20MB），原生窗口
4. **原生 GUI** (WinForms/Qt) — 开发效率低，UI 灵活性差

## 决策

选择 **Wails v2**。

## 理由

- Go 后端直接调用 FFmpeg，无需 IPC 或子进程管理
- 前端可用 Vue 3 + TypeScript，开发体验好
- 内嵌二进制文件（FFmpeg）支持良好（`//go:embed`）
- 包体小，Windows 上单 exe 发布
- 项目已有 Go 生态积累

## 后果

- 前端与后端通过 Wails Bind 直接调用 Go 方法（类似 RPC）
- 前后端紧耦合（方法签名变更需同步更新前端）
- 调试需要 Wails dev server（http://localhost:34115）
- Windows 平台支持好，macOS/Linux 需额外适配
