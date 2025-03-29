// Package reply consts： 对客户端的一些固定的回复
package reply

// PongReply 在客户端发送 PING 命令时的回复是固定的 PONG
type PongReply struct{}

// ToBytes 将回复转换为字节数组
func (r *PongReply) ToBytes() []byte {
	return []byte("+PONG\r\n")
}

// MakePongReply 创建一个 PONG 回复
// 这里使用了工厂模式，将 pongReply 的构造函数隐藏起来
func MakePongReply() *PongReply {
	return &PongReply{}
}

// OKReply 在客户端发送 SET 命令时的回复是固定的 OK
type OKReply struct{}

func (r *OKReply) ToBytes() []byte {
	return []byte("+OK\r\n")
}

func MakeOKReply() *OKReply {
	return &OKReply{}
}

// NullBulkReply 空的 Bulk 回复(字符串 nil)
type NullBulkReply struct{}

func (r *NullBulkReply) ToBytes() []byte {
	return []byte("$-1\r\n") // -1，表示 nil 值
}

func MakeNullBulkReply() *NullBulkReply {
	return &NullBulkReply{}
}

// EmptyBulkReply 空的 Bulk 回复(空字符串)
type EmptyBulkReply struct{}

func (r *EmptyBulkReply) ToBytes() []byte {
	return []byte("$0\r\n\r\n") // 0，表示空字符串
}

func MakeEmptyBulkReply() *EmptyBulkReply {
	return &EmptyBulkReply{}
}

// EmptyMultiBulkReply 空的 MultiBulk 回复(空数组)
type EmptyMultiBulkReply struct{}

func (r *EmptyMultiBulkReply) ToBytes() []byte {
	return []byte("*0\r\n")
}

func MakeEmptyMultiBulkReply() *EmptyMultiBulkReply {
	return &EmptyMultiBulkReply{}
}

// NoReply 无回复
type NoReply struct{}

func (r *NoReply) ToBytes() []byte {
	return []byte("")
}

func MakeNoReply() *NoReply {
	return &NoReply{}
}
