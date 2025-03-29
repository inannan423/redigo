package connection

import (
	"net"
	"redigo/lib/sync/wait"
	"sync"
	"time"
)

// Connection 表示客户端和服务端的连接
type Connection struct {
	conn         net.Conn   // 底层的网络连接
	waitingReply wait.Wait  // 等待完成响应的同步器
	mu           sync.Mutex // 发送响应时的互斥锁
	selectedDB   int        // 选择的数据库的编号
}

// NewConnection 创建一个新的连接
func NewConnection(conn net.Conn) *Connection {
	return &Connection{
		conn: conn,
	}
}

// RemoteAddr 返回远程客户端的地址
func (c *Connection) RemoteAddr() net.Addr {
	return c.conn.RemoteAddr()
}

// Close 关闭连接
func (c *Connection) Close() error {
	c.waitingReply.WaitWithTimeout(10 * time.Second)
	_ = c.conn.Close()
	return nil
}

// Write 向客户端发送数据
func (c *Connection) Write(b []byte) error {
	if len(b) == 0 {
		return nil
	}
	c.mu.Lock()
	c.waitingReply.Add(1)
	defer func() {
		c.waitingReply.Done()
		c.mu.Unlock()
	}()

	_, err := c.conn.Write(b)
	return err
}

// GetDBIndex returns selected db
func (c *Connection) GetDBIndex() int {
	return c.selectedDB
}

// SelectDB selects a database
func (c *Connection) SelectDB(dbNum int) {
	c.selectedDB = dbNum
}
