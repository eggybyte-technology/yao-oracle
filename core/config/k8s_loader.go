package config

import (
	"context"
	"encoding/json"
	"fmt"

	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/clientcmd"
)

// K8sConfigLoader loads configuration directly from Kubernetes API.
//
// This interface defines methods for loading configuration from Kubernetes
// Secrets. It does NOT involve file mounting - all configuration is read
// directly via Kubernetes API calls.
//
// Implementations must:
//   - Use InClusterConfig() when running inside Kubernetes
//   - Support out-of-cluster config for local development
//   - Read from Kubernetes Secret (not ConfigMap for sensitive data)
//   - Parse JSON configuration from Secret data
type K8sConfigLoader interface {
	// LoadProxyConfig reads Secret from Kubernetes API and returns ProxyConfig
	//
	// Parameters:
	//   - ctx: Context for the API call
	//   - namespace: Kubernetes namespace where Secret lives
	//   - secretName: Name of the Secret resource
	//
	// Returns:
	//   - *ProxyConfig: Parsed proxy configuration
	//   - error: Error if Secret cannot be read or parsed
	LoadProxyConfig(ctx context.Context, namespace, secretName string) (*ProxyConfig, error)

	// LoadDashboardConfig reads Secret from Kubernetes API and returns DashboardConfig
	//
	// Parameters:
	//   - ctx: Context for the API call
	//   - namespace: Kubernetes namespace where Secret lives
	//   - secretName: Name of the Secret resource
	//
	// Returns:
	//   - *DashboardConfig: Parsed dashboard configuration
	//   - error: Error if Secret cannot be read or parsed
	LoadDashboardConfig(ctx context.Context, namespace, secretName string) (*DashboardConfig, error)

	// LoadFullConfig reads Secret from Kubernetes API and returns full Config
	//
	// This method loads both proxy and dashboard configuration from a single Secret.
	//
	// Parameters:
	//   - ctx: Context for the API call
	//   - namespace: Kubernetes namespace where Secret lives
	//   - secretName: Name of the Secret resource
	//
	// Returns:
	//   - *Config: Parsed full configuration
	//   - error: Error if Secret cannot be read or parsed
	LoadFullConfig(ctx context.Context, namespace, secretName string) (*Config, error)
}

// K8sConfigLoaderImpl is the default implementation of K8sConfigLoader.
//
// It connects directly to Kubernetes API and reads Secret data without any
// file system involvement. The Secret must contain a key "config-with-secrets.json"
// with the full configuration as JSON.
//
// Thread-safety: Safe for concurrent use after initialization.
type K8sConfigLoaderImpl struct {
	clientset *kubernetes.Clientset
}

