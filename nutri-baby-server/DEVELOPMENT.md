# Nutri Baby Server - 开发指南

## 项目架构

本项目采用 **DDD(领域驱动设计)** + **Clean Architecture(简洁架构)** 模式,遵循 **依赖倒置原则**。

### 架构分层

```
┌─────────────────────────────────────────────┐
│        Interface Layer (接口层)              │
│    HTTP Handlers, Middleware, Router       │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│      Application Layer (应用层)              │
│    Services, DTOs, Assemblers              │
└─────────────────────────────────────────────┘
                    ↓
┌─────────────────────────────────────────────┐
│         Domain Layer (领域层)                │
│    Entities, Value Objects, Repositories   │
└─────────────────────────────────────────────┘
                    ↑
┌─────────────────────────────────────────────┐
│    Infrastructure Layer (基础设施层)          │
│    Persistence, Cache, Logger, Config      │
└─────────────────────────────────────────────┘
```

### 依赖方向

- Interface → Application → Domain
- Infrastructure → Domain (实现Domain定义的接口)
- 依赖注入通过Wire自动生成

## 已完成的核心文件

### 1. 项目配置

- ✅ `go.mod` - Go模块定义
- ✅ `config/config.yaml` - 应用配置
- ✅ `Makefile` - 构建命令
- ✅ `README.md` - 项目说明

### 2. 领域层 (Domain Layer)

**实体 (Entities)**:
- ✅ `internal/domain/entity/user.go` - 用户、家庭、成员、邀请码
- ✅ `internal/domain/entity/baby.go` - 宝宝档案
- ✅ `internal/domain/entity/record.go` - 喂养、睡眠、换尿布、成长记录
- ✅ `internal/domain/entity/vaccine.go` - 疫苗计划、记录、提醒

**仓储接口 (Repository Interfaces)**:
- ✅ `internal/domain/repository/user_repository.go` - 用户相关仓储接口
- ✅ `internal/domain/repository/record_repository.go` - 记录仓储接口
- ✅ `internal/domain/repository/vaccine_repository.go` - 疫苗仓储接口

### 3. 基础设施层 (Infrastructure Layer)

- ✅ `internal/infrastructure/config/config.go` - 配置管理
- ✅ `internal/infrastructure/logger/logger.go` - Zap日志
- ✅ `internal/infrastructure/persistence/database.go` - PostgreSQL连接
- ✅ `internal/infrastructure/persistence/redis.go` - Redis连接

### 4. 公共包 (Pkg)

- ✅ `pkg/errors/errors.go` - 错误定义
- ✅ `pkg/response/response.go` - 响应封装

### 5. Wire依赖注入

- ✅ `wire/wire.go` - Wire配置
- ✅ `wire/app.go` - 应用结构
- ✅ `cmd/server/main.go` - 主入口

## 待实现的功能模块

### 1. 仓储实现 (Infrastructure/Persistence)

需要为每个Repository接口创建GORM实现:

```go
// 示例: internal/infrastructure/persistence/vaccine_repository_impl.go
type vaccineRecordRepositoryImpl struct {
    db *gorm.DB
}

func NewVaccineRecordRepository(db *gorm.DB) repository.VaccineRecordRepository {
    return &vaccineRecordRepositoryImpl{db: db}
}

func (r *vaccineRecordRepositoryImpl) Create(ctx context.Context, record *entity.VaccineRecord) error {
    return r.db.WithContext(ctx).Create(record).Error
}

// ... 实现其他方法
```

需要实现的仓储:
- [ ] `user_repository_impl.go`
- [ ] `family_repository_impl.go`
- [ ] `family_member_repository_impl.go`
- [ ] `invitation_repository_impl.go`
- [ ] `baby_repository_impl.go`
- [ ] `feeding_record_repository_impl.go`
- [ ] `sleep_record_repository_impl.go`
- [ ] `diaper_record_repository_impl.go`
- [ ] `growth_record_repository_impl.go`
- [ ] `vaccine_plan_repository_impl.go`
- [ ] `vaccine_record_repository_impl.go`
- [ ] `vaccine_reminder_repository_impl.go`

### 2. 应用服务 (Application/Service)

创建业务逻辑层:

```go
// 示例: internal/application/service/vaccine_service.go
type VaccineService struct {
    vaccinePlanRepo     repository.VaccinePlanRepository
    vaccineRecordRepo   repository.VaccineRecordRepository
    vaccineReminderRepo repository.VaccineReminderRepository
    babyRepo            repository.BabyRepository
}

func NewVaccineService(
    vaccinePlanRepo repository.VaccinePlanRepository,
    vaccineRecordRepo repository.VaccineRecordRepository,
    vaccineReminderRepo repository.VaccineReminderRepository,
    babyRepo repository.BabyRepository,
) *VaccineService {
    return &VaccineService{...}
}

func (s *VaccineService) GetVaccinePlans(ctx context.Context, babyID string) ([]*dto.VaccinePlanDTO, error) {
    // 业务逻辑
}
```

需要实现的服务:
- [ ] `auth_service.go` - 认证服务
- [ ] `family_service.go` - 家庭服务
- [ ] `baby_service.go` - 宝宝服务
- [ ] `record_service.go` - 记录服务
- [ ] `vaccine_service.go` - 疫苗服务
- [ ] `sync_service.go` - 同步服务
- [ ] `statistics_service.go` - 统计服务

### 3. DTO (Data Transfer Objects)

```go
// internal/application/dto/vaccine_dto.go
type VaccinePlanDTO struct {
    PlanID        string `json:"planId"`
    VaccineName   string `json:"vaccineName"`
    AgeInMonths   int    `json:"ageInMonths"`
    DoseNumber    int    `json:"doseNumber"`
    ScheduledDate int64  `json:"scheduledDate"`
    Status        string `json:"status"`
}
```

