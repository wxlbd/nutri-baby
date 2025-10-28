# 疫苗提醒超期问题修复

## 问题描述

超期疫苗（超过预定接种日期）未在 `/v1/babies/{babyId}/vaccine-reminders` 接口返回，导致前端无法显示超期提醒。

### 问题原因

疫苗提醒系统中，**提醒状态是静态存储在数据库中的**，而不是动态计算的：

1. **创建时存储状态** - 在 `InitializeVaccineReminders()` 时，根据当前时间计算状态并存储为 `upcoming`
2. **查询时直接返回存储值** - `FindByStatus()` 只查询数据库中该状态的记录，不重新计算
3. **状态永不更新** - 当真实时间推进导致疫苗超期时，数据库中的状态仍是旧值

### 具体流程图

```
创建提醒 (初始化时)
  ↓
计算预定日期 = 出生日期 + 月龄
  ↓
调用 UpdateStatus() 计算状态 → "upcoming" (存入数据库)
  ↓
时间推进，当前日期超过预定日期
  ↓
查询提醒接口 (FindByStatus("overdue"))
  ↓
❌ 数据库查询条件: WHERE status = 'overdue'
  ↓
无法找到该疫苗记录 (因为数据库中仍是 'upcoming')
```

## 修复方案

### 修改文件

**文件**: `nutri-baby-server/internal/application/service/vaccine_service.go`

**方法**: `GetVaccineReminders()` (第154-216行)

### 核心变化

修复后的流程：

```go
// 1. 查询出宝宝的所有未完成提醒
reminders, err := s.vaccineReminderRepo.FindByBabyID(ctx, babyID)

// 2. 对每个提醒实时计算状态
for _, reminder := range reminders {
    oldStatus := reminder.Status
    reminder.UpdateStatus()  // ← 实时计算状态

    // 3. 如果状态变化，更新到数据库
    if oldStatus != reminder.Status {
        s.vaccineReminderRepo.UpdateStatus(ctx, reminder.ReminderID, reminder.Status)
    }

    // 4. 过滤出请求的状态
    if status != "" && reminder.Status != status {
        continue
    }
}
```

### 修复前后对比

#### ❌ 修复前 (有问题)
```go
reminders, err := s.vaccineReminderRepo.FindByStatus(ctx, babyID, status, limit)
// 问题: 直接查询指定状态的记录，没有实时计算
// 如果一个记录的 status='upcoming' 但已经超期了，查询 status='overdue' 就找不到
```

#### ✅ 修复后 (正常)
```go
// 1. 获取所有未完成的提醒
reminders, err := s.vaccineReminderRepo.FindByBabyID(ctx, babyID)

// 2. 实时计算每个提醒的状态
for _, reminder := range reminders {
    oldStatus := reminder.Status
    reminder.UpdateStatus()  // 重新计算状态

    // 3. 如果状态变化，同步到数据库
    if oldStatus != reminder.Status && reminder.Status != entity.ReminderStatusCompleted {
        s.vaccineReminderRepo.UpdateStatus(ctx, reminder.ReminderID, reminder.Status)
    }

    // 4. 过滤返回的记录
    if status != "" && reminder.Status != status {
        continue
    }
}
```

## UpdateStatus() 逻辑

`UpdateStatus()` 方法在 `vaccine.go` 中定义 (第124-139行)：

```go
// UpdateStatus 更新提醒状态
func (v *VaccineReminder) UpdateStatus() {
	days := v.DaysUntilDue()  // 计算距离预定日期的天数

	if v.Status == ReminderStatusCompleted {
		return  // 已完成不再更新
	}

	if days > 7 {
		v.Status = ReminderStatusUpcoming    // 超过7天
	} else if days >= 0 {
		v.Status = ReminderStatusDue         // 0-7天内
	} else {
		v.Status = ReminderStatusOverdue     // 已经超期
	}
}
```

**状态转换规则**:
- `days > 7`: `upcoming` (即将到期)
- `0 <= days <= 7`: `due` (应接种)
- `days < 0`: `overdue` (已逾期)

## API 行为变化

