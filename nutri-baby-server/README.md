# Nutri Baby Server

宝宝喂养日志服务端 API

## 技术栈

- **框架**: Gin
- **数据库**: PostgreSQL
- **ORM**: GORM
- **缓存**: Redis
- **日志**: Zap
- **依赖注入**: Wire
- **架构**: DDD + Clean Architecture

## 项目结构

```
nutri-baby-server/
├── cmd/                    # 应用程序入口
│   └── server/
│       └── main.go
├── internal/               # 内部应用代码
│   ├── domain/            # 领域层
│   │   ├── entity/        # 实体
│   │   ├── valueobject/   # 值对象
│   │   └── repository/    # 仓储接口
│   ├── application/       # 应用层
│   │   ├── dto/           # 数据传输对象
│   │   ├── service/       # 应用服务
│   │   └── assembler/     # 组装器
│   ├── infrastructure/    # 基础设施层
│   │   ├── persistence/   # 持久化
│   │   ├── cache/         # 缓存
│   │   ├── logger/        # 日志
│   │   └── config/        # 配置
│   └── interface/         # 接口层
│       ├── http/          # HTTP处理器
│       ├── middleware/    # 中间件
│       └── router/        # 路由
├── pkg/                   # 公共库
│   ├── errors/            # 错误定义
│   ├── response/          # 响应封装
│   └── utils/             # 工具函数
├── wire/                  # Wire依赖注入
│   └── wire.go
├── config/                # 配置文件
│   └── config.yaml
├── migrations/            # 数据库迁移
├── go.mod
├── go.sum
├── Makefile
└── README.md
```

## 快速开始

### 1. 安装依赖

```bash
go mod download
```

### 2. 安装 Wire

```bash
go install github.com/google/wire/cmd/wire@latest
```

### 3. 生成依赖注入代码

```bash
cd wire && wire
```

### 4. 配置数据库

编辑 `config/config.yaml`:

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  dbname: nutri_baby
```

### 5. 运行迁移

```bash
make migrate-up
```

### 6. 启动服务

```bash
make run
```

## 开发命令

```bash
# 生成Wire代码
make wire

# 运行服务
make run

# 运行测试
make test

# 数据库迁移
make migrate-up
make migrate-down

# 代码格式化
make fmt

# 代码检查
make lint
```

## API 文档

详见 [API.md](../nutri-baby-app/API.md)
