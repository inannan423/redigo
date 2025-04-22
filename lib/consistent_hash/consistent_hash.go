package consistenthash

import (
	"hash/crc32"
	"sort"
)

type NodeMap struct {
	hashFunc    func(data []byte) uint32
	nodeHashs   []int
	nodehashMap map[int]string
}

// NewNodeMap creates a new NodeMap instance
func NewNodeMap(hashFunc func(data []byte) uint32) *NodeMap {
	m := &NodeMap{
		hashFunc:    hashFunc,
		nodehashMap: make(map[int]string),
	}
	if m.hashFunc == nil {
		m.hashFunc = crc32.ChecksumIEEE
	}
	return m
}

// IsEmpty checks if the NodeMap is empty
func (m *NodeMap) IsEmpty() bool {
	return len(m.nodehashMap) == 0
}

// AddNodes adds nodes to the NodeMap
func (m *NodeMap) AddNodes(nodes ...string) {
	for _, node := range nodes {
		if node == "" {
			continue
		}
		hash := int(m.hashFunc([]byte(node)))
		m.nodeHashs = append(m.nodeHashs, hash)
		m.nodehashMap[hash] = node
	}
	sort.Ints(m.nodeHashs)
}

// PickNode picks a node based on the key, returning the node that is closest to the hash of the key
func (m *NodeMap) PickNode(key string) string {
	if m.IsEmpty() {
		return ""
	}

	hash := int(m.hashFunc([]byte(key)))
	index := sort.Search(len(m.nodeHashs), func(i int) bool {
		return m.nodeHashs[i] >= hash
	})
	// If the hash is greater than all node hashes, Int.Search returns len(m.nodeHashs)
	// So we need to wrap around to the first node
	if index == len(m.nodeHashs) {
		index = 0
	}
	return m.nodehashMap[m.nodeHashs[index]]
}
