# Helm Chart Upgrade Notes

本文档说明了 Yao-Oracle Helm Chart 的最新改进和配置要点。

## 🎯 主要改进

### 1. **自动服务发现和地址配置**

所有服务现在自动通过 Kubernetes 服务发现连接，无需手动配置地址。

#### Proxy 服务
- 自动发现所有 Cache Node 实例
- 使用 StatefulSet headless service 进行节点路由
- 命令行参数自动生成：
  ```bash
  -port=8080
  -nodes=yao-oracle-node-0.yao-oracle-node.default.svc.cluster.local:8080,
         yao-oracle-node-1.yao-oracle-node.default.svc.cluster.local:8080,
         yao-oracle-node-2.yao-oracle-node.default.svc.cluster.local:8080
  ```

#### Dashboard 服务
- 自动连接到 Proxy 服务
- 自动发现所有 Cache Node 实例
- 命令行参数自动生成：
  ```bash
  -port=8080
  -proxy=yao-oracle-proxy.default.svc.cluster.local:8080
  -nodes=yao-oracle-node-0.yao-oracle-node.default.svc.cluster.local:8080,
         yao-oracle-node-1.yao-oracle-node.default.svc.cluster.local:8080,
         yao-oracle-node-2.yao-oracle-node.default.svc.cluster.local:8080
  ```

#### Node 服务
- 简化配置，只需指定端口
- 通过 StatefulSet headless service 提供稳定的网络标识

### 2. **ConfigMap 动态监听**

#### 启用方式
```yaml
proxy:
  configWatch:
    enabled: true
    configMapName: ""  # 自动生成
    reloadInterval: 10

dashboard:
  configWatch:
    enabled: true
    configMapName: ""  # 自动生成
    reloadInterval: 10
```

#### 工作原理
- 服务自动监听 ConfigMap 变化
- 配置更新后无需重启 Pod
- 适用于：namespace 配置、API Key 更新、节点配置变更

### 3. **增强的日志系统**

所有服务现在输出带颜色的结构化日志：

```
╔═══════════════════════════════════════════════════════╗
║                                                       ║
║          🔮 Yao-Oracle Distributed KV Cache          ║
║                    Proxy Service                      ║
║                                                       ║
╚═══════════════════════════════════════════════════════╝

[INFO]  proxy-main Starting Proxy Service...
[STEP]  proxy-main [1/6] Parsing command line arguments
[INFO]  proxy-main Configuration: port=8080, config=/etc/yao-oracle/config.json
[SUCCESS] proxy-main Initialization complete!
```

颜色方案：
- 🔵 `INFO` - 蓝色（一般信息）
- 🟢 `SUCCESS` - 绿色（成功操作）
- 🟡 `WARN` - 黄色（警告）
- 🔴 `ERROR` - 红色（错误）
- 🔷 `STEP` - 青色（初始化步骤）

### 4. **新的 Dashboard UI**

#### 多 Tab 页面设计
- 📊 **Overview（总览）** - 集群关键指标
- 📁 **Namespaces（业务空间）** - 命名空间列表
- 🔀 **Proxy Instances（Proxy 实例）** - Proxy 详情
- 💾 **Cache Nodes（缓存节点）** - 节点详情

#### 登录流程更新
- 登录页面：`/login.html`
- 登录成功后跳转到：`/dashboard.html`（新的多 Tab 页面）
- 原有的 `/index.html` 仍然保留作为简单视图

## 📝 部署示例

### 完整部署
```bash
# 使用默认配置部署
helm install yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --create-namespace

# 使用自定义配置
helm install yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --create-namespace \
  --set proxy.replicaCount=3 \
  --set node.replicaCount=5 \
  --set config.dashboard.password=MySecurePassword
```

### 生产环境部署
```bash
helm install yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --create-namespace \
  --values ./helm/yao-oracle/values-prod.yaml \
  --set-string config.namespaces[0].apikey=$GAME_API_KEY \
  --set-string config.namespaces[1].apikey=$ADS_API_KEY \
  --set-string config.dashboard.password=$DASHBOARD_PASSWORD
```

