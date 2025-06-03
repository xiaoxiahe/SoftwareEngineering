# 充电桩模拟器 (Simulator)

这是电动汽车充电桩管理系统的充电桩模拟器，基于 Go 语言开发，用于模拟真实充电桩的工作状态和充电过程。

## 技术栈

- **语言**: Go 1.24.3
- **架构**: 微服务架构
- **通信**: HTTP RESTful API
- **配置**: JSON 配置文件
- **日志**: 结构化日志记录

## 项目结构

```
simulator/
├── cmd/
│   └── main.go            # 主程序入口
├── configs/
│   ├── simulator.json     # 标准配置文件
│   └── simulator.docker.json # Docker环境配置
├── internal/
│   ├── config/           # 配置管理
│   ├── models/           # 数据模型
│   ├── services/         # 服务层
│   ├── simulator/        # 模拟器核心逻辑
│   └── utils/            # 工具函数
└── go.mod               # Go模块定义
```

## 核心功能

### 充电桩模拟

- 支持快充桩(F1, F2)和慢充桩(T1, T2, T3)
- 模拟充电桩的各种状态：可用、充电中、故障、维护
- 实时充电进度更新和电量计算
- 充电完成自动通知

### 故障模拟

- 随机故障生成机制
- 可配置故障概率和类型
- 故障恢复模拟
- 故障状态上报

### 与后端系统集成

- 接收后端分配的充电任务
- 实时上报充电进度
- 充电完成状态通知
- 故障状态同步

### 模拟场景

- 自动生成模拟充电请求
- 多种充电模式支持（快充/慢充）
- 可配置的模拟参数

## 快速开始

### 环境要求

- Go 1.24.3+
- 后端服务运行中

### 安装依赖

```bash
go mod download
```

### 配置文件

修改 `configs/simulator.json` 配置：

```json
{
  "backend": {
    "url": "http://localhost:8080",
    "timeout": 30
  },
  "simulator": {
    "updateInterval": 10,
    "logLevel": "info"
  },
  "piles": {
    "fastCharging": {
      "count": 2,
      "power": 7.0,
      "prefix": "F"
    },
    "trickleCharging": {
      "count": 3,
      "power": 3.5,
      "prefix": "T"
    }
  },
  "fault": {
    "randomFault": true,
    "faultChance": 5.0,
    "meanRepairTime": 300
  },
  "simulation": {
    "autoGenerate": true,
    "requestInterval": 60,
    "maxConcurrentUsers": 10
  }
}
```

### 启动模拟器

```bash
go run cmd/main.go
```

或使用自定义配置：

```bash
go run cmd/main.go -config configs/simulator.docker.json
```

或指定后端地址：

```bash
go run cmd/main.go -backend http://192.168.1.100:8080
```

## 命令行参数

- `-config` : 指定配置文件路径 (默认: configs/simulator.json)
- `-backend` : 指定后端 API 地址 (覆盖配置文件中的设置)
- `-help` : 显示帮助信息

## 工作原理

### 充电桩状态机

```
可用 → 充电中 → 完成 → 可用
  ↓      ↓
故障 → 维修中 → 可用
```

### 充电过程模拟

1. **接收充电任务**: 从后端接收充电分配请求
2. **开始充电**: 更新充电桩状态为充电中
3. **进度更新**: 每 10 秒更新一次充电进度
4. **电量计算**: 根据充电功率和时间计算电量增长
5. **完成通知**: 充电完成后通知后端系统

### 故障模拟

- 在充电过程中随机触发故障
- 故障概率可配置(默认 5%)
- 故障后自动进入维修状态
- 维修完成后恢复可用状态

## API 接口

模拟器作为 HTTP 客户端，调用后端 API：

### 充电分配响应

- `POST /api/v1/simulator/assign-charging` - 接收充电任务

### 状态上报

- `POST /api/v1/simulator/charging-progress` - 上报充电进度
- `POST /api/v1/simulator/charging-complete` - 上报充电完成
- `POST /api/v1/simulator/pile-fault` - 上报充电桩故障

## 数据模型

### 充电桩模型

```go
type Pile struct {
    ID       string      // 充电桩ID (F1, F2, T1, T2, T3)
    Type     PileType    // 充电桩类型 (fast/trickle)
    Power    float64     // 充电功率 (kW)
    Status   PileStatus  // 当前状态
    Vehicle  *ChargingVehicle // 当前充电车辆
}
```

### 充电车辆模型

```go
type ChargingVehicle struct {
    UserID            string    // 用户ID
    StartTime         time.Time // 开始充电时间
    RequestedCapacity float64   // 请求充电电量
    CurrentCapacity   float64   // 当前已充电量
    ChargingMode      string    // 充电模式
}
```

