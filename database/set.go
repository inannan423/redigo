package database

import (
	"redigo/datastruct/set"
	"redigo/interface/database"
	"redigo/interface/resp"
	"redigo/lib/utils"
	"redigo/resp/reply"
	"strconv"
)

// strToInt converts string to int
func strToInt(str string) (int, error) {
	value, err := strconv.Atoi(str)
	if err != nil {
		return 0, err
	}
	return value, nil
}

// intToStr converts int to string
func intToStr(n int) string {
	return strconv.Itoa(n)
}

// execSAdd implements SADD key member [member...]
// Add one or more members to a set
func execSAdd(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	members := args[1:]

	var result resp.Reply

	// Use key-level locking to prevent concurrent modification of the same set
	db.WithKeyLock(key, func() {
		// Get or create set
		setObj, isNew, errReply := getOrInitSet(db, key)
		if errReply != nil {
			result = errReply
			return
		}

		// Add all members
		count := 0
		for _, member := range members {
			count += setObj.Add(string(member))
		}

		// Store back to database if it's a new set or any members were added
		if isNew || count > 0 {
			db.PutEntity(key, &database.DataEntity{
				Data: setObj,
			})

			// Add to AOF
			db.addAof(utils.ToCmdLineWithName("SADD", args...))
		}

		result = reply.MakeIntReply(int64(count))
	})

	return result
}

// execSCard implements SCARD key
// Get the number of members in a set
func execSCard(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// Get set
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil {
		return reply.MakeIntReply(0)
	}

	return reply.MakeIntReply(int64(setObj.Len()))
}

// execSIsMember implements SISMEMBER key member
// Determine if a given value is a member of a set
func execSIsMember(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	member := string(args[1])

	// Get set
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil {
		return reply.MakeIntReply(0)
	}

	if setObj.Contains(member) {
		return reply.MakeIntReply(1)
	}
	return reply.MakeIntReply(0)
}

// execSMembers implements SMEMBERS key
// Get all the members in a set
func execSMembers(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// Get set
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil {
		return reply.MakeMultiBulkReply([][]byte{})
	}

	// Convert members to [][]byte
	members := setObj.Members()
	result := make([][]byte, len(members))
	for i, member := range members {
		result[i] = []byte(member)
	}

	return reply.MakeMultiBulkReply(result)
}

// execSRem implements SREM key member [member...]
// Remove one or more members from a set
func execSRem(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	members := args[1:]

	var result resp.Reply

	// Use key-level locking to prevent concurrent modification of the same set
	db.WithKeyLock(key, func() {
		// Get set
		setObj, errReply := getAsSet(db, key)
		if errReply != nil {
			result = errReply
			return
		}
		if setObj == nil {
			result = reply.MakeIntReply(0)
			return
		}

		// Remove all members
		count := 0
		for _, member := range members {
			count += setObj.Remove(string(member))
		}

		// If any members were removed
		if count > 0 {
			// Check if set is now empty
			if setObj.Len() == 0 {
				db.Remove(key)
			} else {
				// Store updated set
				db.PutEntity(key, &database.DataEntity{
					Data: setObj,
				})
			}

			// Add to AOF
			db.addAof(utils.ToCmdLineWithName("SREM", args...))
		}

		result = reply.MakeIntReply(int64(count))
	})

	return result
}

// execSPop implements SPOP key [count]
// Remove and return one or multiple random members from a set
func execSPop(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// Determine count
	count := 1
	if len(args) >= 2 {
		var err error
		count, err = strToInt(string(args[1]))
		if err != nil || count < 0 {
			return reply.MakeStandardErrorReply("ERR value is out of range, must be positive")
		}
	}

	// Get set
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil || setObj.Len() == 0 {
		return reply.MakeNullBulkReply()
	}

	// If count is 0, return empty array
	if count == 0 {
		return reply.MakeMultiBulkReply([][]byte{})
	}

	// Cap count to set size
	if count > setObj.Len() {
		count = setObj.Len()
	}

	// Get random members
	members := setObj.RandomDistinctMembers(count)

	// Remove members
	for _, member := range members {
		setObj.Remove(member)
	}

	// Store updated set or remove if empty
	if setObj.Len() == 0 {
		db.Remove(key)
	} else {
		db.PutEntity(key, &database.DataEntity{
			Data: setObj,
		})
	}

	// Add to AOF
	cmdArgs := make([][]byte, 2)
	cmdArgs[0] = []byte(key)
	cmdArgs[1] = []byte(intToStr(count))
	db.addAof(utils.ToCmdLineWithName("SPOP", cmdArgs...))

	// If only popping one member, return it as a bulk string
	if count == 1 {
		return reply.MakeBulkReply([]byte(members[0]))
	}

	// Otherwise return array of members
	result := make([][]byte, len(members))
	for i, member := range members {
		result[i] = []byte(member)
	}
	return reply.MakeMultiBulkReply(result)
}