### 启用 Ingress
```bash
helm install yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --create-namespace \
  --set dashboard.ingress.enabled=true \
  --set dashboard.ingress.hosts[0].host=dashboard.example.com \
  --set dashboard.ingress.className=nginx
```

## 🔄 升级现有部署

### 添加新的 Namespace
```bash
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --reuse-values \
  --set config.namespaces[3].name=new-service \
  --set config.namespaces[3].apikey=new-secret-key \
  --set config.namespaces[3].description="New Service"
```

### 扩容 Cache Nodes
```bash
# 从 3 个节点扩容到 5 个节点
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --reuse-values \
  --set node.replicaCount=5
```

### 更新 API Key（安全方式）
```bash
# 创建 Secret
kubectl create secret generic custom-apikeys \
  --from-literal=game-app-key=$NEW_GAME_KEY \
  --from-literal=ads-app-key=$NEW_ADS_KEY \
  --namespace yao-oracle

# 升级使用新 Secret
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --set config.namespaces[0].apikey=$NEW_GAME_KEY \
  --set config.namespaces[1].apikey=$NEW_ADS_KEY
```

## 🔍 验证部署

### 检查 Pod 状态
```bash
kubectl get pods -n yao-oracle -l app.kubernetes.io/instance=yao-oracle
```

预期输出：
```
NAME                        READY   STATUS    RESTARTS   AGE
yao-oracle-proxy-0          1/1     Running   0          2m
yao-oracle-proxy-1          1/1     Running   0          2m
yao-oracle-node-0           1/1     Running   0          2m
yao-oracle-node-1           1/1     Running   0          2m
yao-oracle-node-2           1/1     Running   0          2m
yao-oracle-dashboard-xxx    1/1     Running   0          2m
```

### 检查服务发现
```bash
# 检查 Proxy 服务
kubectl get svc -n yao-oracle yao-oracle-proxy

# 检查 Node Headless 服务
kubectl get svc -n yao-oracle yao-oracle-node

# 检查 Dashboard 服务
kubectl get svc -n yao-oracle yao-oracle-dashboard
```

### 查看服务日志
```bash
# Proxy 日志（查看彩色初始化日志）
kubectl logs -n yao-oracle -l app.kubernetes.io/component=proxy --tail=100

# Node 日志
kubectl logs -n yao-oracle -l app.kubernetes.io/component=node --tail=100

# Dashboard 日志（查看服务连接信息）
kubectl logs -n yao-oracle -l app.kubernetes.io/component=dashboard --tail=100
```

### 验证服务连接
```bash
# 查看 Dashboard 的启动日志，确认它正确连接到 Proxy 和 Nodes
kubectl logs -n yao-oracle -l app.kubernetes.io/component=dashboard | grep -E "Connected|proxy|node"
```

预期输出：
```
[INFO]  dashboard-main Proxy address: yao-oracle-proxy.yao-oracle.svc.cluster.local:8080
[INFO]  dashboard-main Node addresses: 3 configured
[INFO]  dashboard-main   Node 1: yao-oracle-node-0.yao-oracle-node.yao-oracle.svc.cluster.local:8080
[INFO]  dashboard-main   Node 2: yao-oracle-node-1.yao-oracle-node.yao-oracle.svc.cluster.local:8080
[INFO]  dashboard-main   Node 3: yao-oracle-node-2.yao-oracle-node.yao-oracle.svc.cluster.local:8080
[SUCCESS] dashboard Connected to proxy: yao-oracle-proxy.yao-oracle.svc.cluster.local:8080
[SUCCESS] dashboard Connected to node: yao-oracle-node-0.yao-oracle-node.yao-oracle.svc.cluster.local:8080
```

## 🔐 安全建议

### 1. 更改默认密码
```bash
# 生成安全密码
NEW_PASSWORD=$(openssl rand -base64 32)

# 更新 Dashboard 密码
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --reuse-values \
  --set config.dashboard.password=$NEW_PASSWORD
```

