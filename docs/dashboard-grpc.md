# Yao-Oracle Dashboard - gRPC Streaming 架构

## 概述

Yao-Oracle Dashboard 现在使用 **gRPC streaming** 实现实时监控和配置管理，完全符合 `new-dashboard.md` 的设计要求。

## 系统架构

```
┌─────────────────────────────────┐
│   Flutter Web Dashboard         │
│  (gRPC Client + Stream Handler) │
└────────────┬────────────────────┘
             │ gRPC Stream (bidirectional)
             │
┌────────────▼────────────────────┐
│   Mock-Admin / Admin Service    │
│  (gRPC Server + StreamMetrics)  │
│  - Aggregates metrics           │
│  - Streams to dashboard         │
│  - Manages secrets              │
└─────────────────────────────────┘
```

## 主要变更

### 1. Protocol Buffers 定义

新的 `dashboard.proto` 提供了完整的 gRPC streaming API：

- **StreamMetrics**: 实时推送集群指标（ClusterMetrics 包含 GlobalStats, NamespaceStats, NodeStats）
- **QueryCache**: 查询特定缓存条目
- **ManageSecret**: 管理 API key
- **GetConfig**: 获取配置信息

### 2. 后端实现

#### Mock-Admin 服务

- **位置**: `cmd/mock-admin/main.go`
- **实现**: `internal/dashboard/grpc_server.go`
- **功能**:
  - 提供 gRPC streaming server
  - 使用 `MockDataGenerator` 生成测试数据
  - 每 N 秒推送一次 ClusterMetrics

#### 启动命令

```bash
# 编译 mock-admin
make build-local

# 运行 mock-admin (gRPC 端口 9090)
./bin/mock-admin --grpc-port=9090 --password=admin123 --refresh-interval=5
```

### 3. 前端实现

#### Dart gRPC Client

- **位置**: `frontend/dashboard/lib/core/grpc_client.dart`
- **功能**:
  - 连接 gRPC server
  - 订阅 StreamMetrics 流
  - 将 protobuf 消息转换为 Dart 模型
  - 提供向后兼容的 HTTP API 接口

#### 应用状态管理

- **位置**: `frontend/dashboard/lib/core/app_state.dart`
- **功能**:
  - 使用 `GrpcClient` 替代 `ApiClient`
  - 自动连接 gRPC stream
  - 实时更新 UI 数据（overview, nodes, namespaces）

## 开发流程

### 1. 生成代码

```bash
# 生成 Go gRPC 代码
make proto-generate

# 生成 Dart gRPC 代码
make proto-dart
```

### 2. 启动后端

```bash
# 启动 mock-admin (gRPC streaming server)
./bin/mock-admin --grpc-port=9090 --refresh-interval=5
```

### 3. 启动前端

```bash
# 进入 dashboard 目录
cd frontend/dashboard

# 安装依赖
flutter pub get

# 运行 Flutter Web (使用默认 localhost:9090)
flutter run -d chrome

# 或指定 gRPC 服务器地址
flutter run -d chrome --dart-define=GRPC_HOST=192.168.1.100 --dart-define=GRPC_PORT=9090
```

## 测试验证

### 1. 验证 gRPC 连接

打开浏览器开发者工具，查看控制台输出：

```
gRPC stream connected
Metrics received: { timestamp: ..., global: {...}, namespaces: [...], nodes: [...] }
```

### 2. 验证实时更新

观察 Dashboard 页面，指标应该每 5 秒自动刷新（根据 `--refresh-interval` 设置）。

### 3. 使用 grpcurl 测试

```bash
# 安装 grpcurl
brew install grpcurl

# 列出服务
grpcurl -plaintext localhost:9090 list

# 调用 StreamMetrics (会持续输出流数据)
grpcurl -plaintext localhost:9090 yao.oracle.v1.DashboardService/StreamMetrics

# 查询配置
grpcurl -plaintext localhost:9090 yao.oracle.v1.DashboardService/GetConfig
```

## 数据流

### Metrics 流

```
1. mock-admin: 启动 MockDataGenerator
   ↓
2. mock-admin: 定期生成 mock metrics (每 5 秒)
   ↓
3. Dashboard: 订阅 StreamMetrics
   ↓
4. mock-admin: 推送 ClusterMetrics protobuf message
   ↓
5. Dashboard: 接收并解析 ClusterMetrics
   ↓
6. Dashboard: 更新 UI (overview, nodes, namespaces)
```

### ClusterMetrics 结构

```json
{
  "timestamp": 1735890000,
  "global": {
    "qps": 150.5,
    "latency_ms": 2.5,
    "hit_rate": 0.92,
    "memory_used_mb": 450.2,
    "health_score": 0.95,
    "total_keys": 53300,
    "total_proxies": 1,
    "total_nodes": 3,
    "healthy_nodes": 3
  },
  "namespaces": [
    {
      "name": "game-app",
      "qps": 45.0,
      "hit_rate": 0.94,
      "ttl_avg": 60.0,
      "keys": 15000,
      "memory_used_mb": 120.0,
      "api_key": "game****3456",
      "description": "Gaming application cache",
      "max_memory_mb": 512,
      "default_ttl": 60,
      "rate_limit_qps": 100
    }
  ],
  "nodes": [
    {
      "id": "cache-node-0:7070",
      "ip": "cache-node-0:7070",
      "namespace": "",
      "memory_used_mb": 150.0,
      "hit_rate": 0.90,
      "latency_ms": 1.5,
      "key_count": 18000,
      "healthy": true,
      "uptime_seconds": 3600,
      "qps": 50.0
    }
  ]
}
```

## 配置

### 环境变量

#### Mock-Admin

- `--grpc-port`: gRPC 服务器端口（默认：9090）
- `--password`: Dashboard 密码（默认：admin123）
- `--refresh-interval`: 指标刷新间隔（秒，默认：5）

#### Dashboard

- `GRPC_HOST`: gRPC 服务器地址（默认：localhost）
- `GRPC_PORT`: gRPC 服务器端口（默认：9090）

### 生产部署

生产环境中，`mock-admin` 应该被真正的 `admin` 服务替代，该服务：

1. 连接到所有 Proxy 和 Node 实例
2. 聚合真实指标数据
3. 通过 gRPC streaming 推送到 Dashboard
4. 支持 Secret 管理和配置更新

## 后续改进

1. **认证/授权**: 添加 JWT 或 mTLS 认证
2. **压缩**: 启用 gRPC 消息压缩
3. **重连机制**: 自动重连断开的 stream
4. **时序数据**: 在前端缓存历史数据用于图表展示
5. **过滤器**: 支持更细粒度的订阅过滤（按 namespace, node 等）

## 故障排查

### 问题: Dashboard 无法连接 gRPC

**检查**:
1. mock-admin 是否运行: `ps aux | grep mock-admin`
2. 端口是否监听: `lsof -i :9090`
3. gRPC 地址是否正确: 检查 `GRPC_HOST` 和 `GRPC_PORT`

### 问题: Metrics 不更新

**检查**:
1. 浏览器控制台是否有错误
2. gRPC stream 是否断开（查看日志）
3. mock-admin 的 refresh-interval 设置

### 问题: proto 代码未生成

```bash
# 重新生成 Go 代码
cd api && buf generate

# 重新生成 Dart 代码
bash scripts/generate_dart_grpc.sh
```

## 参考文档

- [Protocol Buffers 设计文档](./new-dashboard.md)
- [gRPC 官方文档](https://grpc.io/)
- [Flutter gRPC 文档](https://pub.dev/packages/grpc)


