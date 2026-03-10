package rdb

import (
	"redigo/database"
	"redigo/interface/resp"
	"redigo/lib/logger"
	"redigo/resp/reply"
	"sync"
)

// 全局 RDB 处理器
var (
	globalRDBHandler *RDBHandler
	rdbOnce          sync.Once
)

// InitRDB 初始化 RDB 处理器
func InitRDB(db database.Database) *RDBHandler {
	rdbOnce.Do(func() {
		globalRDBHandler = NewRDBHandler(db)
		
		// 注册 RDB 命令
		registerRDBCommands()
		
		// 尝试加载 RDB 文件
		if err := globalRDBHandler.Load(); err != nil {
			logger.Error("RDB: Failed to load: " + err.Error())
		}
	})
	
	return globalRDBHandler
}

// GetRDBHandler 获取全局 RDB 处理器
func GetRDBHandler() *RDBHandler {
	return globalRDBHandler
}

// 注册 RDB 相关命令
func registerRDBCommands() {
	// SAVE 命令
	database.RegisterCommand("save", func(db *database.DB, args [][]byte) resp.Reply {
		logger.Info("RDB: SAVE command received")
		if handler := GetRDBHandler(); handler != nil {
			return handler.Save()
		}
		return reply.MakeStandardErrorReply("ERR RDB not initialized")
	}, 1)
	
	// BGSAVE 命令
	database.RegisterCommand("bgsave", func(db *database.DB, args [][]byte) resp.Reply {
		logger.Info("RDB: BGSAVE command received")
		if handler := GetRDBHandler(); handler != nil {
			return handler.BGSave()
		}
		return reply.MakeStandardErrorReply("ERR RDB not initialized")
	}, 1)
	
	// LASTSAVE 命令
	database.RegisterCommand("lastsave", func(db *database.DB, args [][]byte) resp.Reply {
		if handler := GetRDBHandler(); handler != nil {
			return reply.MakeIntReply(handler.LastSaveTime().Unix())
		}
		return reply.MakeIntReply(0)
	}, 1)
}
