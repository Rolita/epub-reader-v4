package webdav

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/studio-b12/gowebdav"

	"epub-reader/internal/shelf"
)

// ClientWrapper 封装 gowebdav.Client，提供高级同步功能
type ClientWrapper struct {
	client *gowebdav.Client
	cfg    *Config
}

// DeleteRemote 在 WebDAV 上删除指定路径的文件或文件夹
func (c *ClientWrapper) DeleteRemote(remotePath string) error {
	err := c.client.Remove(remotePath)
	if err != nil {
		GlobalLogger.Add("删除文件: "+remotePath, "失败: "+err.Error(), "ERROR")
	} else {
		GlobalLogger.Add("删除文件: "+remotePath, "成功", "SUCCESS")
	}
	return err
}

// DeleteRemoteBook 删除远程书籍文件
func (c *ClientWrapper) DeleteRemoteBook(shelfName, bookID string) error {
	remotePath := "books/" + shelfName + "/" + bookID + ".epub"
	return c.client.Remove(remotePath)
}

// NewClientWrapper 创建客户端包装器
func NewClientWrapper(cfg *Config) (*ClientWrapper, error) {
	fullURL := buildFullURL(cfg)
	client := gowebdav.NewClient(fullURL, cfg.Username, cfg.Password)

	// 确保连接正常
	if _, err := client.ReadDir("."); err != nil {
		// 尝试创建目录
		if err := client.MkdirAll("/", 0755); err != nil {
			return nil, fmt.Errorf("无法连接或创建远程路径 (%s): %w", fullURL, err)
		}
	}

	return &ClientWrapper{client: client, cfg: cfg}, nil
}

// UploadFile 将本地文件上传到 WebDAV 指定远程路径
func (c *ClientWrapper) UploadFile(localPath, remotePath string) error {
	file, err := os.Open(localPath)
	if err != nil {
		GlobalLogger.Add("上传文件: "+remotePath, "失败: "+err.Error(), "ERROR")
		return fmt.Errorf("打开本地文件失败: %w", err)
	}
	defer file.Close()

	err = c.client.WriteStream(remotePath, file, 0644)
	if err != nil {
		GlobalLogger.Add("上传文件: "+remotePath, "失败: "+err.Error(), "ERROR")
	} else {
		GlobalLogger.Add("上传文件: "+remotePath, "成功", "SUCCESS")
	}
	return err
}

// DownloadFile 将远程文件下载到本地（使用临时文件保证原子性）
func (c *ClientWrapper) DownloadFile(remotePath, localPath string) error {
	data, err := c.client.ReadStream(remotePath)
	if err != nil {
		GlobalLogger.Add("下载文件: "+remotePath, "失败: "+err.Error(), "ERROR")
		return fmt.Errorf("读取远程文件失败: %w", err)
	}
	defer data.Close()

	// 确保本地目录存在
	localDir := filepath.Dir(localPath)
	if err := os.MkdirAll(localDir, 0755); err != nil {
		GlobalLogger.Add("下载文件: "+remotePath, "失败: 创建目录失败", "ERROR")
		return fmt.Errorf("创建本地目录失败: %w", err)
	}

	// 使用临时文件，保证下载中途不会读取到损坏的文件
	tmpPath := localPath + ".tmp"
	file, err := os.Create(tmpPath)
	if err != nil {
		GlobalLogger.Add("下载文件: "+remotePath, "失败: 创建临时文件失败", "ERROR")
		return fmt.Errorf("创建临时文件失败: %w", err)
	}

	_, err = io.Copy(file, data)
	file.Close()
	if err != nil {
		os.Remove(tmpPath) // 下载失败，删除临时文件
		GlobalLogger.Add("下载文件: "+remotePath, "失败: "+err.Error(), "ERROR")
		return fmt.Errorf("写入本地文件失败: %w", err)
	}

	// 下载成功后，原子性地重命名临时文件
	if err := os.Rename(tmpPath, localPath); err != nil {
		os.Remove(tmpPath)
		GlobalLogger.Add("下载文件: "+remotePath, "失败: 重命名文件失败", "ERROR")
		return fmt.Errorf("重命名文件失败: %w", err)
	}

	GlobalLogger.Add("下载文件: "+remotePath, "成功", "SUCCESS")
	return nil
}

