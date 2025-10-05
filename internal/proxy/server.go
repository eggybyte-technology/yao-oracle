package proxy

import (
	"context"
	"fmt"
	"net"
	"sync"

	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
	grpchealth "google.golang.org/grpc/health"
	"google.golang.org/grpc/health/grpc_health_v1"

	oraclev1 "github.com/eggybyte-technology/yao-oracle/pb/yao/oracle/v1"

	"github.com/eggybyte-technology/yao-oracle/core/config"
	"github.com/eggybyte-technology/yao-oracle/core/hash"
	"github.com/eggybyte-technology/yao-oracle/core/health"
	"github.com/eggybyte-technology/yao-oracle/core/metrics"
	"github.com/eggybyte-technology/yao-oracle/core/utils"
)

// Server implements the ProxyService gRPC server.
//
// The proxy server acts as the brain of the cluster, handling:
//   - Business namespace isolation via API key authentication
//   - Request routing using consistent hashing
//   - Dynamic configuration reloading from Kubernetes Secret
//   - Health checking and metrics collection
//
// Thread-safety: All methods are safe for concurrent use.
type Server struct {
	oraclev1.UnimplementedProxyServiceServer

	mu            sync.RWMutex
	informer      *config.K8sInformer
	ring          *hash.Ring
	nodeClients   map[string]oraclev1.NodeServiceClient
	metrics       *metrics.Metrics
	healthChecker *health.Checker
	logger        *utils.Logger
	stopCh        chan struct{}
}

// NewServer creates a new proxy server instance with Kubernetes Informer.
//
// The informer provides dynamic configuration reloading without restart.
// Configuration changes (namespace updates, API key rotations) are applied
// automatically when the Kubernetes Secret is updated.
//
// Parameters:
//   - informer: Kubernetes Informer for configuration management
//
// Returns:
//   - *Server: A new proxy server instance ready to start
//
// Example:
//
//	informer, _ := config.NewK8sInformer(...)
//	server := proxy.NewServer(informer)
//	server.Run(8080)
func NewServer(informer *config.K8sInformer) *Server {
	s := &Server{
		informer:      informer,
		ring:          hash.NewRing(150),
		nodeClients:   make(map[string]oraclev1.NodeServiceClient),
		metrics:       metrics.NewMetrics(),
		healthChecker: health.NewChecker(),
		logger:        utils.NewLogger("proxy"),
		stopCh:        make(chan struct{}),
	}

	return s
}

// SetNodes configures the cache nodes for routing.
//
// This method is typically used for:
//   - Static node configuration (testing/development)
//   - Manual node registration
//   - Initial cluster setup
//
// In production, nodes are usually discovered via Kubernetes service discovery.
//
// Parameters:
//   - nodes: List of cache node addresses (e.g., ["node-0:7070", "node-1:7070"])
//
// Side effects:
//   - Clears existing hash ring
//   - Establishes gRPC connections to all nodes
//   - Logs connection errors (but continues for successful nodes)
func (s *Server) SetNodes(nodes []string) {
	s.mu.Lock()
	defer s.mu.Unlock()

	// Clear existing ring
	s.ring = hash.NewRing(150)

	// Add new nodes
	for _, node := range nodes {
		s.ring.AddNode(node)
		s.logger.Info("Added cache node: %s", node)

		// Create gRPC client for this node
		if _, exists := s.nodeClients[node]; !exists {
			conn, err := grpc.Dial(node, grpc.WithTransportCredentials(insecure.NewCredentials()))
			if err != nil {
				s.logger.Error("Failed to connect to node %s: %v", node, err)
				continue
			}
			s.nodeClients[node] = oraclev1.NewNodeServiceClient(conn)
		}
	}

	s.logger.Info("Cache node ring updated: %d nodes", s.ring.Size())
}

