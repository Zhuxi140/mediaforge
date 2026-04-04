package media

import "strings"

type Quality string

const (
	High   = "High"
	Medium = "Medium"
	Low    = "Low"
	custom = "Custom"
)

type FFmpegTask struct {
	ID                string  `json:"ID"`
	InputPath         string  `json:"InputPath"`
	OutputPath        string  `json:"OutputPath"`
	Format            string  `json:"Format"`
	Quality           Quality `json:"Quality"`
	VideoCRF          string  `json:"VideoCRF"`
	AudioVBR          string  `json:"AudioVBR"`
	HwAccel           string  `json:"HwAccel"`
	ForceDropSubtitle bool    `json:"ForceDropSubtitle"`
}

type SubtitleStream struct {
	//流索引
	Index string `json:"Index"`
	// 语言
	Language string `json:"Language"`
	// 编码格式
	Codec string `json:"Codec"`
}

// 常见视频格式
var VideoExts = map[string]struct{}{
	"mp4": {}, "mkv": {},
	"avi": {}, "mov": {},
	"wmv": {}, "flv": {},
	"webm": {}, "mpeg": {},
	"mpg": {}, "3gp": {},
	"m2ts": {}, "mts": {},
	"rmvb": {},
}

var ValidAudioFormats = map[string]struct{}{
	"mp3": {}, "wav": {},
	"aac": {}, "flac": {},
	"ogg": {}, "m4a": {},
	"wma": {},
}

// 判断文件是否为常见视频文件
func IsVideoFile(ext string) bool {

	lower := strings.ToLower(strings.TrimSpace(ext))
	_, exists := VideoExts[lower]
	return exists
}

// 判断文件是否为常见音频文件
func IsAudioFile(ext string) bool {

	lower := strings.ToLower(strings.TrimSpace(ext))
	_, exists := ValidAudioFormats[lower]
	return exists
}
