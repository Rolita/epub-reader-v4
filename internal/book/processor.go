package book

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/timsims/pamphlet"
)

// BookMetadata 对应前端的 BookMetadata 接口
type BookMetadata struct {
	Title       string `json:"title"`
	Author      string `json:"author"`
	Description string `json:"description"` // 简介
}

// BookConfig 对应前端的 BookConfig 接口
type BookConfig struct {
	ID               string `json:"id"`
	Title            string `json:"title"`
	Author           string `json:"author,omitempty"`
	Description      string `json:"description,omitempty"` // 简介
	OriginalFileName string `json:"originalFileName"`
	Md5              string `json:"md5"`
	CoverPath        string `json:"coverPath,omitempty"` // 这里的 CoverPath 是 epub 内部路径，用于在前端显示时作为 fallback
	CreatedAt        int64  `json:"createdAt"`
}

// ImportResult 对应前端的 ImportResult 接口
type ImportResult struct {
	Success  bool   `json:"success"`
	Title    string `json:"title"`
	Author   string `json:"author,omitempty"`
	CoverURL string `json:"coverUrl,omitempty"`
	Md5      string `json:"md5"`
	FilePath string `json:"filePath"`
	Error    string `json:"error,omitempty"`
}

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
		return fullPath, err
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

// extractEpubMetadata 从 EPUB 文件中提取元数据（标题、作者、封面、简介）
func extractEpubMetadata(epubFilePath string) (BookMetadata, []byte, string, error) {
	metadata := BookMetadata{}
	var coverData []byte
	var coverInternalPath string // epub 内部的封面路径

	parser, err := pamphlet.Open(epubFilePath)
	if err != nil {
		return metadata, nil, "", fmt.Errorf("failed to open epub: %w", err)
	}
	book := parser.GetBook()

	metadata.Title = book.Title
	metadata.Author = book.Author
	metadata.Description = book.Description

	// 尝试从 EPUB 中提取封面
	coverPath := ""
	for _, item := range book.ManifestItems {
		// 查找具有 'cover' 属性的 item 或 media-type 为图像的 item
		// 简单的启发式方法：查找文件名中包含 "cover" 或 "thumb" 的图片
		if strings.Contains(strings.ToLower(item.Href), "cover") || strings.Contains(strings.ToLower(item.Href), "thumb") {
			if strings.HasPrefix(item.MediaType, "image/") {
				coverPath = item.Href
				break
			}
		}
	}

	if coverPath != "" {
		zipReader, err := zip.OpenReader(epubFilePath)
		if err != nil {
			return metadata, nil, "", fmt.Errorf("failed to open epub zip: %w", err)
		}
		defer zipReader.Close()

		for _, f := range zipReader.File {
			if strings.HasSuffix(f.Name, filepath.Base(coverPath)) {
				rc, err := f.Open()
				if err != nil {
					continue
				}
				defer rc.Close()

				data, err := io.ReadAll(rc)
				if err != nil {
					continue
				}
				coverData = data
				coverInternalPath = coverPath
				break
			}
		}
	}

	return metadata, coverData, coverInternalPath, nil
}

// ProcessAndImportEpub 整合 EPUB 文件的读取、MD5 计算、元数据提取和文件保存
func ProcessAndImportEpub(filePath, shelfName, booksDir string) ImportResult {
	result := ImportResult{
		Success:  false,
		Title:    "",
		Md5:      "",
		FilePath: "",
	}

	// 1. 读取文件内容 (实际上我们只需要文件路径给 pamphlet)
	fileBytes, err := GetFileBytes(filePath)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to read file: %v", err)
		return result
	}

	// 2. 计算 MD5
	bookMd5 := CalculateMD5(fileBytes)
	result.Md5 = bookMd5

	// 3. 获取书籍目录
	bookDestDir := filepath.Join(booksDir, shelfName, bookMd5)

	// 4. 提取文件名
	originalFileName := filepath.Base(filePath)

	// 5. 提取元数据和封面数据
	metadata, coverData, coverInternalPath, err := extractEpubMetadata(filePath)
	if err != nil {
		fmt.Printf("Error extracting epub metadata: %v. Attempting fallback.\n", err)
		// 即使解析失败，也尝试使用文件名作为标题
		metadata.Title = strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
		metadata.Author = "未知作者" // 兜底
		// 继续，不返回错误，让后续流程尝试保存文件
	}

	// 文件名兜底作者
	if metadata.Author == "" || metadata.Author == "未知作者" {
		nameWithoutExt := strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
		// 常见分隔符：—— _ - 【】 []
		// 这里简化处理，可以根据实际情况增加更复杂的正则表达式
		parts := strings.SplitN(nameWithoutExt, "——", 2)
		if len(parts) > 1 {
			metadata.Author = strings.TrimSpace(parts[0])
			if metadata.Title == "" { // 如果标题也为空，则尝试从文件名中提取
				metadata.Title = strings.TrimSpace(parts[1])
			}
		} else {
			parts = strings.SplitN(nameWithoutExt, "_", 2)
			if len(parts) > 1 {
				metadata.Author = strings.TrimSpace(parts[0])
				if metadata.Title == "" {
					metadata.Title = strings.TrimSpace(parts[1])
				}
			}
		}
		if metadata.Author == "" {
			metadata.Author = "未知作者"
		}
	}

	// 标题兜底
	if metadata.Title == "" {
		metadata.Title = strings.TrimSuffix(originalFileName, filepath.Ext(originalFileName))
	}

	result.Title = metadata.Title
	result.Author = metadata.Author

	// 6. 保存封面
	var coverURL string
	if coverData != nil && len(coverData) > 0 {
		coverFileName := "cover.png" // 统一保存为 png
		_, err := SaveFile(bookDestDir, coverFileName, coverData)
		if err != nil {
			fmt.Printf("Failed to save cover file: %v\n", err)
		} else {
			coverURL = fmt.Sprintf("books/%s/%s/%s", shelfName, bookMd5, coverFileName)
			result.CoverURL = coverURL
		}
	}

	// 7. 复制书籍本体
	bookFilePath, err := CopyFile(filePath, bookDestDir, originalFileName)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to copy book file: %v", err)
		return result
	}
	result.FilePath = bookFilePath

	// 8. 创建配置文件
	config := BookConfig{
		ID:               bookMd5,
		Title:            metadata.Title,
		Author:           metadata.Author,
		Description:      metadata.Description,
		OriginalFileName: originalFileName,
		Md5:              bookMd5,
		CoverPath:        coverInternalPath, // 保存 epub 内部的封面路径，用于前端 fallback
		CreatedAt:        time.Now().UnixMilli(),
	}

	configJson, err := json.MarshalIndent(config, "", "  ")
	if err != nil {
		result.Error = fmt.Sprintf("Failed to marshal config: %v", err)
		return result
	}

	_, err = SaveFile(bookDestDir, "config.json", configJson)
	if err != nil {
		result.Error = fmt.Sprintf("Failed to save config file: %v", err)
		return result
	}

	result.Success = true
	return result
}
