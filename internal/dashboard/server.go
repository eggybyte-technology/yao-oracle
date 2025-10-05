package dashboard

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"

	oraclev1 "github.com/eggybyte-technology/yao-oracle/pb/yao/oracle/v1"

	"github.com/eggybyte-technology/yao-oracle/core/config"
	"github.com/eggybyte-technology/yao-oracle/core/health"
	"github.com/eggybyte-technology/yao-oracle/core/utils"
)

// Server implements the Dashboard HTTP server.
//
// The dashboard provides a web-based monitoring interface for:
//   - Cluster health visualization
//   - Namespace statistics and metrics
//   - Cache node monitoring
//   - Real-time performance charts
//
// Security:
//   - Password authentication via JWT
//   - Session management with configurable timeout
//   - HTTPS support (when configured)
//
// Thread-safety: All methods are safe for concurrent use.
type Server struct {
	mu              sync.RWMutex
	informer        ConfigInformer
	proxyClient     oraclev1.ProxyServiceClient
	nodeClients     map[string]oraclev1.NodeServiceClient
	healthChecker   *health.Checker
	logger          *utils.Logger
	sessions        map[string]time.Time // Simple session management
	sessionsMu      sync.RWMutex
	proxyAddr       string
	nodeAddrs       []string
	refreshInterval int
	stopCh          chan struct{}
	mockGenerator   *MockDataGenerator // For test mode
	testMode        bool               // Whether running in test mode
}

// ConfigInformer is an interface for configuration providers.
// This allows both real Kubernetes Informer and mock implementations.
type ConfigInformer interface {
	GetConfig() config.Config
	Start(ctx context.Context, onChange func(kind string, data map[string][]byte)) error
	Stop()
}

// NewServer creates a new dashboard server instance with configuration informer.
//
// The informer provides dynamic configuration reloading without restart.
// Dashboard password and namespace information are automatically updated
// when the Kubernetes Secret is modified.
//
// Parameters:
//   - informer: Configuration informer (K8sInformer or MockConfigInformer)
//   - proxyAddr: Address of the proxy service (optional, for health checks)
//   - nodeAddrs: List of cache node addresses (optional, for direct monitoring)
//   - refreshInterval: Dashboard auto-refresh interval in seconds
//
// Returns:
//   - *Server: A new dashboard server instance ready to start
//
// Example:
//
//	informer, _ := config.NewK8sInformer(...)
//	server := dashboard.NewServer(informer, "proxy:8080", []string{"node-0:7070"}, 5)
//	server.Run(8081)
func NewServer(informer ConfigInformer, proxyAddr string, nodeAddrs []string, refreshInterval int) *Server {
	s := &Server{
		informer:        informer,
		nodeClients:     make(map[string]oraclev1.NodeServiceClient),
		healthChecker:   health.NewChecker(),
		logger:          utils.NewLogger("dashboard"),
		sessions:        make(map[string]time.Time),
		proxyAddr:       proxyAddr,
		nodeAddrs:       nodeAddrs,
		refreshInterval: refreshInterval,
		stopCh:          make(chan struct{}),
		testMode:        false,
	}

	// Connect to proxy
	if proxyAddr != "" {
		conn, err := grpc.Dial(proxyAddr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			s.logger.Error("Failed to connect to proxy: %v", err)
		} else {
			s.proxyClient = oraclev1.NewProxyServiceClient(conn)
			s.logger.Info("Connected to proxy: %s", proxyAddr)
		}
	}

	// Connect to nodes
	for _, addr := range nodeAddrs {
		conn, err := grpc.Dial(addr, grpc.WithTransportCredentials(insecure.NewCredentials()))
		if err != nil {
			s.logger.Error("Failed to connect to node %s: %v", addr, err)
			continue
		}
		s.nodeClients[addr] = oraclev1.NewNodeServiceClient(conn)
		s.logger.Info("Connected to node: %s", addr)
	}

	return s
}

