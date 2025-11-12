# AI分析任务后台运行机制

## 概述

AI分析任务的后台处理采用**混合模式**：
- **创建阶段**: 立即返回给用户（异步处理）
- **处理阶段**: 由定时任务或手动触发处理
- **状态查询**: 用户随时可查询任务状态和结果

## 任务生命周期

```
┌─────────────────────────────────────────────────────────────┐
│                    AI分析任务生命周期                         │
└─────────────────────────────────────────────────────────────┘

1. 创建任务
   ┌─────────────────────────────────────────────────────────┐
   │ 用户请求: POST /v1/ai-analysis                         │
   │ ↓                                                       │
   │ 1. 验证宝宝是否存在                                    │
   │ 2. 创建分析记录 (Status: pending)                      │
   │ 3. 保存到数据库                                        │
   │ 4. 立即返回 AnalysisID 给用户                          │
   └─────────────────────────────────────────────────────────┘

2. 后台处理 (ProcessPendingAnalyses 或 BatchAnalyze)
   ┌─────────────────────────────────────────────────────────┐
   │ 触发: 定时任务 或 BatchAnalyze 端点                   │
   │ ↓                                                       │
   │ 1. 获取所有 Status=pending 的任务                      │
   │ 2. 对每个任务执行 processAnalysis:                    │
   │    - 更新状态: analyzing                              │
   │    - 收集数据 (宝宝信息、记录等)                      │
   │    - 调用 Eino 链进行AI分析                           │
   │    - 序列化结果                                       │
   │    - 更新状态: completed                              │
   │ 3. 错误处理: 更新状态为 failed                        │
   └─────────────────────────────────────────────────────────┘

3. 结果查询
   ┌─────────────────────────────────────────────────────────┐
   │ 用户请求: GET /v1/ai-analysis/{id}                    │
   │ ↓                                                       │
   │ 返回任务状态和结果 (如已完成)                         │
   └─────────────────────────────────────────────────────────┘
```

## 代码架构

### 1. 任务创建 (CreateAnalysis)

**文件**: `internal/application/service/ai_analysis_service.go:175`

```go
func (s *aiAnalysisServiceImpl) CreateAnalysis(ctx context.Context, req *CreateAnalysisRequest) (*AnalysisResponse, error) {
    // 1. 验证宝宝
    _, err := s.babyRepo.FindByID(ctx, req.BabyID)

    // 2. 创建分析记录
    analysis := &entity.AIAnalysis{
        BabyID:       req.BabyID,
        AnalysisType: req.AnalysisType,
        Status:       entity.AIAnalysisStatusPending,  // ← 初始状态
        StartDate:    req.StartDate.Time(),
        EndDate:      req.EndDate.Time(),
    }

    // 3. 保存到数据库
    if err := s.aiAnalysisRepo.Create(ctx, analysis); err != nil {
        return nil, err
    }

    // 4. 立即返回给用户
    return &AnalysisResponse{
        AnalysisID: strconv.FormatInt(analysis.ID, 10),
        Status:     analysis.Status,
        Message:    "AI分析任务已创建，正在处理中...",
    }, nil
}
```

**特点**:
- ✅ 非阻塞：创建后立即返回
- ✅ 返回分析ID供后续查询
- ✅ 状态初始为 `pending`

### 2. 后台处理 (ProcessPendingAnalyses)

**文件**: `internal/application/service/ai_analysis_service.go:497`

```go
func (s *aiAnalysisServiceImpl) ProcessPendingAnalyses(ctx context.Context) error {
    // 1. 获取待处理任务 (最多10个)
    pendingAnalyses, err := s.aiAnalysisRepo.GetPendingAnalyses(ctx, 10)

    // 2. 遍历处理每个任务
    for _, analysis := range pendingAnalyses {
        if err := s.processAnalysis(ctx, analysis.ID); err != nil {
            s.logger.Error("处理分析任务失败", zap.Error(err))
            continue  // 继续处理其他任务
        }
    }
    return nil
}

func (s *aiAnalysisServiceImpl) processAnalysis(ctx context.Context, analysisID int64) error {
    // 1. 更新状态为 analyzing
    s.aiAnalysisRepo.UpdateStatus(ctx, analysisID, entity.AIAnalysisStatusAnalyzing)

    // 2. 获取分析记录
    analysis, err := s.aiAnalysisRepo.GetByID(ctx, analysisID)

    // 3. 收集数据 (宝宝信息、喂养/睡眠/成长等记录)
    data, err := s.collectAnalysisData(ctx, analysis)

    // 4. 调用AI进行分析
    result, err := s.chainBuilder.Analyze(ctx, analysis, data)

    // 5. 序列化和保存结果
    resultJSON, _ := json.Marshal(result)
    s.aiAnalysisRepo.UpdateResult(ctx, analysisID, string(resultJSON),
        entity.AIAnalysisStatusCompleted)

    return nil
}
```

