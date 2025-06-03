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

// Payload represents the data sent from the client to the server (parsed data)
type Payload struct {
	Data resp.Reply // The data sent between the client and server uses the Reply interface
	Err  error
}

// readState represents the state of the parser
type readState struct {
	readingMultiLine  bool     // Whether it is reading multi-line data
	expectedArgsCount int      // Expected number of arguments
	msgType           byte     // Message type
	args              [][]byte // Arguments
	bulkLen           int64    // Length of Bulk reply
}

// isDone checks if parsing is complete
func (r *readState) isDone() bool {
	return r.expectedArgsCount > 0 && len(r.args) == r.expectedArgsCount
}

// ParseStream parses the stream into individual Payloads
// Implements concurrency
func ParseStream(reader io.Reader) <-chan *Payload {
	ch := make(chan *Payload)
	go parseIt(reader, ch)
	return ch
}

// parseIt parses the input stream and sends Payloads to the channel
func parseIt(reader io.Reader, ch chan<- *Payload) {
	defer func() {
		if err := recover(); err != nil {
			// Print stack trace information
			logger.Error(string(debug.Stack()))
		}
	}()

	bufReader := bufio.NewReader(reader) // Buffered reader
	var state readState                  // Parser state
	var err error
	var msg []byte

	// Read data
	for {
		var ioErr bool // Whether it is an IO error
		msg, ioErr, err = readLine(bufReader, &state)
		if err != nil {
			// If it is an IO error, close the channel and exit the loop
			if ioErr {
				ch <- &Payload{Err: err}
				close(ch)
				return
			}
			ch <- &Payload{Err: err}
			state = readState{} // Reset state
			continue            // Continue the loop to read the next line
		}

		// Non-multi-line reading state
		if !state.readingMultiLine {
			// Multi-bulk reply
			if msg[0] == '*' {
				// Parse the header to get the expected number of arguments
				err = parseMultiBulkHeader(msg, &state)
				if err != nil {
					ch <- &Payload{Err: errors.New("Protocol error" + string(msg))}
					state = readState{} // Reset state
					continue            // Continue the loop to read the next line
				}
				// If the expected number of arguments is 0, return directly
				if state.expectedArgsCount == 0 {
					ch <- &Payload{Data: &reply.EmptyMultiBulkReply{}}
					state = readState{} // Reset state
					continue            // Continue the loop to read the next line
				}
			} else if msg[0] == '$' {
				// Bulk reply
				err = parseBulkHeader(msg, &state) // Parse the Bulk reply header to get the length
				if err != nil {
					ch <- &Payload{Err: errors.New("Protocol error" + string(msg))}
					state = readState{} // Reset state
					continue            // Continue the loop to read the next line
				}
				if state.bulkLen == -1 {
					// If the length of the Bulk reply is 0, return directly
					ch <- &Payload{Data: &reply.NullBulkReply{}}
					state = readState{} // Reset state
					continue            // Continue the loop to read the next line
				}
			} else {
				// Single-line reply
				result, err := parseSingleLineReply(msg)
				ch <- &Payload{Data: result, Err: err}
				state = readState{} // This message is complete, reset state
				continue            // Continue the loop to read the next line
			}
		} else {
			err = readBody(msg, &state)
			if err != nil {
				ch <- &Payload{
					Err: errors.New("protocol error: " + string(msg)),
				}
				state = readState{} // Reset state
				continue
			}
			// If parsing is complete, return the result
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

// readLine reads a line of data
func readLine(bufReader *bufio.Reader, state *readState) ([]byte, bool, error) {
	var line []byte
	var err error
	// Read a normal line
	if state.bulkLen == 0 {
		line, err = bufReader.ReadBytes('\n')
		if err != nil {
			// An error occurred
			return nil, true, err
		}
		if len(line) == 0 || line[len(line)-2] != '\r' {
			// Does not conform to RESP protocol format
			return nil, false, errors.New("Protocol error: " + string(line))
		}
	} else {
		// Read Bulk reply
		line = make([]byte, state.bulkLen+2) // 2 is the length of \r\n
		_, err = io.ReadFull(bufReader, line)
		if err != nil {
			// An error occurred
			return nil, true, err
		}
		if len(line) == 0 || line[len(line)-2] != '\r' || line[len(line)-1] != '\n' {
			// Does not conform to RESP protocol format
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
		// Multi-line reading
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
	if state.bulkLen == -1 { // Null bulk
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

// parseSingleLineReply parses a single-line reply
func parseSingleLineReply(msg []byte) (resp.Reply, error) {
	str := strings.TrimSuffix(string(msg), "\r\n")
	var result resp.Reply
	switch msg[0] {
	case '+': // Status reply
		result = reply.MakeStatusReply(str[1:])
	case '-': // Error reply
		result = reply.MakeStandardErrorReply(str[1:])
	case ':': // Integer reply
		val, err := strconv.ParseInt(str[1:], 10, 64)
		if err != nil {
			return nil, errors.New("protocol error: " + string(msg))
		}
		result = reply.MakeIntReply(val)
	}
	return result, nil
}

// readBody reads the message body
func readBody(msg []byte, state *readState) error {
	if len(msg) < 2 {
		return errors.New("protocol error: message too short")
	}
	line := msg[0 : len(msg)-2]
	var err error
	if line[0] == '$' {
		// Bulk reply
		state.bulkLen, err = strconv.ParseInt(string(line[1:]), 10, 64)
		if err != nil {
			return errors.New("protocol error: " + string(msg))
		}
		if state.bulkLen <= 0 { // Null bulk in multi-bulks
			state.args = append(state.args, []byte{})
			state.bulkLen = 0
		}
	} else {
		state.args = append(state.args, line)
	}
	return nil
}
