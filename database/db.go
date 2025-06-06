package database

import (
	"redigo/datastruct/dict"
	"redigo/datastruct/hash"
	"redigo/datastruct/set"
	"redigo/datastruct/zset"
	"redigo/interface/database"
	"redigo/interface/resp"
	"redigo/resp/reply"
	"strings"
	"sync"
)

// KeyLockManager manages locks for individual keys
type KeyLockManager struct {
	locks sync.Map // map[string]*sync.RWMutex
}

// NewKeyLockManager creates a new KeyLockManager instance
func NewKeyLockManager() *KeyLockManager {
	return &KeyLockManager{}
}

// Lock acquires a write lock for the given key
func (klm *KeyLockManager) Lock(key string) {
	lockInterface, _ := klm.locks.LoadOrStore(key, &sync.RWMutex{})
	lock := lockInterface.(*sync.RWMutex)
	// If the lock is locked, it will block until it can acquire the lock
	lock.Lock()
}

// Unlock releases a write lock for the given key
func (klm *KeyLockManager) Unlock(key string) {
	if lockInterface, ok := klm.locks.Load(key); ok {
		lock := lockInterface.(*sync.RWMutex)
		lock.Unlock()
	}
}

// RLock acquires a read lock for the given key
func (klm *KeyLockManager) RLock(key string) {
	lockInterface, _ := klm.locks.LoadOrStore(key, &sync.RWMutex{})
	lock := lockInterface.(*sync.RWMutex)
	lock.RLock()
}

// RUnlock releases a read lock for the given key
func (klm *KeyLockManager) RUnlock(key string) {
	if lockInterface, ok := klm.locks.Load(key); ok {
		lock := lockInterface.(*sync.RWMutex)
		lock.RUnlock()
	}
}

// CleanupLock removes the lock for a deleted key to prevent memory leaks
// This should be called when a key is permanently deleted from the database
func (klm *KeyLockManager) CleanupLock(key string) {
	// Only delete the lock if no one is using it
	// In practice, this should be called after ensuring no operations are pending on this key
	klm.locks.Delete(key)
}

type DB struct {
	index   int
	data    dict.Dict
	addAof  func(CmdLine)
	lockMgr *KeyLockManager
}

// MakeDB creates a new DB instance
func MakeDB() *DB {
	return &DB{
		index: 0,
		data:  dict.MakeSyncDict(),
		addAof: func(line CmdLine) {
			// No-op by default,
			// can be overridden by the database instance
		},
		lockMgr: NewKeyLockManager(),
	}
}

// ExecFunc is a function type that takes a DB instance and a slice of byte slices as arguments and returns a resp.Reply
// All redis commands like PING, SET, GET, etc. are implemented as functions of this type
type ExecFunc func(db *DB, args [][]byte) resp.Reply

// CmdLine is a type alias for a slice of byte slices
// It is used to represent the command line arguments passed to the ExecFunc
type CmdLine = [][]byte

// Exec executes a command on the DB instance
// It takes a connection and a command line as arguments
// It returns a resp.Reply which is the response to the command
func (db *DB) Exec(c resp.Connection, cmdLine CmdLine) resp.Reply {
	// The first element of cmdLine is the command name, like "PING", "SET", etc.
	// Convert it to lowercase to ensure case-insensitivity
	cmdName := strings.ToLower(string(cmdLine[0]))
	// Get the command from the command table using the command name
	// If the command is not found, return an error reply
	cmd, ok := cmdTable[cmdName]
	if !ok {
		return reply.MakeStandardErrorReply("ERR unknown command '" + cmdName + "'")
	}
	// Validate the number of arguments passed to the command
	if !ValidateArity(cmd.arity, cmdLine) {
		return reply.MakeArgNumErrReply(cmdName)
	}
	// Execute the command and return the response
	return cmd.exec(db, cmdLine[1:])
}

// ValidateArity checks if the number of arguments passed to a command is valid
func ValidateArity(arity int, args [][]byte) bool {
	// Check if the number of arguments is less than the required arity
	if arity >= 0 {
		return len(args) == arity
	} else {
		// If the arity is negative, it means the command takes a variable number of arguments
		// Check if the number of arguments is within the valid range
		return len(args) >= -arity
	}
}

// GetEntity returns DataEntity bind to the given key
func (db *DB) GetEntity(key string) (*database.DataEntity, bool) {
	raw, ok := db.data.Get(key)
	if !ok {
		return nil, false
	}
	entity, _ := raw.(*database.DataEntity)
	return entity, true
}

