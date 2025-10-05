# Yao-Oracle Helm Chart Configuration Guide

## Overview

This guide explains how to configure and manage the Yao-Oracle distributed KV cache system using Helm.

## Key Features

### 1. Dynamic ConfigMap Monitoring

The Proxy service can dynamically watch Kubernetes ConfigMaps for configuration changes. When the ConfigMap is updated, the Proxy automatically reloads the configuration without requiring a pod restart.

**Configuration:**
```yaml
proxy:
  configWatch:
    enabled: true              # Enable ConfigMap watching
    configMapName: ""          # Auto-generated if empty
    reloadInterval: 10         # Reload interval in seconds
```

**How it works:**
1. Proxy service watches the specified ConfigMap using Kubernetes API
2. When changes are detected, configuration is automatically reloaded
3. New API keys, namespaces, and node addresses take effect immediately
4. No downtime required for configuration updates

### 2. Business Namespace Isolation

Each business namespace has:
- **Independent API Key**: Unique authentication token
- **Isolated Cache Storage**: Data is completely separated between namespaces
- **Separate Metrics**: Per-namespace monitoring and statistics
- **Resource Limits**: Optional memory and key count limits

**Configuration Example:**
```yaml
config:
  namespaces:
    - name: game-app
      apikey: "your-secret-api-key-here"
      description: "Gaming application namespace"
      maxMemoryMB: 512           # Optional: max memory in MB
      maxKeys: 100000            # Optional: max number of keys
      defaultTTL: 3600           # Optional: default TTL in seconds
      
    - name: ads-app
      apikey: "another-secret-key"
      description: "Advertisement service"
      maxMemoryMB: 256
      maxKeys: 50000
      defaultTTL: 1800
```

**Access Control:**
- Clients must include `X-API-Key` header with their requests
- Proxy validates API key and maps to namespace
- Each namespace can only access its own cached data
- Cross-namespace access is prevented at the Proxy level

### 3. Dashboard Configuration

The Dashboard provides a web interface to monitor cluster health and statistics.

**Configuration:**
```yaml
config:
  dashboard:
    password: "admin-password"        # Dashboard login password
    refreshInterval: 5                 # UI refresh interval (seconds)
    authEnabled: true                  # Enable password authentication
    sessionTimeout: 30                 # Session timeout (minutes)
    
    display:
      showNodeDetails: true            # Show detailed node statistics
      showNamespaceStats: true         # Show per-namespace metrics
      showConnectionStats: true        # Show connection statistics
      maxRecentOperations: 100         # Max recent operations to display
```

**Dashboard Features:**

1. **Cluster Overview:**
   - Total number of business namespaces
   - Total number of cache nodes
   - Total active connections
   - Overall memory usage and cache hit rate

2. **Namespace Statistics:**
   - Number of keys per namespace
   - Memory usage per namespace
   - Request rate and latency
   - Cache hit/miss ratio

3. **Node Details:**
   - Cache node status (healthy/unhealthy)
   - Memory usage per node
   - Number of cache entries per node
   - Active connections per node
   - CPU and memory metrics

**Accessing the Dashboard:**
```bash
# Port-forward for local access
kubectl port-forward svc/yao-oracle-dashboard 8080:8080 -n yao-oracle

# Access at http://localhost:8080
# Login with configured password
```

## Installation

### Basic Installation

```bash
# Install with default values
helm install yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --create-namespace
```

### Production Installation

```bash
# Install with production values and secrets
helm install yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --create-namespace \
  -f ./helm/yao-oracle/values-prod.yaml \
  --set config.namespaces[0].apikey="your-secret-key-1" \
  --set config.namespaces[1].apikey="your-secret-key-2" \
  --set config.dashboard.password="dashboard-admin-password"
```

### Using External Secrets

For production, store secrets externally (e.g., Kubernetes Secrets, HashiCorp Vault):

```bash
# Create secret manually
kubectl create secret generic yao-oracle-api-keys \
  --from-literal=game-app-key="your-secret-key" \
  --from-literal=ads-app-key="another-secret-key" \
  --from-literal=dashboard-password="admin-password" \
  -n yao-oracle

# Install and reference external secrets
helm install yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --create-namespace \
  -f ./helm/yao-oracle/values-prod.yaml
```

## Configuration Updates

### Updating Namespaces (No Downtime)

Thanks to ConfigMap watching, you can add/modify namespaces without restarting pods:

```bash
# Update values file
vim helm/yao-oracle/values.yaml

# Add new namespace:
# namespaces:
#   - name: new-service
#     apikey: "new-service-key"
#     description: "New service namespace"

# Upgrade Helm release
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --set config.namespaces[2].apikey="new-service-key"

# Configuration is automatically reloaded by Proxy!
```

### Scaling Cache Nodes

```bash
# Scale to 5 nodes
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --set node.replicaCount=5

# Node addresses are automatically updated in ConfigMap
```

### Updating Dashboard Password

```bash
# Update password
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --set config.dashboard.password="new-password"

# Dashboard pods will restart to pick up new password
```

## Configuration Files

### values.yaml Structure

