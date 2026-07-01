# ADR-002: 嵌入式 FFmpeg 方案

**状态**: ✅ 已采纳  
**日期**: 2026-06-29  

## 背景

应用核心功能依赖 FFmpeg 进行音视频转码和字幕处理。需要考虑 FFmpeg 的分发方式。

## 可选方案

1. **系统安装 FFmpeg** — 用户自行安装并配置 PATH，应用动态查找
2. **捆绑 FFmpeg** — 将静态构建的 ffmpeg.exe 嵌入应用二进制，首次运行释放
3. **Go 原生库** — 使用 `goav` 等库直接调用 FFmpeg API

## 决策

选择方案 **2（捆绑 FFmpeg）**。

## 理由

- 用户无需额外安装，开箱即用
- `//go:embed bin/*` 机制简洁，编译时将二进制嵌入
- 版本可控，避免系统 FFmpeg 版本不兼容
- 释放路径为 `%APPDATA%/MediaForge/engine/ffmpeg.exe`，符合 Windows 惯例
- 提取一次后后续启动跳过（`os.Stat` 检查）

## 后果

- 仓库中 `pkg/media/bin/ffmpeg.exe` 较大（~150MB），建议使用 Git LFS
- 不支持多平台分发（当前仅 Windows x64）
- 更新 FFmpeg 版本需要手动替换二进制文件 + 重新编译
- macOS/Linux 需打包对应平台二进制
