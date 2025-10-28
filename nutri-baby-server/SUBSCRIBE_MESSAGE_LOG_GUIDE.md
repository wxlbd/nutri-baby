# 订阅消息发送链路日志追踪指南

## 📋 概述

本文档介绍了如何通过日志追踪订阅消息的完整发送链路,帮助你快速定位订阅消息未收到的问题。

## 🔍 日志追踪链路

订阅消息的发送链路包含以下关键步骤,每个步骤都有详细的日志标记:

### 1️⃣ 定时任务触发 (SchedulerService)

**日志标识**: `[CheckFeedingReminders]`

```log
🔔 [CheckFeedingReminders] ===== START =====
⏰ [CheckFeedingReminders] 定时任务触发时间: 2025-10-25 14:00:00
```

**关键日志点**:
- ✅ 找到宝宝列表
- 📅 查询时间范围
- 👶 处理每个宝宝
- 📊 上次喂养时间分析
- ⚙️ 提醒阈值配置
- ⏰ 是否需要发送提醒

**可能的问题**:
- 如果日志显示 `ℹ️ 系统中没有宝宝,跳过检查` → 数据库中没有宝宝数据
- 如果日志显示 `ℹ️ 该宝宝没有喂养记录,跳过` → 宝宝没有喂养记录
- 如果日志显示 `ℹ️ 距离上次喂养时间未达到提醒阈值,跳过` → 时间未到,等待下次检查

### 2️⃣ 查找协作者 (BabyCollaboratorRepository)

**日志标识**: `[CheckFeedingReminders] STEP 3`

```log
🔍 [CheckFeedingReminders] STEP 3 - 查询宝宝的协作者
✅ [CheckFeedingReminders] 找到协作者, collaboratorCount=2
```

**可能的问题**:
- 如果日志显示 `⚠️ 该宝宝没有协作者,无法发送提醒` → 数据库中没有协作者记录

### 3️⃣ 检查授权状态 (SubscribeService)

**日志标识**: `[CheckFeedingReminders] STEP 4`

```log
🔍 [CheckFeedingReminders] STEP 4 - 检查用户授权状态
   openid=oxxx
   templateType=breast_feeding_reminder
```

**可能的问题**:
- 如果日志显示 `⚠️ 用户没有可用授权,跳过` → 用户未授权或授权已使用

### 4️⃣ 发送订阅消息 - 第一层 (SubscribeService)

**日志标识**: `[SendSubscribeMessage]`

```log
📤 [SendSubscribeMessage] START - 开始发送订阅消息
   openid=oxxx
   templateType=breast_feeding_reminder
   page=pages/record/feeding/feeding
   data={"lastTime":"14:00","sinceTime":"3小时",...}
```

**STEP 1: 查询可用授权记录**
```log
🔍 [SendSubscribeMessage] STEP 1 - 查询可用授权记录
✅ [SendSubscribeMessage] 找到可用授权记录
   templateID=xxxx
   status=available
   authorizeTime=2025-10-25 10:00:00
   expireTime=2025-11-01 10:00:00
```

**可能的问题**:
- `❌ 查询授权记录失败` → 数据库查询错误
- `⚠️ 未找到可用授权记录` → 用户未授权或授权已使用/过期

**STEP 2: 检查授权是否可用**
```log
🔍 [SendSubscribeMessage] STEP 2 - 检查授权是否可用
   status=available
✅ [SendSubscribeMessage] 授权可用,准备调用微信API
```

**可能的问题**:
- `⚠️ 授权不可用` → 授权状态为 used/expired

**STEP 3: 调用微信API**
```log
📞 [SendSubscribeMessage] STEP 3 - 调用微信API发送订阅消息
   openid=oxxx
   templateID=xxxx
   page=pages/record/feeding/feeding
   data={"lastTime":{"value":"14:00"},...}
```

**STEP 4: 标记授权为已使用**
```log
🔄 [SendSubscribeMessage] STEP 4 - 标记授权为已使用
✅ [SendSubscribeMessage] 授权状态已更新为已使用
```

**STEP 5: 保存发送日志**
```log
📝 [SendSubscribeMessage] STEP 5 - 保存发送日志
✅ [SendSubscribeMessage] 订阅消息发送成功
   errCode=0
   errMsg=ok
```

### 5️⃣ 发送订阅消息 - 第二层 (WechatService)

**日志标识**: `[WechatService.SendSubscribeMessage]`

```log
🚀 [WechatService.SendSubscribeMessage] START - 开始发送微信订阅消息
   openid=oxxx
   templateID=xxxx
   page=pages/record/feeding/feeding
   miniprogramState=formal
```

**STEP 1: 获取 access_token**
```log
🔑 [WechatService.SendSubscribeMessage] STEP 1 - 获取access_token
🔑 [getAccessToken] 使用缓存的access_token
   token=abc12...xyz89
   expiry=2025-10-25 16:00:00
```

