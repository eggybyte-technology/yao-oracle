# Helm Chart Changelog

## [Unreleased]

### Added - Enhanced Configuration Management

#### 1. Dynamic ConfigMap Monitoring
- **Proxy service now supports dynamic ConfigMap watching**
  - Automatically detects and reloads configuration changes
  - No pod restart required for config updates
  - Configurable reload interval
  - RBAC permissions for ConfigMap watch operations
  
  **Configuration:**
  ```yaml
  proxy:
    configWatch:
      enabled: true
      configMapName: ""  # Auto-generated
      reloadInterval: 10
  ```

#### 2. Enhanced Business Namespace Configuration
- **Extended namespace configuration with resource limits**
  - `maxMemoryMB`: Maximum memory per namespace
  - `maxKeys`: Maximum number of keys
  - `defaultTTL`: Default TTL for cache entries
  
  **Example:**
  ```yaml
  config:
    namespaces:
      - name: game-app
        apikey: "secret-key"
        description: "Gaming application"
        maxMemoryMB: 512
        maxKeys: 100000
        defaultTTL: 3600
  ```

- **Improved API key isolation**
  - Each namespace has independent authentication
  - Cross-namespace access prevention
  - Separate metrics per namespace

#### 3. Comprehensive Dashboard Configuration
- **Enhanced dashboard settings for monitoring**
  - Authentication control (`authEnabled`)
  - Session timeout configuration
  - UI refresh interval
  - Display preferences (node details, namespace stats, connections)
  
  **Configuration:**
  ```yaml
  config:
    dashboard:
      password: "admin-password"
      refreshInterval: 5
      authEnabled: true
      sessionTimeout: 30
      display:
        showNodeDetails: true
        showNamespaceStats: true
        showConnectionStats: true
        maxRecentOperations: 100
  ```

- **Dashboard features:**
  - Cluster overview (namespaces, nodes, connections)
  - Per-namespace statistics (keys, memory, requests)
  - Per-node details (health, memory, connections)

#### 4. Configuration Management Improvements
- **Automatic service discovery**
  - Node addresses auto-populated from StatefulSet
  - Proxy addresses auto-generated for Dashboard
  - Support for manual override when needed
  
  ```yaml
  config:
    nodes: []  # Auto-populated, or specify manually
    proxyAddresses: []  # Auto-populated, or specify manually
  ```

- **Dual configuration approach**
  - **ConfigMap** (`yao-oracle-config`): Public configuration (no secrets)
  - **Secret** (`yao-oracle-secret`): Sensitive data (API keys, passwords)
  - Both support dynamic updates through Helm upgrades

#### 5. Enhanced Templates
- **ConfigMap template** (`configmap.yaml`)
  - Includes all namespace settings
  - Auto-generates node addresses
  - Includes proxy addresses for dashboard
  - Includes dashboard display settings

- **Secret template** (`secret.yaml`)
  - Complete configuration with secrets
  - Individual keys for easy access
  - Used by Proxy and Dashboard services

- **Proxy deployment** (`proxy/deployment.yaml`)
  - ConfigMap watching environment variables
  - Pod metadata injection
  - Improved volume mounts

- **Dashboard deployment** (`dashboard/deployment.yaml`)
  - Dashboard-specific environment variables
  - Proxy service address injection
  - Authentication settings

#### 6. Documentation
- **CONFIG-GUIDE.md**: Comprehensive configuration guide
  - Feature explanations
  - Configuration examples
  - Security best practices
  - Troubleshooting tips
  
- **Enhanced NOTES.txt**: Post-installation guidance
  - Dynamic configuration update instructions
  - Security reminders for default credentials
  - Step-by-step testing procedures

#### 7. Production Configuration
- **Enhanced values-prod.yaml**
  - ConfigMap watching enabled
  - Resource limits for all namespaces
  - Comprehensive dashboard settings
  - Example multi-namespace setup

### Changed
- **RBAC**: Ensured ConfigMap watch permissions
- **Environment variables**: Added configuration for ConfigMap watching
- **Values structure**: Extended with new configuration options

### Security Enhancements
- **API Key Management**
  - Support for external secret managers
  - Individual keys stored in Secret
  - Validation warnings for default keys

- **Dashboard Authentication**
  - Configurable authentication toggle
  - Session timeout control
  - Strong password recommendations

- **Network Isolation**
  - Namespace-level isolation enforced at Proxy
  - Independent API keys per namespace
  - No cross-namespace data access

## Migration Guide

### From Previous Version

If upgrading from a previous version:

1. **Review new configuration options**
   ```bash
   # Check new values
   helm show values ./helm/yao-oracle
   ```

2. **Enable ConfigMap watching** (optional but recommended)
   ```bash
   helm upgrade yao-oracle ./helm/yao-oracle \
     --set proxy.configWatch.enabled=true \
     -n yao-oracle
   ```

3. **Add namespace resource limits** (optional)
   ```bash
   helm upgrade yao-oracle ./helm/yao-oracle \
     --set config.namespaces[0].maxMemoryMB=512 \
     --set config.namespaces[0].maxKeys=100000 \
     -n yao-oracle --reuse-values
   ```

4. **Update dashboard settings** (optional)
   ```bash
   helm upgrade yao-oracle ./helm/yao-oracle \
     --set config.dashboard.authEnabled=true \
     --set config.dashboard.sessionTimeout=30 \
     -n yao-oracle --reuse-values
   ```

## Notes

### Breaking Changes
- None. All new features are backward compatible.

### Deprecations
- None.

### Known Issues
- ConfigMap watching requires RBAC permissions (automatically created when `rbac.create=true`)
- Dashboard metrics require Proxy service to be accessible

### Future Enhancements
- Support for external ConfigMap watching (e.g., from etcd)
- Enhanced monitoring dashboard with Grafana integration
- Support for namespace quotas and rate limiting
- Multi-cluster configuration support

