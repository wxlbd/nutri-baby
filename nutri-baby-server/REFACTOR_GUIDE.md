# 数据库主键重构 - Repository 层改动指南

## 改动原则

所有 Repository 接口和实现都需要将参数中的 `string` 类型 ID 改为 `int64`,但 `openID` 保留字符串(用于微信登录)。

## 快速改动规则

### BabyRepository

**接口文件**: `internal/domain/repository/baby_repository.go`

改动前后对比:
```go
// 改动前
FindByID(ctx context.Context, babyID string) (*entity.Baby, error)
FindByCreator(ctx context.Context, creatorID string) ([]*entity.Baby, error)

// 改动后
FindByID(ctx context.Context, babyID int64) (*entity.Baby, error)
FindByCreator(ctx context.Context, creatorID int64) ([]*entity.Baby, error)
```

**实现文件**: `internal/infrastructure/persistence/baby_repository_impl.go`

改动规则:
- 参数 `babyID string` → `babyID int64`
- 参数 `creatorID string` → `creatorID int64`
- SQL 查询中: `baby_id = ?` 保持不变,只是参数类型变为 int64
- 软删除检查: `deleted_at IS NULL` → 使用 GORM soft_delete 自动处理

### BabyCollaboratorRepository

**关键变动**:
- `babyID string` → `babyID int64`
- `openid string` 保留(用于微信登录)
- 新增 `userID int64` 参数方法(备用)

### Record Repositories (FeedingRecord, SleepRecord 等)

改动:
- `babyID string` → `babyID int64`
- `recordID string` → `recordID int64` (如有)
- `createBy string` → `createdBy int64`

### Vaccine 和 Subscribe Repositories

改动:
- `babyID string` → `babyID int64`
- `templateID string` → `templateID int64`
- `openid string` → `userID int64`(需要根据上下文调整)

## 批量修改策略

### 步骤1: 修改所有 Repository 接口

使用全局查找替换 (Ctrl+Shift+H or Cmd+Shift+H):

```
查找:ctx context.Context, babyID int64
替换为: (ctx context.Context, babyID int64)

查找: (ctx context.Context, recordID string)
替换为: (ctx context.Context, recordID int64)

查找: (ctx context.Context, creatorID string)
替换为: (ctx context.Context, creatorID int64)

查找: (ctx context.Context, templateID string)
替换为: (ctx context.Context, templateID int64)
```

### 步骤2: 修改所有 Repository 实现中的 SQL 查询条件

```
查找: Where("baby_id = ?, babyID
替换为: Where("baby_id = ?, babyID (参数类型已为int64)
```

### 步骤3: 更新软删除查询

所有使用 `deleted_at IS NULL` 的地方,GORM 会自动处理,但确保没有硬删除逻辑

### 步骤4: 修改 Service 层

Service 层需要处理 DTO (string) 到 Entity (int64) 的转换:

```go
// 在 Service 层添加转换逻辑
import (
    "strconv"
    "github.com/wxlbd/nutri-baby-server/pkg/snowflake"
)

// 示例
func (s *BabyService) CreateBaby(dto *dto.CreateBabyRequest, openID string) (*dto.BabyDTO, error) {
    // 获取用户
    user, err := s.userRepo.FindByOpenID(ctx, openID)
    if err != nil {
        return nil, err
    }

    // 创建实体
    baby := &entity.Baby{
        ID: snowflake.Generate(),  // int64
        Name: dto.Name,
        UserID: user.ID,           // int64 (从User.ID获取)
    }

    // 保存
    err = s.babyRepo.Create(ctx, baby)
    if err != nil {
        return nil, err
    }

    // 转换为 DTO 返回给前端 (ID变为字符串)
    return &dto.BabyDTO{
        BabyID: strconv.FormatInt(baby.ID, 10),  // int64 → string
        Name: baby.Name,
    }, nil
}
```

## 文件列表

### Repository 接口文件 (需要修改)

1. `internal/domain/repository/baby_repository.go`
2. `internal/domain/repository/baby_collaborator_repository.go`
3. `internal/domain/repository/baby_invitation_repository.go`
4. `internal/domain/repository/record_repository.go`
5. `internal/domain/repository/vaccine_repository.go`
6. `internal/domain/repository/subscribe_repository.go`
7. `internal/domain/repository/user_repository.go` (修改 DefaultBabyID 参数)

### Repository 实现文件 (需要修改)

1. `internal/infrastructure/persistence/baby_repository_impl.go`
2. `internal/infrastructure/persistence/baby_collaborator_repository_impl.go`
3. `internal/infrastructure/persistence/baby_invitation_repository_impl.go`
4. `internal/infrastructure/persistence/feeding_record_repository_impl.go`
5. `internal/infrastructure/persistence/sleep_record_repository_impl.go`
6. `internal/infrastructure/persistence/diaper_record_repository_impl.go`
7. `internal/infrastructure/persistence/growth_record_repository_impl.go`
8. `internal/infrastructure/persistence/vaccine_plan_template_repository_impl.go`
9. `internal/infrastructure/persistence/baby_vaccine_schedule_repository_impl.go`
10. `internal/infrastructure/persistence/subscribe_repository_impl.go`
11. `internal/infrastructure/persistence/subscription_cache_repository_impl.go`

### Service 层文件 (需要添加转换逻辑)

1. `internal/application/service/auth_service.go` - User 创建
2. `internal/application/service/baby_service.go` - Baby CRUD
3. `internal/application/service/feeding_record_service.go`
4. `internal/application/service/sleep_record_service.go`
5. `internal/application/service/diaper_record_service.go`
6. `internal/application/service/growth_record_service.go`
7. `internal/application/service/vaccine_schedule_service.go`
8. `internal/application/service/subscribe_service.go`

## 关键注意事项

1. **不要改动 DTO** - DTO 保持字符串类型,让前端无感知
2. **openID 保留字符串** - 微信登录逻辑仍然使用 openID 字符串
3. **Service 层是转换点** - 所有 string ↔ int64 转换在 Service 层完成
4. **软删除处理** - GORM soft_delete 插件会自动处理,无需手动