// Get retrieves a value by key (with API key authentication).
//
// Request flow:
// 1. Validate API key and determine namespace
// 2. Add namespace prefix to key
// 3. Use consistent hashing to select target node
// 4. Forward request to selected node
// 5. Return result to client
func (s *Server) Get(ctx context.Context, req *oraclev1.ProxyGetRequest) (*oraclev1.ProxyGetResponse, error) {
	s.metrics.IncRequests()

	// Authenticate and get namespace
	ns, ok := s.authenticateRequest(req.ApiKey)
	if !ok {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("invalid API key")
	}

	// Add namespace prefix to key
	namespacedKey := s.namespaceKey(ns.Name, req.Key)

	// Route to appropriate node
	targetNode := s.selectNode(namespacedKey)
	if targetNode == "" {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("no cache node available")
	}

	// Get client for target node
	s.mu.RLock()
	client, exists := s.nodeClients[targetNode]
	s.mu.RUnlock()

	if !exists {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("node client not found: %s", targetNode)
	}

	// Forward request to node
	nodeResp, err := client.Get(ctx, &oraclev1.GetRequest{
		Key: namespacedKey,
	})
	if err != nil {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("node error: %w", err)
	}

	if nodeResp.Found {
		s.metrics.IncCacheHits()
	} else {
		s.metrics.IncCacheMisses()
	}
	s.metrics.IncRequestsOK()

	return &oraclev1.ProxyGetResponse{
		Found: nodeResp.Found,
		Value: nodeResp.Value,
		Ttl:   nodeResp.Ttl,
		Node:  targetNode,
	}, nil
}

// Set stores a key-value pair (with API key authentication).
func (s *Server) Set(ctx context.Context, req *oraclev1.ProxySetRequest) (*oraclev1.ProxySetResponse, error) {
	s.metrics.IncRequests()

	// Authenticate and get namespace
	ns, ok := s.authenticateRequest(req.ApiKey)
	if !ok {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("invalid API key")
	}

	// Add namespace prefix to key
	namespacedKey := s.namespaceKey(ns.Name, req.Key)

	// Route to appropriate node
	targetNode := s.selectNode(namespacedKey)
	if targetNode == "" {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("no cache node available")
	}

	// Get client for target node
	s.mu.RLock()
	client, exists := s.nodeClients[targetNode]
	s.mu.RUnlock()

	if !exists {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("node client not found: %s", targetNode)
	}

	// Forward request to node
	nodeResp, err := client.Set(ctx, &oraclev1.SetRequest{
		Key:   namespacedKey,
		Value: req.Value,
		Ttl:   req.Ttl,
	})
	if err != nil {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("node error: %w", err)
	}

	s.metrics.IncRequestsOK()

	return &oraclev1.ProxySetResponse{
		Success: nodeResp.Success,
		Message: nodeResp.Message,
		Node:    targetNode,
	}, nil
}

// Delete removes a key (with API key authentication).
func (s *Server) Delete(ctx context.Context, req *oraclev1.ProxyDeleteRequest) (*oraclev1.ProxyDeleteResponse, error) {
	s.metrics.IncRequests()

	// Authenticate and get namespace
	ns, ok := s.authenticateRequest(req.ApiKey)
	if !ok {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("invalid API key")
	}

	// Add namespace prefix to key
	namespacedKey := s.namespaceKey(ns.Name, req.Key)

	// Route to appropriate node
	targetNode := s.selectNode(namespacedKey)
	if targetNode == "" {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("no cache node available")
	}

	// Get client for target node
	s.mu.RLock()
	client, exists := s.nodeClients[targetNode]
	s.mu.RUnlock()

	if !exists {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("node client not found: %s", targetNode)
	}

	// Forward request to node
	nodeResp, err := client.Delete(ctx, &oraclev1.DeleteRequest{
		Key: namespacedKey,
	})
	if err != nil {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("node error: %w", err)
	}

	s.metrics.IncRequestsOK()

	return &oraclev1.ProxyDeleteResponse{
		Success: nodeResp.Success,
		Existed: nodeResp.Existed,
		Node:    targetNode,
	}, nil
}

// BatchGet retrieves multiple keys in a single request.
func (s *Server) BatchGet(ctx context.Context, req *oraclev1.ProxyBatchGetRequest) (*oraclev1.ProxyBatchGetResponse, error) {
	s.metrics.IncRequests()

	// Authenticate and get namespace
	ns, ok := s.authenticateRequest(req.ApiKey)
	if !ok {
		s.metrics.IncRequestsError()
		return nil, fmt.Errorf("invalid API key")
	}

	results := make(map[string][]byte)
	nodesUsed := make(map[string]bool)

	// Process each key
	for _, key := range req.Keys {
		// Add namespace prefix
		namespacedKey := s.namespaceKey(ns.Name, key)

		// Route to appropriate node
		targetNode := s.selectNode(namespacedKey)
		if targetNode == "" {
			continue
		}

		nodesUsed[targetNode] = true

		// Get client for target node
		s.mu.RLock()
		client, exists := s.nodeClients[targetNode]
		s.mu.RUnlock()

		if !exists {
			continue
		}

		// Forward request to node
		nodeResp, err := client.Get(ctx, &oraclev1.GetRequest{
			Key: namespacedKey,
		})
		if err != nil || !nodeResp.Found {
			continue
		}

		// Store result (using original key, not namespaced)
		results[key] = nodeResp.Value
	}

	// Convert nodes used map to slice
	nodesList := make([]string, 0, len(nodesUsed))
	for node := range nodesUsed {
		nodesList = append(nodesList, node)
	}

	s.metrics.IncRequestsOK()

	return &oraclev1.ProxyBatchGetResponse{
		Results:   results,
		NodesUsed: nodesList,
	}, nil
}

