package dict

type Consumer func(key string, val interface{}) bool // function type for iterating over key-value pairs

type Dict interface {
	Get(key string) (val interface{}, exists bool)        // get value by key, return the value and a boolean indicating if the key exists
	Len() int                                             // get the number of key-value pairs
	Put(key string, val interface{}) (result int)         // put key-value pair, if exists, modify the value, return 0, if doesn't exist, add it, return 1
	PutIfAbsent(key string, val interface{}) (result int) // put key-value pair if absent, return 0, if exists, return 1
	PutIfExists(key string, val interface{}) (result int) // put key-value pair if exists, return 0, if absent, return 1
	Remove(key string) (result int)                       // remove key-value pair, return the count of pairs
	ForEach(consumer Consumer)                            // iterate over all key-value pairs
	Keys() []string                                       // get all keys
	RandomKeys(n int) []string                            // get n random keys
	RandomDistinctKeys(n int) []string                    // get n distinct random keys
	Clear()                                               // clear all key-value pairs
}