// NewTestServer creates a new dashboard server in test mode with mock data.
//
// This is used for testing the dashboard UI without requiring a real backend.
// All metrics and data are simulated.
//
// Parameters:
//   - password: Dashboard password for authentication
//   - refreshInterval: Dashboard auto-refresh interval in seconds
//
// Returns:
//   - *Server: A new dashboard server instance with mock data
//
// Example:
//
//	server := dashboard.NewTestServer("admin123", 5)
//	server.Run(8080)
func NewTestServer(password string, refreshInterval int) *Server {
	mockInformer := NewMockConfigInformer(password)
	mockGenerator := NewMockDataGenerator(refreshInterval)

	s := &Server{
		informer:        mockInformer,
		nodeClients:     make(map[string]oraclev1.NodeServiceClient),
		healthChecker:   health.NewChecker(),
		logger:          utils.NewLogger("dashboard"),
		sessions:        make(map[string]time.Time),
		refreshInterval: refreshInterval,
		stopCh:          make(chan struct{}),
		mockGenerator:   mockGenerator,
		testMode:        true,
	}

	// Setup mock clients
	s.proxyClient = NewMockProxyClient(mockGenerator)
	s.logger.Info("Test mode: Using mock proxy client")

	// Create mock node clients
	for _, nodeData := range mockGenerator.GetNodes() {
		s.nodeClients[nodeData.Address] = NewMockNodeClient(nodeData)
		s.logger.Info("Test mode: Using mock node client for %s", nodeData.Address)
	}

	return s
}

// authenticate checks if the session is valid.
func (s *Server) authenticate(sessionID string) bool {
	s.sessionsMu.RLock()
	defer s.sessionsMu.RUnlock()

	expiry, exists := s.sessions[sessionID]
	if !exists {
		return false
	}

	// Check if session has expired
	return time.Now().Before(expiry)
}

// createSession creates a new session and returns the session ID.
func (s *Server) createSession() string {
	s.sessionsMu.Lock()
	defer s.sessionsMu.Unlock()

	// Generate session ID (simple implementation - use UUID in production)
	sessionID := fmt.Sprintf("session-%d", time.Now().UnixNano())

	// Set session expiry (30 minutes from now)
	cfg := s.informer.GetConfig()
	sessionTimeout := 30 * time.Minute
	if cfg.Dashboard != nil && cfg.Dashboard.RefreshInterval > 0 {
		// Session timeout is 10x the refresh interval (in minutes)
		sessionTimeout = time.Duration(cfg.Dashboard.RefreshInterval*10) * time.Minute
	}
	s.sessions[sessionID] = time.Now().Add(sessionTimeout)

	return sessionID
}

// Run starts the dashboard HTTP server on the specified port.
//
// This method sets up the HTTP routes and blocks until the server is stopped.
//
// Routes:
//   - GET  / - Dashboard HTML page
//   - POST /api/auth/login - Login endpoint
//   - GET  /api/metrics/overview - Cluster overview metrics
//   - GET  /api/metrics/namespaces - Namespace statistics
//   - GET  /api/metrics/nodes - Node health and metrics
//
// Parameters:
//   - port: HTTP port to listen on
//
// Returns:
//   - error: Error if server fails to start or encounters runtime error
func (s *Server) Run(port int) error {
	gin.SetMode(gin.ReleaseMode)
	router := gin.Default()

	// CORS middleware for development (allow all origins)
	router.Use(func(c *gin.Context) {
		c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
		c.Writer.Header().Set("Access-Control-Allow-Credentials", "true")
		c.Writer.Header().Set("Access-Control-Allow-Headers", "Content-Type, Content-Length, Accept-Encoding, X-CSRF-Token, Authorization, accept, origin, Cache-Control, X-Requested-With, X-Session-ID")
		c.Writer.Header().Set("Access-Control-Allow-Methods", "POST, OPTIONS, GET, PUT, DELETE")

		if c.Request.Method == "OPTIONS" {
			c.AbortWithStatus(204)
			return
		}

		c.Next()
	})

	// Serve static files
	router.Static("/css", "./web/css")
	router.Static("/js", "./web/js")
	router.Static("/assets", "./web/assets")

	// Serve HTML pages
	router.GET("/", s.handleIndex)
	router.GET("/login", s.handleLogin)
	router.GET("/dashboard", s.handleDashboard)

	// WebSocket endpoint
	router.GET("/ws", s.handleWebSocket)

	// API routes
	api := router.Group("/api")
	{
		// Authentication (no auth required)
		api.POST("/auth/login", s.handleAPILogin)
		api.POST("/auth/logout", s.handleAPILogout)

		// Overview endpoint (for testing, no auth required in test mode)
		if s.testMode {
			api.GET("/overview", s.handleAPIOverview)
			api.GET("/cluster/timeseries", s.handleClusterTimeseries)
			api.GET("/proxies", s.handleAPIProxies)
			api.GET("/nodes", s.handleAPINodes)
			api.GET("/namespaces", s.handleAPINamespaces)
		}

		// Metrics (auth required)
		api.GET("/metrics/overview", s.authMiddleware(), s.handleMetricsOverview)
		api.GET("/metrics/namespaces", s.authMiddleware(), s.handleMetricsNamespaces)
		api.GET("/metrics/nodes", s.authMiddleware(), s.handleMetricsNodes)
		api.GET("/metrics/proxy", s.authMiddleware(), s.handleMetricsProxy)
	}

	// Mark service as healthy and ready
	s.healthChecker.SetHealthy(true)
	s.healthChecker.SetReady(true)

	s.logger.Info("Dashboard server listening on port %d", port)

	return router.Run(fmt.Sprintf(":%d", port))
}

