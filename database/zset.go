package database

import (
	"redigo/interface/database"
	"redigo/interface/resp"
	"redigo/lib/utils"
	"redigo/resp/reply"
	"strconv"
)

// parseFloat parses a string to float64, handling errors
func parseFloat(val string) (float64, resp.Reply) {
	score, err := strconv.ParseFloat(val, 64)
	if err != nil {
		return 0, reply.MakeStandardErrorReply("value is not a valid float")
	}
	return score, nil
}

// execZAdd implements the ZADD command
// ZADD key [NX|XX] [CH] [INCR] score member [score member ...]
func execZAdd(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 || len(args)%2 == 0 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zadd' command")
	}

	key := string(args[0])

	var result resp.Reply

	// Use key-level locking to prevent concurrent modification of the same zset
	db.WithKeyLock(key, func() {
		// Get or create ZSet
		zsetObj, exists := getAsZSet(db, key)
		if exists && zsetObj == nil {
			result = reply.MakeWrongTypeErrReply()
			return
		}

		added := 0
		for i := 1; i < len(args); i += 2 {
			scoreStr := string(args[i])
			member := string(args[i+1])

			// Parse score
			score, err := parseFloat(scoreStr)
			if err != nil {
				result = err
				return
			}

			// Add member to ZSet
			if zsetObj.Add(member, score) {
				added++
			}
		}

		// Store ZSet in database
		db.PutEntity(key, &database.DataEntity{Data: zsetObj})

		// Add AOF record
		db.addAof(utils.ToCmdLineWithName("ZADD", args...))

		result = reply.MakeIntReply(int64(added))
	})

	return result
}

// execZScore implements the ZSCORE command
// ZSCORE key member
func execZScore(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zscore' command")
	}

	key := string(args[0])
	member := string(args[1])

	zsetObj, exists := getAsZSet(db, key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	if zsetObj == nil {
		return reply.MakeWrongTypeErrReply()
	}

	// Get score
	score, exists := zsetObj.Score(member)
	if !exists {
		return reply.MakeNullBulkReply()
	}

	return reply.MakeBulkReply([]byte(strconv.FormatFloat(score, 'f', -1, 64)))
}

// execZCard implements the ZCARD command
// ZCARD key
func execZCard(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zcard' command")
	}

	key := string(args[0])

	zsetObj, exists := getAsZSet(db, key)
	if !exists {
		return reply.MakeIntReply(0)
	}
	if zsetObj == nil {
		return reply.MakeWrongTypeErrReply()
	}

	return reply.MakeIntReply(int64(zsetObj.Len()))
}

// execZRange implements the ZRANGE command
// ZRANGE key start stop [WITHSCORES]
func execZRange(db *DB, args [][]byte) resp.Reply {
	if len(args) < 3 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zrange' command")
	}

	withScores := false
	if len(args) > 3 && string(args[3]) == "WITHSCORES" {
		withScores = true
	}

	key := string(args[0])

	// Parse start and stop indices
	start, err := strconv.Atoi(string(args[1]))
	if err != nil {
		return reply.MakeStandardErrorReply("value is not an integer or out of range")
	}

	stop, err := strconv.Atoi(string(args[2]))
	if err != nil {
		return reply.MakeStandardErrorReply("value is not an integer or out of range")
	}

	var result resp.Reply

	// Use read lock to allow concurrent reads while preventing concurrent writes
	db.WithKeyRLock(key, func() {
		// Get ZSet
		zsetObj, exists := getAsZSet(db, key)
		if !exists {
			result = reply.MakeEmptyMultiBulkReply()
			return
		}
		if zsetObj == nil {
			result = reply.MakeWrongTypeErrReply()
			return
		}

		// Get range
		members := zsetObj.RangeByRank(start, stop)

		// Prepare result
		if !withScores {
			resultBytes := make([][]byte, len(members))
			for i, member := range members {
				resultBytes[i] = []byte(member)
			}
			result = reply.MakeMultiBulkReply(resultBytes)
		} else {
			resultBytes := make([][]byte, len(members)*2)
			for i, member := range members {
				resultBytes[i*2] = []byte(member)
				score, _ := zsetObj.Score(member)
				resultBytes[i*2+1] = []byte(strconv.FormatFloat(score, 'f', -1, 64))
			}
			result = reply.MakeMultiBulkReply(resultBytes)
		}
	})

	return result
}

