# Yao-Oracle Dashboard

前端监控面板，用于实时监控 Yao-Oracle 分布式缓存集群的运行状态。

## 技术栈

- **React 19** - UI 框架
- **TypeScript** - 类型安全
- **Vite** - 构建工具
- **ECharts** - 图表库
- **Zustand** - 状态管理
- **React Router** - 路由管理
- **Nginx 1.27** - 生产环境服务器

## 功能特性

### 1. 集群总览 (Overview)
- 业务空间数量统计
- Proxy 和 Node 实例健康状态
- 实时 QPS 监控
- 缓存命中率仪表盘
- 请求类型分布饼图
- 平均延迟和总键值数量

### 2. Proxy 实例监控 (Proxies)
- 实时 Proxy 列表和状态
- 每个 Proxy 的活跃连接数
- QPS 时序图 (GET/SET/DELETE)
- 延迟统计 (P50/P90/P99)
- 错误率监控

### 3. Cache Node 监控 (Nodes)
- Node 实例列表和状态
- 内存使用情况
- 缓存键值数量
- 命中率统计
- 热点 Key 排行榜

### 4. 业务空间管理 (Namespaces)
- 业务空间配置信息
- 每个空间的内存占用
- QPS 分布对比
- 命中率统计
- 资源限制配置

## 开发环境

### 前置条件

- Node.js 20+
- npm 或 yarn

### 安装依赖

```bash
cd dashboard
npm install
```

### 启动开发服务器

```bash
npm run dev
```

开发服务器将在 `http://localhost:5173` 启动。

### 启动模拟 Admin 服务

在另一个终端窗口中：

```bash
# 从项目根目录运行
./scripts/run-mock-admin.sh
```

模拟 Admin 服务将在 `http://localhost:8081` 启动，提供：
- REST API: `http://localhost:8081/api`
- WebSocket: `ws://localhost:8081/ws`

### 一键启动测试环境

```bash
# 从项目根目录运行
cd dashboard
npm run dev &
cd ..
./scripts/run-mock-admin.sh
```

然后访问 `http://localhost:5173` 查看 Dashboard。

## 生产环境构建

### 构建前端

```bash
npm run build
```

构建产物将输出到 `dist/` 目录。

### Docker 构建

```bash
# 从项目根目录
docker build -f build/dashboard.Dockerfile -t yao-oracle/dashboard:latest .
```

### Docker 运行

```bash
docker run -p 8080:80 \
  -e ADMIN_SERVICE_URL=http://admin-service:8081 \
  yao-oracle/dashboard:latest
```

## 环境变量

### 开发环境 (`.env.development`)

```env
VITE_ADMIN_URL=http://localhost:8081/api
VITE_ADMIN_WS_URL=ws://localhost:8081/ws
VITE_APP_TITLE=Yao-Oracle Dashboard (Dev)
```

### 生产环境 (`.env.production`)

```env
VITE_ADMIN_URL=/api
VITE_ADMIN_WS_URL=/ws
VITE_APP_TITLE=Yao-Oracle Dashboard
```

生产环境中，Nginx 会将 `/api` 和 `/ws` 请求代理到 Admin 服务。

## 项目结构

```
dashboard/
├── src/
│   ├── api/                  # API 客户端
│   │   ├── client.ts         # REST API
│   │   └── websocket.ts      # WebSocket 客户端
│   ├── components/           # 通用组件
│   │   ├── MetricCard.tsx    # 指标卡片
│   │   ├── StatusBadge.tsx   # 状态徽章
│   │   └── charts/           # 图表组件
│   │       ├── QPSChart.tsx
│   │       ├── GaugeChart.tsx
│   │       ├── BarChart.tsx
│   │       └── PieChart.tsx
│   ├── pages/                # 页面组件
│   │   ├── Overview.tsx      # 总览页
│   │   ├── Proxies.tsx       # Proxy 监控页
│   │   ├── Nodes.tsx         # Node 监控页
│   │   └── Namespaces.tsx    # 业务空间页
│   ├── stores/               # 状态管理
│   │   └── metricsStore.ts   # 指标数据存储
│   ├── types/                # TypeScript 类型
│   │   └── metrics.ts        # 指标数据类型
│   ├── App.tsx               # 根组件
│   ├── App.css               # 全局样式
│   └── main.tsx              # 入口文件
├── public/                   # 静态资源
├── nginx.conf                # Nginx 配置
├── package.json              # 依赖配置
├── tsconfig.json             # TypeScript 配置
├── vite.config.ts            # Vite 配置
└── README.md                 # 本文档
```

## API 接口

Dashboard 与 Admin 服务通信，使用以下 API:

### REST API

- `GET /api/overview` - 集群总览
- `GET /api/proxies` - Proxy 列表
- `GET /api/proxies/:id` - Proxy 详情
- `GET /api/proxies/:id/timeseries` - Proxy 时序数据
- `GET /api/nodes` - Node 列表
- `GET /api/nodes/:id` - Node 详情
- `GET /api/nodes/:id/timeseries` - Node 时序数据
- `GET /api/namespaces` - 业务空间列表
- `GET /api/namespaces/:name` - 业务空间详情
- `GET /api/health` - 健康检查

### WebSocket

连接到 `/ws` 接收实时更新：

```typescript
{
  "type": "overview_update" | "proxy_update" | "node_update" | "event",
  "data": { ... }
}
```

## 性能优化

- ✅ 代码分割和懒加载
- ✅ 图表实例复用，避免内存泄漏
- ✅ WebSocket 自动重连
- ✅ API 请求防抖
- ✅ 静态资源缓存 (1年)
- ✅ Gzip 压缩

## 浏览器支持

- Chrome (最新版)
- Firefox (最新版)
- Safari (最新版)
- Edge (最新版)

## 故障排除

### 无法连接到 Admin 服务

1. 确认 Admin 服务正在运行
2. 检查 `.env.development` 中的 URL 配置
3. 检查浏览器控制台的网络请求

### WebSocket 连接失败

1. 确认 Admin 服务的 WebSocket 端点可访问
2. 检查防火墙设置
3. 查看浏览器开发者工具的 WebSocket 连接状态

### 图表不显示

1. 清除浏览器缓存
2. 检查数据是否正确加载 (Network 标签)
3. 查看控制台错误日志

## 许可证

MIT License
