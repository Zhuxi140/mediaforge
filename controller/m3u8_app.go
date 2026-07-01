package controller

import (
	"context"
	"mediaforge/pkg/media"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type M3U8App struct {
	ctx context.Context
}

func NewM3U8App() *M3U8App {
	return &M3U8App{}
}

func (a *M3U8App) Startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *M3U8App) DownloadM3U8(id, url, outputDir string) error {
	return media.DownloadM3U8(a.ctx, id, url, outputDir)
}

func (a *M3U8App) CancelDownload(id string) {
	media.CancelM3U8Task(id)
}

func (a *M3U8App) ProbeM3U8URL(url string) (*media.MediaInfo, error) {
	return media.ProbeM3U8URL(url)
}

func (a *M3U8App) SelectOutputDir() string {
	selection, err := runtime.OpenDirectoryDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "选择下载目录",
	})
	if err != nil {
		return ""
	}
	return selection
}
