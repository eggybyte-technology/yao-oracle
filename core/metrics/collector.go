package metrics

import (
	"sync"
	"sync/atomic"
	"time"
)

// Metrics holds statistics for a service.
type Metrics struct {
	StartTime time.Time

	// Request counters
	requestsTotal atomic.Int64
	requestsOK    atomic.Int64
	requestsErr   atomic.Int64

	// Cache metrics
	cacheHits   atomic.Int64
	cacheMisses atomic.Int64

	// Per-namespace metrics
	mu               sync.RWMutex
	namespaceMetrics map[string]*NamespaceMetrics
}

// NamespaceMetrics holds per-namespace statistics.
type NamespaceMetrics struct {
	Requests atomic.Int64
	Hits     atomic.Int64
	Misses   atomic.Int64
	Errors   atomic.Int64
}

// NewMetrics creates a new metrics collector.
func NewMetrics() *Metrics {
	return &Metrics{
		StartTime:        time.Now(),
		namespaceMetrics: make(map[string]*NamespaceMetrics),
	}
}

// IncRequests increments the total request counter.
func (m *Metrics) IncRequests() {
	m.requestsTotal.Add(1)
}

// IncRequestsOK increments the successful request counter.
func (m *Metrics) IncRequestsOK() {
	m.requestsOK.Add(1)
}

// IncRequestsErr increments the error request counter.
func (m *Metrics) IncRequestsErr() {
	m.requestsErr.Add(1)
}

// IncRequestsError is an alias for IncRequestsErr for backward compatibility.
func (m *Metrics) IncRequestsError() {
	m.IncRequestsErr()
}

// IncCacheHits increments the cache hit counter.
func (m *Metrics) IncCacheHits() {
	m.cacheHits.Add(1)
}

// IncCacheMisses increments the cache miss counter.
func (m *Metrics) IncCacheMisses() {
	m.cacheMisses.Add(1)
}

// GetRequestsTotal returns the total number of requests.
func (m *Metrics) GetRequestsTotal() int64 {
	return m.requestsTotal.Load()
}

// GetRequestsOK returns the number of successful requests.
func (m *Metrics) GetRequestsOK() int64 {
	return m.requestsOK.Load()
}

// GetRequestsErr returns the number of failed requests.
func (m *Metrics) GetRequestsErr() int64 {
	return m.requestsErr.Load()
}

// GetCacheHits returns the number of cache hits.
func (m *Metrics) GetCacheHits() int64 {
	return m.cacheHits.Load()
}

// GetCacheMisses returns the number of cache misses.
func (m *Metrics) GetCacheMisses() int64 {
	return m.cacheMisses.Load()
}

// GetHitRate returns the cache hit rate (0.0 to 1.0).
func (m *Metrics) GetHitRate() float64 {
	hits := m.GetCacheHits()
	misses := m.GetCacheMisses()
	total := hits + misses

	if total == 0 {
		return 0.0
	}

	return float64(hits) / float64(total)
}

// GetUptime returns the service uptime.
func (m *Metrics) GetUptime() time.Duration {
	return time.Since(m.StartTime)
}

// RecordNamespaceRequest records a request for a specific namespace.
func (m *Metrics) RecordNamespaceRequest(namespace string, hit bool, err error) {
	m.mu.Lock()
	nsMetrics, exists := m.namespaceMetrics[namespace]
	if !exists {
		nsMetrics = &NamespaceMetrics{}
		m.namespaceMetrics[namespace] = nsMetrics
	}
	m.mu.Unlock()

	nsMetrics.Requests.Add(1)

	if err != nil {
		nsMetrics.Errors.Add(1)
	} else if hit {
		nsMetrics.Hits.Add(1)
	} else {
		nsMetrics.Misses.Add(1)
	}
}

// GetNamespaceMetrics returns metrics for a specific namespace.
func (m *Metrics) GetNamespaceMetrics(namespace string) *NamespaceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()
	return m.namespaceMetrics[namespace]
}

// GetAllNamespaceMetrics returns metrics for all namespaces.
func (m *Metrics) GetAllNamespaceMetrics() map[string]*NamespaceMetrics {
	m.mu.RLock()
	defer m.mu.RUnlock()

	result := make(map[string]*NamespaceMetrics)
	for ns, metrics := range m.namespaceMetrics {
		result[ns] = metrics
	}
	return result
}
