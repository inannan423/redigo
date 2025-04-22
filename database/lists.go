package database

import (
	// Use the go standard library's list package
	"container/list"
	"redigo/interface/database"
	"redigo/interface/resp"
	"redigo/lib/utils"
	"redigo/resp/reply"
	"strconv"
)

// getAsList retrieves the list stored at the given key, or creates a new one if it doesn't exist.
// It returns the list and a boolean indicating if the key existed.
func getAsList(db *DB, key string) (*list.List, bool) {
	entity, ok := db.GetEntity(key)
	if !ok {
		// Key doesn't exist, create a new list
		return list.New(), false
	}
	// Key exists, check if it's a list
	lst, ok := entity.Data.(*list.List)
	if !ok {
		// Key exists but is not a list type
		return nil, true // Indicate key exists but is wrong type
	}
	return lst, true
}

// execLPush implements the LPUSH command: Prepends one or multiple values to a list
// LPUSH key value [value ...]
func execLPush(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	values := args[1:]

	// Get or create list
	lst, exists := getAsList(db, key)
	if lst == nil && exists { // Key exists but is not a list
		return reply.MakeWrongTypeErrReply()
	}

	// Prepend values
	for _, value := range values {
		lst.PushFront(value) // Add to the front (left)
	}

	// Store the updated list
	db.PutEntity(key, &database.DataEntity{Data: lst})
	db.addAof(utils.ToCmdLineWithName("LPUSH", args...))

	// Return the new length of the list
	return reply.MakeIntReply(int64(lst.Len()))
}

// execRPush implements the RPUSH command: Appends one or multiple values to a list
// RPUSH key value [value ...]
func execRPush(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	values := args[1:]

	// Get or create list
	lst, exists := getAsList(db, key)
	if lst == nil && exists { // Key exists but is not a list
		return reply.MakeWrongTypeErrReply()
	}

	// Append values
	for _, value := range values {
		lst.PushBack(value) // Add to the back (right)
	}

	// Store the updated list
	db.PutEntity(key, &database.DataEntity{Data: lst})
	db.addAof(utils.ToCmdLineWithName("RPUSH", args...))

	// Return the new length of the list
	return reply.MakeIntReply(int64(lst.Len()))
}

// execLPop implements the LPOP command: Removes and returns the first element of the list stored at key
// LPOP key
func execLPop(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// Get list
	lst, exists := getAsList(db, key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	if lst == nil { // Key exists but is not a list
		return reply.MakeWrongTypeErrReply()
	}

	// Check if list is empty
	if lst.Len() == 0 {
		return reply.MakeNullBulkReply()
	}

	// Remove and get the first element
	element := lst.Front()
	lst.Remove(element)
	value := element.Value.([]byte)

	// If list becomes empty after pop, remove the key
	if lst.Len() == 0 {
		db.Remove(key)
	} else {
		// Otherwise update the list in database
		db.PutEntity(key, &database.DataEntity{Data: lst})
	}

	db.addAof(utils.ToCmdLineWithName("LPOP", args...))
	return reply.MakeBulkReply(value)
}

// execRPop implements the RPOP command: Removes and returns the last element of the list stored at key
// RPOP key
func execRPop(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// Get list
	lst, exists := getAsList(db, key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	if lst == nil { // Key exists but is not a list
		return reply.MakeWrongTypeErrReply()
	}

	// Check if list is empty
	if lst.Len() == 0 {
		return reply.MakeNullBulkReply()
	}

	// Remove and get the last element
	element := lst.Back()
	lst.Remove(element)
	value := element.Value.([]byte)

	// If list becomes empty after pop, remove the key
	if lst.Len() == 0 {
		db.Remove(key)
	} else {
		// Otherwise update the list in database
		db.PutEntity(key, &database.DataEntity{Data: lst})
	}

	db.addAof(utils.ToCmdLineWithName("RPOP", args...))
	return reply.MakeBulkReply(value)
}

// execLRange implements the LRANGE command: Returns the specified elements of the list stored at key
// LRANGE key start stop
func execLRange(db *DB, args [][]byte) resp.Reply {
	// Parse arguments
	key := string(args[0])
	start, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeStandardErrorReply("value is not an integer or out of range")
	}
	stop, err := strconv.ParseInt(string(args[2]), 10, 64)
	if err != nil {
		return reply.MakeStandardErrorReply("value is not an integer or out of range")
	}

	// Get list
	lst, exists := getAsList(db, key)
	if !exists {
		return reply.MakeEmptyMultiBulkReply()
	}
	if lst == nil { // Key exists but is not a list
		return reply.MakeWrongTypeErrReply()
	}

	// Convert negative indices
	size := int64(lst.Len())
	if start < 0 {
		start = size + start
	}
	if stop < 0 {
		stop = size + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= size {
		stop = size - 1
	}
	if start > stop {
		return reply.MakeEmptyMultiBulkReply()
	}

	// Collect elements
	elements := make([][]byte, 0, stop-start+1)
	index := int64(0)
	for e := lst.Front(); e != nil; e = e.Next() {
		if index >= start && index <= stop {
			elements = append(elements, e.Value.([]byte))
		} else if index > stop {
			break
		}
		index++
	}

	return reply.MakeMultiBulkReply(elements)
}

// execLLen implements the LLEN command: Returns the length of the list stored at key
// LLEN key
func execLLen(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])

	// Get list
	lst, exists := getAsList(db, key)
	if !exists {
		return reply.MakeIntReply(0)
	}
	if lst == nil { // Key exists but is not a list
		return reply.MakeWrongTypeErrReply()
	}

	return reply.MakeIntReply(int64(lst.Len()))
}

