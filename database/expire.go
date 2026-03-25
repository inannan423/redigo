package database

import (
	"redigo/interface/resp"
	"redigo/lib/utils"
	"redigo/resp/reply"
	"strconv"
	"time"
)

func parseExpireArg(arg []byte) (int64, resp.Reply) {
	value, err := strconv.ParseInt(string(arg), 10, 64)
	if err != nil {
		return 0, reply.MakeStandardErrorReply("ERR value is not an integer or out of range")
	}
	return value, nil
}

func setExpireAt(db *DB, key string, expireAt int64) resp.Reply {
	if _, ok := db.GetEntity(key); !ok {
		return reply.MakeIntReply(0)
	}
	if expireAt <= time.Now().UnixMilli() {
		db.Remove(key)
	} else {
		db.SetExpire(key, expireAt)
	}
	db.addAof(utils.ToCmdLineWithName("PEXPIREAT", []byte(key), []byte(strconv.FormatInt(expireAt, 10))))
	return reply.MakeIntReply(1)
}

// execExpire handles EXPIRE key seconds
func execExpire(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	seconds, errReply := parseExpireArg(args[1])
	if errReply != nil {
		return errReply
	}
	expireAt := time.Now().UnixMilli() + seconds*1000
	return setExpireAt(db, key, expireAt)
}

// execPExpire handles PEXPIRE key milliseconds
func execPExpire(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	milliseconds, errReply := parseExpireArg(args[1])
	if errReply != nil {
		return errReply
	}
	expireAt := time.Now().UnixMilli() + milliseconds
	return setExpireAt(db, key, expireAt)
}

// execExpireAt handles EXPIREAT key timestamp (seconds)
func execExpireAt(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	seconds, errReply := parseExpireArg(args[1])
	if errReply != nil {
		return errReply
	}
	expireAt := seconds * 1000
	return setExpireAt(db, key, expireAt)
}

// execPExpireAt handles PEXPIREAT key timestamp (milliseconds)
func execPExpireAt(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	milliseconds, errReply := parseExpireArg(args[1])
	if errReply != nil {
		return errReply
	}
	return setExpireAt(db, key, milliseconds)
}

func ttlWithUnit(db *DB, key string, unit int64) resp.Reply {
	if _, ok := db.GetEntity(key); !ok {
		return reply.MakeIntReply(-2)
	}
	expireAt, ok := db.GetExpire(key)
	if !ok {
		return reply.MakeIntReply(-1)
	}
	remaining := expireAt - time.Now().UnixMilli()
	if remaining <= 0 {
		db.Remove(key)
		return reply.MakeIntReply(-2)
	}
	return reply.MakeIntReply(remaining / unit)
}

// execTTL handles TTL key
func execTTL(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	return ttlWithUnit(db, key, 1000)
}

// execPTTL handles PTTL key
func execPTTL(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	return ttlWithUnit(db, key, 1)
}

// execPersist handles PERSIST key
func execPersist(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	if _, ok := db.GetEntity(key); !ok {
		return reply.MakeIntReply(0)
	}
	if _, ok := db.GetExpire(key); !ok {
		return reply.MakeIntReply(0)
	}
	db.ClearExpire(key)
	db.addAof(utils.ToCmdLineWithName("PERSIST", args...))
	return reply.MakeIntReply(1)
}

func init() {
	RegisterCommand("EXPIRE", execExpire, 3)
	RegisterCommand("PEXPIRE", execPExpire, 3)
	RegisterCommand("EXPIREAT", execExpireAt, 3)
	RegisterCommand("PEXPIREAT", execPExpireAt, 3)
	RegisterCommand("TTL", execTTL, 2)
	RegisterCommand("PTTL", execPTTL, 2)
	RegisterCommand("PERSIST", execPersist, 2)
}
