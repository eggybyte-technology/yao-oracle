package main

import (
	"context"
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"strings"
	"syscall"
	"time"

	"github.com/eggybyte-technology/yao-oracle/core/config"
	"github.com/eggybyte-technology/yao-oracle/core/utils"
	"github.com/eggybyte-technology/yao-oracle/internal/dashboard"
)

// Configuration keys - centralized for easy maintenance
const (
	// Infrastructure configuration (from environment variables)
	envHTTPPort    = "HTTP_PORT"    // Business HTTP port
	envHealthPort  = "HEALTH_PORT"  // Health check HTTP port
	envMetricsPort = "METRICS_PORT" // Prometheus metrics port
	envLogLevel    = "LOG_LEVEL"

	// Kubernetes configuration
	envNamespace  = "NAMESPACE"
	envSecretName = "SECRET_NAME"
	envPodName    = "POD_NAME"
	envPodIP      = "POD_IP"

	// Service discovery configuration
	envProxyServiceDNS = "PROXY_SERVICE_DNS"
	envNodeServiceDNS  = "NODE_SERVICE_DNS"
	envDiscoveryMode   = "DISCOVERY_MODE"
	envRefreshInterval = "REFRESH_INTERVAL"

	// Standard port allocation (same across all services)
	defaultHTTPPort        = 8080 // Business gRPC/HTTP port
	defaultHealthPort      = 9090 // Health check port
	defaultMetricsPort     = 9100 // Prometheus metrics port
	defaultLogLevel        = "info"
	defaultNamespace       = "default"
	defaultSecretName      = "yao-oracle-secret"
	defaultDiscoveryMode   = "k8s"
	defaultRefreshInterval = 5
)

// DashboardEnvConfig holds infrastructure configuration loaded from environment variables.
type DashboardEnvConfig struct {
	HTTPPort        int // Business HTTP port (8080)
	HealthPort      int // Health check HTTP port (9090)
	MetricsPort     int // Prometheus metrics port (9100)
	LogLevel        string
	Namespace       string
	SecretName      string
	PodName         string
	PodIP           string
	ProxyServiceDNS string
	NodeServiceDNS  string
	DiscoveryMode   string
	RefreshInterval int
}

// loadEnvConfig loads infrastructure configuration from environment variables.
func loadEnvConfig() DashboardEnvConfig {
	cfg := DashboardEnvConfig{
		HTTPPort:        defaultHTTPPort,
		HealthPort:      defaultHealthPort,
		MetricsPort:     defaultMetricsPort,
		LogLevel:        defaultLogLevel,
		Namespace:       defaultNamespace,
		SecretName:      defaultSecretName,
		DiscoveryMode:   defaultDiscoveryMode,
		RefreshInterval: defaultRefreshInterval,
	}

	// Load HTTP port (business port)
	if portStr := os.Getenv(envHTTPPort); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			cfg.HTTPPort = p
		}
	}

	// Load health check port
	if portStr := os.Getenv(envHealthPort); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			cfg.HealthPort = p
		}
	}

	// Load metrics port
	if portStr := os.Getenv(envMetricsPort); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			cfg.MetricsPort = p
		}
	}

	// Load log level
	if level := os.Getenv(envLogLevel); level != "" {
		cfg.LogLevel = level
	}

	// Load Kubernetes configuration
	if ns := os.Getenv(envNamespace); ns != "" {
		cfg.Namespace = ns
	}
	if secret := os.Getenv(envSecretName); secret != "" {
		cfg.SecretName = secret
	}
	cfg.PodName = os.Getenv(envPodName)
	cfg.PodIP = os.Getenv(envPodIP)

	// Load service discovery configuration
	cfg.ProxyServiceDNS = os.Getenv(envProxyServiceDNS)
	cfg.NodeServiceDNS = os.Getenv(envNodeServiceDNS)
	if mode := os.Getenv(envDiscoveryMode); mode != "" {
		cfg.DiscoveryMode = mode
	}
	if intervalStr := os.Getenv(envRefreshInterval); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			cfg.RefreshInterval = interval
		}
	}

	return cfg
}

