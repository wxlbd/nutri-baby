# 订阅消息快速测试指南

## 🚀 快速测试步骤

### 第1步：检查数据库表是否存在

```bash
# 方法1：使用 psql 命令
psql -h localhost -U wxl -d nutri_baby -f diagnose_subscribe.sql

# 方法2：手动查询
psql -h localhost -U wxl -d nutri_baby -c "SELECT tablename FROM pg_tables WHERE schemaname = 'public' AND tablename LIKE '%subscribe%' OR tablename LIKE '%message%';"
```

**预期结果：**
- `subscribe_records` 表存在
- `message_send_queue` 表存在
- `message_send_logs` 表存在

---

### 第2步：启动后端服务

```bash
cd /Users/wxl/GolandProjects/nutri-baby/nutri-baby-server
make run
```

**检查日志输出：**
```
✓ Database connected successfully
✓ Scheduler service started (TEST MODE: runs every 1 minute)
✓ Server is running addr=:8080
```

---

### 第3步：模拟用户授权订阅

**前端调用或使用 curl：**

```bash
# 替换 YOUR_TOKEN 为真实的 JWT token
# 替换 YOUR_TEMPLATE_ID 为微信公众平台的模板ID

curl -X POST http://localhost:8080/api/v1/subscribe/auth \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "records": [
      {
        "templateId": "YOUR_TEMPLATE_ID",
        "templateType": "breast_feeding_reminder",
        "status": "accept"
      }
    ]
  }'
```

**预期响应：**
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "successCount": 1,
    "failedCount": 0
  },
  "timestamp": 1729856400
}
```

**验证数据库：**
```sql
SELECT * FROM subscribe_records
WHERE template_type = 'breast_feeding_reminder'
ORDER BY created_at DESC
LIMIT 1;
```

---

### 第4步：创建测试喂养记录（触发提醒）

**方法1：添加旧的喂养记录（推荐）**

```bash
# 计算3小时前的时间戳（毫秒）
# 当前时间戳 - 3 * 60 * 60 * 1000

THREE_HOURS_AGO=$(($(date +%s) * 1000 - 3 * 60 * 60 * 1000))

curl -X POST http://localhost:8080/api/v1/babies/{babyId}/feeding-records \
  -H "Authorization: Bearer YOUR_TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"type\": \"breast\",
    \"time\": $THREE_HOURS_AGO,
    \"detail\": {
      \"side\": \"left\",
      \"duration\": 15
    }
  }"
```

**方法2：直接插入测试数据到数据库**

```sql
-- 插入3小时前的喂养记录
INSERT INTO feeding_records (
    record_id,
    baby_id,
    type,
    time,
    detail,
    created_at,
    updated_at
) VALUES (
    gen_random_uuid()::text,
    '你的baby_id',
    'breast',
    EXTRACT(EPOCH FROM (NOW() - INTERVAL '3 hours')) * 1000,
    '{"side": "left", "duration": 15}'::jsonb,
    NOW(),
    NOW()
);
```

---

### 第5步：等待定时任务触发（最多1分钟）

**查看日志：**
```bash
tail -f logs/server.log | grep -E "Feeding reminder|Message queue|Message sent"
```

**预期日志输出：**
```
2025-10-25 00:30:00 INFO  Starting feeding reminder check...
2025-10-25 00:30:00 INFO  Baby needs feeding reminder babyId=xxx hoursSinceLastFeeding=3.0
2025-10-25 00:30:00 INFO  Feeding reminder queued babyId=xxx openid=om8hB12mqHOp1BiTf3KZ_ew8eWH4
2025-10-25 00:31:00 DEBUG Processing message queue...
2025-10-25 00:31:00 INFO  Message sent successfully messageId=1
```

---

### 第6步：检查数据库记录

**检查消息队列：**
```sql
SELECT
    id,
    openid,
    template_type,
    status,
    scheduled_time,
    retry_count,
    created_at
FROM message_send_queue
ORDER BY created_at DESC
LIMIT 5;
```

**检查发送日志：**
```sql
SELECT
    id,
    openid,
    template_type,
    send_status,
    errcode,
    errmsg,
    send_time,
    created_at
