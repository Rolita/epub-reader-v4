package main

import (
	"context"
	"crypto/md5"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"os"
	"os/exec"
	"path/filepath"
	"runtime"
	"strings"
	"sync"

	"epub-reader/internal/book"
	"epub-reader/internal/config"
	"epub-reader/internal/shelf"
	"epub-reader/internal/utils"
	"epub-reader/internal/webdav"

	wailsruntime "github.com/wailsapp/wails/v2/pkg/runtime"
)

// downloadingBooks 用于防止重复下载的内存锁
var downloadingBooks = make(map[string]bool)
var downloadingMu sync.Mutex

type App struct {
	ctx             context.Context
	IsSyncing       bool          // 同步状态标志
	syncLock        chan struct{} // 防抖锁，防止同步风暴
	importLock      chan struct{} // 导入同步专用锁
	pendingEpubPath string        // 启动时待打开的 EPUB 文件路径
}

func NewApp() *App {
	return &App{
		syncLock:   make(chan struct{}, 1), // 普通同步锁
		importLock: make(chan struct{}, 1), // 导入同步专用锁
	}
}

func (a *App) startup(ctx context.Context) {
	a.ctx = ctx

	logPath := filepath.Join(utils.GetShelfDir(), "webdav_logs.json")
	webdav.GlobalLogger.Init(ctx, logPath)

	go a.startIPCServer()
}

func (a *App) startIPCServer() {
	http.HandleFunc("/ipc/open-epub", func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			http.Error(w, "Method not allowed", http.StatusMethodNotAllowed)
			return
		}
		r.ParseForm()
		path := r.FormValue("path")
		if path == "" {
			path = r.URL.Query().Get("path")
		}
		if path != "" {
			a.pendingEpubPath = path
			fmt.Printf("DEBUG: Received EPUB path: %s\n", path)
		}
		w.WriteHeader(http.StatusOK)
	})

	go func() {
		fmt.Println("DEBUG: Starting IPC server on :50001")
		err := http.ListenAndServe(":50001", nil)
		if err != nil {
			fmt.Printf("DEBUG: IPC server error: %v\n", err)
		}
	}()
}

func (a *App) GetPendingEpubPath() string {
	path := a.pendingEpubPath
	a.pendingEpubPath = ""

	if path == "" {
		ipcFile := filepath.Join(os.TempDir(), "epub-reader-ipc.txt")
		data, err := os.ReadFile(ipcFile)
		if err == nil && len(data) > 0 {
			path = strings.TrimSpace(string(data))
			os.Remove(ipcFile)
			fmt.Printf("DEBUG: Read EPUB path from IPC file: %s\n", path)
		}
	}

	return path
}

// ForceQuit 强制退出应用程序
func (a *App) ForceQuit() {
	os.Exit(0)
}

// ============ 路径获取（胶水层） ============

func (a *App) GetShelfDir() string {
	return utils.GetShelfDir()
}

func (a *App) GetBooksDir() string {
	return utils.GetBooksDir()
}

// ============ 书架操作 ============

func (a *App) ScanShelves() (string, error) {
	shelves, err := shelf.Scan(utils.GetShelfDir())
	if err != nil {
		return "[]", err
	}
	data, _ := json.Marshal(shelves)
	return string(data), nil
}

func (a *App) LoadShelfData(shelfName string) (string, error) {
	return shelf.LoadShelfData(utils.GetShelfDir(), shelfName)
}

func (a *App) SaveShelfData(shelfName string, data string) error {
	// 1. 执行原有的本地写入操作
	err := shelf.SaveShelfData(utils.GetShelfDir(), shelfName, data)
	if err != nil {
		return err
	}

	// 2. 触发异步同步 _library.json 到云端
	go a.SyncLibrary(shelfName)

	return nil
}

