package controller

import (
	"context"
	"mediaforge/pkg/media"
	"path/filepath"
	"strings"
)

type MediaApp struct {
	ctx context.Context
}

func NewMediaApp() *MediaApp {
	return &MediaApp{}
}

func (m *MediaApp) Startup(ctx context.Context) {
	m.ctx = ctx
}

// SubmitMediaTask 提交任务
func (a *MediaApp) SubmitMediaTask(task media.FFmpegTask) error {
	return media.ProcessMediaAsync(task, a.ctx)
}

// CancelMediaTask 取消任务
func (a *MediaApp) CancelMediaTask(id string) {
	media.CancelTask(id)
}

// ScanSubtitles 扫描视频字幕
func (a *MediaApp) ScanSubtitles(inputPath string) ([]media.SubtitleStream, error) {
	return media.GetSubtitleStreams(inputPath)
}

// ExtractSubtitle 提取特定字幕
func (a *MediaApp) ExtractSubtitle(id, inputPath, streamIndex, outDir, targetFormat string) error {
	return media.ProcessSubtitleAsync(a.ctx, id, inputPath, streamIndex, outDir, targetFormat)
}

// CheckInputFile 检查输入文件是否是媒体、音频、图片文件
func (a *MediaApp) CheckInputFile(filePath string) string {
	ext := filepath.Ext(filePath)
	if ext == "" {
		return "未知格式文件"
	}

	cleanExt := strings.ToLower(strings.TrimPrefix(ext, "."))

	isValid := media.IsVideoFile(cleanExt) || media.IsAudioFile(cleanExt) || cleanExt == "gif"

	if !isValid {
		return "暂不支持的格式: " + strings.ToUpper(cleanExt)
	}

	return "success"
}

func (a *MediaApp) ConvertSubtitle(inputPath, outDir, outPutName, targetFormat string) error {
	return media.ConvertSubtitle(inputPath, outDir, targetFormat, outPutName)
}
