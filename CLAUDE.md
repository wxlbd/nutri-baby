# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## 项目概述

这是一个**全栈宝宝喂养日志系统**,包含前端小程序和后端 API 服务,名为**《宝宝喂养日志》**(BabyLog+),旨在帮助新手父母记录和追踪婴幼儿的成长数据,包括喂养、睡眠、排泄、生长、疫苗接种等信息,并支持家庭成员之间的数据共享。

### 项目结构

```
nutri-baby/
├── nutri-baby-app/          # 前端小程序 (uni-app)
│   ├── src/                 # 源代码目录
│   ├── API.md              # API 接口文档 (1211行)
│   └── ...
├── nutri-baby-server/       # 后端服务 (Golang)
│   ├── cmd/                # 应用入口
│   ├── internal/           # 内部代码
│   ├── pkg/                # 公共包
│   ├── config/             # 配置文件
│   ├── README.md           # 后端说明
│   ├── DEVELOPMENT.md      # 开发指南
│   ├── PROJECT_SUMMARY.md  # 项目总结
│   └── ...
├── CLAUDE.md               # 本文件 - AI 助手指南
└── prd.md                  # 产品需求文档
```

### 核心特性

- 🏠 **单家庭模式**: 每个用户只能属于一个家庭,通过邀请码邀请家庭成员协作
- 👶 **多宝宝支持**: 一个家庭可以管理多个宝宝的成长数据
- 👨‍👩‍👧‍👦 **家庭协作**: 支持多个家庭成员共同记录和查看宝宝数据
- 📱 **首次登录引导**: 完善的用户引导流程,引导用户创建或加入家庭
- 🔄 **数据同步**: 本地存储 + 云端同步,支持离线操作
- 💉 **疫苗管理**: 完整的疫苗计划、接种记录和智能提醒

### 应用启动流程

应用启动时会自动检查用户状态并重定向 (`src/App.vue`):

```
启动 App
  ↓
检查登录状态
  ↓
├─ 未登录 → 跳转到登录页 (/pages/user/login)
  ↓
登录成功
  ↓
检查家庭状态
  ↓
├─ 无家庭 → 跳转到家庭引导页 (/pages/family/family) ✨ 显示欢迎界面
│            用户选择: [创建家庭] 或 [加入家庭]
  ↓
有家庭 → 跳转到首页 (/pages/index/index)
```

**关键文件**:
- `src/App.vue:6-81` - 启动时用户状态检查和重定向逻辑
- `src/pages/user/login.vue:42-77` - 登录后的家庭检查逻辑
- `src/pages/family/family.vue:147-174` - 首次登录欢迎引导界面

### 技术栈

**前端 (nutri-baby-app)**:
- **开发框架**: uni-app (基于 Vue 3 + TypeScript)
- **UI 组件库**: nutui-uniapp
- **构建工具**: Vite 5.2.8
- **状态管理**: 基于 Vue 3 reactive 的简化状态管理
- **目标平台**: 微信小程序(主要平台,支持多端发布)
- **数据存储**: 本地存储 + 云端同步策略

**后端 (nutri-baby-server)**:
- **开发语言**: Go 1.25
- **Web 框架**: Gin
- **数据库**: PostgreSQL
- **ORM**: GORM
- **缓存**: Redis
- **日志**: Zap (结构化日志)
- **依赖注入**: Wire (编译时依赖注入)
- **架构**: DDD (领域驱动设计) + Clean Architecture (简洁架构)
- **认证**: JWT (JSON Web Token)

## 开发命令

### 前端开发命令 (nutri-baby-app)

**基本命令**:

```bash
# 进入项目目录
cd nutri-baby-app

# 安装依赖
npm install

# H5 开发
npm run dev:h5

# 微信小程序开发(主要目标平台)
npm run dev:mp-weixin

# 类型检查
npm run type-check

# 构建微信小程序
npm run build:mp-weixin

# 构建 H5
npm run build:h5
```

### 其他平台支持

项目支持多个小程序平台和快应用,使用 `dev:mp-*` 或 `build:mp-*` 命令进行开发和构建:
- 支付宝小程序: `npm run dev:mp-alipay`
- 百度小程序: `npm run dev:mp-baidu`
- 抖音小程序: `npm run dev:mp-toutiao`
- QQ 小程序: `npm run dev:mp-qq`
- 小红书小程序: `npm run dev:mp-xhs`

### 后端开发命令 (nutri-baby-server)

**基本命令**:

```bash
# 进入项目目录
cd nutri-baby-server

# 安装依赖
go mod download

# 安装 Wire 工具
go install github.com/google/wire/cmd/wire@latest

# 生成依赖注入代码
cd wire && wire
# 或使用 Makefile
make wire

# 运行服务 (默认端口 8080)
make run

# 运行测试
make test

# 代码格式化
make fmt

# 代码检查
make lint

# 构建二进制文件
make build

# 数据库迁移
make migrate-up    # 执行迁移
make migrate-down  # 回滚迁移
```

**配置说明**:

编辑 `nutri-baby-server/config/config.yaml` 配置数据库和 Redis:

```yaml
database:
  host: localhost
  port: 5432
  user: postgres
  password: your_password
  dbname: nutri_baby

redis:
  host: localhost
  port: 6379
  password: ""
  db: 0

jwt:
  secret: your-secret-key-change-in-production
  expire_hours: 72

wechat:
  app_id: your_wechat_app_id
  app_secret: your_wechat_app_secret
```

## 项目结构

### 前端项目结构 (nutri-baby-app)

```
nutri-baby-app/
├── src/
│   ├── pages/              # 页面目录 (已实现主要功能)
│   │   ├── index/          # 首页 - 今日概览仪表盘
│   │   ├── baby/           # 宝宝管理
│   │   │   ├── list/       # 宝宝列表
│   │   │   └── edit/       # 编辑宝宝信息
│   │   ├── record/         # 记录功能
│   │   │   ├── feeding/    # 喂养记录 (母乳/奶瓶/辅食)
│   │   │   ├── diaper/     # 换尿布记录
│   │   │   ├── sleep/      # 睡眠记录 (计时器功能)
│   │   │   └── growth/     # 成长记录 (身高/体重/头围)
│   │   ├── timeline/       # 时间轴视图
│   │   ├── statistics/     # 统计分析图表
│   │   ├── family/         # 家庭成员管理
│   │   ├── vaccine/        # 疫苗接种提醒
│   │   └── user/           # 用户中心
│   │       ├── login.vue   # 登录页
│   │       └── user.vue    # 个人中心
│   ├── store/              # 状态管理 (9个模块)
│   │   ├── user.ts         # 用户状态
│   │   ├── baby.ts         # 宝宝信息
│   │   ├── feeding.ts      # 喂养记录
│   │   ├── diaper.ts       # 换尿布记录
│   │   ├── sleep.ts        # 睡眠记录
│   │   ├── growth.ts       # 成长记录
│   │   ├── family.ts       # 家庭成员
│   │   ├── vaccine.ts      # 疫苗管理
│   │   └── index.ts        # 统一导出
│   ├── types/              # TypeScript 类型定义
│   │   └── index.ts        # 327行完整类型定义
│   ├── utils/              # 工具类库
│   │   ├── storage.ts      # 本地存储工具
│   │   ├── request.ts      # HTTP 请求封装
│   │   ├── date.ts         # 日期时间工具
│   │   ├── common.ts       # 通用工具函数
│   │   ├── export.ts       # 数据导出导入
│   │   └── index.ts        # 统一导出
│   ├── static/             # 静态资源
│   │   └── tabbar/         # 底部导航栏图标
│   ├── App.vue             # 应用主组件
│   ├── main.ts             # 应用入口
│   ├── pages.json          # 页面路由配置 (13个页面 + tabBar)
│   ├── manifest.json       # 应用配置
│   └── uni.scss            # 全局样式
├── API.md                  # RESTful API 接口文档 (1321行)
├── vite.config.ts          # Vite 配置 (nutui 自动导入)
├── tsconfig.json           # TypeScript 配置
└── package.json            # 项目依赖
```

**重要说明**:
- 项目已完成核心功能页面开发,共 13 个页面已在 `pages.json` 中注册
- 底部 tabBar 包含 4 个主要入口:首页、时间轴、统计、我的
- 不再有 `src/src/` 嵌套目录,所有功能已迁移至 `src/pages/` 下

### 后端项目结构 (nutri-baby-server)

```
nutri-baby-server/
├── cmd/                         # 应用程序入口
│   └── server/
│       └── main.go             # ✅ 主程序入口
├── internal/                    # 内部应用代码 (DDD 四层架构)
│   ├── domain/                 # 领域层 (核心业务逻辑)
│   │   ├── entity/            # 实体定义
│   │   │   ├── user.go        # ✅ 用户、家庭、成员、邀请码实体
│   │   │   ├── baby.go        # ✅ 宝宝档案实体
│   │   │   ├── record.go      # ✅ 各类记录实体(喂养/睡眠/换尿布/成长)
│   │   │   └── vaccine.go     # ✅ 疫苗计划、记录、提醒实体
│   │   └── repository/        # 仓储接口定义
│   │       ├── user_repository.go          # ✅ 用户相关仓储接口
│   │       ├── record_repository.go        # ✅ 记录仓储接口
│   │       └── vaccine_repository.go       # ✅ 疫苗仓储接口
│   ├── application/           # 应用层 (业务服务)
│   │   ├── dto/               # ⏸️ 数据传输对象 (待实现)
│   │   ├── service/           # ⏸️ 应用服务 (待实现)
│   │   └── assembler/         # ⏸️ 组装器 (待实现)
│   ├── infrastructure/        # 基础设施层 (技术实现)
│   │   ├── config/
│   │   │   └── config.go      # ✅ 配置管理 (Viper)
│   │   ├── logger/
│   │   │   └── logger.go      # ✅ 日志系统 (Zap)
│   │   └── persistence/
│   │       ├── database.go    # ✅ PostgreSQL 数据库连接
│   │       ├── redis.go       # ✅ Redis 连接
│   │       └── *_repository_impl.go  # ⏸️ 仓储实现 (待实现)
│   └── interface/             # 接口层 (HTTP API)
│       ├── http/
│       │   ├── handler/       # ⏸️ HTTP 处理器 (待实现)
│       │   └── router/        # ⏸️ 路由配置 (待实现)
│       └── middleware/        # ⏸️ 中间件 (JWT/CORS/日志等,待实现)
├── pkg/                       # 公共库 (可被外部引用)
│   ├── errors/
│   │   └── errors.go          # ✅ 统一错误定义
│   ├── response/
│   │   └── response.go        # ✅ 统一响应封装
│   └── utils/                 # ⏸️ 工具函数 (待扩展)
├── wire/                      # Wire 依赖注入
│   ├── wire.go                # ✅ Wire 配置和 Provider 定义
│   ├── wire_gen.go            # ✅ Wire 自动生成代码
│   └── app.go                 # ✅ 应用结构体定义
├── config/
│   └── config.yaml            # ✅ 配置文件
├── migrations/                # ⏸️ 数据库迁移脚本 (待实现)
├── logs/                      # 日志文件目录
├── go.mod                     # ✅ Go 模块定义
├── go.sum                     # ✅ 依赖版本锁定
├── Makefile                   # ✅ 构建脚本
├── README.md                  # ✅ 项目说明
├── DEVELOPMENT.md             # ✅ 开发指南 (详细的开发流程说明)
└── PROJECT_SUMMARY.md         # ✅ 项目总结
```

