# AI分析API文档

## 概述
本文档描述了宝宝喂养记录系统中AI分析功能的API接口。系统使用Eino AI框架集成大模型能力，提供智能化的宝宝数据分析服务。

## 功能特性
- 喂养数据分析
- 睡眠数据分析
- 成长数据分析
- 健康综合评估
- 行为模式识别
- 每日智能建议
- 批量数据分析

## API端点

### 1. 创建AI分析任务
**POST** `/v1/ai-analysis`

创建指定类型的AI分析任务。

**请求参数：**
```json
{
  "baby_id": 123,
  "analysis_type": "feeding", // feeding, sleep, growth, health, behavior
  "start_date": "2024-01-01T00:00:00Z",
  "end_date": "2024-01-07T23:59:59Z",
  "options": {}
}
```

**响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "analysis_id": "456",
    "status": "pending",
    "created_at": "2024-01-08T10:00:00Z",
    "message": "AI分析任务已创建，正在处理中..."
  }
}
```

### 2. 获取AI分析结果
**GET** `/v1/ai-analysis/{id}`

获取指定分析ID的分析结果。

**路径参数：**
- `id`: 分析任务ID

**响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "analysis_id": "456",
    "status": "completed",
    "result": {
      "score": 85,
      "insights": [
        {
          "type": "feeding",
          "title": "喂养规律良好",
          "description": "宝宝的喂养时间较为规律，建议继续保持",
          "priority": "medium",
          "category": "规律性"
        }
      ],
      "alerts": [],
      "patterns": [
        {
          "pattern_type": "regular_feeding",
          "description": "每3-4小时喂养一次",
          "confidence": 0.9,
          "frequency": "daily"
        }
      ],
      "predictions": []
    },
    "created_at": "2024-01-08T10:00:00Z"
  }
}
```

### 3. 获取最新AI分析结果
**GET** `/v1/ai-analysis/baby/{babyId}/latest`

获取指定宝宝和类型的最新AI分析结果。

**查询参数：**
- `baby_id`: 宝宝ID
- `analysis_type`: 分析类型 (feeding, sleep, growth, health, behavior)

### 4. 获取AI分析历史统计
**GET** `/v1/ai-analysis/baby/{babyId}/history`

获取指定宝宝的AI分析历史统计数据。

**响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "total_analyses": 15,
    "completed_analyses": 12,
    "failed_analyses": 3,
    "average_score": 82.5,
    "analysis_type_counts": {
      "feeding": 5,
      "sleep": 4,
      "growth": 3,
      "health": 3
    },
    "recent_analyses": [
      // 最近分析列表
    ]
  }
}
```

### 5. 批量AI分析
**POST** `/v1/ai-analysis/batch`

对指定时间范围内的宝宝数据进行批量AI分析。

**查询参数：**
- `baby_id`: 宝宝ID
- `start_date`: 开始日期 (YYYY-MM-DD)
- `end_date`: 结束日期 (YYYY-MM-DD)

### 6. 获取每日建议
**GET** `/v1/ai-analysis/daily-tips/{babyId}`

获取指定宝宝的每日建议。

**查询参数：**
- `date`: 日期 (YYYY-MM-DD)，可选，默认为当天

**响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "tips": [
      {
        "id": "tip_1",
        "icon": "🍼",
        "title": "喂养时间建议",
        "description": "建议在上午9-10点之间进行喂养，此时宝宝消化吸收效果最佳",
        "type": "feeding",
        "priority": "high",
        "action_url": "/pages/record/feeding/index"
      }
    ],
    "generated_at": "2024-01-08T09:00:00Z",
    "expired_at": "2024-01-09T09:00:00Z"
  }
}
```

### 7. 生成每日建议
**POST** `/v1/ai-analysis/daily-tips/{babyId}/generate`

为指定宝宝生成新的每日建议。

## 分析类型

### 1. 喂养分析 (feeding)
- **数据基础**: 母乳喂养、奶瓶喂养、辅食喂养记录
- **分析维度**: 喂养频率、喂养量、喂养时间规律性、营养摄入均衡性
- **输出结果**: 喂养规律评分、营养建议、喂养时间优化建议

### 2. 睡眠分析 (sleep)
- **数据基础**: 睡眠记录（入睡时间、睡眠时长、睡眠质量）
- **分析维度**: 睡眠时长充足性、睡眠规律性、夜间觉醒频率
- **输出结果**: 睡眠质量评分、睡眠改善建议、作息调整建议

### 3. 成长分析 (growth)
- **数据基础**: 身高、体重、头围等成长指标记录
- **分析维度**: 生长速度、发育里程碑达成情况、与WHO标准对比
- **输出结果**: 生长发育评估、成长趋势预测、营养建议

### 4. 健康分析 (health)
- **数据基础**: 综合喂养、睡眠、排泄、体温等多维度数据
- **分析维度**: 整体健康状况、潜在风险识别、健康趋势
- **输出结果**: 综合健康评分、健康预警、护理建议

### 5. 行为分析 (behavior)
- **数据基础**: 喂养、睡眠、活动等行为模式数据
- **分析维度**: 行为规律性、异常行为识别、习惯养成评估
- **输出结果**: 行为模式分析、习惯养成建议、异常行为提醒

## 状态码说明

- `pending`: 分析任务已创建，等待处理
- `analyzing`: 正在进行AI分析
- `completed`: 分析已完成，结果可用
- `failed`: 分析失败

## 错误处理

系统使用统一的错误响应格式：

```json
{
  "code": 1001,
  "message": "参数错误",
  "timestamp": 1641600000
}
```

常见错误码：
- `1001`: 参数错误
- `1002`: 未授权
- `1003`: 资源不存在
- `1005`: 权限不足
- `2001`: 服务器内部错误
- `3004`: 宝宝不存在

## AI模型配置

系统支持多种AI模型提供商：

- **OpenAI**: GPT-4, GPT-3.5等模型
- **Claude**: Claude-3系列模型
- **ERNIE**: 百度文心一言模型
- **Mock**: 模拟模型（用于开发和测试）

配置示例：
```yaml
ai:
  provider: "openai" # 支持的提供商: openai, claude, ernie, mock
  openai:
    api_key: "your-openai-api-key"
    model: "gpt-4"
    max_tokens: 2000
    temperature: 0.7
```

## 性能优化

- **缓存机制**: 分析结果缓存1小时，避免重复计算
- **批量处理**: 支持批量分析，提高处理效率
- **异步处理**: 分析任务异步执行，避免阻塞用户操作
- **智能提示**: 每日建议24小时有效期，避免频繁生成

## 使用建议

1. **合理设置分析时间范围**: 建议分析7-30天的数据以获得更准确的结果
2. **定期进行分析**: 建议每周进行一次综合分析
3. **关注异常提醒**: 及时查看AI分析中的预警信息
4. **结合实际情况**: AI建议仅供参考，请结合宝宝实际情况和医生建议