// Package kv provides a thread-safe in-memory key-value cache with TTL support
// for the Yao-Oracle distributed cache system.
//
// This package implements the storage layer for cache nodes, providing:
//   - Thread-safe concurrent access (read-write locks)
//   - TTL (time-to-live) expiration for cache entries
//   - Automatic background cleanup of expired entries
//   - Basic metrics (hits, misses, sets)
//
// # Basic Usage
//
// Create a cache and perform operations:
//
//	cache := kv.NewCache()
//
//	// Set a value with 5 minute TTL
//	cache.Set("user:123", []byte("John Doe"), 5*time.Minute)
//
//	// Get a value
//	if value, ok := cache.Get("user:123"); ok {
//	    fmt.Printf("Value: %s\n", value)
//	}
//
//	// Delete a value
//	cache.Delete("user:123")
//
//	// Get statistics
//	hits, misses, sets := cache.Stats()
//	fmt.Printf("Hit rate: %.2f%%\n", float64(hits)/float64(hits+misses)*100)
//
// # TTL Behavior
//
// TTL (time-to-live) controls how long entries remain valid:
//   - TTL > 0: Entry expires after the specified duration
//   - TTL = 0: Entry never expires (stored indefinitely)
//   - Expired entries are removed on access and during periodic cleanup
//
// # Thread Safety
//
// All Cache methods are safe for concurrent use:
//   - Multiple goroutines can read simultaneously (RLock)
//   - Write operations acquire exclusive lock (Lock)
//   - Background cleanup runs in a separate goroutine
//
// # Automatic Cleanup
//
// A background goroutine runs every minute to remove expired entries.
// This prevents memory leaks from expired but unaccessed entries.
package kv
