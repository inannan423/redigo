package database

import (
	"redigo/aof"
	"redigo/config"
	"redigo/interface/resp"
	"redigo/lib/logger"
	"redigo/resp/reply"
	"strconv"
	"strings"
)

type StandaloneDatabase struct {
	dbSet      []*DB
	aofHandler *aof.AofHandler
}

// NewStandaloneDatabase creates a new StandaloneDatabase instance
func NewStandaloneDatabase() *StandaloneDatabase {
	database := &StandaloneDatabase{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	database.dbSet = make([]*DB, config.Properties.Databases)
	for i := range database.dbSet {
		db := MakeDB()
		db.index = i
		database.dbSet[i] = db
	}

	if config.Properties.AppendOnly {
		aofHandler, err := aof.NewAofHandler(database)
		if err != nil {
			panic(err)
		}
		database.aofHandler = aofHandler
		for _, db := range database.dbSet {
			// create new variable to avoid closure capturing the loop variable
			sdb := db
			sdb.addAof = func(line CmdLine) {
				database.aofHandler.AddAof(sdb.index, line)
			}
		}
	}

	return database
}

// Exec executes a command on the database
func (d *StandaloneDatabase) Exec(client resp.Connection, args [][]byte) resp.Reply {
	defer func() {
		if err := recover(); err != nil {
			logger.Error("Database Exec panic:" + err.(error).Error())
		}
	}()
	cmdName := strings.ToLower(string(args[0]))
	if cmdName == "select" {
		if len(args) != 2 {
			return reply.MakeArgNumErrReply("select")
		}
		return execSelect(client, d, args[1:])
	}
	// Get the current database index from the client connection
	db := d.dbSet[client.GetDBIndex()]
	return db.Exec(client, args)
}

func (d *StandaloneDatabase) AfterClientClose(c resp.Connection) {

}

func (d *StandaloneDatabase) Close() {

}

// execSelect sets the current database for the client connection.
// select x
func execSelect(c resp.Connection, database *StandaloneDatabase, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid DB index")
	}
	if dbIndex < 0 || dbIndex >= len(database.dbSet) {
		return reply.MakeStandardErrorReply("ERR DB index out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeOKReply()
}
