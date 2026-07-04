package main

import (
	"embed"
	"fmt"
	"log"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"sync"
	"syscall"
	"time"
	"unsafe"

	"github.com/wailsapp/wails/v2"
	"github.com/wailsapp/wails/v2/pkg/options"
	"github.com/wailsapp/wails/v2/pkg/options/assetserver"
	"github.com/wailsapp/wails/v2/pkg/options/windows"
)

var (
	user32              = syscall.NewLazyDLL("user32.dll")
	findWindow          = user32.NewProc("FindWindowW")
	setForegroundWindow = user32.NewProc("SetForegroundWindow")
	showWindow          = user32.NewProc("ShowWindow")

	kernel32            = syscall.NewLazyDLL("kernel32.dll")
	createMutexW        = kernel32.NewProc("CreateMutexW")
	waitForSingleObject = kernel32.NewProc("WaitForSingleObject")
	closeHandle         = kernel32.NewProc("CloseHandle")
)

const (
	SW_RESTORE     = 9
	SW_SHOW        = 5
	WAIT_ABANDONED = 0x00000080
	WAIT_OBJECT_0  = 0x00000000
	WAIT_TIMEOUT   = 0x00000102
)

func activateExistingInstance() bool {
	className, _ := syscall.UTF16PtrFromString("wails_application")
	title, _ := syscall.UTF16PtrFromString("EPUB Reader")

	hwnd, _, _ := findWindow.Call(uintptr(unsafe.Pointer(className)), 0)
	if hwnd == 0 {
		log.Println("Trying to find window by title")
		hwnd, _, _ = findWindow.Call(0, uintptr(unsafe.Pointer(title)))
	}
	if hwnd == 0 {
		log.Println("Cannot find existing window")
		return false
	}

	log.Printf("Found window handle: %d\n", hwnd)
	showWindow.Call(hwnd, SW_RESTORE)
	setForegroundWindow.Call(hwnd)
	return true
}

var instanceMutex sync.Mutex
var isFirstInstance bool

func checkSingleInstance() bool {
	instanceMutex.Lock()
	defer instanceMutex.Unlock()

	if isFirstInstance {
		log.Println("isFirstInstance is true, returning true")
		return true
	}

	mutexName, _ := syscall.UTF16PtrFromString("EPUBReaderInstanceMutex")
	mutexHandle, _, _ := createMutexW.Call(0, 0, uintptr(unsafe.Pointer(mutexName)))
	if mutexHandle == 0 {
		log.Println("createMutexW failed")
		return false
	}

	result, _, _ := waitForSingleObject.Call(mutexHandle, 0)
	log.Printf("waitForSingleObject result: %d\n", result)

	if result == WAIT_OBJECT_0 {
		log.Println("Got mutex, first instance")
		isFirstInstance = true
		go func() {
			for {
				time.Sleep(time.Second)
			}
		}()
		return false
	}

	log.Println("Mutex already held, existing instance")
	closeHandle.Call(mutexHandle)
	return true
}

func sendEpubPathToExistingInstance(epubPath string) {
	time.Sleep(300 * time.Millisecond)
	log.Printf("Writing EPUB path to temp file: %s\n", epubPath)

	tempDir := os.TempDir()
	ipcFile := filepath.Join(tempDir, "epub-reader-ipc.txt")

	err := os.WriteFile(ipcFile, []byte(epubPath), 0644)
	if err != nil {
		log.Printf("Failed to write IPC file: %v\n", err)
		return
	}

	log.Println("EPUB path written to IPC file")
}

//go:embed all:frontend/dist
var assets embed.FS

// FileAssetHandler 处理书籍封面等动态资源的请求
type FileAssetHandler struct{}

func (f *FileAssetHandler) ServeHTTP(w http.ResponseWriter, r *http.Request) {
	configDir, err := os.UserConfigDir()
	if err != nil {
		http.Error(w, "无法获取配置目录", http.StatusInternalServerError)
		return
	}

	actualPath := filepath.Join(configDir, "my-epub-reader", strings.TrimPrefix(r.URL.Path, "/"))

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

	path = strings.ReplaceAll(path, "%20", " ")
	path = strings.ReplaceAll(path, "%2F", "/")

	if _, err := os.Stat(path); os.IsNotExist(err) {
		http.Error(w, "文件不存在", http.StatusNotFound)
		return
	}

	w.Header().Set("Content-Type", "application/epub+zip")
	w.Header().Set("Cache-Control", "no-cache")

	http.ServeFile(w, r, path)
}

func main() {
	logFile, err := os.Create(filepath.Join(os.TempDir(), "epub-reader-debug.log"))
	if err != nil {
		fmt.Println("Failed to create log file:", err)
	}
	log.SetOutput(logFile)
	log.SetFlags(log.LstdFlags | log.Lshortfile)

	log.Println("===== EPUB Reader Starting =====")
	log.Printf("Args: %v\n", os.Args)

	if checkSingleInstance() {
		log.Println("Found existing instance")
		if len(os.Args) > 1 {
			epubPath := os.Args[1]
			log.Printf("Sending EPUB path: %s\n", epubPath)
			sendEpubPathToExistingInstance(epubPath)
		}
		activateExistingInstance()
		log.Println("Exiting secondary instance")
		return
	}

	log.Println("First instance, starting app")

	var epubPath string
	if len(os.Args) > 1 {
		epubPath = os.Args[1]
	}

	app := NewApp()
	if epubPath != "" {
		app.pendingEpubPath = epubPath
	}

	assetsHandler := &FileAssetHandler{}
	epubImageHandler := &EpubImageHandler{app: app}
	localFileHandler := &LocalFileHandler{}

	runErr := wails.Run(&options.App{
		Title:     "EPUB Reader",
		Width:     1920,
		Height:    1080,
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

	if runErr != nil {
		panic(runErr)
	}
}