// Stop gracefully shuts down the dashboard server.
func (s *Server) Stop() {
	// Mark as unhealthy to stop receiving traffic
	s.healthChecker.SetReady(false)
	s.healthChecker.SetHealthy(false)

	// Stop health checker
	if err := s.healthChecker.Stop(); err != nil {
		s.logger.Error("Failed to stop health checker: %v", err)
	}

	// Stop mock generator if in test mode
	if s.testMode && s.mockGenerator != nil {
		s.mockGenerator.Stop()
	}

	close(s.stopCh)
	s.logger.Info("Dashboard server stopped")
}

// StartHealthServer starts the HTTP health check server on the specified port.
//
// This should be called in a goroutine to run concurrently with the main HTTP server.
// The health check server uses the standard core/health implementation which provides
// Kubernetes-compatible liveness and readiness probes.
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

// authMiddleware checks authentication for API endpoints.
func (s *Server) authMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		sessionID := c.GetHeader("X-Session-ID")
		if sessionID == "" {
			// Try cookie
			sessionID, _ = c.Cookie("session_id")
		}

		if !s.authenticate(sessionID) {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "unauthorized"})
			c.Abort()
			return
		}

		c.Next()
	}
}

// handleIndex serves the main page.
func (s *Server) handleIndex(c *gin.Context) {
	c.Redirect(http.StatusFound, "/dashboard")
}

// handleLogin serves the login page.
func (s *Server) handleLogin(c *gin.Context) {
	c.File("./web/login.html")
}

// handleDashboard serves the dashboard page.
func (s *Server) handleDashboard(c *gin.Context) {
	c.File("./web/index.html")
}

// handleAPILogin handles login requests.
func (s *Server) handleAPILogin(c *gin.Context) {
	var req struct {
		Password string `json:"password"`
	}

	if err := c.BindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "invalid request"})
		return
	}

	// Get password from configuration
	cfg := s.informer.GetConfig()
	if cfg.Dashboard == nil || cfg.Dashboard.Password == "" {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "dashboard not configured"})
		return
	}

	// Validate password
	if req.Password != cfg.Dashboard.Password {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "invalid password"})
		return
	}

	// Create session
	sessionID := s.createSession()

	c.JSON(http.StatusOK, gin.H{
		"success":    true,
		"session_id": sessionID,
	})
}

// handleAPILogout handles logout requests.
func (s *Server) handleAPILogout(c *gin.Context) {
	sessionID := c.GetHeader("X-Session-ID")
	if sessionID == "" {
		sessionID, _ = c.Cookie("session_id")
	}

	s.sessionsMu.Lock()
	delete(s.sessions, sessionID)
	s.sessionsMu.Unlock()

	c.JSON(http.StatusOK, gin.H{"success": true})
}