**架构说明**:
- ✅ 已完成: 项目框架、领域层实体和仓储接口、基础设施层核心组件
- ⏸️ 待实现: 仓储实现、应用服务、HTTP 处理器、中间件
- 采用 DDD 四层架构,依赖方向: Interface → Application → Domain ← Infrastructure
- 使用 Wire 进行编译时依赖注入,避免运行时反射开销


## 核心架构

### 前端架构

#### 数据模型设计

项目采用 TypeScript 定义了完整的数据模型(参见 `src/types/index.ts` - 327行),核心实体包括:

1. **用户与家庭**
   - `UserInfo`: 用户基本信息(通过微信授权获取 openid)
   - `BabyProfile`: 宝宝档案(姓名、出生日期、性别等)
   - `FamilyInfo`: 家庭信息(家庭名称、成员列表、宝宝列表)
   - `FamilyMember`: 家庭成员(支持多成员协作记录)
   - `InvitationInfo`: 邀请码信息(用于家庭成员邀请)

2. **记录类型** (使用判别联合类型设计)
   - `FeedingRecord`: 喂养记录,支持三种类型:
     - `BreastFeeding`: 母乳喂养(左/右侧、时长)
     - `BottleFeeding`: 奶瓶喂养(配方奶/母乳、奶量)
     - `FoodFeeding`: 辅食记录(名称、备注)
   - `DiaperRecord`: 排泄记录(类型、大便颜色和性状)
   - `SleepRecord`: 睡眠记录(开始/结束时间、小睡/夜间睡眠)
   - `GrowthRecord`: 成长记录(身高/体重/头围)
   - `VaccineRecord`: 疫苗接种记录

3. **疫苗管理系统** (新增功能)
   - `VaccineType`: 疫苗类型枚举(BCG、HepB、OPV、DTaP、MMR 等 16 种)
   - `VaccinePlan`: 疫苗计划项(月龄、剂次、是否必打、提醒天数)
   - `VaccineRecord`: 疫苗接种记录(医院、批号、医生、反应等详细信息)
   - `VaccineReminder`: 疫苗提醒(状态:upcoming/due/overdue/completed)
   - `VaccineReminderStatus`: 提醒状态类型

4. **同步与配置**
   - `SyncStatus`: 同步状态(idle/syncing/success/error)
   - `SyncConfig`: 同步配置(自动同步、同步间隔、仅Wi-Fi)

5. **联合类型设计**
   - 使用 TypeScript 的判别联合类型(`type` 字段作为判别器)
   - 所有记录通过 `Record` 联合类型统一处理
   - 保证类型安全的同时提供灵活性

### 状态管理设计

项目使用基于 Vue 3 `reactive` 的简化状态管理方案,位于 `nutri-baby-app/src/store/` 目录,共 9 个模块:

1. **user.ts** - 用户状态管理
   - 用户登录状态
   - 用户信息(openid、昵称、头像)
   - Token 管理

2. **baby.ts** - 宝宝信息管理
   - 当前选中的宝宝
   - 宝宝列表
   - 宝宝档案 CRUD 操作

3. **feeding.ts** - 喂养记录管理
   - 喂养记录列表
   - 添加/更新/删除喂养记录
   - 今日喂养统计

4. **diaper.ts** - 换尿布记录管理
   - 换尿布记录列表
   - 今日换尿布统计

5. **sleep.ts** - 睡眠记录管理
   - 睡眠记录列表
   - 睡眠计时器状态
   - 今日睡眠统计

6. **growth.ts** - 成长记录管理
   - 成长记录列表
   - 生长曲线数据

7. **family.ts** - 家庭成员管理
   - 家庭信息
   - 成员列表
   - 邀请码生成与加入

8. **vaccine.ts** - 疫苗管理
   - 疫苗计划列表
   - 疫苗接种记录
   - 疫苗提醒列表
   - 接种统计

9. **index.ts** - 统一导出所有 store 模块

**使用模式**:
```typescript
import { useBabyStore } from '@/store'

const babyStore = useBabyStore()
const currentBaby = babyStore.currentBaby
```

### 本地存储策略