func main() {
	logger := utils.NewLogger("dashboard-main")

	// Print banner
	printBanner(logger)

	// Check if running in test mode
	testMode := os.Getenv("TEST_MODE") == "true"

	if testMode {
		runTestMode(logger)
		return
	}

	// Step 1: Load infrastructure config from environment variables
	logger.Step(1, 7, "Loading infrastructure configuration from environment")
	envCfg := loadEnvConfig()

	// Command line flags can override environment variables
	flagPort := flag.Int("port", envCfg.HTTPPort, "HTTP port to listen on (env: HTTP_PORT)")
	flagProxyAddr := flag.String("proxy", "", "Proxy address (e.g., proxy:8080)")
	flagNodes := flag.String("nodes", "", "Comma-separated cache node addresses")
	flag.Parse()

	// Use flag values (which may be env defaults or CLI overrides)
	envCfg.HTTPPort = *flagPort

	logger.Info("HTTP port: %d (business HTTP, from %s)", envCfg.HTTPPort, envOrDefault(envHTTPPort, "default"))
	logger.Info("Health port: %d (health check, from %s)", envCfg.HealthPort, envOrDefault(envHealthPort, "default"))
	logger.Info("Metrics port: %d (Prometheus, from %s)", envCfg.MetricsPort, envOrDefault(envMetricsPort, "default"))
	logger.Info("Log level: %s", envCfg.LogLevel)
	logger.Info("Kubernetes namespace: %s", envCfg.Namespace)
	logger.Info("Secret name: %s", envCfg.SecretName)
	logger.Info("Refresh interval: %d seconds", envCfg.RefreshInterval)
	if envCfg.PodName != "" {
		logger.Info("Pod name: %s", envCfg.PodName)
		logger.Info("Pod IP: %s", envCfg.PodIP)
	}

	// Step 2: Check runtime environment
	logger.Step(2, 7, "Checking runtime environment")
	logger.Info("Go version: %s", runtime.Version())
	logger.Info("OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logger.Info("CPU cores: %d", runtime.NumCPU())

	// Step 3: Initialize Kubernetes configuration loader
	logger.Step(3, 7, "Initializing Kubernetes configuration loader")
	ctx := context.Background()

	k8sLoader, err := config.NewK8sConfigLoader()
	if err != nil {
		logger.Fatal("Failed to create Kubernetes config loader: %v", err)
	}
	logger.Success("Kubernetes config loader initialized")

	// Step 4: Load initial configuration from Kubernetes Secret
	logger.Step(4, 7, "Loading configuration from Kubernetes Secret")
	dashboardCfg, err := k8sLoader.LoadDashboardConfig(ctx, envCfg.Namespace, envCfg.SecretName)
	if err != nil {
		logger.Fatal("Failed to load dashboard configuration: %v", err)
	}
	logger.Success("Dashboard configuration loaded")
	logger.Info("Authentication: enabled")
	logger.Info("JWT secret: configured")
	if dashboardCfg.Theme != "" {
		logger.Info("Theme: %s", dashboardCfg.Theme)
	}

	// Step 5: Initialize Kubernetes Informer for hot reload
	logger.Step(5, 7, "Initializing Kubernetes Informer for config hot reload")
	informer, err := config.NewK8sInformer(config.K8sInformerConfig{
		Namespace:  envCfg.Namespace,
		SecretName: envCfg.SecretName,
	})
	if err != nil {
		logger.Fatal("Failed to create Kubernetes Informer: %v", err)
	}

	// Start informer with reload callback
	go func() {
		err := informer.Start(ctx, func(kind string, data map[string][]byte) {
			logger.Info("🔄 Configuration updated: %s", kind)
			// The informer automatically updates its internal config cache
			newCfg := informer.GetConfig()
			if newCfg.Dashboard != nil {
				logger.Info("Dashboard config reloaded")
			}
			if newCfg.Proxy != nil {
				logger.Info("Monitoring %d namespaces", len(newCfg.Proxy.Namespaces))
			}
		})
		if err != nil {
			logger.Error("Informer error: %v", err)
		}
	}()

	// Wait a bit for initial cache sync
	time.Sleep(time.Second)
	logger.Success("Kubernetes Informer started, watching for config changes")

	// Get current config to display namespace count
	currentCfg := informer.GetConfig()
	if currentCfg.Proxy != nil {
		logger.Info("Monitoring %d namespaces", len(currentCfg.Proxy.Namespaces))
	}

	// Step 6: Create dashboard server with informer
	logger.Step(6, 7, "Creating dashboard server")

	// Parse node addresses
	var nodeAddrs []string
	if *flagNodes != "" {
		nodeAddrs = strings.Split(*flagNodes, ",")
	}

	// Override proxy address if provided via CLI
	proxyAddr := envCfg.ProxyServiceDNS
	if *flagProxyAddr != "" {
		proxyAddr = *flagProxyAddr
	}

	server := dashboard.NewServer(informer, proxyAddr, nodeAddrs, envCfg.RefreshInterval)
	logger.Success("Dashboard server instance created")

	// Log service discovery configuration
	if proxyAddr != "" {
		logger.Info("Proxy service: %s", proxyAddr)
	} else {
		logger.Warn("No Proxy service configured")
	}
	if len(nodeAddrs) > 0 {
		logger.Info("Cache nodes: %d configured", len(nodeAddrs))
		for i, addr := range nodeAddrs {
			logger.Info("  Node %d: %s", i+1, addr)
		}
	} else {
		logger.Info("Using Kubernetes service discovery for cache nodes")
		logger.Info("Node service: %s", envCfg.NodeServiceDNS)
	}

	// Step 7: Setup graceful shutdown
	logger.Step(7, 7, "Setting up graceful shutdown handler")
	setupGracefulShutdown(logger, informer, server)

	// Start health check server (independent HTTP server for K8s probes)
	go func() {
		if err := server.StartHealthServer(envCfg.HealthPort); err != nil {
			logger.Error("Health server error: %v", err)
		}
	}()

	// Start server
	logger.Success("Initialization complete!")
	logger.Info("Starting dashboard HTTP server on port %d (HTTP)", envCfg.HTTPPort)
	logger.Info("Dashboard URL: http://localhost:%d", envCfg.HTTPPort)
	logger.Info("Health check server on port %d (HTTP)", envCfg.HealthPort)
	logger.Info("Metrics available on port %d (HTTP)", envCfg.MetricsPort)
	logger.Info("Press Ctrl+C to stop")
	logger.Info("═══════════════════════════════════════════════════════")

	if err := server.Run(envCfg.HTTPPort); err != nil {
		logger.Fatal("Failed to run dashboard server: %v", err)
	}
}

// setupGracefulShutdown registers signal handlers for graceful termination.
func setupGracefulShutdown(logger *utils.Logger, informer *config.K8sInformer, server *dashboard.Server) {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan

		logger.Warn("Received signal: %v", sig)
		logger.Info("Initiating graceful shutdown...")

		// Stop Kubernetes Informer
		if informer != nil {
			logger.Info("Stopping Kubernetes Informer...")
			informer.Stop()
		}

		// Stop dashboard server
		if server != nil {
			logger.Info("Stopping dashboard server...")
			server.Stop()
		}

		logger.Success("Dashboard server shut down gracefully")
		os.Exit(0)
	}()
}

