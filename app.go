package main

import (
	"context"
	"fmt"
	"mediaforge/pkg/renamer"

	// 引入 Wails 的 runtime 包，专门负责调用系统底层 API
	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type App struct {
	ctx context.Context
}

func NewApp() *App {
	return &App{}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) Run(ctx context.Context) {
	a.ctx = ctx
}

func (a *App) SelectFiles() []string {
	selection, err := runtime.OpenMultipleFilesDialog(a.ctx, runtime.OpenDialogOptions{
		Title: "请选择要重命名的文件",
		Filters: []runtime.FileFilter{
			{DisplayName: "所有文件 (*.*)", Pattern: "*.*"},
			{DisplayName: "视频文件 (*.mp4;*.mkv)", Pattern: "*.mp4;*.mkv"},
		},
	})

	if err != nil {
		fmt.Printf("选择文件时出错", err)
		return nil
	}

	fmt.Printf("选择的文件是：%v", selection)

	return selection
}

func (a *App) Greet(name string) string {
	if name == "" {
		return "请输入你的名字！"
	}
	return fmt.Sprintf("你好，%s！欢迎使用 MediaForge！", name)
}

func (a *App) PreviewRename(Paths []string, rule renamer.RenameRule) []renamer.RenamePreview {
	previews, err := renamer.GeneratePreview(Paths, rule)

	if err != nil {
		fmt.Printf("生成预览时出错", err)
		return nil
	}
	return previews
}

func (a *App) ApplyRename(Previews []renamer.RenamePreview) string {
	err := renamer.ExecuteRename(Previews)
	if err != nil {
		return fmt.Sprintf("重命名时出错: %v", err)
	}
	return "success"
}

// QuickRename 单文件行内快速重命名
func (a *App) QuickRename(oldPath string, newName string) (string, error) {
	return renamer.QuickRenameOnDisk(oldPath, newName)
}
