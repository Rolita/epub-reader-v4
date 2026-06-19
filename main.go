package main

import (
	"embed"
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

func main() {
	app := NewApp()

	// 实例化处理器
	assetsHandler := &FileAssetHandler{}

	err := wails.Run(&options.App{
		Title:     "EPUB Reader",
		Width:     1314,
		Height:    843,
		Frameless: true,
		AssetServer: &assetserver.Options{
			Assets: assets,
			// 注册处理器，拦截 /books/ 的请求
			Handler: http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if strings.HasPrefix(r.URL.Path, "/books/") {
					assetsHandler.ServeHTTP(w, r)
					return
				}
				// 否则交给默认处理器
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
