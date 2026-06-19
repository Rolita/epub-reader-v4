package webdav

import (
	"fmt"
	"strings"

	"github.com/studio-b12/gowebdav"
)

// buildFullURL 将 BaseURL 和 RemotePath 合并为标准化的完整路径
func buildFullURL(cfg *Config) string {
	base := strings.TrimRight(cfg.BaseURL, "/") + "/"
	remote := strings.Trim(cfg.RemotePath, "/")
	if remote == "" {
		return base
	}
	return base + remote + "/"
}

// EnsureClient 连接并确保远程路径存在，如果不存在则自动创建
func EnsureClient(cfg *Config) (*gowebdav.Client, error) {
	fullURL := buildFullURL(cfg)
	client := gowebdav.NewClient(fullURL, cfg.Username, cfg.Password)

	fmt.Printf("WebDAV 连接尝试: %s\n", fullURL)

	// 1. 尝试探测目录是否存在
	_, err := client.ReadDir(".")
	if err != nil {
		fmt.Printf("读取目录失败: %v\n", err)
		// 2. 如果失败，使用绝对路径 "/" 来创建
		// 在 gowebdav 中，MkdirAll 操作的是当前 client 所属的 path
		// 使用 "/" 强制在当前挂载点根部尝试创建
		err = client.MkdirAll("/", 0755)
		if err != nil {
			return nil, fmt.Errorf("路径不存在且无法创建 (URL: %s): %w", fullURL, err)
		}
		fmt.Println("远程路径已确保存在:", fullURL)
	}

	return client, nil
}

// NewClient 仅执行连接探测，不强制创建路径
func NewClient(cfg *Config) (*gowebdav.Client, error) {
	fullURL := buildFullURL(cfg)
	client := gowebdav.NewClient(fullURL, cfg.Username, cfg.Password)

	if _, err := client.ReadDir("."); err != nil {
		return nil, fmt.Errorf("WebDAV 连接失败 (URL: %s): %w", fullURL, err)
	}

	return client, nil
}

// TestConnection 测试连接并返回详细信息
func TestConnection(cfg *Config) (string, error) {
	fullURL := buildFullURL(cfg)
	client := gowebdav.NewClient(fullURL, cfg.Username, cfg.Password)

	fmt.Printf("WebDAV 测试连接: %s\n", fullURL)

	// 尝试读取根目录
	_, err := client.ReadDir(".")
	if err != nil {
		fmt.Printf("读取目录错误: %v\n", err)
		// 尝试创建目录，使用绝对路径 "/"
		createErr := client.MkdirAll("/", 0755)
		if createErr != nil {
			fmt.Printf("创建目录错误: %v\n", createErr)
			return fmt.Sprintf("连接失败\nURL: %s\n用户名: %s\n读取错误: %v\n创建错误: %v", fullURL, cfg.Username, err, createErr), createErr
		}
		return fmt.Sprintf("已创建远程目录: %s", fullURL), nil
	}

	return fmt.Sprintf("连接成功\n远程目录已存在: %s", fullURL), nil
}
