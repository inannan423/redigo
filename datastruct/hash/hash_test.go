package hash

import (
	"strconv"
	"testing"
)

// TestMakeHash tests the creation of a new hash structure
func TestMakeHash(t *testing.T) {
	h := MakeHash()

	if h == nil {
		t.Fatal("Failed to create a new hash")
	}

	if h.encoding != encodingListpack {
		t.Errorf("New hash should use listpack encoding by default, got %d", h.encoding)
	}

	if len(h.listpack) != 0 {
		t.Errorf("New hash should have empty listpack, got %d items", len(h.listpack))
	}
}

// TestSetAndGet tests basic set and get operations
func TestSetAndGet(t *testing.T) {
	h := MakeHash()

	// Test setting a new field
	result := h.Set("name", "test")
	if result != 1 {
		t.Errorf("Expected Set to return 1 for new field, got %d", result)
	}

	// Test getting an existing field
	value, exists := h.Get("name")
	if !exists {
		t.Error("Expected field to exist after Set")
	}
	if value != "test" {
		t.Errorf("Expected value to be 'test', got '%s'", value)
	}

	// Test updating an existing field
	result = h.Set("name", "updated")
	if result != 0 {
		t.Errorf("Expected Set to return 0 for existing field, got %d", result)
	}

	value, exists = h.Get("name")
	if !exists {
		t.Error("Expected field to exist after update")
	}
	if value != "updated" {
		t.Errorf("Expected value to be 'updated', got '%s'", value)
	}

	// Test getting non-existent field
	_, exists = h.Get("nonexistent")
	if exists {
		t.Error("Expected non-existent field to return false")
	}
}

// TestDelete tests the Delete operation
func TestDelete(t *testing.T) {
	h := MakeHash()
	h.Set("field1", "value1")
	h.Set("field2", "value2")

	// Test deleting existing field
	count := h.Delete("field1")
	if count != 1 {
		t.Errorf("Expected Delete to return 1 for existing field, got %d", count)
	}

	// Verify field was deleted
	_, exists := h.Get("field1")
	if exists {
		t.Error("Field should not exist after Delete")
	}

	// Test deleting non-existent field
	count = h.Delete("nonexistent")
	if count != 0 {
		t.Errorf("Expected Delete to return 0 for non-existent field, got %d", count)
	}
}

// TestEncoding tests the encoding conversion from listpack to hashtable
func TestEncoding(t *testing.T) {
	h := MakeHash()

	// Initial encoding should be listpack
	if h.Encoding() != encodingListpack {
		t.Errorf("Initial encoding should be listpack, got %d", h.Encoding())
	}

	// Add entries below threshold
	for i := 0; i < 10; i++ {
		h.Set("key"+strconv.Itoa(i), "value")
	}

	// Encoding should still be listpack
	if h.Encoding() != encodingListpack {
		t.Errorf("Encoding should remain listpack with few entries, got %d", h.Encoding())
	}

	// Add large value to trigger conversion
	largeValue := string(make([]byte, hashMaxListpackValue+1))
	h.Set("largeKey", largeValue)

	// Encoding should now be hashtable
	if h.Encoding() != encodingHashTable {
		t.Errorf("Encoding should be hashtable after large value, got %d", h.Encoding())
	}

	// Verify data integrity after conversion
	for i := 0; i < 10; i++ {
		val, exists := h.Get("key" + strconv.Itoa(i))
		if !exists || val != "value" {
			t.Errorf("Data integrity issue after encoding conversion")
		}
	}
}

// TestLargeNumberOfEntries tests conversion due to many entries
func TestLargeNumberOfEntries(t *testing.T) {
	h := MakeHash()

	// Add entries to trigger conversion based on count
	for i := 0; i < hashMaxListpackEntries+1; i++ {
		h.Set("key"+strconv.Itoa(i), "value")
	}

	// Encoding should now be hashtable
	if h.Encoding() != encodingHashTable {
		t.Errorf("Encoding should be hashtable after exceeding entry limit, got %d", h.Encoding())
	}

	// Verify some random entries
	for i := 0; i < 10; i++ {
		val, exists := h.Get("key" + strconv.Itoa(i))
		if !exists || val != "value" {
			t.Errorf("Data integrity issue after encoding conversion")
		}
	}
}

// TestOtherOperations tests remaining hash operations
func TestOtherOperations(t *testing.T) {
	h := MakeHash()
	h.Set("field1", "value1")
	h.Set("field2", "value2")

	// Test Len
	if h.Len() != 2 {
		t.Errorf("Expected length 2, got %d", h.Len())
	}

	// Test Exists
	if !h.Exists("field1") {
		t.Error("Expected field1 to exist")
	}
	if h.Exists("nonexistent") {
		t.Error("Expected nonexistent field to not exist")
	}

	// Test GetAll
	all := h.GetAll()
	if len(all) != 2 || all["field1"] != "value1" || all["field2"] != "value2" {
		t.Error("GetAll returned incorrect data")
	}

	// Test Fields
	fields := h.Fields()
	if len(fields) != 2 {
		t.Errorf("Expected 2 fields, got %d", len(fields))
	}

	// Fields content check (order may vary)
	fieldMap := make(map[string]bool)
	for _, f := range fields {
		fieldMap[f] = true
	}
	if !fieldMap["field1"] || !fieldMap["field2"] {
		t.Error("Fields returned incorrect data")
	}

	// Test Values
	values := h.Values()
	if len(values) != 2 {
		t.Errorf("Expected 2 values, got %d", len(values))
	}

	// Values content check (order may vary)
	valueMap := make(map[string]bool)
	for _, v := range values {
		valueMap[v] = true
	}
	if !valueMap["value1"] || !valueMap["value2"] {
		t.Error("Values returned incorrect data")
	}
}
