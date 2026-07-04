package book

import (
	"archive/zip"
	"crypto/md5"
	"encoding/hex"
	"encoding/json"
	"encoding/xml"
	"fmt"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"net/url"
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
	var coverInternalPath string

	parser, err := pamphlet.Open(epubFilePath)
	if err != nil {
		return metadata, nil, "", fmt.Errorf("failed to open epub: %w", err)
	}
	book := parser.GetBook()

	metadata.Title = book.Title
	metadata.Author = book.Author
	metadata.Description = book.Description

	coverData, coverInternalPath = extractCover(epubFilePath)

	return metadata, coverData, coverInternalPath, nil
}

func extractFirstImageSrc(htmlContent string) string {
	lowerContent := strings.ToLower(htmlContent)

	imgStart := strings.Index(lowerContent, "<img")
	if imgStart == -1 {
		imgStart = strings.Index(lowerContent, "<image")
	}
	if imgStart == -1 {
		return ""
	}

	srcStart := strings.Index(lowerContent[imgStart:], "src=\"")
	if srcStart == -1 {
		srcStart = strings.Index(lowerContent[imgStart:], "src='")
		if srcStart == -1 {
			srcStart = strings.Index(lowerContent[imgStart:], "src=")
			if srcStart == -1 {
				return ""
			}
			srcStart += 4
			for i := srcStart; i < len(lowerContent[imgStart:]); i++ {
				c := lowerContent[imgStart+i]
				if c != ' ' && c != '\t' && c != '\n' && c != '\r' {
					srcStart = i
					break
				}
			}
		} else {
			srcStart += 5
		}
	} else {
		srcStart += 5
	}

	srcEnd := strings.Index(lowerContent[imgStart+srcStart:], "\"")
	if srcEnd == -1 {
		srcEnd = strings.Index(lowerContent[imgStart+srcStart:], "'")
		if srcEnd == -1 {
			for i := 0; i < len(lowerContent[imgStart+srcStart:]); i++ {
				c := lowerContent[imgStart+srcStart+i]
				if c == ' ' || c == '\t' || c == '\n' || c == '\r' || c == '>' || c == '/' {
					srcEnd = i
					break
				}
			}
			if srcEnd == -1 {
				srcEnd = len(lowerContent[imgStart+srcStart:])
			}
		}
	}

	imgSrc := htmlContent[imgStart+srcStart : imgStart+srcStart+srcEnd]
	return strings.TrimSpace(imgSrc)
}

