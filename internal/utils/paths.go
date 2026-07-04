package utils

import (
	"encoding/base64"
	"os"
	"path/filepath"
)

func Base64Decode(data string) ([]byte, error) {
	return base64.StdEncoding.DecodeString(data)
}

const AppName = "my-epub-reader"

// GetConfigDir 获取应用配置目录
func GetConfigDir() string {
	configDir, _ := os.UserConfigDir()
	return filepath.Join(configDir, AppName)
}

// GetShelfDir 获取书架目录（存放各书架的 library.json）
func GetShelfDir() string {
	shelfDir := filepath.Join(GetConfigDir(), "shelves")
	if _, err := os.Stat(shelfDir); os.IsNotExist(err) {
		os.MkdirAll(shelfDir, 0755)
	}
	return shelfDir
}

// GetBooksDir 获取书籍存储目录
func GetBooksDir() string {
	booksDir := filepath.Join(GetConfigDir(), "books")
	if _, err := os.Stat(booksDir); os.IsNotExist(err) {
		os.MkdirAll(booksDir, 0755)
	}
	return booksDir
}