- 使用统一前缀 `nutri_baby_` 避免命名冲突
- 定义 `StorageKeys` 枚举管理所有存储键
- 支持离线记录队列(`OFFLINE_QUEUE`),网络恢复后同步
- 封装 `storage.ts` 提供类型安全的存储 API

### 数据导出导入功能

项目提供了完整的数据导出导入功能(参见 `src/utils/export.ts`):

1. **导出功能**
   - `exportAllDataToJSON()`: 导出所有数据为 JSON 格式
   - `saveDataToFile()`: 保存数据到本地文件
   - `shareDataFile()`: 分享数据文件
   - `generateExportSummary()`: 生成导出数据摘要

2. **导入功能**
   - `importDataFromJSON()`: 从 JSON 导入数据
   - `readFileContent()`: 读取文件内容
   - 支持数据验证和格式检查

### API 接口文档

项目完整的 RESTful API 接口文档位于 `nutri-baby-app/API.md` (1321行),包含:

1. **用户认证** - 微信登录、Token 刷新
2. **家庭管理** - 创建/获取家庭、邀请/移除成员
3. **宝宝档案** - CRUD 操作、多宝宝支持
4. **喂养记录** - 母乳/奶瓶/辅食记录管理
5. **睡眠记录** - 睡眠时间追踪和统计
6. **换尿布记录** - 排泄类型和详情记录
7. **成长记录** - 身高/体重/头围追踪
8. **疫苗管理** (新增)
   - 获取疫苗计划
   - 疫苗接种记录 CRUD
   - 疫苗提醒列表
   - 疫苗接种统计
   - 标记提醒已发送
9. **数据同步** - 批量上传/拉取更新/同步状态
10. **统计分析** - 各类记录的统计数据
11. **文件上传** - 图片上传功能
12. **WebSocket 实时推送** - 实时数据同步和提醒

**API 设计特点**:
- 统一响应格式 `ApiResponse<T>`
- 标准错误码定义
- 支持分页查询
- 软删除策略
- 时间戳冲突解决
- 数据库索引优化

### HTTP 请求设计

- 基于 `uni.request` 封装统一请求方法
- 自动处理 token 认证(Bearer 模式)
- 统一错误处理和 toast 提示
- 401 状态码自动跳转登录
- 提供 RESTful 风格的快捷方法(`get`, `post`, `put`, `del`)
- 支持文件上传

### UI 组件自动导入

项目配置了 `@uni-helper/vite-plugin-uni-components` 和 `NutResolver`,NutUI 组件可按需自动导入:

```vue
<template>
  <!-- 无需手动导入,直接使用 -->
  <nut-button type="primary">按钮</nut-button>
</template>
```

### 路径别名

TypeScript 和 Vite 均配置了 `@/*` 指向 `src/*`:

```typescript
import { StorageKeys } from '@/utils/storage'
import type { BabyProfile } from '@/types'
```

### 后端架构

#### DDD 四层架构设计

后端采用 **领域驱动设计 (DDD)** + **简洁架构 (Clean Architecture)** 模式,严格遵循 **依赖倒置原则**:

```
┌─────────────────────────────────────────────┐
│     Interface Layer (接口层)                 │
│  HTTP Handlers, Middleware, Router         │
│  依赖: Application Layer                    │
└─────────────────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────────┐
│   Application Layer (应用层)                 │
│  Services, DTOs, Assemblers                │
│  依赖: Domain Layer                         │
└─────────────────────────────────────────────┘
                   ↓
┌─────────────────────────────────────────────┐
│      Domain Layer (领域层) - 核心             │
│  Entities, Value Objects, Repositories     │
│  ⚠️ 不依赖任何其他层,纯业务逻辑                │
└─────────────────────────────────────────────┘
                   ↑
┌─────────────────────────────────────────────┐
│  Infrastructure Layer (基础设施层)           │
│  Persistence, Cache, Logger, Config        │
│  依赖: Domain Layer (实现仓储接口)            │
└─────────────────────────────────────────────┘
```

**核心原则**:
1. **依赖倒置**: Infrastructure 层实现 Domain 层定义的接口,而不是相反
2. **关注点分离**: 每层只关注自己的职责,降低耦合
3. **可测试性**: Domain 层和 Application 层可独立进行单元测试
4. **灵活性**: 更换数据库或框架只需修改 Infrastructure 层

#### 领域层实体设计

位于 `nutri-baby-server/internal/domain/entity/`,核心实体包括:

1. **用户与家庭实体** (`user.go`)
   - `User`: 用户实体(OpenID、昵称、头像)
   - `Family`: 家庭实体(家庭名称、创建者)
   - `FamilyMember`: 家庭成员关系(角色: admin/member)
   - `Invitation`: 邀请码实体(邀请码、状态、有效期)

2. **宝宝档案实体** (`baby.go`)
   - `Baby`: 宝宝基本信息(姓名、性别、出生日期)
   - 支持多宝宝管理

3. **记录实体** (`record.go`)
   - `FeedingRecord`: 喂养记录(母乳/奶瓶/辅食)
   - `SleepRecord`: 睡眠记录(开始/结束时间、类型)
   - `DiaperRecord`: 换尿布记录(类型、颜色、性状)
   - `GrowthRecord`: 成长记录(身高/体重/头围)

