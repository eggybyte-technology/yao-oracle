package dashboard

import (
	"context"
	"fmt"
	"sync"
	"time"

	oraclev1 "github.com/eggybyte-technology/yao-oracle/pb/yao/oracle/v1"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/eggybyte-technology/yao-oracle/core/utils"
)

// DashboardGRPCServer implements the gRPC DashboardService with streaming support.
//
// This server provides:
//   - Real-time metrics streaming via StreamMetrics RPC
//   - Cache query capabilities via QueryCache RPC
//   - Secret management via ManageSecret RPC
//   - Configuration retrieval via GetConfig RPC
//
// Thread-safety: All methods are safe for concurrent use.
type DashboardGRPCServer struct {
	oraclev1.UnimplementedDashboardServiceServer
	mu              sync.RWMutex
	informer        ConfigInformer
	mockGenerator   *MockDataGenerator
	logger          *utils.Logger
	refreshInterval time.Duration
	testMode        bool
}

// NewDashboardGRPCServer creates a new gRPC dashboard server instance.
//
// Parameters:
//   - informer: Configuration informer for dynamic config reloading
//   - refreshInterval: Metrics refresh interval in seconds
//   - testMode: Whether to use mock data generator
//
// Returns:
//   - *DashboardGRPCServer: A new gRPC server instance
func NewDashboardGRPCServer(informer ConfigInformer, refreshInterval int, testMode bool) *DashboardGRPCServer {
	s := &DashboardGRPCServer{
		informer:        informer,
		logger:          utils.NewLogger("dashboard-grpc"),
		refreshInterval: time.Duration(refreshInterval) * time.Second,
		testMode:        testMode,
	}

	if testMode {
		s.mockGenerator = NewMockDataGenerator(refreshInterval)
		s.logger.Info("Test mode: Using mock data generator")
	}

	return s
}

// StreamMetrics implements the StreamMetrics RPC method.
// It streams cluster metrics to the client at regular intervals.
func (s *DashboardGRPCServer) StreamMetrics(req *oraclev1.SubscribeRequest, stream oraclev1.DashboardService_StreamMetricsServer) error {
	namespaceFilter := req.Namespace
	if namespaceFilter == "" {
		namespaceFilter = "all"
	}
	s.logger.Info("📊 Client subscribed to metrics stream (namespace: %s)", namespaceFilter)

	ticker := time.NewTicker(s.refreshInterval)
	defer ticker.Stop()

	// Send initial metrics immediately
	metrics, err := s.collectClusterMetrics(req.Namespace)
	if err == nil {
		if err := stream.Send(metrics); err != nil {
			s.logger.Error("❌ Failed to send initial metrics: %v", err)
			return status.Errorf(codes.Internal, "failed to send metrics: %v", err)
		}
		s.logger.Info("✅ Sent initial metrics snapshot (QPS: %.1f, Hit Rate: %.1f%%, Nodes: %d)",
			metrics.Global.Qps, metrics.Global.HitRate*100, len(metrics.Nodes))
	}

	for {
		select {
		case <-stream.Context().Done():
			s.logger.Info("👋 Client disconnected from metrics stream")
			return nil
		case <-ticker.C:
			metrics, err := s.collectClusterMetrics(req.Namespace)
			if err != nil {
				s.logger.Error("❌ Failed to collect metrics: %v", err)
				return status.Errorf(codes.Internal, "failed to collect metrics: %v", err)
			}

			if err := stream.Send(metrics); err != nil {
				s.logger.Error("❌ Failed to send metrics: %v", err)
				return status.Errorf(codes.Internal, "failed to send metrics: %v", err)
			}

			// Log periodic updates
			s.logger.Info("🔄 Metrics update sent (QPS: %.1f, Hit Rate: %.1f%%, Memory: %.1fMB, Nodes: %d/%d healthy)",
				metrics.Global.Qps,
				metrics.Global.HitRate*100,
				metrics.Global.MemoryUsedMb,
				metrics.Global.HealthyNodes,
				metrics.Global.TotalNodes)
		}
	}
}

