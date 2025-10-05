# 🏗️ 整体架构方案

```
              ┌──────────────┐
 Client ────▶ │   Proxy      │
              │ - 业务空间隔离 │
              │ - 一致性哈希  │
              │ - 鉴权 (API Key)│
              │ - 配置监听    │
              └───────┬──────┘
                      │
       ┌──────────────┼───────────────┐
       ▼              ▼               ▼
  Cache Node 0    Cache Node 1    Cache Node 2
 (StatefulSet)   (StatefulSet)   (StatefulSet)
  - 存储KV        - 存储KV        - 存储KV
  - 独立命名空间   - 独立命名空间   - 独立命名空间
  - 简单API       - 简单API       - 简单API

              ┌──────────────┐
              │  Dashboard    │
              │ - 读取 Proxy 状态 │
              │ - 展示业务空间   │
              │ - 展示节点状态   │
              │ - 密码保护      │
              └──────────────┘
```

* **Proxy**：对外提供统一访问点，负责 **业务空间隔离 / 鉴权 / 一致性哈希路由 / ConfigMap 动态监听**
* **Cache Node**：最小 KV 存储，业务空间由 Proxy 决定，Node 不感知多租户
* **Dashboard**：单独微服务，展示集群 & 业务状态，密码保护
* **Core**：公共模块（KV 接口、Config Loader、HashRing、鉴权、中间件封装等）

---

# 📂 项目文件夹结构

```
yao-oracle/
├── cmd/
│   ├── proxy/        # Proxy 主入口
│   │   └── main.go
│   ├── node/         # Cache Node 主入口
│   │   └── main.go
│   └── dashboard/    # Dashboard 主入口
│       └── main.go
│
├── core/             # 公共核心模块
│   ├── config/       # 配置读取和监听 (ConfigMap)
│   │   └── loader.go
│   ├── hash/         # 一致性哈希实现
│   │   └── ring.go
│   ├── kv/           # KV 存储抽象
│   │   ├── cache.go
│   │   └── shard.go
│   ├── auth/         # APIKey 鉴权
│   │   └── middleware.go
│   ├── metrics/      # 状态收集 (Prometheus 或自定义)
│   │   └── collector.go
│   └── utils/        # 工具函数
│       └── logger.go
│
├── internal/         # 内部逻辑（非公共）
│   ├── proxy/        # Proxy 内部逻辑
│   │   ├── server.go
│   │   └── router.go
│   ├── node/         # Node 内部逻辑
│   │   └── server.go
│   └── dashboard/    # Dashboard 内部逻辑
│       └── server.go
│
├── web/              # Dashboard 前端资源（HTML/JS/CSS）
│   └── index.html
│
├── go.mod
└── README.md
```

---

# 🔑 核心功能实现要点

## 1. ConfigMap 动态监听（Proxy）

Proxy 要根据 Kubernetes ConfigMap 里的配置动态更新：

* 多个 **业务空间 (namespace)**：每个业务空间有独立 API Key
* 每次 ConfigMap 修改，Proxy **重新加载配置**
* ConfigMap 内容示例：

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: yao-oracle-config
data:
  proxy.json: |
    {
      "namespaces": {
        "game-app": {"apikey": "game123"},
        "ads-app": {"apikey": "ads456"}
      }
    }
  dashboard.json: |
    {
      "password": "admin@123"
    }
```

监听实现（core/config/loader.go）：

```go
func WatchConfig(file string, onChange func(Config)) {
    for {
        cfg, _ := load(file)
        onChange(cfg)
        time.Sleep(10 * time.Second) // 简单轮询，也可用 fsnotify
    }
}
```

---

## 2. Proxy（多业务空间 + 路由）

* Proxy 收到请求 → 提取 **API Key**（请求头/连接参数）
* 验证 API Key 属于哪个业务空间
* 在该空间的哈希环中找到目标节点
* 转发请求到 Cache Node

伪代码（internal/proxy/server.go）：

```go
func handleRequest(req Request) Response {
    ns := auth.ValidateAPIKey(req.APIKey)
    if ns == "" {
        return Response{Error: "Unauthorized"}
    }

    ring := rings[ns] // 每个业务空间独立的哈希环
    node := ring.GetNode(req.Key)
    resp := forwardToNode(node, req)
    return resp
}
```

---

## 3. Cache Node

* 最小 KV 存储（GET/SET/DELETE + TTL）
* 不感知多业务空间
* Proxy 决定 namespace，所以 Node 只存原始 key-value

示例：

```go
cache.Set("game-app:user:123", []byte("profile"), 3600)
cache.Get("game-app:user:123")
```

---

## 4. Dashboard

* 独立微服务（cmd/dashboard/main.go）
* 配置中读取 **dashboard 密码**
* 提供 Web UI（HTML/CSS + Ajax）
* 页面展示：

  * 业务空间数量、命中率、总连接数
  * 每个节点的缓存条目数、内存使用、活跃连接数

伪代码（internal/dashboard/server.go）：

```go
http.HandleFunc("/login", loginHandler)
http.HandleFunc("/metrics", authMiddleware(metricsHandler))
http.Handle("/", http.FileServer(http.Dir("./web")))
```

---

## 5. Core 公共模块

* **core/hash/ring.go**：一致性哈希实现（带虚拟节点）
* **core/kv/cache.go**：分片 HashMap 实现 TTL
* **core/auth/middleware.go**：APIKey 鉴权中间件
* **core/metrics/collector.go**：对 Node 和 Proxy 的状态采集

---

# 🌐 项目部署架构

1. **Cache Node**

   * StatefulSet，副本数可水平扩展
   * Headless Service 提供 Proxy 发现

2. **Proxy**

   * Deployment，多副本
   * 读取 ConfigMap，配置多业务空间
   * Service 对外暴露

3. **Dashboard**

   * Deployment，独立微服务
   * 从 Proxy / Node 拉取 metrics
   * ConfigMap 提供管理密码

---

# ✅ 总结

* 项目名：**yao-oracle**
* **三大微服务**：Proxy（集群大脑）、Cache Node（存储）、Dashboard（观测）
* **配置管理**：Proxy 动态监听 ConfigMap，支持多业务空间 + 独立 API Key；Dashboard 使用单独密码
* **数据隔离**：不同业务空间的 Key 前缀区分，互不可见
* **core 模块**：统一封装（KV、哈希环、鉴权、配置、指标）
