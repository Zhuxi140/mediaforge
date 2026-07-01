package media

import (
	"encoding/json"
	"fmt"
	"math"
	"os/exec"
	"path/filepath"
	"regexp"
	"strconv"
	"strings"
)

type MediaInfo struct {
	FilePath string         `json:"filePath"`
	FileSize string         `json:"fileSize"`
	Format   string         `json:"format"`
	Duration string         `json:"duration"`
	BitRate  string         `json:"bitRate"`
	Video    []VideoStream  `json:"video"`
	Audio    []AudioStream  `json:"audio"`
	Subtitle []SubtitleStream `json:"subtitle"`
}

type VideoStream struct {
	Index       int    `json:"index"`
	Codec       string `json:"codec"`
	Width       int    `json:"width"`
	Height      int    `json:"height"`
	FrameRate   string `json:"frameRate"`
	BitRate     string `json:"bitRate"`
	PixelFormat string `json:"pixelFormat"`
	Profile     string `json:"profile"`
}

type AudioStream struct {
	Index      int    `json:"index"`
	Codec      string `json:"codec"`
	SampleRate int    `json:"sampleRate"`
	Channels   int    `json:"channels"`
	BitRate    string `json:"bitRate"`
	Language   string `json:"language"`
}

type ffprobeFormat struct {
	Filename   string `json:"filename"`
	NbStreams  int    `json:"nb_streams"`
	FormatName string `json:"format_name"`
	Duration   string `json:"duration"`
	Size       string `json:"size"`
	BitRate    string `json:"bit_rate"`
	Tags       struct {
		MajorBrand string `json:"major_brand"`
	} `json:"tags"`
}

type ffprobeStream struct {
	Index        int               `json:"index"`
	CodecName    string            `json:"codec_name"`
	CodecType    string            `json:"codec_type"`
	Profile      string            `json:"profile"`
	Width        int               `json:"width"`
	Height       int               `json:"height"`
	PixFmt       string            `json:"pix_fmt"`
	RFrameRate   string            `json:"r_frame_rate"`
	AvgFrameRate string            `json:"avg_frame_rate"`
	BitRate      string            `json:"bit_rate"`
	SampleRate   string            `json:"sample_rate"`
	Channels     int               `json:"channels"`
	Tags         map[string]string `json:"tags"`
}

type ffprobeOutput struct {
	Streams []ffprobeStream `json:"streams"`
	Format  ffprobeFormat   `json:"format"`
}

func GetMediaInfo(inputPath string) (*MediaInfo, error) {
	if localFFprobePath == "" {
		return parseFromFFmpeg(inputPath)
	}

	args := []string{
		"-v", "quiet",
		"-print_format", "json",
		"-show_format",
		"-show_streams",
		inputPath,
	}

	cmd := exec.Command(localFFprobePath, args...)
	out, err := cmd.Output()
	if err != nil {
		return nil, fmt.Errorf("ffprobe 执行失败: %v", err)
	}

	var probe ffprobeOutput
	if err := json.Unmarshal(out, &probe); err != nil {
		return nil, fmt.Errorf("ffprobe 输出解析失败: %v", err)
	}

	info := &MediaInfo{
		FilePath: inputPath,
		FileSize: computeFileSize(probe.Format.Size, probe.Format.BitRate, probe.Format.Duration),
		Format:   cleanFormatName(probe.Format.FormatName, probe.Format.Tags.MajorBrand, inputPath),
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
		case "subtitle":
			lang := s.Tags["language"]
			if lang == "" {
				lang = "未知语言"
			}
			info.Subtitle = append(info.Subtitle, SubtitleStream{
				Index:    strconv.Itoa(s.Index),
				Language: lang,
				Codec:    s.CodecName,
			})
		}
	}

	return info, nil
}

func getStreamBitRate(s ffprobeStream) string {
	if s.BitRate != "" {
		return formatBitRate(s.BitRate)
	}
	for k, v := range s.Tags {
		if strings.HasPrefix(k, "BPS-") {
			if br := formatBitRate(v); br != "" {
				return br
			}
		}
	}
	return ""
}

func computeFileSize(sizeStr, bitRateStr, durationStr string) string {
	n, err := strconv.ParseInt(sizeStr, 10, 64)
	if err == nil && n > 0 {
		return formatSize(sizeStr)
	}
	br, err1 := strconv.ParseFloat(bitRateStr, 64)
	dur, err2 := strconv.ParseFloat(durationStr, 64)
	if err1 == nil && err2 == nil && br > 0 && dur > 0 {
		return formatSize(strconv.FormatInt(int64(math.Round(br/8*dur)), 10))
	}
	return sizeStr
}

