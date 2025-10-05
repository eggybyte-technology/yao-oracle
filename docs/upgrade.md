# 🧭 Yao-Oracle 动态配置与监控实践方案

## 一、系统架构总览

Yao-Oracle 是一个云原生的分布式 KV 缓存系统，部署在 Kubernetes 上，由三大微服务组成：

| 组件            | 角色      | 主要职责                              |
| ------------- | ------- | --------------------------------- |
| **Proxy**     | 控制与网关层  | 命名空间隔离、一致性哈希路由、API Key 鉴权、动态配置热更新 |
| **Node**      | 存储与数据层  | 负责 KV 存储、TTL、LRU 淘汰、资源上限控制        |
| **Dashboard** | 监控与可视化层 | 监控集群状态、节点健康、业务空间使用情况、配置变更历史       |

所有服务在 Kubernetes 内部通过 Helm 部署和 ConfigMap/Secret 配置驱动。

---

## 二、动态配置方案（基于 Kubernetes Informer）

### 🎯 目标

实现**在不重启 Pod 的情况下**，当 ConfigMap 或 Secret 发生更新时：

* Proxy 和 Dashboard 能实时检测到变更；
* 重新加载配置；
* 验证合法性；
* 动态刷新内存配置。

---

### 🧩 实现思路

1. **放弃 fsnotify 文件监听**（Kubernetes 的 symlink 机制复杂且有延迟）；
2. 改用 Kubernetes 官方推荐的 `client-go Informer`；
3. Dashboard 和 Proxy 启动后：

   * 使用 `InClusterConfig()` 连接 Kubernetes API；
   * Watch 指定的 ConfigMap / Secret；
   * 当 Update 事件触发时，自动调用 `onConfigChange()`；
   * 校验后热更新配置（使用 RWMutex 保护）。

---

### ✅ 实现示例：`core/config/informer.go`

```go
package config

import (
    "context"
    "fmt"
    "k8s.io/client-go/informers"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
    "k8s.io/client-go/tools/cache"
    corev1 "k8s.io/api/core/v1"
)

type DynamicConfigWatcher struct {
    Namespace string
    ConfigMap string
    Secret    string
}

func NewDynamicConfigWatcher(ns, cm, sec string) *DynamicConfigWatcher {
    return &DynamicConfigWatcher{Namespace: ns, ConfigMap: cm, Secret: sec}
}

func (w *DynamicConfigWatcher) Start(onChange func(kind string, data map[string]string)) error {
    cfg, err := rest.InClusterConfig()
    if err != nil {
        return err
    }
    clientset, err := kubernetes.NewForConfig(cfg)
    if err != nil {
        return err
    }

    factory := informers.NewSharedInformerFactoryWithOptions(
        clientset, 0, informers.WithNamespace(w.Namespace),
    )

    // Watch ConfigMap
    cmInformer := factory.Core().V1().ConfigMaps().Informer()
    cmInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
        UpdateFunc: func(_, newObj interface{}) {
            cm := newObj.(*corev1.ConfigMap)
            if cm.Name == w.ConfigMap {
                fmt.Println("🔄 ConfigMap updated -> reloading...")
                onChange("ConfigMap", cm.Data)
            }
        },
    })

    // Watch Secret
    secInformer := factory.Core().V1().Secrets().Informer()
    secInformer.AddEventHandler(cache.ResourceEventHandlerFuncs{
        UpdateFunc: func(_, newObj interface{}) {
            sec := newObj.(*corev1.Secret)
            if sec.Name == w.Secret {
                fmt.Println("🔑 Secret updated -> reloading...")
                data := make(map[string]string)
                for k, v := range sec.StringData {
                    data[k] = v
                }
                onChange("Secret", data)
            }
        },
    })

    stop := make(chan struct{})
    go factory.Start(stop)
    return nil
}
```

> 💡 **Proxy 和 Dashboard 都可复用该模块。**

---

### 🔒 并发安全热更新（示例）

```go
var (
    mu   sync.RWMutex
    cfg  *Config
)

func applyNewConfig(newCfg *Config) {
    mu.Lock()
    defer mu.Unlock()
    cfg = newCfg
}

func GetConfig() *Config {
    mu.RLock()
    defer mu.RUnlock()
    return cfg
}
```

---

## 三、配置层设计

### 1️⃣ 环境变量（所有组件）