**流程**:
1. 获取最多10个待处理任务
2. 逐个处理，失败则跳过，继续处理下一个
3. 每个任务的处理步骤：
   - 更新状态为 `analyzing`
   - 收集所需数据
   - 调用Eino链进行AI分析
   - 保存结果，更新状态为 `completed`

### 3. 批量处理 (BatchAnalyze)

**文件**: `internal/application/service/ai_analysis_service.go:290`

```go
func (s *aiAnalysisServiceImpl) BatchAnalyze(ctx context.Context, babyID string,
    startDate, endDate time.Time) (*BatchAnalysisResponse, error) {

    analysisTypes := []entity.AIAnalysisType{
        entity.AIAnalysisTypeFeeding,
        entity.AIAnalysisTypeSleep,
        entity.AIAnalysisTypeGrowth,
        entity.AIAnalysisTypeHealth,
    }

    for _, analysisType := range analysisTypes {
        // 1. 创建任务
        analysis, _ := s.CreateAnalysis(ctx, &CreateAnalysisRequest{
            BabyID:       id,
            AnalysisType: analysisType,
            StartDate:    CustomTime(startDate),
            EndDate:      CustomTime(endDate),
        })

        // 2. 立即处理（同步）
        analysisID, _ := strconv.ParseInt(analysis.AnalysisID, 10, 64)
        if err := s.processAnalysis(ctx, analysisID); err != nil {
            // 处理失败计数
        } else {
            // 处理成功计数
        }
    }

    return &BatchAnalysisResponse{...}
}
```

**特点**:
- 创建4种分析类型的任务（喂养、睡眠、成长、健康）
- 创建后立即处理（同步执行）
- 返回每种分析的状态和结果

### 4. 定时调度 (Scheduler)

**文件**: `internal/application/service/scheduler_service.go`

```go
// 应用启动时
func (s *SchedulerService) Start() {
    s.scheduler.StartAsync()  // 启动异步调度器
}

// 后台任务触发
func ProcessPendingAnalysesJob(aiAnalysisService service.AIAnalysisService) gin.HandlerFunc {
    return func(c *gin.Context) {
        // 处理所有待分析任务
        if err := aiAnalysisService.ProcessPendingAnalyses(c.Request.Context()); err != nil {
            // 错误处理
        }
    }
}
```

## 任务触发方式

### 方式1: 创建时立即处理 (BatchAnalyze)

```bash
# 请求
POST /v1/ai-analysis/batch?baby_id=1&start_date=2025-11-01&end_date=2025-11-08

# 响应 (立即返回结果，因为是同步处理)
{
  "code": 0,
  "data": {
    "completed_count": 4,    // 4种分析都完成了
    "failed_count": 0
  }
}
```

### 方式2: 创建任务后手动触发处理

```bash
# 1. 创建任务
curl -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer TOKEN" \
  -d '{"baby_id":1,"analysis_type":"feeding","start_date":"2025-11-01","end_date":"2025-11-08"}'
# 响应: {"analysis_id":"1","status":"pending"}

# 2. 稍后手动触发处理
curl -X POST http://localhost:8080/jobs/process-pending-analyses \
  -H "Authorization: Bearer TOKEN"
# 响应: 处理结果

# 3. 查询结果
curl -X GET http://localhost:8080/v1/ai-analysis/1 \
  -H "Authorization: Bearer TOKEN"
# 响应: {"status":"completed","result":{...}}
```

### 方式3: 定时自动处理 (未来实现)

```go
// 建议的实现方式
func (s *SchedulerService) Start() {
    s.scheduler.Every(5).Minutes().Do(func() {
        if err := s.aiAnalysisService.ProcessPendingAnalyses(context.Background()); err != nil {
            s.logger.Error("处理待分析任务失败", zap.Error(err))
        }
    })
    s.scheduler.StartAsync()
}
```

## 数据库状态变迁

```
创建时:
┌──────────┐
│ pending  │  ← 初始状态
└─────┬────┘
      │ ProcessPendingAnalyses 触发
      ↓
┌──────────┐
│analyzing │  ← 处理中
└─────┬────┘
      │ 分析完成
      ↓
┌──────────┐
│completed │  ← 完成，result 已保存
└──────────┘

处理失败时:
┌──────────┐
│analyzing │
└─────┬────┘
      │ 分析失败
      ↓
┌──────────┐
│ failed   │  ← 失败，error_message 已保存
└──────────┘
```

