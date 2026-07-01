package controller

import (
	"context"
	"fmt"
	"mediaforge/pkg/renamer"
	// 引入 Wails 的 runtime 包，专门负责调用系统底层 API
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type RenamerApp struct {
	ctx context.Context
}

func NewRenamerApp() *RenamerApp {
	return &RenamerApp{}
}

func (a *RenamerApp) Startup(ctx context.Context) {
	a.ctx = ctx
}

// SelectFiles 选择文件
func (a *RenamerApp) SelectFiles() []string {
	selection, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "请选择要重命名的文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
			{DisplayName: "视频文件 (*.mp4;*.mkv)", Pattern: "*.mp4;*.mkv"},
			{DisplayName: "音频文件 (*.mp3;*.wav)", Pattern: "*.mp3;*.wav"},
			{DisplayName: "图片文件 (*.jpg;*.png)", Pattern: "*.jpg;*.png"},
			{DisplayName: "字幕文件 (*.srt;*.ass)", Pattern: "*.srt;*.ass"},
		},
	})

	if err != nil {
		fmt.Printf("选择文件时出错: %v", err)
		return nil
	}

	fmt.Printf("选择的文件是：%v", selection)

	return selection
}

// PreviewRename 生成预览
func (a *RenamerApp) PreviewRename(Paths []string, rule renamer.RenameRule) []renamer.RenamePreview {
	previews, err := renamer.GeneratePreview(Paths, rule)

	if err != nil {
		fmt.Printf("生成预览时出错: %v", err)
		return nil
	}
	return previews
}

// ApplyRename 执行重命名
func (a *RenamerApp) ApplyRename(Previews []renamer.RenamePreview) string {
	err := renamer.ExecuteRename(Previews)
	if err != nil {
		return fmt.Sprintf("重命名时出错: %v", err)
	}
	return "success"
}

// QuickRename 单文件行内快速重命名
func (a *RenamerApp) QuickRename(oldPath string, newName string) (string, error) {
	return renamer.QuickRenameOnDisk(oldPath, newName)
}
