# 订阅消息发送流程说明

## 📋 完整流程概览

```
用户授权订阅 → 保存授权记录 → 触发业务事件 → 加入发送队列 → 定时任务处理 → 调用微信API → 记录发送日志
```

## 🔄 详细流程说明

### 1️⃣ 用户授权订阅

**前端调用：** `POST /api/v1/subscribe/auth`

**请求体：**
```json
{
  "records": [
    {
      "templateId": "微信模板ID",
      "templateType": "breast_feeding_reminder",
      "status": "accept"
    }
  ]
}
```

**后端处理：** `subscribe_service.go:SaveSubscribeAuth()`
- 保存授权记录到 `subscribe_records` 表
- 设置有效期（默认30天）
- 状态标记为 `active`

**数据库记录示例：**
```sql
INSERT INTO subscribe_records (
    openid,
    template_id,
    template_type,
    status,
    subscribe_time,
    expire_time
) VALUES (
    'om8hB12mqHOp1BiTf3KZ_ew8eWH4',
    'xxx',
    'breast_feeding_reminder',
    'active',
    NOW(),
    NOW() + INTERVAL '30 days'
);
```

---

### 2️⃣ 业务触发消息发送

#### 场景1：喂养提醒（定时任务触发）

**定时任务：** `scheduler_service.go:CheckFeedingReminders()`
- **触发频率：** 每1分钟检查一次（测试模式）
- **检查逻辑：**
  1. 获取所有宝宝
  2. 查询每个宝宝最近24小时的喂养记录
  3. 如果距离上次喂养 >= 3小时（测试模式：0.016小时=1分钟）
  4. 获取宝宝的协作者（家庭成员）
  5. 检查每个协作者的订阅状态
  6. 构造消息数据并加入队列

**消息数据格式：**
```json
{
  "lastTime": "14:30",
  "sinceTime": "3小时",
  "lastSide": "左侧",
  "reminderTip": "该喂奶啦，注意观察宝宝的饥饿信号"
}
```

#### 场景2：疫苗提醒（定时任务触发）

**定时任务：** `scheduler_service.go:CheckVaccineReminders()`
- **触发频率：** 需要配置 cron 表达式
- **检查逻辑：**
  1. 获取即将到期和已逾期的疫苗提醒
  2. 构造消息数据并加入队列

---

### 3️⃣ 消息加入发送队列

**服务方法：** `subscribe_service.go:QueueSubscribeMessage()`

**处理流程：**
1. **验证订阅状态：**
   - 查询 `subscribe_records` 表
   - 检查订阅是否存在
   - 检查订阅是否有效（active 且未过期）

2. **构造队列记录：**
```go
queue := &entity.MessageSendQueue{
    OpenID:        req.OpenID,
    TemplateID:    record.TemplateID,  // 从订阅记录获取
    TemplateType:  req.TemplateType,
    Data:          string(dataJSON),   // 序列化消息数据
    Page:          req.Page,           // 跳转页面
    ScheduledTime: time.Now(),         // 计划发送时间
    Status:        "pending",          // 待发送
}
```

3. **插入数据库：**
```sql
INSERT INTO message_send_queue (
    openid,
    template_id,
    template_type,
    data,
    page,
    scheduled_time,
    status
) VALUES (...);
```

**⚠️ 可能的失败原因：**
- ❌ 用户未订阅该消息类型
- ❌ 订阅已过期或已取消
- ❌ 数据库插入失败（检查约束、字段类型）

---

### 4️⃣ 定时任务处理消息队列

**定时任务：** `scheduler_service.go:ProcessMessageQueue()`
- **触发频率：** 每1分钟执行一次（测试模式）
- **处理逻辑：**

```go
// 1. 获取待发送消息（限制50条）
messages, err := s.subscribeRepo.GetPendingMessages(ctx, 50)

// 2. 过滤条件
for _, msg := range messages {
    // 检查是否到达发送时间
    if msg.ScheduledTime.After(time.Now()) {
        continue
    }

    // 检查重试次数
    if !msg.CanRetry() {
        // 标记为失败
        continue
    }

    // 3. 发送消息
    err := s.subscribeService.SendSubscribeMessage(ctx, sendReq)

    if err != nil {
        // 发送失败，增加重试次数
        msg.IncrementRetry()
    } else {
        // 发送成功，更新状态为 "sent"
    }
}
```

