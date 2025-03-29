package database

import "redigo/interface/resp"

// CmdLine 是 [][]byte 类型的别名,方便使用
type CmdLine = [][]byte

// Database 是数据库接口，定义了数据库的基本操作
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	AfterClientClose(c resp.Connection)
	Close()
}

// DataEntity 将数据封装为 DataEntity 类型
type DataEntity struct {
	Data interface{}
}
