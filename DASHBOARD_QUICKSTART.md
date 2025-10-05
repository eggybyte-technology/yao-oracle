# 🎯 Yao-Oracle Dashboard 快速开始

## 概述

Yao-Oracle Dashboard 是一个基于 **Flutter Web** 构建的实时监控界面，通过 **gRPC streaming** 与 mock-admin 后端通信，实现实时数据可视化。

## 架构设计

```
┌─────────────────────────────────┐
│   Flutter Web Dashboard         │
│   - Real-time charts            │
│   - gRPC streaming client       │
│   - Provider state management   │
└────────────┬────────────────────┘
             │ gRPC Stream (HTTP/2)
             │ bidirectional
┌────────────▼────────────────────┐
│   Mock-Admin Service (Go)       │
│   - gRPC Server                 │
│   - Mock data generator         │
│   - StreamMetrics RPC           │
└─────────────────────────────────┘
```

## 功能特性

### ✅ 已实现功能

1. **实时数据流**
   - 基于 gRPC streaming 的实时指标推送
   - 5 秒间隔自动更新（可配置）
   - 连接状态实时显示

2. **页面导航**
   - **Overview** - 集群总览（QPS、Hit Rate、Memory、Health Score）
   - **Metrics** - 实时图表可视化（QPS、Hit Rate、Memory、Latency）
   - **Proxies** - 代理实例监控
   - **Nodes** - 缓存节点监控（内存使用、健康状态、Key 统计）
   - **Namespaces** - 业务命名空间管理（QPS、Hit Rate、资源配额）

3. **数据可视化**
   - 使用 `fl_chart` 实现实时折线图
   - 滚动时间窗口（最近 30 个数据点 = 2.5 分钟）
   - 颜色编码的健康状态指示器
   - 响应式布局（支持移动端和桌面端）

4. **Mock-Admin 后端**
   - 模拟 3 个命名空间（game-app、ads-service、analytics）
   - 模拟 3 个缓存节点（动态健康状态）
   - 逼真的指标变化（QPS 波动、Hit Rate 变化、内存增长）
   - 周期性更新（默认 5 秒）

## 快速启动

### 方式一：一键启动脚本（推荐）

```bash
# 启动 mock-admin 和 Flutter Dashboard
./scripts/run-dashboard-dev.sh
```

**启动后：**
- Dashboard 会自动在浏览器打开 `http://localhost:8080`
- mock-admin 在后台运行，监听 `localhost:9090`
- 实时数据每 5 秒自动刷新

**停止服务：**
- 按 `Ctrl+C` 即可停止所有服务

---

### 方式二：手动分步启动

#### 1️⃣ 启动 mock-admin

```bash
# 编译 mock-admin
make build-local

# 运行 mock-admin
./bin/mock-admin --grpc-port=9090 --password=admin123 --refresh-interval=5
```

**mock-admin 输出：**
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
```

#### 2️⃣ 启动 Flutter Dashboard

```bash
cd frontend/dashboard

# 安装依赖（首次运行）
flutter pub get

# 启动 Flutter Web 开发服务器
flutter run -d chrome --web-port=8080 \
    --dart-define=GRPC_HOST=localhost \
    --dart-define=GRPC_PORT=9090
