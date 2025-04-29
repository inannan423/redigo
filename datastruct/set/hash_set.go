package set

import (
	"strconv"
	"time"

	"math/rand"
)

const (
	SET_MAX_INTSET_ENTRIES = 512
)

type HashSet struct {
	dict     map[string]struct{}
	intset   *IntSet
	isIntset bool
}

// NewHashSet creates a new HashSet
func NewHashSet() *HashSet {
	return &HashSet{
		dict:     make(map[string]struct{}),
		intset:   NewIntSet(),
		isIntset: true, // Default to intset
	}
}

// Add adds an integer to the set
func (set *HashSet) Add(member string) int {
	if set.isIntset {
		if val, err := strconv.ParseInt(member, 10, 64); err == nil {
			if ok := set.intset.Add(val); ok {
				if set.intset.Len() > SET_MAX_INTSET_ENTRIES {
					set.convertToHashTable()
				}
				return 1
			}
			return 0
		} else {
			// The input is not a valid integer, so we need to convert to hash table
			// to store non-integer values
			set.convertToHashTable()
		}
	}

	if _, exists := set.dict[member]; exists {
		return 0 // Already exists
	}
	set.dict[member] = struct{}{}
	return 1 // Added successfully
}

// Remove removes an integer from the set
func (set *HashSet) Remove(member string) int {
	if set.isIntset {
		// If the input is an integer, we can remove it from the intset
		if val, err := strconv.ParseInt(member, 10, 64); err == nil {
			if ok := set.intset.Remove(val); ok {
				return 1
			}
			return 0
		}
		return 0 // Not an integer, cannot remove from intset
	}

	if _, exists := set.dict[member]; !exists {
		return 0 // Not found
	}

	delete(set.dict, member)
	return 1 // Removed successfully
}

// Contains checks if the set contains the given value
func (set *HashSet) Contains(member string) bool {
	if set.isIntset {
		if val, err := strconv.ParseInt(member, 10, 64); err == nil {
			return set.intset.Contains(val)
		}
		return false // Not an integer
	}
	_, exists := set.dict[member]
	return exists
}

// Members returns all members of the set
func (set *HashSet) Members() []string {
	if set.isIntset {
		members := make([]string, 0, set.intset.Len())
		set.intset.ForEach(func(value int64) bool {
			members = append(members, strconv.FormatInt(value, 10))
			return true
		})
		return members
	}

	members := make([]string, 0, len(set.dict))
	for member := range set.dict {
		members = append(members, member)
	}
	return members
}

// Len returns the number of members in the set
func (set *HashSet) Len() int {
	if set.isIntset {
		return set.intset.Len()
	}
	return len(set.dict)
}

// ForEach iterates over all members of the set
func (set *HashSet) ForEach(consumer func(member string) bool) {
	if set.isIntset {
		set.intset.ForEach(func(value int64) bool {
			return consumer(strconv.FormatInt(value, 10))
		})
	} else {
		for member := range set.dict {
			if !consumer(member) {
				break
			}
		}
	}
}

// RandomMembers returns random members from the set
func (set *HashSet) RandomMembers(count int) []string {
	size := set.Len()
	if count <= 0 || size == 0 {
		return []string{}
	}

	res := make([]string, count)
	members := set.Members()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))

	for i := 0; i < count; i++ {
		res[i] = members[r.Intn(size)] // Use r.Intn(size) to get a random index
	}
	return res
}

// RandomDistinctMembers returns distinct random members from the set
func (set *HashSet) RandomDistinctMembers(count int) []string {
	size := set.Len()
	if count <= 0 || size == 0 {
		return []string{}
	}

	if count >= size {
		return set.Members() // Return all members if count is greater than or equal to size
	}

	members := set.Members()
	r := rand.New(rand.NewSource(time.Now().UnixNano()))
	r.Shuffle(len(members), func(i, j int) {
		members[i], members[j] = members[j], members[i]
	})

	return members[:count] // Return the first 'count' members after shuffling
}

// convertToHashTable converts the intset to a hash table
func (set *HashSet) convertToHashTable() {
	if !set.isIntset {
		return // Already a hash table
	}

	// Copy elements from intset to hash table
	set.intset.ForEach(func(value int64) bool {
		set.dict[strconv.FormatInt(value, 10)] = struct{}{}
		return true
	})

	set.isIntset = false
}

// IsIntSet checks if the set is an IntSet
func (set *HashSet) IsIntSet() bool {
	return set.isIntset
}
