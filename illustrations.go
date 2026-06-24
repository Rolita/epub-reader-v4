package main

import (
	"archive/zip"
	"fmt"
	"io"
	"path/filepath"
	"strings"
	"sync"
	"time"
)

var (
	tabEpubPathMap   = make(map[string]string)
	tabEpubPathMutex sync.RWMutex

	// EPUB 文件缓存
	epubCache   = make(map[string]*epubCacheEntry)
	epubCacheMutex sync.RWMutex
)

type epubCacheEntry struct {
	reader *zip.ReadCloser
	pathMap map[string]*zip.File
	lastUsed time.Time
}

// RegisterEpubTab 注册 tabId 和 EPUB 文件路径的映射
func (a *App) RegisterEpubTab(tabId, epubPath string) error {
	tabEpubPathMutex.Lock()
	defer tabEpubPathMutex.Unlock()

	tabEpubPathMap[tabId] = epubPath
	
	// 预先加载并缓存 EPUB 文件
	_, err := getOrOpenEpub(epubPath)
	if err != nil {
		return fmt.Errorf("缓存 EPUB 文件失败: %w", err)
	}
	
	return nil
}

// UnregisterEpubTab 注销 tabId 和 EPUB 文件路径的映射
func (a *App) UnregisterEpubTab(tabId string) {
	tabEpubPathMutex.Lock()
	defer tabEpubPathMutex.Unlock()

	delete(tabEpubPathMap, tabId)
}

// GetEpubPathByTabId 根据 tabId 获取 EPUB 文件路径
func GetEpubPathByTabId(tabId string) (string, bool) {
	tabEpubPathMutex.RLock()
	defer tabEpubPathMutex.RUnlock()

	path, ok := tabEpubPathMap[tabId]
	return path, ok
}

// getOrOpenEpub 获取或打开 EPUB 文件，带缓存
func getOrOpenEpub(epubPath string) (*epubCacheEntry, error) {
	epubCacheMutex.RLock()
	entry, exists := epubCache[epubPath]
	epubCacheMutex.RUnlock()

	if exists {
		entry.lastUsed = time.Now()
		return entry, nil
	}

	// 打开 EPUB 文件
	r, err := zip.OpenReader(epubPath)
	if err != nil {
		return nil, err
	}

	// 构建路径映射
	pathMap := make(map[string]*zip.File)
	for _, f := range r.File {
		f.Name = strings.ReplaceAll(f.Name, "\\", "/")
		pathMap[f.Name] = f
	}

	entry = &epubCacheEntry{
		reader:   r,
		pathMap:  pathMap,
		lastUsed: time.Now(),
	}

	epubCacheMutex.Lock()
	epubCache[epubPath] = entry
	epubCacheMutex.Unlock()

	return entry, nil
}

// GetEpubImageByPath 根据 tabId 和资源路径获取 EPUB 中的图片数据
func (a *App) GetEpubImageByPath(tabId, resPath string) ([]byte, string, error) {
	// 获取 EPUB 文件路径
	epubPath, ok := GetEpubPathByTabId(tabId)
	if !ok {
		return nil, "", fmt.Errorf("未找到 tabId %s 对应的 EPUB 文件", tabId)
	}

	// 获取缓存的 EPUB
	entry, err := getOrOpenEpub(epubPath)
	if err != nil {
		return nil, "", fmt.Errorf("无法打开 EPUB 文件: %w", err)
	}

	// 查找文件 - 直接在路径映射中查找
	var targetFile *zip.File
	if f, found := entry.pathMap[resPath]; found {
		targetFile = f
	} else {
		// 尝试带前缀匹配
		for path, f := range entry.pathMap {
			if strings.HasSuffix(path, "/"+resPath) {
				targetFile = f
				break
			}
		}
	}

	if targetFile == nil {
		return nil, "", fmt.Errorf("未找到资源路径 %s", resPath)
	}

	rc, err := targetFile.Open()
	if err != nil {
		return nil, "", fmt.Errorf("无法打开图片文件: %w", err)
	}
	defer rc.Close()

	data, err := io.ReadAll(rc)
	if err != nil {
		return nil, "", fmt.Errorf("读取图片数据失败: %w", err)
	}

	ext := strings.ToLower(filepath.Ext(targetFile.Name))
	mimeType := getMimeType(ext)

	return data, mimeType, nil
}

// getMimeType 根据文件扩展名获取 MIME 类型
func getMimeType(ext string) string {
	mimeTypes := map[string]string{
		".jpg":  "image/jpeg",
		".jpeg": "image/jpeg",
		".png":  "image/png",
		".gif":  "image/gif",
		".webp": "image/webp",
		".svg":  "image/svg+xml",
		".bmp":  "image/bmp",
	}
	if mimeType, ok := mimeTypes[ext]; ok {
		return mimeType
	}
	return "image/png"
}
