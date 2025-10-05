package kv

import (
	"sync"
	"time"
)

// Entry represents a cache entry with its value and optional expiration time.
//
// Fields:
//   - Value: The stored byte slice data
//   - ExpiresAt: When this entry expires. Zero time means no expiration.
type Entry struct {
	// Value is the stored data as byte slice
	Value []byte

	// ExpiresAt is the expiration timestamp
	// Zero value (time.Time{}) means the entry never expires
	ExpiresAt time.Time
}

// IsExpired checks if the entry has expired based on current time.
//
// Returns:
//   - bool: True if the entry has expired, false if still valid or has no expiration
//
// An entry is considered expired if:
//  1. It has an expiration time set (ExpiresAt is not zero), AND
//  2. The current time is after ExpiresAt
//
// Entries with zero ExpiresAt never expire and always return false.
func (e *Entry) IsExpired() bool {
	return !e.ExpiresAt.IsZero() && time.Now().After(e.ExpiresAt)
}

// Cache is a thread-safe in-memory key-value store with TTL support.
//
// Cache provides concurrent access to stored key-value pairs with automatic
// expiration handling. All methods are safe for concurrent use.
//
// The cache includes basic metrics tracking (hits, misses, sets) for
// monitoring and diagnostics.
//
// A background goroutine automatically cleans up expired entries every minute
// to prevent memory leaks.
type Cache struct {
	// mu protects concurrent access to all fields
	mu sync.RWMutex

	// store holds the actual cache data
	// Key: cache key (string), Value: Entry pointer
	store map[string]*Entry

	// Metrics for cache performance tracking
	hits   int64 // Number of successful Get operations
	misses int64 // Number of failed Get operations (key not found or expired)
	sets   int64 // Number of Set operations
}

// NewCache creates a new cache instance and starts the background cleanup goroutine.
//
// Returns:
//   - *Cache: A new cache ready for use
//
// Side effects:
//   - Starts a background goroutine that runs cleanup every minute
//   - The cleanup goroutine continues until the program exits
//
// Example:
//
//	cache := kv.NewCache()
//	cache.Set("key1", []byte("value1"), 5*time.Minute)
//	value, ok := cache.Get("key1")
func NewCache() *Cache {
	c := &Cache{
		store: make(map[string]*Entry),
	}

	// Start cleanup goroutine
	go c.cleanupExpired()

	return c
}

