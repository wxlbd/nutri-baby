# AI分析后台任务处理 - 完整实现指南

## 概述

AI分析后台任务处理已完全实现，支持三种处理方式：

1. **创建 + 延后处理**：创建任务后立即返回，需要手动或自动触发处理
2. **批量分析（推荐）**：一次创建并处理4种分析类型，同步返回所有结果
3. **处理待处理任务**：批量处理所有pending状态的任务

## 路由配置

后台任务处理端点已在路由中注册：

**文件**: `internal/interface/http/router/router.go` (第171行)

```go
// 后台任务（需要认证）
handler.RegisterBackgroundJobs(authRequired, aiAnalysisService, logger)
```

### 完整端点列表

| 端点 | 方法 | 用途 | 认证 |
|------|------|------|------|
| `/v1/ai-analysis` | POST | 创建分析任务 | ✅ |
| `/v1/ai-analysis/{id}` | GET | 获取分析结果 | ✅ |
| `/v1/ai-analysis/batch` | POST | 批量分析（推荐） | ✅ |
| `/v1/jobs/process-pending-analyses` | POST | 处理待处理任务 | ✅ |

## 任务生命周期

### 状态流转图

```
创建任务
  ↓
pending (待处理)
  ↓ 触发处理
analyzing (分析中)
  ↓
✓ completed (已完成)    或    ✗ failed (失败)
```

### 状态说明

- **pending**: 任务已创建，等待处理
- **analyzing**: 任务正在被AI分析
- **completed**: 分析成功，结果可用
- **failed**: 分析失败，包含错误信息

## 使用示例

### 方式1：创建 + 手动处理

**场景**: 批量创建多个任务，然后统一处理

**步骤1：创建任务**

```bash
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
```

**步骤2：（可选）检查任务状态**

```bash
curl -X GET http://localhost:8080/v1/ai-analysis/123 \
  -H "Authorization: Bearer $TOKEN"

# 响应示例
{
  "code": 0,
  "data": {
    "analysis_id": "123",
    "status": "pending"  # 状态为pending
  }
}
```

**步骤3：手动触发处理**

```bash
curl -X POST http://localhost:8080/v1/jobs/process-pending-analyses \
  -H "Authorization: Bearer $TOKEN"

# 响应示例
{
  "code": 0,
  "message": "success"
}
```

**步骤4：等待处理完成后查询结果**

```bash
# 等待2-3秒让AI完成分析
sleep 3

curl -X GET http://localhost:8080/v1/ai-analysis/123 \
  -H "Authorization: Bearer $TOKEN"

# 响应示例
{
  "code": 0,
  "data": {
    "analysis_id": "123",
    "status": "completed",
    "result": {
      "score": 85,
      "insights": [...],
      "alerts": [...],
      "patterns": [...]
    }
  }
}
```

### 方式2：批量分析（推荐）

**场景**: 需要获取完整分析结果，批量分析4种类型

**请求**

```bash
curl -X POST "http://localhost:8080/v1/ai-analysis/batch?baby_id=1&start_date=2025-11-01&end_date=2025-11-08" \
  -H "Authorization: Bearer $TOKEN"
```

**响应示例**

```json
{
  "code": 0,
  "data": {
    "total_count": 4,
    "completed_count": 4,
    "failed_count": 0,
    "analyses": [
      {
        "analysis_id": "124",
        "status": "completed",
        "analysis_type": "feeding",
        "result": {
          "score": 85,
          "insights": [
            {
              "type": "feeding",
              "title": "喂养规律良好",
              "description": "宝宝的喂养时间较为规律",
              "priority": "medium"
            }
          ]
        }
      },
      {
        "analysis_id": "125",
        "status": "completed",
        "analysis_type": "sleep",
        "result": {...}
      },
      {
        "analysis_id": "126",
        "status": "completed",
        "analysis_type": "growth",
        "result": {...}
      },
      {
        "analysis_id": "127",
        "status": "completed",
        "analysis_type": "health",
        "result": {...}
      }
    ]
  }
}
```

## 后端实现

### 核心服务

**文件**: `internal/application/service/ai_analysis_service.go`

#### CreateAnalysis (第175行)

创建分析任务，返回pending状态。

