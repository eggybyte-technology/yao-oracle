package dashboard

import (
	"context"
	"fmt"
	"math/rand"
	"sync"
	"time"

	oraclev1 "github.com/eggybyte-technology/yao-oracle/pb/yao/oracle/v1"
	"google.golang.org/grpc"
)

// MockDataGenerator generates simulated metrics for testing the dashboard
// without requiring a real Kubernetes cluster or backend services.
//
// It simulates:
//   - Multiple namespaces with different configurations
//   - Multiple cache nodes with varying statistics
//   - A proxy service with health information
//   - Dynamic metrics that change over time
//
// Thread-safety: All methods are safe for concurrent use.
type MockDataGenerator struct {
	mu              sync.RWMutex
	namespaces      []*MockNamespace
	nodes           []*MockNode
	proxyHealth     *MockProxyHealth
	startTime       time.Time
	refreshInterval time.Duration
	stopCh          chan struct{}
}

// MockNamespace represents a simulated business namespace.
type MockNamespace struct {
	Name         string
	Description  string
	MaxMemoryMB  int32
	DefaultTTL   int32
	RateLimitQPS int32
	KeyCount     int64
	HitRate      float64
}

// MockNode represents a simulated cache node.
type MockNode struct {
	Address         string
	Healthy         bool
	TotalKeys       int64
	MemoryUsedBytes int64
	MemoryMaxBytes  int64
	UptimeSeconds   int64
	RequestsTotal   int64
	Hits            int64
	Misses          int64
	Latency99thMs   float64
	Latency95thMs   float64
	Latency50thMs   float64
}

// MockProxyHealth represents simulated proxy health status.
type MockProxyHealth struct {
	Healthy         bool
	NamespacesCount int32
	NodesHealthy    int32
	NodesTotal      int32
	Message         string
	RequestsPerSec  float64
}

// NewMockDataGenerator creates a new mock data generator with default test data.
//
// The generator creates:
//   - 3 namespaces (game-app, ads-service, analytics)
//   - 3 cache nodes with simulated metrics
//   - 1 proxy with health information
//
// Parameters:
//   - refreshInterval: How often to update dynamic metrics (in seconds)
//
// Returns:
//   - *MockDataGenerator: A new mock data generator ready to use
func NewMockDataGenerator(refreshInterval int) *MockDataGenerator {
	g := &MockDataGenerator{
		namespaces:      createMockNamespaces(),
		nodes:           createMockNodes(),
		proxyHealth:     createMockProxyHealth(),
		startTime:       time.Now(),
		refreshInterval: time.Duration(refreshInterval) * time.Second,
		stopCh:          make(chan struct{}),
	}

	// Start background goroutine to update metrics
	go g.updateMetricsPeriodically()

	return g
}

// createMockNamespaces creates a set of test namespaces with different configurations.
func createMockNamespaces() []*MockNamespace {
	return []*MockNamespace{
		{
			Name:         "game-app",
			Description:  "Gaming application cache",
			MaxMemoryMB:  512,
			DefaultTTL:   60,
			RateLimitQPS: 100,
			KeyCount:     15000,
			HitRate:      0.92,
		},
		{
			Name:         "ads-service",
			Description:  "Advertisement service cache",
			MaxMemoryMB:  256,
			DefaultTTL:   120,
			RateLimitQPS: 50,
			KeyCount:     8000,
			HitRate:      0.85,
		},
		{
			Name:         "analytics",
			Description:  "Analytics data cache",
			MaxMemoryMB:  1024,
			DefaultTTL:   300,
			RateLimitQPS: 200,
			KeyCount:     30000,
			HitRate:      0.78,
		},
	}
}

