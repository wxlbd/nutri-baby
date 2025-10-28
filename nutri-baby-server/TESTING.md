# 订阅消息测试指南

## ⚠️ 当前配置 - 测试模式

定时任务已配置为**每1分钟**执行一次,方便快速测试:

```
喂养提醒检查: 每1分钟执行 (生产环境: 每3分钟)
消息队列处理: 每1分钟执行 (生产环境: 每5分钟)
```

## 🧪 测试步骤

### 1. 启动后端服务

```bash
cd nutri-baby-server
./bin/server
```

**预期输出**:
```
INFO  Starting Nutri Baby Server...
INFO  Database connected successfully
INFO  Scheduler service started (TEST MODE: runs every 1 minute)
INFO  Server is running addr=:8080 mode=debug
```

### 2. 观察定时任务执行

服务启动后,每隔1分钟会看到以下日志:

```
INFO  Starting feeding reminder check...
INFO  Feeding reminder check completed (implementation pending: need to iterate through all babies)
DEBUG Processing message queue...
DEBUG No pending messages in queue
```

### 3. 测试前端授权上传

#### 步骤一: 打开小程序,触发授权

1. 在小程序中添加一条喂养记录(奶瓶喂养)
2. 首次添加时会弹出订阅消息授权引导
3. 点击"立即开启"
4. 在微信授权弹窗中点击"允许"

#### 步骤二: 查看前端日志

打开微信开发者工具控制台,应该看到:

```
[Subscribe] requestSubscribeMessage success: {...}
[Subscribe] Uploading auth records to backend: [...]
[Subscribe] Auth records uploaded successfully: {...}
```

#### 步骤三: 查看后端日志

后端服务控制台会输出:

```
INFO  Received POST /v1/subscribe/auth
INFO  Auth record saved successfully
```

### 4. 测试消息队列

#### 方式一: 通过 API 手动添加消息到队列

```bash
curl -X POST http://localhost:8080/v1/subscribe/queue \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "openid": "test_openid",
    "templateType": "vaccine_reminder",
    "data": {
      "babyName": "测试宝宝",
      "vaccineName": "卡介苗",
      "dueDate": "2025-10-25",
      "location": "社区医院",
      "doseNumber": "1"
    },
    "page": "pages/vaccine/vaccine",
    "scheduledTime": 0
  }'
```

#### 方式二: 创建喂养记录数据

由于当前的喂养提醒功能实现需要宝宝仓储的 `FindAll()` 方法支持,暂时处于待完善状态。

**日志输出**:
```
INFO  Starting feeding reminder check...
INFO  Feeding reminder check completed (implementation pending: need to iterate through all babies)
```

**完善提醒功能的建议**:
1. 在 BabyRepository 接口添加 `FindAll(ctx context.Context) ([]*entity.Baby, error)` 方法
2. 实现该方法以获取所有宝宝列表
3. 完善 CheckFeedingReminders() 函数中的遍历逻辑

### 5. 验证消息发送

对于手动添加的消息队列测试数据,查看后端日志,应该看到:

```
INFO  Processing message queue...
INFO  Processing message queue count=1
INFO  Message sent successfully messageId=1
```

### 6. 查询发送日志

```bash
curl -X GET "http://localhost:8080/v1/subscribe/logs?offset=0&limit=20" \
  -H "Authorization: Bearer YOUR_TOKEN"
```

**预期响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "logs": [
      {
        "id": 1,
        "templateType": "vaccine_reminder",
        "sendStatus": "success",
        "sendTime": 1729788000,
        "createdAt": 1729788000
      }
    ],
    "total": 1
  }
}
```

## 🐛 常见问题

### 1. 定时任务不执行

**问题**: 启动后没有看到定时任务日志

**检查**:
```bash
# 检查服务是否启动
ps aux | grep server

# 检查日志级别是否太高
# 确保 config.yaml 中 log.level = "debug"
```

### 2. 微信 access_token 获取失败

**问题**: 日志显示 "Failed to request access_token"

**原因**:
- AppID 或 AppSecret 配置错误
- 网络无法访问微信 API

**解决**: 检查 `config/config.yaml`:
```yaml
wechat:
  app_id: "wxXXXXXXXXXXXXXXXX"
  app_secret: "your_app_secret_here"
```

### 3. 数据库连接失败

**问题**: "Failed to connect database"

**解决**:
```bash
# 测试数据库连接
psql -h 101.200.47.93 -U postgres -d postgres

# 检查配置文件
cat config/config.yaml | grep -A 5 database
```

### 4. 前端上传失败

**问题**: 控制台显示 "Error uploading auth records"

**原因**:
- Token 未设置或已过期
- API 地址配置错误
- 后端服务未启动

**解决**:
```typescript
// 检查 .env.development 文件
VITE_API_BASE_URL=http://localhost:8080

// 检查 token
console.log(getStorage(StorageKeys.TOKEN))
```

## 📊 监控指标

### 关键日志

```bash
# 实时查看所有日志
tail -f logs/app.log

# 只看错误日志
grep ERROR logs/app.log

# 查看定时任务执行
grep "feeding reminder\|message queue" logs/app.log

# 查看消息发送
grep "Message sent successfully" logs/app.log
```

### 数据库查询

```sql
-- 查看订阅记录
SELECT * FROM subscribe_records ORDER BY created_at DESC LIMIT 10;

-- 查看发送日志
SELECT * FROM message_send_logs ORDER BY created_at DESC LIMIT 10;

-- 查看队列状态
SELECT status, COUNT(*) FROM message_send_queue GROUP BY status;

-- 查看喂养记录(用于测试提醒)
SELECT * FROM feeding_records ORDER BY time DESC LIMIT 10;
```

## 🔄 切换到生产模式

测试完成后,需要修改定时任务间隔:

### 1. 修改代码

编辑 `internal/application/service/scheduler_service.go`:

```go
func (s *SchedulerService) Start() {
	// 生产环境: 每3分钟检查喂养提醒
	s.cron.AddFunc("0 */3 * * * *", func() {
		s.logger.Info("Starting feeding reminder check...")
		if err := s.CheckFeedingReminders(); err != nil {
			s.logger.Error("Feeding reminder check failed", zap.Error(err))
		}
	})

	// 生产环境: 每5分钟处理消息队列
	s.cron.AddFunc("0 */5 * * * *", func() {
		s.logger.Info("Processing message queue...")
		if err := s.ProcessMessageQueue(); err != nil {
			s.logger.Error("Message queue processing failed", zap.Error(err))
		}
	})

	s.cron.Start()
	s.logger.Info("Scheduler service started (PRODUCTION MODE)")
}
```

### 2. 重新编译

```bash
go build -o bin/server ./cmd/server
```

### 3. 更新配置

```yaml
# config/config.yaml
server:
  mode: "release"  # 改为生产模式

log:
  level: "info"    # 生产环境不输出 debug 日志
```

## ✅ 测试检查清单

- [ ] 服务启动成功
- [ ] 定时任务每1分钟执行
- [ ] 前端授权记录成功上传
- [ ] 后端接收并保存授权记录
- [ ] 喂养提醒检查正常运行
- [ ] 消息队列处理正常运行
- [ ] 微信 access_token 获取成功
- [ ] 消息发送成功(或模拟成功)
- [ ] 发送日志正确记录
- [ ] 失败重试机制正常工作

全部测试通过后,即可切换到生产模式! 🚀
