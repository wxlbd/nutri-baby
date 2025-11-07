# 主键重构进度报告

## ✅ 已完成 (约 30% 进度)

### 1. 实体层 (100% 完成)
- ✅ `pkg/snowflake/snowflake.go` - 雪花算法工具包
- ✅ `internal/domain/entity/user.go` - User 实体改造
- ✅ `internal/domain/entity/baby.go` - Baby + BabyCollaborator 实体改造
- ✅ `internal/domain/entity/baby_invitation.go` - BabyInvitation 实体改造
- ✅ `internal/domain/entity/record.go` - 4个记录实体改造
- ✅ `internal/domain/entity/vaccine.go` - 疫苗实体改造
- ✅ `internal/domain/entity/subscribe.go` - 订阅实体改造

### 2. Repository 接口 (部分完成)
- ✅ `internal/domain/repository/user_repository.go` - User + Baby 接口
- ✅ `internal/domain/repository/baby_collaborator_repository.go` - BabyCollaborator 接口
- ⏳ `internal/domain/repository/baby_invitation_repository.go` - 待改
- ⏳ `internal/domain/repository/record_repository.go` - 待改 (4个接口)
- ⏳ `internal/domain/repository/vaccine_repository.go` - 待改
- ⏳ `internal/domain/repository/subscribe_repository.go` - 待改

## 📋 剩余工作 (约 70% 工作量)

### 3. Repository 实现层 (13个文件)
需要修改 `internal/infrastructure/persistence/` 中的:
- baby_repository_impl.go
- baby_collaborator_repository_impl.go
- baby_invitation_repository_impl.go
- feeding_record_repository_impl.go
- sleep_record_repository_impl.go
- diaper_record_repository_impl.go
- growth_record_repository_impl.go
- vaccine_plan_template_repository_impl.go
- baby_vaccine_schedule_repository_impl.go
- subscribe_repository_impl.go
- subscription_cache_repository_impl.go
- user_repository_impl.go

**改动规则**:
```go
// SQL查询中,将参数从string改为int64
// 改动前: Where("baby_id = ?", babyID) // babyID string
// 改动后: Where("baby_id = ?", babyID) // babyID int64

// GORM 软删除自动处理,无需改动查询条件
```

### 4. Service 层 (10个文件)
需要在各 Service 的方法中添加 ID 转换逻辑:

```go
import (
    "strconv"
    "github.com/wxlbd/nutri-baby-server/pkg/snowflake"
)

// 示例:从 DTO 获取ID字符串后转为int64
func (s *Service) Method(babyIDStr string) {
    babyID, _ := strconv.ParseInt(babyIDStr, 10, 64)
    // 使用 int64 型 babyID...
}

// 示例:返回前将int64转回字符串
return &dto.BabyDTO{
    BabyID: strconv.FormatInt(baby.ID, 10),
}
```

待修改文件:
- auth_service.go
- baby_service.go
- feeding_record_service.go
- sleep_record_service.go
- diaper_record_service.go
- growth_record_service.go
- vaccine_schedule_service.go
- subscribe_service.go
- timeline_service.go
- sync_service.go

### 5. 数据库迁移脚本
需要生成 `migrations/008_refactor_to_snowflake_id.sql`

包含:
- 为所有表添加新 `id BIGINT PRIMARY KEY` 字段
- 迁移数据(如有需要)
- 更新外键约束
- 时间字段改为 BIGINT 毫秒时间戳

## 🚀 推荐实施方案

### 选项 A: 继续让我完成 (耗时较长)
我继续逐个修改剩余文件,预计还需要 30-40 分钟

### 选项 B: 您使用IDE批量替换 (快速)
1. 使用 IDE 的全局查找替换功能
2. 按照提供的规则进行替换
3. 我协助处理复杂逻辑部分

### 选项 C: 混合方案 (推荐)
- 我完成所有 Repository 接口和实现的改造 (机械性改动,快速)
- 您或我手工处理 Service 层的转换逻辑 (业务逻辑,需谨慎)
- 我生成数据库迁移脚本

## 编译检查

实体层已验证编译通过 ✓

Repository 接口部分改造后,还需要检查:
1. `make build` - 全量编译
2. `make wire` - 重新生成依赖注入

## 下一步建议

请指示:
1. 是否继续让我完成全部改造?
2. 还是采用其他方案?
3. 优先级: Repository 实现 > Service 层 > 数据库迁移
