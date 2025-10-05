# ✅ Flutter Dashboard + Mock-Admin 实现总结

## 🎉 已完成功能

### 1. **Flutter Web Dashboard**

#### ✨ 页面实现
- **Overview Page** - 集群总览页面
  - 实时连接状态指示器（LIVE/OFFLINE）
  - 组件健康状态（Proxies、Nodes）
  - 集群核心指标（QPS、Keys、Hit Rate、Latency）
  - 浮动操作按钮：Query Cache

- **Metrics Page** - 实时图表可视化 ⭐ NEW
  - QPS 实时折线图
  - Hit Rate 实时折线图
  - Memory Usage 实时折线图
  - Latency 实时折线图
  - 滚动时间窗口（最近 30 个数据点）

- **Nodes Page** - 缓存节点监控
  - 节点健康状态（HEALTHY/UNHEALTHY）
  - 内存使用率进度条
  - Key 统计、Hit Rate、Uptime
  - 实时数据更新

- **Namespaces Page** - 业务命名空间管理
  - 命名空间配置信息
  - 内存使用率进度条
  - QPS、Hit Rate、Rate Limit、TTL
  - 实时数据更新

- **Proxies Page** - 代理实例监控（暂时显示空状态）

#### 🎨 UI 组件
- **MetricsChart** - 实时折线图组件（基于 fl_chart）
- **MetricCard** - 指标卡片
- **QueryCacheDialog** - 缓存查询对话框
- **LoadingWidget** - 加载/错误/空状态

#### 🔗 核心功能
- **gRPC Streaming 订阅**
  - 自动连接 mock-admin
  - 实时接收 ClusterMetrics 更新
  - 连接断开自动重连（5 秒延迟）
  
- **状态管理（Provider）**
  - 全局 AppState 管理
  - 实时数据更新触发 UI 刷新
  - 历史数据存储（用于图表绘制）

---

### 2. **Mock-Admin 后端**

#### 📡 gRPC 服务
- **StreamMetrics** - Server streaming RPC
  - 立即推送初始快照
  - 周期性推送更新（默认 5 秒）
  - 支持命名空间过滤

- **QueryCache** - 单次查询 RPC
  - 返回 mock 数据（key、value、TTL、size 等）

- **ManageSecret** - 配置更新 RPC
  - 模拟 API Key 更新

- **GetConfig** - 配置查询 RPC
  - 返回所有命名空间配置

#### 🎲 Mock 数据生成
- **3 个命名空间**：game-app、ads-service、analytics
- **3 个缓存节点**：cache-node-0/1/2
- **动态指标变化**：
  - QPS 波动（100-200）
  - Hit Rate 波动（85-95%）
  - Memory 逐渐增长（带随机抖动）
  - 偶尔节点健康状态变化（5% 概率）

---

## 🚀 快速启动

### 一键启动脚本
```bash
./scripts/run-dashboard-dev.sh
```

**启动内容：**
1. 编译并启动 mock-admin（gRPC 端口 9090）
2. 安装 Flutter 依赖（首次运行）
3. 启动 Flutter Web 开发服务器（端口 8080）
4. 自动打开浏览器

**访问：** http://localhost:8080

---

## 📐 技术架构

```
┌─────────────────────────────────────────────┐
│          Flutter Web Dashboard              │
│                                             │
│  ┌────────────┐  ┌──────────────────────┐ │
│  │  Provider  │  │   GrpcClient         │ │
│  │  AppState  │◄─┤ (StreamMetrics)      │ │
│  └─────┬──────┘  └──────────────────────┘ │
│        │                                    │
│  ┌─────▼──────────────────────────────┐   │
│  │  Pages: Overview / Metrics /       │   │
│  │         Nodes / Namespaces         │   │
│  └────────────────────────────────────┘   │
└─────────────────┬───────────────────────────┘
                  │ gRPC Stream (HTTP/2)
                  │ ClusterMetrics (5s interval)
┌─────────────────▼───────────────────────────┐
│          Mock-Admin Service (Go)            │
│                                             │
│  ┌──────────────────┐  ┌────────────────┐ │
│  │ gRPC Server      │  │ MockDataGen    │ │
│  │ DashboardService │◄─┤ (周期性更新)  │ │
│  └──────────────────┘  └────────────────┘ │
└─────────────────────────────────────────────┘
```

---

## 📦 项目结构

```
yao-oracle/
├── frontend/dashboard/              # Flutter Web Dashboard
│   ├── lib/
│   │   ├── core/
│   │   │   ├── grpc_client.dart     # gRPC 客户端
│   │   │   └── app_state.dart       # Provider 状态管理
│   │   ├── pages/                   # 页面组件
│   │   │   ├── overview_page.dart
│   │   │   ├── metrics_page.dart    # ⭐ NEW 实时图表
│   │   │   ├── nodes_page.dart
│   │   │   ├── namespaces_page.dart
│   │   │   └── proxies_page.dart
│   │   ├── widgets/                 # UI 组件
│   │   │   ├── metrics_chart.dart   # ⭐ NEW fl_chart 封装
│   │   │   ├── query_cache_dialog.dart  # ⭐ NEW
│   │   │   ├── metric_card.dart
│   │   │   └── loading_widget.dart
│   │   ├── models/                  # 数据模型
│   │   │   └── metrics.dart
│   │   └── generated/               # Dart gRPC 生成代码
│   └── pubspec.yaml
│
├── cmd/mock-admin/                  # Mock-Admin 入口
│   └── main.go
│
├── internal/dashboard/              # Mock-Admin 实现
│   ├── grpc_server.go               # gRPC 服务器
│   ├── mock_data.go                 # Mock 数据生成
│   └── mock_config.go               # Mock 配置
│
├── api/yao/oracle/v1/
│   └── dashboard.proto              # gRPC API 定义
│
└── scripts/
    ├── run-dashboard-dev.sh         # ⭐ NEW 一键启动脚本
    └── generate_dart_grpc.sh        # Dart gRPC 代码生成
```

