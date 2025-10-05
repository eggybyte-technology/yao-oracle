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
	"github.com/eggybyte-technology/yao-oracle/internal/proxy"
)

// Configuration keys - centralized for easy maintenance
const (
	// Infrastructure configuration (from environment variables)
	envGRPCPort    = "GRPC_PORT"    // Business gRPC port
	envHealthPort  = "HEALTH_PORT"  // Health check HTTP port
	envMetricsPort = "METRICS_PORT" // Prometheus metrics port
	envLogLevel    = "LOG_LEVEL"

	// Kubernetes configuration
	envNamespace  = "NAMESPACE"
	envSecretName = "SECRET_NAME"
	envPodName    = "POD_NAME"
	envPodIP      = "POD_IP"

	// Service discovery configuration
	envNodeHeadlessService = "NODE_HEADLESS_SERVICE"
	envDiscoveryMode       = "DISCOVERY_MODE"
	envDiscoveryInterval   = "DISCOVERY_INTERVAL"

	// Standard port allocation (same across all services)
	defaultGRPCPort          = 8080 // Business gRPC/HTTP port
	defaultHealthPort        = 9090 // Health check port
	defaultMetricsPort       = 9100 // Prometheus metrics port
	defaultLogLevel          = "info"
	defaultNamespace         = "default"
	defaultSecretName        = "yao-oracle-secret"
	defaultDiscoveryMode     = "k8s"
	defaultDiscoveryInterval = 10
)

// ProxyEnvConfig holds infrastructure configuration loaded from environment variables.
type ProxyEnvConfig struct {
	GRPCPort          int // Business gRPC port (8080)
	HealthPort        int // Health check HTTP port (9090)
	MetricsPort       int // Prometheus metrics port (9100)
	LogLevel          string
	Namespace         string
	SecretName        string
	PodName           string
	PodIP             string
	NodeService       string
	DiscoveryMode     string
	DiscoveryInterval int
}

// loadEnvConfig loads infrastructure configuration from environment variables.
func loadEnvConfig() ProxyEnvConfig {
	cfg := ProxyEnvConfig{
		GRPCPort:          defaultGRPCPort,
		HealthPort:        defaultHealthPort,
		MetricsPort:       defaultMetricsPort,
		LogLevel:          defaultLogLevel,
		Namespace:         defaultNamespace,
		SecretName:        defaultSecretName,
		DiscoveryMode:     defaultDiscoveryMode,
		DiscoveryInterval: defaultDiscoveryInterval,
	}

	// Load GRPC port (business port)
	if portStr := os.Getenv(envGRPCPort); portStr != "" {
		if p, err := strconv.Atoi(portStr); err == nil && p > 0 {
			cfg.GRPCPort = p
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
	cfg.NodeService = os.Getenv(envNodeHeadlessService)
	if mode := os.Getenv(envDiscoveryMode); mode != "" {
		cfg.DiscoveryMode = mode
	}
	if intervalStr := os.Getenv(envDiscoveryInterval); intervalStr != "" {
		if interval, err := strconv.Atoi(intervalStr); err == nil && interval > 0 {
			cfg.DiscoveryInterval = interval
		}
	}

	return cfg
}

func main() {
	logger := utils.NewLogger("proxy-main")

	// Print banner
	printBanner(logger)

	// Step 1: Load infrastructure config from environment variables
	logger.Step(1, 7, "Loading infrastructure configuration from environment")
	envCfg := loadEnvConfig()

	// Command line flags can override environment variables
	flagPort := flag.Int("port", envCfg.GRPCPort, "gRPC port to listen on (env: GRPC_PORT)")
	flagNodes := flag.String("nodes", "", "Comma-separated cache node addresses (for testing)")
	flag.Parse()

	// Use flag values (which may be env defaults or CLI overrides)
	envCfg.GRPCPort = *flagPort

	logger.Info("GRPC port: %d (business gRPC, from %s)", envCfg.GRPCPort, envOrDefault(envGRPCPort, "default"))
	logger.Info("Health port: %d (health check, from %s)", envCfg.HealthPort, envOrDefault(envHealthPort, "default"))
	logger.Info("Metrics port: %d (Prometheus, from %s)", envCfg.MetricsPort, envOrDefault(envMetricsPort, "default"))
	logger.Info("Log level: %s", envCfg.LogLevel)
	logger.Info("Kubernetes namespace: %s", envCfg.Namespace)
	logger.Info("Secret name: %s", envCfg.SecretName)
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
	proxyCfg, err := k8sLoader.LoadProxyConfig(ctx, envCfg.Namespace, envCfg.SecretName)
	if err != nil {
		logger.Fatal("Failed to load proxy configuration: %v", err)
	}
	logger.Success("Configuration loaded: %d namespaces configured", len(proxyCfg.Namespaces))
	for _, ns := range proxyCfg.Namespaces {
		logger.Info("  - Namespace: %s (%s)", ns.Name, ns.Description)
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
			if newCfg.Proxy != nil {
				logger.Info("Reloaded: %d namespaces", len(newCfg.Proxy.Namespaces))
			}
		})
		if err != nil {
			logger.Error("Informer error: %v", err)
		}
	}()

	// Wait a bit for initial cache sync
	time.Sleep(time.Second)
	logger.Success("Kubernetes Informer started, watching for config changes")

	// Step 6: Create proxy server with informer
	logger.Step(6, 7, "Creating proxy server")
	server := proxy.NewServer(informer)
	logger.Success("Proxy server instance created")

	// Configure cache nodes (for testing)
	if *flagNodes != "" {
		nodeList := strings.Split(*flagNodes, ",")
		server.SetNodes(nodeList)
		logger.Info("Configured %d static cache nodes (testing mode)", len(nodeList))
		for i, node := range nodeList {
			logger.Info("  Node %d: %s", i+1, node)
		}
	} else {
		logger.Info("Using Kubernetes service discovery for cache nodes")
		logger.Info("Node service: %s", envCfg.NodeService)
		logger.Info("Discovery mode: %s", envCfg.DiscoveryMode)
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
	logger.Info("Starting proxy server on port %d (gRPC)", envCfg.GRPCPort)
	logger.Info("Health check server on port %d (HTTP)", envCfg.HealthPort)
	logger.Info("Metrics available on port %d (HTTP)", envCfg.MetricsPort)
	logger.Info("Press Ctrl+C to stop")
	logger.Info("═══════════════════════════════════════════════════════")

	if err := server.Run(envCfg.GRPCPort); err != nil {
		logger.Fatal("Failed to run proxy server: %v", err)
	}
}

// setupGracefulShutdown registers signal handlers for graceful termination.
func setupGracefulShutdown(logger *utils.Logger, informer *config.K8sInformer, server *proxy.Server) {
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

		// Stop proxy server
		if server != nil {
			logger.Info("Stopping proxy server...")
			server.Stop()
		}

		logger.Success("Proxy server shut down gracefully")
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

func printBanner(logger *utils.Logger) {
	banner := `
╔═══════════════════════════════════════════════════════╗
║                                                       ║
║          🔮 Yao-Oracle Distributed KV Cache          ║
║                    Proxy Service                      ║
║                                                       ║
╚═══════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
	logger.Info("Starting Proxy Service...")
}
