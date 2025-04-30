package zset

import (
	"fmt"
	"redigo/datastruct/skiplist"
	"sort"
	"strconv"
)

// ZSet encoding types
const (
	encodingListpack = iota
	encodingSkiplist
)

const listpackMaxSize = 128

// ZSet is the interface that represents a Redis sorted set
type ZSet interface {
	Add(member string, score float64) bool
	Remove(member string) bool
	Score(member string) (float64, bool)
	Exists(member string) bool
	Count(min, max float64) int
	Len() int
	RangeByScore(min, max float64, offset, count int) []string
	RangeByRank(start, stop int) []string
	RemoveRangeByRank(start, stop int) int
	RemoveRangeByScore(min, max float64) int
	Encoding() int
	GetSkiplist() *skiplist.SkipList
}

type zset struct {
	encoding int
	listpack [][2]string
	dict     map[string]float64
	skiplist *skiplist.SkipList
}

// New creates a new zset
func NewZSet() ZSet {
	return &zset{
		encoding: encodingListpack,
		listpack: make([][2]string, 0),
	}
}

// Add adds a member with the given score to the sorted set
// Returns true if the element was added as a new member, false if the score was updated
func (z *zset) Add(member string, score float64) bool {
	// Check if we're using listpack encoding
	if z.encoding == encodingListpack {
		// Check if member already exists in listpack
		for i, pair := range z.listpack {
			if pair[0] == member {
				// Update score if member already exists
				z.listpack[i][1] = formatScore(score)
				return false
			}
		}

		// Add new member to listpack
		z.listpack = append(z.listpack, [2]string{member, formatScore(score)})

		// Convert to skiplist encoding if listpack grows too large
		if len(z.listpack) > listpackMaxSize {
			z.convertToSkiplist()
		}
		return true
	}

	// Using skiplist encoding
	existingScore, exists := z.dict[member]
	if exists {
		// If score changed, update both dict and skiplist
		if existingScore != score {
			// Remove from skiplist with old score
			z.skiplist.Delete(member, existingScore)
			// Insert with new score
			z.skiplist.Insert(member, score)
			// Update score in dict
			z.dict[member] = score
		}
		return false
	}

	// Add new member to both dict and skiplist
	z.dict[member] = score
	z.skiplist.Insert(member, score)
	return true
}

// Helper function to format score as string
func formatScore(score float64) string {
	return fmt.Sprintf("%f", score)
}

// Helper function to parse score string to float64
func parseScore(scoreStr string) (float64, error) {
	return strconv.ParseFloat(scoreStr, 64)
}

// Convert from listpack to skiplist encoding
func (z *zset) convertToSkiplist() {
	if z.encoding == encodingSkiplist {
		return
	}

	// Initialize skiplist and dict
	z.skiplist = skiplist.NewSkipList()
	z.dict = make(map[string]float64, len(z.listpack))

	// Transfer all elements from listpack to skiplist and dict
	for _, pair := range z.listpack {
		member := pair[0]
		score, _ := parseScore(pair[1])
		z.dict[member] = score
		z.skiplist.Insert(member, score)
	}

	// Update encoding and clear listpack
	z.encoding = encodingSkiplist
	z.listpack = nil
}

// Score returns the score of a member, and a boolean indicating if the member exists
func (z *zset) Score(member string) (float64, bool) {
	if z.encoding == encodingListpack {
		for _, pair := range z.listpack {
			if pair[0] == member {
				score, err := parseScore(pair[1])
				if err != nil {
					return 0, false
				}
				return score, true
			}
		}
		return 0, false
	}

	// Using skiplist encoding
	score, exists := z.dict[member]
	return score, exists
}

// Exists checks if a member exists in the sorted set
func (z *zset) Exists(member string) bool {
	if z.encoding == encodingListpack {
		for _, pair := range z.listpack {
			if pair[0] == member {
				return true
			}
		}
		return false
	}

	// Using skiplist encoding
	_, exists := z.dict[member]
	return exists
}

// Count returns the number of elements in the specified score range
func (z *zset) Count(min, max float64) int {
	if z.encoding == encodingListpack {
		count := 0
		for _, pair := range z.listpack {
			score, _ := parseScore(pair[1])
			if score >= min && score <= max {
				count++
			}
		}
		return count
	}

	// Using skiplist encoding
	return z.skiplist.CountInRange(min, max)
}

