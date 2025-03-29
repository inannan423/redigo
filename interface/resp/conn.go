// Package resp Conn: 一个 Redis 的连接
package resp

type Connection interface {
	Write([]byte) error // 写入数据
	GetDBIndex() int    // 获取当前连接的数据库索引
	SelectDB(int)       // 选择数据库
}
