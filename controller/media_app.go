package controller

import (
	"context"
	"mediaforge/pkg/config"
	"mediaforge/pkg/media"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
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
	return media.ConvertSubtitle(inputPath, outDir, outPutName, targetFormat)
}

// SelectMediaFile 选择单个媒体文件
func (a *MediaApp) SelectMediaFile() string {
	selection, err := runtime.OpenFileDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "请选择媒体文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "所有媒体文件", Pattern: "*.mp4;*.mkv;*.avi;*.mov;*.wmv;*.flv;*.webm;*.mpeg;*.mpg;*.3gp;*.m2ts;*.mts;*.rmvb;*.mp3;*.wav;*.aac;*.flac;*.ogg;*.m4a;*.wma;*.gif"},
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
		},
	})
	if err != nil {
		return ""
	}
	return selection
}

// GetMediaInfo 获取媒体文件完整元数据
func (a *MediaApp) GetMediaInfo(inputPath string) *media.MediaInfo {
	info, err := media.GetMediaInfo(inputPath)
	if err != nil {
		runtime.LogError(a.ctx, "获取元数据失败: "+err.Error())
		return nil
	}
	return info
}

// LoadSettings 加载用户配置
func (a *MediaApp) LoadSettings() *config.Settings {
	s, err := config.Load()
	if err != nil {
		return &config.Settings{}
	}
	return s
}

// SaveSettings 保存用户配置
func (a *MediaApp) SaveSettings(s *config.Settings) error {
	return s.Save()
}
