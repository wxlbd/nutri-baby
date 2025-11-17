# APIfox 使用指南 - AI分析API

## 第一步：获取有效Token

### 方式1: 使用Token生成工具

```bash
cd nutri-baby-server
go run generate_token.go \
  -openid "om8hB12mqHOp1BiTf3KZ_ew8eWH4" \
  -secret "your-secret-key-change-in-production" \
  -expire 72
```

输出的token示例:
```
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJvbThoQjEybXFIT3AxQmlUZjNLWl9ldzhlV0g0IiwiZXhwIjoxNzYzMjA1NDkwLCJpYXQiOjE3NjI5NDYyOTB9.w55lGHp6znR4mK1Q40ypk48_Evn9MqiypXI2BrA4Z3A
```

### 方式2: 使用微信登录获取Token

通过 `POST /v1/auth/wechat-login` 端点

## 第二步：在APIfox中配置认证

### 方式A: 使用Auth标签（推荐）

1. 在APIfox中打开任何请求
2. 找到 **Auth** 标签
3. 选择 **Type** 下拉列表 → **Bearer Token**
4. 在 **Token** 输入框中粘贴token（仅token部分，不包括"Bearer"前缀）

![APIfox Auth Configuration](./docs/apifox-auth.png)

### 方式B: 使用Headers（手动）

1. 切换到 **Headers** 标签
2. 添加新Header:
   - **Key**: `Authorization`
   - **Value**: `Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...`

## 第三步：创建测试请求

### 1. 创建AI分析任务

**请求配置**:
- **方法**: POST
- **URL**: `http://localhost:8080/v1/ai-analysis`
- **Auth**: Bearer Token (或 Authorization header)

**请求体** (JSON):
```json
{
  "baby_id": 1,
  "analysis_type": "feeding",
  "start_date": "2025-11-01",
  "end_date": "2025-11-08"
}
```

**预期响应** (200 OK):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "analysis_id": "1",
    "status": "pending",
    "created_at": "2025-11-12T19:21:00Z",
    "message": "AI分析任务已创建，正在处理中..."
  },
  "timestamp": 1762946460
}
```

### 2. 获取分析结果

**请求配置**:
- **方法**: GET
- **URL**: `http://localhost:8080/v1/ai-analysis/1`
- **Auth**: Bearer Token

**预期响应** (200 OK):
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "analysis_id": "1",
    "status": "completed",
    "result": {
      "score": 85,
      "insights": [
        {
          "type": "feeding",
          "title": "喂养规律良好",
          "description": "宝宝的喂养时间较为规律",
          "priority": "medium"
        }
      ],
      "alerts": [],
      "patterns": []
    },
    "created_at": "2025-11-12T19:21:00Z"
  },
  "timestamp": 1762946460
}
```

### 3. 批量分析

**请求配置**:
- **方法**: POST
- **URL**: `http://localhost:8080/v1/ai-analysis/batch?baby_id=1&start_date=2025-11-01&end_date=2025-11-08`
- **Auth**: Bearer Token

### 4. 获取每日建议

**请求配置**:
- **方法**: GET
- **URL**: `http://localhost:8080/v1/ai-analysis/daily-tips/1?date=2025-11-12`
- **Auth**: Bearer Token

## 测试场景

### 场景1: 验证认证

**错误情况 1a: 没有Token**
```bash
curl -X POST http://localhost:8080/v1/ai-analysis \
  -H "Content-Type: application/json" \
  -d '{"baby_id":1,"analysis_type":"feeding","start_date":"2025-11-01","end_date":"2025-11-08"}'
```

预期响应 (401):
```json
{
  "code": 1002,
  "message": "未授权",
  "timestamp": 1762946339
}
```

**错误情况 1b: Token格式错误**
```
Authorization: eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9... // 缺少"Bearer "前缀
```

预期响应 (401):
```json
{
  "code": 1002,
  "message": "未授权",
  "timestamp": 1762946339
}
```