// SyncLibrary 异步同步书架索引文件到 WebDAV（带防抖）
func (a *App) SyncLibrary(shelfName string) {
	// 尝试获取同步锁
	select {
	case a.syncLock <- struct{}{}:
		defer func() { <-a.syncLock }()

		cfg, err := webdav.LoadConfig(shelfName)
		if err != nil {
			return // 没有配置，跳过同步
		}

		client, err := webdav.NewClientWrapper(cfg)
		if err != nil {
			return // 客户端创建失败，跳过同步
		}

		// 构造本地路径和远程路径
		localLibPath := filepath.Join(utils.GetShelfDir(), shelfName+"_library.json")
		remoteLibPath := "shelves/" + shelfName + "_library.json"

		// 执行上传
		err = client.UploadFile(localLibPath, remoteLibPath)
		if err != nil {
			fmt.Printf("书架索引同步失败 [%s]: %v\n", shelfName, err)
		} else {
			fmt.Printf("书架索引已同步到 WebDAV [%s]\n", shelfName)
		}
	default:
		// 已有同步任务在进行中，忽略本次请求
		fmt.Printf("书架索引同步被防抖拦截 [%s]\n", shelfName)
	}
}

func (a *App) CreateShelf(shelfName string) error {
	return shelf.Create(utils.GetShelfDir(), shelfName)
}

func (a *App) DeleteShelf(shelfName string) error {
	return shelf.Delete(utils.GetShelfDir(), utils.GetBooksDir(), shelfName)
}

func (a *App) RenameShelf(oldName string, newName string) error {
	// 1. 执行书架物理重命名
	if err := shelf.Rename(oldName, newName, utils.GetShelfDir(), utils.GetBooksDir()); err != nil {
		return err
	}

	// 2. 同步迁移 WebDAV 配置
	if err := webdav.RenameConfig(oldName, newName); err != nil {
		fmt.Printf("WebDAV 配置迁移失败（可能未配置过）: %v\n", err)
	}

	// 3. 重命名云端文件夹
	cfg, err := webdav.LoadConfig(newName)
	if err != nil {
		fmt.Printf("加载 WebDAV 配置失败，跳过云端重命名: %v\n", err)
		return nil
	}

	client, err := webdav.NewClientWrapper(cfg)
	if err != nil {
		fmt.Printf("创建 WebDAV 客户端失败，跳过云端重命名: %v\n", err)
		return nil
	}

	oldRemotePath := fmt.Sprintf("books/%s", oldName)
	newRemotePath := fmt.Sprintf("books/%s", newName)

	if err := client.Rename(oldRemotePath, newRemotePath); err != nil {
		fmt.Printf("云端文件夹重命名失败: %v\n", err)
	} else {
		fmt.Printf("云端文件夹已重命名: %s -> %s\n", oldRemotePath, newRemotePath)
	}

	return nil
}

// SaveShelfOrder 保存书架顺序
func (a *App) SaveShelfOrder(order []string) error {
	return shelf.SaveShelfOrder(utils.GetShelfDir(), order)
}

// ============ 书籍操作 ============

func (a *App) GetFileBytes(filePath string) ([]byte, error) {
	return book.GetFileBytes(filePath)
}

func (a *App) SaveFile(dirPath, fileName string, data []byte) (string, error) {
	return book.SaveFile(dirPath, fileName, data)
}

func (a *App) CalculateMD5(data []byte) string {
	return book.CalculateMD5(data)
}

func (a *App) CalculateFileMD5(filePath string) (string, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return "", err
	}
	defer file.Close()

	hash := md5.New()
	if _, err := io.Copy(hash, file); err != nil {
		return "", err
	}

	return fmt.Sprintf("%x", hash.Sum(nil)), nil
}

func (a *App) CopyFile(srcPath, destDir, destFileName string) (string, error) {
	return book.CopyFile(srcPath, destDir, destFileName)
}

func (a *App) DeleteDirectory(dirPath string) error {
	return book.DeleteDirectory(dirPath)
}

