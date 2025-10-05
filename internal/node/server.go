package node

import (
	"context"
	"fmt"
	"net"
	"runtime"
	"time"

	"google.golang.org/grpc"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	oraclev1 "github.com/eggybyte-technology/yao-oracle/pb/yao/oracle/v1"

	"github.com/eggybyte-technology/yao-oracle/core/health"
	"github.com/eggybyte-technology/yao-oracle/core/kv"
	"github.com/eggybyte-technology/yao-oracle/core/metrics"
	"github.com/eggybyte-technology/yao-oracle/core/utils"
)

// Server implements the NodeService gRPC server.
type Server struct {
	oraclev1.UnimplementedNodeServiceServer

	cache         *kv.Cache
	metrics       *metrics.Metrics
	healthChecker *health.Checker
	logger        *utils.Logger
	startTime     time.Time
}

// NewServer creates a new node server instance.
func NewServer() *Server {
	return &Server{
		cache:         kv.NewCache(),
		metrics:       metrics.NewMetrics(),
		healthChecker: health.NewChecker(),
		logger:        utils.NewLogger("node"),
		startTime:     time.Now(),
	}
}

// Get retrieves a value by key from the cache.
func (s *Server) Get(ctx context.Context, req *oraclev1.GetRequest) (*oraclev1.GetResponse, error) {
	s.metrics.IncRequests()

	value, found := s.cache.Get(req.Key)
	if !found {
		s.metrics.IncCacheMisses()
		return &oraclev1.GetResponse{
			Found: false,
		}, nil
	}

	s.metrics.IncCacheHits()
	s.metrics.IncRequestsOK()

	return &oraclev1.GetResponse{
		Found: true,
		Value: value,
		Ttl:   s.cache.GetTTL(req.Key),
	}, nil
}

// Set stores a key-value pair with optional TTL.
func (s *Server) Set(ctx context.Context, req *oraclev1.SetRequest) (*oraclev1.SetResponse, error) {
	s.metrics.IncRequests()

	ttl := time.Duration(req.Ttl) * time.Second
	s.cache.Set(req.Key, req.Value, ttl)

	s.metrics.IncRequestsOK()

	return &oraclev1.SetResponse{
		Success: true,
	}, nil
}

// Delete removes a key from the cache.
func (s *Server) Delete(ctx context.Context, req *oraclev1.DeleteRequest) (*oraclev1.DeleteResponse, error) {
	s.metrics.IncRequests()

	existed := s.cache.Delete(req.Key)

	s.metrics.IncRequestsOK()

	return &oraclev1.DeleteResponse{
		Success: true,
		Existed: existed,
	}, nil
}

// Health checks if the node is healthy and ready to serve.
func (s *Server) Health(ctx context.Context, req *oraclev1.HealthRequest) (*oraclev1.HealthResponse, error) {
	return &oraclev1.HealthResponse{
		Healthy: true,
		Message: "Node is healthy",
	}, nil
}

// Stats returns node statistics.
func (s *Server) Stats(ctx context.Context, req *oraclev1.StatsRequest) (*oraclev1.StatsResponse, error) {
	hits, misses, _ := s.cache.Stats()

	var m runtime.MemStats
	runtime.ReadMemStats(&m)

	return &oraclev1.StatsResponse{
		TotalKeys:       int64(s.cache.Size()),
		MemoryUsedBytes: int64(m.Alloc),
		UptimeSeconds:   int64(time.Since(s.startTime).Seconds()),
		RequestsTotal:   s.metrics.GetRequestsTotal(),
		Hits:            hits,
		Misses:          misses,
	}, nil
}

// Run starts the node server on the specified port.
func (s *Server) Run(port int) error {
	lis, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	oraclev1.RegisterNodeServiceServer(grpcServer, s)

	// Register gRPC health check service
	grpcHealthServer := grpchealth.NewServer()
	grpcHealthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)
	grpc_health_v1.RegisterHealthServer(grpcServer, grpcHealthServer)

	// Mark service as healthy and ready
	s.healthChecker.SetHealthy(true)
	s.healthChecker.SetReady(true)

	s.logger.Info("Node server listening on :%d", port)

	if err := grpcServer.Serve(lis); err != nil {
		return fmt.Errorf("failed to serve: %w", err)
	}

	return nil
}

// StartHealthServer starts the HTTP health check server on the specified port.
//
// This should be called in a goroutine to run concurrently with the main gRPC server.
//
// Parameters:
//   - port: HTTP port for health checks (typically 9091)
//
// Returns:
//   - error: Error if health server fails to start
//
// Example:
//
//	go server.StartHealthServer(9091)
func (s *Server) StartHealthServer(port int) error {
	return s.healthChecker.Start(port)
}

// Stop gracefully shuts down the node server.
func (s *Server) Stop() error {
	// Mark as unhealthy to stop receiving traffic
	s.healthChecker.SetReady(false)
	s.healthChecker.SetHealthy(false)

	// Stop health checker
	return s.healthChecker.Stop()
}
