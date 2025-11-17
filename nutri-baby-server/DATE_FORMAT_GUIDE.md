# AI分析API - 日期格式支持指南

## 问题背景

在使用AI分析API创建任务时，可能会遇到日期格式错误：
```
parsing time "2025-11-08" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"
```

这是因为后端日期格式配置不正确导致的。我们已经修复了这个问题。

## 支持的日期格式

系统现在支持以下多种日期格式，无需转换即可直接使用：

### 1. **日期格式（推荐）**
```
格式: YYYY-MM-DD
示例: 2025-11-08
用途: 最简洁，适合大多数场景
```

### 2. **日期时间格式**
```
格式: YYYY-MM-DD HH:MM:SS
示例: 2025-11-08 10:30:00
用途: 需要精确到小时/分钟/秒的情况
```

### 3. **ISO 8601 格式（无时区）**
```
格式: YYYY-MM-DDTHH:MM:SS
示例: 2025-11-08T10:30:00
用途: ISO标准格式，与Web APIs兼容
```

### 4. **RFC3339 格式（带时区）**
```
格式: YYYY-MM-DDTHH:MM:SSZ
示例: 2025-11-08T10:30:00Z
用途: 包含时区信息，适合国际应用
```

### 5. **RFC3339 完整格式**
```
格式: YYYY-MM-DDTHH:MM:SS±HH:MM
示例: 2025-11-08T10:30:00+08:00
用途: 明确指定时区偏移
```

## API调用示例

### 示例1：使用简单日期格式

```bash
curl -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "feeding",
    "start_date": "2025-11-01",
    "end_date": "2025-11-08"
  }'
```

### 示例2：使用ISO 8601格式

```bash
curl -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "sleep",
    "start_date": "2025-11-01T00:00:00",
    "end_date": "2025-11-08T23:59:59"
  }'
```

### 示例3：使用RFC3339格式

```bash
curl -X POST http://localhost:8080/v1/ai-analysis \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "baby_id": 1,
    "analysis_type": "growth",
    "start_date": "2025-11-01T00:00:00Z",
    "end_date": "2025-11-08T23:59:59Z"
  }'
```

## 前端代码示例

### TypeScript/Vue.js

```typescript
// src/api/ai.ts

export interface CreateAnalysisRequest {
  baby_id: number
  analysis_type: 'feeding' | 'sleep' | 'growth' | 'health' | 'behavior'
  start_date: string // 支持多种格式
  end_date: string
  options?: Record<string, any>
}

// 创建分析任务 - 使用简单日期格式
export async function createAnalysis(
  babyId: number,
  analysisType: string,
  startDate: Date,
  endDate: Date
) {
  // 方式1：使用 YYYY-MM-DD 格式（推荐）
  const formatDate = (date: Date): string => {
    return date.toISOString().split('T')[0]
  }

  return request.post('/v1/ai-analysis', {
    baby_id: babyId,
    analysis_type: analysisType,
    start_date: formatDate(startDate),
    end_date: formatDate(endDate)
  })
}

// 使用示例
const handleCreateAnalysis = async () => {
  const startDate = new Date('2025-11-01')
  const endDate = new Date('2025-11-08')

  const result = await createAnalysis(1, 'feeding', startDate, endDate)
  console.log('分析任务已创建:', result)
}
```

### 使用日期选择器

```typescript
// 如果使用日期选择器组件
import { Date } from '@/utils/date'

export async function createAnalysisFromPicker(
  babyId: number,
  analysisType: string,
  startDate: string,  // 从日期选择器直接获取，格式通常是 YYYY-MM-DD
  endDate: string
) {
  // 无需任何格式转换，直接发送
  return request.post('/v1/ai-analysis', {
    baby_id: babyId,
    analysis_type: analysisType,
    start_date: startDate,
    end_date: endDate
  })
}
```

## 常见错误及解决方案

### 错误1：时间戳格式
❌ **错误做法**
```json
{
  "start_date": 1730419200,
  "end_date": 1731024000
}
```

✅ **正确做法**
```json
{
  "start_date": "2025-11-01",
  "end_date": "2025-11-08"
}
```

### 错误2：不支持的日期格式
❌ **错误做法**
```json
{
  "start_date": "11/01/2025",
  "end_date": "11/08/2025"
}
```

✅ **正确做法**
```json
{
  "start_date": "2025-11-01",
  "end_date": "2025-11-08"
}
```

### 错误3：时区不一致
❌ **错误做法**
```json
{
  "start_date": "2025-11-01T00:00:00+08:00",
  "end_date": "2025-11-08T23:59:59-05:00"
}
```

✅ **正确做法**
```json
{
  "start_date": "2025-11-01",
  "end_date": "2025-11-08"
}
```

## 后端实现细节

CustomTime 类型自动处理日期解析，尝试顺序如下：

1. RFC3339 完整格式：`2006-01-02T15:04:05Z07:00`
2. RFC3339 Z格式：`2006-01-02T15:04:05Z`
3. ISO 8601：`2006-01-02T15:04:05`
4. 日期时间：`2006-01-02 15:04:05`
5. 日期（默认）：`2006-01-02`

系统会依次尝试每种格式，直到解析成功。

## 最佳实践

1. **前端上传**：使用最简单的格式 `YYYY-MM-DD`
   ```typescript
   const dateStr = date.toISOString().split('T')[0]
   ```

2. **避免时区混淆**：使用UTC日期而不是本地日期
   ```typescript
   // ✅ 推荐
   const utcDate = new Date().toISOString().split('T')[0]

   // ❌ 避免
   const localDate = new Date().toLocaleDateString('en-CA')
   ```

3. **时间戳到日期**：使用标准库函数
   ```typescript
   // ✅ 推荐
   const date = new Date(timestamp * 1000).toISOString().split('T')[0]

   // ❌ 避免自己拼接
   const dateStr = `${year}-${month}-${day}`
   ```

## 故障排除

如果仍然收到日期格式错误：

1. **检查JSON格式**
   ```bash
   # 验证JSON有效性
   echo '{"start_date": "2025-11-01"}' | jq '.'
   ```

2. **检查日期值**
   - 确保日期字符串不为空
   - 确保日期顺序正确（开始日期 < 结束日期）

3. **使用测试脚本**
   ```bash
   bash test_ai_analysis.sh
   ```

4. **查看后端日志**
   ```bash
   tail -f logs/app.log | grep -i "date\|time\|parse"
   ```
