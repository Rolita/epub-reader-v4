package shelf

import (
	"encoding/json"
	"os"
	"path/filepath"
	"time"
)

// Metadata 书架元数据
type Metadata struct {
	Version    int   `json:"version"`
	LastSynced int64 `json:"last_synced"`
}

// Book 书籍信息
type Book struct {
	ID       string `json:"id"`
	Title    string `json:"title"`
	Author   string `json:"author"`
	CoverURL string `json:"coverUrl"`
	FilePath string `json:"filePath"`
	ShelfID  string `json:"shelfId"`
}

// Library 书架库数据结构
type Library struct {
	Metadata Metadata `json:"metadata"`
	Books    []Book   `json:"books"`
}

// LoadLibrary 加载书架库数据
func LoadLibrary(shelfDir, shelfName string) (*Library, error) {
	filePath := filepath.Join(shelfDir, shelfName+"_library.json")

	if _, err := os.Stat(filePath); os.IsNotExist(err) {
		// 返回空的库结构
		return &Library{
			Metadata: Metadata{
				Version:    1,
				LastSynced: 0,
			},
			Books: []Book{},
		}, nil
	}

	data, err := os.ReadFile(filePath)
	if err != nil {
		return nil, err
	}

	var lib Library
	if err := json.Unmarshal(data, &lib); err != nil {
		return nil, err
	}

	// 如果没有 metadata，初始化默认值
	if lib.Metadata.Version == 0 {
		lib.Metadata.Version = 1
	}

	return &lib, nil
}

// SaveLibrary 保存书架库数据，并更新时间戳
func SaveLibrary(shelfDir, shelfName string, lib *Library) error {
	// 更新时间戳
	lib.Metadata.LastSynced = time.Now().Unix()

	filePath := filepath.Join(shelfDir, shelfName+"_library.json")
	data, err := json.MarshalIndent(lib, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(filePath, data, 0644)
}

// UpdateLibraryTimestamp 更新书架时间戳（用于标记修改）
func UpdateLibraryTimestamp(shelfDir, shelfName string) error {
	lib, err := LoadLibrary(shelfDir, shelfName)
	if err != nil {
		return err
	}

	lib.Metadata.LastSynced = time.Now().Unix()
	return SaveLibrary(shelfDir, shelfName, lib)
}
