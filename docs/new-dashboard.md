# 🧭 Yao-Oracle Dashboard — Flutter Web + gRPC 实时监控与配置中心设计方案（完整落地版）

## 一、系统总体设计

### 🎯 目标概述

| 模块                            | 目标                                                         |
| ----------------------------- | ---------------------------------------------------------- |
| **Admin 后端服务**                | 连接所有 Proxy 与 Node，聚合全局与分区数据，通过 gRPC stream 推送 Dashboard 前端 |
| **Dashboard 前端（Flutter Web）** | 提供命名空间级实时监控、缓存查询、配置与密钥管理、节点可视化、时序图展示                       |
| **Kubernetes 配置源**            | ConfigMap 控制默认 TTL / 限制，Secret 控制每个命名空间的 API Key           |
| **通信机制**                      | gRPC 双向流 + Protobuf 定义统一 schema                            |
| **目标体验**                      | 「一屏洞察全局状态，一键深入命名空间详情」的实时观测控制中心                             |

---

## 二、系统架构与数据流

```
                ┌─────────────────────────────┐
                │   Kubernetes Control Plane  │
                │  - ConfigMap (TTL, Limits)  │
                │  - Secret (API Keys)        │
                └────────────┬────────────────┘
                             │
                   ┌─────────▼───────────┐
                   │   Yao-Oracle Admin   │
                   │  - gRPC Gateway      │
                   │  - Cluster Discovery │
                   │  - Informer Watcher  │
                   │  - Metrics Aggregator│
                   └───────┬──────────────┘
                           │ bidirectional gRPC stream
                           │
        ┌──────────────────▼──────────────────┐
        │     Yao-Oracle Dashboard (Flutter)  │
        │  - 实时监控视图                     │
        │  - 命名空间详情 & 缓存查询           │
        │  - 配置 & Secret 管理               │
        └────────────────────────────────────┘
```

---

## 三、gRPC 通信协议设计（核心）

### 1️⃣ 服务定义 — `dashboard.proto`

```protobuf
syntax = "proto3";

package yao.oracle.dashboard;

service DashboardStream {
  rpc StreamMetrics(SubscribeRequest) returns (stream ClusterMetrics);
  rpc QueryCache(CacheQueryRequest) returns (CacheQueryResponse);
  rpc ManageSecret(SecretUpdateRequest) returns (SecretUpdateResponse);
  rpc GetConfig(ConfigRequest) returns (ConfigResponse);
}

message SubscribeRequest {
  string namespace = 1; // 可为空，为空则订阅全局
}

message ClusterMetrics {
  int64 timestamp = 1;
  GlobalStats global = 2;
  repeated NamespaceStats namespaces = 3;
  repeated NodeStats nodes = 4;
}

message GlobalStats {
  double qps = 1;
  double latency_ms = 2;
  double hit_rate = 3;
  double memory_used_mb = 4;
  double health_score = 5;
}

message NamespaceStats {
  string name = 1;
  double qps = 2;
  double hit_rate = 3;
  double ttl_avg = 4;
  int64 keys = 5;
  double memory_used_mb = 6;
  string api_key = 7;
}

message NodeStats {
  string id = 1;
  string ip = 2;
  string namespace = 3;
  double memory_used_mb = 4;
  double hit_rate = 5;
  double latency_ms = 6;
  int64 key_count = 7;
}

message CacheQueryRequest {
  string namespace = 1;
  string key = 2;
}

message CacheQueryResponse {
  string key = 1;
  string value = 2;
  int64 ttl_seconds = 3;
  int64 size_bytes = 4;
  string created_at = 5;
  string last_access = 6;
}

message SecretUpdateRequest {
  string namespace = 1;
  string new_api_key = 2;
}

message SecretUpdateResponse {
  bool success = 1;
  string updated_at = 2;
}

message ConfigRequest {}
message ConfigResponse {
  repeated NamespaceConfig configs = 1;
}

message NamespaceConfig {
  string namespace = 1;
  int64 default_ttl = 2;
  int64 max_keys = 3;
}
```

---

## 四、前端架构设计（Flutter Web）

### 🧩 项目结构

