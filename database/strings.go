package database

import (
	"redigo/interface/database"
	"redigo/interface/resp"
	"redigo/resp/reply"
)

// execGet retrieves the value associated with the specified key from the database.
func execGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	if entity, ok := db.GetEntity(key); ok {
		// TODO: If we have multiple types, we need to check the conversion if it's not []byte
		return reply.MakeBulkReply(entity.Data.([]byte))
	}
	return reply.MakeNullBulkReply()
}

// execSet stores the specified key-value pair in the database.
func execSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	db.PutEntity(key, entity)
	return reply.MakeOKReply()
}

// execSetNX stores the specified key-value pair in the database only if the key does not already exist.
// If the key already exists, it does not modify the value and returns 0.
// If the key does not exist, it sets the value and returns 1.
// SETNX key value
func execSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]
	entity := &database.DataEntity{
		Data: value,
	}
	result := db.PutIfAbsent(key, entity)
	return reply.MakeIntReply(int64(result))
}

// execGetSet stores the specified key-value pair in the database and returns the old value associated with the key.
func execGetSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	value := args[1]

	entity, ok := db.GetEntity(key)
	db.PutEntity(key, &database.DataEntity{
		Data: value,
	})
	if !ok {
		return reply.MakeNullBulkReply()
	}
	return reply.MakeBulkReply(entity.Data.([]byte))
}

// execStrLen retrieves the length of the value associated with the specified key.
func execStrLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	entity, ok := db.GetEntity(key)
	if !ok {
		return reply.MakeNullBulkReply()
	}
	return reply.MakeIntReply(int64(len(entity.Data.([]byte))))
}

func init() {
	RegisterCommand("GET", execGet, 2)
	RegisterCommand("SET", execSet, 3)
	RegisterCommand("SETNX", execSetNX, 3)
	RegisterCommand("GETSET", execGetSet, 3)
	RegisterCommand("SETEX", execSet, 4)
	RegisterCommand("STRLEN", execStrLen, 2)
}