// execZRem implements the ZREM command
// ZREM key member [member ...]
func execZRem(db *DB, args [][]byte) resp.Reply {
	if len(args) < 2 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zrem' command")
	}

	key := string(args[0])

	var result resp.Reply

	// Use key-level locking to prevent concurrent modification of the same zset
	db.WithKeyLock(key, func() {
		// Get ZSet
		zsetObj, exists := getAsZSet(db, key)
		if !exists {
			result = reply.MakeIntReply(0)
			return
		}
		if zsetObj == nil {
			result = reply.MakeWrongTypeErrReply()
			return
		}

		// Remove members
		removed := 0
		for i := 1; i < len(args); i++ {
			member := string(args[i])
			if zsetObj.Remove(member) {
				removed++
			}
		}

		// Update database if we removed anything
		if removed > 0 {
			db.PutEntity(key, &database.DataEntity{Data: zsetObj})

			// Add AOF record
			db.addAof(utils.ToCmdLineWithName("ZREM", args...))
		}

		result = reply.MakeIntReply(int64(removed))
	})

	return result
}

// execZCount implements the ZCOUNT command
// ZCOUNT key min max
func execZCount(db *DB, args [][]byte) resp.Reply {
	if len(args) != 3 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zcount' command")
	}

	key := string(args[0])

	// Parse min and max scores
	min, err := parseFloat(string(args[1]))
	if err != nil {
		return err
	}

	max, err := parseFloat(string(args[2]))
	if err != nil {
		return err
	}

	var result resp.Reply

	// Use read lock to allow concurrent reads while preventing concurrent writes
	db.WithKeyRLock(key, func() {
		// Get ZSet
		zsetObj, exists := getAsZSet(db, key)
		if !exists {
			result = reply.MakeIntReply(0)
			return
		}
		if zsetObj == nil {
			result = reply.MakeWrongTypeErrReply()
			return
		}

		// Count elements in range
		count := zsetObj.Count(min, max)

		result = reply.MakeIntReply(int64(count))
	})

	return result
}

// execZRank implements the ZRANK command
// ZRANK key member
func execZRank(db *DB, args [][]byte) resp.Reply {
	if len(args) != 2 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'zrank' command")
	}

	key := string(args[0])
	member := string(args[1])

	var result resp.Reply

	// Use read lock to allow concurrent reads while preventing concurrent writes
	db.WithKeyRLock(key, func() {
		// Get ZSet
		zsetObj, exists := getAsZSet(db, key)
		if !exists {
			result = reply.MakeNullBulkReply()
			return
		}
		if zsetObj == nil {
			result = reply.MakeWrongTypeErrReply()
			return
		}

		score, exists := zsetObj.Score(member)
		if !exists {
			result = reply.MakeNullBulkReply()
			return
		}

		// Calculate rank
		rank := -1
		if zsetObj.Encoding() == 1 {
			skiplist := zsetObj.GetSkiplist()
			rank = skiplist.GetRank(member, score)
		} else {
			members := zsetObj.RangeByRank(0, -1)
			for i, m := range members {
				if m == member {
					rank = i
					break
				}
			}
		}

		if rank == -1 {
			result = reply.MakeNullBulkReply()
			return
		}

		result = reply.MakeIntReply(int64(rank))
	})

	return result
}

// execZTYPE implements the ZTYPE command
// ZTYPE key returns the type of the key, 0 for listpack, 1 for skiplist
func execZType(db *DB, args [][]byte) resp.Reply {
	if len(args) != 1 {
		return reply.MakeStandardErrorReply("wrong number of arguments for 'ztype' command")
	}

	key := string(args[0])

	// Get ZSet
	zsetObj, exists := getAsZSet(db, key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	if zsetObj == nil {
		return reply.MakeWrongTypeErrReply()
	}

	return reply.MakeIntReply(int64(zsetObj.Encoding()))
}

// Register ZSET commands
func init() {
	RegisterCommand("ZADD", execZAdd, -4)     // key score member [score member ...]
	RegisterCommand("ZSCORE", execZScore, 3)  // key member
	RegisterCommand("ZCARD", execZCard, 2)    // key
	RegisterCommand("ZRANGE", execZRange, -4) // key start stop [WITHSCORES]
	RegisterCommand("ZREM", execZRem, -3)     // key member [member ...]
	RegisterCommand("ZCOUNT", execZCount, 4)  // key min max
	RegisterCommand("ZRANK", execZRank, 3)    // key member
	RegisterCommand("ZTYPE", execZType, 2)    // key
}