---

## 🎯 核心实现细节

### 1. gRPC Streaming 实现

**前端订阅（grpc_client.dart）：**
```dart
Stream<Map<String, dynamic>> streamMetrics({String namespace = ''}) {
  final request = SubscribeRequest()..namespace = namespace;
  final stream = _client.streamMetrics(request);

  _metricsSubscription = stream.listen(
    (clusterMetrics) {
      final data = _convertClusterMetrics(clusterMetrics);
      _metricsController.add(data);
    },
    onError: (error) {
      print('❌ gRPC stream error: $error');
      // 5 秒后自动重连
      Future.delayed(const Duration(seconds: 5), () {
        streamMetrics(namespace: namespace);
      });
    },
  );

  return _metricsController.stream;
}
```

**后端推送（grpc_server.go）：**
```go
func (s *DashboardGRPCServer) StreamMetrics(
    req *oraclev1.SubscribeRequest,
    stream oraclev1.DashboardService_StreamMetricsServer,
) error {
    ticker := time.NewTicker(s.refreshInterval)
    defer ticker.Stop()

    // 立即发送初始快照
    metrics, _ := s.collectClusterMetrics(req.Namespace)
    stream.Send(metrics)

    for {
        select {
        case <-stream.Context().Done():
            return nil
        case <-ticker.C:
            metrics, _ := s.collectClusterMetrics(req.Namespace)
            stream.Send(metrics)
        }
    }
}
```

---

### 2. 实时图表实现（metrics_chart.dart）

**数据点存储：**
```dart
class _MetricsPageState extends State<MetricsPage> {
  final List<MetricsDataPoint> _qpsHistory = [];
  final List<MetricsDataPoint> _hitRateHistory = [];
  
  static const int _maxDataPoints = 30; // 2.5 分钟

  void _addDataPoint(List<MetricsDataPoint> history, DateTime timestamp, double value) {
    history.add(MetricsDataPoint(timestamp: timestamp, value: value));
    if (history.length > _maxDataPoints) {
      history.removeAt(0); // 滚动窗口
    }
  }
}
```

**fl_chart 封装：**
```dart
LineChart(
  LineChartData(
    lineBarsData: [
      LineChartBarData(
        spots: dataPoints
            .asMap()
            .entries
            .map((e) => FlSpot(e.key.toDouble(), e.value.value))
            .toList(),
        isCurved: true,
        color: lineColor,
        belowBarData: BarAreaData(show: true, color: lineColor.withOpacity(0.1)),
      ),
    ],
  ),
)
```

---

### 3. Mock 数据生成逻辑

**周期性更新（mock_data.go）：**
```go
func (g *MockDataGenerator) updateMetrics() {
    g.mu.Lock()
    defer g.mu.Unlock()

    for _, node := range g.nodes {
        // 请求增长
        increment := int64(rand.Intn(100) + 50)
        node.RequestsTotal += increment

        // Hit Rate 波动
        hitRate := 0.85 + rand.Float64()*0.1
        hits := int64(float64(increment) * hitRate)
        node.Hits += hits
        node.Misses += increment - hits

        // 内存增长
        node.TotalKeys += int64(rand.Intn(150) - 25)
        node.MemoryUsedBytes = node.TotalKeys * (8 * 1024) // 8KB/key

        // 偶尔节点不健康（5% 概率）
        if rand.Float64() < 0.05 {
            node.Healthy = false
        } else {
            node.Healthy = true
        }
    }
}
```

---

## 📝 使用说明

### 查询缓存条目

1. 访问 Overview 页面
2. 点击右下角浮动按钮 "Query Cache"
3. 输入 Namespace 和 Key
4. 点击 "Query" 查看结果

**示例：**
- Namespace: `game-app`
- Key: `user:12345`

**返回数据：**
```
Key: user:12345
Value: {"mock":"data for user:12345"}
TTL: 60s
Size: 35 bytes
```

---

## 🛠️ 开发调试

### 重新生成 Dart gRPC 代码

```bash
./scripts/generate_dart_grpc.sh
```

### Flutter 热重载

在 Flutter Web 运行时：
- `r` - 热重载
- `R` - 热重启
- `q` - 退出

### 查看日志

**Frontend（浏览器控制台）：**
```
✅ Received metrics update: QPS=152.3, Nodes=3, Namespaces=3
```

**Backend（终端）：**
```
[INFO] 📊 Client subscribed to metrics stream (namespace: all)
[INFO] 🔄 Metrics update sent (QPS: 152.3, Hit Rate: 90.1%)
```

---

## 🎯 下一步计划

- [ ] ManageSecret UI 实现
- [ ] 热点 Key 排行榜
- [ ] 多集群切换
- [ ] 告警配置
- [ ] 真实 Admin Service 对接

---

## 📚 相关文档

- [快速开始指南](./DASHBOARD_QUICKSTART.md)
- [gRPC API 定义](./api/yao/oracle/v1/dashboard.proto)
- [项目架构说明](./docs/new-dashboard.md)

---

**✅ Dashboard 已就绪，立即体验实时监控！**

```bash
./scripts/run-dashboard-dev.sh
```

访问 http://localhost:8080 🚀