或者首次获取:
```log
🔄 [getAccessToken] 开始获取新的access_token
   appid=wxxx
📞 [getAccessToken] 调用微信API获取access_token
📥 [getAccessToken] 收到微信API响应
   response={"access_token":"xxx","expires_in":7200}
✅ [getAccessToken] access_token获取成功
```

**可能的问题**:
- `❌ 请求access_token失败` → 网络问题或AppID/AppSecret错误
- `❌ 微信API返回错误` → 检查 errcode 和 errmsg

**STEP 2: 格式化模板数据**
```log
🔄 [WechatService.SendSubscribeMessage] STEP 2 - 格式化模板数据
✅ [WechatService.SendSubscribeMessage] 数据格式化完成
   formattedData={"lastTime":{"value":"14:00"},...}
```

**STEP 3: 构建请求体**
```log
📦 [WechatService.SendSubscribeMessage] STEP 3 - 构建请求体
✅ [WechatService.SendSubscribeMessage] 请求体构建完成
   requestBody={"touser":"oxxx","template_id":"xxxx",...}
```

**STEP 4: 调用微信API**
```log
📞 [WechatService.SendSubscribeMessage] STEP 4 - 调用微信订阅消息API
   url=https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=***
   requestBodySize=256
📥 [WechatService.SendSubscribeMessage] 收到HTTP响应
   statusCode=200
   status=200 OK
📥 [WechatService.SendSubscribeMessage] 响应内容
   responseBody={"errcode":0,"errmsg":"ok"}
```

**可能的问题**:
- `❌ HTTP请求失败` → 网络问题
- `statusCode != 200` → HTTP层错误

**STEP 5: 检查发送结果**
```log
🔍 [WechatService.SendSubscribeMessage] STEP 5 - 检查发送结果
   errcode=0
   errmsg=ok
✅ [WechatService.SendSubscribeMessage] 订阅消息发送成功
🏁 [WechatService.SendSubscribeMessage] END - 订阅消息发送完成
```

**可能的问题**:
- `errcode=43101` → 用户拒绝接收消息
- `errcode=47003` → 模板ID不正确
- `errcode=41030` → 不合法的page路径
- 其他errcode → 参考微信官方文档

### 6️⃣ 任务完成

```log
🏁 [CheckFeedingReminders] ===== END =====
   endTime=2025-10-25 14:00:05
```

## 🐛 常见问题排查

### 问题1: 定时任务没有触发

**日志特征**: 没有看到 `🔔 [CheckFeedingReminders] ===== START =====`

**排查步骤**:
1. 检查服务是否启动成功
2. 查看日志中是否有 `Scheduler service started`
3. 检查cron表达式配置 (当前为每1分钟触发一次)

### 问题2: 定时任务触发但跳过所有宝宝

**日志特征**:
```log
ℹ️ [CheckFeedingReminders] 系统中没有宝宝,跳过检查
```

**排查步骤**:
1. 检查数据库中 `babies` 表是否有数据
2. 使用SQL查询: `SELECT * FROM babies WHERE deleted_at IS NULL;`

### 问题3: 找到宝宝但没有喂养记录

**日志特征**:
```log
ℹ️ [CheckFeedingReminders] 该宝宝没有喂养记录,跳过
```

**排查步骤**:
1. 检查数据库中 `feeding_records` 表是否有数据
2. 使用SQL查询: `SELECT * FROM feeding_records WHERE baby_id='xxx' ORDER BY time DESC LIMIT 1;`
3. 确保有最近24小时内的喂养记录

### 问题4: 时间未达到提醒阈值

**日志特征**:
```log
ℹ️ [CheckFeedingReminders] 距离上次喂养时间未达到提醒阈值,跳过
   hoursSinceLastFeeding=0.5
   thresholdHours=0.016
```

**排查步骤**:
1. 检查 `hoursSinceLastFeeding` 和 `thresholdHours` 的值
2. 当前测试环境阈值为 0.016 小时 (~1分钟)
3. 如果需要立即触发,可以添加一条旧的喂养记录

### 问题5: 没有找到协作者

**日志特征**:
```log
⚠️ [CheckFeedingReminders] 该宝宝没有协作者,无法发送提醒
```

**排查步骤**:
1. 检查数据库中 `baby_collaborators` 表
2. 使用SQL查询: `SELECT * FROM baby_collaborators WHERE baby_id='xxx';`
3. 确保至少有一条协作者记录,且包含正确的 `openid`

### 问题6: 用户没有可用授权

**日志特征**:
```log
⚠️ [CheckFeedingReminders] 用户没有可用授权,跳过
```