func extractCover(epubFilePath string) ([]byte, string) {
	zipReader, err := zip.OpenReader(epubFilePath)
	if err != nil {
		return nil, ""
	}
	defer zipReader.Close()

	var opfContent []byte
	var opfDir string
	fileMap := make(map[string]*zip.File)
	var firstImage *zip.File

	for _, f := range zipReader.File {
		if f.FileInfo().IsDir() {
			continue
		}

		fileMap[f.Name] = f

		lowerName := strings.ToLower(f.Name)

		if firstImage == nil {
			if strings.HasSuffix(lowerName, ".jpg") || strings.HasSuffix(lowerName, ".jpeg") ||
				strings.HasSuffix(lowerName, ".png") || strings.HasSuffix(lowerName, ".gif") ||
				strings.HasSuffix(lowerName, ".webp") || strings.HasSuffix(lowerName, ".svg") {
				firstImage = f
			}
		}

		if opfContent == nil && strings.HasSuffix(lowerName, ".opf") {
			rc, err := f.Open()
			if err == nil {
				opfContent, _ = io.ReadAll(rc)
				rc.Close()

				idx := strings.LastIndex(f.Name, "/")
				if idx >= 0 {
					opfDir = f.Name[:idx] + "/"
				}
			}
		}
	}

	coverPatterns := []string{
		"cover.png", "cover.jpg", "cover.jpeg", "cover.svg",
		"cover-art.png", "cover-art.jpg", "cover-art.jpeg", "cover-art.svg",
		"cover-image.png", "cover-image.jpg", "cover-image.jpeg", "cover-image.svg",
		"frontcover.png", "frontcover.jpg", "frontcover.jpeg", "frontcover.svg",
		"front-cover.png", "front-cover.jpg", "front-cover.jpeg", "front-cover.svg",
		"titlepage.png", "titlepage.jpg", "titlepage.jpeg", "titlepage.svg",
		"title-page.png", "title-page.jpg", "title-page.jpeg", "title-page.svg",
		"title.png", "title.jpg", "title.jpeg", "title.svg",
		"00_cover.png", "00_cover.jpg", "00_cover.jpeg", "00_cover.svg",
		"cover01.png", "cover01.jpg", "cover01.jpeg", "cover01.svg",
		"cover-01.png", "cover-01.jpg", "cover-01.jpeg", "cover-01.svg",
		"bkcover.png", "bkcover.jpg", "bkcover.jpeg", "bkcover.svg",
		"bookcover.png", "bookcover.jpg", "bookcover.jpeg", "bookcover.svg",
	}

	for _, pattern := range coverPatterns {
		lowerPattern := strings.ToLower(pattern)
		for name, f := range fileMap {
			lowerName := strings.ToLower(name)
			baseName := filepath.Base(lowerName)
			if baseName == lowerPattern {
				rc, err := f.Open()
				if err == nil {
					data, _ := io.ReadAll(rc)
					rc.Close()
					if len(data) > 0 {
						return data, name
					}
				}
			}
		}
	}

	if opfContent != nil {
		type OPF struct {
			XMLName  xml.Name `xml:"package"`
			Metadata struct {
				Meta []struct {
					Name    string `xml:"name,attr"`
					Content string `xml:"content,attr"`
				} `xml:"meta"`
			} `xml:"metadata"`
			Guide struct {
				References []struct {
					Type string `xml:"type,attr"`
					Href string `xml:"href,attr"`
				} `xml:"reference"`
			} `xml:"guide"`
			Manifest struct {
				Items []struct {
					ID         string `xml:"id,attr"`
					Href       string `xml:"href,attr"`
					MediaType  string `xml:"media-type,attr"`
					Properties string `xml:"properties,attr"`
				} `xml:"item"`
			} `xml:"manifest"`
		}

		var opf OPF
		if err := xml.Unmarshal(opfContent, &opf); err == nil {
			for _, ref := range opf.Guide.References {
				if strings.EqualFold(ref.Type, "cover") {
					coverPath := opfDir + ref.Href
					if f := findFileInMap(fileMap, coverPath); f != nil {
						lowerHref := strings.ToLower(ref.Href)
						isHtml := strings.HasSuffix(lowerHref, ".html") ||
							strings.HasSuffix(lowerHref, ".xhtml") ||
							strings.HasSuffix(lowerHref, ".htm")

						if isHtml {
							rc, err := f.Open()
							if err == nil {
								content, _ := io.ReadAll(rc)
								rc.Close()
								imgSrc := extractFirstImageSrc(string(content))
								if imgSrc != "" {
									if strings.HasPrefix(imgSrc, "/") {
										imgSrc = imgSrc[1:]
									}
									if !strings.HasPrefix(imgSrc, "http://") && !strings.HasPrefix(imgSrc, "https://") {
										imgDir := ""
										lastSlash := strings.LastIndex(coverPath, "/")
										if lastSlash >= 0 {
											imgDir = coverPath[:lastSlash+1]
										}
										if !strings.Contains(imgSrc, "/") {
											imgSrc = imgDir + imgSrc
										}
										if imgFile := findFileInMap(fileMap, imgSrc); imgFile != nil {
											rc, err := imgFile.Open()
											if err == nil {
												data, _ := io.ReadAll(rc)
												rc.Close()
												if len(data) > 0 {
													return data, imgFile.Name
												}
											}
										}
									}
								}
							}
						} else {
							rc, err := f.Open()
							if err == nil {
								data, _ := io.ReadAll(rc)
								rc.Close()
								if len(data) > 0 {
									return data, f.Name
								}
							}
						}
					}
				}
			}

			var coverID string
			for _, meta := range opf.Metadata.Meta {
				if strings.EqualFold(meta.Name, "cover") {
					coverID = meta.Content
					break
				}
			}

			if coverID != "" {
				for _, item := range opf.Manifest.Items {
					if item.ID == coverID {
						lowerHref := strings.ToLower(item.Href)
						isHtml := strings.HasSuffix(lowerHref, ".html") ||
							strings.HasSuffix(lowerHref, ".xhtml") ||
							strings.HasSuffix(lowerHref, ".htm")

						coverPath := opfDir + item.Href
						if f := findFileInMap(fileMap, coverPath); f != nil {
							if isHtml {
								rc, err := f.Open()
								if err == nil {
									content, _ := io.ReadAll(rc)
									rc.Close()
									imgSrc := extractFirstImageSrc(string(content))
									if imgSrc != "" {
										if strings.HasPrefix(imgSrc, "/") {
											imgSrc = imgSrc[1:]
										}
										if !strings.HasPrefix(imgSrc, "http://") && !strings.HasPrefix(imgSrc, "https://") {
											imgDir := ""
											lastSlash := strings.LastIndex(coverPath, "/")
											if lastSlash >= 0 {
												imgDir = coverPath[:lastSlash+1]
											}
											if !strings.Contains(imgSrc, "/") {
												imgSrc = imgDir + imgSrc
											}
											if imgFile := findFileInMap(fileMap, imgSrc); imgFile != nil {
												rc, err := imgFile.Open()
												if err == nil {
													data, _ := io.ReadAll(rc)
													rc.Close()
													if len(data) > 0 {
														return data, imgFile.Name
													}
												}
											}
										}
									}
								}
							} else if strings.HasPrefix(item.MediaType, "image/") {
								rc, err := f.Open()
								if err == nil {
									data, _ := io.ReadAll(rc)
									rc.Close()
									if len(data) > 0 {
										return data, f.Name
									}
								}
							}
						}
					}
				}
			}

			for _, item := range opf.Manifest.Items {
				if strings.HasPrefix(item.MediaType, "image/") &&
					strings.Contains(strings.ToLower(item.Properties), "cover-image") {
					coverPath := opfDir + item.Href
					if f := findFileInMap(fileMap, coverPath); f != nil {
						rc, err := f.Open()
						if err == nil {
							data, _ := io.ReadAll(rc)
							rc.Close()
							if len(data) > 0 {
								return data, f.Name
							}
						}
					}
				}
			}

			for _, item := range opf.Manifest.Items {
				lowerID := strings.ToLower(item.ID)
				if strings.HasPrefix(item.MediaType, "image/") &&
					(strings.Contains(lowerID, "cover") || strings.Contains(lowerID, "front") ||
						strings.Contains(lowerID, "title") || strings.Contains(lowerID, "thumbnail")) {
					coverPath := opfDir + item.Href
					if f := findFileInMap(fileMap, coverPath); f != nil {
						rc, err := f.Open()
						if err == nil {
							data, _ := io.ReadAll(rc)
							rc.Close()
							if len(data) > 0 {
								return data, f.Name
							}
						}
					}
				}
			}
		}
	}

	if firstImage != nil {
		rc, err := firstImage.Open()
		if err == nil {
			data, _ := io.ReadAll(rc)
			rc.Close()
			if len(data) > 0 {
				return data, firstImage.Name
			}
		}
	}

	return nil, ""
}

