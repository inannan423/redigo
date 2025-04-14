package database

import "redigo/interface/resp"

// CmdLine is a type alias for a slice of byte slices
type CmdLine = [][]byte

// Database is an interface that defines the methods for a database
type Database interface {
	Exec(client resp.Connection, args [][]byte) resp.Reply
	AfterClientClose(c resp.Connection)
	Close()
}

// DataEntity 将数据封装为 DataEntity 类型
type DataEntity struct {
	Data interface{}
}
