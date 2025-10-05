# Yao-Oracle Cursor Rules

This directory contains Cursor Rules that guide AI-assisted development for the Yao-Oracle project.

## Available Rules

### Core Architecture Rules (Always Applied)

1. **[configuration-management.mdc](configuration-management.mdc)** 
   - Two-tier configuration system (Environment Variables + ConfigMap/Secret)
   - Configuration loading strategies per service
   - Hot reload implementation with fsnotify
   - Configuration validation patterns

2. **[project-structure.mdc](project-structure.mdc)**
   - Directory structure and module boundaries
- Import path conventions
   - File naming standards
   - Service architecture overview

### Implementation Rules (Fetched by Description)

3. **[microservices-implementation.mdc](microservices-implementation.mdc)** _(NEW)_
   - Service initialization patterns
   - Configuration loading implementation for each service
   - Hot reload patterns with mutex protection
   - Server implementation with graceful shutdown

4. **[infrastructure.mdc](infrastructure.mdc)** _(UPDATED)_
   - Build scripts and Makefile standards
   - Docker multi-stage builds and buildx
   - Helm chart structure and templates
   - **NEW Section**: Configuration Management in Helm
     - ConfigMap and Secret templates
     - Deployment configuration injection
     - Environment-specific values files
     - Configuration update workflows

5. **[protobuf-and-buf.mdc](protobuf-and-buf.mdc)**
   - Protocol Buffers API design
   - Buf configuration and generation
   - Breaking change detection

6. **[dashboard.mdc](dashboard.mdc)**
   - Dashboard architecture and pages
   - API endpoints and WebSocket
   - Chart.js visualization patterns

7. **[go-code-quality.mdc](go-code-quality.mdc)**
   - Go coding standards
   - Error handling patterns
   - Testing best practices

8. **[web-frontend.mdc](web-frontend.mdc)**
   - Frontend HTML/CSS/JS standards
   - Dashboard UI implementation

## Configuration Management Architecture

The project uses a **two-tier configuration system**:

### Tier 1: Environment Variables (Infrastructure Config)
```
GRPC_PORT, HTTP_PORT        → Service ports
LOG_LEVEL, LOG_FORMAT       → Logging configuration
CONFIG_PATH                 → Path to mounted config file
NODE_SERVICE, PROXY_SERVICE → Service discovery
MAX_MEMORY_MB, MAX_KEYS     → Resource limits
```

**Deployment**: Defined in `values.yaml` → Injected as `env` in Deployment spec

### Tier 2: ConfigMap/Secret (Business Config)
```json
{
  "proxy": {
    "namespaces": [
      {
        "name": "game-app",
        "apikey": "secret-key",
        "maxMemoryMB": 512,
        "maxKeys": 100000
      }
    ]
  },
  "dashboard": {
    "password": "admin-password",
    "refreshInterval": 5
  }
}
```

**Deployment**: Defined in `values.yaml` → Rendered to Secret → Mounted as file

### Service-Specific Configuration

| Service   | Env Vars | Secret | Hot Reload | Config File Path                          |
|-----------|----------|--------|------------|-------------------------------------------|
| Proxy     | ✅       | ✅     | ✅         | `/etc/yao-oracle/config-with-secrets.json`|
| Node      | ✅       | ❌     | ❌         | N/A (stateless)                           |
| Dashboard | ✅       | ✅     | ✅         | `/etc/yao-oracle/config-with-secrets.json`|

## Key Changes in This Update

### 1. Enhanced infrastructure.mdc
- **New Section 4**: Configuration Management in Helm
  - Complete ConfigMap and Secret template examples
  - Deployment manifests showing env var injection and volume mounts
  - Configuration update workflows (infrastructure vs business config)
  - Best practices for environment-specific deployments
  - External secret management integration examples

### 2. New microservices-implementation.mdc
- Complete `main.go` implementations for all three services
- Configuration loading patterns with hot reload
- Server implementation with `UpdateConfig()` methods
- Thread-safe config access using `sync.RWMutex`
- File watcher implementation with debouncing
- Testing patterns for configuration hot reload
- Service initialization and graceful shutdown

### 3. Updated configuration-management.mdc
- Already comprehensive, no changes needed
- Referenced by both new rules

## Usage Guide

### For AI Assistants

1. **Project Structure Questions**: Reference `project-structure.mdc` (always applied)
2. **Configuration Questions**: Reference `configuration-management.mdc` (always applied)
3. **Implementing Services**: Fetch `microservices-implementation.mdc`
4. **Helm Deployments**: Fetch `infrastructure.mdc`
5. **API Design**: Fetch `protobuf-and-buf.mdc`

### For Developers

When working on:
- **Proxy/Node/Dashboard services**: Read `microservices-implementation.mdc`
- **Helm charts and deployment**: Read `infrastructure.mdc` section 4
- **Configuration changes**: Read `configuration-management.mdc`
- **Adding new namespaces**: Update `values.yaml` → `helm upgrade` (hot reload)
- **Changing ports**: Update `values.yaml` → `helm upgrade` (rolling restart)

## Configuration Update Examples

### Update Business Config (Hot Reload, No Restart)
```bash
# Edit values.yaml - add/remove namespaces or change API keys
vim helm/yao-oracle/values.yaml

# Apply changes
helm upgrade yao-oracle ./helm/yao-oracle --namespace yao-oracle

# Kubernetes updates Secret → Mounted files update → Services hot reload
```

### Update Infrastructure Config (Rolling Restart)
```bash
# Edit values.yaml - change ports, resource limits, replica counts
vim helm/yao-oracle/values.yaml

# Apply changes (triggers rolling restart)
helm upgrade yao-oracle ./helm/yao-oracle --namespace yao-oracle --wait
```

## Best Practices Summary

### Configuration Management
- ✅ Use environment variables for infrastructure config
- ✅ Use Secret for sensitive business config
- ✅ Implement hot reload for business config changes
- ✅ Validate config before applying
- ✅ Use mutex for thread-safe config access
- ✅ Keep Node service stateless (no config file)

### Helm Deployments
- ✅ Mount entire JSON files, not individual keys
- ✅ Use `readOnly: true` for config mounts
- ✅ Create environment-specific values files
- ✅ Use external secret management for production
- ✅ Validate templates with `helm template`

### Service Implementation
- ✅ Load env vars first, then config files
- ✅ Implement graceful shutdown
- ✅ Log config changes with timestamps
- ✅ Use context for cancellation
- ✅ Debounce file watcher events

## Related Documentation

- [Configuration Management Rule](configuration-management.mdc) - Two-tier config system
- [Infrastructure Rule](infrastructure.mdc) - Helm integration and deployment
- [Microservices Implementation Rule](microservices-implementation.mdc) - Service code patterns
- [Project Structure Rule](project-structure.mdc) - Overall architecture

## Version History

- **v1.2.0** (2025-10-03): Added microservices-implementation.mdc, enhanced infrastructure.mdc with config management
- **v1.1.0**: Added configuration-management.mdc with two-tier system
- **v1.0.0**: Initial rules for project structure and code quality
