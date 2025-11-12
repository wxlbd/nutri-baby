# AI分析任务自动处理 - 当前状态和改进方案

## 当前状态 ❌

### 现在的情况

**文件**: `cmd/server/main.go` (第54行)

```go
app.Scheduler.Start()
```

启动了SchedulerService，但：

**文件**: `internal/application/service/scheduler_service.go` (第51-56行)

```go
func (s *SchedulerService) Start() {
    // 启动调度器(用于一次性定时任务)
    s.scheduler.StartAsync()
    s.logger.Info("Scheduler service started (one-time task mode)")
}
```

**问题**: Start()方法只启动了调度器的异步模式，但**没有注册任何自动处理AI分析任务的定时任务**。

### 现状总结

| 处理方式 | 当前状态 | 说明 |
|---------|--------|------|
| 创建任务 | ✅ 可用 | POST /v1/ai-analysis |
| 手动处理 | ✅ 可用 | POST /v1/jobs/process-pending-analyses |
| 自动处理 | ❌ **未实现** | **需要添加定时任务** |

## 改进方案

### 方案A：添加定时处理任务（推荐）

**修改位置**: `internal/application/service/scheduler_service.go`

修改构造函数添加AIAnalysisService依赖：

```go
type SchedulerService struct {
    scheduler           *gocron.Scheduler
    vaccineScheduleRepo repository.BabyVaccineScheduleRepository
    feedingRecordRepo   repository.FeedingRecordRepository
    userRepo            repository.UserRepository
    subscribeService    *SubscribeService
    aiAnalysisService   AIAnalysisService  // ← 添加这行
    strategyFactory     *FeedingReminderStrategyFactory
    logger              *zap.Logger
}

func NewSchedulerService(
    vaccineScheduleRepo repository.BabyVaccineScheduleRepository,
    feedingRecordRepo repository.FeedingRecordRepository,
    userRepo repository.UserRepository,
    subscribeService *SubscribeService,
    aiAnalysisService AIAnalysisService,  // ← 添加这行
    cfg *config.Config,
    logger *zap.Logger,
) *SchedulerService {
    return &SchedulerService{
        // ... 其他字段
        aiAnalysisService: aiAnalysisService,  // ← 添加这行
    }
}
```

修改Start方法添加定时任务：

```go
func (s *SchedulerService) Start() {
    // 启动调度器(用于一次性定时任务)
    s.scheduler.StartAsync()

    // 🆕 添加: 每5分钟自动处理一次待分析任务
    s.scheduler.Every(5).Minutes().Do(func() {
        ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
        defer cancel()

        if err := s.aiAnalysisService.ProcessPendingAnalyses(ctx); err != nil {
            s.logger.Error("自动处理待分析AI任务失败", zap.Error(err))
        } else {
            s.logger.Info("自动处理待分析AI任务成功")
        }
    })

    s.logger.Info("Scheduler service started",
        zap.String("mode", "one-time task + auto processing"))
}
```

**优点**:
- ✅ 自动处理，用户无需手动触发
- ✅ 可配置处理频率（默认5分钟）
- ✅ 单线程执行，不会并发处理
- ✅ 性能可控

### 方案B：在HTTP层添加端点触发定时任务

保持现状，添加一个管理员端点来配置定时任务。

**优点**:
- ✅ 灵活，可动态调整
- ✅ 不需要重启服务

**缺点**:
- ❌ 需要额外的管理界面
- ❌ 用户需要手动配置

## 现有的三种处理方式对比

### 1️⃣ 创建 + 延后处理

```
流程:
POST /v1/ai-analysis (status: pending)
  ↓
(手动或自动触发)
POST /v1/jobs/process-pending-analyses
  ↓
(轮询或等待)
GET /v1/ai-analysis/{id} (status: completed)

当前: ⚠️ 需要手动触发
改进后: ✅ 自动触发（5分钟一次）
```

### 2️⃣ 批量分析（推荐）

```
流程:
POST /v1/ai-analysis/batch (同步处理)
  ↓
立即返回所有4种分析结果

当前: ✅ 完全可用
改进: 无需改进
```

### 3️⃣ 处理待处理任务

```
流程:
创建多个任务
  ↓
(手动或自动触发)
POST /v1/jobs/process-pending-analyses
  ↓
批量处理最多10个任务

当前: ⚠️ 需要手动触发
改进后: ✅ 自动触发（5分钟一次）
```

## 建议方案（实施步骤）

### 步骤1：修改SchedulerService

**文件**: `internal/application/service/scheduler_service.go`

