// Package config implements a two-tier configuration system for Yao-Oracle
// distributed cache microservices with direct Kubernetes API integration.
//
// # Configuration Architecture
//
// Yao-Oracle uses a modern, cloud-native two-tier configuration strategy:
//
// **1. Environment Variables** (Infrastructure/Static Config - Layer 1)
//   - Read at service startup from container environment
//   - Used for: Port bindings, log levels, Kubernetes resource names
//   - Examples: GRPC_PORT, HTTP_PORT, LOG_LEVEL, NAMESPACE, SECRET_NAME
//   - Static during service lifetime (requires restart to change)
//
// **2. Kubernetes API Direct Access** (Business/Dynamic Config - Layer 2)
//   - Services connect directly to Kubernetes API using InClusterConfig()
//   - Read from Kubernetes Secret (NOT mounted as files)
//   - Contains: Namespace definitions, API keys, dashboard credentials
//   - Dynamically reloadable without restart via Kubernetes Informer
//   - No file system involvement - pure API-based approach
//
// # Key Differences from File-Based Approach
//
// This package implements a **file-less configuration system**:
//
//   - ❌ NO Volume/VolumeMount in Kubernetes Deployments
//   - ❌ NO file watching with fsnotify
//   - ❌ NO CONFIG_PATH environment variable
//   - ❌ NO /etc/yao-oracle mount path
//   - ✅ Direct Kubernetes API access with InClusterConfig()
//   - ✅ Kubernetes Informer for instant change detection
//   - ✅ ServiceAccount + RBAC for API access control
//   - ✅ Simpler, more reliable, truly cloud-native
//
// # Configuration Loading
//
// ## Proxy Service Configuration
//
// Proxy service needs dynamic namespace and API key configuration:
//
//	// Load infrastructure config from environment variables
//	infraCfg := config.LoadInfrastructureConfig()
//	logger.SetLevel(infraCfg.LogLevel)
//
//	// Create Kubernetes config loader (no file I/O)
//	loader, err := config.NewK8sConfigLoader()
//	if err != nil {
//	    log.Fatal("Failed to create config loader:", err)
//	}
//
//	// Load configuration directly from Kubernetes API
//	ctx := context.Background()
//	cfg, err := loader.LoadProxyConfig(ctx, infraCfg.Namespace, infraCfg.SecretName)
//	if err != nil {
//	    log.Fatal("Failed to load proxy config:", err)
//	}
//
//	// Start Kubernetes Informer for hot reload
//	informer, err := config.NewK8sInformer(config.K8sInformerConfig{
//	    Namespace:  infraCfg.Namespace,
//	    SecretName: infraCfg.SecretName,
//	})
//	if err != nil {
//	    log.Fatal("Failed to create informer:", err)
//	}
//
//	err = informer.Start(ctx, func(kind string, data map[string][]byte) {
//	    log.Printf("Configuration updated: %s", kind)
//	    // Reload and apply new configuration
//	    newCfg := informer.GetConfig()
//	    server.UpdateNamespaces(newCfg.Proxy.Namespaces)
//	})
//
// ## Cache Node Service Configuration
//
// Nodes are stateless and only need environment variables (NO Kubernetes API access):
//
//	// Cache nodes don't need business config
//	nodeCfg := config.LoadNodeConfig()
//
//	// Start server with static config
//	server := node.NewServer(nodeCfg)
//	server.Run()
//
// Key points for Cache Node:
//   - NO K8sConfigLoader (no Kubernetes API access needed)
//   - NO Informer (stateless design)
//   - NO ServiceAccount/RBAC (no API permissions needed)
//   - Pure environment variable configuration
//
// ## Dashboard Service Configuration
//
// Dashboard needs both namespace info and credentials:
//
//	// Load infrastructure config
//	infraCfg := config.LoadInfrastructureConfig()
//
//	// Create Kubernetes config loader
//	loader, err := config.NewK8sConfigLoader()
//	if err != nil {
//	    log.Fatal("Failed to create config loader:", err)
//	}
//
//	// Load dashboard configuration from Secret
//	ctx := context.Background()
//	cfg, err := loader.LoadDashboardConfig(ctx, infraCfg.Namespace, infraCfg.SecretName)
//	if err != nil {
//	    log.Fatal("Failed to load dashboard config:", err)
//	}
//
//	// Start Kubernetes Informer for hot reload
//	informer, err := config.NewK8sInformer(config.K8sInformerConfig{
//	    Namespace:  infraCfg.Namespace,
//	    SecretName: infraCfg.SecretName,
//	})
//
// # Secret Structure
//
// Services read from a single Kubernetes Secret with this structure:
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
//	            "description": "Gaming application",
//	            "maxMemoryMB": 512,
//	            "defaultTTL": 60,
//	            "rateLimitQPS": 100
//	          }
//	        ]
//	      },
//	      "dashboard": {
//	        "password": "super-secure-password",
//	        "jwtSecret": "jwt-signing-key",
//	        "refreshInterval": 5
//	      }
//	    }
//
// # Hot Reload with Kubernetes Informer
//
// When Secret is updated in Kubernetes:
//  1. Administrator runs `kubectl apply` or `helm upgrade`
//  2. Kubernetes API server updates the Secret
//  3. Informer **immediately** detects the change (no file system delay)
//  4. UpdateFunc callback is triggered
//  5. Service extracts and validates new config
//  6. Service applies new config without restart:
//     - Proxy: Updates namespace list and API keys in memory
//     - Dashboard: Updates password and namespace info (invalidates old tokens)
//     - Nodes: Not affected (stateless)
//
// # RBAC Requirements
//
// Services using K8sConfigLoader and Informer need RBAC permissions:
//
//	# ServiceAccount
//	apiVersion: v1
//	kind: ServiceAccount
//	metadata:
//	  name: yao-oracle
//	  namespace: yao-system
//
//	---
//	# Role
//	apiVersion: rbac.authorization.k8s.io/v1
//	kind: Role
//	metadata:
//	  name: yao-oracle-config-reader
//	  namespace: yao-system
//	rules:
//	- apiGroups: [""]
//	  resources: ["secrets"]
//	  verbs: ["get", "list", "watch"]
//	- apiGroups: [""]
//	  resources: ["endpoints"]
//	  verbs: ["get", "list", "watch"]  # For service discovery
//
//	---
//	# RoleBinding
//	apiVersion: rbac.authorization.k8s.io/v1
//	kind: RoleBinding
//	metadata:
//	  name: yao-oracle-config-reader
//	  namespace: yao-system
//	roleRef:
//	  apiGroup: rbac.authorization.k8s.io
//	  kind: Role
//	  name: yao-oracle-config-reader
//	subjects:
//	- kind: ServiceAccount
//	  name: yao-oracle
//	  namespace: yao-system
//
// And Deployment must reference the ServiceAccount:
//
//	spec:
//	  template:
//	    spec:
//	      serviceAccountName: yao-oracle
//
// # Configuration Validation
//
// All configuration loading includes automatic validation:
//   - At least one namespace must be defined (for proxy)
//   - Namespace names must be unique and non-empty
//   - API keys must be non-empty for each namespace
//   - Dashboard password must be at least 8 characters
//   - Resource limits must be non-negative
//
// Invalid configuration is rejected before being applied:
//
//	cfg, err := loader.LoadProxyConfig(ctx, namespace, secretName)
//	if err != nil {
//	    // Configuration validation failed
//	    log.Fatal("Invalid configuration:", err)
//	}
//	// cfg is guaranteed to be valid here
//
// # Best Practices
//
// DO:
//   - ✅ Use environment variables for infrastructure config (ports, limits, resource names)
//   - ✅ Use Kubernetes Secret for sensitive data (API keys, passwords, JWT secrets)
//   - ✅ Use Kubernetes API direct access (NOT file mounting)
//   - ✅ Use Kubernetes Informer for hot reload (NOT fsnotify)
//   - ✅ Setup proper RBAC (ServiceAccount + Role + RoleBinding)
//   - ✅ Validate all configuration before applying
//   - ✅ Use RWMutex to protect config during hot reload
//   - ✅ Log configuration changes with timestamps
//
// DON'T:
//   - ❌ Mount ConfigMap/Secret as files (use direct API access)
//   - ❌ Use CONFIG_PATH environment variable (not needed)
//   - ❌ Use fsnotify or file watchers (use Informer)
//   - ❌ Skip configuration validation
//   - ❌ Forget ServiceAccount/RBAC for services using Kubernetes API
//   - ❌ Give Cache Node Kubernetes API access (it doesn't need it)
//   - ❌ Commit production secrets to Git
//   - ❌ Restart services for business config changes (use hot reload)
//
// # Package Structure
//
// This package contains:
//   - config.go      - Configuration structures (Namespace, ProxyConfig, DashboardConfig)
//   - k8s_loader.go  - Kubernetes API loader (InClusterConfig, direct Secret reading)
//   - informer.go    - Kubernetes Informer for hot reload
//   - validator.go   - Configuration validation logic
//   - parser.go      - JSON parsing utilities
//   - env.go         - Environment variable helpers
//   - doc.go         - This documentation file
//
// Files NOT included (removed from old architecture):
//   - ❌ loader.go   - File-based configuration loading (replaced by k8s_loader.go)
//   - ❌ watcher.go  - fsnotify file watching (replaced by informer.go)
package config
