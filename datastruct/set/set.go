package set

// Set represents a Redis set
type Set interface {
	Add(member string) int                     // Add a member to the set, return 1 if added, 0 if already exists
	Remove(member string) int                  // Remove a member from the set, return 1 if removed, 0 if not exists
	Contains(member string) bool               // Check if the set contains a member
	Members() []string                         // Get all members of the set
	Len() int                                  // Get the number of members in the set
	ForEach(consumer func(member string) bool) // Iterate over all members
	RandomMembers(count int) []string          // Get random members from the set
	RandomDistinctMembers(count int) []string  // Get distinct random members
	IsIntSet() bool                            // Check if the set is an IntSet
}
