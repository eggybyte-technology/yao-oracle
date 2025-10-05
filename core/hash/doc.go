// Package hash implements consistent hashing with virtual nodes for
// distributed cache node selection in the Yao-Oracle system.
//
// The hash ring distributes keys across multiple nodes using the CRC32
// hash function. Virtual nodes (replicas) are added to improve distribution
// uniformity and reduce hotspots when nodes are added or removed.
//
// # Basic Usage
//
// Create a ring and add nodes:
//
//	ring := hash.NewRing(150) // 150 virtual nodes per physical node
//	ring.AddNode("cache-node-1:8080")
//	ring.AddNode("cache-node-2:8080")
//	ring.AddNode("cache-node-3:8080")
//
// Find the node responsible for a key:
//
//	node := ring.GetNode("user:12345")
//	fmt.Printf("Key 'user:12345' maps to node: %s\n", node)
//
// # Thread Safety
//
// All Ring methods are safe for concurrent use. Reads can proceed in
// parallel, while writes (AddNode, RemoveNode) acquire an exclusive lock.
//
// # Virtual Nodes
//
// Virtual nodes improve distribution uniformity. The default is 150 virtual
// nodes per physical node, which provides good balance between memory usage
// and distribution quality. Higher values (e.g., 500) provide better
// distribution but increase memory consumption.
//
// # Hash Function
//
// The ring uses CRC32 (IEEE polynomial) as the hash function. This provides
// fast hashing with good distribution properties. The hash values are uint32,
// so the ring has 2^32 possible positions.
package hash
