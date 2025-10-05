# Yao-Oracle Helm Chart

[![Helm](https://img.shields.io/badge/Helm-v3-blue)](https://helm.sh/)
[![Kubernetes](https://img.shields.io/badge/Kubernetes-1.19+-blue)](https://kubernetes.io/)

A distributed KV cache system with namespace isolation, powered by Kubernetes.

## Features

### 🔄 Dynamic Configuration Management
- **ConfigMap Watching**: Proxy automatically reloads configuration on changes
- **Zero Downtime Updates**: Add namespaces or scale nodes without restart
- **Kubernetes-Native**: Uses native Kubernetes APIs for configuration sync

### 🔐 Business Namespace Isolation
- **Independent Authentication**: Each namespace has its own API key
- **Complete Data Isolation**: Namespaces cannot access each other's data
- **Resource Limits**: Optional memory and key count limits per namespace
- **Separate Metrics**: Per-namespace monitoring and statistics

### 📊 Comprehensive Dashboard
- **Cluster Overview**: Total namespaces, nodes, connections
- **Namespace Statistics**: Keys, memory, requests per namespace
- **Node Details**: Health, memory, connections per node
- **Password Protected**: Configurable authentication and session timeout

### 🚀 High Availability
- **Horizontal Scaling**: Scale Proxy and Cache Nodes independently
- **Auto-Scaling**: Built-in HPA support for Proxy
- **Pod Disruption Budgets**: Ensure minimum availability during updates
- **Anti-Affinity**: Distribute pods across nodes

### 📈 Production Ready
- **Prometheus Integration**: Built-in metrics and ServiceMonitors
- **Security Hardening**: Pod security contexts, read-only filesystems
- **Network Policies**: Optional network isolation
- **Persistent Storage**: Optional persistent volumes for cache data

## Quick Start

### Prerequisites

- Kubernetes 1.19+
- Helm 3.0+
- (Optional) Prometheus Operator for metrics

### Installation

```bash
# Add namespace
kubectl create namespace yao-oracle

# Install chart
helm install yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --set config.namespaces[0].apikey="your-secret-key" \
  --set config.dashboard.password="dashboard-password"

# Check deployment
kubectl get pods -n yao-oracle
```

### Access Dashboard

```bash
# Port-forward to access dashboard
kubectl port-forward -n yao-oracle svc/yao-oracle-dashboard 8080:8080

# Open browser to http://localhost:8080
# Login with configured password
```

### Test Cache

```bash
# Set a key
kubectl run -it --rm test --image=curlimages/curl --restart=Never -- \
  curl -H "X-API-Key: your-secret-key" \
  http://yao-oracle-proxy.yao-oracle:8080/set?key=test&value=hello

# Get the key
kubectl run -it --rm test --image=curlimages/curl --restart=Never -- \
  curl -H "X-API-Key: your-secret-key" \
  http://yao-oracle-proxy.yao-oracle:8080/get?key=test
```

## Architecture

```
┌─────────────────────────────────────────────────────────────┐
│                      Kubernetes Cluster                      │
├─────────────────────────────────────────────────────────────┤
│                                                               │
│  ┌──────────────┐        ┌──────────────────────────────┐  │
│  │  ConfigMap   │◄───────┤  Proxy (Deployment)          │  │
│  │  & Secret    │ Watch  │  - API Key Auth              │  │
│  │              │        │  - Consistent Hashing        │  │
│  │ - Namespaces │        │  - Request Routing           │  │
│  │ - API Keys   │        │  - Auto Config Reload        │  │
│  │ - Nodes      │        └─────┬───────────────────┬────┘  │
│  └──────────────┘              │                   │        │
│                                 │                   │        │
│                          gRPC   │                   │        │
│                                 ▼                   ▼        │
│                    ┌──────────────────┐   ┌──────────────┐ │
│                    │  Cache Node 0    │   │ Cache Node 1 │ │
│                    │  (StatefulSet)   │   │ (StatefulSet)│ │
│                    │                  │   │              │ │
│                    │  - KV Storage    │   │ - KV Storage │ │
│                    │  - Namespace     │   │ - Namespace  │ │
│                    │    Agnostic      │   │   Agnostic   │ │
│                    └──────────────────┘   └──────────────┘ │
│                                                               │
│  ┌───────────────────────────────────────────────────────┐  │
│  │  Dashboard (Deployment)                               │  │
│  │  - Password Protected Web UI                          │  │
│  │  - Cluster Statistics                                 │  │
│  │  - Namespace Metrics                                  │  │
│  │  - Node Health Monitoring                             │  │
│  └───────────────────────────────────────────────────────┘  │
│                              │                                │
│                              ▼                                │
│                    ┌──────────────────┐                      │
│                    │  Ingress         │                      │
│                    │  (Optional)      │                      │
│                    └──────────────────┘                      │
└─────────────────────────────────────────────────────────────┘
```

## Configuration

### Minimal Configuration

```yaml
# values.yaml
config:
  namespaces:
    - name: my-app
      apikey: "my-secret-key"
      description: "My application cache"
  
  dashboard:
    password: "dashboard-password"
```

### Production Configuration

```yaml
# values-prod.yaml
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

config:
  namespaces:
    - name: production-app
      apikey: ""  # Set via --set or external secrets
      description: "Production application"
      maxMemoryMB: 2048
      maxKeys: 500000
      defaultTTL: 3600
  
  dashboard:
    password: ""  # Set via --set or external secrets
    authEnabled: true
    sessionTimeout: 60
```

### Available Values

See [values.yaml](values.yaml) for all available configuration options.

Key configuration sections:
- `proxy.*` - Proxy service configuration
- `node.*` - Cache node configuration
- `dashboard.*` - Dashboard configuration
- `config.namespaces` - Business namespace definitions
- `config.dashboard` - Dashboard settings

## Dynamic Configuration Updates

The Proxy service watches the ConfigMap for changes and automatically reloads:

```bash
# Add a new namespace without downtime
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --reuse-values \
  --set config.namespaces[2].name=new-app \
  --set config.namespaces[2].apikey=new-key \
  --set config.namespaces[2].description="New application"

# The Proxy automatically detects and applies the changes!
```

## Security

### API Key Management

- Never commit API keys to version control
- Use `--set` flags or external secret managers
- Rotate keys regularly

```bash
# Generate secure API key
openssl rand -base64 32
```

### Dashboard Password

- Use strong passwords (16+ characters)
- Enable `authEnabled: true` in production
- Configure appropriate `sessionTimeout`

### RBAC

The chart creates minimal RBAC permissions:
- ConfigMap: get, watch, list (for dynamic config)
- Secret: get, list (for API keys and passwords)

### Network Policies

Enable network policies for production:

```yaml
networkPolicy:
  enabled: true
```

## Monitoring

### Prometheus

Enable ServiceMonitors for Prometheus scraping:

```yaml
proxy:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true

node:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
```

### Metrics Endpoint

Metrics are exposed on port 9090:

```bash
kubectl port-forward -n yao-oracle svc/yao-oracle-proxy 9090:9090
curl http://localhost:9090/metrics
```

## Upgrading

### Upgrade Chart

```bash
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  -f values-prod.yaml
```

### Scale Nodes

```bash
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --set node.replicaCount=5
```

### Update Configuration

```bash
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --set config.dashboard.refreshInterval=10
```

## Troubleshooting

### Check Pod Status

```bash
kubectl get pods -n yao-oracle
kubectl describe pod <pod-name> -n yao-oracle
```

### View Logs

```bash
# Proxy logs
kubectl logs -l app.kubernetes.io/component=proxy -n yao-oracle

# Node logs
kubectl logs -l app.kubernetes.io/component=node -n yao-oracle

# Dashboard logs
kubectl logs -l app.kubernetes.io/component=dashboard -n yao-oracle
```

### Test Connectivity

```bash
# Test proxy
kubectl run test --rm -it --image=curlimages/curl --restart=Never -- \
  curl -v http://yao-oracle-proxy.yao-oracle:8080/health

# Test node
kubectl run test --rm -it --image=curlimages/curl --restart=Never -- \
  curl -v http://yao-oracle-node-0.yao-oracle-node.yao-oracle:8080/health
```

## Documentation

- [Configuration Guide](CONFIG-GUIDE.md) - Detailed configuration documentation
- [Changelog](CHANGELOG.md) - Version history and changes

## Support

- GitHub: [https://github.com/eggybyte/yao-oracle](https://github.com/eggybyte/yao-oracle)
- Issues: [https://github.com/eggybyte/yao-oracle/issues](https://github.com/eggybyte/yao-oracle/issues)

## License

See the main project LICENSE file.
