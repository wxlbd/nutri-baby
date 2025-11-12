# 自动处理验证指南

## 🎯 快速验证自动处理是否已启用

### 方法1：检查日志（推荐）

启动服务器并查看日志：

```bash
# 启动服务器
./nutri-baby-server

# 预期日志输出（应在启动时看到）：
# INFO: AI分析自动处理任务已启用 (每5分钟一次)
# INFO: Scheduler service started with auto-processing enabled
```

### 方法2：代码检查

查看修改后的SchedulerService：

```bash
grep -A 5 "processAIAnalysisTasks" nutri-baby-server/internal/application/service/scheduler_service.go
```

预期输出：显示定时任务注册和处理方法

### 方法3：完整验证流程

```bash
# 1. 生成测试token
TOKEN=$(go run generate_token.go -openid "om8hB12mqHOp1BiTf3KZ_ew8eWH4" | tail -1)

# 2. 创建分析任务
RESPONSE=$(curl -s -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "feeding",
    "start_date": "2025-11-01",
    "end_date": "2025-11-08"
  }')

ANALYSIS_ID=$(echo "$RESPONSE" | jq -r '.data.analysis_id')
echo "Created analysis: $ANALYSIS_ID"

# 3. 检查初始状态（应该是pending）
curl -s -X GET "http://localhost:8080/v1/ai-analysis/$ANALYSIS_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.data.status'
# 输出: "pending"

# 4. 等待5分钟或手动触发处理
# 选项A: 等待自动处理（5分钟）
echo "等待定时任务处理..."
sleep 300  # 5分钟

# 选项B: 手动触发处理（立即）
curl -X POST http://localhost:8080/v1/jobs/process-pending-analyses \
  -H "Authorization: Bearer $TOKEN"

# 5. 检查处理后状态（应该是completed）
curl -s -X GET "http://localhost:8080/v1/ai-analysis/$ANALYSIS_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.data.status'
# 输出: "completed"
```

## 📊 三种处理方式对比

### 完整对比表

| 特性 | 创建+自动处理 | 批量分析 | 手动触发处理 |
|------|-------------|--------|-----------|
| **处理时机** | 5分钟后 | 立即 | 按需 |
| **用户操作** | 创建后无需操作 | 创建即完成 | 需手动触发 |
| **适用场景** | 后台处理 | 需要结果 | 调试/紧急 |
| **响应时间** | 延迟 | 快速 | 快速 |
| **并发处理** | ✅ 支持 | ✅ 支持 | ✅ 支持 |

## 🔍 监控自动处理

### 查看当前待处理任务

```bash
# 使用API查询
curl -s -X GET http://localhost:8080/v1/ai-analysis/baby/1/history \
  -H "Authorization: Bearer $TOKEN" | \
  jq '.data.analyses[] | select(.status=="pending")'

# 直接查询数据库
psql -U postgres -d nutri_baby -c "
SELECT id, baby_id, analysis_type, status, created_at
FROM ai_analyses
WHERE status IN ('pending', 'analyzing')
ORDER BY created_at DESC
LIMIT 10;"
```

### 监控处理日志

```bash
# 查看最近的处理日志
tail -f logs/app.log | grep -i "自动处理\|AI分析\|pending"

# 预期看到的日志：
# INFO: 自动处理待分析AI任务成功
# INFO: AI分析任务完成
# ERROR: 自动处理待分析AI任务失败 (如果有错误)
```

## ⚙️ 常见配置修改

### 1. 改变处理频率

**文件**: `internal/application/service/scheduler_service.go:60`

```go
// 原始: 每5分钟
_, err := s.scheduler.Every(5).Minutes().Do(s.processAIAnalysisTasks)

// 改为每3分钟（更快）
_, err := s.scheduler.Every(3).Minutes().Do(s.processAIAnalysisTasks)

// 改为每10分钟（更慢）
_, err := s.scheduler.Every(10).Minutes().Do(s.processAIAnalysisTasks)
```

修改后重新编译：
```bash
go build -o nutri-baby-server
```

### 2. 改变批处理大小

**文件**: `internal/application/service/ai_analysis_service.go:499`

```go
// 原始: 最多10个
pendingAnalyses, err := s.aiAnalysisRepo.GetPendingAnalyses(ctx, 10)

// 改为20个（处理更多）
pendingAnalyses, err := s.aiAnalysisRepo.GetPendingAnalyses(ctx, 20)
```

### 3. 改变处理超时时间

**文件**: `internal/application/service/scheduler_service.go:79`

