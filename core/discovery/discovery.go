// Package discovery implements Kubernetes-native service discovery using
// Endpoints API for real-time cluster node detection.
//
// This package provides efficient service discovery for Yao-Oracle cluster
// nodes without relying on DNS lookups. It uses Kubernetes Endpoints API
// to discover service instances in real-time.
//
// Key features:
//   - Direct Kubernetes API access (no DNS caching issues)
//   - Real-time endpoint updates via Informer
//   - Support for headless services
//   - Automatic handling of pod additions/removals
//
// Example usage:
//
//	disco, err := discovery.NewK8sServiceDiscovery(discovery.Config{
//	    Namespace:   "yao-system",
//	    ServiceName: "yao-oracle-node",
//	})
//	if err != nil {
//	    log.Fatal(err)
//	}
//
//	// Get current endpoints
//	endpoints := disco.GetEndpoints()
//	for _, ep := range endpoints {
//	    log.Printf("Discovered node: %s", ep)
//	}
package discovery

import (
	"context"
	"fmt"
	"sync"
	"time"

	corev1 "k8s.io/api/core/v1"
	metav1 "k8s.io/apimachinery/pkg/apis/meta/v1"
	"k8s.io/client-go/informers"
	"k8s.io/client-go/kubernetes"
	"k8s.io/client-go/rest"
	"k8s.io/client-go/tools/cache"
	"k8s.io/client-go/tools/clientcmd"
)

// ServiceDiscovery defines the interface for service discovery.
//
// Implementations watch for service endpoint changes and provide
// real-time lists of available service instances.
type ServiceDiscovery interface {
	// Start begins watching for service endpoint changes
	//
	// The onChange callback is called whenever endpoints are added/removed.
	//
	// Parameters:
	//   - ctx: Context for lifecycle management
	//   - onChange: Callback function called when endpoints change
	//
	// Returns:
	//   - error: Error if discovery cannot be started
	Start(ctx context.Context, onChange func(endpoints []string)) error

	// Stop gracefully shuts down the discovery watcher
	Stop()

	// GetEndpoints returns the current list of service endpoints
	//
	// Returns:
	//   - []string: List of endpoints in "IP:PORT" format
	GetEndpoints() []string
}

// K8sServiceDiscovery implements service discovery using Kubernetes Endpoints API.
//
// This implementation uses Kubernetes SharedInformer to watch Endpoints resources.
// It maintains a cache of current endpoints and notifies listeners when changes occur.
//
// Advantages over DNS-based discovery:
//   - No DNS caching issues (immediate updates)
//   - No need to parse SRV records
//   - Efficient watching with Kubernetes Informer
//   - Automatic handling of network partitions
//
// Thread-safety: All methods are safe for concurrent use.
type K8sServiceDiscovery struct {
	// mu protects concurrent access to endpoints
	mu sync.RWMutex

	// endpoints holds the current list of service endpoints
	endpoints []string

	// clientset is the Kubernetes client
	clientset *kubernetes.Clientset

	// namespace is the Kubernetes namespace
	namespace string

	// serviceName is the name of the Service to discover
	serviceName string

	// factory is the SharedInformerFactory
	factory informers.SharedInformerFactory

	// stopCh signals the informer to stop
	stopCh chan struct{}

	// onChange callback function
	onChange func(endpoints []string)
}

// Config holds configuration for Kubernetes service discovery.
type Config struct {
	// Namespace is the Kubernetes namespace where the Service lives
	Namespace string

	// ServiceName is the name of the Service to discover
	// This should typically be a headless service for StatefulSets
	ServiceName string

	// Port is the service port number (optional)
	// If not specified, the first port from endpoints will be used
	Port int

	// KubeconfigPath is the path to kubeconfig file (for out-of-cluster use)
	// Leave empty to use in-cluster config
	KubeconfigPath string
}

// NewK8sServiceDiscovery creates a new Kubernetes service discovery instance.
//
// This function should be called when running inside a Kubernetes pod.
// It uses InClusterConfig() to authenticate with the Kubernetes API server.
//
// Requirements:
//   - Service must have appropriate RBAC permissions to list/watch Endpoints
//   - ServiceAccount must be attached to the Pod
//   - Role/RoleBinding must grant "get", "list", "watch" permissions on Endpoints
//
// Parameters:
//   - cfg: Service discovery configuration
//
// Returns:
//   - *K8sServiceDiscovery: A new discovery instance ready to start
//   - error: Error if Kubernetes client cannot be created
//
// Example:
//
//	disco, err := discovery.NewK8sServiceDiscovery(discovery.Config{
//	    Namespace:   "yao-system",
//	    ServiceName: "yao-oracle-node",
//	    Port:        7070,
//	})
//	if err != nil {
//	    log.Fatal("Failed to create discovery:", err)
//	}
func NewK8sServiceDiscovery(cfg Config) (*K8sServiceDiscovery, error) {
	// Create Kubernetes client
	var config *rest.Config
	var err error

	if cfg.KubeconfigPath != "" {
		// Use kubeconfig file (for local development)
		config, err = clientcmd.BuildConfigFromFlags("", cfg.KubeconfigPath)
		if err != nil {
			return nil, fmt.Errorf("failed to load kubeconfig from %s: %w", cfg.KubeconfigPath, err)
		}
	} else {
		// Use in-cluster config (for production)
		config, err = rest.InClusterConfig()
		if err != nil {
			return nil, fmt.Errorf("failed to load in-cluster config: %w", err)
		}
	}

	clientset, err := kubernetes.NewForConfig(config)
	if err != nil {
		return nil, fmt.Errorf("failed to create Kubernetes clientset: %w", err)
	}

	return &K8sServiceDiscovery{
		clientset:   clientset,
		namespace:   cfg.Namespace,
		serviceName: cfg.ServiceName,
		stopCh:      make(chan struct{}),
		endpoints:   []string{},
	}, nil
}

