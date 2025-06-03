# 电动汽车充电桩管理系统

一个完整的电动汽车充电桩管理系统，包含后端服务、前端用户界面和充电桩模拟器。系统支持充电桩管理、用户充电服务、智能调度、实时监控和计费管理等功能。

## 系统架构

```
┌─────────────────┐    ┌─────────────────┐    ┌─────────────────┐
│   前端应用       │    │    后端服务      │    │   充电桩模拟器   │
│   (SvelteKit)   │◄──►│   (Go + PG)     │◄──►│    (Go)        │
│                 │    │                 │    │                 │
│ • 用户界面       │    │ • API服务       │    │ • 充电桩模拟     │
│ • 管理控制台     │    │ • 业务逻辑       │    │ • 状态上报       │
│ • 实时监控       │    │ • 数据存储       │    │ • 故障模拟       │
└─────────────────┘    └─────────────────┘    └─────────────────┘
```

## 技术栈

### 后端 (Backend)

- **语言**: Go 1.24.3
- **数据库**: PostgreSQL
- **架构**: Clean Architecture
- **认证**: JWT
- **API**: RESTful

### 前端 (Frontend)

- **框架**: SvelteKit 2.21.1
- **语言**: TypeScript
- **样式**: Tailwind CSS
- **构建**: Vite
- **包管理**: Bun

### 模拟器 (Simulator)

- **语言**: Go 1.24.3
- **通信**: HTTP Client
- **配置**: JSON
- **日志**: 结构化日志

## 核心功能

### 🔌 充电桩管理

- 支持快充桩和慢充桩
- 实时状态监控（可用、占用、故障、维护、离线）
- 故障报告和维修管理
- 统计信息跟踪

### 👥 用户管理

- 用户注册和认证
- 个人信息管理
- 充电历史记录
- 多角色权限控制

### ⚡ 充电服务

- 智能充电桩分配
- 排队机制和等待时间预估
- 实时充电监控
- 充电会话管理

### 📊 数据分析

- 使用统计报表
- 收入分析
- 设备利用率分析
- 用户行为分析

### 💰 计费系统

- 灵活的计费模式
- 实时费用计算
- 详单生成
- 支付集成准备

## 快速开始

### 环境要求

- Go 1.24.3+
- Node.js 18+
- PostgreSQL 12+
- Bun (推荐) 或 npm/yarn

### 1. 克隆项目

```bash
git clone <repository-url>
cd my-app
```

### 2. 启动后端服务

```bash
cd backend
go mod download
# 配置数据库连接 (configs/config.json)
go run cmd/server/main.go
```

### 3. 启动前端应用

```bash
cd frontend
bun install
bun dev
```

### 4. 启动充电桩模拟器

```bash
cd simulator
go mod download
go run cmd/main.go
```

### 5. 访问应用

- 前端应用: <http://localhost:5173>
- 后端 API: <http://localhost:8080>

## 项目结构

```
my-app/
├── backend/                 # 后端服务
│   ├── cmd/server/         # 主程序入口
│   ├── configs/            # 配置文件
│   ├── internal/           # 内部代码
│   │   ├── api/           # API路由和处理器
│   │   ├── config/        # 配置管理
│   │   ├── database/      # 数据库连接
│   │   ├── middleware/    # 中间件
│   │   ├── model/         # 数据模型
│   │   ├── repository/    # 数据访问层
│   │   └── service/       # 业务逻辑层
│   ├── migrations/        # 数据库迁移
│   └── README.md
├── frontend/               # 前端应用
│   ├── src/
│   │   ├── lib/           # 共享库
│   │   ├── routes/        # 页面路由
│   │   └── ...
│   ├── static/            # 静态文件
│   └── README.md
├── simulator/              # 充电桩模拟器
│   ├── cmd/               # 主程序入口
│   ├── configs/           # 配置文件
│   ├── internal/          # 内部代码
│   │   ├── config/        # 配置管理
│   │   ├── models/        # 数据模型
│   │   ├── services/      # 服务层
│   │   ├── simulator/     # 模拟器核心
│   │   └── utils/         # 工具函数
│   └── README.md
└── README.md              # 项目总览
```

## 数据库设计

### 核心数据表

- `users` - 用户信息
- `charging_piles` - 充电桩信息
- `charging_requests` - 充电请求
- `charging_sessions` - 充电会话
- `queue_status` - 排队状态
- `billing_details` - 计费详单
- `fault_records` - 故障记录
- `system_config` - 系统配置

### 关系图

```
Users ──┐
        ├── ChargingRequests ── ChargingSessions ── BillingDetails
        └── QueueStatus
                │
ChargingPiles ──┴── FaultRecords
```

## API 接口文档

### 认证服务

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/auth/profile` - 获取用户信息

### 充电桩管理

- `GET /api/v1/charging-piles` - 获取充电桩列表
- `PUT /api/v1/charging-piles/{id}/control` - 控制充电桩
- `GET /api/v1/charging-piles/queue` - 获取排队信息

### 充电服务

- `POST /api/v1/charging/request` - 提交充电请求
- `GET /api/v1/charging/queue-status` - 查询排队状态
- `POST /api/v1/charging/cancel` - 取消充电请求

### 模拟器接口

- `POST /api/v1/simulator/assign-charging` - 分配充电任务
- `POST /api/v1/simulator/charging-progress` - 上报充电进度
- `POST /api/v1/simulator/charging-complete` - 上报充电完成

## 配置说明

### 后端配置 (backend/configs/config.json)

```json
{
  "server": {
    "port": "8080",
    "host": "0.0.0.0"
  },
  "database": {
    "host": "localhost",
    "port": "5432",
    "user": "postgres",
    "password": "password",
    "dbname": "charging_system"
  },
  "charging": {
    "fastChargingPileNum": 2,
    "trickleChargingPileNum": 3,
    "fastChargingPower": 7.0,
    "trickleChargingPower": 3.5
  }
}
```

### 模拟器配置 (simulator/configs/simulator.json)

```json
{
  "backend": {
    "url": "http://localhost:8080"
  },
  "piles": {
    "fastCharging": {
      "count": 2,
      "power": 7.0
    },
    "trickleCharging": {
      "count": 3,
      "power": 3.5
    }
  }
}
```
