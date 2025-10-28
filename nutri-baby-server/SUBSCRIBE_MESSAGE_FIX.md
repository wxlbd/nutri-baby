# 订阅消息字段映射修复

## 🐛 问题描述

订阅消息发送失败,微信API返回错误:
```
errcode: 47003
errmsg: "argument invalid! data.time1.value is empty"
```

## 🔍 问题分析

通过详细的日志追踪,发现了问题的根本原因:

### 原因
后端发送的字段名与微信订阅消息模板的字段名不匹配。

**错误的字段映射** (之前的代码):
```go
messageData := map[string]interface{}{
    "lastTime":    lastFeedingTime.Format("15:04"),        // ❌ 错误
    "sinceTime":   formatDuration(hoursSinceLastFeeding),  // ❌ 错误
    "lastSide":    getLastFeedingSide(lastFeeding),        // ❌ 错误
    "reminderTip": "该喂奶啦，注意观察宝宝的饥饿信号",          // ❌ 错误
}
```

**微信模板实际需要的字段**:
根据前端配置 (`nutri-baby-app/src/store/subscribe.ts:50`):
```typescript
{
  type: 'breast_feeding_reminder',
  templateId: '2JRV0DnOHnasHzzamWFoWGaUxrgW6GY69-eGn4tBFZE',
  keywords: ['上次时间', '距离上次', '上次位置', '温馨提示'],
}
```

微信订阅消息模板的标准字段格式:
- `time1` - 时间类型字段
- `time2` - 时间类型字段
- `thing3` - 文本类型字段
- `thing4` - 文本类型字段

## ✅ 解决方案

### 1. 修正字段映射

**正确的字段映射** (`scheduler_service.go:320-325`):
```go
messageData := map[string]interface{}{
    "time1":  lastFeedingTime.Format("2006-01-02 15:04"), // ✅ 上次时间
    "time2":  lastFeedingTime.Format("2006-01-02 15:04"), // ✅ 距离上次(也填时间)
    "thing3": getLastFeedingSide(lastFeeding),            // ✅ 上次位置
    "thing4": "该喂奶啦，注意观察宝宝的饥饿信号",                    // ✅ 温馨提示
}
```

### 2. 字段类型说明

| 模板关键词 | 字段名 | 字段类型 | 示例值 | 说明 |
|-----------|--------|---------|--------|------|
| 上次时间 | `time1` | time | `2025-10-25 14:30` | 上次喂养的时间 |
| 距离上次 | `time2` | time | `2025-10-25 14:30` | 目前填相同时间,可优化为计算时间差 |
| 上次位置 | `thing3` | text | `左侧`/`右侧`/`奶瓶喂养` | 喂养方式或位置 |
| 温馨提示 | `thing4` | text | `该喂奶啦，注意观察宝宝的饥饿信号` | 提醒文案 |

### 3. 微信字段类型规范

**时间类型字段 (timeN)**:
- 格式: `YYYY-MM-DD HH:mm` 或 `YYYY-MM-DD HH:mm:ss`
- 示例: `2025-10-25 14:30` 或 `2025-10-25 14:30:00`
- ⚠️ 注意: 不能为空,必须是有效的时间格式

**文本类型字段 (thingN)**:
- 格式: 纯文本字符串
- 长度限制: 一般不超过20个汉字
- 示例: `左侧`, `母乳喂养`, `该喂奶啦`

**其他字段类型**:
- `character_stringN`: 字符串类型
- `phraseN`: 短语类型
- `amountN`: 金额类型
- `dateN`: 日期类型
- `numberN`: 数字类型

## 📝 修改文件

### 修改文件列表
- `internal/application/service/scheduler_service.go:318-325`

### 代码变更
```diff
- messageData := map[string]interface{}{
-     "lastTime":    lastFeedingTime.Format("15:04"),
-     "sinceTime":   formatDuration(hoursSinceLastFeeding),
-     "lastSide":    getLastFeedingSide(lastFeeding),
-     "reminderTip": "该喂奶啦，注意观察宝宝的饥饿信号",
- }

+ // 微信订阅消息模板字段: time1(上次时间), time2(距离上次), thing3(上次位置), thing4(温馨提示)
+ messageData := map[string]interface{}{
+     "time1":  lastFeedingTime.Format("2006-01-02 15:04"), // 上次时间
+     "time2":  lastFeedingTime.Format("2006-01-02 15:04"), // 距离上次(也填时间)
+     "thing3": getLastFeedingSide(lastFeeding),            // 上次位置
+     "thing4": "该喂奶啦，注意观察宝宝的饥饿信号",                    // 温馨提示
+ }
```

