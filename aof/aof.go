package aof

import (
	"io"
	"os"
	"redigo/config"
	"redigo/interface/database"
	"redigo/lib/logger"
	"redigo/lib/utils"
	"redigo/resp/connection"
	"redigo/resp/parser"
	"redigo/resp/reply"
	"strconv"
)

const aofBufferSize = 1 << 16 // 65536 bytes

type CmdLine = [][]byte

type payload struct {
	cmdLine CmdLine
	dbIndex int
}

// AofHandler handles the Append-Only File (AOF) functionality for Redis.
type AofHandler struct {
	db          database.Database
	aofChan     chan *payload
	aofFile     *os.File
	aofFilename string
	currentDB   int
}

// NewAofHandler creates a new AofHandler instance.
func NewAofHandler(db database.Database) (*AofHandler, error) {
	handler := &AofHandler{}
	handler.aofFilename = config.Properties.AppendFilename
	handler.db = db
	// Load the AOF file if it exists
	handler.LoadAof()
	// Open the AOF file for reading and writing
	aofFile, err := os.OpenFile(handler.aofFilename, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0600)
	if err != nil {
		return nil, err
	}
	handler.aofFile = aofFile
	// Make a chan for aof
	handler.aofChan = make(chan *payload, aofBufferSize)
	// Start a goroutine to handle the AOF file writing
	go func() {
		handler.handleAof()
	}()
	return handler, nil
}

// AddAof adds a command line to the AOF file. It will push the command line to the aofChan channel.
func (h *AofHandler) AddAof(dbIndex int, cmdLine CmdLine) {
	if h.aofChan == nil || !config.Properties.AppendOnly {
		h.aofChan = make(chan *payload, 100)
	}
	h.aofChan <- &payload{
		cmdLine: cmdLine,
		dbIndex: dbIndex,
	}
}

// handleAof handles the AOF file writing. It will write the command line to the AOF file.
func (h *AofHandler) handleAof() {
	h.currentDB = 0
	for p := range h.aofChan {
		var dataToWrite []byte

		// 原子性地准备所有要写入的数据
		if p.dbIndex != h.currentDB {
			h.currentDB = p.dbIndex
			// 准备 SELECT 命令数据
			selectData := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			cmdData := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
			// 合并为一次写入
			dataToWrite = append(selectData, cmdData...)
		} else {
			dataToWrite = reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		}

		// 原子性写入
		_, err := h.aofFile.Write(dataToWrite)
		if err != nil {
			logger.Error("AOF write error: " + err.Error())
			continue
		}

		// 确保数据立即刷新到磁盘
		h.aofFile.Sync()
	}
}

// LoadAof loads commands from the AOF file and executes them on the database.
func (h *AofHandler) LoadAof() {
	// Open the AOF file for reading
	aofFile, err := os.Open(h.aofFilename)
	if err != nil {
		logger.Error("AOF file open error: " + err.Error())
		return
	}
	defer aofFile.Close()

	ch := parser.ParseStream(aofFile)
	fakeConn := &connection.Connection{}
	for p := range ch {
		if p.Err != nil {
			// If the error is EOF or unexpected EOF, break the loop
			if p.Err == io.EOF || p.Err == io.ErrUnexpectedEOF {
				// End of file
				break
			}
			// Other errors
			logger.Error("AOF file parse error: " + p.Err.Error())
			continue
		}
		if p.Data == nil {
			logger.Error("AOF file empty payload")
			continue
		}
		// Attempt to parse the payload as a MultiBulkReply
		// If it fails, log an error and continue to the next payload
		r, ok := p.Data.(*reply.MultiBulkReply)
		if !ok {
			logger.Error("AOF file require multi bulk reply")
			continue
		}
		// Execute the command on the database
		rep := h.db.Exec(fakeConn, r.Args)
		if reply.IsErrReply(rep) {
			logger.Error("Execute AOF command error")
		}
	}
}