### 2. 更改默认 API Keys
```bash
# 为每个 namespace 生成唯一的 API Key
GAME_KEY=$(openssl rand -base64 32)
ADS_KEY=$(openssl rand -base64 32)
ANALYTICS_KEY=$(openssl rand -base64 32)

# 更新 API Keys
helm upgrade yao-oracle ./helm/yao-oracle \
  --namespace yao-oracle \
  --reuse-values \
  --set config.namespaces[0].apikey=$GAME_KEY \
  --set config.namespaces[1].apikey=$ADS_KEY \
  --set config.namespaces[2].apikey=$ANALYTICS_KEY
```

### 3. 启用网络策略
```yaml
# values-prod.yaml
networkPolicy:
  enabled: true
  policyTypes:
    - Ingress
    - Egress
```

### 4. 使用 TLS
```yaml
# Dashboard Ingress with TLS
dashboard:
  ingress:
    enabled: true
    className: nginx
    annotations:
      cert-manager.io/cluster-issuer: letsencrypt-prod
    tls:
      - secretName: yao-oracle-dashboard-tls
        hosts:
          - dashboard.example.com
```

## 📊 监控和观测

### 访问 Dashboard
```bash
# Port-forward 到 Dashboard
kubectl port-forward -n yao-oracle svc/yao-oracle-dashboard 8080:8080

# 打开浏览器访问
open http://localhost:8080/login.html
```

### Prometheus 集成
如果启用了 metrics（默认启用）：

```bash
# Port-forward 到 Proxy metrics 端点
kubectl port-forward -n yao-oracle svc/yao-oracle-proxy 9090:9090

# 访问 metrics
curl http://localhost:9090/metrics
```

### 启用 ServiceMonitor
```yaml
# values.yaml
proxy:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true  # 需要 Prometheus Operator
      interval: 30s

node:
  metrics:
    enabled: true
    serviceMonitor:
      enabled: true
      interval: 30s
```

## 🐛 故障排查

### Dashboard 无法连接到 Proxy

**症状：** Dashboard 显示 "No proxy configured"

**解决方法：**
```bash
# 检查 Dashboard Pod 的环境变量和命令行参数
kubectl describe pod -n yao-oracle -l app.kubernetes.io/component=dashboard

# 查看完整的启动命令
kubectl get pod -n yao-oracle -l app.kubernetes.io/component=dashboard -o yaml | grep -A 10 args

# 验证 Proxy 服务可达性
kubectl run -it --rm debug --image=nicolaka/netshoot -n yao-oracle -- \
  nc -zv yao-oracle-proxy.yao-oracle.svc.cluster.local 8080
```

### Proxy 无法连接到 Nodes

**症状：** Proxy 日志显示连接错误

**解决方法：**
```bash
# 检查 Proxy Pod 的命令行参数
kubectl logs -n yao-oracle -l app.kubernetes.io/component=proxy | grep nodes

# 测试 Node headless service
kubectl run -it --rm debug --image=nicolaka/netshoot -n yao-oracle -- \
  nslookup yao-oracle-node-0.yao-oracle-node.yao-oracle.svc.cluster.local
```

### 配置更新不生效

**症状：** 更新 Helm values 后服务仍使用旧配置

**原因：** ConfigMap watching 未启用或 Pod 未重启

**解决方法：**
```bash
# 如果 configWatch.enabled = false，需要重启 Pod
kubectl rollout restart deployment -n yao-oracle yao-oracle-proxy
kubectl rollout restart deployment -n yao-oracle yao-oracle-dashboard

# 如果 configWatch.enabled = true，检查 ConfigMap
kubectl get configmap -n yao-oracle yao-oracle-config -o yaml

# 查看 Proxy 日志确认配置重载
kubectl logs -n yao-oracle -l app.kubernetes.io/component=proxy | grep "Configuration updated"
```

## 📚 相关文档

- [Helm Chart 结构](./README.md)
- [配置指南](./CONFIG-GUIDE.md)
- [变更日志](./CHANGELOG.md)
- [Dashboard 功能说明](../../docs/dashboard.md)

## 🆘 获取帮助

- GitHub Issues: https://github.com/eggybyte/yao-oracle/issues
- 文档: https://github.com/eggybyte/yao-oracle/docs

