# AI分析任务后台运行 - 快速参考

## 🎯 三种任务处理方式对比

| 处理方式 | 触发时机 | 执行模式 | 返回时间 | 适用场景 |
|---------|---------|---------|--------|---------|
| **创建+立即处理** | 创建任务时 | 同步 | 立即 | 单个任务，快速结果 |
| **创建+延后处理** | 手动触发 | 异步 | 延迟 | 大量任务，非实时 |
| **批量处理** | 批量端点 | 同步 | 立即 | 多种分析类型，完整分析 |

## 📊 任务状态流转

```
Created (创建)
    ↓
    ├─ batch=true → BatchAnalyze (同步处理)
    │   ├─ Analyzing (分析中)
    │   ├─ Completed (完成) ✓
    │   └─ Failed (失败) ✗
    │
    └─ batch=false → Pending (等待处理)
        ↓
        ProcessPendingAnalyses 触发
        ↓
        Analyzing (分析中)
        ├─ Completed (完成) ✓
        └─ Failed (失败) ✗
```

## 🚀 快速使用指南

### 1️⃣ 简单场景 - 创建单个任务

```bash
# 创建分析任务（立即返回，后续需要手动处理）
TOKEN="your-jwt-token"

curl -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "feeding",
    "start_date": "2025-11-01",
    "end_date": "2025-11-08"
  }'

# 响应示例
{
  "code": 0,
  "data": {
    "analysis_id": "123",
    "status": "pending",
    "message": "AI分析任务已创建，正在处理中..."
  }
}

# 2秒后检查结果
sleep 2
curl -X GET "http://localhost:8080/v1/ai-analysis/123" \
  -H "Authorization: Bearer $TOKEN"
```

### 2️⃣ 标准场景 - 创建后手动触发处理

```bash
# 1. 创建任务（会加入待处理队列）
curl -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer $TOKEN" \
  -d '{...}'
# 返回: analysis_id="123", status="pending"

# 2. (稍后) 手动触发处理所有待处理任务
curl -X POST http://localhost:8080/jobs/process-pending-analyses \
  -H "Authorization: Bearer $TOKEN"
# 返回: {"code":0,"message":"success"}

# 3. 查询结果（等待完成）
curl -X GET "http://localhost:8080/v1/ai-analysis/123" \
  -H "Authorization: Bearer $TOKEN"
# status=completed, result={...}
```

### 3️⃣ 推荐场景 - 批量分析（立即处理）

```bash
# 一次创建并处理所有分析类型（喂养、睡眠、成长、健康）
curl -X POST "http://localhost:8080/v1/ai-analysis/batch" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '' \
  -G \
  --data-urlencode "baby_id=1" \
  --data-urlencode "start_date=2025-11-01" \
  --data-urlencode "end_date=2025-11-08"

# 立即返回所有分析结果
{
  "code": 0,
  "data": {
    "total_count": 4,
    "completed_count": 4,    // 所有分析都完成了
    "failed_count": 0,
    "analyses": [
      {
        "analysis_id": "124",
        "status": "completed",
        "result": {...}       // 已包含分析结果
      },
      ...
    ]
  }
}
```

## 📝 代码调用示例

### TypeScript/前端

```typescript
// 方式1: 创建后轮询查询
async function analyzeWithPolling(babyId: number, startDate: string, endDate: string) {
  // 1. 创建任务
  const createRes = await fetch('/v1/ai-analysis', {
    method: 'POST',
    headers: {
      'Authorization': `Bearer ${token}`,
      'Content-Type': 'application/json'
    },
    body: JSON.stringify({
      baby_id: babyId,
      analysis_type: 'feeding',
      start_date: startDate,
      end_date: endDate
    })
  });

  const createData = await createRes.json();
  const analysisId = createData.data.analysis_id;

  // 2. 轮询查询结果
  for (let i = 0; i < 30; i++) {
    const queryRes = await fetch(`/v1/ai-analysis/${analysisId}`, {
      headers: { 'Authorization': `Bearer ${token}` }
    });
    const queryData = await queryRes.json();

    if (queryData.data.status === 'completed') {
      return queryData.data.result;
    } else if (queryData.data.status === 'failed') {
      throw new Error(queryData.data.message);
    }

    // 等待2秒再查询
    await new Promise(r => setTimeout(r, 2000));
  }

  throw new Error('分析超时');
}

// 使用
const result = await analyzeWithPolling(1, '2025-11-01', '2025-11-08');
console.log('分析完成:', result);
```

```typescript
// 方式2: 批量分析（推荐，立即返回结果）
async function batchAnalyze(babyId: number, startDate: string, endDate: string) {
  const response = await fetch(`/v1/ai-analysis/batch?baby_id=${babyId}&start_date=${startDate}&end_date=${endDate}`, {
    method: 'POST',
    headers: { 'Authorization': `Bearer ${token}` }
  });

  const data = await response.json();

  if (data.code === 0) {
    // 所有分析都已完成，直接使用结果
    data.data.analyses.forEach(analysis => {
      console.log(`${analysis.analysis_type}: ${analysis.result.score}`);
    });

    return data.data;
  }

  throw new Error(data.message);
}

// 使用
const results = await batchAnalyze(1, '2025-11-01', '2025-11-08');
```