func (a *App) ProcessAndImportEpub(filePath, shelfName string) book.ImportResult {
	booksDir := utils.GetBooksDir()
	return book.ProcessAndImportEpub(filePath, shelfName, booksDir)
}

// FileExists 检查本地文件是否存在
func (a *App) FileExists(filePath string) bool {
	_, err := os.Stat(filePath)
	return !os.IsNotExist(err)
}

// GetBookLocalPath 根据书架名称和书籍 MD5 获取书籍在本地的正确路径
// 这用于处理从云端同步后，book.filePath 可能是另一台设备的绝对路径的情况
func (a *App) GetBookLocalPath(shelfName, bookMD5 string) (string, error) {
	booksDir := utils.GetBooksDir()
	bookDir := filepath.Join(booksDir, shelfName, bookMD5)
	configPath := filepath.Join(bookDir, "config.json")

	if _, err := os.Stat(configPath); os.IsNotExist(err) {
		return "", fmt.Errorf("配置文件不存在: %s", configPath)
	}

	configData, err := os.ReadFile(configPath)
	if err != nil {
		return "", fmt.Errorf("读取配置文件失败: %w", err)
	}

	var config struct {
		OriginalFileName string `json:"originalFileName"`
	}
	if err := json.Unmarshal(configData, &config); err != nil {
		return "", fmt.Errorf("解析配置文件失败: %w", err)
	}

	if config.OriginalFileName == "" {
		return "", fmt.Errorf("配置文件中未找到原始文件名")
	}

	return filepath.Join(bookDir, config.OriginalFileName), nil
}

// FileInfo 文件信息结构
type FileInfo struct {
	Path    string `json:"path"`
	Size    int64  `json:"size"`
	ModTime int64  `json:"modTime"` // Unix timestamp
}

// OpenFileLocation 打开文件所在的文件夹并选中文件
func (a *App) OpenFileLocation(filePath string) error {
	// 规范化路径：将所有斜杠统一为反斜杠（Windows）
	normalizedPath := filepath.FromSlash(filePath)

	// 检查文件是否存在
	if _, err := os.Stat(normalizedPath); err != nil {
		return fmt.Errorf("文件不存在: %s", normalizedPath)
	}

	// 使用 Windows explorer 命令打开文件夹并选中文件
	cmd := exec.Command("explorer", "/select,", normalizedPath)
	return cmd.Start()
}

// GetFileInfo 获取文件信息
func (a *App) GetFileInfo(filePath string) (*FileInfo, error) {
	info, err := os.Stat(filePath)
	if err != nil {
		return nil, err
	}

	return &FileInfo{
		Path:    filePath,
		Size:    info.Size(),
		ModTime: info.ModTime().Unix() * 1000, // 转换为毫秒时间戳
	}, nil
}

// ReadFile 读取文件内容
func (a *App) ReadFile(filePath string) (string, error) {
	content, err := os.ReadFile(filePath)
	if err != nil {
		return "", err
	}
	return string(content), nil
}

// DeleteBook 删除书籍（本地 + WebDAV 联动）
func (a *App) DeleteBook(shelfName string, bookID string, bookMD5 string) error {
	// 1. 先执行本地删除
	booksDir := utils.GetBooksDir()
	bookDir := filepath.Join(booksDir, shelfName, bookMD5)
	if err := book.DeleteDirectory(bookDir); err != nil {
		return err
	}

	// 2. 联动删除 WebDAV 远程书籍（异步执行，不阻塞主线程）
	go func() {
		cfg, err := webdav.LoadConfig(shelfName)
		if err != nil {
			return // 没有配置，跳过
		}

		client, err := webdav.NewClientWrapper(cfg)
		if err != nil {
			return // 连接失败，跳过
		}

		// 删除远程书籍文件
		remoteBookPath := "books/" + shelfName + "/" + bookMD5 + ".epub"
		_ = client.DeleteRemote(remoteBookPath)

		// 删除远程书籍目录
		remoteBookDir := "books/" + shelfName + "/" + bookMD5
		_ = client.DeleteRemote(remoteBookDir)

		// 更新云端 _library.json
		localLibPath := filepath.Join(utils.GetShelfDir(), shelfName+"_library.json")
		remoteLibPath := "shelves/" + shelfName + "_library.json"
		_ = client.UploadFile(localLibPath, remoteLibPath)
	}()

	return nil
}

