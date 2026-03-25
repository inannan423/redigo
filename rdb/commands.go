package rdb

import (
	"redigo/database"
	"redigo/interface/resp"
	"redigo/lib/logger"
	"redigo/resp/reply"
)

// RDBCommandHandler 处理 RDB 相关命令
type RDBCommandHandler struct {
	rdbHandler *RDBHandler
}

// NewRDBCommandHandler 创建 RDB 命令处理器
func NewRDBCommandHandler(rdbHandler *RDBHandler) *RDBCommandHandler {
	return &RDBCommandHandler{
		rdbHandler: rdbHandler,
	}
}

// RegisterCommands 注册 RDB 命令到命令表
func (h *RDBCommandHandler) RegisterCommands() {
	// SAVE 命令
	database.RegisterCommand("save", execSave, 1)
	
	// BGSAVE 命令
	database.RegisterCommand("bgsave", execBGSave, 1)
	
	// LASTSAVE 命令
	database.RegisterCommand("lastsave", execLastSave, 1)
}

// execSave 执行 SAVE 命令
func execSave(db *database.DB, args [][]byte) resp.Reply {
	logger.Info("RDB: SAVE command received")
	// 这里需要获取 RDBHandler 实例，通过其他方式传递
	return reply.MakeStandardErrorReply("ERR SAVE command not properly initialized")
}

// execBGSave 执行 BGSAVE 命令
func execBGSave(db *database.DB, args [][]byte) resp.Reply {
	logger.Info("RDB: BGSAVE command received")
	// 这里需要获取 RDBHandler 实例，通过其他方式传递
	return reply.MakeStandardErrorReply("ERR BGSAVE command not properly initialized")
}

// execLastSave 执行 LASTSAVE 命令，返回上次成功保存的 Unix 时间戳
func execLastSave(db *database.DB, args [][]byte) resp.Reply {
	// 这里需要获取 RDBHandler 实例，通过其他方式传递
	// 暂时返回 0
	return reply.MakeIntReply(0)
}
