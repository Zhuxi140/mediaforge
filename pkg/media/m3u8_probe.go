package media

import (
	"context"
	"encoding/json"
	"fmt"
	"os/exec"
	"time"
)

func ProbeM3U8URL(url string) (*MediaInfo, error) {
	if localFFprobePath == "" {
		return nil, fmt.Errorf("ffprobe 未初始化")
	}

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		url,
	}

	cmd := exec.CommandContext(ctx, localFFprobePath, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("分析失败（链接无效或超时）: %v", err)
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(out, &probe); err != nil {
		return nil, fmt.Errorf("输出解析失败: %v", err)
	}

	info := &MediaInfo{
		FilePath: url,
		Format:   cleanFormatName(probe.Format.FormatName, probe.Format.Tags.MajorBrand, url),
		Duration: formatDuration(probe.Format.Duration),
		BitRate:  formatBitRate(probe.Format.BitRate),
		Video:    []VideoStream{},
		Audio:    []AudioStream{},
		Subtitle: []SubtitleStream{},
	}

	for _, s := range probe.Streams {
		switch s.CodecType {
		case "video":
			fps := s.RFrameRate
			if fps == "0/0" {
				fps = s.AvgFrameRate
			}
			info.Video = append(info.Video, VideoStream{
				Index:       s.Index,
				Codec:       s.CodecName,
				Width:       s.Width,
				Height:      s.Height,
				FrameRate:   formatFrameRate(fps),
				BitRate:     getStreamBitRate(s),
				PixelFormat: s.PixFmt,
				Profile:     s.Profile,
			})
		case "audio":
			lang := s.Tags["language"]
			if lang == "" {
				lang = "未知"
			}
			info.Audio = append(info.Audio, AudioStream{
				Index:      s.Index,
				Codec:      s.CodecName,
				SampleRate: parseSampleRate(s.SampleRate),
				Channels:   s.Channels,
				BitRate:    getStreamBitRate(s),
				Language:   lang,
			})
		}
	}

	return info, nil
}