### 4. HTTP处理器 (Interface/Handler)

```go
// internal/interface/http/handler/vaccine_handler.go
type VaccineHandler struct {
    vaccineService *service.VaccineService
}

func NewVaccineHandler(vaccineService *service.VaccineService) *VaccineHandler {
    return &VaccineHandler{vaccineService: vaccineService}
}

func (h *VaccineHandler) GetVaccinePlans(c *gin.Context) {
    babyID := c.Param("babyId")

    plans, err := h.vaccineService.GetVaccinePlans(c.Request.Context(), babyID)
    if err != nil {
        response.Error(c, err)
        return
    }

    response.Success(c, plans)
}
```

### 5. 路由 (Interface/Router)

```go
// internal/interface/router/router.go
func NewRouter(
    authHandler *handler.AuthHandler,
    vaccineHandler *handler.VaccineHandler,
    // ... 其他handlers
) *gin.Engine {
    r := gin.Default()

    // 中间件
    r.Use(middleware.CORS())
    r.Use(middleware.Logger())
    r.Use(middleware.Recovery())

    v1 := r.Group("/v1")
    {
        // 认证
        auth := v1.Group("/auth")
        {
            auth.POST("/wechat/login", authHandler.WechatLogin)
            auth.POST("/refresh", authHandler.RefreshToken)
        }

        // 疫苗管理(需要认证)
        vaccines := v1.Group("/babies/:babyId/vaccine-plans")
        vaccines.Use(middleware.Auth())
        {
            vaccines.GET("", vaccineHandler.GetVaccinePlans)
        }

        // ... 其他路由
    }

    return r
}
```

### 6. 中间件 (Interface/Middleware)

```go
// internal/interface/middleware/auth.go
func Auth() gin.HandlerFunc {
    return func(c *gin.Context) {
        token := c.GetHeader("Authorization")
        // 验证JWT token
        // 设置用户信息到context
        c.Next()
    }
}

// internal/interface/middleware/cors.go
func CORS() gin.HandlerFunc {
    return func(c *gin.Context) {
        c.Writer.Header().Set("Access-Control-Allow-Origin", "*")
        // ... 其他CORS配置
        c.Next()
    }
}
```

## 开发流程

### 1. 实现一个完整的功能模块

以疫苗管理为例:

```bash
# 1. 创建仓储实现
internal/infrastructure/persistence/vaccine_repository_impl.go

# 2. 创建DTO
internal/application/dto/vaccine_dto.go

# 3. 创建应用服务
internal/application/service/vaccine_service.go

# 4. 创建HTTP处理器
internal/interface/http/handler/vaccine_handler.go

# 5. 更新路由
internal/interface/router/router.go

# 6. 运行Wire生成依赖注入代码
make wire

# 7. 运行服务
make run
```

### 2. 测试API

```bash
# 获取疫苗计划
curl -X GET http://localhost:8080/v1/babies/{babyId}/vaccine-plans \
  -H "Authorization: Bearer {token}"

# 创建疫苗记录
curl -X POST http://localhost:8080/v1/babies/{babyId}/vaccine-records \
  -H "Authorization: Bearer {token}" \
  -H "Content-Type: application/json" \
  -d '{
    "planId": "xxx",
    "vaccineName": "乙肝疫苗",
    "vaccineDate": 1234567890,
    "hospital": "市妇幼保健院"
  }'
```

## Wire使用说明

### 1. 定义Provider

```go
// Provider就是返回依赖的函数
func NewVaccineService(...) *VaccineService {
    return &VaccineService{...}
}
```

### 2. 在wire.go中注册

```go
wire.Build(
    // 仓储
    persistence.NewVaccineRecordRepository,

    // 服务
    service.NewVaccineService,

    // 处理器
    handler.NewVaccineHandler,
)
```

### 3. 生成代码

```bash
cd wire && wire
```

生成的`wire_gen.go`会自动处理所有依赖注入。

## 数据库迁移

初始化数据库后,需要插入默认疫苗计划:

```sql
-- 插入中国国家免疫规划疫苗
INSERT INTO vaccine_plans (plan_id, vaccine_type, vaccine_name, age_in_months, dose_number, is_required, reminder_days)
VALUES
  ('plan_hepb_1', 'HepB', '乙肝疫苗', 0, 1, true, 3),
  ('plan_hepb_2', 'HepB', '乙肝疫苗', 1, 2, true, 7),
  -- ... 更多疫苗计划
```

## 最佳实践

1. **错误处理**: 使用`pkg/errors`中定义的错误类型
2. **日志记录**: 使用`logger.Info/Error`记录关键操作
3. **事务处理**: 在Service层使用GORM事务
4. **参数验证**: 使用gin的binding进行参数验证
5. **上下文传递**: 所有Repository方法都应接收context
6. **单元测试**: 为Service层编写单元测试

## 下一步

1. 实现疫苗管理模块的完整功能
2. 实现用户认证和JWT中间件
3. 实现其他业务模块
4. 添加单元测试和集成测试
5. 完善API文档
6. 实现WebSocket实时推送
7. 添加数据同步功能

## 参考资料

- [Clean Architecture](https://blog.cleancoder.com/uncle-bob/2012/08/13/the-clean-architecture.html)
- [Domain-Driven Design](https://martinfowler.com/bliki/DomainDrivenDesign.html)
- [Wire User Guide](https://github.com/google/wire/blob/main/docs/guide.md)
- [GORM Documentation](https://gorm.io/docs/)
- [Gin Documentation](https://gin-gonic.com/docs/)
