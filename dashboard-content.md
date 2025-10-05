# Yao-Oracle Dashboard 已实现内容总结

## 📋 概述

Yao-Oracle Dashboard 是一个基于纯 HTML + JavaScript + Chart.js 的 Web 监控界面，提供集群状态监控、命名空间管理和缓存数据浏览功能。无需构建步骤，可直接嵌入 Go 二进制文件。

## 🎨 技术栈

- **HTML5**: 语义化标签
- **Vanilla JavaScript (ES6+)**: 模块化设计，无框架依赖
- **CSS3**: 现代样式系统，支持暗黑/明亮主题切换
- **Chart.js 4.x**: 轻量级图表库
- **Fetch API**: HTTP 请求处理

## 📁 文件结构

```
web/
├── login.html              # 登录页面（2 个 HTML 文件）
├── index.html              # 主监控面板
├── css/                    # 样式文件（4 个）
│   ├── common.css          # 共享样式和 CSS 变量
│   ├── login.css           # 登录页面样式
│   ├── dashboard.css       # 面板布局和组件
│   └── cache-explorer.css  # 缓存浏览器样式
├── js/                     # JavaScript 模块（6 个）
│   ├── config.js           # 配置常量
│   ├── auth.js             # 认证逻辑
│   ├── api.js              # API 客户端封装
│   ├── charts.js           # Chart.js 管理和配置
│   ├── dashboard.js        # 主面板逻辑
│   └── cache-explorer.js   # 缓存数据管理
└── assets/                 # 图标和图片资源
```

## 🔐 认证系统（login.html）

### 功能特性
- ✅ 密码认证登录
- ✅ 错误消息显示
- ✅ 测试模式支持（默认密码：`admin123`）
- ✅ 登录成功自动跳转到主面板
- ✅ Session 令牌管理（localStorage）
- ✅ 响应式设计

### 实现细节
- 使用 `POST /api/auth/login` 进行认证
- Session ID 存储在 `localStorage` 的 `yao-oracle-session` 键中
- 所有后续 API 请求自动携带 `Authorization: Bearer <token>` 头

## 📊 主监控面板（index.html）

### 页面布局
```
┌─────────────────────────────────────────────────────┐
│ Header: Logo | Yao-Oracle  [主题] [刷新] [登出]    │
├────────┬────────────────────────────────────────────┤
│ 导航栏 │                                            │
│ ├ 概览 │        Tab 内容区域                        │
│ ├ 命名 │        （动态切换）                        │
│ ├ 代理 │                                            │
│ ├ 节点 │                                            │
│ └ 缓存 │                                            │
└────────┴────────────────────────────────────────────┘
```

### 全局功能
- ✅ **主题切换**: 暗黑/明亮模式切换，偏好设置持久化
- ✅ **自动刷新**: 每 5 秒自动更新数据
- ✅ **手动刷新**: 刷新按钮强制更新所有数据
- ✅ **登出功能**: 清除 token 并跳转登录页
- ✅ **响应式设计**: 支持移动端、平板和桌面端

## 📈 Tab 1: 概览（Overview）

### 展示数据

#### 核心指标卡片（4 个）
1. **Namespaces**: 配置的命名空间数量
2. **Cache Nodes**: 缓存节点总数
3. **Total Keys**: 缓存 Key 总数
4. **Memory Usage**: 内存使用量和百分比

#### 可视化图表（4 个）

**1. Cache Hit Rate（缓存命中率）- 仪表盘图**
- 类型: 半圆仪表盘（Doughnut）
- 数据: 命中率百分比（0-100%）
- 颜色: 绿色（命中）vs 红色（未命中）

**2. QPS Trend（QPS 趋势）- 折线图**
- 类型: 时间序列折线图
- 数据: 最近 20 分钟的 QPS 数据
- X 轴: 时间标签（分钟）
- Y 轴: 每秒请求数

**3. Memory Distribution（内存分布）- 环形图**
- 类型: 环形图（Doughnut）
- 数据: 各命名空间内存使用量
- 显示: 命名空间名称 + 内存大小（MB）

**4. Response Time（响应时间）- 折线图**
- 类型: 时间序列折线图
- 数据: 最近 20 分钟的延迟数据
- X 轴: 时间标签
- Y 轴: 延迟时间（ms）

#### 集群健康状态表
- 组件名称（Proxy、Cache Nodes、Dashboard）
- 实例数量
- 健康状态（healthy/warning/error）
- 健康度百分比
- 运行时长（Uptime）
- 最后检查时间

## 🏷️ Tab 2: 命名空间（Namespaces）

### 展示数据
- **卡片布局**: 每个命名空间一张卡片
- **命名空间信息**:
  - 名称和描述
  - 总 Key 数量
  - 内存使用量（MB）
  - 缓存命中率（%）
  - QPS（每秒请求数）
  - 状态指示器（健康/警告/错误）

### 功能特性
- ✅ 网格布局展示所有命名空间
- ✅ 实时指标更新
- ✅ 状态色彩编码（绿/黄/红）

## 🔌 Tab 3: 代理（Proxies）

### 展示数据
- **表格布局**: 显示所有 Proxy 实例
- **列字段**:
  - Proxy ID/名称
  - IP 地址
  - 当前 QPS
  - 平均延迟（ms）
  - 错误率（%）
  - 健康状态
  - 运行时长

### 功能特性
- ✅ 表格排序支持
- ✅ 健康状态图标
- ✅ 实时数据更新

## 🗄️ Tab 4: 缓存节点（Nodes）

