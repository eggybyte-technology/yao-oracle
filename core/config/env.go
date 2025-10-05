package config

import (
	"os"
	"strconv"
)

// Standard environment variable keys used across all services.
//
// These constants define the infrastructure configuration that is loaded
// from environment variables at service startup. Changes to these values
// require a Pod restart.
//
// NOTE: Configuration files are NOT used. Services read configuration
// directly from Kubernetes API using the Secret specified by EnvSecretName.
const (
	// Infrastructure Configuration (Layer 1)
	EnvGRPCPort    = "GRPC_PORT"
	EnvHTTPPort    = "HTTP_PORT"
	EnvLogLevel    = "LOG_LEVEL"
	EnvMetricsPort = "METRICS_PORT"

	// Kubernetes Resource Names (for direct API access)
	EnvNamespace     = "NAMESPACE"      // Kubernetes namespace
	EnvSecretName    = "SECRET_NAME"    // Name of Secret to read config from
	EnvConfigMapName = "CONFIGMAP_NAME" // Name of ConfigMap (optional)

	// Service Discovery Configuration
	EnvProxyHeadlessService = "PROXY_HEADLESS_SERVICE" // Proxy headless service DNS
	EnvNodeHeadlessService  = "NODE_HEADLESS_SERVICE"  // Node headless service DNS
	EnvDiscoveryMode        = "DISCOVERY_MODE"         // Discovery mode: "k8s" or "dns"
	EnvDiscoveryInterval    = "DISCOVERY_INTERVAL"     // Discovery refresh interval in seconds

	// Cache Node specific configuration
	EnvMaxMemoryMB    = "MAX_MEMORY_MB"
	EnvMaxKeys        = "MAX_KEYS"
	EnvEvictionPolicy = "EVICTION_POLICY" // Eviction policy: "LRU", "LFU", etc.

	// Pod metadata (auto-injected by Kubernetes Downward API)
	EnvPodName = "POD_NAME"
	EnvPodIP   = "POD_IP"
)

// GetEnv retrieves an environment variable value with a default fallback.
//
// This is the primary function for reading string configuration values.
//
// Parameters:
//   - key: Environment variable name
//   - defaultValue: Value to return if environment variable is not set or empty
//
// Returns:
//   - string: The environment variable value or default value
//
// Example:
//
//	logLevel := config.GetEnv(config.EnvLogLevel, "info")
//	configPath := config.GetEnv(config.EnvConfigPath, "/etc/yao-oracle/config.json")
func GetEnv(key, defaultValue string) string {
	if value := os.Getenv(key); value != "" {
		return value
	}
	return defaultValue
}

// GetEnvInt retrieves an integer environment variable with a default fallback.
//
// If the environment variable is not set, empty, or cannot be parsed as an integer,
// the default value is returned instead. Negative values are considered invalid
// and will trigger the default value.
//
// Parameters:
//   - key: Environment variable name
//   - defaultValue: Value to return if parsing fails or value is invalid
//
// Returns:
//   - int: The parsed integer value or default value
//
// Example:
//
//	grpcPort := config.GetEnvInt(config.EnvGRPCPort, 8080)
//	maxMemory := config.GetEnvInt(config.EnvMaxMemoryMB, 512)
func GetEnvInt(key string, defaultValue int) int {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	value, err := strconv.Atoi(valueStr)
	if err != nil || value <= 0 {
		return defaultValue
	}

	return value
}

// GetEnvBool retrieves a boolean environment variable with a default fallback.
//
// Recognizes "true", "1", "yes" as true (case-insensitive).
// All other values (including empty) are considered false.
//
// Parameters:
//   - key: Environment variable name
//   - defaultValue: Value to return if environment variable is not set or empty
//
// Returns:
//   - bool: The parsed boolean value or default value
//
// Example:
//
//	authEnabled := config.GetEnvBool(config.EnvDashboardAuthEnabled, true)
//	watchEnabled := config.GetEnvBool(config.EnvConfigWatchEnabled, false)
func GetEnvBool(key string, defaultValue bool) bool {
	valueStr := os.Getenv(key)
	if valueStr == "" {
		return defaultValue
	}

	// Recognize common truthy values
	switch valueStr {
	case "true", "1", "yes", "True", "TRUE", "Yes", "YES":
		return true
	case "false", "0", "no", "False", "FALSE", "No", "NO":
		return false
	default:
		return defaultValue
	}
}

// EnvSource returns a string describing the source of a configuration value.
//
// This is useful for logging to show where a configuration value came from.
//
// Parameters:
//   - key: Environment variable name to check
//   - defaultValue: Name of the default source
//
// Returns:
//   - string: Description of the value source (e.g., "env:GRPC_PORT" or "default")
//
// Example:
//
//	logger.Info("GRPC port: %d (from %s)", port, config.EnvSource(config.EnvGRPCPort, "default"))
func EnvSource(key, defaultValue string) string {
	if os.Getenv(key) != "" {
		return "env:" + key
	}
	return defaultValue
}