```
lib/
├── main.dart
├── core/
│   ├── grpc/
│   │   ├── dashboard.pb.dart
│   │   ├── dashboard.pbgrpc.dart
│   │   └── grpc_client.dart
│   ├── state/
│   │   ├── global_state.dart
│   │   ├── namespace_state.dart
│   │   └── node_state.dart
│   └── utils/
│       ├── format.dart
│       └── theme.dart
├── pages/
│   ├── overview/
│   │   ├── overview_page.dart
│   │   └── widgets/
│   │       ├── global_chart.dart
│   │       ├── health_gauge.dart
│   │       └── topology_view.dart
│   ├── namespace/
│   │   ├── namespace_page.dart
│   │   └── namespace_detail.dart
│   ├── node/
│   │   ├── node_page.dart
│   │   └── node_detail.dart
│   ├── cache/
│   │   └── cache_query_page.dart
│   ├── config/
│   │   └── config_log_page.dart
│   └── settings/
│       └── settings_page.dart
└── widgets/
    ├── common_card.dart
    ├── data_table.dart
    ├── chart_line.dart
    └── toast.dart
```

---

## 五、📊 页面功能与交互设计（详细版）

---

### **1️⃣ Global Overview Page**

#### 🎯 功能目标

实时展示集群总体健康状态与性能曲线。

#### **展示内容**

| 模块    | 内容                                |
| ----- | --------------------------------- |
| 健康仪表盘 | Health Score（由命中率、延迟、内存综合）        |
| 实时性能图 | 折线图：QPS / Latency / Hit Rate      |
| 资源概览卡 | Proxy 总数 / Node 总数 / Namespace 总数 |
| 总资源使用 | Memory Used、Keys 总数、连接总数          |
| 动态拓扑图 | Proxy ↔ Node ↔ Namespace 连线关系     |

#### **交互**

* 点击节点跳转详情页
* Hover 展示节点 metrics
* WebSocket/gRPC 流更新时平滑动画刷新

---

### **2️⃣ Namespace Explorer Page**

#### 🎯 功能目标

展示每个命名空间的实时性能指标、配置、API Key 状态及历史变更。

#### **展示区块**

| 模块           | 内容                                           |
| ------------ | -------------------------------------------- |
| 命名空间卡片列表     | Name / QPS / Hit Rate / Memory / TTL / Key 数 |
| 详情展开区        | 折线图（QPS, HitRate, Latency）+ Secret 历史表格      |
| Secret 操作区   | 显示当前 API Key（mask）+ 按钮重置/更新                  |
| ConfigMap 信息 | Default TTL、Max Keys（来自 ConfigMap）           |

#### **交互**

* 点击命名空间展开详情
* 点击「🔑 重生成 API Key」→ gRPC 调用 `ManageSecret`
* 每当 Secret 更新 → 实时在卡片上打「已更新」标签
* TTL / Memory 实时动画增长曲线

---

### **3️⃣ Node Inspector Page**

#### 🎯 功能目标

分析单节点的行为趋势与资源使用。

#### **展示内容**

| 模块      | 内容                              |
| ------- | ------------------------------- |
| Node 信息 | ID / IP / Namespace / 状态 / 启动时间 |
| 实时图     | 折线图（内存使用、命中率、延迟）                |
| 热力图     | Key 数分布（时间 vs 数量）               |
| 指标卡     | QPS、Key Count、Memory、Hit Rate   |

#### **交互**

* 实时曲线平滑刷新
* 点击节点→ 打开浮层展示最近10分钟日志（gRPC stream）
* 异常状态节点标红闪烁

---

### **4️⃣ Cache Query Center Page**

#### 🎯 功能目标

提供通过 Namespace + Key 直接查询缓存的界面，用于调试或验证。

#### **布局**

| 区域    | 内容                                                  |
| ----- | --------------------------------------------------- |
| 查询输入区 | Namespace Dropdown + Key TextField + 查询按钮           |
| 查询结果区 | Key / Value(JSON Pretty) / TTL / Size / Last Access |
| 操作按钮  | 删除、刷新 TTL                                           |
| 动态状态  | TTL 倒计时动画更新                                         |

#### **交互**

* 输入 Key → 点击查询 → 调用 `QueryCache`
* Value 自动高亮 JSON（支持折叠）
* 删除缓存 → 调用 `/cache/delete` gRPC
* 若 TTL 更新 → 倒计时立即重置动画