// handleMetricsOverview returns overall cluster metrics.
func (s *Server) handleMetricsOverview(c *gin.Context) {
	ctx := context.Background()
	cfg := s.informer.GetConfig()

	overview := map[string]interface{}{
		"timestamp":  time.Now().Unix(),
		"namespaces": 0,
		"nodes":      len(s.nodeClients),
		"totalKeys":  0,
		"healthy":    true,
	}

	if cfg.Proxy != nil {
		overview["namespaces"] = len(cfg.Proxy.Namespaces)
	}

	// Query proxy health if available
	if s.proxyClient != nil {
		healthResp, err := s.proxyClient.Health(ctx, &oraclev1.ProxyHealthRequest{})
		if err == nil {
			overview["proxyHealthy"] = healthResp.Healthy
			overview["nodesTotal"] = healthResp.NodesTotal
			overview["nodesHealthy"] = healthResp.NodesHealthy
		}
	}

	// Query node statistics
	totalKeys := int64(0)
	for _, client := range s.nodeClients {
		statsResp, err := client.Stats(ctx, &oraclev1.StatsRequest{})
		if err == nil {
			totalKeys += statsResp.TotalKeys
		}
	}
	overview["totalKeys"] = totalKeys

	c.JSON(http.StatusOK, overview)
}

// handleMetricsNamespaces returns namespace statistics.
func (s *Server) handleMetricsNamespaces(c *gin.Context) {
	cfg := s.informer.GetConfig()

	namespaces := []map[string]interface{}{}
	if cfg.Proxy != nil {
		for _, ns := range cfg.Proxy.Namespaces {
			nsInfo := map[string]interface{}{
				"name":         ns.Name,
				"description":  ns.Description,
				"maxMemoryMB":  ns.MaxMemoryMB,
				"defaultTTL":   ns.DefaultTTL,
				"rateLimitQPS": ns.RateLimitQPS,
			}
			namespaces = append(namespaces, nsInfo)
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"namespaces": namespaces,
	})
}

// handleMetricsNodes returns node health and statistics.
func (s *Server) handleMetricsNodes(c *gin.Context) {
	ctx := context.Background()

	nodes := []map[string]interface{}{}
	for addr, client := range s.nodeClients {
		nodeInfo := map[string]interface{}{
			"address": addr,
			"healthy": false,
		}

		// Query health
		healthResp, err := client.Health(ctx, &oraclev1.HealthRequest{})
		if err == nil {
			nodeInfo["healthy"] = healthResp.Healthy
		}

		// Query stats
		statsResp, err := client.Stats(ctx, &oraclev1.StatsRequest{})
		if err == nil {
			nodeInfo["totalKeys"] = statsResp.TotalKeys
			nodeInfo["memoryUsedBytes"] = statsResp.MemoryUsedBytes
			nodeInfo["uptimeSeconds"] = statsResp.UptimeSeconds
			nodeInfo["requestsTotal"] = statsResp.RequestsTotal
			nodeInfo["hits"] = statsResp.Hits
			nodeInfo["misses"] = statsResp.Misses
			if statsResp.RequestsTotal > 0 {
				nodeInfo["hitRate"] = float64(statsResp.Hits) / float64(statsResp.RequestsTotal)
			}
		}

		nodes = append(nodes, nodeInfo)
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodes,
	})
}

// handleMetricsProxy returns proxy health and statistics.
func (s *Server) handleMetricsProxy(c *gin.Context) {
	ctx := context.Background()

	if s.proxyClient == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "proxy not configured"})
		return
	}

	healthResp, err := s.proxyClient.Health(ctx, &oraclev1.ProxyHealthRequest{})
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": err.Error()})
		return
	}

	proxyInfo := map[string]interface{}{
		"healthy":         healthResp.Healthy,
		"namespacesCount": healthResp.NamespacesCount,
		"nodesHealthy":    healthResp.NodesHealthy,
		"nodesTotal":      healthResp.NodesTotal,
		"message":         healthResp.Message,
	}

	c.JSON(http.StatusOK, gin.H{
		"proxy": proxyInfo,
	})
}

// handleWebSocket handles WebSocket connections (stub for now).
func (s *Server) handleWebSocket(c *gin.Context) {
	// WebSocket support is planned but not yet implemented
	// For now, return a friendly error message
	c.JSON(http.StatusNotImplemented, gin.H{
		"error":   "WebSocket not implemented",
		"message": "WebSocket streaming is planned for future release",
	})
}