**排查步骤**:
1. 检查数据库中 `subscribe_records` 表
2. 使用SQL查询授权记录:
```sql
SELECT * FROM subscribe_records
WHERE openid='xxx'
AND template_type='breast_feeding_reminder'
AND status='available'
AND expire_time > NOW()
ORDER BY authorize_time DESC;
```
3. 如果没有记录,需要用户在小程序中重新授权
4. 可以手动插入测试授权记录:
```sql
INSERT INTO subscribe_records (openid, template_id, template_type, status, authorize_time, expire_time)
VALUES ('oxxx', 'your_template_id', 'breast_feeding_reminder', 'available', NOW(), NOW() + INTERVAL '7 days');
```

### 问题7: 微信API返回错误

**日志特征**:
```log
⚠️ [WechatService.SendSubscribeMessage] 微信API返回业务错误
   errcode=43101
   errmsg=用户拒绝接受消息
```

**常见错误码**:
- `errcode=40001` - AppSecret错误或access_token已过期
- `errcode=40003` - touser字段openid不正确
- `errcode=41030` - page路径不合法
- `errcode=43101` - 用户拒绝接受消息,需重新订阅
- `errcode=47003` - 模板ID不存在或不正确
- `errcode=43102` - 订阅关系已过期,需重新订阅

**排查步骤**:
1. 查看完整的日志,找到具体的 errcode 和 errmsg
2. 参考微信官方文档: https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-message-management/subscribe-message/sendMessage.html
3. 检查配置文件中的 AppID 和 AppSecret 是否正确
4. 检查模板ID是否正确
5. 检查用户是否已授权

### 问题8: 网络问题

**日志特征**:
```log
❌ [WechatService.SendSubscribeMessage] HTTP请求失败
   error=connection timeout
```

**排查步骤**:
1. 检查服务器网络连接
2. 尝试手动访问微信API: `curl https://api.weixin.qq.com`
3. 检查防火墙设置

## 📊 完整日志示例

### 成功发送的完整日志链路:

```log
2025-10-25 14:00:00 INFO 🔔 [CheckFeedingReminders] ===== START =====
2025-10-25 14:00:00 INFO ⏰ [CheckFeedingReminders] 定时任务触发时间
2025-10-25 14:00:00 INFO 🔍 [CheckFeedingReminders] STEP 1 - 获取所有宝宝列表
2025-10-25 14:00:00 INFO ✅ [CheckFeedingReminders] 找到宝宝 babyCount=1
2025-10-25 14:00:00 INFO 👶 [CheckFeedingReminders] 处理宝宝 babyId=baby123 babyName=小明
2025-10-25 14:00:00 INFO 🔍 [CheckFeedingReminders] STEP 2 - 查询最近喂养记录
2025-10-25 14:00:00 INFO 📊 [CheckFeedingReminders] 上次喂养时间分析 hoursSinceLastFeeding=3.5
2025-10-25 14:00:00 INFO ⚙️ [CheckFeedingReminders] 提醒阈值配置 thresholdHours=0.016 shouldRemind=true
2025-10-25 14:00:00 INFO ⏰ [CheckFeedingReminders] 需要发送喂养提醒
2025-10-25 14:00:00 INFO 🔍 [CheckFeedingReminders] STEP 3 - 查询宝宝的协作者
2025-10-25 14:00:00 INFO ✅ [CheckFeedingReminders] 找到协作者 collaboratorCount=1
2025-10-25 14:00:00 INFO 👤 [CheckFeedingReminders] 处理协作者 openid=oABC123
2025-10-25 14:00:00 INFO 🔍 [CheckFeedingReminders] STEP 4 - 检查用户授权状态
2025-10-25 14:00:00 INFO ✅ [CheckFeedingReminders] 用户有可用授权,准备发送提醒
2025-10-25 14:00:00 INFO 📦 [CheckFeedingReminders] STEP 5 - 构造消息数据
2025-10-25 14:00:00 INFO 📤 [CheckFeedingReminders] STEP 6 - 发送订阅消息
2025-10-25 14:00:00 INFO 📤 [SendSubscribeMessage] START - 开始发送订阅消息
2025-10-25 14:00:00 INFO 🔍 [SendSubscribeMessage] STEP 1 - 查询可用授权记录
2025-10-25 14:00:00 INFO ✅ [SendSubscribeMessage] 找到可用授权记录
2025-10-25 14:00:00 INFO 🔍 [SendSubscribeMessage] STEP 2 - 检查授权是否可用
2025-10-25 14:00:00 INFO ✅ [SendSubscribeMessage] 授权可用,准备调用微信API
2025-10-25 14:00:00 INFO 📞 [SendSubscribeMessage] STEP 3 - 调用微信API发送订阅消息
2025-10-25 14:00:00 INFO 🚀 [WechatService.SendSubscribeMessage] START - 开始发送微信订阅消息
2025-10-25 14:00:00 INFO 🔑 [WechatService.SendSubscribeMessage] STEP 1 - 获取access_token
2025-10-25 14:00:00 INFO 🔑 [getAccessToken] 使用缓存的access_token
2025-10-25 14:00:00 INFO 🔄 [WechatService.SendSubscribeMessage] STEP 2 - 格式化模板数据
2025-10-25 14:00:00 INFO ✅ [WechatService.SendSubscribeMessage] 数据格式化完成
2025-10-25 14:00:00 INFO 📦 [WechatService.SendSubscribeMessage] STEP 3 - 构建请求体
2025-10-25 14:00:00 INFO ✅ [WechatService.SendSubscribeMessage] 请求体构建完成
2025-10-25 14:00:00 INFO 📞 [WechatService.SendSubscribeMessage] STEP 4 - 调用微信订阅消息API
2025-10-25 14:00:01 INFO 📥 [WechatService.SendSubscribeMessage] 收到HTTP响应 statusCode=200
2025-10-25 14:00:01 INFO 📥 [WechatService.SendSubscribeMessage] 响应内容 responseBody={"errcode":0,"errmsg":"ok"}
2025-10-25 14:00:01 INFO 🔍 [WechatService.SendSubscribeMessage] STEP 5 - 检查发送结果 errcode=0
2025-10-25 14:00:01 INFO ✅ [WechatService.SendSubscribeMessage] 订阅消息发送成功
2025-10-25 14:00:01 INFO 🏁 [WechatService.SendSubscribeMessage] END - 订阅消息发送完成
2025-10-25 14:00:01 INFO 🔄 [SendSubscribeMessage] STEP 4 - 标记授权为已使用
2025-10-25 14:00:01 INFO ✅ [SendSubscribeMessage] 授权状态已更新为已使用
2025-10-25 14:00:01 INFO 📝 [SendSubscribeMessage] STEP 5 - 保存发送日志
2025-10-25 14:00:01 INFO ✅ [SendSubscribeMessage] 发送日志已保存
2025-10-25 14:00:01 INFO 🏁 [SendSubscribeMessage] END - 订阅消息发送流程结束
2025-10-25 14:00:01 INFO ✅ [CheckFeedingReminders] 喂养提醒发送成功
2025-10-25 14:00:01 INFO 🏁 [CheckFeedingReminders] ===== END =====
```