FROM message_send_logs
ORDER BY created_at DESC
LIMIT 5;
```

---

## ⚠️ 常见问题处理

### 问题1：订阅授权失败（原错误）

**错误信息：**
```
ERROR: there is no unique or exclusion constraint matching the ON CONFLICT specification (SQLSTATE 42P10)
```

**解决方案：** ✅ 已修复
- 修改了 `UpdateOrCreateSubscribe` 方法
- 使用显式的查询-创建-更新模式
- 文件：`subscribe_repository_impl.go:61-86`

---

### 问题2：消息队列添加失败

**可能原因1：template_id 为空**

检查：
```sql
SELECT * FROM subscribe_records WHERE template_id IS NULL OR template_id = '';
```

解决：确保前端调用时传递正确的 `templateId`

**可能原因2：订阅记录不存在**

检查：
```sql
SELECT * FROM subscribe_records
WHERE openid = '你的openid'
  AND template_type = 'breast_feeding_reminder';
```

解决：先调用订阅授权接口

**可能原因3：订阅已过期**

检查：
```sql
SELECT
    openid,
    template_type,
    status,
    expire_time,
    NOW() AS current_time,
    CASE
        WHEN expire_time < NOW() THEN '已过期'
        ELSE '有效'
    END AS validity
FROM subscribe_records
WHERE openid = '你的openid';
```

解决：重新授权订阅（有效期30天）

---

### 问题3：消息未发送

**检查定时任务是否运行：**
```bash
# 查看日志，应该每分钟输出一次
tail -f logs/server.log | grep "Processing message queue"
```

**检查是否有待发送消息：**
```sql
SELECT * FROM message_send_queue
WHERE status = 'pending'
  AND scheduled_time <= NOW();
```

**检查业务触发条件：**
```sql
-- 检查是否有超过3小时未喂养的宝宝
SELECT
    baby_id,
    MAX(time) AS last_feeding_time,
    EXTRACT(EPOCH FROM (NOW() - TO_TIMESTAMP(MAX(time)/1000)))/3600 AS hours_since_last
FROM feeding_records
GROUP BY baby_id
HAVING EXTRACT(EPOCH FROM (NOW() - TO_TIMESTAMP(MAX(time)/1000)))/3600 >= 3;
```

---

### 问题4：微信API调用失败

**检查配置：**
```yaml
# config/config.yaml
wechat:
  app_id: "wx1234567890abcdef"  # 确保正确
  app_secret: "your_app_secret"  # 确保正确
```

**测试 access_token：**
```bash
curl "https://api.weixin.qq.com/cgi-bin/token?grant_type=client_credential&appid=你的AppID&secret=你的AppSecret"
```

**查看错误日志：**
```sql
SELECT
    openid,
    template_type,
    errcode,
    errmsg,
    send_time,
    created_at
FROM message_send_logs
WHERE send_status = 'failed'
ORDER BY created_at DESC;
```

**常见微信错误码：**
- `40003`: 无效的 openid
- `43101`: 用户拒绝接收消息（需要重新授权）
- `47003`: 模板参数不正确（检查字段名称）
- `41030`: page路径不正确

---

## 🧪 完整测试脚本

**一键测试脚本（需要有 psql 和 curl）：**

```bash
#!/bin/bash

echo "========================================="
echo "  订阅消息系统测试脚本"
echo "========================================="
echo ""

# 配置
DATABASE="nutri_baby"
USER="wxl"
HOST="localhost"
API_URL="http://localhost:8080/api/v1"
TOKEN="YOUR_JWT_TOKEN"  # 替换为真实 token
TEMPLATE_ID="YOUR_TEMPLATE_ID"  # 替换为微信模板ID
BABY_ID="YOUR_BABY_ID"  # 替换为宝宝ID

# 1. 检查数据库表
echo ">>> 步骤1：检查数据库表"
psql -h $HOST -U $USER -d $DATABASE -c "SELECT tablename FROM pg_tables WHERE tablename IN ('subscribe_records', 'message_send_queue', 'message_send_logs');"
echo ""

# 2. 授权订阅
echo ">>> 步骤2：授权订阅"
curl -X POST "$API_URL/subscribe/auth" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"records\": [
      {
        \"templateId\": \"$TEMPLATE_ID\",
        \"templateType\": \"breast_feeding_reminder\",
        \"status\": \"accept\"
      }
    ]
  }"
