package rdb

import (
	"fmt"
	"os"
	"path/filepath"
	"redigo/config"
	"redigo/interface/database"
	"redigo/interface/resp"
	"redigo/lib/logger"
	"redigo/resp/reply"
	"sync"
	"time"
)

// RDBHandler 处理 RDB 持久化
type RDBHandler struct {
	db           database.Database
	saving       bool
	saveMutex    sync.Mutex
	lastSaveTime time.Time
	saveCount    int // 距离上次 save 后的修改次数
}

// NewRDBHandler 创建一个新的 RDB 处理器
func NewRDBHandler(db database.Database) *RDBHandler {
	return &RDBHandler{
		db:           db,
		saving:       false,
		lastSaveTime: time.Now(),
		saveCount:    0,
	}
}

// Save 执行同步保存（SAVE 命令）
func (h *RDBHandler) Save() resp.Reply {
	h.saveMutex.Lock()
	defer h.saveMutex.Unlock()

	if h.saving {
		return reply.MakeStandardErrorReply("ERR Background save already in progress")
	}

	h.saving = true
	defer func() {
		h.saving = false
	}()

	// 执行保存
	if err := h.performSave(); err != nil {
		logger.Error("RDB SAVE failed: " + err.Error())
		return reply.MakeStandardErrorReply("ERR " + err.Error())
	}

	h.lastSaveTime = time.Now()
	h.saveCount = 0

	return reply.MakeOkReply()
}

// BGSave 执行后台保存（BGSAVE 命令）
func (h *RDBHandler) BGSave() resp.Reply {
	h.saveMutex.Lock()
	if h.saving {
		h.saveMutex.Unlock()
		return reply.MakeStandardErrorReply("ERR Background save already in progress")
	}
	h.saving = true
	h.saveMutex.Unlock()

	// 在后台执行保存
	go func() {
		defer func() {
			h.saveMutex.Lock()
			h.saving = false
			h.saveMutex.Unlock()
		}()

		if err := h.performSave(); err != nil {
			logger.Error("RDB BGSAVE failed: " + err.Error())
			return
		}

		h.saveMutex.Lock()
		h.lastSaveTime = time.Now()
		h.saveCount = 0
		h.saveMutex.Unlock()

		logger.Info("RDB: Background saving terminated with success")
	}()

	return reply.MakeOkReply()
}

// performSave 执行实际的保存操作
func (h *RDBHandler) performSave() error {
	// 确定 RDB 文件路径
	dir := config.Properties.Dir
	if dir == "" {
		dir = "."
	}
	
	filename := config.Properties.DBFilename
	if filename == "" {
		filename = "dump.rdb"
	}
	
	filepath := filepath.Join(dir, filename)
	
	// 创建临时文件
	tmpFile := filepath + ".tmp"
	
	// 执行编码
	if err := SaveToFile(tmpFile, h.db); err != nil {
		os.Remove(tmpFile)
		return err
	}
	
	// 重命名临时文件到目标文件
	if err := os.Rename(tmpFile, filepath); err != nil {
		os.Remove(tmpFile)
		return err
	}
	
	logger.Info("RDB: Saved to " + filepath)
	return nil
}

// Load 从 RDB 文件加载数据
func (h *RDBHandler) Load() error {
	dir := config.Properties.Dir
	if dir == "" {
		dir = "."
	}
	
	filename := config.Properties.DBFilename
	if filename == "" {
		filename = "dump.rdb"
	}
	
	filepath := filepath.Join(dir, filename)
	
	// 检查文件是否存在
	if _, err := os.Stat(filepath); os.IsNotExist(err) {
		logger.Info("RDB: File not found: " + filepath)
		return nil
	}
	
	// 加载文件
	if err := LoadFromFile(filepath, h.db); err != nil {
		return err
	}
	
	logger.Info("RDB: Loaded from " + filepath)
	return nil
}

// ShouldSave 检查是否需要执行自动保存
func (h *RDBHandler) ShouldSave() bool {
	// 检查 save 配置
	if config.Properties.Save == "" {
		return false
	}
	
	// 解析 save 配置（简化实现，支持单个配置如 "900 1"）
	var seconds int
	var changes int
	_, err := fmt.Sscanf(config.Properties.Save, "%d %d", &seconds, &changes)
	if err != nil {
		return false
	}
	
	h.saveMutex.Lock()
	defer h.saveMutex.Unlock()
	
	// 检查是否满足保存条件
	elapsed := time.Since(h.lastSaveTime).Seconds()
	if elapsed >= float64(seconds) && h.saveCount >= changes {
		return true
	}
	
	return false
}

// IncrementSaveCount 增加修改计数
func (h *RDBHandler) IncrementSaveCount() {
	h.saveMutex.Lock()
	h.saveCount++
	h.saveMutex.Unlock()
}

// IsSaving 检查是否正在保存
func (h *RDBHandler) IsSaving() bool {
	h.saveMutex.Lock()
	defer h.saveMutex.Unlock()
	return h.saving
}

// LastSaveTime 返回上次保存时间
func (h *RDBHandler) LastSaveTime() time.Time {
	h.saveMutex.Lock()
	defer h.saveMutex.Unlock()
	return h.lastSaveTime
}

// CmdLine is a type alias for a slice of byte slices
type CmdLine = [][]byte