**数据库查询：**
```sql
SELECT * FROM message_send_queue
WHERE status = 'pending'
  AND scheduled_time <= NOW()
ORDER BY scheduled_time ASC
LIMIT 50;
```

---

### 5️⃣ 调用微信API发送消息

**服务方法：** `subscribe_service.go:SendSubscribeMessage()`

**调用链：**
```
SubscribeService.SendSubscribeMessage()
  ↓
WechatService.SendSubscribeMessage()
  ↓
获取 access_token (带缓存)
  ↓
格式化模板数据
  ↓
POST https://api.weixin.qq.com/cgi-bin/message/subscribe/send
```

**微信API请求体：**
```json
{
  "touser": "om8hB12mqHOp1BiTf3KZ_ew8eWH4",
  "template_id": "模板ID",
  "page": "pages/record/feeding/feeding",
  "miniprogram_state": "formal",
  "lang": "zh_CN",
  "data": {
    "lastTime": {"value": "14:30"},
    "sinceTime": {"value": "3小时"},
    "lastSide": {"value": "左侧"},
    "reminderTip": {"value": "该喂奶啦，注意观察宝宝的饥饿信号"}
  }
}
```

**微信API响应：**
```json
{
  "errcode": 0,
  "errmsg": "ok"
}
```

**常见错误码：**
- `40003`: 无效的 openid
- `43101`: 用户拒绝接收消息
- `47003`: 模板参数不正确
- `41030`: page路径不正确

---

### 6️⃣ 记录发送日志

**数据库表：** `message_send_logs`

**记录内容：**
```sql
INSERT INTO message_send_logs (
    openid,
    template_id,
    template_type,
    data,
    page,
    miniprogram_state,
    send_status,  -- 'success' 或 'failed'
    errcode,      -- 微信返回的错误码
    errmsg,       -- 错误信息
    send_time     -- 实际发送时间
) VALUES (...);
```

---

## 🔍 问题排查指南

### 问题1：订阅授权成功但队列添加失败

**可能原因：**
1. ❌ `TemplateID` 字段为空
2. ❌ 数据库唯一约束冲突（已修复）
3. ❌ 订阅记录未正确保存

**排查步骤：**
```sql
-- 1. 检查订阅记录是否存在
SELECT * FROM subscribe_records
WHERE openid = 'om8hB12mqHOp1BiTf3KZ_ew8eWH4'
  AND template_type = 'breast_feeding_reminder';

-- 2. 检查 template_id 是否为空
SELECT openid, template_id, template_type, status
FROM subscribe_records
WHERE template_id IS NULL OR template_id = '';

-- 3. 检查消息队列表
SELECT * FROM message_send_queue
WHERE openid = 'om8hB12mqHOp1BiTf3KZ_ew8eWH4';
```

---

### 问题2：消息未发送给用户

**可能原因：**
1. ❌ 定时任务未启动（已启动：`main.go:41`）
2. ❌ 消息队列为空
3. ❌ 发送条件不满足（未到发送时间、超过重试次数）
4. ❌ 微信API调用失败

**排查步骤：**
```sql
-- 1. 检查队列中是否有待发送消息
SELECT * FROM message_send_queue
WHERE status = 'pending'
  AND scheduled_time <= NOW()
ORDER BY created_at DESC;

-- 2. 检查发送日志
SELECT * FROM message_send_logs
WHERE openid = 'om8hB12mqHOp1BiTf3KZ_ew8eWH4'
ORDER BY created_at DESC
LIMIT 10;

-- 3. 查看失败的消息
SELECT * FROM message_send_logs
WHERE send_status = 'failed'
ORDER BY created_at DESC;
```

**检查日志：**
```bash
# 查看定时任务日志
tail -f logs/server.log | grep "Processing message queue"
tail -f logs/server.log | grep "Message sent successfully"
tail -f logs/server.log | grep "Failed to send message"
```

---

### 问题3：微信API调用失败

**排查步骤：**

1. **检查配置：**
```yaml
# config/config.yaml
wechat:
  app_id: "你的小程序AppID"
  app_secret: "你的小程序AppSecret"
```

2. **测试 access_token：**
```bash
curl "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=你的AppID&secret=你的AppSecret"
```

