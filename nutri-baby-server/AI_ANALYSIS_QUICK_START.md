# AI分析API快速开始指南

## 问题解决总结

### 认证问题 ✅ 已解决

**问题**: 使用Bearer Token调用AI分析API时收到401 Unauthorized错误

**根本原因**: `AIAnalysisHandler` 的权限检查方法 `checkPermission()` 尝试从上下文中读取 `user_id` 键，但认证中间件实际上设置的是 `openid` 键。

**解决方案**: 修改 `internal/interface/http/handler/ai_analysis_handler.go` 中的 `checkPermission()` 方法，使用正确的上下文键 `openid`。

**修改前**:
```go
func (h *AIAnalysisHandler) checkPermission(c *gin.Context, babyID int64) error {
    userID, exists := c.Get("user_id")  // ❌ 错误的键
    if !exists {
        return errors.ErrUnauthorized
    }
    // ...
}
```

**修改后**:
```go
func (h *AIAnalysisHandler) checkPermission(c *gin.Context, babyID int64) error {
    openid, exists := c.Get("openid")  // ✅ 正确的键
    if !exists {
        return errors.ErrUnauthorized
    }
    // ...
}
```

## 如何使用AI分析API

### 1. 生成JWT Token

使用项目根目录的 `generate_token.go` 工具生成有效的JWT token:

```bash
cd nutri-baby-server
go run generate_token.go -openid "om8hB12mqHOp1BiTf3KZ_ew8eWH4" -secret "your-secret-key-change-in-production" -expire 72
```

输出示例:
```
✅ Token生成成功:

Bearer Token:
eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJvbThoQjEybXFIT3AxQmlUZjNLWl9ldzhlV0g0IiwiZXhwIjoxNzYzMjA1NDkwLCJpYXQiOjE3NjI5NDYyOTB9.w55lGHp6znR4mK1Q40ypk48_Evn9MqiypXI2BrA4Z3A

Token信息:
  - OpenID: om8hB12mqHOp1BiTf3KZ_ew8eWH4
  - Secret: your-secret-key-change-in-production
  - 过期时间: 72小时
  - 生成时间: 2025-11-12 19:18:10
  - 过期时间: 2025-11-15 19:18:10

在APIfox中使用:
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9.eyJzdWIiOiJvbThoQjEybXFIT3AxQmlUZjNLWl9ldzhlV0g0IiwiZXhwIjoxNzYzMjA1NDkwLCJpYXQiOjE3NjI5NDYyOTB9.w55lGHp6znR4mK1Q40ypk48_Evn9MqiypXI2BrA4Z3A
```

### 2. 在APIfox中使用Token

1. 在APIfox中创建请求
2. 选择 **Auth** 标签
3. 选择 **Type** → **Bearer Token**
4. 在 **Token** 字段中粘贴上面生成的token（不需要`Bearer`前缀）

或者直接在Headers中添加:
```
Authorization: Bearer eyJhbGciOiJIUzI1NiIsInR5cCI6IkpXVCJ9...
```

### 3. 创建AI分析任务

**端点**: `POST /v1/ai-analysis`

**请求体** (支持多种日期格式):

```json
{
  "baby_id": 1,
  "analysis_type": "feeding",
  "start_date": "2025-11-01",
  "end_date": "2025-11-08"
}
```

**支持的日期格式**:
- ✅ 简单日期: `2025-11-01` (推荐)
- ✅ 日期时间: `2025-11-01 10:30:00`
- ✅ ISO 8601: `2025-11-01T10:30:00`
- ✅ RFC3339: `2025-11-01T10:30:00Z`
- ✅ RFC3339完整: `2025-11-01T10:30:00+08:00`

**支持的分析类型**:
- `feeding` - 喂养分析
- `sleep` - 睡眠分析
- `growth` - 成长分析
- `health` - 健康分析
- `behavior` - 行为分析

**成功响应示例** (HTTP 200):
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

**认证失败响应** (HTTP 401):
```json
{
  "code": 1002,
  "message": "未授权",
  "timestamp": 1762946339
}
```

如果收到此错误，请检查:
- ✅ Bearer token是否正确
- ✅ Token是否过期
- ✅ Token的Secret是否与config.yaml中的JWT.Secret匹配

### 4. 获取分析结果

**端点**: `GET /v1/ai-analysis/{id}`

```bash
curl -H "Authorization: Bearer YOUR_TOKEN" \
  http://localhost:8080/v1/ai-analysis/1
```

### 5. 批量分析

**端点**: `POST /v1/ai-analysis/batch`

**查询参数**:
```
baby_id=1&start_date=2025-11-01&end_date=2025-11-08
```

### 6. 获取每日建议

**端点**: `GET /v1/ai-analysis/daily-tips/{babyId}`

**查询参数**:
```
date=2025-11-12  // 可选，默认为当前日期
```

## 测试脚本

使用 `test_ai_analysis.sh` 快速测试所有端点:

```bash
bash test_ai_analysis.sh
```

## 重要提示

1. **密钥管理**: 确保 `config/config.yaml` 中的 JWT.Secret 与生成token时使用的密钥一致

2. **日期格式**: 所有日期格式都会自动转换为内部的UTC时间，确保在不同时区上的准确性

3. **权限验证**: API会检查当前用户是否有权访问指定的宝宝数据。暂时权限检查直接返回成功，需要实现完整的权限检查逻辑。

4. **异步处理**: AI分析任务是异步执行的，创建任务后会立即返回，实际分析结果可能需要几秒到几分钟

## 故障排除

### 问题1: 401 Unauthorized

**原因**: Token无效或不匹配
**解决**: 重新生成token，确保使用正确的openid和secret

### 问题2: 404 Baby Not Found (错误码1003)

**原因**: 数据库中不存在该ID的宝宝
**解决**: 先创建宝宝，再进行分析

### 问题3: Token过期

**错误信息**: `无效的令牌`
**解决**: 使用refresh token刷新或重新生成token

## 相关文档

- [DATE_FORMAT_GUIDE.md](DATE_FORMAT_GUIDE.md) - 日期格式完整指南
- [AI_ANALYSIS_API.md](AI_ANALYSIS_API.md) - 完整API文档
- [config/config.yaml](config/config.yaml) - 配置文件参考
