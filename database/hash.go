package database

import (
	"redigo/interface/resp"
	"redigo/lib/utils"
	"redigo/resp/reply"
)

// HSet sets field in the hash stored at key to value
func execHSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	field := string(args[1])
	value := string(args[2])

	var result resp.Reply

	// Use key-level locking to prevent concurrent modification of the same hash
	db.WithKeyLock(key, func() {
		hashObj, _ := db.getOrCreateHash(key)
		res := hashObj.Set(field, value)

		db.addAof(utils.ToCmdLineWithName("HSET", args...))

		result = reply.MakeIntReply(int64(res))
	})

	return result
}

// HGet returns field value in hash
func execHGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	field := string(args[1])

	hash, exists := db.getAsHash(key)
	if !exists {
		return reply.MakeNullBulkReply()
	}

	value, exists := hash.Get(field)
	if !exists {
		return reply.MakeNullBulkReply()
	}

	return reply.MakeBulkReply([]byte(value))
}

// HExists checks if field exists in hash
func execHExists(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	field := string(args[1])

	hash, exists := db.getAsHash(key)
	if !exists {
		return reply.MakeIntReply(0)
	}

	exists = hash.Exists(field)
	if exists {
		return reply.MakeIntReply(1)
	}
	return reply.MakeIntReply(0)
}

// HDel deletes fields from hash
func execHDel(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	var result resp.Reply

	// Use key-level locking to prevent concurrent modification of the same hash
	db.WithKeyLock(key, func() {
		hash, exists := db.getAsHash(key)
		if !exists {
			result = reply.MakeIntReply(0)
			return
		}

		deleted := 0
		for _, field := range args[1:] {
			deleted += hash.Delete(string(field))
		}

		if hash.Len() == 0 {
			db.Remove(key)
		}

		if deleted > 0 {
			db.addAof(utils.ToCmdLineWithName("hdel", args...))
		}

		result = reply.MakeIntReply(int64(deleted))
	})

	return result
}

// HLen returns number of fields in hash
func execHLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	hash, exists := db.getAsHash(key)
	if !exists {
		return reply.MakeIntReply(0)
	}

	return reply.MakeIntReply(int64(hash.Len()))
}

// HGetAll returns all fields and values in hash
func execHGetAll(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	var result resp.Reply

	// Use read lock to allow concurrent reads while preventing concurrent writes
	db.WithKeyRLock(key, func() {
		hash, exists := db.getAsHash(key)
		if !exists {
			result = reply.MakeEmptyMultiBulkReply()
			return
		}

		allMap := hash.GetAll()
		resultBytes := make([][]byte, 0, len(allMap)*2)
		for field, value := range allMap {
			resultBytes = append(resultBytes, []byte(field))
			resultBytes = append(resultBytes, []byte(value))
		}

		result = reply.MakeMultiBulkReply(resultBytes)
	})

	return result
}

// HKeys returns all fields in hash
func execHKeys(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	var result resp.Reply

	// Use read lock to allow concurrent reads while preventing concurrent writes
	db.WithKeyRLock(key, func() {
		hash, exists := db.getAsHash(key)
		if !exists {
			result = reply.MakeEmptyMultiBulkReply()
			return
		}

		fields := hash.Fields()
		resultBytes := make([][]byte, len(fields))
		for i, field := range fields {
			resultBytes[i] = []byte(field)
		}

		result = reply.MakeMultiBulkReply(resultBytes)
	})

	return result
}

// HVals returns all values in hash
func execHVals(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	var result resp.Reply

	// Use read lock to allow concurrent reads while preventing concurrent writes
	db.WithKeyRLock(key, func() {
		hash, exists := db.getAsHash(key)
		if !exists {
			result = reply.MakeEmptyMultiBulkReply()
			return
		}

		values := hash.Values()
		resultBytes := make([][]byte, len(values))
		for i, value := range values {
			resultBytes[i] = []byte(value)
		}

		result = reply.MakeMultiBulkReply(resultBytes)
	})

	return result
}

// HMGet returns values for multiple fields in hash
func execHMGet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	var result resp.Reply

	// Use read lock to allow concurrent reads while preventing concurrent writes
	db.WithKeyRLock(key, func() {
		hash, exists := db.getAsHash(key)
		if !exists {
			results := make([][]byte, len(args)-1)
			for i := range results {
				results[i] = nil
			}
			result = reply.MakeMultiBulkReply(results)
			return
		}

		results := make([][]byte, len(args)-1)
		for i, field := range args[1:] {
			value, exists := hash.Get(string(field))
			if exists {
				results[i] = []byte(value)
			} else {
				results[i] = nil
			}
		}

		result = reply.MakeMultiBulkReply(results)
	})

	return result
}

// HMSet sets multiple fields in hash
// HMSET key field value [field value ...]
func execHMSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	if len(args)%2 == 0 {
		return reply.MakeStandardErrorReply("ERR wrong number of arguments for 'hmset' command")
	}

	var result resp.Reply

	// Use key-level locking to prevent concurrent modification of the same hash
	db.WithKeyLock(key, func() {
		hash, _ := db.getOrCreateHash(key)

		for i := 1; i < len(args); i += 2 {
			field := string(args[i])
			value := string(args[i+1])
			hash.Set(field, value)
		}

		db.addAof(utils.ToCmdLineWithName("hmset", args...))

		result = reply.MakeOKReply()
	})

	return result
}

// HEncoding returns the encoding of the hash.
// 0 for listpack, 1 for dict.
// This is a diy function to check the encoding of the hash.
func execHEncoding(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	hash, exists := db.getAsHash(key)
	if !exists {
		return reply.MakeNullBulkReply()
	}

	return reply.MakeIntReply(int64(hash.Encoding()))
}

// execHSetNX sets field in the hash stored at key to value, only if field does not exist
// HSETNX key field value
func execHSetNX(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	field := string(args[1])
	value := string(args[2])

	var result resp.Reply

	// Use key-level locking to prevent concurrent modification of the same hash
	db.WithKeyLock(key, func() {
		hash, _ := db.getOrCreateHash(key)

		_, exists := hash.Get(field)
		if exists {
			result = reply.MakeIntReply(0)
			return
		}

		hash.Set(field, value)

		db.addAof(utils.ToCmdLineWithName("HSETNX", args...))

		result = reply.MakeIntReply(1)
	})

	return result
}

func init() {
	// Register hash commands
	RegisterCommand("HSET", execHSet, 4)           // HSET key field value
	RegisterCommand("HGET", execHGet, 3)           // HGET key field
	RegisterCommand("HEXISTS", execHExists, 3)     // HEXISTS key field
	RegisterCommand("HDEL", execHDel, -3)          // HDEL key field [field ...] (at least 2 args plus command name)
	RegisterCommand("HLEN", execHLen, 2)           // HLEN key
	RegisterCommand("HGETALL", execHGetAll, 2)     // HGETALL key
	RegisterCommand("HKEYS", execHKeys, 2)         // HKEYS key
	RegisterCommand("HVALS", execHVals, 2)         // HVALS key
	RegisterCommand("HMGET", execHMGet, -3)        // HMGET key field [field ...] (at least 2 args plus command name)
	RegisterCommand("HMSET", execHMSet, -4)        // HMSET key field value [field value ...] (at least 3 args plus command name)
	RegisterCommand("HENCODING", execHEncoding, 2) // HENCODING key
	RegisterCommand("HSETNX", execHSetNX, 4)       // HSETNX key field value
}