// InfrastructureConfig holds common infrastructure configuration for all services.
//
// This struct encapsulates the Layer 1 (infrastructure) configuration that is
// loaded from environment variables at service startup.
//
// All fields have sensible defaults and can be overridden via environment variables.
//
// NOTE: No ConfigPath field - services read configuration directly from
// Kubernetes API using the Secret/ConfigMap names specified here.
type InfrastructureConfig struct {
	// GRPCPort is the port for gRPC server to listen on
	GRPCPort int

	// HTTPPort is the port for HTTP server to listen on (dashboard only)
	HTTPPort int

	// MetricsPort is the port for Prometheus metrics endpoint
	MetricsPort int

	// LogLevel controls logging verbosity (debug, info, warn, error)
	LogLevel string

	// Namespace is the Kubernetes namespace
	Namespace string

	// SecretName is the name of the Secret to read configuration from
	SecretName string

	// ConfigMapName is the name of the ConfigMap to read from (optional)
	ConfigMapName string
}

// LoadInfrastructureConfig loads infrastructure configuration from environment variables.
//
// This function reads Layer 1 configuration (infrastructure settings) from
// environment variables with sensible defaults for all values.
//
// Default values:
//   - GRPCPort: 8080
//   - HTTPPort: 8080
//   - MetricsPort: 9090
//   - LogLevel: "info"
//   - Namespace: "default"
//   - SecretName: "yao-oracle-secret"
//   - ConfigMapName: "yao-oracle-config"
//
// Returns:
//   - InfrastructureConfig: Populated configuration struct
//
// Example:
//
//	cfg := config.LoadInfrastructureConfig()
//	logger.Info("Starting server on port %d", cfg.GRPCPort)
//	logger.Info("Reading config from Secret: %s/%s", cfg.Namespace, cfg.SecretName)
func LoadInfrastructureConfig() InfrastructureConfig {
	return InfrastructureConfig{
		GRPCPort:      GetEnvInt(EnvGRPCPort, 8080),
		HTTPPort:      GetEnvInt(EnvHTTPPort, 8080),
		MetricsPort:   GetEnvInt(EnvMetricsPort, 9090),
		LogLevel:      GetEnv(EnvLogLevel, "info"),
		Namespace:     GetEnv(EnvNamespace, "default"),
		SecretName:    GetEnv(EnvSecretName, "yao-oracle-secret"),
		ConfigMapName: GetEnv(EnvConfigMapName, "yao-oracle-config"),
	}
}

// NodeConfig holds cache node specific configuration.
//
// Cache nodes are stateless and only read configuration from environment variables.
// They do not use ConfigMap or Secret files.
type NodeConfig struct {
	InfrastructureConfig

	// MaxMemoryMB is the maximum memory usage in megabytes
	MaxMemoryMB int

	// MaxKeys is the maximum number of keys the node can store
	MaxKeys int
}

// LoadNodeConfig loads cache node configuration from environment variables.
//
// Default values:
//   - MaxMemoryMB: 512
//   - MaxKeys: 100000
//
// Returns:
//   - NodeConfig: Populated node configuration struct
//
// Example:
//
//	cfg := config.LoadNodeConfig()
//	logger.Info("Max memory: %d MB", cfg.MaxMemoryMB)
//	logger.Info("Max keys: %d", cfg.MaxKeys)
func LoadNodeConfig() NodeConfig {
	return NodeConfig{
		InfrastructureConfig: LoadInfrastructureConfig(),
		MaxMemoryMB:          GetEnvInt(EnvMaxMemoryMB, 512),
		MaxKeys:              GetEnvInt(EnvMaxKeys, 100000),
	}
}

// PodMetadata holds Kubernetes pod metadata.
//
// These values are automatically injected by Kubernetes via the Downward API
// and are used for service discovery and logging.
type PodMetadata struct {
	// Name is the pod name
	Name string

	// Namespace is the Kubernetes namespace
	Namespace string

	// IP is the pod IP address
	IP string
}

// LoadPodMetadata loads Kubernetes pod metadata from environment variables.
//
// These environment variables are typically injected by Kubernetes using the
// Downward API via the deployment YAML.
//
// Returns:
//   - PodMetadata: Populated pod metadata struct
//
// Example:
//
//	metadata := config.LoadPodMetadata()
//	if metadata.Name != "" {
//	    logger.Info("Running in pod: %s", metadata.Name)
//	}
func LoadPodMetadata() PodMetadata {
	return PodMetadata{
		Name:      os.Getenv(EnvPodName),
		Namespace: os.Getenv(EnvNamespace),
		IP:        os.Getenv(EnvPodIP),
	}
}
