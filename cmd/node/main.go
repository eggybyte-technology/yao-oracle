package main

import (
	"flag"
	"fmt"
	"os"
	"os/signal"
	"runtime"
	"strconv"
	"syscall"

	"github.com/eggybyte-technology/yao-oracle/core/utils"
	"github.com/eggybyte-technology/yao-oracle/internal/node"
)

// Configuration keys - centralized for easy maintenance
const (
	// Infrastructure configuration (from environment variables)
	envGRPCPort    = "GRPC_PORT"    // Business gRPC port
	envHealthPort  = "HEALTH_PORT"  // Health check HTTP port
	envMetricsPort = "METRICS_PORT" // Prometheus metrics port
	envLogLevel    = "LOG_LEVEL"
	envMaxMemoryMB = "MAX_MEMORY_MB"
	envMaxKeys     = "MAX_KEYS"

	// Pod metadata (auto-injected by Kubernetes)
	envPodName      = "POD_NAME"
	envPodNamespace = "POD_NAMESPACE"

	// Standard port allocation (same across all services)
	defaultGRPCPort    = 8080 // Business gRPC/HTTP port
	defaultHealthPort  = 9090 // Health check port
	defaultMetricsPort = 9100 // Prometheus metrics port
	defaultLogLevel    = "info"
	defaultMaxMemoryMB = 512
	defaultMaxKeys     = 100000
)

// NodeConfig holds the cache node configuration.
type NodeConfig struct {
	GRPCPort    int // Business gRPC port (8080)
	HealthPort  int // Health check HTTP port (9090)
	MetricsPort int // Prometheus metrics port (9100)
	LogLevel    string
	MaxMemoryMB int
	MaxKeys     int
}

// loadEnvConfig loads infrastructure configuration from environment variables.
func loadEnvConfig() NodeConfig {
	cfg := NodeConfig{
		GRPCPort:    defaultGRPCPort,
		HealthPort:  defaultHealthPort,
		MetricsPort: defaultMetricsPort,
		LogLevel:    defaultLogLevel,
		MaxMemoryMB: defaultMaxMemoryMB,
		MaxKeys:     defaultMaxKeys,
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

	// Load max memory
	if memStr := os.Getenv(envMaxMemoryMB); memStr != "" {
		if mem, err := strconv.Atoi(memStr); err == nil && mem > 0 {
			cfg.MaxMemoryMB = mem
		}
	}

	// Load max keys
	if keysStr := os.Getenv(envMaxKeys); keysStr != "" {
		if keys, err := strconv.Atoi(keysStr); err == nil && keys > 0 {
			cfg.MaxKeys = keys
		}
	}

	return cfg
}

func main() {
	logger := utils.NewLogger("node-main")

	// Print banner
	printBanner(logger)

	// Step 1: Load infrastructure config from environment variables
	logger.Step(1, 4, "Loading infrastructure configuration from environment")
	cfg := loadEnvConfig()

	// Command line flags can override environment variables
	flagPort := flag.Int("port", cfg.GRPCPort, "gRPC port to listen on (env: GRPC_PORT)")
	flagMaxMemory := flag.Int("max-memory", cfg.MaxMemoryMB, "Max memory in MB (env: MAX_MEMORY_MB)")
	flagMaxKeys := flag.Int("max-keys", cfg.MaxKeys, "Max number of keys (env: MAX_KEYS)")
	flag.Parse()

	// Use flag values (which may be env defaults or CLI overrides)
	cfg.GRPCPort = *flagPort
	cfg.MaxMemoryMB = *flagMaxMemory
	cfg.MaxKeys = *flagMaxKeys

	logger.Info("GRPC port: %d (business gRPC, from %s)", cfg.GRPCPort, envOrDefault(envGRPCPort, "default"))
	logger.Info("Health port: %d (health check, from %s)", cfg.HealthPort, envOrDefault(envHealthPort, "default"))
	logger.Info("Metrics port: %d (Prometheus, from %s)", cfg.MetricsPort, envOrDefault(envMetricsPort, "default"))
	logger.Info("Log level: %s (from %s)", cfg.LogLevel, envOrDefault(envLogLevel, "default"))
	logger.Info("Max memory: %d MB (from %s)", cfg.MaxMemoryMB, envOrDefault(envMaxMemoryMB, "default"))
	logger.Info("Max keys: %d (from %s)", cfg.MaxKeys, envOrDefault(envMaxKeys, "default"))

	// Step 2: Check runtime environment
	logger.Step(2, 4, "Checking runtime environment")
	logger.Info("Go version: %s", runtime.Version())
	logger.Info("OS/Arch: %s/%s", runtime.GOOS, runtime.GOARCH)
	logger.Info("CPU cores: %d", runtime.NumCPU())
	logger.Info("Memory allocator: Go runtime")
	logger.Info("Cache type: In-memory with TTL support")

	// Get pod information if running in Kubernetes
	hostname, _ := os.Hostname()
	if hostname != "" {
		logger.Info("Hostname: %s", hostname)
	}

	podName := os.Getenv(envPodName)
	podNamespace := os.Getenv(envPodNamespace)
	if podName != "" {
		logger.Info("Pod name: %s", podName)
		logger.Info("Pod namespace: %s", podNamespace)
	}

	// Step 3: Create cache node server
	logger.Step(3, 4, "Creating cache node server")
	server := node.NewServer()
	logger.Success("Cache node server instance created")

	// Step 4: Setup graceful shutdown
	logger.Step(4, 4, "Setting up graceful shutdown handler")
	setupGracefulShutdown(logger, server)

	// Start health check server (independent HTTP server for K8s probes)
	go func() {
		if err := server.StartHealthServer(cfg.HealthPort); err != nil {
			logger.Error("Health server error: %v", err)
		}
	}()

	// Start server
	logger.Success("Initialization complete!")
	logger.Info("Starting cache node server on port %d (gRPC)", cfg.GRPCPort)
	logger.Info("Health check server on port %d (HTTP)", cfg.HealthPort)
	logger.Info("Metrics available on port %d (HTTP)", cfg.MetricsPort)
	logger.Info("Press Ctrl+C to stop")
	logger.Info("═══════════════════════════════════════════════════════")

	if err := server.Run(cfg.GRPCPort); err != nil {
		logger.Fatal("Failed to run node server: %v", err)
	}
}

// setupGracefulShutdown registers signal handlers for graceful termination.
func setupGracefulShutdown(logger *utils.Logger, server *node.Server) {
	go func() {
		sigChan := make(chan os.Signal, 1)
		signal.Notify(sigChan, syscall.SIGINT, syscall.SIGTERM)
		sig := <-sigChan

		logger.Warn("Received signal: %v", sig)
		logger.Info("Initiating graceful shutdown...")

		// Stop server gracefully
		if server != nil {
			logger.Info("Stopping node server...")
			if err := server.Stop(); err != nil {
				logger.Error("Failed to stop server: %v", err)
			}
		}

		logger.Success("Cache node server shut down gracefully")
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
║                   Cache Node Service                  ║
║                                                       ║
╚═══════════════════════════════════════════════════════╝
`
	fmt.Println(banner)
	logger.Info("Starting Cache Node Service...")
}
