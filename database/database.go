package database

import (
	"redigo/config"
	"redigo/interface/resp"
	"redigo/lib/logger"
	"redigo/resp/reply"
	"strconv"
	"strings"
)

type Database struct {
	dbSet []*DB
}

// NewDatabase creates a new Database instance
func NewDatabase() *Database {
	database := &Database{}
	if config.Properties.Databases == 0 {
		config.Properties.Databases = 16
	}
	database.dbSet = make([]*DB, config.Properties.Databases)
	for i := range database.dbSet {
		db := MakeDB()
		db.index = i
		database.dbSet[i] = db
	}
	return database
}

// Exec executes a command on the database
func (d *Database) Exec(client resp.Connection, args [][]byte) resp.Reply {
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

func (d *Database) AfterClientClose(c resp.Connection) {

}

func (d *Database) Close() {

}

// execSelect sets the current database for the client connection.
// select x
func execSelect(c resp.Connection, database *Database, args [][]byte) resp.Reply {
	dbIndex, err := strconv.Atoi(string(args[0]))
	if err != nil {
		return reply.MakeStandardErrorReply("ERR invalid DB index")
	}
	if dbIndex < 0 || dbIndex >= len(database.dbSet) {
		return reply.MakeStandardErrorReply("ERR DB index out of range")
	}
	c.SelectDB(dbIndex)
	return reply.MakeIntReply(int64(dbIndex))
}
