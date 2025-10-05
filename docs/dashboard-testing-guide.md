# 🎯 Dashboard Testing Guide

## Quick Start

**One-command test:**
```bash
make test-dashboard
```

This command will:
1. ✅ Check all dependencies (Go, Flutter, protoc)
2. 🔄 Generate Dart gRPC code from proto files
3. 📦 Install Flutter dependencies
4. 🚀 Start mock-admin gRPC service (port 9090)
5. 🎨 Start Flutter web dashboard (port 8080)
6. 📺 Display real-time console output from both services

## What You'll See

### Mock Admin Service Output

```
╔══════════════════════════════════════════════════════════════╗
║         🎯 Yao-Oracle Mock Admin Service (Test Mode)       ║
╚══════════════════════════════════════════════════════════════╝

[INFO] Starting mock-admin service...
[INFO] Configuration:
[INFO]   - gRPC Port: 9090
[INFO]   - Refresh Interval: 5 seconds
[INFO]   - Dashboard Password: admin123
[INFO]   - Test Mode: Enabled (Mock Data)

[INFO] ✅ gRPC server listening on localhost:9090
[INFO] 📡 Dashboard clients can now connect and stream metrics
[INFO] 🔄 Mock data refreshing every 5 seconds
[INFO] Ready to accept connections...
─────────────────────────────────────────────────────────────
[INFO] 📊 Client subscribed to metrics stream (namespace: all)
[INFO] ✅ Sent initial metrics snapshot (QPS: 150.5, Hit Rate: 89.2%, Nodes: 3)
[INFO] 🔄 Metrics update sent (QPS: 152.3, Hit Rate: 90.1%, Memory: 435.0MB, Nodes: 3/3 healthy)
```

### Flutter Dashboard Output

```
🔌 Initializing gRPC connection to localhost:9090
✅ gRPC client initialized
🔄 Connecting to gRPC metrics stream...
📊 Subscribing to metrics stream (namespace: all)
✅ Metrics stream subscription active
✅ gRPC stream connected successfully
✅ Received metrics update: QPS=150.5, Nodes=3, Namespaces=3
```

## Access Points

Once started, you can access:

- **Dashboard UI**: http://localhost:8080
- **Mock Admin gRPC**: localhost:9090
- **Admin Logs**: `/tmp/yao-mock-admin.log`
- **Dashboard Logs**: `/tmp/yao-dashboard-flutter.log`

## Testing Features

### 1. Overview Page
- Global cluster metrics (QPS, latency, hit rate, health score)
- Component counts (proxies, nodes, namespaces)
- Real-time updates every 5 seconds

### 2. Nodes Page
- 3 mock cache nodes (cache-node-0, cache-node-1, cache-node-2)
- Memory usage, key count, hit rate per node
- Node health status (green = healthy, red = unhealthy)

### 3. Namespaces Page
- 3 mock namespaces:
  - `game-app`: Gaming application cache
  - `ads-service`: Advertisement service cache
  - `analytics`: Analytics data cache
- Per-namespace QPS, hit rate, memory usage
- Configuration: default TTL, max memory, rate limits

### 4. Proxies Page
- Proxy health status
- Namespace count
- Node health summary

## Mock Data Characteristics

The mock data generator creates realistic behavior:

- **Dynamic Metrics**: Values change over time (simulated traffic patterns)
- **Realistic Distributions**: Nodes have different load characteristics
- **Health Events**: Nodes occasionally become unhealthy (5% chance)
- **Memory Growth**: Keys grow/shrink based on simulated eviction
- **Hit Rate Variations**: Different namespaces have different hit rates:
  - Game app: ~92% (high, cached user sessions)
  - Ads service: ~85% (moderate)
  - Analytics: ~78% (lower, more unique queries)

## Debugging

### View Real-time Logs

**Admin Service:**
```bash
tail -f /tmp/yao-mock-admin.log
```

**Flutter Dashboard:**
```bash
tail -f /tmp/yao-dashboard-flutter.log
```

### Test gRPC Endpoint

**Using grpcurl:**
```bash
# List services
grpcurl -plaintext localhost:9090 list

# List methods
grpcurl -plaintext localhost:9090 list yao.oracle.v1.DashboardService

# Test StreamMetrics
grpcurl -plaintext localhost:9090 yao.oracle.v1.DashboardService/StreamMetrics
```

### Common Issues

**Port 9090 already in use:**
```bash
# Find and kill process using port 9090
lsof -ti:9090 | xargs kill -9
```

**Port 8080 already in use:**
```bash
# Find and kill process using port 8080
lsof -ti:8080 | xargs kill -9
```

**Flutter dependencies not installed:**
```bash
cd frontend/dashboard
flutter pub get
```

**Dart gRPC code not generated:**
```bash
bash scripts/generate_dart_grpc.sh
```

## Stopping Services

Press **Ctrl+C** in the terminal where you ran `make test-dashboard`.

The script will automatically:
1. Stop Flutter web server
2. Stop mock-admin gRPC service
3. Clean up background processes
4. Display shutdown confirmation

## Architecture

```
┌─────────────────────────────────────────────────────┐
│                Flutter Web Browser                  │
│              (http://localhost:8080)                │
└───────────────────┬─────────────────────────────────┘
                    │ gRPC stream (port 9090)
                    │
┌───────────────────▼─────────────────────────────────┐
│              Mock Admin Service                     │
│  - gRPC DashboardService implementation             │
│  - Mock data generator (updates every 5s)           │
│  - 3 namespaces, 3 nodes, 1 proxy                  │
└─────────────────────────────────────────────────────┘
```

## Next Steps

After testing the dashboard:

1. **Implement Real Admin Service**: Replace mock data with real metrics collection
2. **Add Kubernetes Integration**: Deploy to cluster and test with real pods
3. **Add Authentication**: Implement JWT-based dashboard login
4. **Add More Features**: Cache query, secret management, config editing

## Related Documentation

- [Dashboard Architecture](dashboard.mdc) - Full design specification
- [Admin Service](admin.mdc) - Admin service implementation details
- [Protocol Buffers](protobuf-and-buf.mdc) - gRPC API definitions