// FileExists 检查远程文件是否存在
func (c *ClientWrapper) FileExists(remotePath string) (bool, error) {
	_, err := c.client.Stat(remotePath)
	if err != nil {
		if os.IsNotExist(err) {
			return false, nil
		}
		return false, err
	}
	return true, nil
}

// GetRemoteFileTime 获取远程文件修改时间
func (c *ClientWrapper) GetRemoteFileTime(remotePath string) (time.Time, error) {
	stat, err := c.client.Stat(remotePath)
	if err != nil {
		return time.Time{}, err
	}
	return stat.ModTime(), nil
}

// BackupRemoteFile 备份远程文件
func (c *ClientWrapper) BackupRemoteFile(remotePath string) error {
	backupPath := remotePath + ".bak"

	// 删除旧备份
	c.client.Remove(backupPath)

	// 复制文件到备份
	data, err := c.client.ReadStream(remotePath)
	if err != nil {
		return err
	}
	defer data.Close()

	return c.client.WriteStream(backupPath, data, 0644)
}

// RestoreBackupFile 从备份恢复文件
func (c *ClientWrapper) RestoreBackupFile(remotePath string) error {
	c.client.Remove(remotePath)
	return c.client.Rename(remotePath+".bak", remotePath, false)
}

// UploadDir 递归上传本地文件夹（带时间戳智能过滤）
func (c *ClientWrapper) UploadDir(localDir, remoteDir string) error {
	return filepath.Walk(localDir, func(path string, info os.FileInfo, err error) error {
		if err != nil {
			return err
		}

		// 计算相对于 localDir 的相对路径
		relPath, _ := filepath.Rel(localDir, path)
		if relPath == "." {
			return nil // 跳过根目录本身
		}

		// 将 Windows 的反斜杠转为 WebDAV 的正斜杠
		remotePath := strings.ReplaceAll(filepath.Join(remoteDir, relPath), "\\", "/")

		if info.IsDir() {
			// 在 WebDAV 上创建对应的目录
			return c.client.MkdirAll(remotePath, 0755)
		} else {
			// --- 【修正版】检查远程文件信息 ---
			remoteInfo, err := c.client.Stat(remotePath)

			// 如果 err != nil，说明 Stat 失败（通常是 404），必须上传！
			if err != nil {
				fmt.Printf("远程不存在该文件，准备上传: %s\n", relPath)
			} else {
				// 检查远程时间戳是否为零值（某些服务器对不存在的文件返回零值而非错误）
				if remoteInfo.ModTime().Unix() == 0 {
					fmt.Printf("远程文件时间戳为零值，准备上传: %s\n", relPath)
				} else {
					// 如果没有 err，再进行时间戳对比
					if info.ModTime().Before(remoteInfo.ModTime().Add(2 * time.Second)) {
						fmt.Printf("跳过文件 (时间戳判断无需上传): %s (本地:%d, 远程:%d)\n", relPath, info.ModTime().Unix(), remoteInfo.ModTime().Unix())
						GlobalLogger.Add("上传文件: "+remotePath, "文件未变，已跳过", "SKIP")
						return nil
					}
				}
			}

			// 执行上传
			file, err := os.Open(path)
			if err != nil {
				return err
			}
			defer file.Close()
			fmt.Println("正在上传:", relPath)
			return c.client.WriteStream(remotePath, file, 0644)
		}
	})
}

// DownloadDir 递归下载远程文件夹到本地
func (c *ClientWrapper) DownloadDir(remoteDir, localDir string) error {
	// 获取远程目录列表
	files, err := c.client.ReadDir(remoteDir)
	if err != nil {
		return err
	}

	// 确保本地目录存在
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

	for _, file := range files {
		remotePath := strings.ReplaceAll(filepath.Join(remoteDir, file.Name()), "\\", "/")
		localPath := filepath.Join(localDir, file.Name())

		if file.IsDir() {
			// 递归下载子目录
			if err := c.DownloadDir(remotePath, localPath); err != nil {
				return err
			}
		} else {
			// 下载文件
			if err := c.DownloadFile(remotePath, localPath); err != nil {
				return err
			}
		}
	}

	return nil
}

