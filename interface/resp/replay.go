// Package resp reply: Redis 对客户端的回复
package resp

type Reply interface {
	ToBytes() []byte // 将回复转换为字节数组
}