// execSRandMember implements SRANDMEMBER key [count]
// Get one or multiple random members from a set
func execSRandMember(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// Get set
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil || setObj.Len() == 0 {
		return reply.MakeNullBulkReply()
	}

	// Determine count
	count := 1
	withReplacement := false
	if len(args) >= 2 {
		var err error
		count, err = strToInt(string(args[1]))
		if err != nil {
			return reply.MakeStandardErrorReply("ERR value is not an integer")
		}

		// Negative count means return with replacement (can have duplicates)
		if count < 0 {
			withReplacement = true
			count = -count
		}
	}

	// Get random members
	var members []string
	if withReplacement {
		members = setObj.RandomMembers(count)
	} else {
		members = setObj.RandomDistinctMembers(count)
	}

	// If only returning one member, return it as a bulk string
	if len(args) == 1 || (count == 1 && len(members) > 0) {
		return reply.MakeBulkReply([]byte(members[0]))
	}

	// Otherwise return array of members
	result := make([][]byte, len(members))
	for i, member := range members {
		result[i] = []byte(member)
	}
	return reply.MakeMultiBulkReply(result)
}

// execSUnion implements SUNION key [key...]
// Return the union of multiple sets
func execSUnion(db *DB, args [][]byte) resp.Reply {
	// Create empty result set
	result := set.NewHashSet()

	// Process each set
	for _, arg := range args {
		key := string(arg)
		setObj, errReply := getAsSet(db, key)
		if errReply != nil {
			return errReply
		}
		if setObj == nil {
			continue
		}

		// Add all members to result
		setObj.ForEach(func(member string) bool {
			result.Add(member)
			return true
		})
	}

	// Convert set to reply
	members := result.Members()
	resultBytes := make([][]byte, len(members))
	for i, member := range members {
		resultBytes[i] = []byte(member)
	}

	return reply.MakeMultiBulkReply(resultBytes)
}

// execSUnionStore implements SUNIONSTORE destination key [key...]
// Store the union of multiple sets in a new set
func execSUnionStore(db *DB, args [][]byte) resp.Reply {
	destKey := string(args[0])
	keys := args[1:]

	// Execute union
	unionReply := execSUnion(db, keys)
	if _, ok := unionReply.(reply.ErrorReply); ok {
		return unionReply
	}

	// Create new set with union result
	unionResult := unionReply.(*reply.MultiBulkReply)
	newSet := set.NewHashSet()
	for _, member := range unionResult.Args {
		newSet.Add(string(member))
	}

	// Store set in database
	db.PutEntity(destKey, &database.DataEntity{
		Data: newSet,
	})

	// Add to AOF
	db.addAof(utils.ToCmdLineWithName("SUNIONSTORE", args...))

	return reply.MakeIntReply(int64(newSet.Len()))
}

// execSInter implements SINTER key [key...]
// Return the intersection of multiple sets
func execSInter(db *DB, args [][]byte) resp.Reply {
	if len(args) == 0 {
		return reply.MakeEmptyMultiBulkReply()
	}

	// Get first set as base
	key := string(args[0])
	firstSet, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if firstSet == nil {
		return reply.MakeEmptyMultiBulkReply()
	}

	// Create result set with members of first set
	result := set.NewHashSet()
	firstSet.ForEach(func(member string) bool {
		result.Add(member)
		return true
	})

	// Intersect with each other set
	for i := 1; i < len(args); i++ {
		key := string(args[i])
		currentSet, errReply := getAsSet(db, key)
		if errReply != nil {
			return errReply
		}

		// Empty set or key doesn't exist means empty intersection
		if currentSet == nil {
			return reply.MakeEmptyMultiBulkReply()
		}

		// Keep only members that exist in current set
		toRemove := make([]string, 0)
		result.ForEach(func(member string) bool {
			if !currentSet.Contains(member) {
				toRemove = append(toRemove, member)
			}
			return true
		})

		// Remove non-intersecting members
		for _, member := range toRemove {
			result.Remove(member)
		}

		// Early termination if result is already empty
		if result.Len() == 0 {
			return reply.MakeEmptyMultiBulkReply()
		}
	}

	// Convert result to reply
	members := result.Members()
	resultBytes := make([][]byte, len(members))
	for i, member := range members {
		resultBytes[i] = []byte(member)
	}

	return reply.MakeMultiBulkReply(resultBytes)
}