```go
func (s *aiAnalysisServiceImpl) CreateAnalysis(ctx context.Context, req *CreateAnalysisRequest) (*AnalysisResponse, error) {
    // 创建分析记录
    analysis := &entity.AIAnalysis{
        BabyID:       req.BabyID,
        AnalysisType: req.AnalysisType,
        Status:       entity.AIAnalysisStatusPending,  // 初始状态
        StartDate:    req.StartDate.Time(),
        EndDate:      req.EndDate.Time(),
    }
    // 保存到数据库并立即返回
    return &AnalysisResponse{
        AnalysisID: strconv.FormatInt(analysis.ID, 10),
        Status:     analysis.Status,
        Message:    "AI分析任务已创建，正在处理中...",
    }, nil
}
```

**特点**:
- 非阻塞：立即返回，不等待分析完成
- 返回分析ID用于查询进度

#### ProcessPendingAnalyses (第497行)

批量处理待处理的分析任务。

```go
func (s *aiAnalysisServiceImpl) ProcessPendingAnalyses(ctx context.Context) error {
    // 最多取10个待处理任务
    pendingAnalyses, err := s.aiAnalysisRepo.GetPendingAnalyses(ctx, 10)

    // 逐个处理
    for _, analysis := range pendingAnalyses {
        if err := s.processAnalysis(ctx, analysis.ID); err != nil {
            // 继续处理其他任务，不因为一个失败而停止
            continue
        }
    }
    return nil
}
```

**特点**:
- 一次处理最多10个任务
- 错误隔离：一个任务失败不影响其他任务

#### processAnalysis (第519行)

处理单个分析任务的核心逻辑。

```go
func (s *aiAnalysisServiceImpl) processAnalysis(ctx context.Context, analysisID int64) error {
    // 1. 更新状态为analyzing
    s.aiAnalysisRepo.UpdateStatus(ctx, analysisID, entity.AIAnalysisStatusAnalyzing)

    // 2. 收集分析所需数据
    data, err := s.collectAnalysisData(ctx, analysis)

    // 3. 使用Eino链进行AI分析
    result, err := s.chainBuilder.Analyze(ctx, analysis, data)

    // 4. 保存结果
    s.aiAnalysisRepo.UpdateResult(ctx, analysisID, string(resultJSON),
        entity.AIAnalysisStatusCompleted)
    return nil
}
```

**处理流程**:
1. 状态: pending → analyzing
2. 收集数据：宝宝信息、喂养/睡眠/成长记录
3. 调用Eino AI链进行分析
4. 状态: analyzing → completed (或 failed)

#### BatchAnalyze (第290行)

一次创建并处理4种分析类型（喂养、睡眠、成长、健康）。

```go
func (s *aiAnalysisServiceImpl) BatchAnalyze(ctx context.Context, babyID string,
    startDate, endDate time.Time) (*BatchAnalysisResponse, error) {

    analysisTypes := []entity.AIAnalysisType{
        entity.AIAnalysisTypeFeeding,
        entity.AIAnalysisTypeSleep,
        entity.AIAnalysisTypeGrowth,
        entity.AIAnalysisTypeHealth,
    }

    var analyses []*AnalysisResponse
    for _, analysisType := range analysisTypes {
        // 创建任务
        analysis, _ := s.CreateAnalysis(ctx, ...)

        // 立即处理（同步）
        s.processAnalysis(ctx, analysisID)

        analyses = append(analyses, analysis)
    }

    return &BatchAnalysisResponse{
        Analyses:       analyses,
        TotalCount:     len(analyses),
        CompletedCount: completedCount,
    }, nil
}
```

**特点**:
- 同步处理：等待所有分析完成后返回
- 返回完整的分析结果
- 推荐用于需要立即获取完整分析的场景

### HTTP处理器

**文件**: `internal/interface/http/handler/ai_analysis_handler.go`

#### ProcessPendingAnalysesJob (第403行)

HTTP处理器，调用ProcessPendingAnalyses服务。

```go
func ProcessPendingAnalysesJob(aiAnalysisService service.AIAnalysisService, logger *zap.Logger) gin.HandlerFunc {
    return func(c *gin.Context) {
        if err := aiAnalysisService.ProcessPendingAnalyses(c.Request.Context()); err != nil {
            logger.Error("处理待分析AI任务失败", zap.Error(err))
            response.Error(c, err)
            return
        }
        response.Success(c, nil)
    }
}
```

#### RegisterBackgroundJobs (第415行)

注册后台任务路由。

```go
func RegisterBackgroundJobs(router *gin.RouterGroup, aiAnalysisService service.AIAnalysisService, logger *zap.Logger) {
    router.POST("/jobs/process-pending-analyses", ProcessPendingAnalysesJob(aiAnalysisService, logger))
}
```

