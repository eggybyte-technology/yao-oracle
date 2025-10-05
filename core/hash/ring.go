package hash

import (
	"fmt"
	"hash/crc32"
	"sort"
	"sync"
)

// Ring represents a consistent hash ring that distributes keys across
// multiple nodes using virtual nodes for better balance.
//
// Ring is safe for concurrent use. Read operations (GetNode, Nodes, Size)
// can proceed in parallel, while write operations (AddNode, RemoveNode)
// acquire an exclusive lock.
//
// Virtual nodes (replicas) improve distribution uniformity. Higher replica
// counts provide better load balancing but increase memory usage.
type Ring struct {
	// mu protects concurrent access to all fields
	mu sync.RWMutex

	// nodes contains the identifiers of all physical nodes in the ring
	nodes []string

	// virtualNodes is the number of virtual nodes per physical node
	// Higher values provide better distribution but increase memory
	virtualNodes int

	// ring contains all virtual node hashes in sorted order
	// for efficient binary search during key lookup
	ring []uint32

	// hashMap maps each virtual node hash to its physical node identifier
	hashMap map[uint32]string
}

// NewRing creates a new consistent hash ring with the specified number of
// virtual nodes per physical node.
//
// Parameters:
//   - virtualNodes: Number of virtual nodes to create for each physical node.
//     If <= 0, defaults to 150. Typical values range from 100 to 500.
//     Higher values provide better key distribution but increase memory usage.
//
// Returns:
//   - *Ring: A new empty ring ready to accept nodes via AddNode
//
// Example:
//
//	// Create a ring with default 150 virtual nodes per physical node
//	ring := hash.NewRing(150)
//
//	// Create a ring with higher replication for better distribution
//	highRepRing := hash.NewRing(500)
func NewRing(virtualNodes int) *Ring {
	if virtualNodes <= 0 {
		virtualNodes = 150
	}

	return &Ring{
		virtualNodes: virtualNodes,
		hashMap:      make(map[uint32]string),
	}
}

// AddNode registers a new physical node in the hash ring by creating
// multiple virtual nodes (replicas) for better key distribution.
//
// Parameters:
//   - node: Unique identifier for the physical node (e.g., "cache-0:8080",
//     "10.0.1.5:8080"). Must not be empty. Duplicate node identifiers
//     are silently ignored (idempotent operation).
//
// Side effects:
//   - Acquires write lock on the ring
//   - Creates r.virtualNodes virtual nodes in the hash space
//   - Re-sorts the hash ring (O(n log n) operation where n is total virtual nodes)
//   - Keys may be remapped to the new node (consistent hashing property)
//
// Thread-safety: Safe for concurrent calls
//
// Example:
//
//	ring := hash.NewRing(150)
//	ring.AddNode("cache-node-1:8080")
//	ring.AddNode("cache-node-2:8080")
//	ring.AddNode("cache-node-1:8080") // Ignored - already exists
func (r *Ring) AddNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Check if node already exists
	for _, n := range r.nodes {
		if n == node {
			return
		}
	}

	r.nodes = append(r.nodes, node)

	// Add virtual nodes
	for i := 0; i < r.virtualNodes; i++ {
		virtualKey := fmt.Sprintf("%s#%d", node, i)
		hash := r.hashKey(virtualKey)
		r.ring = append(r.ring, hash)
		r.hashMap[hash] = node
	}

	// Sort ring
	sort.Slice(r.ring, func(i, j int) bool {
		return r.ring[i] < r.ring[j]
	})
}

// RemoveNode unregisters a physical node from the hash ring by removing
// all of its virtual nodes.
//
// Parameters:
//   - node: Identifier of the physical node to remove. If the node doesn't
//     exist, this is a no-op (idempotent operation).
//
// Side effects:
//   - Acquires write lock on the ring
//   - Removes all virtual nodes associated with this physical node
//   - Keys previously mapped to this node will be remapped to other nodes
//   - Does NOT rebuild the ring (only removes entries)
//
// Thread-safety: Safe for concurrent calls
//
// Example:
//
//	ring.RemoveNode("cache-node-1:8080")
//	// Keys previously on node-1 are now distributed to remaining nodes
func (r *Ring) RemoveNode(node string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Remove from nodes list
	for i, n := range r.nodes {
		if n == node {
			r.nodes = append(r.nodes[:i], r.nodes[i+1:]...)
			break
		}
	}

	// Remove virtual nodes
	newRing := make([]uint32, 0)
	for _, hash := range r.ring {
		if r.hashMap[hash] != node {
			newRing = append(newRing, hash)
		} else {
			delete(r.hashMap, hash)
		}
	}
	r.ring = newRing
}

// GetNode returns the physical node responsible for storing the given key
// using consistent hashing algorithm.
//
// Parameters:
//   - key: The cache key to look up (e.g., "user:12345", "session:abc")
//
// Returns:
//   - string: The physical node identifier that should handle this key.
//     Returns empty string if no nodes are available in the ring.
//
// The algorithm:
//  1. Hashes the key using CRC32
//  2. Uses binary search to find the first virtual node with hash >= key hash
//  3. Wraps around to the first node if we're past the end (ring property)
//  4. Returns the physical node associated with that virtual node
//
// Thread-safety: Safe for concurrent calls (read lock only)
//
// Example:
//
//	node := ring.GetNode("user:12345")
//	if node == "" {
//	    return errors.New("no cache nodes available")
//	}
//	// Forward request to this node
func (r *Ring) GetNode(key string) string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	if len(r.ring) == 0 {
		return ""
	}

	hash := r.hashKey(key)

	// Binary search for the first node with hash >= key hash
	idx := sort.Search(len(r.ring), func(i int) bool {
		return r.ring[i] >= hash
	})

	// Wrap around if we're past the end
	if idx == len(r.ring) {
		idx = 0
	}

	return r.hashMap[r.ring[idx]]
}

// Nodes returns a copy of all physical nodes currently in the ring.
//
// Returns:
//   - []string: Slice of physical node identifiers. Returns a copy to prevent
//     external modification. Order is not guaranteed.
//
// Thread-safety: Safe for concurrent calls (read lock only)
//
// Example:
//
//	nodes := ring.Nodes()
//	fmt.Printf("Ring has %d nodes: %v\n", len(nodes), nodes)
func (r *Ring) Nodes() []string {
	r.mu.RLock()
	defer r.mu.RUnlock()

	nodes := make([]string, len(r.nodes))
	copy(nodes, r.nodes)
	return nodes
}

// Size returns the number of physical nodes currently in the ring.
//
// Returns:
//   - int: Number of physical nodes (not virtual nodes)
//
// Thread-safety: Safe for concurrent calls (read lock only)
//
// Example:
//
//	if ring.Size() == 0 {
//	    log.Println("Warning: No cache nodes available")
//	}
func (r *Ring) Size() int {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return len(r.nodes)
}

// hashKey computes the hash for a given key.
func (r *Ring) hashKey(key string) uint32 {
	return crc32.ChecksumIEEE([]byte(key))
}