// execSInterStore implements SINTERSTORE destination key [key...]
// Store the intersection of multiple sets in a new set
func execSInterStore(db *DB, args [][]byte) resp.Reply {
	destKey := string(args[0])
	keys := args[1:]

	// Execute intersection
	interReply := execSInter(db, keys)
	if _, ok := interReply.(reply.ErrorReply); ok {
		return interReply
	}

	// Create new set with intersection result
	interResult, ok := interReply.(*reply.MultiBulkReply)
	if !ok {
		return reply.MakeEmptyMultiBulkReply()
	}

	newSet := set.NewHashSet()
	for _, member := range interResult.Args {
		newSet.Add(string(member))
	}

	// Store set in database
	db.PutEntity(destKey, &database.DataEntity{
		Data: newSet,
	})

	// Add to AOF
	db.addAof(utils.ToCmdLineWithName("SINTERSTORE", args...))

	return reply.MakeIntReply(int64(newSet.Len()))
}

// execSDiff implements SDIFF key [key...]
// Return the difference between sets
func execSDiff(db *DB, args [][]byte) resp.Reply {
	// Get first set as base
	key := string(args[0])
	firstSet, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if firstSet == nil {
		return reply.MakeEmptyMultiBulkReply()
	}

	// Create result set with members of first set
	result := set.NewHashSet()
	firstSet.ForEach(func(member string) bool {
		result.Add(member)
		return true
	})

	// Remove members that appear in subsequent sets
	for i := 1; i < len(args); i++ {
		key := string(args[i])
		currentSet, errReply := getAsSet(db, key)
		if errReply != nil {
			return errReply
		}
		if currentSet == nil {
			continue
		}

		// Remove members that exist in current set
		currentSet.ForEach(func(member string) bool {
			result.Remove(member)
			return true
		})

		// Early termination if result is already empty
		if result.Len() == 0 {
			return reply.MakeEmptyMultiBulkReply()
		}
	}

	// Convert result to reply
	members := result.Members()
	resultBytes := make([][]byte, len(members))
	for i, member := range members {
		resultBytes[i] = []byte(member)
	}

	return reply.MakeMultiBulkReply(resultBytes)
}

// execSDiffStore implements SDIFFSTORE destination key [key...]
// Store the difference between sets in a new set
func execSDiffStore(db *DB, args [][]byte) resp.Reply {
	destKey := string(args[0])
	keys := args[1:]

	// Execute difference
	diffReply := execSDiff(db, keys)
	if _, ok := diffReply.(reply.ErrorReply); ok {
		return diffReply
	}

	// Create new set with difference result
	diffResult, ok := diffReply.(*reply.MultiBulkReply)
	if !ok {
		return reply.MakeIntReply(0)
	}

	newSet := set.NewHashSet()
	for _, member := range diffResult.Args {
		newSet.Add(string(member))
	}

	// Store set in database
	db.PutEntity(destKey, &database.DataEntity{
		Data: newSet,
	})

	// Add to AOF
	db.addAof(utils.ToCmdLineWithName("SDIFFSTORE", args...))

	return reply.MakeIntReply(int64(newSet.Len()))
}

// SetType represents the type of the set (intset or hashset)
func execSetType(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// Get set
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return errReply
	}
	if setObj == nil {
		return reply.MakeNullBulkReply()
	}

	// Determine set type
	if setObj.IsIntSet() {
		return reply.MakeStatusReply("intset")
	}
	return reply.MakeStatusReply("hashset")
}

func init() {
	RegisterCommand("SADD", execSAdd, -3)
	RegisterCommand("SCARD", execSCard, 2)
	RegisterCommand("SISMEMBER", execSIsMember, 3)
	RegisterCommand("SMEMBERS", execSMembers, 2)
	RegisterCommand("SREM", execSRem, -3)
	RegisterCommand("SPOP", execSPop, -2)
	RegisterCommand("SRANDMEMBER", execSRandMember, -2)
	RegisterCommand("SUNION", execSUnion, -2)
	RegisterCommand("SUNIONSTORE", execSUnionStore, -3)
	RegisterCommand("SINTER", execSInter, -2)
	RegisterCommand("SINTERSTORE", execSInterStore, -3)
	RegisterCommand("SDIFF", execSDiff, -2)
	RegisterCommand("SDIFFSTORE", execSDiffStore, -3)
	RegisterCommand("SETTYPE", execSetType, 2)
}
