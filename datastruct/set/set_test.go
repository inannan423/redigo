package set

import (
	"strconv"
	"testing"
)

// TestNewHashSet tests the creation of a new set structure
func TestNewHashSet(t *testing.T) {
	set := NewHashSet()

	if set == nil {
		t.Fatal("Failed to create a new hash set")
	}

	if !set.isIntset {
		t.Errorf("New hash set should use intset encoding by default")
	}

	if set.Len() != 0 {
		t.Errorf("New hash set should be empty, got %d items", set.Len())
	}
}

// TestAddAndContains tests basic add and contains operations
func TestAddAndContains(t *testing.T) {
	set := NewHashSet()

	// Test adding integers
	result := set.Add("42")
	if result != 1 {
		t.Errorf("Expected Add to return 1 for new member, got %d", result)
	}

	// Test contains for existing member
	exists := set.Contains("42")
	if !exists {
		t.Error("Expected member to exist after Add")
	}

	// Test adding duplicate
	result = set.Add("42")
	if result != 0 {
		t.Errorf("Expected Add to return 0 for existing member, got %d", result)
	}

	// Test contains for non-existing member
	exists = set.Contains("99")
	if exists {
		t.Error("Expected non-existent member to return false")
	}
}

// TestRemove tests the Remove operation
func TestRemove(t *testing.T) {
	set := NewHashSet()
	set.Add("100")
	set.Add("200")

	// Test removing existing member
	count := set.Remove("100")
	if count != 1 {
		t.Errorf("Expected Remove to return 1 for existing member, got %d", count)
	}

	// Verify member was removed
	exists := set.Contains("100")
	if exists {
		t.Error("Member should not exist after Remove")
	}

	// Test removing non-existent member
	count = set.Remove("nonexistent")
	if count != 0 {
		t.Errorf("Expected Remove to return 0 for non-existent member, got %d", count)
	}
}

// TestEncoding tests the encoding conversion from intset to hashtable
func TestEncoding(t *testing.T) {
	set := NewHashSet()

	// Initial encoding should be intset
	if !set.isIntset {
		t.Error("Initial encoding should be intset")
	}

	// Add string member to trigger conversion
	set.Add("abc")

	// Encoding should now be hashtable
	if set.isIntset {
		t.Error("Encoding should be hashtable after adding non-integer")
	}

	// Verify data integrity after conversion
	if !set.Contains("abc") {
		t.Error("Member should still exist after encoding conversion")
	}

	// Test adding mixed data
	set.Add("123")
	set.Add("def")

	if !set.Contains("123") || !set.Contains("def") {
		t.Error("Set should contain all members after mixed adds")
	}
}

// TestLargeNumberOfEntries tests conversion due to many entries
func TestLargeNumberOfEntries(t *testing.T) {
	set := NewHashSet()

	// Add entries to trigger conversion based on count
	for i := 0; i < SET_MAX_INTSET_ENTRIES+1; i++ {
		set.Add(strconv.Itoa(i))
	}

	// Encoding should now be hashtable
	if set.isIntset {
		t.Error("Encoding should be hashtable after exceeding entry limit")
	}

	// Verify some random entries
	for i := 0; i < 10; i++ {
		if !set.Contains(strconv.Itoa(i)) {
			t.Error("Data integrity issue after encoding conversion")
		}
	}
}

// TestMembers tests the Members method
func TestMembers(t *testing.T) {
	set := NewHashSet()
	set.Add("100")
	set.Add("200")
	set.Add("300")

	members := set.Members()
	if len(members) != 3 {
		t.Errorf("Expected 3 members, got %d", len(members))
	}

	// Check members content (set has no guaranteed order)
	memberMap := make(map[string]bool)
	for _, m := range members {
		memberMap[m] = true
	}
	if !memberMap["100"] || !memberMap["200"] || !memberMap["300"] {
		t.Error("Members returned incorrect data")
	}
}

// TestForEach tests the ForEach method
func TestForEach(t *testing.T) {
	set := NewHashSet()
	set.Add("100")
	set.Add("200")
	set.Add("300")

	count := 0
	set.ForEach(func(member string) bool {
		count++
		return true
	})

	if count != 3 {
		t.Errorf("ForEach should iterate through all 3 members, got %d", count)
	}

	// Test early termination
	count = 0
	set.ForEach(func(member string) bool {
		count++
		return count < 2 // Stop after first member
	})

	if count != 2 {
		t.Errorf("ForEach should stop after 2 members with early termination, got %d", count)
	}
}

