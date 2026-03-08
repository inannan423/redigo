package database

import (
	"redigo/interface/database"
	"redigo/interface/resp"
	"redigo/lib/utils"
	"redigo/resp/reply"
	"strconv"
	"strings"
	"time"
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

	var expireAt int64
	hasExpire := false
	if len(args) > 2 {
		if (len(args)-2)%2 != 0 {
			return reply.MakeStandardErrorReply("ERR syntax error")
		}
		for i := 2; i < len(args); i += 2 {
			option := strings.ToLower(string(args[i]))
			if option != "ex" && option != "px" {
				return reply.MakeStandardErrorReply("ERR syntax error")
			}
			if hasExpire {
				return reply.MakeStandardErrorReply("ERR syntax error")
			}
			ttlValue, err := strconv.ParseInt(string(args[i+1]), 10, 64)
			if err != nil {
				return reply.MakeStandardErrorReply("ERR value is not an integer or out of range")
			}
			if ttlValue <= 0 {
				return reply.MakeStandardErrorReply("ERR invalid expire time in set")
			}
			if option == "ex" {
				expireAt = time.Now().UnixMilli() + ttlValue*1000
			} else {
				expireAt = time.Now().UnixMilli() + ttlValue
			}
			hasExpire = true
		}
	}

	entity := &database.DataEntity{
		Data: value,
	}
	db.PutEntity(key, entity)
	db.addAof(utils.ToCmdLineWithName("SET", []byte(key), value))
	if hasExpire {
		db.SetExpire(key, expireAt)
		db.addAof(utils.ToCmdLineWithName("PEXPIREAT", []byte(key), []byte(strconv.FormatInt(expireAt, 10))))
	}
	return reply.MakeOKReply()
}

// execSetEX stores the specified key-value pair with expiration time in seconds.
func execSetEX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	seconds, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeStandardErrorReply("ERR value is not an integer or out of range")
	}
	if seconds <= 0 {
		return reply.MakeStandardErrorReply("ERR invalid expire time in setex")
	}
	value := args[2]

	entity := &database.DataEntity{
		Data: value,
	}
	db.PutEntity(key, entity)
	expireAt := time.Now().UnixMilli() + seconds*1000
	db.SetExpire(key, expireAt)

	db.addAof(utils.ToCmdLineWithName("SET", []byte(key), value))
	db.addAof(utils.ToCmdLineWithName("PEXPIREAT", []byte(key), []byte(strconv.FormatInt(expireAt, 10))))
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
	db.addAof(utils.ToCmdLineWithName("SETNX", args...))
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
	db.addAof(utils.ToCmdLineWithName("GETSET", args...))
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
	RegisterCommand("SET", execSet, -3)
	RegisterCommand("SETNX", execSetNX, 3)
	RegisterCommand("GETSET", execGetSet, 3)
	RegisterCommand("SETEX", execSetEX, 4)
	RegisterCommand("STRLEN", execStrLen, 2)
}