```

**访问地址：**
- **Dashboard**: http://localhost:8080

---

## 目录结构

```
frontend/dashboard/
├── lib/
│   ├── core/                    # 核心功能模块
│   │   ├── grpc_client.dart     # gRPC 客户端（StreamMetrics 订阅）
│   │   └── app_state.dart       # 全局状态管理（Provider）
│   ├── pages/                   # 页面组件
│   │   ├── overview_page.dart   # 总览页面
│   │   ├── metrics_page.dart    # 实时图表页面（NEW ✨）
│   │   ├── nodes_page.dart      # 节点监控页面
│   │   ├── namespaces_page.dart # 命名空间管理页面
│   │   └── proxies_page.dart    # 代理监控页面
│   ├── widgets/                 # 可复用组件
│   │   ├── metrics_chart.dart   # 实时折线图组件（NEW ✨）
│   │   ├── metric_card.dart     # 指标卡片
│   │   └── loading_widget.dart  # 加载状态
│   ├── models/                  # 数据模型
│   │   └── metrics.dart         # 指标数据结构
│   ├── generated/               # gRPC 生成代码（自动生成，勿手动修改）
│   │   └── yao/oracle/v1/
│   │       ├── dashboard.pb.dart
│   │       ├── dashboard.pbgrpc.dart
│   │       └── dashboard.pbjson.dart
│   └── main.dart                # 应用入口
├── pubspec.yaml                 # Flutter 依赖配置
└── web/                         # Web 资源
```

---

## 技术栈

### 前端（Flutter Web）

- **框架**: Flutter 3.9+
- **状态管理**: Provider 6.1+
- **gRPC 客户端**: grpc-dart 4.1.0
- **图表库**: fl_chart 1.1.1
- **Protobuf**: protobuf 4.2.0

### 后端（Mock-Admin）

- **语言**: Go 1.23+
- **gRPC 服务器**: google.golang.org/grpc
- **数据生成**: 周期性 mock 数据更新

---

## gRPC API 说明

### StreamMetrics（Server Streaming）

**请求：**
```protobuf
message SubscribeRequest {
  string namespace = 1; // 可选，过滤特定命名空间
}
```

**响应流：**
```protobuf
message ClusterMetrics {
  int64 timestamp = 1;
  GlobalStats global = 2;
  repeated NamespaceStats namespaces = 3;
  repeated NodeStats nodes = 4;
}
```

**流程：**
1. Dashboard 连接 mock-admin 的 `StreamMetrics` RPC
2. mock-admin 立即发送初始快照
3. 每隔 5 秒推送更新的 ClusterMetrics
4. Dashboard 接收到数据后自动更新 UI

---

## 开发调试

### 查看 gRPC 日志

**Flutter 端（浏览器控制台）：**
```
✅ Received metrics update: QPS=152.3, Nodes=3, Namespaces=3
```

**mock-admin 端（终端）：**
```
[INFO] 📊 Client subscribed to metrics stream (namespace: all)
[INFO] ✅ Sent initial metrics snapshot (QPS: 150.5, Hit Rate: 89.2%, Nodes: 3)
[INFO] 🔄 Metrics update sent (QPS: 152.3, Hit Rate: 90.1%, Memory: 435.0MB, Nodes: 3/3 healthy)
```

### 重新生成 Dart gRPC 代码

如果修改了 `api/yao/oracle/v1/dashboard.proto`：

```bash
# 重新生成 Dart gRPC 代码
./scripts/generate_dart_grpc.sh
```

### 热重载

Flutter Web 支持热重载，修改代码后按 `r` 刷新：
```bash
# 在 Flutter Web 运行时
r      # 热重载
R      # 热重启
q      # 退出
```

---

## 常见问题

### Q1: 连接失败 "Failed to connect: Connection refused"

**解决：**
1. 确认 mock-admin 是否正在运行：
   ```bash
   lsof -i :9090
   ```
2. 检查 gRPC 端口配置是否正确：
   ```bash
   # 应为 localhost:9090
   flutter run -d chrome --dart-define=GRPC_HOST=localhost --dart-define=GRPC_PORT=9090
   ```

### Q2: 图表不显示数据

**解决：**
1. 检查浏览器控制台是否有 gRPC 错误
2. 确认 `isStreamConnected` 状态为 `true`（右上角显示 "LIVE"）
3. 等待 5-10 秒让数据积累（至少需要 2 个数据点才能绘制图表）

### Q3: 编译错误 "Undefined name 'DashboardServiceClient'"

**解决：**
```bash
# 重新生成 Dart gRPC 代码
cd frontend/dashboard
flutter pub get
../../scripts/generate_dart_grpc.sh
```

---

## 下一步计划

### 🚧 待实现功能

1. **配置管理**
   - [ ] QueryCache UI（查询缓存条目）
   - [ ] ManageSecret UI（更新 API Key）
   - [ ] Namespace 配置编辑

2. **增强可视化**
   - [ ] 热点 Key 排行榜
   - [ ] 集群拓扑图
   - [ ] 告警配置界面

3. **生产集成**
   - [ ] 真实 Admin Service 对接（替换 mock-admin）
   - [ ] 身份认证（JWT Token）
   - [ ] 多集群切换

---

## 参考文档

- [Flutter gRPC 官方文档](https://grpc.io/docs/languages/dart/)
- [fl_chart 图表库](https://pub.dev/packages/fl_chart)
- [Yao-Oracle 架构设计](./docs/new-dashboard.md)
- [Protobuf 定义](./api/yao/oracle/v1/dashboard.proto)

---

## 贡献者

如需贡献代码，请遵循以下步骤：

1. Fork 本仓库
2. 创建功能分支 (`git checkout -b feature/new-chart`)
3. 提交更改 (`git commit -am 'Add real-time alerts'`)
4. 推送到分支 (`git push origin feature/new-chart`)
5. 创建 Pull Request

---

**🎉 现在你可以开始体验 Yao-Oracle Dashboard 了！**

```bash
# 一键启动
./scripts/run-dashboard-dev.sh
```

访问 http://localhost:8080 查看实时监控界面！