// TestMixedDataTypes tests set with both integer and string members
func TestMixedDataTypes(t *testing.T) {
	set := NewHashSet()

	// Add integers first
	set.Add("10")
	set.Add("20")

	// Now add a string to force conversion
	set.Add("abc")

	if set.isIntset {
		t.Error("Set should use hashtable encoding after adding non-integer")
	}

	// Check all values are preserved
	if !set.Contains("10") || !set.Contains("20") || !set.Contains("abc") {
		t.Error("Set should contain all members after conversion")
	}

	// Now add more mixed data
	set.Add("30")
	set.Add("def")

	// Check length
	if set.Len() != 5 {
		t.Errorf("Expected 5 members, got %d", set.Len())
	}
}

// TestRandomMembers tests getting random members with replacement
func TestRandomMembers(t *testing.T) {
	set := NewHashSet()
	for i := 0; i < 100; i++ {
		set.Add(strconv.Itoa(i))
	}

	// Get 10 random members (may contain duplicates)
	random := set.RandomMembers(10)
	if len(random) != 10 {
		t.Errorf("Expected 10 random members, got %d", len(random))
	}

	// All returned values should be in the set
	for _, m := range random {
		if !set.Contains(m) {
			t.Errorf("Random member %s not found in original set", m)
		}
	}

	// Get 0 random members
	random = set.RandomMembers(0)
	if len(random) != 0 {
		t.Errorf("Expected empty result for count 0, got %d items", len(random))
	}

	// Get more random members than set size
	random = set.RandomMembers(200)
	if len(random) != 200 {
		t.Errorf("Expected 200 random members, got %d", len(random))
	}
}

// TestRandomDistinctMembers tests getting distinct random members
func TestRandomDistinctMembers(t *testing.T) {
	set := NewHashSet()
	for i := 0; i < 100; i++ {
		set.Add(strconv.Itoa(i))
	}

	// Get 50 distinct random members
	random := set.RandomDistinctMembers(50)
	if len(random) != 50 {
		t.Errorf("Expected 50 random members, got %d", len(random))
	}

	// Check for uniqueness
	uniqueCheck := make(map[string]bool)
	for _, m := range random {
		if uniqueCheck[m] {
			t.Error("RandomDistinctMembers returned duplicate values")
		}
		uniqueCheck[m] = true

		// Verify member exists in original set
		if !set.Contains(m) {
			t.Errorf("Random member %s not found in original set", m)
		}
	}

	// Get 0 random members
	random = set.RandomDistinctMembers(0)
	if len(random) != 0 {
		t.Errorf("Expected empty result for count 0, got %d items", len(random))
	}

	// Get more random members than set size
	random = set.RandomDistinctMembers(200)
	if len(random) != 100 {
		t.Errorf("Expected all 100 members when count > size, got %d", len(random))
	}
}

// TestEmptySet tests operations on an empty set
func TestEmptySet(t *testing.T) {
	set := NewHashSet()

	// Test length
	if set.Len() != 0 {
		t.Errorf("Empty set should have length 0, got %d", set.Len())
	}

	// Test contains
	if set.Contains("anything") {
		t.Error("Empty set should not contain any members")
	}

	// Test members
	members := set.Members()
	if len(members) != 0 {
		t.Errorf("Empty set should return empty members list, got %d items", len(members))
	}

	// Test remove
	result := set.Remove("anything")
	if result != 0 {
		t.Errorf("Remove on empty set should return 0, got %d", result)
	}

	// Test random members
	random := set.RandomMembers(10)
	if len(random) != 0 {
		t.Errorf("RandomMembers on empty set should return empty list, got %d items", len(random))
	}
}

// TestIntSetSpecificBehavior tests behaviors that are specific to the intset implementation
func TestIntSetSpecificBehavior(t *testing.T) {
	set := NewHashSet()

	// Add integers that should be stored in intset
	set.Add("1")
	set.Add("2")
	set.Add("3")

	// Verify it's still using intset
	if !set.isIntset {
		t.Error("Set should still be using intset encoding")
	}

	// Remove a value
	set.Remove("2")

	// Check remaining values
	if !set.Contains("1") || set.Contains("2") || !set.Contains("3") {
		t.Error("Remove operation not working correctly in intset mode")
	}

	// Add a very large integer that should still work with intset
	set.Add("9223372036854775807") // Max int64
	if !set.isIntset {
		t.Error("Set should still be using intset encoding with large integers")
	}

	if !set.Contains("9223372036854775807") {
		t.Error("Set should contain max int64 value")
	}

	// Now add a non-integer to force conversion
	set.Add("abc")
	if set.isIntset {
		t.Error("Set should convert to hashtable after adding non-integer")
	}

	// Verify all values survived the conversion
	if !set.Contains("1") || !set.Contains("3") ||
		!set.Contains("9223372036854775807") || !set.Contains("abc") {
		t.Error("Some values were lost during conversion from intset to hashtable")
	}
}
