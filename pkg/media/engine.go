package media

import (
	"embed"
	"fmt"
	"io"
	"os"
	"path/filepath"
)

//go:embed bin/*
var embeddedFiles embed.FS

var localFFmpegPath string
var localFFprobePath string

func init() {
	if err := setupEmbeddedBinaries(); err != nil {
		fmt.Printf("初始化引擎时出错: %v\n", err)
	}
}

func setupEmbeddedBinaries() error {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("无法获取系统配置目录：%v", err)
	}

	appDir := filepath.Join(configDir, "MediaForge", "engine")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("无法创建引擎目录：%v", err)
	}

	localFFmpegPath = filepath.Join(appDir, "ffmpeg.exe")
	localFFprobePath = filepath.Join(appDir, "ffprobe.exe")

	if err := extractBinary("bin/ffmpeg.exe", localFFmpegPath); err != nil {
		return err
	}
	if err := extractBinary("bin/ffprobe.exe", localFFprobePath); err != nil {
		return err
	}

	return nil
}

func extractBinary(embeddedPath, targetPath string) error {
	if _, err := os.Stat(targetPath); err == nil {
		return nil
	}

	fmt.Printf("正在释放 %s ...\n", targetPath)
	data, err := embeddedFiles.Open(embeddedPath)
	if err != nil {
		return fmt.Errorf("找不到内嵌文件 %s: %v", embeddedPath, err)
	}
	defer data.Close()

	outFile, err := os.OpenFile(targetPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("创建文件 %s 失败：%v", targetPath, err)
	}
	defer outFile.Close()

	if _, err := io.Copy(outFile, data); err != nil {
		return fmt.Errorf("写入文件 %s 失败：%v", targetPath, err)
	}

	fmt.Printf("释放完成: %s\n", targetPath)
	return nil
}