// ============ 进度配置 ============

func (a *App) SaveProgress(filePath string, progressJSON string) error {
	// 1. 执行原有的本地写入操作
	err := config.SaveProgress(filePath, progressJSON)
	if err != nil {
		return err
	}

	// 2. 触发异步同步 (只同步 config.json)
	go a.SyncBookConfig(filePath)

	return nil
}

// SyncBookConfig 异步同步单个书籍的 config.json 到 WebDAV（带防抖）
func (a *App) SyncBookConfig(filePath string) {
	// 从文件路径中提取 shelfName 和 bookID (md5)
	// filePath: C:\Users\Kotori\AppData\Roaming\my-epub-reader\books\{shelfName}\{bookID}\xxx.epub
	dir := filepath.Dir(filePath)
	bookID := filepath.Base(dir)
	parentDir := filepath.Dir(dir)
	shelfName := filepath.Base(parentDir)

	// 尝试获取同步锁
	select {
	case a.syncLock <- struct{}{}:
		defer func() { <-a.syncLock }()

		cfg, err := webdav.LoadConfig(shelfName)
		if err != nil {
			return // 没有配置，跳过同步
		}

		client, err := webdav.NewClientWrapper(cfg)
		if err != nil {
			return // 客户端创建失败，跳过同步
		}

		// 构造本地路径和远程路径
		localPath := filepath.Join(dir, "config.json")
		remotePath := "books/" + shelfName + "/" + bookID + "/config.json"

		// 执行上传
		err = client.UploadFile(localPath, remotePath)
		if err != nil {
			fmt.Printf("配置同步失败 [%s/%s]: %v\n", shelfName, bookID, err)
		} else {
			fmt.Printf("阅读进度已同步到 WebDAV [%s/%s]\n", shelfName, bookID)
		}
	default:
		// 已有同步任务在进行中，忽略本次请求
		fmt.Printf("阅读进度同步被防抖拦截 [%s/%s]\n", shelfName, bookID)
	}
}

func (a *App) GetProgress(filePath string) string {
	return config.GetProgress(filePath)
}

// ============ 系统操作 ============

func (a *App) OpenConfigDir() {
	path := utils.GetConfigDir()
	var cmd *exec.Cmd
	switch runtime.GOOS {
	case "windows":
		cmd = exec.Command("explorer", path)
	case "darwin":
		cmd = exec.Command("open", path)
	default:
		cmd = exec.Command("xdg-open", path)
	}
	cmd.Run()
}

func (a *App) SelectEpubFiles() (string, error) {
	result, err := wailsruntime.OpenFileDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择 EPUB 文件",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "EPUB 文件 (*.epub)",
				Pattern:     "*.epub",
			},
		},
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

// SelectMultipleEpubFiles 选择多个 EPUB 文件
func (a *App) SelectMultipleEpubFiles() ([]string, error) {
	results, err := wailsruntime.OpenMultipleFilesDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择多个 EPUB 文件",
		Filters: []wailsruntime.FileFilter{
			{
				DisplayName: "EPUB 文件 (*.epub)",
				Pattern:     "*.epub",
			},
		},
	})
	if err != nil {
		return nil, err
	}
	return results, nil
}

// SelectEpubFolder 选择包含 EPUB 文件的文件夹
func (a *App) SelectEpubFolder() (string, error) {
	result, err := wailsruntime.OpenDirectoryDialog(a.ctx, wailsruntime.OpenDialogOptions{
		Title: "选择包含 EPUB 文件的文件夹",
	})
	if err != nil {
		return "", err
	}
	return result, nil
}