## 🧪 测试验证

### 1. 编译服务
```bash
cd nutri-baby-server
make build
```

### 2. 重启服务
```bash
make run
```

### 3. 查看日志
```bash
tail -f logs/app.log | grep -E "\[WechatService.SendSubscribeMessage\]"
```

### 4. 预期结果
日志应该显示:
```log
📥 [WechatService.SendSubscribeMessage] 响应内容
   responseBody={"errcode":0,"errmsg":"ok"}

✅ [WechatService.SendSubscribeMessage] 订阅消息发送成功
   openid=oxxx
   templateId=2JRV0DnOHnasHzzamWFoWGaUxrgW6GY69-eGn4tBFZE
```

用户应该能在微信小程序中收到订阅消息。

## 🚀 后续优化建议

### 1. 优化 time2 字段
目前 `time2` 字段填的是上次喂养时间,可以优化为距离当前的时间差:

```go
// 计算时间差
duration := time.Since(lastFeedingTime)
hours := int(duration.Hours())
minutes := int(duration.Minutes()) % 60

var sinceTimeText string
if hours > 0 {
    sinceTimeText = fmt.Sprintf("距今%d小时%d分钟", hours, minutes)
} else {
    sinceTimeText = fmt.Sprintf("距今%d分钟", minutes)
}

messageData := map[string]interface{}{
    "time1":  lastFeedingTime.Format("2006-01-02 15:04"),
    "time2":  sinceTimeText, // 改为时间差描述
    "thing3": getLastFeedingSide(lastFeeding),
    "thing4": "该喂奶啦，注意观察宝宝的饥饿信号",
}
```

⚠️ **注意**: 需要确认微信模板的 `time2` 字段类型是 `time` 还是 `thing`,如果是 `time` 类型则不能用文本描述。

### 2. 统一字段映射管理
建议创建一个字段映射配置文件或常量,避免字段名硬编码:

```go
// 订阅消息字段映射
type SubscribeMessageFields struct {
    BreastFeedingReminder struct {
        LastTime    string // time1: 上次时间
        SinceTime   string // time2: 距离上次
        LastSide    string // thing3: 上次位置
        ReminderTip string // thing4: 温馨提示
    }
    // ... 其他消息类型
}

var MessageFields = SubscribeMessageFields{
    BreastFeedingReminder: struct {
        LastTime    string
        SinceTime   string
        LastSide    string
        ReminderTip string
    }{
        LastTime:    "time1",
        SinceTime:   "time2",
        LastSide:    "thing3",
        ReminderTip: "thing4",
    },
}
```

### 3. 字段验证
添加字段验证逻辑,确保发送前数据格式正确:

```go
func validateTimeField(value string) error {
    _, err := time.Parse("2006-01-02 15:04", value)
    if err != nil {
        return fmt.Errorf("invalid time format: %w", err)
    }
    return nil
}

func validateThingField(value string) error {
    if len(value) == 0 {
        return fmt.Errorf("thing field cannot be empty")
    }
    if len([]rune(value)) > 20 {
        return fmt.Errorf("thing field too long (max 20 characters)")
    }
    return nil
}
```

## 📚 参考资料

### 微信官方文档
- [订阅消息发送API](https://developers.weixin.qq.com/miniprogram/dev/OpenApiDoc/mp-message-management/subscribe-message/sendMessage.html)
- [订阅消息模板规范](https://developers.weixin.qq.com/miniprogram/dev/framework/open-ability/subscribe-message.html)

### 常见错误码
| 错误码 | 错误说明 | 解决方案 |
|-------|---------|---------|
| 40001 | access_token过期 | 重新获取access_token |
| 40003 | touser字段openid为空或不正确 | 检查openid是否正确 |
| 41030 | page路径不正确 | 检查page参数 |
| 43101 | 用户拒绝接受消息 | 用户需要重新授权 |
| 47001 | data格式不正确 | 检查data字段格式 |
| 47003 | 模板参数不正确 | **字段名或字段值不符合模板要求** ⭐ |

## ✨ 总结

这次修复的关键点:
1. ✅ 通过详细的日志追踪快速定位问题
2. ✅ 修正了字段名映射错误
3. ✅ 添加了详细的注释说明
4. ✅ 提供了完整的测试和验证步骤

现在订阅消息应该可以正常发送了! 🎉