## 🛠️ 日志查看命令

### 查看最近的日志
```bash
tail -f logs/app.log
```

### 过滤订阅消息相关日志
```bash
tail -f logs/app.log | grep -E "\[CheckFeedingReminders\]|\[SendSubscribeMessage\]|\[WechatService.SendSubscribeMessage\]"
```

### 查看某个openid的日志
```bash
tail -f logs/app.log | grep "openid=oABC123"
```

### 查看错误日志
```bash
tail -f logs/app.log | grep "ERROR"
```

### 查看警告和错误日志
```bash
tail -f logs/app.log | grep -E "WARN|ERROR"
```

## 📝 数据库诊断SQL

### 检查宝宝数据
```sql
SELECT baby_id, name, created_at FROM babies WHERE deleted_at IS NULL;
```

### 检查喂养记录
```sql
SELECT baby_id, time, type, created_at
FROM feeding_records
WHERE baby_id='your_baby_id'
ORDER BY time DESC
LIMIT 10;
```

### 检查协作者
```sql
SELECT baby_id, openid, role, created_at
FROM baby_collaborators
WHERE baby_id='your_baby_id';
```

### 检查订阅授权记录
```sql
SELECT openid, template_type, status, authorize_time, expire_time
FROM subscribe_records
WHERE openid='your_openid'
ORDER BY authorize_time DESC;
```

### 检查消息发送日志
```sql
SELECT openid, template_type, send_status, err_msg, send_time, created_at
FROM message_send_logs
WHERE openid='your_openid'
ORDER BY created_at DESC
LIMIT 10;
```

## 🎯 快速定位问题的步骤

1. **查看最新日志**: `tail -f logs/app.log`
2. **确认定时任务是否触发**: 查找 `[CheckFeedingReminders] ===== START =====`
3. **确认是否找到宝宝**: 查找 `找到宝宝 babyCount=`
4. **确认是否需要提醒**: 查找 `需要发送喂养提醒`
5. **确认是否找到协作者**: 查找 `找到协作者 collaboratorCount=`
6. **确认授权状态**: 查找 `用户有可用授权` 或 `用户没有可用授权`
7. **确认API调用**: 查找 `[WechatService.SendSubscribeMessage]`
8. **确认发送结果**: 查找 `errcode=` 和 `errmsg=`

通过这些详细的日志,你可以准确定位订阅消息未收到的原因!