| 环境变量             | 示例值                 | 说明     |
| ---------------- | ------------------- | ------ |
| `NAMESPACE`      | `yao-system`        | 当前命名空间 |
| `LOG_LEVEL`      | `info`              | 日志级别   |
| `CONFIGMAP_NAME` | `yao-oracle-config` | 动态配置来源 |
| `SECRET_NAME`    | `yao-oracle-secret` | 敏感配置来源 |

---

### 2️⃣ Proxy 特有环境变量

| 变量名                      | 示例值                                               | 说明                   |
| ------------------------ | ------------------------------------------------- | -------------------- |
| `GRPC_PORT`              | `8080`                                            | gRPC 服务端口            |
| `HTTP_PORT`              | `9090`                                            | 管理接口端口               |
| `METRICS_PORT`           | `9100`                                            | Prometheus 指标端口      |
| `PROXY_HEADLESS_SERVICE` | `yao-proxy-headless.yao-system.svc.cluster.local` | Dashboard 发现 Proxy 用 |
| `NODE_HEADLESS_SERVICE`  | `yao-node-headless.yao-system.svc.cluster.local`  | Proxy 发现 Node 用      |
| `DISCOVERY_MODE`         | `k8s`                                             | 启用 K8s API 发现        |
| `DISCOVERY_INTERVAL`     | `10`                                              | 集群发现刷新间隔秒            |

---

### 3️⃣ Node 特有环境变量

| 变量名               | 示例值       | 说明       |
| ----------------- | --------- | -------- |
| `GRPC_PORT`       | `7070`    | gRPC 端口  |
| `MAX_MEMORY_MB`   | `1024`    | 最大内存     |
| `MAX_KEYS`        | `1000000` | 最大 key 数 |
| `EVICTION_POLICY` | `LRU`     | 淘汰策略     |
| `METRICS_PORT`    | `9101`    | 指标端口     |

---

### 4️⃣ Dashboard 特有环境变量

| 变量名                 | 示例值                                               | 说明                            |
| ------------------- | ------------------------------------------------- | ----------------------------- |
| `HTTP_PORT`         | `8081`                                            | Web 服务端口                      |
| `METRICS_PORT`      | `9102`                                            | Prometheus 指标端口               |
| `PROXY_SERVICE_DNS` | `yao-proxy-headless.yao-system.svc.cluster.local` | Proxy 发现 DNS                  |
| `NODE_SERVICE_DNS`  | `yao-node-headless.yao-system.svc.cluster.local`  | Node 发现 DNS                   |
| `DISCOVERY_MODE`    | `k8s`                                             | 使用 Kubernetes API 查询 Endpoint |
| `REFRESH_INTERVAL`  | `5`                                               | 页面刷新间隔（秒）                     |

---

### 5️⃣ ConfigMap 示例

```yaml
apiVersion: v1
kind: ConfigMap
metadata:
  name: yao-oracle-config
  namespace: yao-system
data:
  config.json: |
    {
      "proxy": {
        "namespaces": [
          {"name": "game-app", "maxMemoryMB": 512, "defaultTTL": 60, "rateLimitQPS": 100},
          {"name": "ads-service", "maxMemoryMB": 256, "defaultTTL": 120, "rateLimitQPS": 50}
        ]
      },
      "dashboard": {
        "refreshInterval": 5,
        "theme": "dark",
        "enableRealtime": true
      }
    }
```

---

### 6️⃣ Secret 示例

```yaml
apiVersion: v1
kind: Secret
metadata:
  name: yao-oracle-secret
  namespace: yao-system
type: Opaque
stringData:
  config-with-secrets.json: |
    {
      "proxy": {
        "apikeys": {
          "game-app": "game-secret-123",
          "ads-service": "ads-secret-456"
        }
      },
      "dashboard": {
        "password": "super-secure-password",
        "jwtSecret": "jwt-signing-key"
      }
    }
```

---

## 四、集群发现机制（Proxy & Dashboard）

### 🎯 目标

* Proxy 能发现所有 Node 实例；
* Dashboard 能发现所有 Proxy 与 Node；
* 使用 **Kubernetes Endpoints API**（而非 DNS）以便判断健康状态。

---

### 🧠 实现逻辑（Go）