func formatSize(sizeStr string) string {
	size, err := strconv.ParseInt(sizeStr, 10, 64)
	if err != nil || size <= 0 {
		return sizeStr
	}
	const unit = 1024
	if size < unit {
		return fmt.Sprintf("%d B", size)
	}
	div, exp := int64(unit), 0
	for n := size / unit; n >= unit; n /= unit {
		div *= unit
		exp++
	}
	return fmt.Sprintf("%.1f %cB", float64(size)/float64(div), "KMGTPE"[exp])
}

func formatDuration(durStr string) string {
	sec, err := strconv.ParseFloat(durStr, 64)
	if err != nil || sec <= 0 {
		return durStr
	}
	h := int(sec) / 3600
	m := (int(sec) % 3600) / 60
	s := int(sec) % 60
	ms := int(math.Round((sec - float64(int(sec))) * 1000))
	if h > 0 {
		return fmt.Sprintf("%d:%02d:%02d.%03d", h, m, s, ms)
	}
	if m > 0 {
		return fmt.Sprintf("%d:%02d.%03d", m, s, ms)
	}
	return fmt.Sprintf("%d.%03ds", s, ms)
}

func formatBitRate(brStr string) string {
	br, err := strconv.ParseInt(brStr, 10, 64)
	if err != nil || br <= 0 {
		return brStr
	}
	if br < 1000 {
		return fmt.Sprintf("%d bps", br)
	}
	if br < 1000000 {
		return fmt.Sprintf("%d Kbps", br/1000)
	}
	return fmt.Sprintf("%.1f Mbps", float64(br)/1000000)
}

func formatFrameRate(fps string) string {
	if !strings.Contains(fps, "/") {
		return fps
	}
	parts := strings.Split(fps, "/")
	if len(parts) != 2 {
		return fps
	}
	num, err1 := strconv.ParseFloat(parts[0], 64)
	den, err2 := strconv.ParseFloat(parts[1], 64)
	if err1 != nil || err2 != nil || den == 0 {
		return fps
	}
	return fmt.Sprintf("%.3f fps", num/den)
}

func cleanFormatName(raw string, majorBrand string, filePath string) string {
	majorBrand = strings.TrimSpace(majorBrand)
	if majorBrand != "" {
		switch majorBrand {
		case "isom", "mp42", "mp41", "3gp4", "3gp5":
			// Major Brand 是纯 MP4 系列
			return "MP4"
		case "qt  ":
			// QuickTime MOV
			return "MOV"
		}
	}
	parts := strings.Split(raw, ",")
	first := strings.TrimSpace(parts[0])
	if first != "" {
		return strings.ToUpper(first)
	}
	ext := strings.TrimPrefix(filepath.Ext(filePath), ".")
	if ext != "" {
		return strings.ToUpper(ext)
	}
	return "未知"
}

func parseSampleRate(sr string) int {
	n, _ := strconv.Atoi(sr)
	return n
}

var reFFmpegDuration = regexp.MustCompile(`Duration: (\d{2}:\d{2}:\d{2}\.\d{2})`)
var reFFmpegBitrate = regexp.MustCompile(`bitrate: (\d+ kb/s)`)
var reFFmpegVideo = regexp.MustCompile(`Stream #0:(\d+)(?:\(.*?\))?.*?: Video: (\w+)`)
var reFFmpegAudio = regexp.MustCompile(`Stream #0:(\d+)(?:\(.*?\))?.*?: Audio: (\w+)`)

func parseFromFFmpeg(inputPath string) (*MediaInfo, error) {
	if localFFmpegPath == "" {
		return nil, fmt.Errorf("引擎未初始化")
	}

	cmd := exec.Command(localFFmpegPath, "-i", inputPath)
	out, _ := cmd.CombinedOutput()
	output := string(out)

	info := &MediaInfo{
		FilePath: inputPath,
		Video:    []VideoStream{},
		Audio:    []AudioStream{},
	}

	lines := strings.Split(output, "\n")
	for _, line := range lines {
		if m := reFFmpegDuration.FindStringSubmatch(line); len(m) > 1 {
			info.Duration = m[1]
			continue
		}
		if m := reFFmpegBitrate.FindStringSubmatch(line); len(m) > 1 {
			info.BitRate = m[1]
			continue
		}
		if strings.Contains(line, "Video:") {
			if m := reFFmpegVideo.FindStringSubmatch(line); len(m) > 2 {
				codec := strings.Split(m[2], " ")[0]
				codec = strings.Split(codec, "(")[0]
				info.Video = append(info.Video, VideoStream{
					Index: parseInt(m[1]),
					Codec: codec,
				})
			}
		}
		if strings.Contains(line, "Audio:") {
			if m := reFFmpegAudio.FindStringSubmatch(line); len(m) > 2 {
				info.Audio = append(info.Audio, AudioStream{
					Index: parseInt(m[1]),
					Codec: strings.Split(m[2], " ")[0],
				})
			}
		}
	}

	return info, nil
}

func parseInt(s string) int {
	n, _ := strconv.Atoi(s)
	return n
}
