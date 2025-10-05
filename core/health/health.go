// Package health provides a unified HTTP health check server for microservices.
//
// This package implements standard Kubernetes liveness and readiness probes via
// HTTP endpoints. Services can start a lightweight HTTP health server alongside
// their main service (gRPC or HTTP).
//
// Example usage:
//
//	checker := health.NewChecker()
//	checker.SetHealthy(true)
//	checker.SetReady(true)
//	go checker.Start(9090) // Start health server on port 9090
package health

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"sync/atomic"
	"time"

	"github.com/eggybyte-technology/yao-oracle/core/utils"
)

// Checker implements health and readiness checks for Kubernetes probes.
//
// The checker maintains two independent states:
//   - Healthy: Indicates if the service is alive (liveness probe)
//   - Ready: Indicates if the service can handle requests (readiness probe)
//
// Thread-safety: All methods are safe for concurrent use via atomic operations.
type Checker struct {
	healthy    atomic.Bool // Liveness state
	ready      atomic.Bool // Readiness state
	startTime  time.Time   // Service start timestamp
	logger     *utils.Logger
	httpServer *http.Server
}

// HealthResponse represents the JSON response for health check endpoints.
type HealthResponse struct {
	Status  string `json:"status"`            // "healthy" or "unhealthy"
	Uptime  int64  `json:"uptime_seconds"`    // Service uptime in seconds
	Message string `json:"message,omitempty"` // Optional status message
}

// ReadyResponse represents the JSON response for readiness check endpoint.
type ReadyResponse struct {
	Ready   bool   `json:"ready"`             // Readiness state
	Uptime  int64  `json:"uptime_seconds"`    // Service uptime in seconds
	Message string `json:"message,omitempty"` // Optional status message
}

// NewChecker creates a new health checker instance.
//
// The checker is initialized with healthy=false and ready=false.
// Services should explicitly set these states after initialization is complete.
//
// Returns:
//   - *Checker: A new health checker instance ready to start
//
// Example:
//
//	checker := health.NewChecker()
//	checker.SetHealthy(true)
//	checker.SetReady(true)
func NewChecker() *Checker {
	c := &Checker{
		startTime: time.Now(),
		logger:    utils.NewLogger("health"),
	}
	// Default to unhealthy/not ready until service initializes
	c.healthy.Store(false)
	c.ready.Store(false)
	return c
}

// SetHealthy updates the liveness state.
//
// This should be called to indicate whether the service is alive.
// Set to false to signal that the service should be restarted (Kubernetes will kill the pod).
//
// Parameters:
//   - healthy: true if service is alive, false if service should be restarted
//
// Thread-safety: Safe for concurrent use.
func (c *Checker) SetHealthy(healthy bool) {
	c.healthy.Store(healthy)
	if healthy {
		c.logger.Info("Service marked as HEALTHY (liveness probe)")
	} else {
		c.logger.Warn("Service marked as UNHEALTHY (liveness probe)")
	}
}

// SetReady updates the readiness state.
//
// This should be called to indicate whether the service can handle requests.
// Set to false to temporarily remove the pod from load balancing without restarting it.
//
// Parameters:
//   - ready: true if service can handle requests, false to remove from load balancer
//
// Thread-safety: Safe for concurrent use.
func (c *Checker) SetReady(ready bool) {
	c.ready.Store(ready)
	if ready {
		c.logger.Info("Service marked as READY (readiness probe)")
	} else {
		c.logger.Warn("Service marked as NOT READY (readiness probe)")
	}
}

// IsHealthy returns the current liveness state.
//
// Returns:
//   - bool: true if service is alive, false otherwise
//
// Thread-safety: Safe for concurrent use.
func (c *Checker) IsHealthy() bool {
	return c.healthy.Load()
}

// IsReady returns the current readiness state.
//
// Returns:
//   - bool: true if service can handle requests, false otherwise
//
// Thread-safety: Safe for concurrent use.
func (c *Checker) IsReady() bool {
	return c.ready.Load()
}