// createMockNodes creates a set of test cache nodes with varying characteristics.
func createMockNodes() []*MockNode {
	return []*MockNode{
		{
			Address:         "cache-node-0:7070",
			Healthy:         true,
			TotalKeys:       18000,
			MemoryUsedBytes: 150 * 1024 * 1024, // 150 MB
			MemoryMaxBytes:  512 * 1024 * 1024, // 512 MB
			UptimeSeconds:   3600,
			RequestsTotal:   50000,
			Hits:            45000,
			Misses:          5000,
			Latency99thMs:   5.2,
			Latency95thMs:   3.8,
			Latency50thMs:   1.5,
		},
		{
			Address:         "cache-node-1:7070",
			Healthy:         true,
			TotalKeys:       17500,
			MemoryUsedBytes: 140 * 1024 * 1024, // 140 MB
			MemoryMaxBytes:  512 * 1024 * 1024, // 512 MB
			UptimeSeconds:   3550,
			RequestsTotal:   48000,
			Hits:            43000,
			Misses:          5000,
			Latency99thMs:   4.8,
			Latency95thMs:   3.5,
			Latency50thMs:   1.4,
		},
		{
			Address:         "cache-node-2:7070",
			Healthy:         true,
			TotalKeys:       17800,
			MemoryUsedBytes: 145 * 1024 * 1024, // 145 MB
			MemoryMaxBytes:  512 * 1024 * 1024, // 512 MB
			UptimeSeconds:   3580,
			RequestsTotal:   49000,
			Hits:            44000,
			Misses:          5000,
			Latency99thMs:   5.0,
			Latency95thMs:   3.6,
			Latency50thMs:   1.5,
		},
	}
}

// createMockProxyHealth creates simulated proxy health information.
func createMockProxyHealth() *MockProxyHealth {
	return &MockProxyHealth{
		Healthy:         true,
		NamespacesCount: 3,
		NodesHealthy:    3,
		NodesTotal:      3,
		Message:         "All systems operational",
		RequestsPerSec:  150.5,
	}
}

// updateMetricsPeriodically runs in the background and simulates metric changes.
func (g *MockDataGenerator) updateMetricsPeriodically() {
	ticker := time.NewTicker(g.refreshInterval)
	defer ticker.Stop()

	for {
		select {
		case <-ticker.C:
			g.updateMetrics()
		case <-g.stopCh:
			return
		}
	}
}