```go
type SchedulerService struct {
    scheduler           *gocron.Scheduler
    vaccineScheduleRepo repository.BabyVaccineScheduleRepository
    feedingRecordRepo   repository.FeedingRecordRepository
    userRepo            repository.UserRepository
    subscribeService    *SubscribeService
    aiAnalysisService   AIAnalysisService    // ← 新增
    strategyFactory     *FeedingReminderStrategyFactory
    logger              *zap.Logger
}

func NewSchedulerService(
    vaccineScheduleRepo repository.BabyVaccineScheduleRepository,
    feedingRecordRepo repository.FeedingRecordRepository,
    userRepo repository.UserRepository,
    subscribeService *SubscribeService,
    aiAnalysisService AIAnalysisService,  // ← 新增
    cfg *config.Config,
    logger *zap.Logger,
) *SchedulerService {
    return &SchedulerService{
        scheduler:           gocron.NewScheduler(time.Local),
        vaccineScheduleRepo: vaccineScheduleRepo,
        feedingRecordRepo:   feedingRecordRepo,
        userRepo:            userRepo,
        subscribeService:    subscribeService,
        aiAnalysisService:   aiAnalysisService,  // ← 新增
        strategyFactory:     NewFeedingReminderStrategyFactory(cfg),
        logger:              logger,
    }
}

func (s *SchedulerService) Start() {
    s.scheduler.StartAsync()

    // 每5分钟自动处理一次待分析任务
    _, err := s.scheduler.Every(5).Minutes().Do(s.processAIAnalysisTasks)
    if err != nil {
        s.logger.Error("添加AI分析定时任务失败", zap.Error(err))
    }

    s.logger.Info("Scheduler service started with auto-processing enabled")
}

// 新增: AI分析任务处理方法
func (s *SchedulerService) processAIAnalysisTasks() {
    ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
    defer cancel()

    if err := s.aiAnalysisService.ProcessPendingAnalyses(ctx); err != nil {
        s.logger.Error("自动处理待分析AI任务失败", zap.Error(err))
        return
    }

    s.logger.Info("自动处理待分析AI任务成功")
}
```

### 步骤2：更新Wire配置

**文件**: `wire/wire.go`

无需修改，AIAnalysisService已配置。但NewSchedulerService的参数需要添加AIAnalysisService。

### 步骤3：重新生成Wire依赖

```bash
cd wire && wire
```

### 步骤4：编译和测试

```bash
go build -o /tmp/app
# 检查是否编译通过
```

## 性能考虑

### 处理能力

- **频率**: 每5分钟处理一次
- **单次处理**: 最多10个任务
- **平均处理时间**: 50-100秒（取决于数据量）

### 优化建议

如果发现处理速度跟不上任务创建速度：

1. **减少处理间隔**
   ```go
   s.scheduler.Every(2).Minutes().Do(...)  // 改为2分钟
   ```

2. **增加批处理大小**
   在 `ai_analysis_service.go:499` 修改：
   ```go
   pendingAnalyses, err := s.aiAnalysisRepo.GetPendingAnalyses(ctx, 20)  // 改为20
   ```

3. **使用工作队列**
   替换为更高效的任务队列系统（如Redis Queue）

## 可配置选项

建议在config.yaml中添加配置：

```yaml
scheduler:
  ai_analysis:
    enabled: true          # 是否启用自动处理
    interval_minutes: 5    # 处理间隔（分钟）
    batch_size: 10         # 单次处理任务数
    timeout_minutes: 4     # 处理超时时间
```

修改代码以读取这个配置：

```go
func (s *SchedulerService) Start() {
    s.scheduler.StartAsync()

    if s.cfg.Scheduler.AIAnalysis.Enabled {
        interval := s.cfg.Scheduler.AIAnalysis.IntervalMinutes
        _, err := s.scheduler.Every(interval).Minutes().Do(s.processAIAnalysisTasks)
        if err != nil {
            s.logger.Error("添加AI分析定时任务失败", zap.Error(err))
        }
    }
}
```

## 总结

**当前状态**:
- ❌ 自动处理：未实现
- ✅ 手动处理：可用
- ✅ 批量分析：可用

**改进方案**:
- ✅ 添加定时处理任务（推荐）
- 每5分钟自动处理待处理任务
- 无需用户手动干预

**实施难度**: 🟢 低（只需添加一个定时任务）

**建议优先级**: 🔴 高（提升用户体验）

---

**关键问题**: 你要我立即实现这个自动处理的定时任务吗？