4. **疫苗管理实体** (`vaccine.go`)
   - `VaccinePlan`: 疫苗计划(疫苗类型、月龄、剂次)
   - `VaccineRecord`: 疫苗接种记录(医院、批号、医生)
   - `VaccineReminder`: 疫苗提醒(状态、提醒时间)

所有实体均包含:
- 基础字段: ID, CreatedAt, UpdatedAt, DeletedAt (软删除)
- GORM 标签定义数据库映射
- 业务逻辑验证方法

#### 仓储模式 (Repository Pattern)

位于 `nutri-baby-server/internal/domain/repository/`,定义数据访问接口:

```go
// 示例: VaccineRecordRepository 接口
type VaccineRecordRepository interface {
    Create(ctx context.Context, record *entity.VaccineRecord) error
    GetByID(ctx context.Context, id string) (*entity.VaccineRecord, error)
    Update(ctx context.Context, record *entity.VaccineRecord) error
    Delete(ctx context.Context, id string) error
    ListByBabyID(ctx context.Context, babyID string, offset, limit int) ([]*entity.VaccineRecord, int64, error)
    GetByPlanAndBaby(ctx context.Context, planID, babyID string) (*entity.VaccineRecord, error)
}
```

**设计优势**:
- Domain 层只定义接口,Infrastructure 层负责实现
- 方便进行单元测试 (可用 Mock 替代真实数据库)
- 易于切换数据存储方案 (PostgreSQL → MongoDB)

#### Wire 依赖注入

使用 Google Wire 进行**编译时依赖注入**,位于 `nutri-baby-server/wire/`:

```go
// wire.go - Provider 定义
var infrastructureSet = wire.NewSet(
    config.NewConfig,
    logger.NewLogger,
    persistence.NewDatabase,
    persistence.NewRedis,
)

var repositorySet = wire.NewSet(
    // 仓储实现 Provider
    persistence.NewUserRepository,
    persistence.NewBabyRepository,
    // ... 更多仓储
)

var serviceSet = wire.NewSet(
    // 应用服务 Provider
    service.NewAuthService,
    service.NewBabyService,
    // ... 更多服务
)

// InitializeApp - Wire 自动生成依赖注入代码
func InitializeApp() (*App, error) {
    wire.Build(
        infrastructureSet,
        repositorySet,
        serviceSet,
        // ... handlers, routers
        NewApp,
    )
    return nil, nil
}
```

**优势**:
- 编译时生成代码,无运行时反射开销
- 类型安全,编译期发现依赖问题
- 代码可读性强,依赖关系清晰

#### 统一响应与错误处理

**响应封装** (`pkg/response/response.go`):
```go
type Response struct {
    Code      int         `json:"code"`
    Message   string      `json:"message"`
    Data      interface{} `json:"data,omitempty"`
    Timestamp int64       `json:"timestamp"`
}

// 成功响应
func Success(c *gin.Context, data interface{})

// 错误响应
func Error(c *gin.Context, err error)
```

**错误定义** (`pkg/errors/errors.go`):
```go
var (
    ErrInvalidParam   = errors.New(1001, "参数错误")
    ErrUnauthorized   = errors.New(1002, "未授权")
    ErrNotFound       = errors.New(1003, "资源不存在")
    ErrConflict       = errors.New(1004, "数据冲突")
    ErrForbidden      = errors.New(1005, "权限不足")
    ErrInternalServer = errors.New(2001, "服务器内部错误")
)
```

#### 数据库与缓存

**PostgreSQL 配置** (`internal/infrastructure/persistence/database.go`):
- 使用 GORM 作为 ORM
- 支持连接池配置 (MaxOpenConns, MaxIdleConns, ConnMaxLifetime)
- 自动迁移表结构
- 软删除支持

**Redis 配置** (`internal/infrastructure/persistence/redis.go`):
- 用于缓存热点数据 (用户 Session、家庭信息)
- Token 黑名单管理
- 实时数据推送

#### 日志系统

使用 **Uber Zap** 提供结构化日志 (`internal/infrastructure/logger/logger.go`):

```go
logger.Info("用户登录",
    zap.String("openid", openid),
    zap.String("ip", clientIP),
)

logger.Error("数据库查询失败",
    zap.Error(err),
    zap.String("query", sql),
)
```

**特性**:
- 高性能结构化日志
- 支持日志分级 (Debug, Info, Warn, Error)
- 日志轮转 (lumberjack)
- 生产环境自动输出 JSON 格式

## 核心功能与实现状态

根据 PRD 文档(`prd.md`)和当前开发进度:

### 1. 用户与家庭管理 (FR-101 ~ FR-103)
- ✅ 微信一键授权登录 (`pages/user/login.vue`)
- ✅ 宝宝档案管理 (`pages/baby/list` 和 `pages/baby/edit`)
- ✅ 家庭成员邀请与协作 (`pages/family/family.vue`)
- ✅ 支持多孩家庭
- ✅ 家庭成员角色管理(admin/member)

### 2. 核心记录功能 (FR-201 ~ FR-204)
- ✅ 喂养记录 (`pages/record/feeding/feeding.vue`)
  - 母乳喂养(左/右侧计时)
  - 奶瓶喂养(配方奶/母乳、奶量记录)
  - 辅食记录