## 数据库表结构

```sql
-- ai_analyses 表
CREATE TABLE ai_analyses (
    id BIGINT PRIMARY KEY,
    baby_id BIGINT,
    analysis_type VARCHAR(50),          -- feeding, sleep, growth, health, behavior
    status VARCHAR(20),                 -- pending, analyzing, completed, failed
    start_date TIMESTAMP,
    end_date TIMESTAMP,
    result TEXT,                        -- JSON 格式的分析结果
    error_message TEXT,                 -- 失败原因
    created_at TIMESTAMP,
    updated_at TIMESTAMP
);

-- 状态字段的含义
pending    → 已创建，等待处理
analyzing  → 正在进行AI分析
completed  → 分析完成，result 包含结果
failed     → 分析失败，error_message 包含错误原因
```

## 最佳实践

### 1. 监听分析进度

```typescript
// 前端代码
async function waitForAnalysisCompletion(analysisId: string, maxRetries = 60) {
    for (let i = 0; i < maxRetries; i++) {
        const response = await fetch(`/v1/ai-analysis/${analysisId}`, {
            headers: { 'Authorization': `Bearer ${token}` }
        });
        const data = await response.json();

        if (data.data.status === 'completed') {
            console.log('分析完成:', data.data.result);
            return data.data.result;
        } else if (data.data.status === 'failed') {
            throw new Error('分析失败: ' + data.data.message);
        }

        // 等待2秒再查询
        await new Promise(r => setTimeout(r, 2000));
    }
    throw new Error('分析超时');
}

// 使用
const result = await waitForAnalysisCompletion('123');
```

### 2. 处理长运行时间

```typescript
// 如果分析可能需要很长时间，使用 WebSocket 推送更新（未来实现）
const ws = new WebSocket(`wss://localhost:8080/v1/ai-analysis/123/subscribe?token=${token}`);
ws.onmessage = (event) => {
    const update = JSON.parse(event.data);
    console.log('状态更新:', update.status);
    if (update.status === 'completed') {
        console.log('结果:', update.result);
    }
};
```

### 3. 错误处理

```typescript
// 获取失败的原因
const response = await fetch(`/v1/ai-analysis/123`, {
    headers: { 'Authorization': `Bearer ${token}` }
});
const data = await response.json();

if (data.data.status === 'failed') {
    // 记录错误信息
    console.error('分析失败原因:', data.data.message);

    // 重新创建任务
    const newAnalysis = await createAnalysis({...});
}
```

## 性能考虑

### 1. 任务批大小

```go
// 当前实现：每次处理最多 10 个任务
pendingAnalyses, _ := s.aiAnalysisRepo.GetPendingAnalyses(ctx, 10)

// 建议：根据服务器能力调整
// - 高性能服务器: 调整为 20-50
// - 普通服务器: 保持 10
// - 低功率服务器: 调整为 1-5
```

### 2. 超时设置

```go
// 当前: 无超时限制
// 建议: 添加超时保护

ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
defer cancel()

result, err := s.chainBuilder.Analyze(ctx, analysis, data)
```

### 3. 并发控制

```go
// 当前: 串行处理
for _, analysis := range pendingAnalyses {
    s.processAnalysis(ctx, analysis.ID)
}

// 建议: 使用 goroutine 并发处理（需要防止竞态条件）
for _, analysis := range pendingAnalyses {
    go s.processAnalysis(ctx, analysis.ID)
}
```

## 后续改进方向

- [ ] 实现后台定时任务 (每5分钟自动处理)
- [ ] 添加 WebSocket 实时进度推送
- [ ] 实现分析结果缓存 (1小时)
- [ ] 添加任务重试机制 (失败自动重试3次)
- [ ] 实现任务优先级队列
- [ ] 添加任务处理监控面板
- [ ] 支持取消正在处理的任务

## 故障排除

### 问题1: 任务一直是 pending 状态

**原因**: 没有触发 ProcessPendingAnalyses
**解决**:
```bash
# 手动触发一次
curl -X POST http://localhost:8080/jobs/process-pending-analyses \
  -H "Authorization: Bearer TOKEN"

# 或检查后台日志
tail -f logs/app.log | grep "处理分析任务"
```

### 问题2: 任务分析失败

**检查**:
- 宝宝ID是否有效
- 日期范围是否合理
- 是否有足够的数据进行分析
- AI服务是否正常

### 问题3: 内存占用过高

**原因**: 一次性处理过多任务
**解决**:
- 减少 `GetPendingAnalyses` 的数量
- 增加处理频率（更小的批次）
- 实现流式处理而不是全量加载

---

**最后更新**: 2025-11-12
