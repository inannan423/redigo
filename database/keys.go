package database

import (
	"redigo/interface/resp"
	"redigo/lib/utils"
	"redigo/lib/wildcard"
	"redigo/resp/reply"
)

// Handle the DEL command.
// It deletes the specified keys from the database
func execDel(db *DB, args [][]byte) resp.Reply {
	keys := make([]string, len(args))
	for i, arg := range args {
		keys[i] = string(arg)
	}
	deleted := db.Removes(keys...)
	if deleted > 0 {
		db.addAof(utils.ToCmdLineWithName("DEL", args...))
	}
	return reply.MakeIntReply(int64(deleted))
}

// Handle the EXISTS command.
// It checks if the specified keys exist in the database
func execExists(db *DB, args [][]byte) resp.Reply {
	result := int64(0)
	for _, arg := range args {
		key := string(arg)
		if _, ok := db.GetEntity(key); ok {
			result++
		}
	}
	return reply.MakeIntReply(result)
}

// Handle the FLUSHDB command.
// It clears all keys from the database
func execFlushDB(db *DB, args [][]byte) resp.Reply {
	db.Flush()
	db.addAof(utils.ToCmdLineWithName("FLUSHDB", args...))
	return reply.MakeOKReply()
}

// Handle the TYPE command.
// It returns the type of the specified key
func execType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	if entity, ok := db.GetEntity(key); ok {
		switch entity.Data.(type) {
		// If the entity is []byte, return the type as "string"
		case []byte:
			return reply.MakeBulkReply([]byte("string"))
		}
		// TODO: Add more types as needed
	} else {
		return reply.MakeStatusReply("none")
	}
	return reply.MakeUnknownReply()
}

// Handle the RENAME command.
// It renames a key in the database.
// RENAME key newkey
func execRename(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dst := string(args[1])
	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeStandardErrorReply("ERR no such key")
	}
	db.PutEntity(dst, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLineWithName("RENAME", args...))
	return reply.MakeOKReply()
}

// Handle the RENAMENX command.
// It renames a key in the database only if the new key does not exist.
// RENAME key newkey
func execRenameNX(db *DB, args [][]byte) resp.Reply {
	src := string(args[0])
	dst := string(args[1])
	entity, ok := db.GetEntity(src)
	if !ok {
		return reply.MakeStandardErrorReply("ERR no such key")
	}
	if _, ok := db.GetEntity(dst); ok {
		return reply.MakeIntReply(0)
	}
	db.PutEntity(dst, entity)
	db.Remove(src)
	db.addAof(utils.ToCmdLineWithName("RENAMENX", args...))
	return reply.MakeIntReply(1)
}

// Handle the KEYS command.
// It returns all keys in the database that match the specified pattern.
func execKeys(db *DB, args [][]byte) resp.Reply {
	pattern := wildcard.CompilePattern(string(args[0]))
	result := make([][]byte, 0) // Store all matching keys
	db.data.ForEach(func(key string, val interface{}) bool {
		if pattern.IsMatch(key) {
			result = append(result, []byte(key))
		}
		return true
	})
	return reply.MakeMultiBulkReply(result)
}

func init() {
	RegisterCommand("DEL", execDel, -2)
	RegisterCommand("EXISTS", execExists, -2)
	RegisterCommand("FLUSHDB", execFlushDB, -1)
	RegisterCommand("TYPE", execType, 2)
	RegisterCommand("RENAME", execRename, 3)
	RegisterCommand("RENAMENX", execRenameNX, 3)
	RegisterCommand("KEYS", execKeys, 2)
}
