## 🧭 一、整体页面结构（建议）

```
Dashboard
├── Overview（总览页）
├── Proxy Monitor（Proxy 监控页）
├── Node Monitor（Node 监控页）
├── Namespace & Key Query（命名空间/Key 查询页）
├── Events & Health（事件/健康状态页）
```

每个页面都从 **Admin 微服务** 获取数据（REST + WebSocket）：

* WebSocket：实时推送（overview、proxy_update、node_update、event）
* REST：历史数据查询、分页查询、Key 内容查询等

---

## 📡 二、Dashboard 所需数据（来自 Admin）

| 类别       | 数据项                                    | 来源                  |
| -------- | -------------------------------------- | ------------------- |
| 集群总览     | Proxy 数量、Node 数量、总 QPS、整体命中率、整体延迟      | Admin 聚合 Proxy/Node |
| Proxy 状态 | 每个 Proxy 的 QPS、延迟（p50/p90/p99）、错误率、连接数 | Proxy → Admin       |
| Node 状态  | 内存使用、Key 总数、命中/未命中次数、热点 Key            | Node → Admin        |
| 时间序列     | QPS 曲线、延迟曲线、命中率曲线、内存使用曲线               | Proxy/Node → Admin  |
| Key 查询   | namespace、key 的当前值、TTL、命中统计、所在 Node    | Admin → Node/Proxy  |
| 事件流      | 节点上下线、Proxy 异常、命中率突变等                  | Admin event stream  |

---

## 📈 三、图表类型与用途

| 图表类型               | 使用位置                       | 说明                                   |
| ------------------ | -------------------------- | ------------------------------------ |
| 🟦 **数字卡片**        | Overview                   | 展示核心指标：Proxy 数量、Node 数量、总 QPS、整体命中率等 |
| 📈 **折线图**         | Proxy / Node 监控页           | 展示 QPS、延迟（p50/p90/p99）、命中率、内存使用的时间序列 |
| 🕸 **饼图 / 环形图**    | Overview / Node页           | 展示命中率分布、各命名空间占比                      |
| 📊 **柱状图**         | Proxy页 / Node页             | 展示热点 Key 排行，或命名空间访问量                 |
| 🌡 **仪表盘图（Gauge）** | Overview / Proxy页          | 展示实时 QPS 峰值占比、整体健康度、命中率              |
| 🧾 **表格**          | Proxy列表 / Node列表 / Key 查询页 | 展示详细状态、支持排序/过滤/点击下钻                  |

推荐使用：

* 📌 **Chart.js**（轻量，Vite + React 集成方便）
* 或 ECharts（如果你需要更强大的地理/交互）

---

## 📋 四、各页面详细内容与功能

### 🟣 1️⃣ **Overview 总览页**

> 目标：一眼看到整个集群的健康状态 & 核心指标

#### 📊 数据展示：

* Proxy 总数、Node 总数
* 全局 QPS（get / set / delete）
* 总体命中率
* 平均延迟（p50/p90/p99）
* 总内存使用（所有 Node 汇总）

#### 📈 图表：

* QPS 总曲线（折线图）
* 命中率曲线（折线图）
* 命名空间访问量占比（饼图）
* 总内存使用变化（折线图）

#### 🧠 功能：

* 实时自动刷新（WebSocket）
* 历史回放（时间区间选择 → 调用 REST）
* 各指标鼠标悬停显示 tooltip

---

### 🟩 2️⃣ **Proxy Monitor 页面**

> 目标：观察每个 Proxy 的请求负载、延迟、错误情况

#### 📊 数据展示（表格）：

| Proxy ID | QPS(get/set/del) | 延迟(p50/p90/p99) | 错误率 | 连接数 | 状态 |
| -------- | ---------------- | --------------- | --- | --- | -- |

#### 📈 图表：

* 每个 Proxy 的 QPS 曲线（可选 Proxy ID）
* 延迟曲线（p50/p90/p99）
* 错误率曲线

#### 🧠 功能：

* 点击 Proxy ID → 下钻查看历史趋势（调用 REST）
* 多选 Proxy 对比曲线
* 健康状态颜色标识（绿色=正常，红色=异常）

---

### 🟦 3️⃣ **Node Monitor 页面**

> 目标：监控存储层的使用情况、命中率与热点 Key

#### 📊 数据展示（表格）：