3. **检查模板ID：**
- 登录微信公众平台
- 进入"订阅消息" → "模板库"
- 确认模板ID正确
- 确认模板参数字段名称正确

4. **检查用户授权状态：**
- 用户必须主动触发订阅（点击订阅按钮）
- 订阅有效期30天，过期需重新授权

---

## 🧪 测试流程

### 1. 启动服务
```bash
cd /Users/wxl/GolandProjects/nutri-baby/nutri-baby-server
make run
```

### 2. 授权订阅（前端或API测试）
```bash
curl -X POST http://localhost:8080/api/v1/subscribe/auth \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "records": [
      {
        "templateId": "你的微信模板ID",
        "templateType": "breast_feeding_reminder",
        "status": "accept"
      }
    ]
  }'
```

### 3. 触发业务事件（添加喂养记录）
```bash
# 添加一条3小时前的喂养记录
curl -X POST http://localhost:8080/api/v1/babies/{babyId}/feeding-records \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "type": "breast",
    "time": 1729843200000,  # 3小时前的时间戳
    "detail": {
      "side": "left",
      "duration": 15
    }
  }'
```

### 4. 等待定时任务执行（1分钟）
```bash
# 查看日志
tail -f logs/server.log | grep "Feeding reminder"
```

### 5. 检查数据库
```sql
-- 检查消息队列
SELECT * FROM message_send_queue ORDER BY created_at DESC LIMIT 5;

-- 检查发送日志
SELECT * FROM message_send_logs ORDER BY created_at DESC LIMIT 5;
```

---

## 📊 数据流转示意图

```
┌─────────────────┐
│  前端用户授权    │
└────────┬────────┘
         │
         ▼
┌─────────────────────────┐
│  subscribe_records      │  ← 保存授权记录
│  (订阅记录表)            │
└─────────────────────────┘
         │
         │ (业务触发)
         ▼
┌─────────────────────────┐
│  定时任务检查            │  ← CheckFeedingReminders()
│  (距上次喂养>=3小时)     │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  message_send_queue     │  ← 加入发送队列
│  (消息发送队列表)        │     status='pending'
└────────┬────────────────┘
         │
         │ (每1分钟)
         ▼
┌─────────────────────────┐
│  定时任务处理            │  ← ProcessMessageQueue()
│  (从队列取消息发送)      │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  调用微信API             │  ← WechatService.SendSubscribeMessage()
│  (发送订阅消息)          │
└────────┬────────────────┘
         │
         ▼
┌─────────────────────────┐
│  message_send_logs      │  ← 记录发送结果
│  (消息发送日志表)        │     success/failed
└─────────────────────────┘
         │
         ▼
┌─────────────────────────┐
│  用户手机收到消息        │
└─────────────────────────┘
```

---

## 🎯 关键配置参数

### 定时任务频率（测试模式）
```go
// scheduler_service.go:63
s.cron.AddFunc("0 */1 * * * *", func() { // 每1分钟检查喂养提醒
    s.CheckFeedingReminders()
})

// scheduler_service.go:71
s.cron.AddFunc("0 */1 * * * *", func() { // 每1分钟处理消息队列
    s.ProcessMessageQueue()
})
```

### 喂养提醒阈值（测试模式）
```go
// scheduler_service.go:299
if hoursSinceLastFeeding >= 0.016 { // 测试模式：1分钟，生产环境改为3小时
    // 发送提醒
}
```

### 重试策略
```go
// entity/subscribe.go:69-70
MaxRetry:   3,  // 最大重试3次
RetryCount: 0,  // 当前重试次数
```

---

## ✅ 自检清单

- [ ] 数据库表已创建（`subscribe_records`, `message_send_queue`, `message_send_logs`）
- [ ] 定时任务已启动（查看日志：`Scheduler service started`）
- [ ] 微信配置正确（`app_id` 和 `app_secret`）
- [ ] 用户已授权订阅（检查 `subscribe_records` 表）
- [ ] 业务触发条件满足（例如：距上次喂养>=3小时）
- [ ] 消息已加入队列（检查 `message_send_queue` 表）
- [ ] 定时任务正在处理队列（查看日志）
- [ ] 微信API调用成功（检查 `message_send_logs` 表）

---

## 📞 获取帮助

如果以上步骤仍无法解决问题，请提供：
1. 完整的错误日志
2. 相关表的数据截图
3. 微信公众平台的模板配置截图
