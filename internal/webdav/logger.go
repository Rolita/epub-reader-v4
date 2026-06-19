package webdav

import (
	"context"
	"encoding/json"
	"os"
	"path/filepath"
	"sync"
	"time"

	"github.com/wailsapp/wails/v2/pkg/runtime"
)

type LogEntry struct {
	Timestamp string `json:"timestamp"`
	Action    string `json:"action"`
	Result    string `json:"result"`
	Type      string `json:"type"` // "SUCCESS", "ERROR", "SKIP"
}

type Logger struct {
	mu      sync.Mutex
	LogPath string
	ctx     context.Context
}

var GlobalLogger = &Logger{}

func (l *Logger) Init(ctx context.Context, logPath string) {
	l.ctx = ctx
	l.LogPath = logPath
	// 确保目录存在
	dir := filepath.Dir(logPath)
	os.MkdirAll(dir, 0755)
}

func (l *Logger) Add(action, result, entryType string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	// 1. 读取旧日志
	var logs []LogEntry
	data, err := os.ReadFile(l.LogPath)
	if err == nil {
		json.Unmarshal(data, &logs)
	}

	// 2. 追加新日志
	newEntry := LogEntry{
		Timestamp: time.Now().Format("2006-01-02 15:04:05"),
		Action:    action,
		Result:    result,
		Type:      entryType,
	}
	logs = append([]LogEntry{newEntry}, logs...)

	// 3. 超过 200 条则进行截断
	if len(logs) > 200 {
		logs = logs[:200]
	}

	// 4. 保存回 JSON 文件
	newData, _ := json.MarshalIndent(logs, "", "  ")
	os.WriteFile(l.LogPath, newData, 0644)

	// 5. 发送事件通知前端
	if l.ctx != nil {
		runtime.EventsEmit(l.ctx, "webdav-log-updated", newEntry)
	}
}

func (l *Logger) GetLogs() []LogEntry {
	l.mu.Lock()
	defer l.mu.Unlock()

	var logs []LogEntry
	data, err := os.ReadFile(l.LogPath)
	if err == nil {
		json.Unmarshal(data, &logs)
	}
	return logs
}

func (l *Logger) Clear() {
	l.mu.Lock()
	defer l.mu.Unlock()

	os.WriteFile(l.LogPath, []byte("[]"), 0644)
}
