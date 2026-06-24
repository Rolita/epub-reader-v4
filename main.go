package main

import (
	"embed"
	"fmt"
	"net/http"
	"os"
	"path/filepath"
	"strings"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

//go:embed all:frontend/dist
var assets embed.FS

// FileAssetHandler 处理书籍封面等动态资源的请求
type FileAssetHandler struct{}

func (f *FileAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	// 获取配置目录路径
	configDir, err := os.UserConfigDir()
	if err != nil {
		http.Error(w, "无法获取配置目录", http.StatusInternalServerError)
		return
	}

	// 构建实际物理路径: AppData/my-epub-reader/books/...
	actualPath := filepath.Join(configDir, "my-epub-reader", strings.TrimPrefix(r.URL.Path, "/"))

	// 使用 http.ServeFile 提供文件服务
	http.ServeFile(w, r, actualPath)
}

// EpubImageHandler 处理 EPUB 内部图片的请求
type EpubImageHandler struct {
	app *App
}

func (h *EpubImageHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/epub-img/")

	parts := strings.SplitN(path, "/", 2)
	if len(parts) != 2 {
		http.Error(w, "无效的请求路径", http.StatusBadRequest)
		return
	}

	tabId := parts[0]
	resPath := parts[1]

	if tabId == "" || resPath == "" {
		http.Error(w, "缺少必要参数", http.StatusBadRequest)
		return
	}

	data, mimeType, err := h.app.GetEpubImageByPath(tabId, resPath)
	if err != nil {
		http.Error(w, fmt.Sprintf("获取图片失败: %v", err), http.StatusInternalServerError)
		return
	}

	w.Header().Set("Content-Type", mimeType)
	w.Header().Set("Cache-Control", "public, max-age=3600")

	w.Write(data)
}

// LocalFileHandler 处理本地 EPUB 文件的请求
type LocalFileHandler struct{}

func (h *LocalFileHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	path := strings.TrimPrefix(r.URL.Path, "/local-file/")
	
	// 解码路径（处理特殊字符）
	path = strings.ReplaceAll(path, "%20", " ")
	path = strings.ReplaceAll(path, "%2F", "/")
	
	// 安全检查：确保路径存在
	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}
	
	// 设置适当的响应头
	w.Header().Set("Content-Type", "application/epub+zip")
	w.Header().Set("Cache-Control", "no-cache")
	
	// 提供文件服务
	http.ServeFile(w, r, path)
}

func main() {
	app := NewApp()

	assetsHandler := &FileAssetHandler{}
	epubImageHandler := &EpubImageHandler{app: app}
	localFileHandler := &LocalFileHandler{}

	err := wails.Run(&options.App{
		Title:     "EPUB Reader",
		Width:     1314,
		Height:    843,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/books/") {
					assetsHandler.ServeHTTP(w, r)
					return
				}
				if strings.HasPrefix(r.URL.Path, "/epub-img/") {
					epubImageHandler.ServeHTTP(w, r)
					return
				}
				if strings.HasPrefix(r.URL.Path, "/local-file/") {
					localFileHandler.ServeHTTP(w, r)
					return
				}
				http.FileServer(http.FS(assets)).ServeHTTP(w, r)
			}),
		},
		OnStartup: app.startup,
		Bind: []interface{}{
			app,
		},
		Windows: &windows.Options{
			WebviewIsTransparent: false,
			WindowIsTranslucent:  false,
			DisableWindowIcon:    false,
		},
	})

	if err != nil {
		panic(err)
	}
}