| Node ID | Key 数量 | 内存使用 | Hit Count | Miss Count | 命中率 | 状态 |
| ------- | ------ | ---- | --------- | ---------- | --- | -- |

#### 📈 图表：

* 内存使用曲线（按 Node）
* 命中率曲线（按 Node）
* 热点 Key 柱状图（Top N）
* 命名空间占比饼图（可选）

#### 🧠 功能：

* 点击 Node → 查看热点 Key & 历史命中率
* 实时刷新（WebSocket）
* 选定时间范围查看历史趋势

---

### 🟡 4️⃣ **Namespace & Key 查询页面**

> 目标：提供精确的 KV 查询 & 命名空间分析

#### 🔍 查询表单：

* Namespace（下拉框自动补全）
* Key（文本框）
* 模糊匹配选项（前缀 / contains）

#### 📊 查询结果：

| Key | Value（截断显示） | TTL | 所在 Node | Hit Count | Miss Count | 上次访问时间 |

#### 🧠 功能：

* 支持分页、排序、导出（CSV）
* 点击 Key → 查看详细（弹窗显示完整 Value / JSON）
* 实时搜索（前端过滤 + 后端分页）
* 命名空间下热点 Key 列表（Top N）

---

### 🟠 5️⃣ **Events & Health 页面**

> 目标：追踪系统运行中的异常事件、上下线、性能波动

#### 📊 数据：

* 节点上下线事件
* Proxy 异常（如错误率上升）
* 命中率异常波动

#### 📝 展示方式：

* 滚动事件列表（时间倒序）
* 按类型筛选（Node/Proxy/Global）
* 点击事件 → 跳转相关 Node/Proxy 页面

---

## 🧠 五、交互逻辑与数据同步方式

### 🌐 与 Admin API 对接

| 类型        | 用途                                                  | 特点                |
| --------- | --------------------------------------------------- | ----------------- |
| WebSocket | 实时更新 overview / proxy_update / node_update / events | Admin 主动推送，前端更新状态 |
| REST      | 历史查询、分页、Key 查询                                      | 用户交互时触发一次性请求      |

例如：

```ts
// WebSocket
const ws = new WebSocket("/ws");
ws.onmessage = (e) => {
  const msg = JSON.parse(e.data);
  switch (msg.type) {
    case "overview": updateOverview(msg.data); break;
    case "proxy_update": updateProxy(msg.data); break;
    case "node_update": updateNode(msg.data); break;
    case "event": addEvent(msg.data); break;
  }
};

// REST 查询 Key
fetch(`/api/query?ns=${ns}&key=${key}`).then(r => r.json());
```

---

## 📝 六、图表与 UI 技术建议

| 组件   | 推荐库                               | 用途                  |
| ---- | --------------------------------- | ------------------- |
| 图表   | Chart.js                          | 折线图、柱状图、饼图、仪表盘      |
| 表格   | Ant Design Table / TanStack Table | Proxy/Node/Key 列表展示 |
| 样式   | Tailwind CSS                      | 快速布局 & 响应式          |
| 状态管理 | React Context + Hooks             | 轻量共享全局 Admin 状态     |

---

## ☸️ 七、部署与运维

* `npm run build` → 产出 `/dist`
* 用 `nginx:1.29.1-alpine` 容器 serve 静态文件
* 配置 `/api`、`/ws` 反向代理到 Admin
* 前后端独立部署在 K8s，Dashboard 为无状态微服务，可水平扩展

---

## ✅ 最终结果（Dashboard 能做到的）

| 功能点                | 是否覆盖 |
| ------------------ | ---- |
| 集群总览 & 健康状态        | ✅    |
| Proxy 监控（QPS/延迟）   | ✅    |
| Node 监控（内存/命中率/热点） | ✅    |
| Namespace & Key 查询 | ✅    |
| 实时事件流              | ✅    |
| 实时更新 + 历史查询        | ✅    |
| 独立前端微服务 + K8s 部署   | ✅    |

---

### 📌 TL;DR

> 📊 **Dashboard 实现内容 = 实时监控 + 历史查询 + Key 检索 + 健康事件**
> 使用 React + Vite + Nginx 实现独立前端，通过 WebSocket + REST 与 Admin 通信，展示 Proxy/Node 的各类 Metrics 与状态，是一个完整的轻量级集群可视化监控系统 ✅