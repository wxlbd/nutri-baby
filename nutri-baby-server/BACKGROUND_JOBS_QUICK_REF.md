# AI分析API - 快速参考卡

## 核心端点速查表

### 1. 创建分析任务

```bash
POST /v1/ai-analysis
Authorization: Bearer <token>
Content-Type: application/json

{
  "baby_id": 1,
  "analysis_type": "feeding|sleep|growth|health|behavior",
  "start_date": "2025-11-01",
  "end_date": "2025-11-08"
}

# 响应: { analysis_id, status: "pending" }
```

### 2. 获取分析结果

```bash
GET /v1/ai-analysis/{id}
Authorization: Bearer <token>

# 响应: { analysis_id, status, result: {...} }
```

### 3. 批量分析（推荐）⭐

```bash
POST /v1/ai-analysis/batch?baby_id=1&start_date=2025-11-01&end_date=2025-11-08
Authorization: Bearer <token>

# 立即返回: { analyses: [{分析1}, {分析2}, {分析3}, {分析4}] }
```

### 4. 处理待处理任务

```bash
POST /v1/jobs/process-pending-analyses
Authorization: Bearer <token>

# 响应: { code: 0, message: "success" }
```

### 5. 获取最新分析

```bash
GET /v1/ai-analysis/baby/{babyId}/latest?analysis_type=feeding
Authorization: Bearer <token>

# 响应: { analysis_id, status, result }
```

### 6. 获取分析统计

```bash
GET /v1/ai-analysis/baby/{babyId}/history
Authorization: Bearer <token>

# 响应: { total, completed, failed, analyses: [...] }
```

### 7. 获取每日建议

```bash
GET /v1/ai-analysis/daily-tips/{babyId}?date=2025-11-12
Authorization: Bearer <token>

# 响应: { tips, generated_at, expired_at }
```

### 8. 生成每日建议

```bash
POST /v1/ai-analysis/daily-tips/{babyId}?date=2025-11-12
Authorization: Bearer <token>

# 响应: { tips, generated_at, expired_at }
```

## 任务状态

| 状态 | 说明 | 可查询 | 有结果 |
|------|------|--------|--------|
| pending | 等待处理 | ✅ | ❌ |
| analyzing | 分析中 | ✅ | ❌ |
| completed | 已完成 | ✅ | ✅ |
| failed | 失败 | ✅ | ❌ |

## 分析类型

- `feeding` - 喂养分析
- `sleep` - 睡眠分析
- `growth` - 成长分析
- `health` - 健康分析
- `behavior` - 行为分析

## 日期格式支持

✅ 简单日期: `2025-11-01`
✅ 日期时间: `2025-11-01 10:30:00`
✅ ISO 8601: `2025-11-01T10:30:00`
✅ RFC3339: `2025-11-01T10:30:00Z`
✅ RFC3339+TZ: `2025-11-01T10:30:00+08:00`

## 常见组合

### 获取完整分析（推荐）

```bash
# 一次调用，立即返回所有4种分析
curl -X POST \
  "http://localhost:8080/v1/ai-analysis/batch?baby_id=1&start_date=2025-11-01&end_date=2025-11-08" \
  -H "Authorization: Bearer $TOKEN"
```

### 轮询等待结果

```bash
# 创建任务
ANALYSIS_ID=$(curl -s -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"baby_id":1,"analysis_type":"feeding","start_date":"2025-11-01","end_date":"2025-11-08"}' \
  | jq -r '.data.analysis_id')

# 轮询检查结果（每2秒检查一次，最多30次）
for i in {1..30}; do
  STATUS=$(curl -s -X GET "http://localhost:8080/v1/ai-analysis/$ANALYSIS_ID" \
    -H "Authorization: Bearer $TOKEN" \
    | jq -r '.data.status')

  if [ "$STATUS" = "completed" ]; then
    echo "✓ 分析完成"
    curl -s -X GET "http://localhost:8080/v1/ai-analysis/$ANALYSIS_ID" \
      -H "Authorization: Bearer $TOKEN" | jq '.data.result'
    break
  fi

  echo "⏳ 等待中... 状态: $STATUS"
  sleep 2
done
```

### 批量创建后处理

```bash
# 创建多个分析任务
for i in {1..5}; do
  curl -s -X POST http://localhost:8080/v1/ai-analysis \
    -H "Authorization: Bearer $TOKEN" \
    -H "Content-Type: application/json" \
    -d "{\"baby_id\":1,\"analysis_type\":\"feeding\",\"start_date\":\"2025-11-0$i\",\"end_date\":\"2025-11-08\"}" \
    > /dev/null
done

# 统一处理所有任务
curl -X POST http://localhost:8080/v1/jobs/process-pending-analyses \
  -H "Authorization: Bearer $TOKEN"
```

## 错误响应

### 401 未授权

```json
{
  "code": 1002,
  "message": "未授权"
}
```

原因: Token缺失、无效或过期

### 1003 资源不存在

```json
{
  "code": 1003,
  "message": "获取宝宝信息失败"
}
```

原因: baby_id 不存在

### 日期格式错误

```json
{
  "code": 1001,
  "message": "日期格式错误"
}
```

原因: 日期格式不支持

## 性能指标

| 操作 | 耗时 |
|------|------|
| 创建任务 | ~50ms |
| 批量分析 (4种) | ~200-500ms |
| 单个AI分析 | 1-10秒 |
| 处理10个任务 | 10-100秒 |

## 生成Token

```bash
go run generate_token.go \
  -openid "om8hB12mqHOp1BiTf3KZ_ew8eWH4" \
  -secret "your-secret-key" \
  -expire 72
```

## TypeScript使用示例

```typescript
const api = {
  // 批量分析（推荐）
  async batchAnalyze(babyId: number, startDate: string, endDate: string) {
    const response = await fetch(
      `/v1/ai-analysis/batch?baby_id=${babyId}&start_date=${startDate}&end_date=${endDate}`,
      {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json'
        }
      }
    );
    return response.json();
  },

  // 创建分析
  async createAnalysis(
    babyId: number,
    type: string,
    startDate: string,
    endDate: string
  ) {
    const response = await fetch('/v1/ai-analysis', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`,
        'Content-Type': 'application/json'
      },
      body: JSON.stringify({
        baby_id: babyId,
        analysis_type: type,
        start_date: startDate,
        end_date: endDate
      })
    });
    return response.json();
  },

  // 获取结果
  async getResult(analysisId: string) {
    const response = await fetch(`/v1/ai-analysis/${analysisId}`, {
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    return response.json();
  },

  // 处理待处理任务
  async processPending() {
    const response = await fetch('/v1/jobs/process-pending-analyses', {
      method: 'POST',
      headers: {
        'Authorization': `Bearer ${token}`
      }
    });
    return response.json();
  }
};
```

## 最佳实践

✅ **DO:**
- 批量分析时使用 POST /v1/ai-analysis/batch
- 为长时间操作实现超时机制
- 检查 status 字段而非假设完成
- 定期刷新 token

❌ **DON'T:**
- 频繁轮询同一个任务（推荐间隔 2+ 秒）
- 在前端直接调用处理待处理任务的端点
- 忽略错误响应码

## 相关文档

- [完整实现指南](./BACKGROUND_JOBS_COMPLETE.md)
- [实现总结](./BACKGROUND_JOBS_SUMMARY.md)
- [快速启动](./AI_ANALYSIS_QUICK_START.md)
- [API完整参考](./AI_ANALYSIS_API.md)

---

**最后更新**: 2025-11-12