// NewK8sConfigLoader creates a new Kubernetes config loader using in-cluster config.
//
// This function should be called when running inside a Kubernetes pod. It uses
// InClusterConfig() to authenticate with the Kubernetes API server.
//
// Requirements:
//   - Service must have appropriate RBAC permissions to read Secrets
//   - ServiceAccount must be attached to the Pod
//   - Role/RoleBinding must grant "get" permission on Secrets
//
// Returns:
//   - *K8sConfigLoaderImpl: A new loader instance ready to use
//   - error: Error if in-cluster config cannot be loaded or client creation fails
//
// Example:
//
//	loader, err := config.NewK8sConfigLoader()
//	if err != nil {
//	    log.Fatal("Failed to create config loader:", err)
//	}
//
//	cfg, err := loader.LoadFullConfig(ctx, "yao-system", "yao-oracle-secret")
//	if err != nil {
//	    log.Fatal("Failed to load config:", err)
//	}
func NewK8sConfigLoader() (*K8sConfigLoaderImpl, error) {
	cfg, err := rest.InClusterConfig()
	if err != nil {
		return nil, fmt.Errorf("failed to get in-cluster config: %w", err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return &K8sConfigLoaderImpl{clientset: clientset}, nil
}

// NewK8sConfigLoaderFromKubeconfig creates a loader using a kubeconfig file.
//
// This function should be used for local development and testing outside
// of a Kubernetes cluster. It loads credentials from a kubeconfig file.
//
// Parameters:
//   - kubeconfigPath: Path to kubeconfig file (e.g., "~/.kube/config")
//
// Returns:
//   - *K8sConfigLoaderImpl: A new loader instance ready to use
//   - error: Error if kubeconfig cannot be loaded or client creation fails
//
// Example:
//
//	loader, err := config.NewK8sConfigLoaderFromKubeconfig("~/.kube/config")
//	if err != nil {
//	    log.Fatal("Failed to create config loader:", err)
//	}
func NewK8sConfigLoaderFromKubeconfig(kubeconfigPath string) (*K8sConfigLoaderImpl, error) {
	cfg, err := clientcmd.BuildConfigFromFlags("", kubeconfigPath)
	if err != nil {
		return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", kubeconfigPath, err)
	}

	clientset, err := kubernetes.NewForConfig(cfg)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return &K8sConfigLoaderImpl{clientset: clientset}, nil
}

// LoadProxyConfig loads proxy configuration from Kubernetes Secret.
//
// The Secret must contain a key "config-with-secrets.json" with JSON data
// that includes a "proxy" section.
func (l *K8sConfigLoaderImpl) LoadProxyConfig(ctx context.Context, namespace, secretName string) (*ProxyConfig, error) {
	fullConfig, err := l.LoadFullConfig(ctx, namespace, secretName)
	if err != nil {
		return nil, err
	}

	if fullConfig.Proxy == nil {
		return nil, fmt.Errorf("proxy configuration not found in Secret %s/%s", namespace, secretName)
	}

	return fullConfig.Proxy, nil
}

// LoadDashboardConfig loads dashboard configuration from Kubernetes Secret.
//
// The Secret must contain a key "config-with-secrets.json" with JSON data
// that includes a "dashboard" section.
func (l *K8sConfigLoaderImpl) LoadDashboardConfig(ctx context.Context, namespace, secretName string) (*DashboardConfig, error) {
	fullConfig, err := l.LoadFullConfig(ctx, namespace, secretName)
	if err != nil {
		return nil, err
	}

	if fullConfig.Dashboard == nil {
		return nil, fmt.Errorf("dashboard configuration not found in Secret %s/%s", namespace, secretName)
	}

	return fullConfig.Dashboard, nil
}

// LoadFullConfig loads the complete configuration from Kubernetes Secret.
//
// Expected Secret structure:
//
//	apiVersion: v1
//	kind: Secret
//	metadata:
//	  name: yao-oracle-secret
//	  namespace: yao-system
//	type: Opaque
//	stringData:
//	  config-with-secrets.json: |
//	    {
//	      "proxy": {
//	        "namespaces": [...]
//	      },
//	      "dashboard": {
//	        "password": "...",
//	        "jwtSecret": "..."
//	      }
//	    }
//
// Side effects:
//   - Makes API call to Kubernetes API server
//   - Validates configuration structure before returning
func (l *K8sConfigLoaderImpl) LoadFullConfig(ctx context.Context, namespace, secretName string) (*Config, error) {
	// Read Secret from Kubernetes API
	secret, err := l.clientset.CoreV1().Secrets(namespace).Get(ctx, secretName, metav1.GetOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to get Secret %s/%s: %w", namespace, secretName, err)
	}

	// Parse config-with-secrets.json key
	configJSON, ok := secret.Data["config-with-secrets.json"]
	if !ok {
		return nil, fmt.Errorf("key 'config-with-secrets.json' not found in Secret %s/%s", namespace, secretName)
	}

	// Unmarshal JSON
	var cfg Config
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		return nil, fmt.Errorf("failed to parse configuration JSON: %w", err)
	}

	// Validate configuration before returning
	if err := ValidateConfig(&cfg); err != nil {
		return nil, fmt.Errorf("invalid configuration: %w", err)
	}

	return &cfg, nil
}