## 配置说明

### 后端连接配置

```json
{
  "backend": {
    "url": "http://localhost:8080", // 后端API地址
    "timeout": 30 // 请求超时时间(秒)
  }
}
```

### 充电桩配置

```json
{
  "piles": {
    "fastCharging": {
      "count": 2, // 快充桩数量
      "power": 7.0, // 充电功率(kW)
      "prefix": "F" // ID前缀
    },
    "trickleCharging": {
      "count": 3, // 慢充桩数量
      "power": 3.5, // 充电功率(kW)
      "prefix": "T" // ID前缀
    }
  }
}
```

### 故障模拟配置

```json
{
  "fault": {
    "randomFault": true, // 启用随机故障
    "faultChance": 5.0, // 故障概率(%)
    "meanRepairTime": 300 // 平均维修时间(秒)
  }
}
```

### 自动模拟配置

```json
{
  "simulation": {
    "autoGenerate": true, // 启用自动生成请求
    "requestInterval": 60, // 请求间隔(秒)
    "maxConcurrentUsers": 10 // 最大并发用户数
  }
}
```

## 日志记录

模拟器提供详细的日志记录：

```
2024-05-31 10:00:00 [INFO] 模拟器启动成功
2024-05-31 10:00:01 [INFO] 充电桩初始化完成: F1, F2, T1, T2, T3
2024-05-31 10:00:02 [INFO] 与后端连接建立: http://localhost:8080
2024-05-31 10:05:30 [INFO] 用户 user123 开始在充电桩 F1 充电
2024-05-31 10:05:40 [INFO] 充电桩 F1 进度更新: 10.5kWh/30.0kWh
2024-05-31 10:15:20 [INFO] 充电桩 F1 充电完成: 30.0kWh
2024-05-31 10:20:15 [WARN] 充电桩 T2 发生故障: 电源异常
```

## 监控和调试

### 状态查询

模拟器运行时会在控制台显示当前状态：

- 所有充电桩的实时状态
- 当前充电任务进度
- 故障和维修状态

### 调试模式

设置日志级别为 `debug` 可获得更详细的调试信息：

```json
{
  "simulator": {
    "logLevel": "debug"
  }
}
```

## 性能特性

- **低延迟**: 实时状态更新，延迟小于 1 秒
- **高并发**: 支持多个充电桩同时工作
- **容错性**: 网络异常时自动重试
- **资源效率**: 内存占用小，CPU 使用率低

## 故障排除

常见问题：

1. **连接后端失败**: 检查后端服务状态和网络连接
2. **配置文件错误**: 验证 JSON 格式和必需字段
3. **充电任务不响应**: 检查后端 API 接口和模拟器日志
4. **故障模拟异常**: 检查故障配置参数

## Docker 部署

### 单独部署模拟器

模拟器目录包含独立的 docker-compose.yml，可以单独部署：

```bash
cd simulator

# 启动模拟器服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看实时日志
docker-compose logs -f simulator
```

### 构建说明

Docker 部署使用多阶段构建：

1. **构建阶段**: 使用 Go 1.24.3 编译模拟器程序
2. **运行阶段**: 使用 Alpine Linux 轻量级镜像运行

### 环境变量

Docker 部署使用以下配置：

```bash
# 运行环境
ENVIRONMENT=production
LOG_LEVEL=info

# 后端连接
BACKEND_URL=http://localhost:8080
```

### 访问地址

部署后的访问地址：

- **模拟器监控**: <http://localhost:8081>
- **健康检查**: <http://localhost:8081/health>

### 配置文件

模拟器配置文件通过 Docker 卷挂载：

```bash
# 配置文件位置
./configs:/app/configs
```

可以修改 `configs/simulator.json` 来调整模拟器参数。

### 健康检查

模拟器服务包含健康检查：

```bash
# 检查服务健康状态
curl http://localhost:8081/health
```

### 常用命令

```bash
# 重新构建模拟器镜像
docker-compose build simulator

# 查看详细日志
docker-compose logs -f simulator

# 重启模拟器
docker-compose restart simulator

# 停止服务
docker-compose down
```

### 多实例部署

可以通过修改 docker-compose.yml 部署多个模拟器实例：

```yaml
simulator:
  # ...existing config...
  deploy:
    replicas: 3 # 运行3个实例
```

### 与后端集成

确保后端服务已启动：

```bash
# 先启动后端
cd ../backend
docker-compose up -d

# 再启动模拟器
cd ../simulator
docker-compose up -d
```