func findFileInMap(fileMap map[string]*zip.File, coverPath string) *zip.File {
	coverPath = strings.TrimPrefix(coverPath, "/")

	if decodedPath, err := url.QueryUnescape(coverPath); err == nil {
		coverPath = decodedPath
	}

	if f, ok := fileMap[coverPath]; ok {
		return f
	}

	for name, f := range fileMap {
		name = strings.TrimPrefix(name, "/")
		if name == coverPath || strings.HasSuffix(name, "/"+coverPath) {
			return f
		}

		lowerName := strings.ToLower(name)
		lowerPath := strings.ToLower(coverPath)
		if lowerName == lowerPath || strings.HasSuffix(lowerName, "/"+lowerPath) {
			return f
		}
	}

	return nil
}

func getCoverFileName(coverInternalPath string) string {
	lowerPath := strings.ToLower(coverInternalPath)
	if strings.HasSuffix(lowerPath, ".jpg") || strings.HasSuffix(lowerPath, ".jpeg") {
		return "cover.jpg"
	}
	if strings.HasSuffix(lowerPath, ".png") {
		return "cover.png"
	}
	if strings.HasSuffix(lowerPath, ".gif") {
		return "cover.gif"
	}
	if strings.HasSuffix(lowerPath, ".webp") {
		return "cover.webp"
	}
	if strings.HasSuffix(lowerPath, ".svg") {
		return "cover.svg"
	}
	return "cover.png"
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
		coverFileName := getCoverFileName(coverInternalPath)
		_, err := SaveFile(bookDestDir, coverFileName, coverData)
		if err != nil {
			fmt.Printf("Failed to save cover file: %v\n", err)
		} else {
			coverURL = fmt.Sprintf("books/%s/%s/%s", shelfName, bookMd5, coverFileName)
			result.CoverURL = coverURL
		}
	}

	// 7. 复制书籍本体（先检查是否已存在）
	targetBookPath := filepath.Join(bookDestDir, originalFileName)
	if _, err := os.Stat(targetBookPath); err == nil {
		fmt.Printf("书籍文件已存在，跳过复制: %s\n", targetBookPath)
		result.FilePath = targetBookPath
	} else {
		bookFilePath, err := CopyFile(filePath, bookDestDir, originalFileName)
		if err != nil {
			result.Error = fmt.Sprintf("Failed to copy book file: %v", err)
			return result
		}
		result.FilePath = bookFilePath
	}

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
