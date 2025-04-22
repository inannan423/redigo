package reply

import (
	"bytes"
	"redigo/interface/resp"
	"strconv"
)

var (
	nullBUlkReplyBytes = []byte("$-1") // -1，表示 nil 值
	CRLF               = "\r\n"
)

// ErrorReply 错误回复，实现了 Reply 的 ToBytes 方法，也实现了系统的 error 接口
// 这里使用了接口组合，将 error 接口和 Reply 接口组合在一起
type ErrorReply interface {
	Error() string
	ToBytes() []byte
}

// BulkReply 字符串回复
type BulkReply struct {
	Arg []byte // 回复的内容，此时是不符合 RESP 协议的
}

// ToBytes 将不符合 RESP 协议的字符串转换为符合 RESP 协议的字符串
func (r *BulkReply) ToBytes() []byte {
	// 如果字符串为空，返回空字符串
	if len(r.Arg) == 0 {
		return nullBUlkReplyBytes
	}
	// 将 BulkReply 转换为符合 RESP 协议的字节数组
	return []byte("$" + strconv.Itoa(len(r.Arg)) + CRLF + string(r.Arg) + CRLF)
}

func MakeBulkReply(arg []byte) *BulkReply {
	return &BulkReply{Arg: arg}
}

// MultiBulkReply 多个字符串回复
type MultiBulkReply struct {
	Args [][]byte
}

func (r *MultiBulkReply) ToBytes() []byte {
	argLen := len(r.Args)
	var buf bytes.Buffer
	buf.WriteString("*" + strconv.Itoa(argLen) + CRLF)
	for _, arg := range r.Args {
		if arg == nil {
			// *-1\r\n\r\n 表示空数组
			buf.WriteString(string(nullBUlkReplyBytes) + CRLF)
		} else {
			// *3\r\n$3\r\nfoo\r\n$3\r\nbar\r\n$5\r\nhello\r\n
			buf.WriteString("$" + strconv.Itoa(len(arg)) + CRLF + string(arg) + CRLF)
		}
	}
	return buf.Bytes()
}

func MakeMultiBulkReply(args [][]byte) *MultiBulkReply {
	return &MultiBulkReply{Args: args}
}

// StatusReply 状态回复
type StatusReply struct {
	Status string
}

// MakeStatusReply creates StatusReply
func MakeStatusReply(status string) *StatusReply {
	return &StatusReply{
		Status: status,
	}
}

// ToBytes marshal redis.Reply
func (r *StatusReply) ToBytes() []byte {
	return []byte("+" + r.Status + CRLF)
}

// StandardErrorReply 状态回复(通用错误回复)
type StandardErrorReply struct {
	Status string
}

func (r *StandardErrorReply) ToBytes() []byte {
	return []byte("-" + r.Status + CRLF)
}

func MakeStandardErrorReply(status string) *StandardErrorReply {
	return &StandardErrorReply{Status: status}
}

// Error implements the ErrorReply interface for StandardErrorReply
func (r *StandardErrorReply) Error() string {
	return r.Status
}

// IntReply 整数回复
type IntReply struct {
	Code int64
}

// ToBytes marshal redis.Reply
func (r *IntReply) ToBytes() []byte {
	return []byte(":" + strconv.FormatInt(r.Code, 10) + CRLF)
}

// MakeIntReply creates int reply
func MakeIntReply(code int64) *IntReply {
	return &IntReply{
		Code: code,
	}
}

// IsErrReply 判断是否是“错误回复”
// 如果第一个字符是 -，则表示是错误回复
func IsErrReply(reply resp.Reply) bool {
	return reply.ToBytes()[0] == '-'
}
