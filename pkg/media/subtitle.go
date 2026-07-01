package media

import (
	"context"
	"fmt"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

// 检测视频文件是否包含内封字幕流
// 返回值： hasSubtitle: 是否包含字幕流， isCompatible : 是否兼容
func CheckSubtitleCompatibility(inputPath string, targetFormat string) (bool, bool) {

	if localFFmpegPath == "" {
		return false, true
	}

	cmd := exec.Command(localFFmpegPath, "-i", inputPath)

	out, _ := cmd.CombinedOutput()

	HasSubtitle := false
	IsCompatible := true

	lines := strings.Split(string(out), "\n")

	for _, line := range lines {
		if strings.Contains(line, "Subtitle") {
			HasSubtitle = true

			if strings.ToLower(targetFormat) == "mp4" {

				//mp4只支持 mov_text(tx3g)
				if !strings.Contains(line, "mov_text") && !strings.Contains(line, "tx3g") {
					IsCompatible = false
				}

			}
		}

	}

	return HasSubtitle, IsCompatible
}

// 扫描视频、获取字幕流信息
func GetSubtitleStreams(inputPath string) ([]SubtitleStream, error) {
	if localFFmpegPath == "" {
		return nil, fmt.Errorf("ffmpeg未初始化")
	}

	ext := filepath.Ext(inputPath)

	if !IsVideoFile(strings.TrimPrefix(ext, ".")) {
		return nil, fmt.Errorf("该文件不是视频文件，或暂不支持该视频格式")
	}

	cmd := exec.Command(localFFmpegPath, "-i", inputPath)
	out, _ := cmd.CombinedOutput()

	var streams []SubtitleStream
	lines := strings.Split(string(out), "\n")

	re := regexp.MustCompile(`Stream #0:(\d+)(?:\((.*?)\))?.*?: Subtitle: (.*)`)

	for _, line := range lines {
		if strings.Contains(line, "Subtitle") {
			matches := re.FindStringSubmatch(line)
			if len(matches) >= 4 {
				codecInfo := strings.Split(strings.TrimSpace(matches[3]), " ")[0]

				lang := matches[2]
				if lang == "" {
					lang = "未知语言"
				}

				streams = append(streams, SubtitleStream{
					Index:    matches[1],
					Language: lang,
					Codec:    codecInfo,
				})
			}
		}
	}

	return streams, nil
}

// 异步提取指定的字幕流
func ProcessSubtitleAsync(crx context.Context, id, inputPath, streamIndex, outDir string, targetFormat string) error {
	base := filepath.Base(inputPath)
	ext := filepath.Ext(base)
	nameWioutExt := strings.TrimSuffix(base, ext)

	targetFormat = strings.TrimPrefix(targetFormat, ".")
	var codec string

	codec, targetFormat = selectCodec(targetFormat)

	outName := fmt.Sprintf("%s_sub_stream%s.%s", nameWioutExt, streamIndex, targetFormat)

	var finalOutPutPath string

	if outDir == "" {
		finalOutPutPath = filepath.Join(filepath.Dir(inputPath), outName)
	} else {
		finalOutPutPath = filepath.Join(outDir, outName)
	}

	ctx, cancel := context.WithCancel(crx)
	activeTasks.Store(id, cancel)

	go func() {
		defer activeTasks.Delete(id)

		args := []string{
			"-y", "-nostdin", "-i", inputPath,
			"-map", "0:" + streamIndex,
			"-c:s", codec,
			finalOutPutPath,
		}

		fmt.Printf("\n[字幕提取] ffmpeg %s\n", strings.Join(args, " "))
		cmd := exec.CommandContext(ctx, localFFmpegPath, args...)

		done := make(chan struct{})
		go func() {
			select {
			case <-ctx.Done():
				if cmd.Process != nil {
					cmd.Process.Kill()
				}
			case <-done:
			}
		}()

		out, err := cmd.CombinedOutput()
		close(done)

		if err != nil {
			if ctx.Err() == context.Canceled {
				runtime.EventsEmit(ctx, "ffmpeg-error-"+id, "任务已取消")
			} else {
				runtime.EventsEmit(ctx, "ffmpeg-error-"+id, fmt.Sprintf("提取失败：%v \n 日志信息: %s", err, string(out)))
			}
			return
		}

		runtime.EventsEmit(ctx, "ffmpeg-done-"+id, "success")
	}()
	return nil
}

func ConvertSubtitle(inputPath, outputDir, outPutName, targetFormat string) error {
	if localFFmpegPath == "" {
		return fmt.Errorf("引擎未初始化")
	}

	ext := filepath.Ext(inputPath)

	if !IsSubtitleFile(strings.TrimPrefix(ext, ".")) {
		return fmt.Errorf("该文件不是字幕文件或暂不支持该类型字幕: %s", ext)
	}

	base := filepath.Base(inputPath)
	var outName string
	if outPutName != "" {
		outName = outPutName
	} else {
		outName = strings.TrimSuffix(base, ext)
	}
	targetFormat = strings.ToLower(strings.TrimPrefix(targetFormat, "."))
	outName = fmt.Sprintf("%s.%s", outName, targetFormat)

	var finalOutPutPath string

	if outputDir == "" {
		finalOutPutPath = filepath.Join(filepath.Dir(inputPath), outName)
	} else {
		finalOutPutPath = filepath.Join(outputDir, outName)
	}

	var codec string

	codec, targetFormat = selectCodec(targetFormat)

	args := []string{
		"-y", "-nostdin", "-i", inputPath,
		"-c:s", codec,
		finalOutPutPath,
	}

	fmt.Printf("\n[字幕转换] ffmpeg %s\n", strings.Join(args, " "))
	cmd := exec.Command(localFFmpegPath, args...)

	out, err := cmd.CombinedOutput()

	if err != nil {
		return fmt.Errorf("命令执行失败：%v \n 执行命令为: %s \n ffmpeg输出信息: %s ", err, cmd.Args, string(out))
	}
	return nil
}

func selectCodec(targetFormat string) (string, string) {
	var codec string

	switch targetFormat {
	case "ass", "ssa":
		codec = "ass"
	case "vtt":
		codec = "webvtt"
	case "xml", "ttml", "dfxp":
		codec = "ttml"
	case "sami", "smi":
		codec = "sami"
	default:
		codec = "subrip"
		targetFormat = "srt"

	}
	return codec, targetFormat
}