// Get retrieves a value from the cache by key.
//
// Parameters:
//   - key: The cache key to look up
//
// Returns:
//   - []byte: The cached value if found and not expired, nil otherwise
//   - bool: True if the key was found and not expired, false otherwise
//
// Behavior:
//   - Increments hits counter if key found and valid
//   - Increments misses counter if key not found or expired
//   - Removes expired entries on access
//
// Thread-safety: Safe for concurrent calls
//
// Example:
//
//	if value, ok := cache.Get("user:123"); ok {
//	    fmt.Printf("User data: %s\n", value)
//	} else {
//	    fmt.Println("User not found in cache")
//	}
func (c *Cache) Get(key string) ([]byte, bool) {
	c.mu.RLock()
	entry, exists := c.store[key]
	c.mu.RUnlock()

	if !exists {
		c.mu.Lock()
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	if entry.IsExpired() {
		c.mu.Lock()
		delete(c.store, key)
		c.misses++
		c.mu.Unlock()
		return nil, false
	}

	c.mu.Lock()
	c.hits++
	c.mu.Unlock()

	return entry.Value, true
}

// Set stores a key-value pair with optional TTL (time-to-live).
//
// Parameters:
//   - key: The cache key
//   - value: The data to store (byte slice)
//   - ttl: Time-to-live duration. Use 0 for no expiration.
//
// Behavior:
//   - If ttl > 0: Entry expires after the specified duration
//   - If ttl = 0: Entry never expires
//   - Overwrites existing entry if key already exists
//   - Increments sets counter
//
// Thread-safety: Safe for concurrent calls
//
// Example:
//
//	// Set with 5 minute expiration
//	cache.Set("session:abc", []byte("user123"), 5*time.Minute)
//
//	// Set with no expiration
//	cache.Set("config:version", []byte("1.0"), 0)
func (c *Cache) Set(key string, value []byte, ttl time.Duration) {
	c.mu.Lock()
	defer c.mu.Unlock()

	entry := &Entry{
		Value: value,
	}

	if ttl > 0 {
		entry.ExpiresAt = time.Now().Add(ttl)
	}

	c.store[key] = entry
	c.sets++
}

// Delete removes a key from the cache.
//
// Parameters:
//   - key: The cache key to remove
//
// Returns:
//   - bool: True if the key existed and was deleted, false if key was not found
//
// Thread-safety: Safe for concurrent calls
//
// Example:
//
//	if cache.Delete("user:123") {
//	    fmt.Println("User removed from cache")
//	} else {
//	    fmt.Println("User was not in cache")
//	}
func (c *Cache) Delete(key string) bool {
	c.mu.Lock()
	defer c.mu.Unlock()

	_, exists := c.store[key]
	if exists {
		delete(c.store, key)
	}
	return exists
}

// Size returns the current number of entries in the cache.
//
// Returns:
//   - int: Number of entries (includes expired but not yet cleaned entries)
//
// Note: This count includes expired entries that haven't been cleaned yet.
// The actual number of valid entries may be lower.
//
// Thread-safety: Safe for concurrent calls
//
// Example:
//
//	fmt.Printf("Cache contains %d entries\n", cache.Size())
func (c *Cache) Size() int {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return len(c.store)
}

// Stats returns cache performance statistics.
//
// Returns:
//   - hits: Number of successful Get operations
//   - misses: Number of failed Get operations (key not found or expired)
//   - sets: Number of Set operations
//
// Thread-safety: Safe for concurrent calls
//
// Example:
//
//	hits, misses, sets := cache.Stats()
//	total := hits + misses
//	if total > 0 {
//	    hitRate := float64(hits) / float64(total) * 100
//	    fmt.Printf("Hit rate: %.2f%% (hits: %d, misses: %d, sets: %d)\n",
//	        hitRate, hits, misses, sets)
//	}
func (c *Cache) Stats() (hits, misses, sets int64) {
	c.mu.RLock()
	defer c.mu.RUnlock()
	return c.hits, c.misses, c.sets
}

// Clear removes all entries from the cache and resets it to empty state.
//
// Side effects:
//   - Removes all cached data
//   - Does NOT reset statistics (hits, misses, sets)
//   - Memory is released for garbage collection
//
// Thread-safety: Safe for concurrent calls
//
// Example:
//
//	// Clear all cache data on configuration reload
//	cache.Clear()
func (c *Cache) Clear() {
	c.mu.Lock()
	defer c.mu.Unlock()
	c.store = make(map[string]*Entry)
}

// cleanupExpired periodically removes expired entries from the cache.
//
// This method runs in a background goroutine started by NewCache.
// It wakes up every minute to scan for and remove expired entries.
//
// Side effects:
//   - Acquires write lock during cleanup (may briefly block other operations)
//   - Removes expired entries to prevent memory leaks
//   - Runs indefinitely until program termination
//
// This is an internal method and should not be called directly by users.
func (c *Cache) cleanupExpired() {
	ticker := time.NewTicker(1 * time.Minute)
	defer ticker.Stop()

	for range ticker.C {
		c.mu.Lock()
		for key, entry := range c.store {
			if entry.IsExpired() {
				delete(c.store, key)
			}
		}
		c.mu.Unlock()
	}
}

// GetTTL returns the remaining time-to-live for a cache key in seconds.
//
// Parameters:
//   - key: The cache key to check
//
// Returns:
//   - int32: Remaining TTL in seconds. Returns 0 if:
//   - Key doesn't exist
//   - Entry has no expiration (TTL was 0)
//   - Entry has already expired
//
// Thread-safety: Safe for concurrent calls
//
// Example:
//
//	ttl := cache.GetTTL("session:abc")
//	if ttl > 0 {
//	    fmt.Printf("Session expires in %d seconds\n", ttl)
//	} else {
//	    fmt.Println("Session not found or expired")
//	}
func (c *Cache) GetTTL(key string) int32 {
	c.mu.RLock()
	defer c.mu.RUnlock()

	entry, exists := c.store[key]
	if !exists {
		return 0
	}

	if entry.ExpiresAt.IsZero() {
		return 0 // No expiration
	}

	remaining := time.Until(entry.ExpiresAt)
	if remaining <= 0 {
		return 0
	}

	return int32(remaining.Seconds())
}