// updateMetrics simulates metric changes (adds random variations to simulate real behavior).
func (g *MockDataGenerator) updateMetrics() {
	g.mu.Lock()
	defer g.mu.Unlock()

	uptimeSeconds := time.Since(g.startTime).Seconds()

	// Update nodes with small random variations
	for i, node := range g.nodes {
		// Simulate request growth (varies by node - simulate uneven load distribution)
		baseIncrement := 50 + (i * 20) // Different base load per node
		increment := int64(rand.Intn(100) + baseIncrement)
		node.RequestsTotal += increment

		// Simulate hits (85-95% hit rate with slight variations per node)
		baseHitRate := 0.85 + float64(i)*0.02 // Node 0: 85%, Node 1: 87%, Node 2: 89%
		hitRate := baseHitRate + rand.Float64()*0.05
		hits := int64(float64(increment) * hitRate)
		node.Hits += hits
		node.Misses += increment - hits

		// Simulate key count changes (gradual growth with occasional drops)
		change := int64(rand.Intn(150) - 25) // Bias toward growth
		node.TotalKeys += change
		if node.TotalKeys < 10000 {
			node.TotalKeys = 10000
		}
		// Cap at reasonable maximum
		if node.TotalKeys > 50000 {
			node.TotalKeys = 50000 - int64(rand.Intn(5000)) // Occasional eviction
		}

		// Simulate memory changes (proportional to key count with some overhead)
		avgKeySize := 8*1024 + int64(rand.Intn(2*1024)) // 8-10KB per key
		node.MemoryUsedBytes = node.TotalKeys * avgKeySize

		// Simulate uptime
		node.UptimeSeconds = int64(uptimeSeconds)

		// Simulate latency variations (more realistic patterns)
		// P99 typically 3-8ms, P95 2-5ms, P50 1-2ms
		baseLatency := 1.0 + float64(i)*0.2 // Different baseline per node
		node.Latency99thMs = baseLatency*5.0 + rand.Float64()*2.0
		node.Latency95thMs = baseLatency*3.0 + rand.Float64()*1.5
		node.Latency50thMs = baseLatency + rand.Float64()*0.5

		// Simulate occasional node health issues (5% chance)
		if rand.Float64() < 0.05 {
			node.Healthy = false
			g.proxyHealth.NodesHealthy--
		} else {
			node.Healthy = true
			g.proxyHealth.NodesHealthy = g.proxyHealth.NodesTotal
		}
	}

	// Update namespace stats (aggregate from nodes with realistic distribution)
	totalKeys := int64(0)
	totalHits := int64(0)
	totalRequests := int64(0)

	for _, node := range g.nodes {
		totalKeys += node.TotalKeys
		totalHits += node.Hits
		totalRequests += node.RequestsTotal
	}

	// Distribute keys across namespaces (weighted with some variation)
	g.namespaces[0].KeyCount = totalKeys*40/100 + int64(rand.Intn(1000)-500) // game-app: ~40%
	g.namespaces[1].KeyCount = totalKeys*25/100 + int64(rand.Intn(500)-250)  // ads-service: ~25%
	g.namespaces[2].KeyCount = totalKeys*35/100 + int64(rand.Intn(800)-400)  // analytics: ~35%

	// Update hit rates with realistic per-namespace variations
	if totalRequests > 0 {
		overallHitRate := float64(totalHits) / float64(totalRequests)
		// Game app typically has higher hit rate (cached user sessions)
		g.namespaces[0].HitRate = min(0.98, overallHitRate+0.05+rand.Float64()*0.03)
		// Ads service has moderate hit rate
		g.namespaces[1].HitRate = min(0.95, overallHitRate+rand.Float64()*0.05-0.025)
		// Analytics has lower hit rate (more unique queries)
		g.namespaces[2].HitRate = min(0.92, overallHitRate-0.05+rand.Float64()*0.03)
	}

	// Update proxy health with realistic QPS calculation
	if uptimeSeconds > 0 {
		// Calculate instantaneous QPS (requests in last interval)
		intervalRequests := float64(0)
		for _, node := range g.nodes {
			intervalRequests += float64(node.RequestsTotal)
		}
		g.proxyHealth.RequestsPerSec = intervalRequests / uptimeSeconds

		// Add some variation to simulate traffic patterns
		g.proxyHealth.RequestsPerSec *= (0.9 + rand.Float64()*0.2)
	}

	// Update proxy message based on health
	if g.proxyHealth.NodesHealthy < g.proxyHealth.NodesTotal {
		g.proxyHealth.Healthy = true // Degraded but still operational
		g.proxyHealth.Message = fmt.Sprintf("Degraded: %d/%d nodes healthy", g.proxyHealth.NodesHealthy, g.proxyHealth.NodesTotal)
	} else {
		g.proxyHealth.Healthy = true
		g.proxyHealth.Message = "All systems operational"
	}
}

// min returns the smaller of two float64 values.
func min(a, b float64) float64 {
	if a < b {
		return a
	}
	return b
}

// GetNamespaces returns the list of mock namespaces.
func (g *MockDataGenerator) GetNamespaces() []*MockNamespace {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]*MockNamespace, len(g.namespaces))
	copy(result, g.namespaces)
	return result
}

// GetNodes returns the list of mock nodes.
func (g *MockDataGenerator) GetNodes() []*MockNode {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Return a copy to prevent external modification
	result := make([]*MockNode, len(g.nodes))
	copy(result, g.nodes)
	return result
}

// GetProxyHealth returns mock proxy health status.
func (g *MockDataGenerator) GetProxyHealth() *MockProxyHealth {
	g.mu.RLock()
	defer g.mu.RUnlock()

	// Return a copy
	result := *g.proxyHealth
	return &result
}

// Stop stops the background metric update goroutine.
func (g *MockDataGenerator) Stop() {
	close(g.stopCh)
}

// MockProxyClient implements a mock gRPC proxy client for testing.
type MockProxyClient struct {
	generator *MockDataGenerator
}

// NewMockProxyClient creates a new mock proxy client.
func NewMockProxyClient(generator *MockDataGenerator) oraclev1.ProxyServiceClient {
	return &MockProxyClient{generator: generator}
}

