package reply

// UnknownReply 未知错误回复
type UnknownReply struct{}

func (r *UnknownReply) Error() string {
	return "Unknown Error"
}

func (r *UnknownReply) ToBytes() []byte {
	return []byte("-ERR unknown\r\n")
}

func MakeUnknownReply() *UnknownReply {
	return &UnknownReply{}
}

// ArgNumErrReply 参数数量错误回复
type ArgNumErrReply struct {
	Cmd string // 提示用户使用的哪个命令有问题
}

func (r *ArgNumErrReply) Error() string {
	return "ERR wrong number of arguments for '" + r.Cmd + "' command"
}

func (r *ArgNumErrReply) ToBytes() []byte {
	return []byte("-ERR wrong number of arguments for '" + r.Cmd + "' command\r\n")
}

func MakeArgNumErrReply(cmd string) *ArgNumErrReply {
	return &ArgNumErrReply{Cmd: cmd}
}

// SyntaxErrReply 语法错误回复
type SyntaxErrReply struct{}

func (r *SyntaxErrReply) Error() string {
	return "ERR syntax error"
}

func (r *SyntaxErrReply) ToBytes() []byte {
	return []byte("-ERR syntax error\r\n")
}

func MakeSyntaxErrReply() *SyntaxErrReply {
	return &SyntaxErrReply{}
}

// WrongTypeErrReply 类型错误回复
type WrongTypeErrReply struct{}

func (r *WrongTypeErrReply) Error() string {
	return "WRONG TYPE Operation against a key holding the wrong kind of value"
}

func (r *WrongTypeErrReply) ToBytes() []byte {
	return []byte("-WRONG TYPE Operation against a key holding the wrong kind of value\r\n")
}

func MakeWrongTypeErrReply() *WrongTypeErrReply {
	return &WrongTypeErrReply{}
}

// ProtocolErrReply 协议错误回复
type ProtocolErrReply struct {
	Msg string
}

func (r *ProtocolErrReply) Error() string {
	return "PROTOCOL ERROR: " + r.Msg
}

func (r *ProtocolErrReply) ToBytes() []byte {
	return []byte("-PROTOCOL ERROR: " + r.Msg + "\r\n")
}

func MakeProtocolErrReply(msg string) *ProtocolErrReply {
	return &ProtocolErrReply{Msg: msg}
}
