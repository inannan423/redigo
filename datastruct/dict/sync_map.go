package dict

import (
	"sync"
)

type SyncDict struct {
	m sync.Map
}

// MakeSyncDict creates a new SyncDict instance
func MakeSyncDict() *SyncDict {
	return &SyncDict{}
}

// Get returns the value associated with the given key and a boolean indicating if the key exists
func (dict *SyncDict) Get(key string) (val interface{}, exists bool) {
	// Get the value by key
	if value, ok := dict.m.Load(key); ok {
		return value, true
	}
	return nil, false
}

// Len returns the number of key-value pairs in the dictionary
func (dict *SyncDict) Len() int {
	count := 0
	// Iterate over all key-value pairs and count them
	dict.m.Range(func(key, value interface{}) bool {
		count++
		return true
	})
	return count
}

// Put adds a key-value pair to the dictionary, if the key already exists, return 1 else 0
func (dict *SyncDict) Put(key string, val interface{}) (result int) {
	_, exists := dict.m.Load(key)
	// Store the key-value pair
	dict.m.Store(key, val)
	// Return the count of pairs
	if exists {
		return 0
	}
	return 1
}

// PutIfAbsent adds a key-value pair to the dictionary if the key does not exist, return 1 if it exists, else 0
func (dict *SyncDict) PutIfAbsent(key string, val interface{}) (result int) {
	_, exists := dict.m.Load(key)
	if exists {
		return 0
	}
	// Store the key-value pair
	dict.m.Store(key, val)
	return 1
}

// PutIfExists adds a key-value pair to the dictionary if the key exists, return 1 if it does not exist, else 0
func (dict *SyncDict) PutIfExists(key string, val interface{}) (result int) {
	_, exists := dict.m.Load(key)
	if !exists {
		return 0
	}
	// Store the key-value pair
	dict.m.Store(key, val)
	return 1
}

// Remove removes a key-value pair from the dictionary, return the count of pairs were removed
func (dict *SyncDict) Remove(key string) (result int) {
	_, exists := dict.m.Load(key)
	if !exists {
		return 0
	}
	// Delete the key-value pair
	dict.m.Delete(key)
	return 1
}

// ForEach iterates over all key-value pairs in the dictionary and applies the consumer function to each pair
func (dict *SyncDict) ForEach(consumer Consumer) {
	// Iterate over all key-value pairs and apply the consumer function
	dict.m.Range(func(key, value interface{}) bool {
		consumer(key.(string), value)
		// Always return true to continue iteration
		return true
	})
}

// Keys returns a slice of all keys in the dictionary
func (dict *SyncDict) Keys() []string {
	keys := make([]string, dict.Len())
	// Iterate over all key-value pairs and collect the keys
	dict.m.Range(func(key, value interface{}) bool {
		keys = append(keys, key.(string))
		return true
	})
	return keys
}

// RandomKeys returns a slice of n random keys from the dictionary
// Due to m.Range doesn't guarantee the order of iteration, we can use this feature to get random keys
// Note: This method may not be truly random, but it will give different keys each time
// Duplicate keys may be returned
func (dict *SyncDict) RandomKeys(n int) []string {
	keys := make([]string, dict.Len())
	for i := 0; i < n; i++ {
		// Randomly select a key from the dictionary
		dict.m.Range(func(key, value interface{}) bool {
			keys = append(keys, key.(string))
			return false
		})
	}
	return keys
}

// RandomDistinctKeys returns a slice of n distinct random keys from the dictionary
func (dict *SyncDict) RandomDistinctKeys(n int) []string {
	result := make([]string, dict.Len())

	i := 0

	// Iterate over all key-value pairs and collect the keys
	dict.m.Range(func(key, value interface{}) bool {
		result[i] = key.(string)
		i++
		return i != n
	})
	return result
}

// Clear clears all key-value pairs in the dictionary
func (dict *SyncDict) Clear() {
	*dict = *MakeSyncDict()
}