// DownloadDirExcluding 递归下载目录，跳过指定后缀的文件
func (c *ClientWrapper) DownloadDirExcluding(remoteDir, localDir string, excludeExt string) error {
	// 获取远程目录列表
	files, err := c.client.ReadDir(remoteDir)
	if err != nil {
		return err
	}

	// 确保本地目录存在
	if err := os.MkdirAll(localDir, 0755); err != nil {
		return err
	}

	for _, file := range files {
		remotePath := strings.ReplaceAll(filepath.Join(remoteDir, file.Name()), "\\", "/")
		localPath := filepath.Join(localDir, file.Name())

		if file.IsDir() {
			// 递归下载子目录（仍然排除指定后缀）
			if err := c.DownloadDirExcluding(remotePath, localPath, excludeExt); err != nil {
				return err
			}
		} else {
			// 如果是文件，判断是否需要跳过
			if strings.HasSuffix(strings.ToLower(file.Name()), strings.ToLower(excludeExt)) {
				fmt.Printf("跳过文件: %s\n", file.Name())
				continue
			}

			// 执行下载
			fmt.Printf("下载文件: %s\n", file.Name())
			if err := c.DownloadFile(remotePath, localPath); err != nil {
				fmt.Printf("下载失败 [%s]: %v\n", file.Name(), err)
			}
		}
	}

	return nil
}

// SyncLibrary 同步书架库文件
// 返回值：true 表示上传，false 表示下载，error 表示失败
func (c *ClientWrapper) SyncLibrary(localLibPath string) (bool, error) {
	remotePath := "_library.json"

	// 检查远程文件是否存在
	exists, err := c.FileExists(remotePath)
	if err != nil {
		return false, fmt.Errorf("检查远程文件失败: %w", err)
	}

	// 获取本地文件修改时间
	localFileInfo, err := os.Stat(localLibPath)
	if err != nil {
		return false, fmt.Errorf("获取本地文件信息失败: %w", err)
	}
	localTime := localFileInfo.ModTime()

	if !exists {
		// 远程文件不存在，直接上传
		if err := c.UploadFile(localLibPath, remotePath); err != nil {
			return false, fmt.Errorf("上传库文件失败: %w", err)
		}
		return true, nil
	}

	// 远程文件存在，比较时间戳
	remoteTime, err := c.GetRemoteFileTime(remotePath)
	if err != nil {
		return false, fmt.Errorf("获取远程文件时间失败: %w", err)
	}

	if localTime.After(remoteTime) {
		// 本地更新，上传（带备份）
		if err := c.BackupRemoteFile(remotePath); err != nil {
			return false, fmt.Errorf("备份远程文件失败: %w", err)
		}

		if err := c.UploadFile(localLibPath, remotePath); err != nil {
			// 上传失败，恢复备份
			c.RestoreBackupFile(remotePath)
			return false, fmt.Errorf("上传库文件失败: %w", err)
		}

		// 上传成功，删除备份
		c.client.Remove(remotePath + ".bak")
		return true, nil
	} else if remoteTime.After(localTime) {
		// 远程更新，下载
		if err := c.DownloadFile(remotePath, localLibPath); err != nil {
			return false, fmt.Errorf("下载库文件失败: %w", err)
		}
		return false, nil
	}

	// 时间相同，无需同步
	return false, nil
}

// SyncBookFile 同步书籍文件
func (c *ClientWrapper) SyncBookFile(localPath, remoteSubPath string) error {
	remotePath := filepath.Join("books", remoteSubPath)

	// 检查远程文件是否存在
	exists, err := c.FileExists(remotePath)
	if err != nil {
		return fmt.Errorf("检查远程文件失败: %w", err)
	}

	if exists {
		// 获取时间戳比较
		localInfo, err := os.Stat(localPath)
		if err != nil {
			return fmt.Errorf("获取本地文件信息失败: %w", err)
		}

		remoteTime, err := c.GetRemoteFileTime(remotePath)
		if err != nil {
			return fmt.Errorf("获取远程文件时间失败: %w", err)
		}

		if localInfo.ModTime().Before(remoteTime) {
			// 远程更新，下载
			return c.DownloadFile(remotePath, localPath)
		}
		// 本地更新或相同，不上传（书籍文件通常只在添加时上传一次）
		return nil
	}

	// 远程不存在，上传
	return c.UploadFile(localPath, remotePath)
}

