package config

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// BookConfig 书籍配置结构体（包含元数据和进度）
type BookConfig struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Author           string `json:"author"`
	OriginalFileName string `json:"originalFileName"`
	MD5              string `json:"md5"`
	CoverPath        string `json:"coverPath"`
	CreatedAt        int64  `json:"createdAt"`
	LastCFI          string `json:"last_cfi"`
	// 以下字段用于保留 EPUB 元数据
	Description string   `json:"description,omitempty"`
	Publisher   string   `json:"publisher,omitempty"`
	PublishDate string   `json:"publishDate,omitempty"`
	Language    string   `json:"language,omitempty"`
	Subjects    []string `json:"subjects,omitempty"`
}

// ProgressInfo 阅读进度信息
type ProgressInfo struct {
	CFI        string  `json:"cfi"`
	Percentage float64 `json:"percentage"`
	Timestamp  int64   `json:"timestamp"`
}

// SaveProgress 保存阅读进度到书籍同目录的 config.json（保留原有元数据）
func SaveProgress(filePath, progressJSON string) error {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	var config BookConfig
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	config.LastCFI = progressJSON

	prettyData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, prettyData, 0644)
}

// GetProgress 从书籍同目录的 config.json 读取阅读进度
func GetProgress(filePath string) string {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return ""
	}

	var config BookConfig
	json.Unmarshal(data, &config)
	return config.LastCFI
}
