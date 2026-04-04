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

/*var localFFprobePath string*/

// 记录到用户电脑上的真实物理路径
var localFFmpegPath string

func init() {
	if err := setupEmbeddedFFmpeg(); err != nil {
		fmt.Printf("初始化ffmpeg时出错: %v\n", err)
	}
}

// 释放FFmpeg引擎
func setupEmbeddedFFmpeg() error {
	// 获取系统配置目录
	configDir, err := os.UserConfigDir()
	if err != nil {
		return fmt.Errorf("无法获取系统配置目录：%v", err)
	}

	//为软件建立专属缓存目录
	appDir := filepath.Join(configDir, "MediaForge", "engine")
	if err := os.MkdirAll(appDir, 0755); err != nil {
		return fmt.Errorf("无法创建软件缓存目录：%v", err)
	}

	localFFmpegPath = filepath.Join(appDir, "ffmpeg.exe")
	/*	localFFprobePath = filepath.Join(appDir, "ffprobe.exe")*/

	// 如果文件已经存在，直接跳过释放进程
	if _, err := os.Stat(localFFmpegPath); err == nil {
		fmt.Printf("已存在ffmpeg文件，跳过释放进程")
		return nil
	}

	// 读取内嵌的二进制数据并写入硬盘
	fmt.Println("首次运行.... 正在向系统释放FFmpeg引擎")
	exeData, err := embeddedFiles.Open("bin/ffmpeg.exe")
	if err != nil {
		return fmt.Errorf("找不到内嵌的ffmpeg.exe(请检查打包前bin目录下是否有该文件): %v", err)
	}

	defer exeData.Close()

	//
	outFile1, err := os.OpenFile(localFFmpegPath, os.O_CREATE|os.O_WRONLY, 0755)
	if err != nil {
		return fmt.Errorf("创建ffmpeg底层文件：%v", err)
	}
	/*	outFile2, err := os.OpenFile(localFFprobePath, os.O_CREATE|os.O_WRONLY, 0755)
		if err != nil {
			return fmt.Errorf("创建ffprobe底层文件：%v", err)
		}*/

	defer outFile1.Close()

	if _, err := io.Copy(outFile1, exeData); err != nil {
		return fmt.Errorf("写入引擎底层文件失败：%v", err)
	}

	fmt.Println("释放完成")
	return nil
}