- ✅ 换尿布记录 (`pages/record/diaper/diaper.vue`)
  - 快捷按钮(小便/大便/两者)
  - 大便颜色和性状详细信息
- ✅ 睡眠记录 (`pages/record/sleep/sleep.vue`)
  - 计时器功能
  - 小睡/夜间睡眠区分
- ✅ 成长记录 (`pages/record/growth/growth.vue`)
  - 身高、体重、头围记录

### 3. 数据呈现 (FR-301 ~ FR-303)
- ✅ "今日"仪表盘 (`pages/index/index.vue` - 667行主页面)
  - 核心数据摘要(奶量、睡眠、换尿布统计)
  - 距离上次喂奶时间提示
  - 快捷记录按钮
- ✅ 时间轴视图 (`pages/timeline/timeline.vue`)
  - 时间倒序展示所有事件
  - 不同类型事件图标区分
- ✅ 统计图表 (`pages/statistics/statistics.vue`)
  - 趋势分析(按周/月)
  - ⏳ WHO 生长曲线待实现

### 4. 疫苗管理 (新增功能)
- ✅ 疫苗接种提醒 (`pages/vaccine/vaccine.vue`)
- ✅ 疫苗计划管理
- ✅ 疫苗接种记录
- ✅ 接种统计和提醒状态

### 5. 辅助功能 (FR-401 ~ FR-403)
- ✅ 数据导出导入(JSON 格式)
- ⏳ 智能提醒(微信订阅消息) - 待后端集成
- ⏳ 育儿知识库(可选) - 未实现

## 开发注意事项

### 前端开发 (uni-app)

#### uni-app 特性

1. **单位系统**: 使用 `rpx` 作为响应式单位(750rpx = 屏幕宽度)
2. **生命周期**: 使用 `@dcloudio/uni-app` 提供的组合式 API
   ```typescript
   import { onLaunch, onShow } from '@dcloudio/uni-app'
   ```
3. **API 调用**: 统一使用 `uni.*` API,兼容多端
4. **条件编译**: 使用 `#ifdef` 和 `#ifndef` 实现平台差异化

### 页面注册流程

新增页面必须在 `src/pages.json` 中注册:

```json
{
  "pages": [
    {
      "path": "pages/新页面/index",
      "style": {
        "navigationBarTitleText": "页面标题"
      }
    }
  ]
}
```

### 组件规范

- 所有组件使用 `<script setup lang="ts">` 语法
- Vue 3 Composition API 为主
- 遵循 nutui-uniapp 的设计规范

### 样式处理

- 全局样式在 `uni.scss` 中定义
- 组件样式使用 `<style scoped>` 或 `<style lang="scss">`
- nutui-uniapp 变量已在 Vite 配置中自动导入

### 微信小程序特定配置

- `manifest.json` 中 `mp-weixin.appid` 需填写微信小程序 AppID
- 订阅消息需在微信公众平台配置模板 ID
- 云开发需在微信开发者工具中开通并初始化

### 后端开发 (Golang)

#### 代码规范

1. **命名规范**
   - 包名: 小写单词,不使用下划线或驼峰 (`persistence`, `repository`)
   - 接口名: 大写字母开头,描述能力 (`UserRepository`, `Logger`)
   - 实体名: 大写字母开头,清晰明了 (`User`, `Baby`, `VaccineRecord`)
   - 私有方法: 小写字母开头 (`createToken`, `validateUser`)

2. **错误处理**
   - 使用 `pkg/errors` 中定义的业务错误
   - 数据库错误需转换为业务错误
   - 记录详细的错误日志,包含上下文信息

3. **Context 传递**
   - 所有 Repository 方法必须接收 `context.Context`
   - HTTP Handler 使用 `c.Request.Context()`
   - 用于超时控制和请求取消

#### Wire 使用最佳实践

1. **Provider 函数规范**
   ```go
   // ✅ 正确: 清晰的依赖注入
   func NewUserService(
       userRepo repository.UserRepository,
       logger *zap.Logger,
   ) *UserService {
       return &UserService{
           userRepo: userRepo,
           logger:   logger,
       }
   }

   // ❌ 错误: 隐藏依赖
   func NewUserService() *UserService {
       logger := zap.NewProduction() // 不要在构造函数中创建依赖
       return &UserService{logger: logger}
   }
   ```

2. **Wire 代码生成**
   - 修改 `wire/wire.go` 后必须运行 `cd wire && wire` 重新生成
   - 不要手动修改 `wire_gen.go`
   - 编译错误通常是缺少 Provider 或循环依赖

3. **依赖注入顺序**
   ```
   基础设施 (Config, Logger, DB, Redis)
       ↓
   仓储实现 (Repository Implementations)
       ↓
   应用服务 (Services)
       ↓
   HTTP 处理器 (Handlers)
       ↓
   路由和应用 (Router, App)
   ```

#### GORM 使用注意事项

1. **预加载关联**
   ```go
   // ✅ 使用 Preload 避免 N+1 查询
   db.Preload("Babies").Preload("Members").First(&family, id)

   // ❌ 不要逐个查询关联
   db.First(&family, id)
   for _, member := range family.Members { // N+1 问题
       db.First(&member)
   }
   ```

