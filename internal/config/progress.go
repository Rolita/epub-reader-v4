package config

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Bookmark 书签信息
type Bookmark struct {
	CFI          string  `json:"cfi"`
	Percentage   float64 `json:"percentage"`
	Timestamp    int64   `json:"timestamp"`
	ChapterTitle string  `json:"chapterTitle,omitempty"`
	Snippet      string  `json:"snippet,omitempty"`
}

// SearchHistoryItem 搜索历史项
type SearchHistoryItem struct {
	Keyword   string `json:"keyword"`
	Timestamp int64  `json:"timestamp"`
}

// Note 笔记信息
type Note struct {
	CFI          string  `json:"cfi"`
	Percentage   float64 `json:"percentage"`
	Timestamp    int64   `json:"timestamp"`
	ChapterTitle string  `json:"chapterTitle,omitempty"`
	Content      string  `json:"content"`
	SelectedText string  `json:"selectedText,omitempty"`
	Color        string  `json:"color,omitempty"`
}

// BookConfig 书籍配置结构体（包含元数据和进度）
type BookConfig struct {
	ID               string              `json:"id"`
	Title            string              `json:"title"`
	Author           string              `json:"author"`
	OriginalFileName string              `json:"originalFileName"`
	MD5              string              `json:"md5"`
	CoverPath        string              `json:"coverPath"`
	CreatedAt        int64               `json:"createdAt"`
	LastCFI          string              `json:"last_cfi"`
	Bookmarks        []Bookmark          `json:"bookmarks,omitempty"`
	Notes            []Note              `json:"notes,omitempty"`
	SearchHistory    []SearchHistoryItem `json:"searchHistory,omitempty"`
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

// ClearProgress 清除书籍的阅读进度（保留其他元数据）
func ClearProgress(filePath string) error {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config BookConfig
	json.Unmarshal(data, &config)

	config.LastCFI = ""

	prettyData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, prettyData, 0644)
}

// SaveBookmark 保存书签到书籍同目录的 config.json
func SaveBookmark(filePath, bookmarkJSON string) error {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	var config BookConfig
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	var bookmark Bookmark
	if err := json.Unmarshal([]byte(bookmarkJSON), &bookmark); err != nil {
		return err
	}

	// 检查是否已存在相同位置的书签，存在则更新时间戳
	for i, b := range config.Bookmarks {
		if b.CFI == bookmark.CFI {
			config.Bookmarks[i] = bookmark
			goto save
		}
	}

	// 不存在则添加新书签
	config.Bookmarks = append(config.Bookmarks, bookmark)

save:
	prettyData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, prettyData, 0644)
}

// GetBookmarks 从书籍同目录的 config.json 读取书签列表
func GetBookmarks(filePath string) ([]Bookmark, error) {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config BookConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config.Bookmarks, nil
}

// DeleteBookmark 删除指定 CFI 的书签
func DeleteBookmark(filePath, cfi string) error {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config BookConfig
	json.Unmarshal(data, &config)

	// 删除指定 CFI 的书签
	newBookmarks := make([]Bookmark, 0, len(config.Bookmarks))
	for _, b := range config.Bookmarks {
		if b.CFI != cfi {
			newBookmarks = append(newBookmarks, b)
		}
	}

	config.Bookmarks = newBookmarks

	prettyData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, prettyData, 0644)
}

// SaveSearchHistory 保存搜索历史到书籍同目录的 config.json
func SaveSearchHistory(filePath, keyword string) error {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	var config BookConfig
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	// 检查是否已存在相同关键词，存在则更新时间戳
	for i, h := range config.SearchHistory {
		if h.Keyword == keyword {
			config.SearchHistory[i].Timestamp = time.Now().UnixMilli()
			goto save
		}
	}

	// 不存在则添加新搜索历史
	config.SearchHistory = append(config.SearchHistory, SearchHistoryItem{
		Keyword:   keyword,
		Timestamp: time.Now().UnixMilli(),
	})

	// 限制最多保存20条搜索历史
	if len(config.SearchHistory) > 20 {
		config.SearchHistory = config.SearchHistory[len(config.SearchHistory)-20:]
	}

save:
	prettyData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, prettyData, 0644)
}

// GetSearchHistory 从书籍同目录的 config.json 读取搜索历史
func GetSearchHistory(filePath string) ([]SearchHistoryItem, error) {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config BookConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config.SearchHistory, nil
}

// ClearSearchHistory 清除书籍的搜索历史
func ClearSearchHistory(filePath string) error {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config BookConfig
	json.Unmarshal(data, &config)

	config.SearchHistory = nil

	prettyData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, prettyData, 0644)
}

// SaveNote 保存笔记到书籍同目录的 config.json
func SaveNote(filePath, noteJSON string) error {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	var config BookConfig
	data, err := os.ReadFile(configPath)
	if err == nil {
		json.Unmarshal(data, &config)
	}

	var note Note
	if err := json.Unmarshal([]byte(noteJSON), &note); err != nil {
		return err
	}

	for i, n := range config.Notes {
		if n.CFI == note.CFI {
			config.Notes[i] = note
			goto save
		}
	}

	config.Notes = append(config.Notes, note)

save:
	prettyData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, prettyData, 0644)
}

// GetNotes 从书籍同目录的 config.json 读取笔记列表
func GetNotes(filePath string) ([]Note, error) {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config BookConfig
	if err := json.Unmarshal(data, &config); err != nil {
		return nil, err
	}

	return config.Notes, nil
}

// DeleteNote 删除指定 CFI 的笔记
func DeleteNote(filePath, cfi string) error {
	dir := filepath.Dir(filePath)
	configPath := filepath.Join(dir, "config.json")

	data, err := os.ReadFile(configPath)
	if err != nil {
		return err
	}

	var config BookConfig
	json.Unmarshal(data, &config)

	newNotes := make([]Note, 0, len(config.Notes))
	for _, n := range config.Notes {
		if n.CFI != cfi {
			newNotes = append(newNotes, n)
		}
	}

	config.Notes = newNotes

	prettyData, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(configPath, prettyData, 0644)
}
