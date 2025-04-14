package database

import (
	"redigo/interface/resp"
	"redigo/resp/reply"
)

// Ping responds to the PING command
func Ping(db *DB, args [][]byte) resp.Reply {
	return reply.MakePongReply()
}

// Register the PING command to the command table
func init() {
	// Register the PING command with the command table
	RegisterCommand("ping", Ping, 1)
}