// collectClusterMetrics gathers cluster metrics from all sources.
func (s *DashboardGRPCServer) collectClusterMetrics(namespaceFilter string) (*oraclev1.ClusterMetrics, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	if s.testMode && s.mockGenerator != nil {
		return s.collectMockMetrics(namespaceFilter)
	}

	// In production, this would collect real metrics from proxy and nodes
	// For now, return empty metrics
	return &oraclev1.ClusterMetrics{
		Timestamp: time.Now().Unix(),
		Global: &oraclev1.GlobalStats{
			Qps:          0,
			LatencyMs:    0,
			HitRate:      0,
			MemoryUsedMb: 0,
			HealthScore:  1.0,
			TotalKeys:    0,
			TotalProxies: 0,
			TotalNodes:   0,
			HealthyNodes: 0,
		},
		Namespaces: []*oraclev1.NamespaceStats{},
		Nodes:      []*oraclev1.NodeStats{},
	}, nil
}

// collectMockMetrics generates mock cluster metrics for testing.
func (s *DashboardGRPCServer) collectMockMetrics(namespaceFilter string) (*oraclev1.ClusterMetrics, error) {
	nodes := s.mockGenerator.GetNodes()
	namespaces := s.mockGenerator.GetNamespaces()
	_ = s.mockGenerator.GetProxyHealth() // Not used for now, but kept for future reference

	// Calculate global stats
	totalKeys := int64(0)
	totalMemoryMB := 0.0
	totalHits := int64(0)
	totalRequests := int64(0)
	healthyNodes := int32(0)

	for _, node := range nodes {
		totalKeys += node.TotalKeys
		totalMemoryMB += float64(node.MemoryUsedBytes) / (1024 * 1024)
		totalHits += node.Hits
		totalRequests += node.RequestsTotal
		if node.Healthy {
			healthyNodes++
		}
	}

	hitRate := 0.0
	if totalRequests > 0 {
		hitRate = float64(totalHits) / float64(totalRequests)
	}

	// Calculate QPS based on uptime
	qps := 0.0
	if len(nodes) > 0 && nodes[0].UptimeSeconds > 0 {
		qps = float64(totalRequests) / float64(nodes[0].UptimeSeconds)
	}

	// Calculate average latency from nodes
	avgLatency := 0.0
	if len(nodes) > 0 {
		for _, node := range nodes {
			avgLatency += node.Latency50thMs
		}
		avgLatency /= float64(len(nodes))
	}

	// Calculate health score (weighted: 50% hit rate, 30% latency, 20% memory)
	healthScore := hitRate*0.5 + (1.0-min(avgLatency/10.0, 1.0))*0.3 + (1.0-min(totalMemoryMB/1024.0, 1.0))*0.2

	globalStats := &oraclev1.GlobalStats{
		Qps:          qps,
		LatencyMs:    avgLatency,
		HitRate:      hitRate,
		MemoryUsedMb: totalMemoryMB,
		HealthScore:  healthScore,
		TotalKeys:    totalKeys,
		TotalProxies: 1,
		TotalNodes:   int32(len(nodes)),
		HealthyNodes: healthyNodes,
	}

	// Build namespace stats
	namespaceStats := make([]*oraclev1.NamespaceStats, 0, len(namespaces))
	for _, ns := range namespaces {
		// Filter by namespace if specified
		if namespaceFilter != "" && ns.Name != namespaceFilter {
			continue
		}

		nsStats := &oraclev1.NamespaceStats{
			Name:         ns.Name,
			Qps:          qps * 0.3, // Approximate distribution
			HitRate:      ns.HitRate,
			TtlAvg:       float64(ns.DefaultTTL),
			Keys:         ns.KeyCount,
			MemoryUsedMb: float64(ns.KeyCount) * 8 / 1024, // Approximate
			ApiKey:       maskAPIKey("api-key-" + ns.Name),
			Description:  ns.Description,
			MaxMemoryMb:  int32(ns.MaxMemoryMB),
			DefaultTtl:   int32(ns.DefaultTTL),
			RateLimitQps: int32(ns.RateLimitQPS),
		}
		namespaceStats = append(namespaceStats, nsStats)
	}

	// Build node stats
	nodeStats := make([]*oraclev1.NodeStats, 0, len(nodes))
	for _, node := range nodes {
		ns := &oraclev1.NodeStats{
			Id:            node.Address,
			Ip:            node.Address,
			Namespace:     "",
			MemoryUsedMb:  float64(node.MemoryUsedBytes) / (1024 * 1024),
			HitRate:       float64(node.Hits) / max(float64(node.RequestsTotal), 1.0),
			LatencyMs:     node.Latency50thMs,
			KeyCount:      node.TotalKeys,
			Healthy:       node.Healthy,
			UptimeSeconds: node.UptimeSeconds,
			Qps:           float64(node.RequestsTotal) / max(float64(node.UptimeSeconds), 1.0),
		}
		nodeStats = append(nodeStats, ns)
	}

	return &oraclev1.ClusterMetrics{
		Timestamp:  time.Now().Unix(),
		Global:     globalStats,
		Namespaces: namespaceStats,
		Nodes:      nodeStats,
	}, nil
}

