package config

// Namespace represents a business namespace with its API key and resource limits.
//
// Each namespace provides data isolation for different tenants or applications.
// The API key is used for authentication and namespace identification.
type Namespace struct {
	// Name is the unique identifier for the namespace
	// Example: "game-app", "ads-service"
	Name string `json:"name"`

	// APIKey is the secret key used for authentication
	// This is sensitive data and should be stored in Kubernetes Secret
	APIKey string `json:"apikey"`

	// Description provides human-readable information about the namespace
	Description string `json:"description,omitempty"`

	// MaxMemoryMB is the memory limit in megabytes for this namespace
	// Optional: Default value depends on service configuration
	MaxMemoryMB int `json:"maxMemoryMB,omitempty"`

	// DefaultTTL is the default time-to-live in seconds for cache entries
	// Optional: 0 means no expiration
	DefaultTTL int `json:"defaultTTL,omitempty"`

	// RateLimitQPS is the queries-per-second limit for this namespace
	// Optional: 0 means no rate limiting
	RateLimitQPS int `json:"rateLimitQPS,omitempty"`
}

// ProxyConfig holds the proxy service configuration.
//
// This configuration is loaded from Kubernetes Secret and can be hot-reloaded
// without service restart using Kubernetes Informer.
type ProxyConfig struct {
	// Namespaces is the list of configured business namespaces
	// At least one namespace must be defined
	Namespaces []Namespace `json:"namespaces"`

	// Port is deprecated and should be configured via environment variables
	// This field is kept for backward compatibility
	Port int `json:"port,omitempty"`
}

// DashboardConfig holds the dashboard service configuration.
//
// This configuration contains sensitive data (password, JWT secret) and
// must be stored in Kubernetes Secret, not ConfigMap.
type DashboardConfig struct {
	// Password is the authentication password for dashboard access
	// Must be at least 8 characters long
	Password string `json:"password"`

	// JWTSecret is the secret key used for JWT token signing
	// Should be a long, random string for security
	JWTSecret string `json:"jwtSecret,omitempty"`

	// RefreshInterval is the dashboard auto-refresh interval in seconds
	// Default: 5 seconds
	RefreshInterval int `json:"refreshInterval,omitempty"`

	// Theme is the dashboard UI theme ("light" or "dark")
	Theme string `json:"theme,omitempty"`
}

// Config holds all configuration including proxy and dashboard settings.
//
// This is the root configuration structure that is loaded from Kubernetes
// Secret. The Secret should contain a single key "config-with-secrets.json"
// that contains the full configuration as JSON.
//
// Example Secret structure:
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
//	        "namespaces": [
//	          {
//	            "name": "game-app",
//	            "apikey": "game-secret-123",
//	            "description": "Gaming application"
//	          }
//	        ]
//	      },
//	      "dashboard": {
//	        "password": "super-secure-password",
//	        "jwtSecret": "jwt-signing-key"
//	      }
//	    }
type Config struct {
	// Proxy configuration (required for proxy service)
	Proxy *ProxyConfig `json:"proxy,omitempty"`

	// Dashboard configuration (required for dashboard service)
	Dashboard *DashboardConfig `json:"dashboard,omitempty"`
}

// GetNamespaceByAPIKey returns the namespace for the given API key.
//
// This is a helper method that searches through all configured namespaces
// to find one with a matching API key.
//
// Parameters:
//   - apiKey: The API key to look up
//
// Returns:
//   - namespace: The Namespace object if found
//   - ok: True if API key was found, false otherwise
//
// Example:
//
//	ns, ok := config.GetNamespaceByAPIKey("game-secret-123")
//	if !ok {
//	    return errors.New("invalid API key")
//	}
//	log.Printf("Authenticated as namespace: %s", ns.Name)
func (c *Config) GetNamespaceByAPIKey(apiKey string) (*Namespace, bool) {
	if c.Proxy == nil {
		return nil, false
	}

	for i := range c.Proxy.Namespaces {
		if c.Proxy.Namespaces[i].APIKey == apiKey {
			return &c.Proxy.Namespaces[i], true
		}
	}
	return nil, false
}

// GetNamespaceByName returns the namespace with the given name.
//
// Parameters:
//   - name: The namespace name to look up
//
// Returns:
//   - namespace: The Namespace object if found
//   - ok: True if namespace was found, false otherwise
func (c *Config) GetNamespaceByName(name string) (*Namespace, bool) {
	if c.Proxy == nil {
		return nil, false
	}

	for i := range c.Proxy.Namespaces {
		if c.Proxy.Namespaces[i].Name == name {
			return &c.Proxy.Namespaces[i], true
		}
	}
	return nil, false
}
