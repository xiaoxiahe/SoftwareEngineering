# 后端服务 (Backend)

这是电动汽车充电桩管理系统的后端服务，基于 Go 语言开发，提供充电桩管理、用户管理、充电调度和计费等核心功能。

## 技术栈

- **语言**: Go 1.24.3
- **数据库**: PostgreSQL
- **主要依赖**:
  - `github.com/golang-jwt/jwt` - JWT 认证
  - `github.com/golang-migrate/migrate/v4` - 数据库迁移
  - `github.com/lib/pq` - PostgreSQL 驱动
  - `golang.org/x/crypto` - 密码加密

## 项目结构

```
backend/
├── cmd/server/          # 主程序入口
├── configs/             # 配置文件
├── internal/            # 内部代码包
│   ├── api/            # API路由和处理器
│   ├── config/         # 配置管理
│   ├── database/       # 数据库连接
│   ├── middleware/     # 中间件
│   ├── model/          # 数据模型
│   ├── repository/     # 数据访问层
│   └── service/        # 业务逻辑层
└── migrations/         # 数据库迁移文件
```

## 核心功能

### 充电桩管理

- 支持快充桩和慢充桩两种类型
- 充电桩状态监控（可用、占用、故障、维护、离线）
- 故障报告和维修管理
- 统计信息跟踪（充电次数、时长、电量）

### 充电调度

- 智能充电桩分配算法
- 排队机制和等待时间预估
- 优先级调度支持
- 实时充电会话管理

### 用户管理

- 用户注册和认证
- JWT 令牌验证
- 用户充电历史记录

### 计费系统

- 基于时长和电量的计费模式
- 动态价格配置
- 详单生成和管理

## 快速开始

### 环境要求

- Go 1.24.3+
- PostgreSQL 12+

### 安装依赖

```bash
go mod download
```

### 配置数据库

1. 创建 PostgreSQL 数据库
2. 修改 `configs/config.json` 中的数据库连接配置
3. 运行数据库迁移：

```bash
go run cmd/server/main.go -migrate
```

### 启动服务

```bash
go run cmd/server/main.go
```

或编译后运行：

```bash
go build -o main.exe cmd/server/main.go
./main.exe
```

## API 接口

### 认证相关

- `POST /api/v1/auth/register` - 用户注册
- `POST /api/v1/auth/login` - 用户登录
- `GET /api/v1/auth/profile` - 获取用户信息

### 充电桩相关

- `GET /api/v1/charging-piles` - 获取所有充电桩
- `PUT /api/v1/charging-piles/{id}/control` - 控制充电桩状态
- `GET /api/v1/charging-piles/queue` - 获取排队信息

### 充电服务

- `POST /api/v1/charging/request` - 提交充电请求
- `GET /api/v1/charging/queue-status` - 查询排队状态
- `POST /api/v1/charging/cancel` - 取消充电请求

### 模拟器接口

- `POST /api/v1/simulator/assign-charging` - 分配充电任务
- `POST /api/v1/simulator/charging-complete` - 上报充电完成
- `POST /api/v1/simulator/charging-progress` - 上报充电进度

## 配置说明

配置文件位于 `configs/config.json`：

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
    "dbname": "charging_system",
    "sslmode": "disable"
  },
  "charging": {
    "fastChargingPileNum": 2,
    "trickleChargingPileNum": 3,
    "fastChargingPower": 7.0,
    "trickleChargingPower": 3.5,
    "maxQueueLength": 10
  }
}
```

## 数据库架构

系统包含以下主要数据表：

- `users` - 用户信息
- `charging_piles` - 充电桩信息
- `charging_requests` - 充电请求
- `charging_sessions` - 充电会话
- `queue_status` - 排队状态
- `billing_details` - 计费详单
- `fault_records` - 故障记录
- `system_config` - 系统配置

## 部署

1. 编译项目：

```bash
go build -o charging-backend cmd/server/main.go
```

2. 设置环境变量或配置文件
3. 运行数据库迁移
4. 启动服务

## 故障排除

常见问题：

1. **数据库连接失败**: 检查数据库配置和连接状态
2. **端口占用**: 修改配置中的端口号
3. **迁移失败**: 检查数据库权限和表结构

## Docker 部署

### 单独部署后端服务

后端目录包含独立的 docker-compose.yml，可以单独部署后端服务及其依赖：

```bash
cd backend

# 启动后端服务、PostgreSQL 和 Redis
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f backend
```

### 服务说明

部署后包含以下服务：

- **backend**: Go 后端服务 (端口: 8080)
- **postgres**: PostgreSQL 数据库 (端口: 5432)
- **redis**: Redis 缓存 (端口: 6379)

### 环境变量

Docker 部署使用以下默认配置：

```bash
# 数据库配置
DB_HOST=postgres
DB_PORT=5432
DB_NAME=ev_charging
DB_USER=ev_user
DB_PASSWORD=secure_password123

# JWT配置
JWT_SECRET=your_super_secret_jwt_key_here_change_in_production

# 应用配置
ENVIRONMENT=production
LOG_LEVEL=info
```

### 健康检查

后端服务包含健康检查端点：

```bash
# 检查服务健康状态
curl http://localhost:8080/health
```

### 数据持久化

PostgreSQL 和 Redis 数据会持久化到 Docker 卷中：

- `postgres_data`: PostgreSQL 数据
- `redis_data`: Redis 数据

### 常用命令

```bash
# 重新构建后端镜像
docker-compose build backend

# 查看数据库日志
docker-compose logs -f postgres

# 连接到数据库
docker-compose exec postgres psql -U ev_user -d ev_charging

# 清理所有数据
docker-compose down -v
```
