package set

import (
	"encoding/binary"
	"fmt"
	"math"
)

// Dymamic encoding for intset
// According to the size of the intset, we can use different encoding
// types to save space.
const (
	INTSET_ENC_INT16 = 2
	INTSET_ENC_INT32 = 4
	INTSET_ENC_INT64 = 8
)

type IntSet struct {
	encoding uint32
	length   uint32
	contents []byte
}

// NewIntSet creates a new IntSet with the given encoding
func NewIntSet() *IntSet {
	return &IntSet{
		encoding: INTSET_ENC_INT16, // Default encoding is 16-bit integer
		length:   0,
		contents: make([]byte, 0),
	}
}

// Len returns the number of elements in the set
func (is *IntSet) Len() int {
	return int(is.length)
}

// Add adds an integer to the set
func (is *IntSet) Add(value int64) bool {
	// Check if we need to upgrade encoding
	var requiredEncoding uint32
	if value < math.MinInt16 || value > math.MaxInt16 {
		if value < math.MinInt32 || value > math.MaxInt32 {
			requiredEncoding = INTSET_ENC_INT64
		} else {
			requiredEncoding = INTSET_ENC_INT32
		}
	} else {
		requiredEncoding = INTSET_ENC_INT16
	}

	// Upgrade encoding if necessary
	if requiredEncoding > is.encoding {
		is.upgradeEncoding(requiredEncoding)
	}

	// Check if value already exists
	pos := is.findPosition(value)
	if pos >= 0 {
		return false // Value already exists
	}

	// Insert value at the correct position
	pos = -pos - 1 // Convert to insertion position
	is.insertAt(pos, value)
	return true
}

// Contains checks if the set contains the given value
func (is *IntSet) Contains(value int64) bool {
	pos := is.findPosition(value)
	return pos >= 0 // Value found
}

// Remove removes a value from the set
func (is *IntSet) Remove(value int64) bool {
	pos := is.findPosition(value)
	if pos < 0 {
		return false // Value not found
	}

	is.removeAt(pos)
	return true
}

// Helper Methods

// upgradeEncoding upgrades the encoding of the IntSet if necessary
func (is *IntSet) upgradeEncoding(newEncoding uint32) {
	if newEncoding <= is.encoding {
		return
	}

	// Save old values
	oldValues := is.ToSlice()

	// Reset and use new encoding
	is.encoding = newEncoding
	is.length = 0
	is.contents = make([]byte, 0, len(oldValues)*int(newEncoding))

	// Re-add all values with new encoding
	for _, v := range oldValues {
		is.Add(v)
	}
}

// ToSlice returns all elements as a slice
func (is *IntSet) ToSlice() []int64 {
	result := make([]int64, is.length)
	for i := uint32(0); i < is.length; i++ {
		result[i] = is.getValueAt(i)
	}
	return result
}

// findPosition finds the position of the value in the set
// Returns the index of the value if found, or the index where it should be inserted
func (is *IntSet) findPosition(value int64) int {
	// Binary search to find position
	low, high := 0, int(is.length)-1

	for low <= high {
		mid := (low + high) / 2
		midVal := is.getValueAt(uint32(mid))

		if midVal < value {
			low = mid + 1
		} else if midVal > value {
			high = mid - 1
		} else {
			return mid // Found
		}
	}

	return -(low + 1) // Not found, return insertion point
}

// getValueAt retrieves the value at the given index
func (is *IntSet) getValueAt(index uint32) int64 {
	if index >= is.length {
		panic(fmt.Sprintf("Index out of bounds: %d", index))
	}

	offset := index * is.encoding
	switch is.encoding {
	case INTSET_ENC_INT16:
		return int64(int16(binary.LittleEndian.Uint16(is.contents[offset:])))
	case INTSET_ENC_INT32:
		return int64(int32(binary.LittleEndian.Uint32(is.contents[offset:])))
	case INTSET_ENC_INT64:
		return int64(binary.LittleEndian.Uint64(is.contents[offset:]))
	}

	panic("Invalid encoding")
}

// insertAt inserts a value at the specified position
func (is *IntSet) insertAt(pos int, value int64) {
	// Expand contents
	oldLen := len(is.contents)
	newLen := oldLen + int(is.encoding)
	if cap(is.contents) < newLen {
		newContents := make([]byte, newLen, newLen*2)
		copy(newContents, is.contents)
		is.contents = newContents
	} else {
		is.contents = is.contents[:newLen]
	}

	// Shift elements to make space
	offset := pos * int(is.encoding)
	if pos < int(is.length) {
		// Move all the elements after pos backward
		copy(is.contents[offset+int(is.encoding):], is.contents[offset:oldLen])
	}

	// Insert new element
	switch is.encoding {
	case INTSET_ENC_INT16:
		binary.LittleEndian.PutUint16(is.contents[offset:], uint16(value))
	case INTSET_ENC_INT32:
		binary.LittleEndian.PutUint32(is.contents[offset:], uint32(value))
	case INTSET_ENC_INT64:
		binary.LittleEndian.PutUint64(is.contents[offset:], uint64(value))
	}

	is.length++
}

// removeAt removes the value at the specified position
func (is *IntSet) removeAt(pos int) {
	if pos < 0 || pos >= int(is.length) {
		return
	}

	offset := pos * int(is.encoding)
	endOffset := int(is.length) * int(is.encoding)

	// Move elements after pos forward
	copy(is.contents[offset:], is.contents[offset+int(is.encoding):endOffset])

	// Shrink contents
	is.contents = is.contents[:endOffset-int(is.encoding)]
	is.length--
}

// ForEach iterates over all elements in the set
func (is *IntSet) ForEach(consumer func(value int64) bool) {
	for i := uint32(0); i < is.length; i++ {
		if !consumer(is.getValueAt(i)) {
			break
		}
	}
}
