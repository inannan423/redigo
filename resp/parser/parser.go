package parser

import (
	"bufio"
	"errors"
	"io"
	"redigo/interface/resp"
	"redigo/lib/logger"
	"redigo/resp/reply"
	"runtime/debug"
	"strconv"
	"strings"
)

// Payload 客户端给服务端发送的数据（解析完成的）
type Payload struct {
	Data resp.Reply // 客户端发给服务端的和服务端发给客户端的数据使用的是一个结构，因此也能用 Reply 接口
	Err  error
}

// Parser 解析器的状态
type readState struct {
	readingMultiLine  bool     // 是否正在读取多行数据
	expectedArgsCount int      // 期望的参数数量
	msgType           byte     // 消息类型
	args              [][]byte // 参数
	bulkLen           int64    // Bulk 回复的长度
}

// finished 检查是否解析完成
func (r *readState) isDone() bool {
	return r.expectedArgsCount > 0 && len(r.args) == r.expectedArgsCount
}

// ParseStream 解析流，将流解析为一个个的 Payload
// 实现并发
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parseIt(reader, ch)
	return ch
}

func parseIt(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			// 打印调用栈信息
			logger.Error(string(debug.Stack()))
		}
	}()

	bufReader := bufio.NewReader(reader) // 读取缓冲区
	var state readState                  // 解析器的状态
	var err error
	var msg []byte

	// 读取数据
	for {
		var ioErr bool // 是否是 IO 错误
		msg, ioErr, err = readLine(bufReader, &state)

		if err != nil {
			// 如果是 IO 错误，关闭通道，退出循环
			if ioErr {
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
			ch <- &Payload{Err: err}
			state = readState{} // 重置状态
			continue            // 继续循环，读取下一行
		}

		// 非多行读取状态
		if !state.readingMultiLine {
			// 多条批量回复
			if msg[0] == '*' {
				// 解析头部，获取期望的参数数量
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: errors.New("Protocol error" + string(msg))}
					state = readState{} // 重置状态
					continue            // 继续循环，读取下一行
				}
				// 需要的参数数量为 0，直接返回
				if state.expectedArgsCount == 0 {
					ch <- &Payload{Data: &reply.EmptyMultiBulkReply{}}
					state = readState{} // 重置状态
					continue            // 继续循环，读取下一行
				}
			} else if msg[0] == '$' {
				// Bulk 回复
				err = parseBulkHeader(msg, &state) // 解析 Bulk 回复的头部，获取 Bulk 回复的长度
				if err != nil {
					ch <- &Payload{Err: errors.New("Protocol error" + string(msg))}
					state = readState{} // 重置状态
					continue            // 继续循环，读取下一行
				}
				if state.bulkLen == -1 {
					// Bulk 回复的长度为 0，直接返回
					ch <- &Payload{Data: &reply.NullBulkReply{}}
					state = readState{} // 重置状态
					continue            // 继续循环，读取下一行
				}
			} else {
				// 单行回复
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{Data: result, Err: err}
				state = readState{} // 本条消息已结束，重置状态
				continue            // 继续循环，读取下一行
			}
		} else {
			err = readBody(msg, &state)
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{} // reset state
				continue
			}
			// 如果解析完成，返回结果
			if state.isDone() {
				var result resp.Reply
				if state.msgType == '*' {
					result = reply.MakeMultiBulkReply(state.args)
				} else if state.msgType == '$' {
					result = reply.MakeBulkReply(state.args[0])
				}
				ch <- &Payload{
					Data: result,
					Err:  err,
				}
				state = readState{}
			}
		}
	}
}

// readLine 读取一行数据
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var line []byte
	var err error
	// 读取普通的行
	if state.bulkLen == 0 {
		line, err = bufReader.ReadBytes('\n')
		if err != nil {
			// 发生错误
			return nil, true, err
		}
		if len(line) == 0 || line[len(line)-2] != '\r' {
			// 不符合 RESP 协议格式
			return nil, false, errors.New("Protocol error: " + string(line))
		}
	} else {
		// 读取 Bulk 回复
		line = make([]byte, state.bulkLen+2) // 2 是 \r\n 的长度
		_, err = io.ReadFull(bufReader, line)
		if err != nil {
			// 发生错误
			return nil, true, err
		}
		if len(line) == 0 || line[len(line)-2] != '\r' || line[len(line)-1] != '\n' {
			// 不符合 RESP 协议格式
			return nil, false, errors.New("Protocol error: " + string(line))
		}
		state.bulkLen = 0
	}
	return line, false, nil
}

func parseMultiBulkHeader(msg []byte, state *readState) error {
	var err error
	var expectedLine uint64
	expectedLine, err = strconv.ParseUint(string(msg[1:len(msg)-2]), 10, 32)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if expectedLine == 0 {
		state.expectedArgsCount = 0
		return nil
	} else if expectedLine > 0 {
		// 多行读取的
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = int(expectedLine)
		state.args = make([][]byte, 0, expectedLine)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

func parseBulkHeader(msg []byte, state *readState) error {
	var err error
	state.bulkLen, err = strconv.ParseInt(string(msg[1:len(msg)-2]), 10, 64)
	if err != nil {
		return errors.New("protocol error: " + string(msg))
	}
	if state.bulkLen == -1 { // null bulk
		return nil
	} else if state.bulkLen > 0 {
		state.msgType = msg[0]
		state.readingMultiLine = true
		state.expectedArgsCount = 1
		state.args = make([][]byte, 0, 1)
		return nil
	} else {
		return errors.New("protocol error: " + string(msg))
	}
}

// parseSingleLineReply 解析单行回复
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply
	switch msg[0] {
	case '+': // status reply
		result = reply.MakeStatusReply(str[1:])
	case '-': // err reply
		result = reply.MakeStandardErrorReply(str[1:])
	case ':': // int reply
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// readBody 读取消息体
func readBody(msg []byte, state *readState) error {
	if len(msg) < 2 {
		return errors.New("protocol error: message too short")
	}
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' {
		// bulk reply
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // null bulk in multi bulks
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
