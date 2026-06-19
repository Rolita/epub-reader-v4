package webdav

import (
	"encoding/json"
	"os"
	"path/filepath"
)

// Config WebDAV 配置结构
type Config struct {
	BaseURL    string `json:"base_url"`
	RemotePath string `json:"remote_path"`
	Username   string `json:"username"`
	Password   string `json:"password"`
}

// LoadConfig 读取 webdav.json 并返回指定书架的配置
func LoadConfig(shelfName string) (*Config, error) {
	configDir, _ := os.UserConfigDir()
	path := filepath.Join(configDir, "my-epub-reader", "webdav.json")

	data, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	var allConfigs map[string]Config
	if err := json.Unmarshal(data, &allConfigs); err != nil {
		return nil, err
	}

	config, ok := allConfigs[shelfName]
	if !ok {
		return nil, os.ErrNotExist
	}
	return &config, nil
}

// SaveConfig 保存指定书架的 WebDAV 配置
func SaveConfig(shelfName string, cfg *Config) error {
	configDir, _ := os.UserConfigDir()
	configPath := filepath.Join(configDir, "my-epub-reader")

	// 确保目录存在
	if err := os.MkdirAll(configPath, 0755); err != nil {
		return err
	}

	path := filepath.Join(configPath, "webdav.json")

	// 读取现有配置
	var allConfigs map[string]Config
	if data, err := os.ReadFile(path); err == nil {
		if err := json.Unmarshal(data, &allConfigs); err != nil {
			return err
		}
	} else {
		allConfigs = make(map[string]Config)
	}

	// 更新配置
	allConfigs[shelfName] = *cfg

	// 写入文件
	data, err := json.MarshalIndent(allConfigs, "", "  ")
	if err != nil {
		return err
	}

	return os.WriteFile(path, data, 0644)
}

// RenameConfig 在重命名书架时，同步更新 webdav.json 中的 Key
func RenameConfig(oldName, newName string) error {
	configDir, _ := os.UserConfigDir()
	path := filepath.Join(configDir, "my-epub-reader", "webdav.json")

	// 1. 读取完整配置
	data, err := os.ReadFile(path)
	if err != nil {
		return err // 如果文件不存在，说明没配置过，直接返回即可
	}

	var allConfigs map[string]Config
	if err := json.Unmarshal(data, &allConfigs); err != nil {
		return err
	}

	// 2. 执行 Key 迁移
	if cfg, exists := allConfigs[oldName]; exists {
		allConfigs[newName] = cfg   // 将旧配置复制给新 Key
		delete(allConfigs, oldName) // 删除旧 Key
	} else {
		return nil // 原来没配置过 WebDAV，不需要做任何操作
	}

	// 3. 覆盖写入
	newData, err := json.MarshalIndent(allConfigs, "", "  ")
	if err != nil {
		return err
	}
	return os.WriteFile(path, newData, 0644)
}