// Len returns the number of elements in the sorted set
func (z *zset) Len() int {
	if z.encoding == encodingListpack {
		return len(z.listpack)
	}
	return len(z.dict)
}

// RangeByScore returns members with scores between min and max
// Limit: if offset >=0 and count > 0, return at most count members starting from offset
func (z *zset) RangeByScore(min, max float64, offset, count int) []string {
	if z.encoding == encodingListpack {
		// Get matching elements from listpack
		var matches [][2]string
		for _, pair := range z.listpack {
			score, _ := parseScore(pair[1])
			if score >= min && score <= max {
				matches = append(matches, pair)
			}
		}

		// Sort matches by score
		sort.Slice(matches, func(i, j int) bool {
			scoreI, _ := parseScore(matches[i][1])
			scoreJ, _ := parseScore(matches[j][1])
			return scoreI < scoreJ
		})

		// Apply offset and count if specified
		if offset >= 0 && count > 0 {
			end := offset + count
			if end > len(matches) {
				end = len(matches)
			}
			if offset < len(matches) {
				matches = matches[offset:end]
			} else {
				matches = nil
			}
		}

		// Extract member names
		result := make([]string, len(matches))
		for i, pair := range matches {
			result[i] = pair[0]
		}
		return result
	}

	// Using skiplist encoding
	return z.skiplist.RangeByScore(min, max, offset, count)
}

// RangeByRank returns members ordered by rank (position)
// Returns members between start and stop ranks (inclusive, 0-based)
func (z *zset) RangeByRank(start, stop int) []string {
	if z.encoding == encodingListpack {
		// Copy and sort listpack elements by score
		pairs := make([][2]string, len(z.listpack))
		copy(pairs, z.listpack)

		sort.Slice(pairs, func(i, j int) bool {
			scoreI, _ := parseScore(pairs[i][1])
			scoreJ, _ := parseScore(pairs[j][1])
			return scoreI < scoreJ
		})

		// Handle negative indices and out of range
		size := len(pairs)
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
		if start > stop || start >= size {
			return []string{}
		}

		// Extract member names
		result := make([]string, 0, stop-start+1)
		for i := start; i <= stop; i++ {
			result = append(result, pairs[i][0])
		}
		return result
	}

	// Using skiplist encoding
	return z.skiplist.RangeByRank(start, stop)
}

// Remove removes a member from the sorted set
// Returns true if the member was removed, false if it didn't exist
func (z *zset) Remove(member string) bool {
	if z.encoding == encodingListpack {
		for i, pair := range z.listpack {
			if pair[0] == member {
				// Remove the member by slicing it out
				z.listpack = append(z.listpack[:i], z.listpack[i+1:]...)
				return true
			}
		}
		return false
	}

	// Using skiplist encoding
	score, exists := z.dict[member]
	if exists {
		z.skiplist.Delete(member, score)
		delete(z.dict, member)
		return true
	}
	return false
}

// RemoveRangeByRank removes all members between start and stop ranks (inclusive, 0-based)
// Returns number of members removed
func (z *zset) RemoveRangeByRank(start, stop int) int {
	members := z.RangeByRank(start, stop)
	count := 0
	for _, member := range members {
		if z.Remove(member) {
			count++
		}
	}
	return count
}

// RemoveRangeByScore removes all members with scores between min and max
// Returns number of members removed
func (z *zset) RemoveRangeByScore(min, max float64) int {
	if z.encoding == encodingListpack {
		// Find members to remove
		toRemove := make([]string, 0)
		for _, pair := range z.listpack {
			score, _ := parseScore(pair[1])
			if score >= min && score <= max {
				toRemove = append(toRemove, pair[0])
			}
		}

		// Remove the identified members
		count := 0
		for _, member := range toRemove {
			if z.Remove(member) {
				count++
			}
		}
		return count
	}

	// Using skiplist encoding
	members := z.skiplist.RangeByScore(min, max, 0, -1)
	count := 0
	for _, member := range members {
		if z.Remove(member) {
			count++
		}
	}
	return count
}

// Encoding returns the current encoding type of the zset (0 for listpack, 1 for skiplist)
func (z *zset) Encoding() int {
	return z.encoding
}

// GetSkiplist returns the skiplist used by the zset when in skiplist encoding
// Returns nil if the zset is using listpack encoding
func (z *zset) GetSkiplist() *skiplist.SkipList {
	if z.encoding == encodingSkiplist {
		return z.skiplist
	}
	return nil
}
