package aof

import (
	"os"
	"redigo/config"
	"redigo/interface/database"
	"redigo/lib/logger"
	"redigo/lib/utils"
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
		// Write the SELECT command if the database index has changed
		if p.dbIndex != h.currentDB {
			h.currentDB = p.dbIndex
			// Write the SELECT command to the AOF file
			data := reply.MakeMultiBulkReply(utils.ToCmdLine("SELECT", strconv.Itoa(p.dbIndex))).ToBytes()
			_, err := h.aofFile.Write(data)
			if err != nil {
				logger.Error("AOF write error: " + err.Error())
				// Continue to the next command
				continue
			}
		}

		// Write the command line to the AOF file
		data := reply.MakeMultiBulkReply(p.cmdLine).ToBytes()
		_, err := h.aofFile.Write(data)
		if err != nil {
			logger.Error("AOF write error: " + err.Error())
			// Continue to the next command
			continue
		}
	}
}

// LoadAof loads commands from the AOF file and executes them on the database.
func (h *AofHandler) LoadAof() {
	// TODO: Implement loading commands from AOF file
}