// QueryCache implements the QueryCache RPC method.
func (s *DashboardGRPCServer) QueryCache(ctx context.Context, req *oraclev1.CacheQueryRequest) (*oraclev1.CacheQueryResponse, error) {
	s.logger.Info("Cache query: namespace=%s, key=%s", req.Namespace, req.Key)

	if req.Namespace == "" || req.Key == "" {
		return nil, status.Errorf(codes.InvalidArgument, "namespace and key are required")
	}

	// In test mode, return mock data
	if s.testMode {
		return &oraclev1.CacheQueryResponse{
			Key:        req.Key,
			Value:      fmt.Sprintf(`{"mock":"data for %s"}`, req.Key),
			TtlSeconds: 60,
			SizeBytes:  int64(len(req.Key) + 20),
			CreatedAt:  time.Now().Add(-5 * time.Minute).Format(time.RFC3339),
			LastAccess: time.Now().Format(time.RFC3339),
			Found:      true,
		}, nil
	}

	// In production, query the actual cache through proxy
	return &oraclev1.CacheQueryResponse{
		Found: false,
	}, nil
}

// ManageSecret implements the ManageSecret RPC method.
func (s *DashboardGRPCServer) ManageSecret(ctx context.Context, req *oraclev1.SecretUpdateRequest) (*oraclev1.SecretUpdateResponse, error) {
	s.logger.Info("Secret update request: namespace=%s", req.Namespace)

	if req.Namespace == "" || req.NewApiKey == "" {
		return nil, status.Errorf(codes.InvalidArgument, "namespace and new_api_key are required")
	}

	// In test mode, always succeed
	if s.testMode {
		return &oraclev1.SecretUpdateResponse{
			Success:   true,
			UpdatedAt: time.Now().Format(time.RFC3339),
			Message:   "API key updated successfully (test mode)",
		}, nil
	}

	// In production, update the Kubernetes Secret
	// This would trigger the Informer to reload configuration
	return &oraclev1.SecretUpdateResponse{
		Success:   false,
		UpdatedAt: time.Now().Format(time.RFC3339),
		Message:   "Not implemented in production mode",
	}, nil
}

// GetConfig implements the GetConfig RPC method.
func (s *DashboardGRPCServer) GetConfig(ctx context.Context, req *oraclev1.ConfigRequest) (*oraclev1.ConfigResponse, error) {
	s.logger.Info("Config request received")

	cfg := s.informer.GetConfig()

	configs := make([]*oraclev1.NamespaceConfig, 0)
	if cfg.Proxy != nil {
		for _, ns := range cfg.Proxy.Namespaces {
			configs = append(configs, &oraclev1.NamespaceConfig{
				Namespace:    ns.Name,
				DefaultTtl:   int64(ns.DefaultTTL),
				MaxKeys:      1000000, // Default value
				MaxMemoryMb:  int32(ns.MaxMemoryMB),
				RateLimitQps: int32(ns.RateLimitQPS),
				Description:  ns.Description,
			})
		}
	}

	return &oraclev1.ConfigResponse{
		Configs: configs,
	}, nil
}

// Stop gracefully shuts down the gRPC server.
func (s *DashboardGRPCServer) Stop() {
	if s.testMode && s.mockGenerator != nil {
		s.mockGenerator.Stop()
	}
	s.logger.Info("Dashboard gRPC server stopped")
}

// maskAPIKey masks an API key for security (shows only first and last 4 characters).
func maskAPIKey(key string) string {
	if len(key) <= 8 {
		return "****"
	}
	return key[:4] + "****" + key[len(key)-4:]
}

// max returns the maximum of two float64 values.
func max(a, b float64) float64 {
	if a > b {
		return a
	}
	return b
}

// RegisterDashboardServer registers the dashboard service with a gRPC server.
func RegisterDashboardServer(grpcServer *grpc.Server, dashboardServer *DashboardGRPCServer) {
	oraclev1.RegisterDashboardServiceServer(grpcServer, dashboardServer)
}
