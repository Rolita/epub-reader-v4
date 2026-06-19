package book

import (
	"crypto/md5"
	"encoding/hex"
	"io"
	"os"
	"path/filepath"
)

// GetFileBytes 读取文件内容
func GetFileBytes(filePath string) ([]byte, error) {
	return os.ReadFile(filePath)
}

// SaveFile 保存文件到指定目录
func SaveFile(dirPath, fileName string, data []byte) (string, error) {
	if err := os.MkdirAll(dirPath, 0755); err != nil {
		return "", err
	}
	fullPath := filepath.Join(dirPath, fileName)
	if err := os.WriteFile(fullPath, data, 0644); err != nil {
		return "", err
	}
	return fullPath, nil
}

// CalculateMD5 计算数据的 MD5 哈希值
func CalculateMD5(data []byte) string {
	hash := md5.Sum(data)
	return hex.EncodeToString(hash[:])
}

// CopyFile 复制文件到目标目录
func CopyFile(srcPath, destDir, destFileName string) (string, error) {
	srcFile, err := os.Open(srcPath)
	if err != nil {
		return "", err
	}
	defer srcFile.Close()

	if err := os.MkdirAll(destDir, 0755); err != nil {
		return "", err
	}

	destPath := filepath.Join(destDir, destFileName)
	destFile, err := os.Create(destPath)
	if err != nil {
		return "", err
	}
	defer destFile.Close()

	_, err = io.Copy(destFile, srcFile)
	return destPath, err
}

// DeleteDirectory 删除目录及其内容
func DeleteDirectory(dirPath string) error {
	return os.RemoveAll(dirPath)
}
