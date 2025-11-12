## 日期格式快速参考

### 支持的格式

| 格式 | 示例 | 用途 |
|------|------|------|
| **YYYY-MM-DD** | `2025-11-08` | 👍 推荐，最简洁 |
| YYYY-MM-DD HH:MM:SS | `2025-11-08 10:30:00` | 需要精确时间 |
| YYYY-MM-DDTHH:MM:SS | `2025-11-08T10:30:00` | ISO 8601 |
| YYYY-MM-DDTHH:MM:SSZ | `2025-11-08T10:30:00Z` | RFC3339 |
| YYYY-MM-DDTHH:MM:SS±HH:MM | `2025-11-08T10:30:00+08:00` | 带时区 |

### 创建分析任务（最常用）

```bash
# 简单格式（推荐）
curl -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "feeding",
    "start_date": "2025-11-01",
    "end_date": "2025-11-08"
  }'
```

### 前端转换代码

```typescript
// 将Date对象转换为支持的格式
const formatDate = (date: Date): string => {
  return date.toISOString().split('T')[0]  // 返回 YYYY-MM-DD
}

// 使用
const startDate = formatDate(new Date('2025-11-01'))
const endDate = formatDate(new Date('2025-11-08'))
```

### 分析类型

```
feeding   - 喂养分析
sleep     - 睡眠分析
growth    - 成长分析
health    - 健康分析
behavior  - 行为分析
```

### 常见API端点

```bash
# 创建分析任务
POST /v1/ai-analysis

# 获取分析结果
GET /v1/ai-analysis/{analysis_id}

# 获取最新分析
GET /v1/ai-analysis/latest?baby_id=1&analysis_type=feeding

# 批量分析
POST /v1/ai-analysis/batch?baby_id=1&start_date=2025-11-01&end_date=2025-11-08

# 获取每日建议
GET /v1/ai-analysis/daily-tips?baby_id=1

# 生成每日建议
POST /v1/ai-analysis/daily-tips?baby_id=1

# 获取分析统计
GET /v1/ai-analysis/stats?baby_id=1
```

### 错误处理

```typescript
try {
  const result = await createAnalysis({
    baby_id: 1,
    analysis_type: 'feeding',
    start_date: '2025-11-01',
    end_date: '2025-11-08'
  })
  console.log('成功:', result)
} catch (error: any) {
  if (error.response?.status === 400) {
    console.error('日期格式错误:', error.response.data.message)
  } else if (error.response?.status === 401) {
    console.error('认证失败，请检查token')
  } else {
    console.error('服务器错误:', error)
  }
}
```