### 展示数据
- **卡片布局**: 每个节点一张卡片
- **节点信息**:
  - 节点 ID
  - IP 地址
  - 内存使用进度条（当前/最大）
  - 存储的 Key 数量
  - 命中次数
  - 未命中次数
  - 命中率百分比
  - 运行时长
  - 健康状态

### 功能特性
- ✅ 内存使用可视化进度条
- ✅ 节点健康状态监控
- ✅ 详细统计信息

## 🔍 Tab 5: 缓存浏览器（Cache Explorer）

### 功能特性

#### 1. 命名空间选择
- 下拉菜单显示所有配置的命名空间
- 显示命名空间名称和描述
- 选择后加载该命名空间的缓存数据

#### 2. Key 列表（分页）
- **显示字段**:
  - Key 名称
  - TTL 剩余时间
  - 最后修改时间
- **分页控制**: 每页 50 条记录
- **搜索功能**: 按 Key 前缀过滤

#### 3. Key 详情查看
- 点击 Key 查看详细信息
- **显示内容**:
  - Key 名称
  - Value 值（JSON 格式化）
  - TTL 剩余时间
  - 数据大小（bytes）
  - 创建时间
- JSON 值语法高亮显示

### 交互流程
```
选择命名空间 → 加载 Key 列表 → 搜索/分页 → 点击 Key → 查看详情
```

## 🎯 Mock 数据模式（测试用）

### 测试配置
- **TEST_MODE**: `true` - 启用测试模式
- **DEFAULT_PASSWORD**: `admin123` - 默认登录密码
- **useMockData**: `true` - 使用模拟数据

### 模拟数据内容
- **4 个命名空间**: game-app, ads-service, user-cache, api-cache
- **3 个 Proxy 实例**: 模拟负载和健康状态
- **6 个 Cache Node**: 不同内存和 Key 分布
- **时间序列数据**: 20 个数据点的历史趋势
- **健康状态**: 所有组件状态模拟

### 测试优势
- ✅ 无需后端即可测试完整 UI
- ✅ 快速验证设计和交互
- ✅ 前端独立开发和调试

## 🔌 后端 API 接口要求

### 认证接口
- `POST /api/auth/login` - 用户登录
- `POST /api/auth/logout` - 用户登出

### 监控数据接口
- `GET /api/dashboard/cluster-status` - 集群概览数据
- `GET /api/dashboard/namespaces` - 命名空间列表
- `GET /api/dashboard/proxies` - Proxy 实例列表
- `GET /api/dashboard/nodes` - Cache Node 列表

### 缓存浏览接口
- `GET /api/cache/namespaces` - 可用命名空间列表
- `GET /api/cache/keys?namespace={ns}&apikey={key}&page={n}&prefix={p}` - Key 列表（分页）
- `GET /api/cache/value?namespace={ns}&apikey={key}&key={k}` - Key 值查询

## 🎨 样式和主题

### 设计系统
- **CSS 变量**: 定义在 `common.css` 中的设计令牌
- **颜色方案**: 
  - Primary: #667eea（紫色渐变）
  - Success: #48bb78（绿色）
  - Warning: #ed8936（橙色）
  - Error: #f56565（红色）
  - Info: #4299e1（蓝色）

### 主题支持
- ✅ 暗黑模式（默认）
- ✅ 明亮模式
- ✅ 主题偏好持久化
- ✅ 图表颜色自动适配主题

### 响应式断点
- 移动端: 320px+
- 平板: 768px+
- 桌面: 1024px+

## ⚡ 性能特性

- **轻量级**: 总页面大小 < 500KB（不含 Chart.js CDN）
- **模块化**: 每个 JS 文件 < 500 行
- **Chart 管理**: 防止内存泄漏的实例销毁机制
- **懒加载**: 按需加载 Tab 数据
- **并发控制**: API 请求去重和缓存

## 🚀 部署方式

### Go Embed 集成
```go
//go:embed web
var webFS embed.FS

// 嵌入到 Go 二进制文件中
```

### 独立测试
```bash
# 使用 Python 或 Node.js 启动本地服务器
python3 -m http.server 8080
# 或
npx serve web -p 8080
```

## 📝 已完成的工作

### ✅ 页面实现
- [x] 登录页面（login.html）
- [x] 主监控面板（index.html）

### ✅ 样式实现
- [x] 共享样式系统（common.css）
- [x] 登录页面样式（login.css）
- [x] 面板布局样式（dashboard.css）
- [x] 缓存浏览器样式（cache-explorer.css）

### ✅ JavaScript 模块
- [x] 配置管理（config.js）
- [x] 认证逻辑（auth.js）
- [x] API 客户端（api.js）
- [x] 图表管理（charts.js）
- [x] 主面板逻辑（dashboard.js）
- [x] 缓存浏览器（cache-explorer.js）

### ✅ 功能特性
- [x] 密码认证
- [x] Session 管理
- [x] 5 个监控 Tab
- [x] 4 种可视化图表
- [x] 暗黑/明亮主题切换
- [x] 自动刷新机制
- [x] 缓存数据浏览
- [x] Mock 数据测试模式
- [x] 响应式设计

## 🎯 下一步工作

### 待完成
- [ ] 后端 API 实现并连接
- [ ] 关闭 TEST_MODE 和 Mock 数据
- [ ] WebSocket 实时推送优化
- [ ] 更多图表类型（柱状图、饼图等）
- [ ] 导出数据功能
- [ ] 告警通知功能

---

**文档生成时间**: 2025-10-04  
**项目**: Yao-Oracle Distributed KV Cache System  
**版本**: v1.0

