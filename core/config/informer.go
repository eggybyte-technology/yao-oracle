package config

import (
	"context"
	"encoding/json"
	"fmt"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"

	"github.com/eggybyte-technology/yao-oracle/core/utils"
)

// DynamicConfigWatcher watches for configuration changes using Kubernetes Informer.
//
// This interface defines the contract for hot-reloading configuration without
// service restart. It uses Kubernetes Informer API to watch Secret resources
// for changes and triggers callbacks when updates are detected.
//
// Key features:
//   - Instant change detection (no file system delay)
//   - Efficient watching using Kubernetes Informer
//   - Automatic reconnection on errors
//   - Thread-safe configuration access
type DynamicConfigWatcher interface {
	// Start begins watching ConfigMap/Secret for changes
	//
	// The onChange callback is called whenever the watched resource is updated.
	// Start should be called at most once per watcher instance.
	//
	// Parameters:
	//   - ctx: Context for lifecycle management
	//   - onChange: Callback function called with resource kind and data
	//
	// Returns:
	//   - error: Error if watcher cannot be started
	Start(ctx context.Context, onChange func(kind string, data map[string][]byte)) error

	// Stop gracefully shuts down the watcher
	//
	// This method is safe to call multiple times.
	Stop()

	// GetConfig returns the current cached configuration
	//
	// This method is thread-safe and returns the last successfully loaded config.
	//
	// Returns:
	//   - Config: Current configuration
	GetConfig() Config
}

// K8sInformer watches Kubernetes Secret for configuration changes.
//
// It uses Kubernetes SharedInformer for efficient watching and caching.
// The informer watches a specific Secret in a namespace and calls the
// onChange callback when the Secret is updated.
//
// Advantages over file watching:
//   - No ~60s delay waiting for Kubernetes to update mounted files
//   - Instant notification when Secret is updated
//   - No symlink complexity
//   - More reliable and cloud-native
//
// Thread-safety: All methods are safe for concurrent use.
type K8sInformer struct {
	// mu protects concurrent access to config
	mu sync.RWMutex

	// config holds the currently loaded configuration
	config Config

	// clientset is the Kubernetes client
	clientset *kubernetes.Clientset

	// namespace is the Kubernetes namespace
	namespace string

	// secretName is the name of the Secret to watch
	secretName string

	// factory is the SharedInformerFactory
	factory informers.SharedInformerFactory

	// stopCh signals the informer to stop
	stopCh chan struct{}

	// logger for configuration loading events
	logger *utils.Logger

	// onChange callback function
	onChange func(kind string, data map[string][]byte)
}

// K8sInformerConfig holds configuration for creating a Kubernetes informer.
type K8sInformerConfig struct {
	// Namespace is the Kubernetes namespace where the Secret lives
	Namespace string

	// SecretName is the name of the Secret to watch
	SecretName string

	// KubeconfigPath is the path to kubeconfig file (for out-of-cluster use)
	// Leave empty to use in-cluster config
	KubeconfigPath string
}

// NewK8sInformer creates a new Kubernetes Secret informer.
//
// This function should be called when running inside a Kubernetes pod.
// It uses InClusterConfig() to authenticate with the Kubernetes API server.
//
// Requirements:
//   - Service must have appropriate RBAC permissions to watch Secrets
//   - ServiceAccount must be attached to the Pod
//   - Role/RoleBinding must grant "get", "list", "watch" permissions on Secrets
//
// Parameters:
//   - cfg: Informer configuration
//
// Returns:
//   - *K8sInformer: A new informer instance ready to start
//   - error: Error if Kubernetes client cannot be created
//
// Example:
//
//	informer, err := config.NewK8sInformer(config.K8sInformerConfig{
//	    Namespace:  "yao-system",
//	    SecretName: "yao-oracle-secret",
//	})
//	if err != nil {
//	    log.Fatal("Failed to create informer:", err)
//	}
//
//	err = informer.Start(ctx, func(kind string, data map[string][]byte) {
//	    log.Printf("Configuration updated: %s", kind)
//	    // Reload configuration
//	})
func NewK8sInformer(cfg K8sInformerConfig) (*K8sInformer, error) {
	logger := utils.NewLogger("k8s-informer")

	// Create Kubernetes client
	var config *rest.Config
	var err error

	if cfg.KubeconfigPath != "" {
		// Use kubeconfig file (for local development)
		config, err = clientcmd.BuildConfigFromFlags("", cfg.KubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", cfg.KubeconfigPath, err)
		}
		logger.Info("Using kubeconfig: %s", cfg.KubeconfigPath)
	} else {
		// Use in-cluster config (for production)
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
		}
		logger.Info("Using in-cluster Kubernetes configuration")
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return &K8sInformer{
		clientset:  clientset,
		namespace:  cfg.Namespace,
		secretName: cfg.SecretName,
		stopCh:     make(chan struct{}),
		logger:     logger,
	}, nil
}