## 路由注册

**文件**: `internal/interface/http/router/router.go` (第171行)

```go
// 后台任务（需要认证）
handler.RegisterBackgroundJobs(authRequired, aiAnalysisService, logger)
```

在NewRouter函数中，将RegisterBackgroundJobs注册到认证的路由组中。

## 依赖注入

**文件**: `wire/wire.go`

所有依赖已在Wire中配置：

- `logger.NewLogger` - 日志系统
- `service.NewAIAnalysisService` - AI分析服务
- `handler.NewAIAnalysisHandler` - AI分析处理器
- `persistence.NewAIAnalysisRepository` - 数据访问层

Wire自动生成完整的依赖图，在NewRouter时注入所需依赖。

## 数据库

### 表结构

**表**: `ai_analyses`

| 字段 | 类型 | 说明 |
|------|------|------|
| id | INT64 | 主键 |
| baby_id | INT64 | 宝宝ID |
| analysis_type | VARCHAR | 分析类型（feeding/sleep/growth/health/behavior） |
| status | VARCHAR | 状态（pending/analyzing/completed/failed） |
| start_date | TIMESTAMP | 分析开始日期 |
| end_date | TIMESTAMP | 分析结束日期 |
| result | TEXT | 分析结果（JSON格式） |
| created_at | TIMESTAMP | 创建时间 |
| updated_at | TIMESTAMP | 更新时间 |

### 查询待处理任务

```sql
SELECT * FROM ai_analyses
WHERE status = 'pending'
ORDER BY created_at ASC
LIMIT 10;
```

### 查看任务状态分布

```sql
SELECT status, COUNT(*) as count
FROM ai_analyses
GROUP BY status;
```

## 性能考虑

### 处理能力

- **单次处理**: 最多10个任务
- **AI分析耗时**: 1-10秒/任务
- **单轮处理耗时**: 10-100秒（10个任务）

### 优化建议

1. **调整批处理大小**（`ai_analysis_service.go:499`）
   ```go
   pendingAnalyses, err := s.aiAnalysisRepo.GetPendingAnalyses(ctx, 10)  // ← 修改这里
   ```
   - 低功率服务器: 1-5
   - 普通服务器: 10 (默认)
   - 高性能服务器: 20-50

2. **实现自动定时处理**
   在SchedulerService中添加定时任务：
   ```go
   s.scheduler.Every(5).Minutes().Do(func() {
       ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)
       defer cancel()
       s.aiAnalysisService.ProcessPendingAnalyses(ctx)
   })
   ```

3. **添加任务队列优先级**
   支持按优先级处理任务

## 测试

### 运行测试脚本

```bash
# 生成测试token
TOKEN=$(go run generate_token.go -openid "om8hB12mqHOp1BiTf3KZ_ew8eWH4" | grep "Bearer")

# 运行测试脚本
bash test_background_jobs.sh "$TOKEN"
```

### 手动测试

详见本文档上方的"使用示例"部分。

## 常见问题

### Q: 为什么任务创建后还是pending状态？
A: 这是正常的。CreateAnalysis返回pending状态，需要调用ProcessPendingAnalyses处理。

### Q: 批量分析需要多长时间？
A: 取决于数据量和AI模型响应时间，通常10-100秒。

### Q: 能否取消正在处理的任务？
A: 当前不支持。任务失败后可重新创建。

### Q: 如何监控任务处理进度？
A: 通过GET /v1/ai-analysis/{id}查询任务状态。

## 监控和调试

### 查看待处理任务数

```bash
curl http://localhost:8080/v1/ai-analysis/baby/1/history \
  -H "Authorization: Bearer $TOKEN" | jq '.data | map(select(.status=="pending")) | length'
```

### 查看处理日志

```bash
tail -f logs/app.log | grep -i "ai分析\|处理\|分析任务"
```

## 总结

✅ **完整实现**：三种处理方式都已实现并可用
✅ **路由注册**：后台任务端点已注册在路由中
✅ **依赖注入**：所有依赖都通过Wire配置
✅ **数据持久化**：任务状态和结果持久化到数据库
✅ **错误处理**：完整的错误处理和日志记录

**推荐使用**：
- 需要立即返回结果 → **批量分析**
- 需要批量处理任务 → **创建 + 处理**
- 需要灵活控制 → **手动触发处理**

---

**最后更新**: 2025-11-12
**状态**: ✅ 完全可用