// Health checks proxy health and cluster status.
func (s *Server) Health(ctx context.Context, req *oraclev1.ProxyHealthRequest) (*oraclev1.ProxyHealthResponse, error) {
	cfg := s.informer.GetConfig()

	// Count healthy nodes
	healthyNodes := 0
	totalNodes := s.ring.Size()

	s.mu.RLock()
	for nodeAddr, client := range s.nodeClients {
		healthResp, err := client.Health(ctx, &oraclev1.HealthRequest{})
		if err == nil && healthResp.Healthy {
			healthyNodes++
		} else {
			s.logger.Warn("Node %s is unhealthy: %v", nodeAddr, err)
		}
	}
	s.mu.RUnlock()

	namespacesCount := 0
	if cfg.Proxy != nil {
		namespacesCount = len(cfg.Proxy.Namespaces)
	}

	return &oraclev1.ProxyHealthResponse{
		Healthy:         healthyNodes > 0,
		NamespacesCount: int32(namespacesCount),
		NodesHealthy:    int32(healthyNodes),
		NodesTotal:      int32(totalNodes),
		Message:         fmt.Sprintf("%d of %d nodes healthy", healthyNodes, totalNodes),
	}, nil
}

// Run starts the proxy server on the specified port.
//
// This method blocks until the server is stopped via Stop() or encounters an error.
//
// Parameters:
//   - port: gRPC port to listen on
//
// Returns:
//   - error: Error if server fails to start or encounters runtime error
func (s *Server) Run(port int) error {
	listener, err := net.Listen("tcp", fmt.Sprintf(":%d", port))
	if err != nil {
		return fmt.Errorf("failed to listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	oraclev1.RegisterProxyServiceServer(grpcServer, s)

	// Register gRPC health check
	grpcHealthServer := grpchealth.NewServer()
	grpc_health_v1.RegisterHealthServer(grpcServer, grpcHealthServer)
	grpcHealthServer.SetServingStatus("", grpc_health_v1.HealthCheckResponse_SERVING)

	// Mark service as healthy and ready
	s.healthChecker.SetHealthy(true)
	s.healthChecker.SetReady(true)

	s.logger.Info("Proxy server listening on port %d", port)

	return grpcServer.Serve(listener)
}

// Stop gracefully shuts down the proxy server.
func (s *Server) Stop() {
	// Mark as unhealthy to stop receiving traffic
	s.healthChecker.SetReady(false)
	s.healthChecker.SetHealthy(false)

	// Stop health checker
	if err := s.healthChecker.Stop(); err != nil {
		s.logger.Error("Failed to stop health checker: %v", err)
	}

	close(s.stopCh)
	s.logger.Info("Proxy server stopped")
}

// StartHealthServer starts the HTTP health check server on the specified port.
//
// This should be called in a goroutine to run concurrently with the main gRPC server.
//
// Parameters:
//   - port: HTTP port for health checks (typically 9090)
//
// Returns:
//   - error: Error if health server fails to start
//
// Example:
//
//	go server.StartHealthServer(9090)
func (s *Server) StartHealthServer(port int) error {
	return s.healthChecker.Start(port)
}

// authenticateRequest validates the API key and returns the corresponding namespace.
func (s *Server) authenticateRequest(apiKey string) (*config.Namespace, bool) {
	return s.informer.GetNamespaceByAPIKey(apiKey)
}

// namespaceKey adds the namespace prefix to a key.
func (s *Server) namespaceKey(namespace, key string) string {
	return fmt.Sprintf("%s:%s", namespace, key)
}

// selectNode uses consistent hashing to select a target cache node.
func (s *Server) selectNode(key string) string {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return s.ring.GetNode(key)
}