// SmartUploadLibrary 智能上传书架库（基于时间戳对比）
// 返回值：true 表示执行了上传，false 表示无需同步
func (c *ClientWrapper) SmartUploadLibrary(localLibPath, remotePath string) (bool, error) {
	// 1. 加载本地库
	localData, err := os.ReadFile(localLibPath)
	if err != nil {
		return false, fmt.Errorf("读取本地库失败: %w", err)
	}

	var localLib shelf.Library
	if err := json.Unmarshal(localData, &localLib); err != nil {
		return false, fmt.Errorf("解析本地库失败: %w", err)
	}

	// 2. 尝试读取远程库
	remoteData, err := c.client.ReadStream(remotePath)
	if err != nil {
		// 远程文件不存在，直接上传
		return c.forceUploadLibrary(localLibPath, remotePath)
	}
	defer remoteData.Close()

	var remoteLib shelf.Library
	if err := json.NewDecoder(remoteData).Decode(&remoteLib); err != nil {
		// 远程 JSON 解析失败，直接上传
		return c.forceUploadLibrary(localLibPath, remotePath)
	}

	// 3. 核心对比逻辑：如果本地时间戳不大于远程，则跳过
	if localLib.Metadata.LastSynced <= remoteLib.Metadata.LastSynced {
		fmt.Printf("无需同步：本地时间戳(%d) <= 远程时间戳(%d)\n", localLib.Metadata.LastSynced, remoteLib.Metadata.LastSynced)
		return false, nil
	}

	// 4. 执行上传
	return c.forceUploadLibrary(localLibPath, remotePath)
}

// forceUploadLibrary 强制上传库文件
func (c *ClientWrapper) forceUploadLibrary(localLibPath, remotePath string) (bool, error) {
	// 先备份远程文件
	backupPath := remotePath + ".bak"
	c.client.Remove(backupPath)

	// 复制到备份
	backupData, err := c.client.ReadStream(remotePath)
	if err == nil {
		defer backupData.Close()
		c.client.WriteStream(backupPath, backupData, 0644)
	}

	// 上传新文件
	file, err := os.Open(localLibPath)
	if err != nil {
		return false, err
	}
	defer file.Close()

	if err := c.client.WriteStream(remotePath, file, 0644); err != nil {
		// 上传失败，恢复备份
		c.client.Remove(remotePath)
		c.client.Rename(backupPath, remotePath, false)
		return false, fmt.Errorf("上传失败，已恢复备份: %w", err)
	}

	// 上传成功，删除备份
	c.client.Remove(backupPath)

	fmt.Printf("已上传库文件: %s -> %s\n", localLibPath, remotePath)
	return true, nil
}

// SmartDownloadLibrary 智能下载书架库（基于时间戳对比）
// 返回值：true 表示执行了下载，false 表示无需同步
func (c *ClientWrapper) SmartDownloadLibrary(remotePath, localLibPath string) (bool, error) {
	// 1. 尝试读取远程库
	remoteData, err := c.client.ReadStream(remotePath)
	if err != nil {
		return false, fmt.Errorf("读取远程库失败: %w", err)
	}
	defer remoteData.Close()

	var remoteLib shelf.Library
	if err := json.NewDecoder(remoteData).Decode(&remoteLib); err != nil {
		return false, fmt.Errorf("解析远程库失败: %w", err)
	}

	// 2. 加载本地库（如果存在）
	var localLib shelf.Library
	localData, err := os.ReadFile(localLibPath)
	if err == nil {
		json.Unmarshal(localData, &localLib)
	}

	// 3. 核心对比逻辑：如果远程时间戳不大于本地，则跳过
	if remoteLib.Metadata.LastSynced <= localLib.Metadata.LastSynced {
		fmt.Printf("无需同步：远程时间戳(%d) <= 本地时间戳(%d)\n", remoteLib.Metadata.LastSynced, localLib.Metadata.LastSynced)
		return false, nil
	}

	// 4. 执行下载（带临时文件保护）
	tmpPath := localLibPath + ".tmp"
	if err := c.DownloadFile(remotePath, tmpPath); err != nil {
		os.Remove(tmpPath)
		return false, fmt.Errorf("下载失败: %w", err)
	}

	// 原子替换
	if err := os.Rename(tmpPath, localLibPath); err != nil {
		os.Remove(tmpPath)
		return false, fmt.Errorf("替换本地文件失败: %w", err)
	}

	fmt.Printf("已下载库文件: %s -> %s\n", remotePath, localLibPath)
	return true, nil
}