### Go/后端

```go
// 创建分析任务
analysisReq := &service.CreateAnalysisRequest{
    BabyID:       1,
    AnalysisType: entity.AIAnalysisTypeFeeding,
    StartDate:    service.CustomTime(time.Now().AddDate(0, 0, -7)),
    EndDate:      service.CustomTime(time.Now()),
}

analysis, err := aiAnalysisService.CreateAnalysis(ctx, analysisReq)
if err != nil {
    return err
}

// 如果需要立即处理
analysisID, _ := strconv.ParseInt(analysis.AnalysisID, 10, 64)
err = aiAnalysisService.ProcessPendingAnalyses(ctx)

// 查询结果
result, err := aiAnalysisService.GetAnalysisResult(ctx, analysis.AnalysisID)
if result.Status == entity.AIAnalysisStatusCompleted {
    fmt.Printf("分析得分: %f\n", result.Result.Score)
}
```

## 🔍 监控和调试

### 查看待处理任务

```bash
# 查询数据库中的待处理任务
psql -U postgres -d nutri_baby -c "
SELECT id, baby_id, analysis_type, status, created_at
FROM ai_analyses
WHERE status IN ('pending', 'analyzing')
ORDER BY created_at DESC
LIMIT 10;
"
```

### 查看处理日志

```bash
# 查看AI分析相关的日志
tail -f logs/app.log | grep -i "ai分析\|分析任务\|处理失败"

# 示例输出:
# 2025-11-12T19:21:34.123+0800	info	ai_analysis_service	创建AI分析任务	{"analysis_id":123,"baby_id":1,"analysis_type":"feeding"}
# 2025-11-12T19:21:36.456+0800	info	ai_analysis_service	AI分析任务完成	{"analysis_id":123,"score":85.5}
```

### 强制触发任务处理

```bash
# 使用后台任务端点
curl -X POST http://localhost:8080/jobs/process-pending-analyses \
  -H "Authorization: Bearer $TOKEN" \
  -v  # 显示详细日志

# 查看返回状态码
# 200 OK = 处理成功
# 400 Bad Request = 参数错误
# 401 Unauthorized = 无权限
# 500 Internal Server Error = 服务器错误（查看日志）
```

## ⚙️ 配置调整

### 调整批处理大小

**当前**: 每次最多处理10个任务

```go
// 文件: internal/application/service/ai_analysis_service.go:499
pendingAnalyses, err := s.aiAnalysisRepo.GetPendingAnalyses(ctx, 10)  // ← 修改这里

// 建议值:
// - 低功率服务器: 1-5
// - 普通服务器: 10 (默认)
// - 高性能服务器: 20-50
```

### 添加自动定时处理

```go
// 文件: internal/application/service/scheduler_service.go:52
func (s *SchedulerService) Start() {
    // 原代码
    s.scheduler.StartAsync()

    // 添加以下代码实现每5分钟自动处理一次
    s.scheduler.Every(5).Minutes().Do(func() {
        ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
        defer cancel()

        if err := s.aiAnalysisService.ProcessPendingAnalyses(ctx); err != nil {
            s.logger.Error("后台处理待分析任务失败", zap.Error(err))
        }
    })

    s.logger.Info("AI分析自动处理任务已启用 (每5分钟一次)")
}
```

## 🐛 常见问题

**Q: 为什么任务一直是 pending？**
A: 因为没有手动触发 ProcessPendingAnalyses。调用端点：
```bash
curl -X POST http://localhost:8080/jobs/process-pending-analyses -H "Authorization: Bearer $TOKEN"
```

**Q: 分析需要多长时间？**
A: 取决于：
- 数据量（记录数）
- AI模型响应时间（1-10秒）
- 通常总耗时: 10秒 - 2分钟

**Q: 能否取消正在处理的任务？**
A: 当前不支持。可以在处理失败后重新创建任务。

**Q: 同时创建多个任务效率高吗？**
A: 不高。建议使用批量分析端点一次处理所有类型。

## 📚 相关文档

- [AI_ANALYSIS_API.md](./AI_ANALYSIS_API.md) - 完整API参考
- [AI_ANALYSIS_QUICK_START.md](./AI_ANALYSIS_QUICK_START.md) - 快速开始
- [TEST_REPORT_20251112.md](./TEST_REPORT_20251112.md) - 测试报告

---

**关键要点**:
1. ✅ 分析任务是**非阻塞**的（创建后立即返回）
2. ✅ 后台处理可以**手动触发**或**自动定时执行**（未来实现）
3. ✅ 推荐使用**批量分析端点**获得最佳性能
4. ✅ 始终**检查任务状态**而不是假设立即完成

**最后更新**: 2025-11-12