2. **软删除**
   - 所有实体已包含 `gorm.DeletedAt`,使用 `Delete()` 自动软删除
   - 硬删除使用 `Unscoped().Delete()`
   - 查询时 GORM 自动过滤已软删除记录

3. **事务处理**
   ```go
   err := r.db.Transaction(func(tx *gorm.DB) error {
       if err := tx.Create(&user).Error; err != nil {
           return err // 自动回滚
       }
       if err := tx.Create(&family).Error; err != nil {
           return err
       }
       return nil // 提交事务
   })
   ```

#### API 开发流程

**实现一个完整的 API 接口** (以疫苗管理为例):

1. **创建 DTO** (`internal/application/dto/vaccine_dto.go`)
   ```go
   type CreateVaccineRecordRequest struct {
       PlanID      string `json:"planId" binding:"required"`
       BabyID      string `json:"babyId" binding:"required"`
       VaccineDate int64  `json:"vaccineDate" binding:"required"`
       Hospital    string `json:"hospital"`
   }
   ```

2. **实现仓储** (`internal/infrastructure/persistence/vaccine_repository_impl.go`)
   ```go
   func (r *vaccineRecordRepositoryImpl) Create(ctx context.Context, record *entity.VaccineRecord) error {
       return r.db.WithContext(ctx).Create(record).Error
   }
   ```

3. **实现服务** (`internal/application/service/vaccine_service.go`)
   ```go
   func (s *VaccineService) CreateRecord(ctx context.Context, req *dto.CreateVaccineRecordRequest) (*dto.VaccineRecordResponse, error) {
       // 业务逻辑验证
       // 调用仓储
       // 返回 DTO
   }
   ```

4. **实现 Handler** (`internal/interface/http/handler/vaccine_handler.go`)
   ```go
   func (h *VaccineHandler) CreateRecord(c *gin.Context) {
       var req dto.CreateVaccineRecordRequest
       if err := c.ShouldBindJSON(&req); err != nil {
           response.Error(c, errs.ErrInvalidParam)
           return
       }

       result, err := h.vaccineService.CreateRecord(c.Request.Context(), &req)
       if err != nil {
           response.Error(c, err)
           return
       }

       response.Success(c, result)
   }
   ```

5. **注册路由** (`internal/interface/router/router.go`)
   ```go
   vaccines := v1.Group("/babies/:babyId/vaccine-records")
   vaccines.Use(middleware.Auth())
   {
       vaccines.POST("", vaccineHandler.CreateRecord)
       vaccines.GET("", vaccineHandler.ListRecords)
   }
   ```

6. **更新 Wire** (`wire/wire.go`)
   ```go
   var serviceSet = wire.NewSet(
       service.NewVaccineService,
   )

   var handlerSet = wire.NewSet(
       handler.NewVaccineHandler,
   )
   ```

7. **运行 Wire 生成代码**
   ```bash
   cd wire && wire
   ```

#### 中间件开发

**JWT 认证中间件示例**:
```go
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        if token == "" {
            response.Error(c, errs.ErrUnauthorized)
            c.Abort()
            return
        }

        // 验证 token,解析用户信息
        claims, err := parseToken(token)
        if err != nil {
            response.Error(c, errs.ErrUnauthorized)
            c.Abort()
            return
        }

        // 设置用户信息到 context
        c.Set("userID", claims.UserID)
        c.Set("openid", claims.OpenID)
        c.Next()
    }
}
```

## 性能与用户体验要求

根据 PRD 的非功能性需求:

- 首次加载时间 < 3 秒
- 核心操作 ≤ 3 次点击完成
- 支持单手操作,按钮区域足够大
- 支持离线记录,网络恢复自动同步
- 界面简洁温馨,色彩柔和

## 数据安全

- 所有 HTTP 通信使用 HTTPS
- 敏感数据(token)加密存储
- 云数据库权限严格控制,仅家庭成员可访问
- 不得在代码中硬编码敏感信息(API keys、tokens)

## 当前开发状态

### 前端 (nutri-baby-app) - 已完成 ✅
- 项目基础架构搭建
- 完整的 TypeScript 类型定义系统 (327行)
- 9 个状态管理模块 (user, baby, feeding, diaper, sleep, growth, family, vaccine)
- 工具类库完整封装 (storage, request, date, common, export)
- nutui-uniapp UI 框架集成和自动导入配置
- 13 个功能页面开发完成
- 数据导出导入功能
- 底部 tabBar 导航配置

### 后端 (nutri-baby-server) - 部分完成 ⏳

**已完成 ✅**:
- DDD 四层架构搭建
- 领域层实体定义 (User, Family, Baby, Records, Vaccine)
- 仓储接口定义 (Repository Interfaces)
- 基础设施层核心组件:
  - 配置管理 (Viper)
  - 日志系统 (Zap + Lumberjack)
  - 数据库连接 (PostgreSQL + GORM)
  - Redis 连接
- 统一错误定义和响应封装
- Wire 依赖注入框架配置
- Makefile 构建脚本
- 完整的开发文档 (README, DEVELOPMENT, PROJECT_SUMMARY)

**进行中 ⏳**:
- 仓储实现 (Repository Implementations)
- 应用服务层 (Application Services)
- HTTP 处理器 (Handlers)
- 路由配置 (Router)
- 中间件 (JWT 认证、CORS、日志)