// execLIndex implements the LINDEX command: Returns the element at index in the list stored at key
// LINDEX key index
func execLIndex(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	index, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeStandardErrorReply("value is not an integer or out of range")
	}

	// Get list
	lst, exists := getAsList(db, key)
	if !exists {
		return reply.MakeNullBulkReply()
	}
	if lst == nil { // Key exists but is not a list
		return reply.MakeWrongTypeErrReply()
	}

	size := int64(lst.Len())
	if index < 0 {
		index = size + index
	}
	if index < 0 || index >= size {
		return reply.MakeNullBulkReply()
	}

	// Find the element at the specified index
	var element *list.Element
	if index < size/2 {
		// If index is in the first half, iterate from front
		element = lst.Front()
		for i := int64(0); i < index; i++ {
			element = element.Next()
		}
	} else {
		// If index is in the second half, iterate from back
		element = lst.Back()
		for i := size - 1; i > index; i-- {
			element = element.Prev()
		}
	}

	return reply.MakeBulkReply(element.Value.([]byte))
}

// execLSet implements the LSET command: Sets the list element at index to value
// LSET key index value
func execLSet(db *DB, args [][]byte) resp.Reply {
	key := string(args[0])
	index, err := strconv.ParseInt(string(args[1]), 10, 64)
	if err != nil {
		return reply.MakeStandardErrorReply("value is not an integer or out of range")
	}
	value := args[2]

	// Get list
	lst, exists := getAsList(db, key)
	if !exists {
		return reply.MakeStandardErrorReply("no such key")
	}
	if lst == nil { // Key exists but is not a list
		return reply.MakeWrongTypeErrReply()
	}

	size := int64(lst.Len())
	if index < 0 {
		index = size + index
	}
	if index < 0 || index >= size {
		return reply.MakeStandardErrorReply("index out of range")
	}

	// Find and update the element at the specified index
	var element *list.Element
	if index < size/2 {
		element = lst.Front()
		for i := int64(0); i < index; i++ {
			element = element.Next()
		}
	} else {
		element = lst.Back()
		for i := size - 1; i > index; i-- {
			element = element.Prev()
		}
	}
	element.Value = value

	db.PutEntity(key, &database.DataEntity{Data: lst})
	db.addAof(utils.ToCmdLineWithName("LSET", args...))
	return reply.MakeOKReply()
}

func init() {
	// Register list commands
	// Arity is negative because the command takes a variable number of arguments (key + at least one value)
	RegisterCommand("LPUSH", execLPush, -3)  // key value [value ...] -> at least 3 args
	RegisterCommand("RPUSH", execRPush, -3)  // key value [value ...] -> at least 3 args
	RegisterCommand("LPOP", execLPop, 2)     // key
	RegisterCommand("RPOP", execRPop, 2)     // key
	RegisterCommand("LRANGE", execLRange, 4) // key start stop
	RegisterCommand("LLEN", execLLen, 2)     // LLEN key -> exactly 2 args
	RegisterCommand("LINDEX", execLIndex, 3) // LINDEX key index -> exactly 3 args
	RegisterCommand("LSET", execLSet, 4)     // LSET key index value -> exactly 4 args
}