---

### **5️⃣ Config & Secret Logs Page**

#### 🎯 功能目标

查看 Kubernetes Informer 推送的 ConfigMap / Secret 变更日志。

#### **展示内容**

| 字段        | 内容                         |
| --------- | -------------------------- |
| Namespace | 命名空间                       |
| 类型        | Secret / ConfigMap         |
| 更新时间      | 时间戳                        |
| 变更摘要      | 如 “game-app: TTL 60 → 120” |
| 来源        | informer / manual          |
| 状态        | ✅ 已同步 / ⚠️ 等待刷新            |

#### **交互**

* 滚动加载历史（无限滚动）
* Diff 高亮（旧 vs 新值）
* 过滤器：按 Namespace / 类型筛选

---

### **6️⃣ Settings Page**

| 模块   | 内容                                 |
| ---- | ---------------------------------- |
| 用户信息 | 当前登录管理员 / Token 有效期                |
| 刷新频率 | 下拉选项（5s / 10s / 30s）               |
| 主题   | Dark / Light                       |
| 连接状态 | 🟢 Connected / 🔴 Disconnected（实时） |
| 日志导出 | 下载当前 metrics JSON                  |

---

## 六、UI / UX 风格规范

| 元素   | 风格描述                             |
| ---- | -------------------------------- |
| 主色调  | 深蓝 + 霓虹渐变线条（科技监控风）               |
| 字体   | Inter / JetBrains Mono           |
| 组件形态 | 玻璃拟态半透明背景 + 阴影 + 渐变描边            |
| 动画   | 平滑曲线刷新、闪烁提示更新、展开折叠动效             |
| 图表库  | `fl_chart`（折线 / 柱状 / 仪表 / 热力）    |
| 布局   | 响应式 `LayoutBuilder` + `GridView` |

---

## 七、状态管理与实时逻辑

| 模块        | 状态管理方式                                  |
| --------- | --------------------------------------- |
| 全局指标      | Riverpod / Provider 单例状态                |
| 命名空间      | Scoped Provider（按命名空间区分）                |
| 节点数据      | StreamBuilder + gRPC Stream 绑定          |
| 查询结果      | FutureBuilder + gRPC 单次请求               |
| Secret 变更 | 事件流监听 informer → Admin → gRPC → Flutter |

---

## 八、示意布局（文字版）

```
───────────────────────────────────────────────
 Yao-Oracle Dashboard  |  🟢 Connected  | ⚙️ Settings
───────────────────────────────────────────────
[Global Health Score] [Global QPS Trend] [Memory Usage]
───────────────────────────────────────────────
Namespace Overview
┌─────────────────────────────────────────────┐
│ game-app      QPS:120  Hit:94%  TTL:60s     │ 🔑 Updated
│ ads-service   QPS:40   Hit:89%  TTL:120s    │
└─────────────────────────────────────────────┘
───────────────────────────────────────────────
Cache Query  |  Node Inspector  |  Config Logs
───────────────────────────────────────────────
```

---

## 九、部署与运维建议

| 部署项     | 说明                             |
| ------- | ------------------------------ |
| 构建命令    | `flutter build web --release`  |
| 容器化部署   | 通过 nginx 或 Higress 暴露 `/`      |
| 访问认证    | JWT + API Key 双层认证（由 Admin 颁发） |
| gRPC 网关 | 支持 HTTP/2 + TLS                |
| 性能优化    | Flutter Web CanvasKit 渲染模式     |
| 缓存层     | 使用 IndexedDB 缓存最近指标数据          |

---

## 🔚 最终成效

| 维度    | 效果                               |
| ----- | -------------------------------- |
| 动态感知  | Dashboard 实时显示全局与命名空间指标          |
| 配置热更新 | Secret / ConfigMap 变更实时反映到前端     |
| 操作性   | 支持 API Key 管理、缓存查询与删除            |
| 高级可视化 | GPU 加速曲线、健康仪表盘、拓扑图               |
| 云原生融合 | 完全依托 Kubernetes API 与 gRPC 双流架构  |
| 技术统一  | Flutter Web + gRPC → 高性能跨平台可视化系统 |