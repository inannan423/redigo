package skiplist

import (
	"math/rand"
	"time"
)

const maxLevel = 16 // Maximum number of levels in the skip list

// Node represents a node in the skip list
type Node struct {
	Member  string
	Score   float64
	Forward []*Node // Forward points at different levels
}

// SkipList represents a skip list
type SkipList struct {
	header *Node // Header node
	tail   *Node // Tail node
	level  int   // Current max level of the skip list
	length int   // Length of the skip list
	rand   *rand.Rand
}

// New SkipList creates a new skip list
func NewSkipList() *SkipList {
	header := &Node{
		Forward: make([]*Node, maxLevel),
	}
	return &SkipList{
		header: header,
		level:  1,
		rand:   rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// randomLevel generates a random level for the new node
func (sl *SkipList) randomLevel() int {
	level := 1
	// Increase level with 25% probability
	for level < maxLevel && sl.rand.Float32() < 0.25 {
		level++
	}
	return level
}

// Insert inserts a new member with the given score into the skip list
func (sl *SkipList) Insert(member string, score float64) {
	update := make([]*Node, maxLevel)
	x := sl.header

	// Find position to insert
	for i := sl.level - 1; i >= 0; i-- {
		for x.Forward[i] != nil &&
			(x.Forward[i].Score < score ||
				(x.Forward[i].Score == score && x.Forward[i].Member < member)) {
			x = x.Forward[i]
		}
		update[i] = x
	}

	// Generate random level for new node
	level := sl.randomLevel()

	// If new level is higher than current, update header's forward pointers
	if level > sl.level {
		for i := sl.level; i < level; i++ {
			update[i] = sl.header
		}
		sl.level = level
	}

	// Create new node
	x = &Node{
		Member:  member,
		Score:   score,
		Forward: make([]*Node, level),
	}

	// Insert node at all levels
	for i := 0; i < level; i++ {
		x.Forward[i] = update[i].Forward[i]
		update[i].Forward[i] = x
	}

	// Update tail if necessary
	if x.Forward[0] == nil {
		sl.tail = x
	}

	sl.length++
}

// Delete removes an element from the skip list
func (sl *SkipList) Delete(member string, score float64) bool {
	update := make([]*Node, maxLevel)
	x := sl.header

	// Find position to delete
	for i := sl.level - 1; i >= 0; i-- {
		for x.Forward[i] != nil &&
			(x.Forward[i].Score < score ||
				(x.Forward[i].Score == score && x.Forward[i].Member < member)) {
			x = x.Forward[i]
		}
		update[i] = x
	}

	// Move to first node on level 0
	x = x.Forward[0]

	// Make sure we found the right node
	if x != nil && x.Score == score && x.Member == member {
		// Remove node at all levels
		for i := 0; i < sl.level; i++ {
			if update[i].Forward[i] != x {
				break
			}
			update[i].Forward[i] = x.Forward[i]
		}

		// Update tail if necessary
		if x == sl.tail {
			sl.tail = update[0]
		}

		// Update level if necessary
		for sl.level > 1 && sl.header.Forward[sl.level-1] == nil {
			sl.level--
		}

		sl.length--
		return true
	}

	return false
}

// CountInRange counts elements with score between min and max
func (sl *SkipList) CountInRange(min, max float64) int {
	count := 0
	x := sl.header

	// Find first node with score >= min
	for i := sl.level - 1; i >= 0; i-- {
		for x.Forward[i] != nil && x.Forward[i].Score < min {
			x = x.Forward[i]
		}
	}

	// Traverse nodes with score <= max
	x = x.Forward[0]
	for x != nil && x.Score <= max {
		count++
		x = x.Forward[0]
	}

	return count
}

// RangeByScore returns members with scores between min and max
func (sl *SkipList) RangeByScore(min, max float64, offset, count int) []string {
	result := []string{}
	x := sl.header

	// Find first node with score >= min
	for i := sl.level - 1; i >= 0; i-- {
		for x.Forward[i] != nil && x.Forward[i].Score < min {
			x = x.Forward[i]
		}
	}

	// Traverse nodes with score <= max
	x = x.Forward[0]
	skipped := 0

	for x != nil && x.Score <= max {
		if offset < 0 || skipped >= offset {
			result = append(result, x.Member)
			// Stop if we've collected enough elements
			if count > 0 && len(result) >= count {
				break
			}
		} else {
			skipped++
		}
		x = x.Forward[0]
	}

	return result
}

// RangeByRank returns members by rank (position)
func (sl *SkipList) RangeByRank(start, stop int) []string {
	result := []string{}

	// Handle negative indices and out of range
	if start < 0 {
		start = sl.length + start
	}
	if stop < 0 {
		stop = sl.length + stop
	}
	if start < 0 {
		start = 0
	}
	if stop >= sl.length {
		stop = sl.length - 1
	}
	if start > stop || start >= sl.length {
		return result
	}

	// Traverse to start position
	x := sl.header.Forward[0]
	for i := 0; i < start && x != nil; i++ {
		x = x.Forward[0]
	}

	// Collect members between start and stop
	for i := start; i <= stop && x != nil; i++ {
		result = append(result, x.Member)
		x = x.Forward[0]
	}

	return result
}

// GetRank returns the rank of a member
func (sl *SkipList) GetRank(member string, score float64) int {
	rank := 0
	x := sl.header

	for i := sl.level - 1; i >= 0; i-- {
		for x.Forward[i] != nil &&
			(x.Forward[i].Score < score ||
				(x.Forward[i].Score == score && x.Forward[i].Member < member)) {
			rank += 1 // Count nodes we're skipping
			x = x.Forward[i]
		}
	}

	x = x.Forward[0]
	if x != nil && x.Member == member {
		return rank
	}

	return -1 // Member not found
}