```go
import (
    "context"
    "fmt"
    "k8s.io/client-go/kubernetes"
    "k8s.io/client-go/rest"
)

func DiscoverPods(namespace, serviceName string) ([]string, error) {
    cfg, _ := rest.InClusterConfig()
    clientset, _ := kubernetes.NewForConfig(cfg)

    ep, err := clientset.CoreV1().Endpoints(namespace).Get(context.TODO(), serviceName, metav1.GetOptions{})
    if err != nil {
        return nil, err
    }

    var addrs []string
    for _, subset := range ep.Subsets {
        for _, addr := range subset.Addresses {
            addrs = append(addrs, addr.IP)
        }
    }
    return addrs, nil
}
```

Dashboard 使用：

```go
proxies, _ := DiscoverPods("yao-system", "yao-proxy-headless")
nodes, _ := DiscoverPods("yao-system", "yao-node-headless")
```

---

### 🧩 Headless Service 示例

```yaml
apiVersion: v1
kind: Service
metadata:
  name: yao-proxy-headless
  namespace: yao-system
spec:
  clusterIP: None
  selector:
    app: yao-proxy
  ports:
  - port: 8080
    name: grpc
---
apiVersion: v1
kind: Service
metadata:
  name: yao-node-headless
  namespace: yao-system
spec:
  clusterIP: None
  selector:
    app: yao-node
  ports:
  - port: 7070
    name: grpc
```

---

## 五、Dashboard 页面与指标设计

### 1️⃣ 页面结构

```
Dashboard Web UI
├── Login (Password via Secret)
├── Cluster Overview
├── Proxy Nodes
├── Cache Nodes
├── Namespaces
├── Configuration History
```

---

### 2️⃣ Cluster Overview

**展示内容：**

* Proxy 总数、Node 总数
* 在线率、平均响应延迟
* 请求 QPS（全局）
* 缓存命中率
* 内存占用趋势

**图表：**

* 📈 折线图：QPS vs Time
* 📊 柱状图：命中率 per namespace
* 🌐 拓扑图：Proxy ↔ Node 连接关系
* 💡 仪表盘：集群健康评分

---

### 3️⃣ Proxy Nodes 页面

**展示：**

* 每个 Proxy 的 IP、启动时间、命名空间数量
* QPS、错误率、平均延迟
* 最近配置变更时间

**图表：**

* 折线图：每秒请求数 / 延迟变化
* 饼图：命名空间流量分布
* 时间线：配置更新事件

---

### 4️⃣ Cache Nodes 页面

**展示：**

* IP、内存使用率、key 数量
* TTL 平均值、过期率、命中率

**图表：**

* 热力图：内存使用分布
* 折线图：命中率 vs 时间
* 柱状图：TTL 分布区间

---

### 5️⃣ Namespaces 页面

**展示：**

* 各命名空间 key 数、内存占比、限流阈值
* 命中率、QPS、错误率
* API Key 状态与上次刷新时间

**图表：**

* 柱状图：key 数 vs 命名空间
* 折线图：命中率趋势
* 表格：API Key 列表与状态

---

### 6️⃣ 配置变更页面

**展示：**

* ConfigMap / Secret 更新日志
* 更新来源、时间、摘要
* Proxy/Dashboard 自动 reload 状态

**图表：**

* 时间轴 (timeline)
* 日志表格

---

## 六、最佳实践总结

| 类别               | 最佳实践                              | 理由         |
| ---------------- | --------------------------------- | ---------- |
| **动态配置**         | 使用 Informer 代替 fsnotify           | 低延迟、稳定可靠   |
| **并发安全**         | RWMutex 包装配置状态                    | 防止读写冲突     |
| **集群发现**         | Kubernetes API + Headless Service | 精准、支持健康检测  |
| **配置管理**         | ConfigMap（非敏感）+ Secret（敏感）        | 职责清晰、安全合规  |
| **Dashboard 设计** | 分层图表展示（集群/节点/命名空间）                | 可视化清晰，便于扩展 |
| **监控指标**         | QPS、延迟、命中率、TTL、内存                 | 完整覆盖缓存系统性能 |
| **部署方式**         | Helm Chart 统一管理                   | 环境一致性、易于升级 |

---

✅ **最终效果：**

* Proxy 与 Dashboard 都能在运行中自动感知配置变化；
* 集群节点自动发现、健康状态实时展示；
* Dashboard 提供清晰的多维指标与交互式可视化；
* 整个系统完全原生地与 Kubernetes 控制面集成。