// ScanEpubFilesInFolder 扫描文件夹中的所有 EPUB 文件
func (a *App) ScanEpubFilesInFolder(folderPath string) ([]string, error) {
	var epubFiles []string

	err := filepath.Walk(folderPath, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return nil // 跳过无法访问的文件
		}
		if !info.IsDir() && filepath.Ext(strings.ToLower(path)) == ".epub" {
			epubFiles = append(epubFiles, path)
		}
		return nil
	})

	if err != nil {
		return nil, err
	}
	return epubFiles, nil
}

// ============ WebDAV 操作 ============

// TestShelfWebDav 连接测试接口，给 UI 调用
func (a *App) TestShelfWebDav(shelfName string) (string, error) {
	cfg, err := webdav.LoadConfig(shelfName)
	if err != nil {
		return "", fmt.Errorf("找不到该书架的 WebDAV 配置: %w", err)
	}

	// 使用 TestConnection 获取详细的连接信息
	result, err := webdav.TestConnection(cfg)
	if err != nil {
		return result, err
	}

	return result, nil
}

// LoadWebDavConfig 加载指定书架的 WebDAV 配置
func (a *App) LoadWebDavConfig(shelfName string) (string, error) {
	cfg, err := webdav.LoadConfig(shelfName)
	if err != nil {
		if os.IsNotExist(err) {
			return "{}", nil
		}
		return "", err
	}

	data, err := json.Marshal(cfg)
	if err != nil {
		return "", err
	}

	return string(data), nil
}

// SaveWebDavConfig 保存指定书架的 WebDAV 配置
func (a *App) SaveWebDavConfig(shelfName string, configJSON string) error {
	var cfg webdav.Config
	if err := json.Unmarshal([]byte(configJSON), &cfg); err != nil {
		return err
	}

	if err := webdav.SaveConfig(shelfName, &cfg); err != nil {
		return err
	}

	client, err := webdav.NewClientWrapper(&cfg)
	if err != nil {
		return nil
	}

	configDir := utils.GetConfigDir()
	localWebdavJson := filepath.Join(configDir, "webdav.json")
	if err := client.UploadFile(localWebdavJson, "webdav.json"); err != nil {
		return nil
	}

	fmt.Printf("webdav.json 已上传到云端\n")
	return nil
}

// SyncShelf 同步指定书架到 WebDAV
func (a *App) SyncShelf(shelfName string) (string, error) {
	// 1. 加载配置
	cfg, err := webdav.LoadConfig(shelfName)
	if err != nil {
		return "", fmt.Errorf("加载 WebDAV 配置失败: %w", err)
	}

	// 2. 创建客户端
	client, err := webdav.NewClientWrapper(cfg)
	if err != nil {
		return "", fmt.Errorf("创建 WebDAV 客户端失败: %w", err)
	}

	// 3. 获取本地库文件路径
	localLibPath := filepath.Join(utils.GetShelfDir(), shelfName+"_library.json")

	// 4. 同步库文件
	uploaded, err := client.SyncLibrary(localLibPath)
	if err != nil {
		return "", fmt.Errorf("同步库文件失败: %w", err)
	}

	if uploaded {
		return "已上传本地更新到云端", nil
	}
	return "已从云端下载最新数据", nil
}

