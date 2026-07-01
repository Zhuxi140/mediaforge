package config

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"
)

type Settings struct {
	MediaOutputDir   string `json:"mediaOutputDir"`
	MediaHwAccel     string `json:"mediaHwAccel"`
	MediaQuality     string `json:"mediaQuality"`
	MediaVideoCRF    string `json:"mediaVideoCRF"`
	MediaAudioVBR    string `json:"mediaAudioVBR"`
	MediaTargetFmt   string `json:"mediaTargetFmt"`
	SubOutputDir     string `json:"subOutputDir"`
	SubFormat        string `json:"subFormat"`
	SubConvertDir    string `json:"subConvertDir"`
	SubConvertFmt    string `json:"subConvertFmt"`
}

func configPath() string {
	configDir, err := os.UserConfigDir()
	if err != nil {
		return ""
	}
	appDir := filepath.Join(configDir, "MediaForge")
	return filepath.Join(appDir, "config.json")
}

func Load() (*Settings, error) {
	path := configPath()
	if path == "" {
		return nil, fmt.Errorf("无法获取用户配置目录")
	}
	data, err := os.ReadFile(path)
	if err != nil {
		if os.IsNotExist(err) {
			return &Settings{}, nil
		}
		return nil, err
	}
	var s Settings
	if err := json.Unmarshal(data, &s); err != nil {
		return &Settings{}, nil
	}
	return &s, nil
}

func (s *Settings) Save() error {
	path := configPath()
	if path == "" {
		return fmt.Errorf("无法获取用户配置目录")
	}
	dir := filepath.Dir(path)
	if err := os.MkdirAll(dir, 0755); err != nil {
		return err
	}
	data, err := json.MarshalIndent(s, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, data, 0644)
}
