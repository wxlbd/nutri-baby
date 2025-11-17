# AI分析API日期格式修复总结

## 问题描述

创建AI分析任务时报错：
```
parsing time "2025-11-08" as "2006-01-02T15:04:05Z07:00": cannot parse "" as "T"
```

**原因**: 后端期望RFC3339格式的完整时间戳，但前端发送的是简单的日期格式。

## 解决方案

### 1. **实现自定义日期类型 (CustomTime)**

在 `internal/application/service/ai_analysis_service.go` 中：

```go
type CustomTime time.Time

func (ct *CustomTime) UnmarshalJSON(b []byte) error {
    // 支持5种日期格式的自动解析
    layouts := []string{
        "2006-01-02T15:04:05Z07:00",  // RFC3339
        "2006-01-02T15:04:05Z",       // RFC3339 without offset
        "2006-01-02T15:04:05",        // ISO 8601
        "2006-01-02 15:04:05",        // 日期时间
        "2006-01-02",                 // 日期（推荐）
    }
    // ... 尝试每个格式直到成功
}
```

### 2. **更新请求结构体**

```go
type CreateAnalysisRequest struct {
    BabyID       int64                  `json:"baby_id" binding:"required"`
    AnalysisType entity.AIAnalysisType  `json:"analysis_type" binding:"required"`
    StartDate    CustomTime             `json:"start_date" binding:"required"`  // 改为 CustomTime
    EndDate      CustomTime             `json:"end_date" binding:"required"`    // 改为 CustomTime
    Options      map[string]interface{} `json:"options,omitempty"`
}
```

### 3. **添加辅助方法**

- `Time()`: 转换为 `time.Time` 类型
- `Before()`: 日期比较
- `MarshalJSON()`: JSON序列化

## 支持的日期格式

现在API支持以下5种日期格式，无需格式转换：

✅ **推荐**
```json
"start_date": "2025-11-01",
"end_date": "2025-11-08"
```

✅ **ISO 8601**
```json
"start_date": "2025-11-01T00:00:00",
"end_date": "2025-11-08T23:59:59"
```

✅ **RFC3339**
```json
"start_date": "2025-11-01T00:00:00Z",
"end_date": "2025-11-08T23:59:59Z"
```

✅ **日期时间**
```json
"start_date": "2025-11-01 00:00:00",
"end_date": "2025-11-08 23:59:59"
```

✅ **带时区**
```json
"start_date": "2025-11-01T00:00:00+08:00",
"end_date": "2025-11-08T23:59:59+08:00"
```

## 前端使用示例

### TypeScript

```typescript
// 方式1: 使用简单日期格式（推荐）
const createAnalysis = async (babyId: number, startDate: Date, endDate: Date) => {
  const formatDate = (date: Date) => date.toISOString().split('T')[0]

  return request.post('/v1/ai-analysis', {
    baby_id: babyId,
    analysis_type: 'feeding',
    start_date: formatDate(startDate),
    end_date: formatDate(endDate)
  })
}

// 方式2: 直接使用日期字符串
const createAnalysis2 = async () => {
  return request.post('/v1/ai-analysis', {
    baby_id: 1,
    analysis_type: 'feeding',
    start_date: '2025-11-01',
    end_date: '2025-11-08'
  })
}
```

## 修改的文件

1. **internal/application/service/ai_analysis_service.go**
   - 添加 `CustomTime` 类型及其方法
   - 更新 `CreateAnalysisRequest` 结构体
   - 修改日期转换逻辑

2. **新增文档**
   - `DATE_FORMAT_GUIDE.md` - 详细使用指南
   - `DATE_FORMAT_QUICK_REF.md` - 快速参考
   - `test_ai_analysis.sh` - API测试脚本

## 测试

执行测试脚本验证各种日期格式：

```bash
bash test_ai_analysis.sh
```

## 编译状态

✅ 所有包编译成功
✅ 无类型错误
✅ 向后兼容

## 使用建议

1. **前端优先使用 `YYYY-MM-DD` 格式**
   - 最简洁
   - 减少序列化/反序列化开销
   - 避免时区混淆

2. **日期选择器集成**
   ```typescript
   const selectedDate = datePicker.value // "2025-11-08"
   request.post('/v1/ai-analysis', {
     start_date: selectedDate,
     end_date: selectedDate
   })
   ```

3. **错误处理**
   ```typescript
   try {
     await createAnalysis(...)
   } catch (error) {
     if (error.status === 400) {
       console.error('日期格式错误:', error.message)
     }
   }
   ```
