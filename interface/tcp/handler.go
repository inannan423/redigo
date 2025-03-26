package tcp

import (
	"context"
	"net"
)

// Handler 提供处理 TCP 连接的接口
type Handler interface {
	Handle(ctx context.Context, conn net.Conn)
	Close() error
}