// Start begins watching the Secret for changes.
//
// This method creates a SharedInformerFactory and starts watching the Secret.
// The onChange callback is called whenever the Secret is updated.
//
// The informer automatically handles:
//   - Initial configuration load
//   - Watching for updates
//   - Cache synchronization
//   - Reconnection on errors
//
// Side effects:
//   - Loads initial configuration from Secret
//   - Starts background goroutines
//   - Calls onChange immediately with initial config
func (i *K8sInformer) Start(ctx context.Context, onChange func(kind string, data map[string][]byte)) error {
	i.onChange = onChange

	// Load initial configuration
	if err := i.loadInitialConfig(ctx); err != nil {
		i.logger.Error("Failed to load initial configuration: %v", err)
		return err
	}

	// Call onChange with initial config
	if onChange != nil {
		i.mu.RLock()
		configJSON, _ := json.Marshal(i.config)
		data := map[string][]byte{
			"config-with-secrets.json": configJSON,
		}
		i.mu.RUnlock()
		onChange("Secret", data)
	}

	// Create SharedInformerFactory with namespace filter
	i.factory = informers.NewSharedInformerFactoryWithOptions(
		i.clientset,
		time.Minute,
		informers.WithNamespace(i.namespace),
	)

	// Watch Secret resources
	secretInformer := i.factory.Core().V1().Secrets().Informer()
	secretInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		UpdateFunc: func(oldObj, newObj interface{}) {
			secret := newObj.(*corev1.Secret)
			if secret.Name == i.secretName {
				i.logger.Info("🔑 Secret %s updated, reloading configuration...", i.secretName)
				i.handleSecretUpdate(secret)
			}
		},
	})

	// Start informers
	i.factory.Start(i.stopCh)

	// Wait for cache sync
	synced := i.factory.WaitForCacheSync(i.stopCh)
	for typ, ok := range synced {
		if !ok {
			return fmt.Errorf("failed to sync cache for %v", typ)
		}
	}

	i.logger.Info("✅ Kubernetes Informer started, watching Secret: %s/%s", i.namespace, i.secretName)
	return nil
}

// Stop gracefully shuts down the informer.
//
// This method is safe to call multiple times.
func (i *K8sInformer) Stop() {
	if i.stopCh != nil {
		close(i.stopCh)
		i.stopCh = nil
		i.logger.Info("🛑 Stopped Kubernetes Informer")
	}
}

// GetConfig returns the current cached configuration.
//
// This method is thread-safe and does not perform I/O.
func (i *K8sInformer) GetConfig() Config {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.config
}

// loadInitialConfig loads the initial configuration from Secret.
func (i *K8sInformer) loadInitialConfig(ctx context.Context) error {
	loader, err := NewK8sConfigLoader()
	if err != nil {
		return fmt.Errorf("failed to create config loader: %w", err)
	}

	cfg, err := loader.LoadFullConfig(ctx, i.namespace, i.secretName)
	if err != nil {
		return fmt.Errorf("failed to load config: %w", err)
	}

	i.mu.Lock()
	i.config = *cfg
	i.mu.Unlock()

	i.logger.Info("Initial configuration loaded from Secret %s/%s", i.namespace, i.secretName)
	return nil
}

// handleSecretUpdate processes Secret update events.
func (i *K8sInformer) handleSecretUpdate(secret *corev1.Secret) {
	// Parse configuration from Secret
	configJSON, ok := secret.Data["config-with-secrets.json"]
	if !ok {
		i.logger.Error("Key 'config-with-secrets.json' not found in Secret")
		return
	}

	var cfg Config
	if err := json.Unmarshal(configJSON, &cfg); err != nil {
		i.logger.Error("Failed to parse configuration: %v", err)
		return
	}

	// Validate configuration before applying
	if err := ValidateConfig(&cfg); err != nil {
		i.logger.Error("Invalid configuration, keeping old config: %v", err)
		return
	}

	// Apply new configuration atomically
	i.mu.Lock()
	i.config = cfg
	i.mu.Unlock()

	// Call onChange callback
	if i.onChange != nil {
		data := make(map[string][]byte)
		for k, v := range secret.Data {
			data[k] = v
		}
		i.onChange("Secret", data)
	}

	i.logger.Info("✅ Configuration updated from Secret at %s", time.Now().Format(time.RFC3339))
}

// GetNamespaceByAPIKey is a convenience method for API key authentication.
//
// This method is thread-safe.
func (i *K8sInformer) GetNamespaceByAPIKey(apiKey string) (*Namespace, bool) {
	i.mu.RLock()
	defer i.mu.RUnlock()
	return i.config.GetNamespaceByAPIKey(apiKey)
}