// PutEntity stores the given DataEntity in the database
func (db *DB) PutEntity(key string, entity *database.DataEntity) int {
	return db.data.Put(key, entity)
}

// PutIfExists edit the given DataEntity in the database
func (db *DB) PutIfExists(key string, entity *database.DataEntity) int {
	return db.data.PutIfExists(key, entity)
}

// PutIfAbsent stores the given DataEntity in the database if it doesn't already exist
func (db *DB) PutIfAbsent(key string, entity *database.DataEntity) int {
	return db.data.PutIfAbsent(key, entity)
}

// Remove deletes the DataEntity associated with the given key from the database
func (db *DB) Remove(key string) int {
	result := db.data.Remove(key)
	// Clean up the lock for the deleted key to prevent memory leaks
	if result > 0 {
		db.lockMgr.CleanupLock(key)
	}
	return result
}

// GetAsHash retrieves the DataEntity associated with the given key and checks if it is a hash
func (db *DB) getAsHash(key string) (*hash.Hash, bool) {
	entity, ok := db.GetEntity(key)
	if !ok {
		return nil, false
	}
	hash, ok := entity.Data.(*hash.Hash)
	if !ok {
		return nil, true // Key exists but is not a hash
	}
	return hash, true
}

// getOrCreateHash retrieves the DataEntity associated with the given key and creates a new hash if it doesn't exist
func (db *DB) getOrCreateHash(key string) (*hash.Hash, bool) {
	hashObj, ok := db.getAsHash(key)
	if ok {
		return hashObj, true
	}

	hashObj = hash.MakeHash()
	db.PutEntity(key, &database.DataEntity{
		Data: hashObj,
	})
	return hashObj, false
}

// getAsSet returns a set.Set from database
func getAsSet(db *DB, key string) (set.Set, reply.ErrorReply) {
	entity, exists := db.GetEntity(key)
	if !exists {
		return nil, nil
	}

	setObj, ok := entity.Data.(set.Set)
	if !ok {
		return nil, reply.MakeWrongTypeErrReply()
	}
	return setObj, nil
}

// getOrInitSet returns a set.Set for the given key
// creates a new one if it doesn't exist
func getOrInitSet(db *DB, key string) (set.Set, bool, reply.ErrorReply) {
	setObj, errReply := getAsSet(db, key)
	if errReply != nil {
		return nil, false, errReply
	}

	isNew := false
	if setObj == nil {
		setObj = set.NewHashSet()
		isNew = true
	}

	return setObj, isNew, nil
}

// getAsZSet retrieves the ZSet stored at key, or creates a new one if it doesn't exist
func getAsZSet(db *DB, key string) (zset.ZSet, bool) {
	// Get entity from database
	entity, exists := db.GetEntity(key)
	if !exists {
		return zset.NewZSet(), false
	}

	// Check if entity is a ZSet
	zsetObj, ok := entity.Data.(zset.ZSet)
	if !ok {
		return nil, true // Key exists but is not a ZSet
	}

	return zsetObj, true
}

// Removes deletes the DataEntity associated with the given keys from the database
func (db *DB) Removes(keys ...string) int {
	deleted := 0
	for _, key := range keys {
		_, ok := db.data.Get(key)
		if ok {
			db.data.Remove(key)
			// Clean up the lock for the deleted key to prevent memory leaks
			db.lockMgr.CleanupLock(key)
			deleted++
		}
	}
	return deleted
}

// Flush clears the database by removing all DataEntity objects
func (db *DB) Flush() {
	db.data.Clear()
	// Clear all locks when flushing the database
	db.lockMgr.locks = sync.Map{}
}

// WithKeyLock executes the given function with a write lock on the specified key
func (db *DB) WithKeyLock(key string, fn func()) {
	db.lockMgr.Lock(key)
	defer db.lockMgr.Unlock(key)
	fn()
}

// WithKeyRLock executes the given function with a read lock on the specified key
func (db *DB) WithKeyRLock(key string, fn func()) {
	db.lockMgr.RLock(key)
	defer db.lockMgr.RUnlock(key)
	fn()
}

// WithKeyLockReturn executes the given function with a write lock on the specified key and returns the result
func (db *DB) WithKeyLockReturn(key string, fn func() interface{}) interface{} {
	db.lockMgr.Lock(key)
	defer db.lockMgr.Unlock(key)
	return fn()
}