### 场景2: 验证日期格式

在APIfox中创建4个请求，使用相同的auth但不同的日期格式:

**请求2a: 简单日期**
```json
{
  "baby_id": 1,
  "analysis_type": "feeding",
  "start_date": "2025-11-01",
  "end_date": "2025-11-08"
}
```

**请求2b: 日期时间**
```json
{
  "baby_id": 1,
  "analysis_type": "feeding",
  "start_date": "2025-11-01 00:00:00",
  "end_date": "2025-11-08 23:59:59"
}
```

**请求2c: ISO 8601**
```json
{
  "baby_id": 1,
  "analysis_type": "feeding",
  "start_date": "2025-11-01T00:00:00",
  "end_date": "2025-11-08T23:59:59"
}
```

**请求2d: RFC3339**
```json
{
  "baby_id": 1,
  "analysis_type": "feeding",
  "start_date": "2025-11-01T00:00:00Z",
  "end_date": "2025-11-08T23:59:59Z"
}
```

所有请求应该返回相同的结果（除了因为宝宝不存在的1003错误）

### 场景3: 验证分析类型

测试所有支持的分析类型:

```
- feeding (喂养分析)
- sleep (睡眠分析)
- growth (成长分析)
- health (健康分析)
- behavior (行为分析)
```

## 故障排除

### 问题1: Token返回401

**原因可能**:
1. Token已过期
2. Token的Secret不匹配
3. Token格式不正确

**解决方法**:
```bash
# 重新生成token
go run generate_token.go -openid "om8hB12mqHOp1BiTf3KZ_ew8eWH4" \
  -secret "your-secret-key-change-in-production" \
  -expire 72
```

### 问题2: 返回1003错误（宝宝不存在）

**原因**: 数据库中不存在该ID的宝宝

**解决方法**:
1. 先通过其他接口创建宝宝
2. 使用正确的宝宝ID进行测试

### 问题3: 日期格式错误

如果收到"日期格式错误"，检查:
- 日期不为空
- 日期格式在支持列表中
- 开始日期 < 结束日期

## APIfox快捷键

- `Ctrl+Shift+L` - 打开/关闭左侧面板
- `Ctrl+Shift+R` - 打开/关闭右侧面板
- `Ctrl+Enter` - 发送请求
- `Ctrl+S` - 保存请求

## 导出和分享

### 导出为cURL

1. 选择任何请求
2. 右键 → **生成代码** → **cURL**
3. 复制生成的命令

### 导出为Postman

1. **File** → **Export**
2. 选择 **Postman Collection**
3. 保存JSON文件

## 常用API端点速查表

| 操作 | 方法 | 端点 | 认证 |
|------|------|------|------|
| 创建分析 | POST | `/v1/ai-analysis` | ✅ |
| 获取结果 | GET | `/v1/ai-analysis/{id}` | ✅ |
| 批量分析 | POST | `/v1/ai-analysis/batch` | ✅ |
| 获取建议 | GET | `/v1/ai-analysis/daily-tips/{babyId}` | ✅ |
| 生成建议 | POST | `/v1/ai-analysis/daily-tips/{babyId}/generate` | ✅ |
| 获取统计 | GET | `/v1/ai-analysis/baby/{babyId}/history` | ✅ |
| 最新分析 | GET | `/v1/ai-analysis/baby/{babyId}/latest` | ✅ |

## 有用的链接

- 📖 [API_ANALYSIS_QUICK_START.md](./AI_ANALYSIS_QUICK_START.md) - 快速开始
- 📖 [DATE_FORMAT_GUIDE.md](./DATE_FORMAT_GUIDE.md) - 日期格式详解
- 📖 [AI_ANALYSIS_API.md](./AI_ANALYSIS_API.md) - 完整API文档
- 📖 [JWT_AUTH_FIX_REPORT.md](./JWT_AUTH_FIX_REPORT.md) - 认证问题报告
