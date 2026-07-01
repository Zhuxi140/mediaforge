package media

import (
	"bufio"
	"context"
	"fmt"
	"io"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"sync"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// 最大并发转码任务数
const maxConcurrentTasks = 3

var taskSem = make(chan struct{}, maxConcurrentTasks)

// 存储当前正在处理的任务
var activeTasks sync.Map

// 从日志中抓取时间并转码
var timeRegex = regexp.MustCompile(`time=(\d{2}:\d{2}:\d{2}.\d{2})`)

// 视频处理
func videoDeal(task FFmpegTask) []string {
	// 视频编码
	var args []string

	if task.ForceDropSubtitle {
		args = append(args, "-sn")
	}

	vCode := "libx264"
	qFlag := "-crf"

	switch task.HwAccel {
	case "nvidia":
		vCode = "h264_nvenc"
		qFlag = "-cq"
	case "intel":
		vCode = "h264_qsv"
		qFlag = "-q"
	case "amd":
		vCode = "h264_amf"
		qFlag = "-qp"
	case "mac":
		vCode = "h264_videotoolbox"
		qFlag = "-q:v"
	}

	args = append(args, "-c:v", vCode)
	if vCode == "libx264" {
		args = append(args, "-preset", "fast")
	}
	args = append(args, "-c:a", "aac")

	//视频质量控制
	switch task.Quality {
	case High:
		args = append(args, qFlag, "18")
	case Medium:
		args = append(args, qFlag, "23")
	case Low:
		args = append(args, qFlag, "28")
	case custom:
		args = append(args, qFlag, task.VideoCRF)
	}

	return args
}

// 音频处理
func audioDeal(task FFmpegTask, lowerFormat string) []string {

	var args []string

	args = append(args, "-vn")
	switch lowerFormat {
	case "mp3":
		args = append(args, "-c:a", "libmp3lame")
		switch task.Quality {
		case High:
			args = append(args, "-q:a", "0")
		case Medium:
			args = append(args, "-q:a", "2")
		case Low:
			args = append(args, "-q:a", "5")
		case custom:
			args = append(args, "-q:a", task.AudioVBR)
		}
	case "aac":
		args = append(args, "-c:a", "aac")
		switch task.Quality {
		case High:
			args = append(args, "-b:a", "320k")
		case Medium:
			args = append(args, "-b:a", "192k")
		case Low:
			args = append(args, "-b:a", "128k")
		case custom:
			args = append(args, "-b:a", task.AudioVBR+"k")
		}
	case "flac":
		args = append(args, "-c:a", "flac")
	case "wav":
		args = append(args, "-c:a", "pcm_s16le")
	}

	return args
}

// 执行格式转换，并用Wails的context向前端广播进度
func RunConvert(ctx context.Context, task FFmpegTask) error {

	if localFFmpegPath == "" {
		return fmt.Errorf("ffmpeg未初始化,无法执行任务")
	}

	// 组装命令
	args := []string{"-y", "-i", task.InputPath}

	lowerFormat := strings.ToLower(task.Format)

	if IsVideoFile(lowerFormat) {
		args = append(args, videoDeal(task)...)
	} else if IsAudioFile(lowerFormat) {
		args = append(args, audioDeal(task, lowerFormat)...)
	} else if lowerFormat == "gif" {
		args = append(args, "-vf", "fps=10,scale=480:-1:flags=lanczos", "-loop", "0")
	} else {
		return fmt.Errorf("暂不支持的格式： %s", lowerFormat)
	}

	args = append(args, task.OutputPath)

	cmd := exec.CommandContext(ctx, localFFmpegPath, args...)

	// 劫持错误流
	stderr, err := cmd.StderrPipe()
	if err != nil {
		return fmt.Errorf("无法获取错误流：%v", err)
	}

	//异步启动
	if err := cmd.Start(); err != nil {
		return fmt.Errorf("无法启动命令：%v", err)
	}

	// 取消时关闭 stderr 以解除 scanner 阻塞
	done := make(chan struct{})
	go func() {
		select {
		case <-ctx.Done():
			cmd.Process.Kill()
			stderr.Close()
		case <-done:
		}
	}()

	// 监听进度
	lastOutPut := MonitorProgress(stderr, ctx, task)
	close(done)

	if err := cmd.Wait(); err != nil {
		return fmt.Errorf("命令执行失败：%v \n 最后输出为: %s \n 执行命令为: %s", err, lastOutPut, cmd.Args)
	}

	runtime.EventsEmit(ctx, "ffmpeg-done-"+task.ID, "success")

	return nil
}

// 取消任务
func CancelTask(id string) {
	if cancelFunc, exist := activeTasks.Load(id); exist {
		cancelFunc.(context.CancelFunc)()
		activeTasks.Delete(id)
		fmt.Println("任务取消成功")
	}
}

// 自定义分割器
func scanCR(data []byte, atEOF bool) (advance int, token []byte, err error) {
	if atEOF && len(data) == 0 {
		return 0, nil, nil
	}

	if i := strings.IndexAny(string(data), "\r\n"); i >= 0 {
		return i + 1, data[0:i], nil
	}

	if atEOF {
		return len(data), data, nil
	}

	return 0, nil, nil
}

// 接受基础参数，组装业务实体，异步拉起底层转码任务
func ProcessMediaAsync(task FFmpegTask, crx context.Context) error {

	if !task.ForceDropSubtitle {
		hasSub, isCompatible := CheckSubtitleCompatibility(task.InputPath, task.Format)
		// 如果有字幕，且格式不兼容，才拦截报警
		if hasSub && !isCompatible {
			return fmt.Errorf("WARN_SUBTITLE")
		}
	}

	var finalOutPutPath string

	base := filepath.Base(task.InputPath)
	ext := filepath.Ext(base)
	nameWithoutExt := strings.TrimSuffix(base, ext)
	outName := fmt.Sprintf("%s_converted.%s", nameWithoutExt, task.Format)

	if task.OutputPath != "" {

		// 采用用户指定输出路径
		finalOutPutPath = filepath.Join(task.OutputPath, outName)

	} else {
		// 默认输出到同文件夹下
		dir := filepath.Dir(task.InputPath)
		finalOutPutPath = filepath.Join(dir, outName)
	}

	task.OutputPath = finalOutPutPath

	// 确保输出目录存在
	outDir := filepath.Dir(finalOutPutPath)
	if err := os.MkdirAll(outDir, 0755); err != nil {
		return fmt.Errorf("无法创建输出目录: %v", err)
	}

	ctx, cancel := context.WithCancel(crx)
	activeTasks.Store(task.ID, cancel)

	go func() {
		// 获取并发令牌，达到上限时阻塞等待
		taskSem <- struct{}{}

		defer func() {
			<-taskSem // 释放令牌
		}()

		// 任务结束后删除任务
		defer activeTasks.Delete(task.ID)

		err := RunConvert(ctx, task)
		if err != nil {
			//如果用户手动取消,提示并关闭任务
			if ctx.Err() == context.Canceled {
				runtime.EventsEmit(crx, "ffmpeg-error-"+task.ID, "任务已取消")
			} else {
				runtime.LogError(crx, fmt.Sprintf("任务执行失败: %v", err))
				runtime.EventsEmit(crx, "ffmpeg-error-"+task.ID, err.Error())
			}
		}
	}()

	return nil

}

// 监视进度
func MonitorProgress(stderr io.ReadCloser, ctx context.Context, task FFmpegTask) string {
	// 实时监控ffmpeg日志
	scanner := bufio.NewScanner(stderr)
	// 使用自定义分割器
	scanner.Split(scanCR)

	// 记录FFmpeg的 最后输出  用于转换失败或出错时的错误展示信息
	var lastOutPut string

	// 读取日志 捕获时间
	for scanner.Scan() {
		line := scanner.Text()

		if strings.Contains(line, "time=") {
			matches := timeRegex.FindStringSubmatch(line)
			if len(matches) > 1 {
				currentTime := matches[1]
				runtime.EventsEmit(ctx, "ffmpeg-progress-"+task.ID, currentTime)
			}

		} else if strings.TrimSpace(line) != "" {
			lastOutPut = line
		}
	}
	return lastOutPut
}