// Start begins the HTTP health check server on the specified port.
//
// This method blocks until the server is stopped via Stop() or encounters an error.
// The server provides the following endpoints:
//
//   - GET /health - Liveness probe (returns 200 if healthy, 503 if unhealthy)
//   - GET /ready - Readiness probe (returns 200 if ready, 503 if not ready)
//   - GET /healthz - Alias for /health (common Kubernetes convention)
//   - GET /readyz - Alias for /ready (common Kubernetes convention)
//
// Parameters:
//   - port: HTTP port to listen on for health checks
//
// Returns:
//   - error: Error if server fails to start or encounters runtime error
//
// Example:
//
//	checker := health.NewChecker()
//	checker.SetHealthy(true)
//	checker.SetReady(true)
//	go checker.Start(9090) // Non-blocking start
func (c *Checker) Start(port int) error {
	mux := http.NewServeMux()

	// Liveness probe endpoints
	mux.HandleFunc("/health", c.handleHealth)
	mux.HandleFunc("/healthz", c.handleHealth) // Kubernetes convention alias

	// Readiness probe endpoints
	mux.HandleFunc("/ready", c.handleReady)
	mux.HandleFunc("/readyz", c.handleReady) // Kubernetes convention alias

	c.httpServer = &http.Server{
		Addr:         fmt.Sprintf(":%d", port),
		Handler:      mux,
		ReadTimeout:  5 * time.Second,
		WriteTimeout: 5 * time.Second,
		IdleTimeout:  15 * time.Second,
	}

	c.logger.Info("Starting health check server on port %d", port)
	c.logger.Info("  - Liveness probe: http://:%d/health", port)
	c.logger.Info("  - Readiness probe: http://:%d/ready", port)

	if err := c.httpServer.ListenAndServe(); err != nil && err != http.ErrServerClosed {
		return fmt.Errorf("health server error: %w", err)
	}

	return nil
}

// Stop gracefully shuts down the health check server.
//
// This method blocks until all active connections are closed or the timeout is reached.
//
// Thread-safety: Safe for concurrent use.
func (c *Checker) Stop() error {
	if c.httpServer == nil {
		return nil
	}

	c.logger.Info("Stopping health check server...")
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()

	if err := c.httpServer.Shutdown(ctx); err != nil {
		return fmt.Errorf("failed to shutdown health server: %w", err)
	}

	c.logger.Info("Health check server stopped")
	return nil
}

// handleHealth handles liveness probe requests at /health and /healthz.
func (c *Checker) handleHealth(w http.ResponseWriter, r *http.Request) {
	uptime := int64(time.Since(c.startTime).Seconds())

	if c.IsHealthy() {
		resp := HealthResponse{
			Status:  "healthy",
			Uptime:  uptime,
			Message: "Service is alive",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Service is unhealthy - Kubernetes will restart the pod
	resp := HealthResponse{
		Status:  "unhealthy",
		Uptime:  uptime,
		Message: "Service is not alive, restart required",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable) // 503
	json.NewEncoder(w).Encode(resp)
}

// handleReady handles readiness probe requests at /ready and /readyz.
func (c *Checker) handleReady(w http.ResponseWriter, r *http.Request) {
	uptime := int64(time.Since(c.startTime).Seconds())

	if c.IsReady() {
		resp := ReadyResponse{
			Ready:   true,
			Uptime:  uptime,
			Message: "Service is ready to handle requests",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		json.NewEncoder(w).Encode(resp)
		return
	}

	// Service is not ready - Kubernetes will remove from load balancer
	resp := ReadyResponse{
		Ready:   false,
		Uptime:  uptime,
		Message: "Service is not ready, temporarily unavailable",
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusServiceUnavailable) // 503
	json.NewEncoder(w).Encode(resp)
}
