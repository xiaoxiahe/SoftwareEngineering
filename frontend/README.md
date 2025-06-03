# 前端应用 (Frontend)

这是电动汽车充电桩管理系统的前端应用，基于SvelteKit开发，提供用户友好的充电桩管理界面。

## 技术栈

- **框架**: SvelteKit 2.21.1
- **语言**: TypeScript
- **构建工具**: Vite
- **包管理器**: Bun
- **UI组件库**:
  - Tailwind CSS - 样式框架
  - bits-ui - UI组件库
  - shadcn/ui 风格组件
- **图表库**: LayerChart
- **表格组件**: TanStack Table
- **图标库**: Lucide Svelte

## 核心功能

### 用户功能

- 用户注册和登录
- 个人信息管理
- 充电请求提交
- 排队状态查询
- 充电历史记录
- 实时充电监控

### 管理员功能

- 充电桩状态监控
- 充电桩控制（启停、维护）
- 系统统计报表
- 用户管理
- 故障处理
- 计费管理

## 快速开始

### 环境要求

- Bun (推荐) 或 npm/yarn

### 安装依赖

```bash
bun install
```

### 开发模式

```bash
bun dev
```

### 构建项目

```bash
bun run build
```

### 预览构建结果

```bash
bun run preview
```

## 开发脚本

- `bun dev` - 启动开发服务器
- `bun run build` - 构建生产版本
- `bun run preview` - 预览构建结果
- `bun run check` - TypeScript类型检查
- `bun run format` - 代码格式化
- `bun run lint` - 代码检查
- `bun run test` - 运行测试

## 项目结构

```
frontend/
├── src/
│   ├── lib/                # 共享库文件
│   │   ├── components/     # 可复用组件
│   │   ├── stores/         # Svelte stores
│   │   ├── types/          # TypeScript类型定义
│   │   └── utils/          # 工具函数
│   ├── routes/             # 页面路由
│   │   ├── auth/          # 认证相关页面
│   │   ├── charging/      # 充电相关页面
│   │   ├── admin/         # 管理员页面
│   │   └── dashboard/     # 仪表板
│   ├── app.html           # HTML模板
│   ├── app.css            # 全局样式
│   └── app.d.ts           # 类型声明
├── static/                # 静态文件
├── tests/                 # 测试文件
└── 配置文件
```

## 页面路由

### 公开页面

- `/` - 首页
- `/auth/login` - 登录页面
- `/auth/register` - 注册页面

### 用户页面

- `/dashboard` - 用户仪表板
- `/charging/request` - 充电请求
- `/charging/status` - 充电状态
- `/charging/history` - 充电历史
- `/profile` - 个人信息

### 管理员页面

- `/admin/dashboard` - 管理员仪表板
- `/admin/piles` - 充电桩管理
- `/admin/users` - 用户管理
- `/admin/billing` - 计费管理
- `/admin/reports` - 统计报表

## API集成

前端通过RESTful API与后端服务通信，基础地址: `http://localhost:8080/api/v1`

主要API端点：

- 认证: `/auth/*`
- 充电桩: `/charging-piles/*`
- 充电服务: `/charging/*`
- 用户管理: `/users/*`

## Docker 部署

### 单独部署前端应用

前端目录包含独立的 docker-compose.yml，可以单独部署前端服务：

```bash
cd frontend

# 启动前端服务
docker-compose up -d

# 查看服务状态
docker-compose ps

# 查看日志
docker-compose logs -f frontend
```

### 构建说明

Docker 部署使用多阶段构建：

1. **构建阶段**: 使用 Bun 安装依赖并构建应用
2. **运行阶段**: 使用轻量级镜像运行构建后的应用

### 环境变量

Docker 部署使用以下配置：

```bash
# 运行环境
NODE_ENV=production
PORT=3000

# 后端API地址
PUBLIC_API_URL=http://localhost:8080
```

### 访问地址

部署后的访问地址：

- **前端应用**: <http://localhost:3000>

### 健康检查

前端服务包含健康检查：

```bash
# 检查服务健康状态
curl http://localhost:3000/
```

### 常用命令

```bash
# 重新构建前端镜像
docker-compose build frontend

# 查看构建日志
docker-compose logs -f frontend

# 停止服务
docker-compose down
```

### 开发与生产环境

```bash
# 开发环境
bun dev              # 开发服务器 (http://localhost:5173)

# 生产环境
docker-compose up -d # Docker 部署 (http://localhost:3000)
```