// Health implements the mock Health RPC call.
func (m *MockProxyClient) Health(ctx context.Context, in *oraclev1.ProxyHealthRequest, opts ...grpc.CallOption) (*oraclev1.ProxyHealthResponse, error) {
	health := m.generator.GetProxyHealth()
	return &oraclev1.ProxyHealthResponse{
		Healthy:         health.Healthy,
		NamespacesCount: health.NamespacesCount,
		NodesHealthy:    health.NodesHealthy,
		NodesTotal:      health.NodesTotal,
		Message:         health.Message,
	}, nil
}

// Implement other required methods as no-ops (dashboard doesn't use these)

// Get implements the mock Get RPC call.
func (m *MockProxyClient) Get(ctx context.Context, in *oraclev1.ProxyGetRequest, opts ...grpc.CallOption) (*oraclev1.ProxyGetResponse, error) {
	return &oraclev1.ProxyGetResponse{}, fmt.Errorf("not implemented in mock")
}

// Set implements the mock Set RPC call.
func (m *MockProxyClient) Set(ctx context.Context, in *oraclev1.ProxySetRequest, opts ...grpc.CallOption) (*oraclev1.ProxySetResponse, error) {
	return &oraclev1.ProxySetResponse{}, fmt.Errorf("not implemented in mock")
}

// Delete implements the mock Delete RPC call.
func (m *MockProxyClient) Delete(ctx context.Context, in *oraclev1.ProxyDeleteRequest, opts ...grpc.CallOption) (*oraclev1.ProxyDeleteResponse, error) {
	return &oraclev1.ProxyDeleteResponse{}, fmt.Errorf("not implemented in mock")
}

// BatchGet implements the mock BatchGet RPC call.
func (m *MockProxyClient) BatchGet(ctx context.Context, in *oraclev1.ProxyBatchGetRequest, opts ...grpc.CallOption) (*oraclev1.ProxyBatchGetResponse, error) {
	return &oraclev1.ProxyBatchGetResponse{}, fmt.Errorf("not implemented in mock")
}

// MockNodeClient implements a mock gRPC node client for testing.
type MockNodeClient struct {
	nodeData *MockNode
}

// NewMockNodeClient creates a new mock node client for a specific node.
func NewMockNodeClient(nodeData *MockNode) oraclev1.NodeServiceClient {
	return &MockNodeClient{nodeData: nodeData}
}

// Health implements the mock Health RPC call.
func (m *MockNodeClient) Health(ctx context.Context, in *oraclev1.HealthRequest, opts ...grpc.CallOption) (*oraclev1.HealthResponse, error) {
	return &oraclev1.HealthResponse{
		Healthy: m.nodeData.Healthy,
		Message: "Mock node healthy",
	}, nil
}

// Stats implements the mock Stats RPC call.
func (m *MockNodeClient) Stats(ctx context.Context, in *oraclev1.StatsRequest, opts ...grpc.CallOption) (*oraclev1.StatsResponse, error) {
	return &oraclev1.StatsResponse{
		TotalKeys:       m.nodeData.TotalKeys,
		MemoryUsedBytes: m.nodeData.MemoryUsedBytes,
		UptimeSeconds:   m.nodeData.UptimeSeconds,
		RequestsTotal:   m.nodeData.RequestsTotal,
		Hits:            m.nodeData.Hits,
		Misses:          m.nodeData.Misses,
	}, nil
}

// Get implements the mock Get RPC call (not used in dashboard).
func (m *MockNodeClient) Get(ctx context.Context, in *oraclev1.GetRequest, opts ...grpc.CallOption) (*oraclev1.GetResponse, error) {
	return &oraclev1.GetResponse{}, fmt.Errorf("not implemented in mock")
}

// Set implements the mock Set RPC call (not used in dashboard).
func (m *MockNodeClient) Set(ctx context.Context, in *oraclev1.SetRequest, opts ...grpc.CallOption) (*oraclev1.SetResponse, error) {
	return &oraclev1.SetResponse{}, fmt.Errorf("not implemented in mock")
}

// Delete implements the mock Delete RPC call (not used in dashboard).
func (m *MockNodeClient) Delete(ctx context.Context, in *oraclev1.DeleteRequest, opts ...grpc.CallOption) (*oraclev1.DeleteResponse, error) {
	return &oraclev1.DeleteResponse{}, fmt.Errorf("not implemented in mock")
}