// UploadShelf 上传书架到云端（智能同步，基于时间戳对比）
func (a *App) UploadShelf(shelfName string) (string, error) {
	// 1. 加载配置
	cfg, err := webdav.LoadConfig(shelfName)
	if err != nil {
		return "", fmt.Errorf("加载 WebDAV 配置失败: %w", err)
	}

	// 2. 创建客户端（确保远程路径存在）
	client, err := webdav.NewClientWrapper(cfg)
	if err != nil {
		return "", fmt.Errorf("创建 WebDAV 客户端失败: %w", err)
	}

	configDir := utils.GetConfigDir()

	// 3. 上传 webdav.json (全局配置)
	localWebdavJson := filepath.Join(configDir, "webdav.json")
	if err := client.UploadFile(localWebdavJson, "webdav.json"); err != nil {
		return "", fmt.Errorf("上传 webdav.json 失败: %w", err)
	}

	// 4. 智能上传当前书架的 _library.json
	localLibPath := filepath.Join(utils.GetShelfDir(), shelfName+"_library.json")
	remoteLibPath := "shelves/" + shelfName + "_library.json"

	// 使用智能同步：先对比时间戳
	uploaded, err := client.SmartUploadLibrary(localLibPath, remoteLibPath)
	if err != nil {
		return "", fmt.Errorf("上传书架配置失败: %w", err)
	}

	// 5. 上传当前书架的整个书籍文件夹（无条件上传，书籍文件使用文件时间戳）
	localBooksDir := filepath.Join(utils.GetBooksDir(), shelfName)
	if err := client.UploadDir(localBooksDir, "books/"+shelfName); err != nil {
		return "", fmt.Errorf("上传书籍文件夹失败: %w", err)
	}

	if uploaded {
		return "同步完成：配置文件和书籍已全部上传", nil
	}
	return "同步完成：书架配置已是最新，仅上传书籍文件", nil
}

// TriggerAutoSync 触发后台自动同步（带防抖机制，使用专用锁）
func (a *App) TriggerAutoSync(shelfName string) {
	// 尝试获取导入专用锁，如果锁已被占用则忽略
	select {
	case a.importLock <- struct{}{}:
		// 获得锁，开始同步
		go func() {
			defer func() { <-a.importLock }() // 释放锁
			a.performAutoSync(shelfName)
		}()
	default:
		// 锁已被占用，忽略本次请求
		fmt.Printf("导入同步被防抖拦截：shelf=%s 已有导入同步任务在进行中\n", shelfName)
	}
}

// performAutoSync 执行实际的同步操作
func (a *App) performAutoSync(shelfName string) {
	a.IsSyncing = true
	defer func() { a.IsSyncing = false }()

	fmt.Printf("检测到新书籍，后台启动自动同步: %s\n", shelfName)

	_, err := a.UploadShelf(shelfName)
	if err != nil {
		fmt.Printf("自动同步失败 [%s]: %v\n", shelfName, err)
	} else {
		fmt.Printf("自动同步完成 [%s]\n", shelfName)
	}
}

// DownloadShelf 从云端下载书架（优化版：跳过 .epub 文件，只下载配置和封面）
func (a *App) DownloadShelf(shelfName string) (string, error) {
	// 1. 加载配置
	cfg, err := webdav.LoadConfig(shelfName)
	if err != nil {
		return "", fmt.Errorf("加载 WebDAV 配置失败: %w", err)
	}

	// 2. 创建客户端
	client, err := webdav.NewClientWrapper(cfg)
	if err != nil {
		return "", fmt.Errorf("创建 WebDAV 客户端失败: %w", err)
	}

	configDir := utils.GetConfigDir()

	// 3. 下载 webdav.json (全局配置) - DownloadFile 已内置临时文件保护
	localWebdavJson := filepath.Join(configDir, "webdav.json")
	if err := client.DownloadFile("webdav.json", localWebdavJson); err != nil {
		return "", fmt.Errorf("下载 webdav.json 失败: %w", err)
	}

	// 4. 下载当前书架的 shelves/{shelfName}_library.json - DownloadFile 已内置临时文件保护
	localLibPath := filepath.Join(utils.GetShelfDir(), shelfName+"_library.json")
	remoteLibPath := "shelves/" + shelfName + "_library.json"
	if err := client.DownloadFile(remoteLibPath, localLibPath); err != nil {
		return "", fmt.Errorf("下载书架配置失败: %w", err)
	}

	// 5. 递归下载整个书籍文件夹，排除 .epub 文件
	// 只下载 config.json、cover.png 等配置文件，不下载大型 EPUB 文件
	localBooksDir := filepath.Join(utils.GetBooksDir(), shelfName)
	if err := client.DownloadDirExcluding("books/"+shelfName, localBooksDir, ".epub"); err != nil {
		return "", fmt.Errorf("下载书籍配置文件夹失败: %w", err)
	}

	return "同步完成：书架配置和封面已下载（EPUB 文件已跳过）", nil
}