// handleAPIOverview returns simplified overview data (test mode).
func (s *Server) handleAPIOverview(c *gin.Context) {
	if !s.testMode || s.mockGenerator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not available"})
		return
	}

	proxyHealth := s.mockGenerator.GetProxyHealth()
	nodes := s.mockGenerator.GetNodes()
	namespaces := s.mockGenerator.GetNamespaces()

	totalKeys := int64(0)
	totalMemory := int64(0)
	for _, node := range nodes {
		totalKeys += node.TotalKeys
		totalMemory += node.MemoryUsedBytes
	}

	c.JSON(http.StatusOK, gin.H{
		"timestamp":    time.Now().Unix(),
		"status":       "healthy",
		"namespaces":   len(namespaces),
		"nodes":        len(nodes),
		"nodesHealthy": proxyHealth.NodesHealthy,
		"totalKeys":    totalKeys,
		"totalMemory":  totalMemory,
		"qps":          proxyHealth.RequestsPerSec,
	})
}

// handleClusterTimeseries returns mock timeseries data (test mode).
func (s *Server) handleClusterTimeseries(c *gin.Context) {
	if !s.testMode || s.mockGenerator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not available"})
		return
	}

	// Generate some mock timeseries data
	now := time.Now().Unix()
	points := make([]map[string]interface{}, 10)
	for i := 0; i < 10; i++ {
		points[i] = map[string]interface{}{
			"timestamp": now - int64((10-i)*60),
			"qps":       100.0 + float64(i*10),
			"latency":   2.5 + float64(i)*0.2,
			"hitRate":   0.85 + float64(i)*0.01,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"timeseries": points,
	})
}

// handleAPIProxies returns proxy information (test mode).
func (s *Server) handleAPIProxies(c *gin.Context) {
	if !s.testMode || s.mockGenerator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not available"})
		return
	}

	proxyHealth := s.mockGenerator.GetProxyHealth()
	c.JSON(http.StatusOK, gin.H{
		"proxies": []map[string]interface{}{
			{
				"address":    "proxy-0:8080",
				"healthy":    proxyHealth.Healthy,
				"namespaces": proxyHealth.NamespacesCount,
				"nodes":      proxyHealth.NodesTotal,
				"qps":        proxyHealth.RequestsPerSec,
			},
		},
	})
}

// handleAPINodes returns node information (test mode).
func (s *Server) handleAPINodes(c *gin.Context) {
	if !s.testMode || s.mockGenerator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not available"})
		return
	}

	nodes := s.mockGenerator.GetNodes()
	nodesList := make([]map[string]interface{}, len(nodes))
	for i, node := range nodes {
		hitRate := 0.0
		if node.RequestsTotal > 0 {
			hitRate = float64(node.Hits) / float64(node.RequestsTotal)
		}

		nodesList[i] = map[string]interface{}{
			"address":    node.Address,
			"healthy":    node.Healthy,
			"totalKeys":  node.TotalKeys,
			"memoryUsed": node.MemoryUsedBytes,
			"memoryMax":  node.MemoryMaxBytes,
			"hitRate":    hitRate,
			"uptime":     node.UptimeSeconds,
			"qps":        float64(node.RequestsTotal) / float64(node.UptimeSeconds),
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"nodes": nodesList,
	})
}

// handleAPINamespaces returns namespace information (test mode).
func (s *Server) handleAPINamespaces(c *gin.Context) {
	if !s.testMode || s.mockGenerator == nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "not available"})
		return
	}

	namespaces := s.mockGenerator.GetNamespaces()
	nsList := make([]map[string]interface{}, len(namespaces))
	for i, ns := range namespaces {
		nsList[i] = map[string]interface{}{
			"name":        ns.Name,
			"description": ns.Description,
			"keyCount":    ns.KeyCount,
			"hitRate":     ns.HitRate,
			"maxMemory":   ns.MaxMemoryMB * 1024 * 1024,
			"defaultTTL":  ns.DefaultTTL,
			"rateLimit":   ns.RateLimitQPS,
		}
	}

	c.JSON(http.StatusOK, gin.H{
		"namespaces": nsList,
	})
}

// MarshalJSON converts the server config to JSON for debugging.
func (s *Server) MarshalJSON() ([]byte, error) {
	s.mu.RLock()
	defer s.mu.RUnlock()

	return json.Marshal(map[string]interface{}{
		"proxyAddr":       s.proxyAddr,
		"nodeAddrs":       s.nodeAddrs,
		"refreshInterval": s.refreshInterval,
	})
}