// Start begins watching for service endpoint changes.
//
// This method creates a SharedInformerFactory and starts watching Endpoints.
// The onChange callback is called whenever endpoints are added or removed.
func (d *K8sServiceDiscovery) Start(ctx context.Context, onChange func(endpoints []string)) error {
	d.onChange = onChange

	// Load initial endpoints
	if err := d.loadInitialEndpoints(ctx); err != nil {
		return fmt.Errorf("failed to load initial endpoints: %w", err)
	}

	// Call onChange with initial endpoints
	if onChange != nil {
		onChange(d.GetEndpoints())
	}

	// Create SharedInformerFactory with namespace filter
	d.factory = informers.NewSharedInformerFactoryWithOptions(
		d.clientset,
		time.Minute,
		informers.WithNamespace(d.namespace),
	)

	// Watch Endpoints resources
	endpointsInformer := d.factory.Core().V1().Endpoints().Informer()
	endpointsInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
		AddFunc: func(obj interface{}) {
			ep := obj.(*corev1.Endpoints)
			if ep.Name == d.serviceName {
				d.handleEndpointsUpdate(ep)
			}
		},
		UpdateFunc: func(oldObj, newObj interface{}) {
			ep := newObj.(*corev1.Endpoints)
			if ep.Name == d.serviceName {
				d.handleEndpointsUpdate(ep)
			}
		},
		DeleteFunc: func(obj interface{}) {
			ep := obj.(*corev1.Endpoints)
			if ep.Name == d.serviceName {
				d.mu.Lock()
				d.endpoints = []string{}
				d.mu.Unlock()

				if d.onChange != nil {
					d.onChange([]string{})
				}
			}
		},
	})

	// Start informers
	d.factory.Start(d.stopCh)

	// Wait for cache sync
	synced := d.factory.WaitForCacheSync(d.stopCh)
	for typ, ok := range synced {
		if !ok {
			return fmt.Errorf("failed to sync cache for %v", typ)
		}
	}

	return nil
}

// Stop gracefully shuts down the discovery watcher.
func (d *K8sServiceDiscovery) Stop() {
	if d.stopCh != nil {
		close(d.stopCh)
		d.stopCh = nil
	}
}

// GetEndpoints returns the current list of service endpoints.
//
// Thread-safe: Safe for concurrent calls.
func (d *K8sServiceDiscovery) GetEndpoints() []string {
	d.mu.RLock()
	defer d.mu.RUnlock()

	// Return a copy to prevent external modifications
	result := make([]string, len(d.endpoints))
	copy(result, d.endpoints)
	return result
}

// loadInitialEndpoints loads the initial list of endpoints.
func (d *K8sServiceDiscovery) loadInitialEndpoints(ctx context.Context) error {
	ep, err := d.clientset.CoreV1().Endpoints(d.namespace).Get(ctx, d.serviceName, metav1.GetOptions{})
	if err != nil {
		return fmt.Errorf("failed to get endpoints: %w", err)
	}

	d.handleEndpointsUpdate(ep)
	return nil
}

// handleEndpointsUpdate processes Endpoints update events.
func (d *K8sServiceDiscovery) handleEndpointsUpdate(ep *corev1.Endpoints) {
	// Extract IP addresses from endpoints
	var newEndpoints []string

	for _, subset := range ep.Subsets {
		// Get port
		port := 0
		if len(subset.Ports) > 0 {
			port = int(subset.Ports[0].Port)
		}

		// Get addresses
		for _, addr := range subset.Addresses {
			if port > 0 {
				newEndpoints = append(newEndpoints, fmt.Sprintf("%s:%d", addr.IP, port))
			} else {
				newEndpoints = append(newEndpoints, addr.IP)
			}
		}
	}

	// Update endpoints atomically
	d.mu.Lock()
	d.endpoints = newEndpoints
	d.mu.Unlock()

	// Call onChange callback
	if d.onChange != nil {
		d.onChange(newEndpoints)
	}
}