// envOrDefault returns the source of an environment variable value.
func envOrDefault(envKey string, defaultValue string) string {
	if os.Getenv(envKey) != "" {
		return fmt.Sprintf("env:%s", envKey)
	}
	return defaultValue
}

// runTestMode starts the dashboard in test mode with mock data.
func runTestMode(logger *utils.Logger) {
	logger.Info("═══════════════════════════════════════════════════════")
	logger.Info("🧪 TEST MODE ENABLED")
	logger.Info("═══════════════════════════════════════════════════════")

	// Load test configuration from environment
	envCfg := loadEnvConfig()

	// Parse command line flags
	flagPort := flag.Int("port", envCfg.HTTPPort, "HTTP port to listen on")
	flag.Parse()
	envCfg.HTTPPort = *flagPort

	// Get password from environment (with default for test mode)
	password := os.Getenv("DASHBOARD_PASSWORD")
	if password == "" {
		password = "admin123"
		logger.Info("Using default test password: admin123")
	}

	logger.Step(1, 3, "Initializing test server with mock data")
	logger.Info("HTTP port: %d", envCfg.HTTPPort)
	logger.Info("Health port: %d", envCfg.HealthPort)
	logger.Info("Metrics port: %d", envCfg.MetricsPort)
	logger.Info("Refresh interval: %d seconds", envCfg.RefreshInterval)
	logger.Info("Mock data: 3 namespaces, 3 cache nodes")
	logger.Success("Test configuration loaded")

	// Create test server with mock data
	logger.Step(2, 3, "Creating test server with mock data generator")
	server := dashboard.NewTestServer(password, envCfg.RefreshInterval)
	logger.Success("Test server created with mock data")

	// Setup graceful shutdown
	logger.Step(3, 3, "Setting up graceful shutdown handler")
	setupTestGracefulShutdown(logger, server)

	// Start health check server
	go func() {
		if err := server.StartHealthServer(envCfg.HealthPort); err != nil {
			logger.Error("Health server error: %v", err)
		}
	}()

	// Start server
	logger.Success("Initialization complete!")
	logger.Info("Starting dashboard HTTP server on port %d (HTTP)", envCfg.HTTPPort)
	logger.Info("Dashboard URL: http://localhost:%d", envCfg.HTTPPort)
	logger.Info("Health check server on port %d (HTTP)", envCfg.HealthPort)
	logger.Info("Metrics available on port %d (HTTP)", envCfg.MetricsPort)
	logger.Info("═══════════════════════════════════════════════════════")
	logger.Warn("⚠️  TEST MODE - Using mock data (no real backend)")
	logger.Info("═══════════════════════════════════════════════════════")
	logger.Info("Press Ctrl+C to stop")
	logger.Info("═══════════════════════════════════════════════════════")

	if err := server.Run(envCfg.HTTPPort); err != nil {
		logger.Fatal("Failed to run dashboard server: %v", err)
	}
}

// setupTestGracefulShutdown registers signal handlers for test mode.
func setupTestGracefulShutdown(logger *utils.Logger, server *dashboard.Server) {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan

		logger.Warn("Received signal: %v", sig)
		logger.Info("Initiating graceful shutdown...")

		// Stop dashboard server
		if server != nil {
			logger.Info("Stopping test server...")
			server.Stop()
		}

		logger.Success("Test server shut down gracefully")
		os.Exit(0)
	}()
}

func printBanner(logger *utils.Logger) {
	banner := `
╔═══════════════════════════════════════════════════════╗
║                                                       ║
║          🔮 Yao-Oracle Distributed KV Cache          ║
║                  Dashboard Service                    ║
║                                                       ║
╚═══════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
	logger.Info("Starting Dashboard Service...")
}