**待实现 ⏸️**:
- 数据库迁移脚本
- 微信登录集成
- JWT Token 生成和验证
- WebSocket 实时推送
- 单元测试和集成测试
- 疫苗提醒定时任务

### API 文档 - 已完成 ✅
- RESTful API 接口文档编写 (1211行,包含疫苗管理)
- 50+ API 接口设计
- 数据库表结构设计
- WebSocket 推送协议设计

## 项目关键文件说明

### 前端关键文件
- `nutri-baby-app/src/pages/index/index.vue` (667行) - 核心主页面,包含今日仪表盘
- `nutri-baby-app/src/types/index.ts` (327行) - 完整的 TypeScript 类型定义
- `nutri-baby-app/src/store/vaccine.ts` (10758字节) - 疫苗管理状态
- `nutri-baby-app/src/pages.json` - 13个页面 + tabBar 配置
- `nutri-baby-app/src/utils/request.ts` - HTTP 请求封装

### 后端关键文件
- `nutri-baby-server/cmd/server/main.go` - 应用程序入口
- `nutri-baby-server/internal/domain/entity/*.go` - 领域实体定义
- `nutri-baby-server/internal/domain/repository/*.go` - 仓储接口定义
- `nutri-baby-server/internal/infrastructure/config/config.go` - 配置管理
- `nutri-baby-server/internal/infrastructure/logger/logger.go` - 日志系统
- `nutri-baby-server/internal/infrastructure/persistence/database.go` - 数据库连接
- `nutri-baby-server/pkg/errors/errors.go` - 统一错误定义
- `nutri-baby-server/pkg/response/response.go` - 统一响应封装
- `nutri-baby-server/wire/wire.go` - Wire 依赖注入配置
- `nutri-baby-server/config/config.yaml` - 配置文件

### API 文档
- `nutri-baby-app/API.md` (1211行) - 完整的 RESTful API 接口文档

### 项目文档
- `nutri-baby-server/README.md` - 后端项目说明
- `nutri-baby-server/DEVELOPMENT.md` - 详细的开发指南
- `nutri-baby-server/PROJECT_SUMMARY.md` - 项目总结
- `prd.md` - 产品需求文档

## 重要提示

### 前端配置
1. **API 基础 URL 配置**: 在 `nutri-baby-app/src/utils/request.ts` 中配置 `BASE_URL`,默认为 `import.meta.env.VITE_API_BASE_URL`
2. **微信小程序 AppID**: 在 `nutri-baby-app/src/manifest.json` 的 `mp-weixin.appid` 字段填写
3. **疫苗管理是新增核心功能**: 包含计划、记录、提醒、统计四大模块
4. **状态管理非 Vuex/Pinia**: 使用基于 Vue 3 reactive 的简化方案
5. **数据持久化**: 本地使用 uni.setStorageSync,云端通过 RESTful API 同步

### 后端配置
1. **数据库配置**: 编辑 `nutri-baby-server/config/config.yaml` 配置 PostgreSQL 和 Redis 连接信息
2. **微信配置**: 在 config.yaml 中填写 `wechat.app_id` 和 `wechat.app_secret`
3. **JWT Secret**: 生产环境必须修改 `jwt.secret` 为强随机密钥
4. **Wire 依赖注入**: 修改 wire.go 后必须运行 `cd wire && wire` 重新生成代码
5. **数据库迁移**: 首次运行需要执行 GORM 自动迁移或手动创建表结构

### 前后端联调
1. **API Base URL**: 前端 `.env` 文件中配置后端服务地址
2. **CORS 配置**: 后端需要配置 CORS 中间件允许前端域名
3. **Token 传递**: 前端使用 Bearer Token 方式,后端 JWT 中间件验证
4. **时间戳格式**: 统一使用 Unix 时间戳 (秒级)
5. **错误码对齐**: 前后端错误码定义保持一致

## 参考资料

### 前端开发
- [uni-app 官方文档](https://uniapp.dcloud.net.cn/)
- [nutui-uniapp 组件库](https://nutui.jd.com/uniapp)
- [微信小程序开发文档](https://developers.weixin.qq.com/miniprogram/dev/framework/)
- [微信小程序云开发文档](https://developers.weixin.qq.com/miniprogram/dev/wxcloud/basis/getting-started.html)
- [Vue 3 官方文档](https://cn.vuejs.org/)
- [TypeScript 官方文档](https://www.typescriptlang.org/zh/)

### 后端开发
- [Go 官方文档](https://go.dev/doc/)
- [Gin Web Framework](https://gin-gonic.com/zh-cn/docs/)
- [GORM ORM 框架](https://gorm.io/zh_CN/docs/)
- [Wire 依赖注入](https://github.com/google/wire/blob/main/docs/guide.md)
- [Uber Zap 日志库](https://github.com/uber-go/zap)
- [Redis 官方文档](https://redis.io/docs/)
- [PostgreSQL 官方文档](https://www.postgresql.org/docs/)

### 架构设计
- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Repository Pattern](https://martinfowler.com/eaaCatalog/repository.html)

### 工具与资源
- [Viper 配置管理](https://github.com/spf13/viper)
- [JWT 官方文档](https://jwt.io/)
- [Lumberjack 日志轮转](https://github.com/natefinch/lumberjack)