```
values.yaml
├── global                    # Global settings (registry, imagePullSecrets)
├── proxy                     # Proxy service configuration
│   ├── configWatch          # ConfigMap watching settings
│   ├── resources            # CPU/memory limits
│   └── autoscaling          # HPA settings
├── node                      # Cache node configuration
│   ├── replicaCount         # Number of nodes
│   ├── persistence          # Persistent storage
│   └── resources            # CPU/memory limits
├── dashboard                 # Dashboard configuration
│   ├── ingress              # Ingress settings
│   └── resources            # CPU/memory limits
├── config                    # Application configuration
│   ├── namespaces           # Business namespaces
│   └── dashboard            # Dashboard settings
└── rbac                      # RBAC settings
```

## Security Best Practices

### 1. API Key Management

- **Never commit API keys to Git**
- Use `--set` flags or external secret managers
- Rotate API keys regularly
- Use strong, random keys (32+ characters)

```bash
# Generate secure API key
openssl rand -base64 32
```

### 2. Dashboard Password

- Use strong passwords (16+ characters)
- Enable `authEnabled: true` in production
- Set appropriate `sessionTimeout`
- Use TLS/HTTPS with ingress

### 3. RBAC

The chart creates minimal RBAC permissions:
- ServiceAccount for Proxy/Dashboard
- Role with ConfigMap read/watch permissions
- Role with Secret read permissions

```yaml
rbac:
  create: true  # Always enable in production
```

### 4. Network Policies

Enable network policies to restrict pod-to-pod communication:

```yaml
networkPolicy:
  enabled: true
  policyTypes:
    - Ingress
    - Egress
```

## Monitoring

### Prometheus Integration

Enable ServiceMonitors for Prometheus:

```yaml
proxy:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
      interval: 30s

node:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
      interval: 30s
```

### Key Metrics

- `yao_oracle_cache_hits_total` - Cache hit count per namespace
- `yao_oracle_cache_misses_total` - Cache miss count per namespace
- `yao_oracle_keys_total` - Total keys per namespace
- `yao_oracle_memory_bytes` - Memory usage per namespace
- `yao_oracle_connections_active` - Active connections per node

## Troubleshooting

### Proxy not reloading configuration

Check ConfigMap watch permissions:
```bash
kubectl get role yao-oracle -n yao-oracle -o yaml
# Should have: configmaps [get, watch, list]
```

Check Proxy logs:
```bash
kubectl logs -l app.kubernetes.io/component=proxy -n yao-oracle
```

### Dashboard not showing data

Check if Dashboard can reach Proxy:
```bash
kubectl exec -it deploy/yao-oracle-dashboard -n yao-oracle -- \
  curl http://yao-oracle-proxy:8080/health
```

Check environment variables:
```bash
kubectl get pod -l app.kubernetes.io/component=dashboard -n yao-oracle -o yaml | grep -A 10 "env:"
```

### Namespace isolation not working

Verify API key configuration:
```bash
kubectl get secret yao-oracle-secret -n yao-oracle -o yaml
```

Test with different API keys:
```bash
# Should succeed
curl -H "X-API-Key: game-app-key" http://proxy:8080/set?key=test&value=123

# Should fail (wrong namespace)
curl -H "X-API-Key: ads-app-key" http://proxy:8080/get?key=test
```

## Examples

### Example 1: Three-Tier Application Setup

```yaml
config:
  namespaces:
    - name: web-frontend
      apikey: "web-frontend-secret"
      description: "Web frontend cache"
      maxMemoryMB: 256
      defaultTTL: 300
      
    - name: api-backend
      apikey: "api-backend-secret"
      description: "API backend cache"
      maxMemoryMB: 512
      defaultTTL: 600
      
    - name: analytics
      apikey: "analytics-secret"
      description: "Analytics data cache"
      maxMemoryMB: 1024
      defaultTTL: 3600
```

### Example 2: High-Availability Production

```yaml
proxy:
  replicaCount: 3
  configWatch:
    enabled: true
  autoscaling:
    enabled: true
    minReplicas: 3
    maxReplicas: 10

node:
  replicaCount: 5
  persistence:
    enabled: true
    storageClass: fast-ssd
    size: 50Gi
  podDisruptionBudget:
    enabled: true
    minAvailable: 3
```

### Example 3: Development Environment

```yaml
proxy:
  replicaCount: 1
  configWatch:
    enabled: true
  resources:
    requests:
      memory: "128Mi"
      cpu: "100m"

node:
  replicaCount: 2
  persistence:
    enabled: false

dashboard:
  ingress:
    enabled: false  # Use port-forward

config:
  dashboard:
    password: "dev-password"
    authEnabled: false  # Disable for dev
```

## References

- [Helm Documentation](https://helm.sh/docs/)
- [Kubernetes ConfigMaps](https://kubernetes.io/docs/concepts/configuration/configmap/)
- [Kubernetes RBAC](https://kubernetes.io/docs/reference/access-authn-authz/rbac/)
- [Prometheus Operator](https://github.com/prometheus-operator/prometheus-operator)

