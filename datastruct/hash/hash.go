package hash

const (
	// If the number of entries in the hash exceeds this value, it will be converted to a hash table
	hashMaxListpackEntries = 512
	// If the length of the value in the hash exceeds this value, it will be converted to a hash table
	hashMaxListpackValue = 64
)

// The encoding types for the hash
const (
	encodingListpack = iota
	encodingHashTable
)

type Hash struct {
	encoding int         // The encoding type of the hash
	listpack [][2]string // Using Go slice to simulate the listpack
	dict     map[string]string
}

// MakeHash creates a new Hash instance
func MakeHash() *Hash {
	return &Hash{
		encoding: encodingListpack, // Use listpack encoding by default
		listpack: make([][2]string, 0),
		dict:     make(map[string]string),
	}
}

// Get retrieves the value associated with the given key from the hash
func (h *Hash) Get(field string) (val string, exists bool) {
	// If using listpack encoding, search in the listpack
	if h.encoding == encodingListpack {
		for _, entry := range h.listpack {
			if entry[0] == field {
				return entry[1], true
			}
		}
		return "", false
	}

	val, exists = h.dict[field]
	return
}

// Set sets the value for the given key in the hash
// If the field already exists, it updates the value and returns 0 else 1 if it is a new entry
func (h *Hash) Set(field, value string) int {
	if h.encoding == encodingListpack {
		// If the size of the listpack exceeds the maximum entries or the length of the field or value exceeds the maximum value, convert to hash table
		if len(h.listpack) >= hashMaxListpackEntries || len(field) > hashMaxListpackValue || len(value) > hashMaxListpackValue {
			h.convertToHashTable()
		}
	}

	if h.encoding == encodingListpack {
		// Check if the field already exists in the listpack
		for i, entry := range h.listpack {
			if entry[0] == field {
				h.listpack[i][1] = value
				return 0 // Updated existing entry
			}
		}

		// Add new entry
		h.listpack = append(h.listpack, [2]string{field, value})
		return 1
	}

	_, exsists := h.dict[field]
	h.dict[field] = value
	if exsists {
		return 0 // Updated existing entry
	}
	return 1 // Added new entry
}

// Delete removes the given field from the hash
func (h *Hash) Delete(field string) int {
	count := 0

	if h.encoding == encodingListpack {
		for i, entry := range h.listpack {
			if entry[0] == field {
				// Delete the entry and move the last entry to the current position to reduce the size
				// Because hash doesn't need to maintain order, we can just swap the last entry with the current one
				lastIndex := len(h.listpack) - 1
				h.listpack[i] = h.listpack[lastIndex]
				h.listpack = h.listpack[:lastIndex]
				count++
				break
			}
		}
	} else {
		// Delete the field from the hash table
		if _, exists := h.dict[field]; exists {
			delete(h.dict, field)
			count++
		}
	}

	return count
}

// Len returns the number of entries in the hash
func (h *Hash) Len() int {
	if h.encoding == encodingListpack {
		return len(h.listpack)
	}
	return len(h.dict)
}

// GetAll returns all the fields and values in the hash
func (h *Hash) GetAll() map[string]string {
	result := make(map[string]string)

	if h.encoding == encodingListpack {
		for _, entry := range h.listpack {
			result[entry[0]] = entry[1]
		}
	} else {
		for field, value := range h.dict {
			result[field] = value
		}
	}
	return result
}

// Fields returns all the fields in the hash
func (h *Hash) Fields() []string {
	if h.encoding == encodingListpack {
		fields := make([]string, len(h.listpack))
		for i, entry := range h.listpack {
			fields[i] = entry[0]
		}
		return fields
	}

	fields := make([]string, 0, len(h.dict))
	for field := range h.dict {
		fields = append(fields, field)
	}
	return fields
}

// Values returns all the values in the hash
func (h *Hash) Values() []string {
	if h.encoding == encodingListpack {
		values := make([]string, len(h.listpack))
		for i, entry := range h.listpack {
			values[i] = entry[1]
		}
		return values
	}

	values := make([]string, 0, len(h.dict))
	for _, value := range h.dict {
		values = append(values, value)
	}
	return values
}

// Exists checks if the field exists in the hash
func (h *Hash) Exists(field string) bool {
	_, exists := h.Get(field)
	return exists
}

// convertToHashTable converts the hash from listpack to hash table encoding
func (h *Hash) convertToHashTable() {
	if h.encoding == encodingHashTable {
		return
	}

	h.dict = make(map[string]string, len(h.listpack))

	for _, entry := range h.listpack {
		h.dict[entry[0]] = entry[1]
	}

	h.encoding = encodingHashTable

	h.listpack = nil // Clear the listpack to free up memory
}

// Encoding returns the encoding type of the hash
func (h *Hash) Encoding() int {
	return h.encoding
}

// Clear clears all entries in the hash
func (h *Hash) Clear() {
	h.listpack = nil
	h.dict = nil
	h.encoding = encodingListpack
}