// DownloadSingleEpub 从 WebDAV 下载指定的书籍文件（按需加载）
func (a *App) DownloadSingleEpub(shelfName string, bookID string, fileName string) error {
	// 检查是否正在下载
	lockKey := shelfName + ":" + bookID
	downloadingMu.Lock()
	if downloadingBooks[lockKey] {
		downloadingMu.Unlock()
		return fmt.Errorf("该书籍正在下载中")
	}
	downloadingBooks[lockKey] = true
	downloadingMu.Unlock()
	defer func() {
		downloadingMu.Lock()
		delete(downloadingBooks, lockKey)
		downloadingMu.Unlock()
	}()

	// 加载配置
	cfg, err := webdav.LoadConfig(shelfName)
	if err != nil {
		return fmt.Errorf("加载 WebDAV 配置失败: %w", err)
	}

	// 创建客户端
	client, err := webdav.NewClientWrapper(cfg)
	if err != nil {
		return fmt.Errorf("创建 WebDAV 客户端失败: %w", err)
	}

	// 获取正确的 EPUB 文件名：优先从本地 config.json 读取，其次从云端下载 config.json
	// 这是为了处理从云端同步书架后，book.filePath 可能是另一台设备的绝对路径的情况
	booksDir := utils.GetBooksDir()
	bookDir := filepath.Join(booksDir, shelfName, bookID)
	configPath := filepath.Join(bookDir, "config.json")

	var realFileName string

	// 尝试从本地读取 config.json
	if _, err := os.Stat(configPath); err == nil {
		configData, err := os.ReadFile(configPath)
		if err == nil {
			var config struct {
				OriginalFileName string `json:"originalFileName"`
			}
			if json.Unmarshal(configData, &config) == nil && config.OriginalFileName != "" {
				realFileName = config.OriginalFileName
			}
		}
	}

	// 如果本地没有或读取失败，从云端下载 config.json
	if realFileName == "" {
		if err := os.MkdirAll(bookDir, 0755); err != nil {
			return fmt.Errorf("创建本地目录失败: %w", err)
		}
		remoteConfigPath := fmt.Sprintf("books/%s/%s/config.json", shelfName, bookID)
		if err := client.DownloadFile(remoteConfigPath, configPath); err != nil {
			return fmt.Errorf("下载配置文件失败: %w", err)
		}

		configData, err := os.ReadFile(configPath)
		if err != nil {
			return fmt.Errorf("读取配置文件失败: %w", err)
		}

		var config struct {
			OriginalFileName string `json:"originalFileName"`
		}
		if err := json.Unmarshal(configData, &config); err != nil {
			return fmt.Errorf("解析配置文件失败: %w", err)
		}

		if config.OriginalFileName == "" {
			return fmt.Errorf("配置文件中未找到原始文件名")
		}
		realFileName = config.OriginalFileName
	}

	// 使用正确的文件名构建远程路径和本地路径
	remotePath := fmt.Sprintf("books/%s/%s/%s", shelfName, bookID, realFileName)
	localPath := filepath.Join(bookDir, realFileName)

	// 检查本地是否已存在该书籍文件
	if _, err := os.Stat(localPath); err == nil {
		fmt.Printf("书籍文件已存在，跳过下载: %s\n", localPath)
		return nil
	}

	// 执行下载
	if err := client.DownloadFile(remotePath, localPath); err != nil {
		return fmt.Errorf("下载电子书失败: %w", err)
	}

	fmt.Printf("书籍下载完成: %s/%s/%s\n", shelfName, bookID, realFileName)
	return nil
}

