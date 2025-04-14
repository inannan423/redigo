// Package resp Conn: 一个 Redis 的连接
package resp

type Connection interface {
	Write([]byte) error // Write data to the connection
	GetDBIndex() int    // Get database index
	SelectDB(int)       // Select database
}
