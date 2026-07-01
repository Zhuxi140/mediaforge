package media

import (
	"bufio"
	"context"
	"fmt"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

var speedRegex = regexp.MustCompile(`speed=\s*([\d.]+)x`)
var durationRegex = regexp.MustCompile(`Duration:\s*(\d{2}:\d{2}:\d{2}\.\d{2})`)

var m3u8Processes sync.Map

const m3u8OutputName = "m3u8_download"

func CancelM3U8Task(id string) {
	if proc, exist := m3u8Processes.Load(id); exist {
		if p, ok := proc.(*os.Process); ok {
			p.Kill()
		}
		m3u8Processes.Delete(id)
	}
	CancelTask(id)
}

func DownloadM3U8(crx context.Context, id, url, outputDir string) error {
	if localFFmpegPath == "" {
		return fmt.Errorf("引擎未初始化")
	}

	finalOutDir := outputDir
	if finalOutDir == "" {
		home, err := os.UserHomeDir()
		if err != nil {
			return fmt.Errorf("无法获取用户目录: %v", err)
		}
		finalOutDir = filepath.Join(home, "Downloads")
	}

	if err := os.MkdirAll(finalOutDir, 0755); err != nil {
		return fmt.Errorf("无法创建输出目录: %v", err)
	}

	outputPath := filepath.Join(finalOutDir, fmt.Sprintf("%s_%s.mp4", m3u8OutputName, time.Now().Format("20060102_150405")))

	ctx, cancel := context.WithCancel(crx)
	activeTasks.Store(id, cancel)

	go func() {
		defer activeTasks.Delete(id)
		defer m3u8Processes.Delete(id)

		args := []string{
			"-y", "-i", url,
			"-c", "copy",
			"-bsf:a", "aac_adtstoasc",
			outputPath,
		}

		cmd := exec.Command(localFFmpegPath, args...)
		stderr, err := cmd.StderrPipe()
		if err != nil {
			runtime.EventsEmit(crx, "m3u8-error-"+id, fmt.Sprintf("无法获取错误流: %v", err))
			return
		}

		if err := cmd.Start(); err != nil {
			runtime.EventsEmit(crx, "m3u8-error-"+id, fmt.Sprintf("无法启动命令: %v", err))
			return
		}

		m3u8Processes.Store(id, cmd.Process)

		done := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				cmd.Process.Kill()
				stderr.Close()
			case <-done:
			}
		}()
		defer close(done)

		var totalDuration string

		scanner := bufio.NewScanner(stderr)
		scanner.Split(scanCR)
		for scanner.Scan() {
			line := strings.TrimSpace(scanner.Text())
			if line == "" {
				continue
			}

			// 捕获总时长
			if totalDuration == "" {
				if dm := durationRegex.FindStringSubmatch(line); len(dm) > 1 {
					totalDuration = strings.TrimRight(dm[1], ".")
				}
			}

			speedMatch := speedRegex.FindStringSubmatch(line)
			timeMatch := timeRegex.FindStringSubmatch(line)
			if len(speedMatch) == 0 && len(timeMatch) == 0 {
				continue
			}

			var parts []string
			if totalDuration != "" {
				parts = append(parts, "总时长 "+totalDuration)
			}
			if len(speedMatch) > 1 {
				parts = append(parts, speedMatch[1]+"x")
			}
			if len(timeMatch) > 1 {
				parts = append(parts, timeMatch[1])
			}

			runtime.EventsEmit(crx, "m3u8-progress-"+id, strings.Join(parts, " | "))
		}

		if err := cmd.Wait(); err != nil {
			if ctx.Err() == context.Canceled {
				runtime.EventsEmit(crx, "m3u8-error-"+id, "任务已取消")
				os.Remove(outputPath)
			} else {
				runtime.LogError(crx, fmt.Sprintf("M3U8 下载失败: %v", err))
				runtime.EventsEmit(crx, "m3u8-error-"+id, err.Error())
			}
			return
		}

		runtime.EventsEmit(crx, "m3u8-done-"+id, outputPath)
	}()

	return nil
}