### 查询超期疫苗

**请求**:
```
GET /v1/babies/7430d098-7b37-499a-8440-83a21b992bf9/vaccine-reminders?status=overdue
```

**修复前**: 返回空列表（如果超期疫苗的 `status` 还是 `upcoming`）

**修复后**: 返回所有已超期的疫苗，包括那些状态刚从 `upcoming` 转为 `overdue` 的

### 查询所有未完成提醒

**请求**:
```
GET /v1/babies/7430d098-7b37-499a-8440-83a21b992bf9/vaccine-reminders
```

**修复前**: 返回各种状态的提醒，但状态可能不准确

**修复后**: 返回所有未完成的提醒，且状态始终是最新的实时计算值

## 性能考虑

### 优化建议

当前实现会对每个查询都重新计算状态。如果需要进一步优化：

1. **定时后台任务** - 定期运行任务更新所有过期的提醒
   ```go
   // CheckVaccineReminders() 可以定期执行
   // 更新所有状态变化的提醒
   ```

2. **数据库查询优化** - 直接在 SQL 层计算状态
   ```sql
   SELECT *,
          CASE
              WHEN scheduled_date > NOW() + interval '7 days' THEN 'upcoming'
              WHEN scheduled_date BETWEEN NOW() AND NOW() + interval '7 days' THEN 'due'
              WHEN scheduled_date < NOW() THEN 'overdue'
              ELSE 'completed'
          END as actual_status
   FROM vaccine_reminders
   WHERE baby_id = ? AND actual_status = ?
   ```

目前的修复采取折中方案：在应用层实时计算，按需更新数据库，保证数据最终一致。

## 测试验证

### 测试场景

1. **新创建的疫苗提醒** - 状态应为 `upcoming`
2. **接近预定日期的提醒** (7天内) - 状态应为 `due`
3. **超过预定日期的提醒** - 状态应为 `overdue`
4. **已接种记录对应的提醒** - 状态应为 `completed`

### 手动测试步骤

```bash
# 1. 查询某个宝宝的所有提醒
curl "http://localhost:8080/v1/babies/{babyId}/vaccine-reminders"

# 2. 查询超期提醒
curl "http://localhost:8080/v1/babies/{babyId}/vaccine-reminders?status=overdue"

# 3. 查询应接种提醒
curl "http://localhost:8080/v1/babies/{babyId}/vaccine-reminders?status=due"

# 4. 查询即将到期的提醒
curl "http://localhost:8080/v1/babies/{babyId}/vaccine-reminders?status=upcoming"
```

## 相关文件

- `nutri-baby-server/internal/application/service/vaccine_service.go` - 修复的服务类
- `nutri-baby-server/internal/domain/entity/vaccine.go` - UpdateStatus() 定义
- `nutri-baby-server/internal/infrastructure/persistence/vaccine_reminder_repository_impl.go` - 数据库操作
- `nutri-baby-app/src/pages/vaccine/vaccine.vue` - 前端显示逻辑
- `nutri-baby-app/src/api/vaccine.ts` - 前端 API 调用

## 部署建议

1. **数据库一致性** - 修复后，建议运行一次清理任务，更新所有过期的提醒状态
   ```sql
   -- 更新所有已超期但状态不正确的提醒
   UPDATE vaccine_reminders
   SET status = 'overdue'
   WHERE scheduled_date < EXTRACT(EPOCH FROM NOW()) * 1000
   AND status IN ('upcoming', 'due')
   AND status != 'completed'
   ```

2. **灰度发布** - 先在开发环境验证，再逐步发布到生产环境

3. **日志监控** - 监控 `GetVaccineReminders` 的状态变更日志，确认修复生效

## 总结

| 问题 | 原因 | 解决方案 |
|------|------|--------|
| 超期疫苗未返回 | 状态静态存储，不动态计算 | 每次查询时实时调用 UpdateStatus() |
| 数据库状态过期 | 状态一次性存储后永不更新 | 状态变化时同步更新数据库 |
| 提醒状态不准确 | 没有考虑时间推进 | 在应用层实现实时计算 |