echo -e "\n"

# 3. 创建喂养记录（3小时前）
echo ">>> 步骤3：创建喂养记录（3小时前）"
THREE_HOURS_AGO=$(($(date +%s) * 1000 - 3 * 60 * 60 * 1000))
curl -X POST "$API_URL/babies/$BABY_ID/feeding-records" \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d "{
    \"type\": \"breast\",
    \"time\": $THREE_HOURS_AGO,
    \"detail\": {
      \"side\": \"left\",
      \"duration\": 15
    }
  }"
echo -e "\n"

# 4. 等待1分钟（定时任务触发）
echo ">>> 步骤4：等待定时任务触发（60秒）"
for i in {60..1}; do
    echo -ne "剩余时间: $i 秒\r"
    sleep 1
done
echo ""

# 5. 检查消息队列
echo ">>> 步骤5：检查消息队列"
psql -h $HOST -U $USER -d $DATABASE -c "SELECT id, template_type, status, scheduled_time FROM message_send_queue ORDER BY created_at DESC LIMIT 5;"
echo ""

# 6. 检查发送日志
echo ">>> 步骤6：检查发送日志"
psql -h $HOST -U $USER -d $DATABASE -c "SELECT id, template_type, send_status, errcode, send_time FROM message_send_logs ORDER BY created_at DESC LIMIT 5;"
echo ""

echo "========================================="
echo "  测试完成！"
echo "========================================="
```

**保存为 `test_subscribe.sh` 并执行：**
```bash
chmod +x test_subscribe.sh
./test_subscribe.sh
```

---

## 📊 成功标准

测试成功的标志：

1. ✅ 订阅授权返回 `successCount: 1`
2. ✅ `subscribe_records` 表有记录，`status = 'active'`
3. ✅ 创建喂养记录后1分钟内，日志输出 `Feeding reminder queued`
4. ✅ `message_send_queue` 表有记录，`status = 'pending'`
5. ✅ 再等1分钟，日志输出 `Message sent successfully`
6. ✅ `message_send_logs` 表有记录，`send_status = 'success'`
7. ✅ 微信小程序收到订阅消息

---

## 🔧 调试技巧

### 1. 实时查看日志
```bash
# 过滤订阅相关日志
tail -f logs/server.log | grep -i subscribe

# 过滤定时任务日志
tail -f logs/server.log | grep -E "Scheduler|Feeding|Message"
```

### 2. 手动插入测试消息
```sql
-- 直接插入一条待发送消息（跳过业务逻辑）
INSERT INTO message_send_queue (
    openid,
    template_id,
    template_type,
    data,
    page,
    scheduled_time,
    status,
    created_at,
    updated_at
) VALUES (
    'om8hB12mqHOp1BiTf3KZ_ew8eWH4',
    '你的模板ID',
    'breast_feeding_reminder',
    '{"lastTime": "14:30", "sinceTime": "3小时", "lastSide": "左侧", "reminderTip": "该喂奶啦"}',
    'pages/record/feeding/feeding',
    NOW(),
    'pending',
    NOW(),
    NOW()
);
```

### 3. 强制触发定时任务
```go
// 临时修改 scheduler_service.go:299
if hoursSinceLastFeeding >= 0.001 { // 改为很小的值，立即触发
```

---

## 📞 需要帮助？

如果测试失败，请提供：

1. **完整的错误日志**
   ```bash
   tail -100 logs/server.log > debug.log
   ```

2. **数据库诊断结果**
   ```bash
   psql -h localhost -U wxl -d nutri_baby -f diagnose_subscribe.sql > db_status.txt
   ```

3. **相关表的数据**
   ```sql
   -- 导出订阅记录
   \copy (SELECT * FROM subscribe_records) TO 'subscribe_records.csv' CSV HEADER;

   -- 导出消息队列
   \copy (SELECT * FROM message_send_queue) TO 'queue.csv' CSV HEADER;

   -- 导出发送日志
   \copy (SELECT * FROM message_send_logs) TO 'logs.csv' CSV HEADER;
   ```

将这些信息提供给开发团队进行诊断。