// GetWebDAVLogs 获取 WebDAV 交互日志
func (a *App) GetWebDAVLogs() []webdav.LogEntry {
	return webdav.GlobalLogger.GetLogs()
}

// ClearWebDAVLogs 清空 WebDAV 交互日志
func (a *App) ClearWebDAVLogs() {
	webdav.GlobalLogger.Clear()
}

// SaveBookCover 将图片保存为书籍封面
func (a *App) SaveBookCover(shelfName, bookMd5, imageDataBase64 string) error {
	bookDir := filepath.Join(utils.GetBooksDir(), shelfName, bookMd5)
	if err := os.MkdirAll(bookDir, 0755); err != nil {
		return fmt.Errorf("failed to create book directory: %w", err)
	}

	imageData, err := utils.Base64Decode(imageDataBase64)
	if err != nil {
		return fmt.Errorf("failed to decode base64: %w", err)
	}

	coverPath := filepath.Join(bookDir, "cover.png")
	if err := os.WriteFile(coverPath, imageData, 0644); err != nil {
		return fmt.Errorf("failed to write cover file: %w", err)
	}

	cfg, err := webdav.LoadConfig(shelfName)
	if err != nil {
		fmt.Printf("加载 WebDAV 配置失败，跳过封面上传: %v\n", err)
		return nil
	}

	client, err := webdav.NewClientWrapper(cfg)
	if err != nil {
		fmt.Printf("创建 WebDAV 客户端失败，跳过封面上传: %v\n", err)
		return nil
	}

	remotePath := fmt.Sprintf("books/%s/%s/cover.png", shelfName, bookMd5)
	if err := client.UploadFile(coverPath, remotePath); err != nil {
		fmt.Printf("封面上传失败: %v\n", err)
	} else {
		fmt.Printf("封面已上传到云端: %s\n", remotePath)
	}

	return nil
}

// CopyImageToClipboard 复制图片到剪贴板
func (a *App) CopyImageToClipboard(imageURL string) error {
	// 解析图片 URL，格式: /epub-img/{tabId}/{resPath}
	parts := strings.Split(strings.TrimPrefix(imageURL, "/epub-img/"), "/")
	if len(parts) < 2 {
		return fmt.Errorf("invalid image URL format")
	}

	tabId := parts[0]
	resPath := strings.Join(parts[1:], "/")

	// 从 EPUB 文件中读取图片（直接返回二进制数据，无需 Base64）
	imageData, _, err := a.GetEpubImageByPath(tabId, resPath)
	if err != nil {
		return fmt.Errorf("failed to read image: %w", err)
	}

	// 使用系统命令复制图片到剪贴板（直接使用二进制位图数据）
	if runtime.GOOS == "windows" {
		return copyImageToClipboardWindows(imageData)
	}

	return fmt.Errorf("unsupported platform: %s", runtime.GOOS)
}

// copyImageToClipboardWindows 在 Windows 上复制图片到剪贴板
func copyImageToClipboardWindows(imageData []byte) error {
	// 创建临时文件
	tmpFile, err := os.CreateTemp("", "clipboard-*.png")
	if err != nil {
		return fmt.Errorf("failed to create temp file: %w", err)
	}
	defer os.Remove(tmpFile.Name())

	if _, err := tmpFile.Write(imageData); err != nil {
		tmpFile.Close()
		return fmt.Errorf("failed to write temp file: %w", err)
	}
	tmpFile.Close()

	// 使用 PowerShell 复制图片到剪贴板
	psScript := fmt.Sprintf(`
Add-Type -AssemblyName System.Windows.Forms
Add-Type -AssemblyName System.Drawing
$image = [System.Drawing.Image]::FromFile('%s')
[Windows.Forms.Clipboard]::SetImage($image)
$image.Dispose()
`, tmpFile.Name())

	cmd := exec.Command("powershell", "-Command", psScript)
	if err := cmd.Run(); err != nil {
		return fmt.Errorf("failed to copy to clipboard: %w", err)
	}

	return nil
}