```go
// 原始: 4分钟超时
ctx, cancel := context.WithTimeout(context.Background(), 4*time.Minute)

// 改为6分钟（给复杂分析更多时间）
ctx, cancel := context.WithTimeout(context.Background(), 6*time.Minute)
```

## 🧪 测试场景

### 场景1: 验证基础自动处理

```bash
# 1. 启动服务器，查看日志
./nutri-baby-server

# 2. 在日志中应看到:
# INFO: AI分析自动处理任务已启用 (每5分钟一次)

# ✅ 通过
```

### 场景2: 创建任务等待自动处理

```bash
# 1. 创建分析任务
TOKEN="..."
ANALYSIS_ID=$(curl -s -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "feeding",
    "start_date": "2025-11-01",
    "end_date": "2025-11-08"
  }' | jq -r '.data.analysis_id')

echo "Created: $ANALYSIS_ID"

# 2. 检查状态是pending
curl -s -X GET "http://localhost:8080/v1/ai-analysis/$ANALYSIS_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.data.status'
# 输出: "pending"

# 3. 等待5分钟（或手动触发）
curl -X POST http://localhost:8080/v1/jobs/process-pending-analyses \
  -H "Authorization: Bearer $TOKEN"

# 4. 立即检查状态应变为completed
curl -s -X GET "http://localhost:8080/v1/ai-analysis/$ANALYSIS_ID" \
  -H "Authorization: Bearer $TOKEN" | jq '.data.status'
# 输出: "completed"

# ✅ 通过
```

### 场景3: 并发处理验证

```bash
# 1. 创建多个任务
for i in {1..5}; do
  curl -s -X POST http://localhost:8080/v1/ai-analysis \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"baby_id\":1,\"analysis_type\":\"feeding\",\"start_date\":\"2025-11-0$i\",\"end_date\":\"2025-11-08\"}" \
    > /dev/null
done

# 2. 手动或自动处理
curl -X POST http://localhost:8080/v1/jobs/process-pending-analyses \
  -H "Authorization: Bearer $TOKEN"

# 3. 验证所有任务都被处理
curl -s -X GET http://localhost:8080/v1/ai-analysis/baby/1/history \
  -H "Authorization: Bearer $TOKEN" | \
  jq '.data.analyses | map(select(.status!="completed")) | length'
# 输出: 0（所有任务都完成）

# ✅ 通过
```

## 🐛 故障排除

### 问题1: 自动处理没有触发

**检查清单**:
1. 查看日志是否显示"AI分析自动处理任务已启用"
2. 检查是否有错误日志："添加AI分析定时任务失败"
3. 验证AIAnalysisService是否正确注入

```bash
# 查看日志
tail -f logs/app.log | grep "AI分析"

# 检查代码
grep -A 3 "processAIAnalysisTasks" scheduler_service.go
```

### 问题2: 处理失败

**检查清单**:
1. 检查数据库是否可以连接
2. 检查是否有待处理任务
3. 查看具体错误日志

```bash
# 查看错误日志
tail -f logs/app.log | grep -i error

# 检查数据库
psql -U postgres -d nutri_baby -c "
SELECT COUNT(*) FROM ai_analyses WHERE status='pending';"
```

### 问题3: 处理太慢

**优化建议**:
1. 减少处理间隔 (从5分钟改为3分钟)
2. 增加批处理大小 (从10改为20)
3. 检查AI模型响应时间

```bash
# 查看处理时间
tail -f logs/app.log | grep "自动处理"
# 查看每个分析的耗时
tail -f logs/app.log | grep "AI分析任务完成"
```

## 📚 相关代码位置

| 功能 | 文件 | 行号 |
|------|------|------|
| SchedulerService定义 | scheduler_service.go | 18-27 |
| Start方法 | scheduler_service.go | 54-68 |
| 处理方法 | scheduler_service.go | 76-88 |
| ProcessPendingAnalyses | ai_analysis_service.go | 497-516 |
| GetPendingAnalyses | ai_analysis_repository.go | - |

## ✅ 最后检查清单

- [x] 代码已修改并编译通过
- [x] Wire依赖已更新
- [x] Git已提交
- [x] 日志会显示启动信息
- [x] 定时任务已注册
- [x] 所有三种处理方式都可用

## 🎉 总结

**自动处理现在完全可用！**

项目启动时，系统会自动：
1. ✅ 启动定时任务调度器
2. ✅ 注册每5分钟执行一次的处理任务
3. ✅ 批量处理所有待处理的AI分析任务

用户无需任何手动操作，系统会自动处理已创建的分析任务！

---

**实现日期**: 2025-11-12
**状态**: ✅ 生产就绪
**提交**: bbd2a83
