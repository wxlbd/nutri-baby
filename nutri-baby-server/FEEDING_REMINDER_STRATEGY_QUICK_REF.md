# 喂养提醒策略模式快速参考

## 如何添加新的喂养类型

### 步骤 1: 实现策略接口

```go
// 例如: 添加"水果"喂养类型
type FruitFeedingReminderStrategy struct{}

func NewFruitFeedingReminderStrategy() *FruitFeedingReminderStrategy {
    return &FruitFeedingReminderStrategy{}
}

func (s *FruitFeedingReminderStrategy) GetTemplateType() string {
    return "fruit_feeding_reminder"
}

func (s *FruitFeedingReminderStrategy) BuildMessageData(
    record *entity.FeedingRecord,
    lastFeedingTime time.Time,
    hoursSinceLastFeeding float64,
) map[string]interface{} {
    fruitName := "水果"
    if name, ok := record.Detail["fruitName"].(string); ok && name != "" {
        fruitName = name
    }

    return map[string]interface{}{
        "time1":             lastFeedingTime.Format("2006-01-02 15:04"),
        "thing2":            formatTimeSince(hoursSinceLastFeeding),
        "character_string3": fruitName,
        "phrase4":           "水果",
        "thing5":            "该给宝宝吃水果啦，注意清洗干净",
    }
}

func (s *FruitFeedingReminderStrategy) CanHandle(record *entity.FeedingRecord) bool {
    feedingType, ok := record.Detail["type"].(string)
    return ok && feedingType == "fruit"
}
```

### 步骤 2: 注册到工厂

```go
// 在 feeding_reminder_strategy.go 中修改工厂
func NewFeedingReminderStrategyFactory() *FeedingReminderStrategyFactory {
    return &FeedingReminderStrategyFactory{
        strategies: []FeedingReminderStrategy{
            NewBreastFeedingReminderStrategy(),
            NewBottleFeedingReminderStrategy(),
            NewFoodFeedingReminderStrategy(),
            NewFruitFeedingReminderStrategy(), // 添加新策略
        },
    }
}
```

### 步骤 3: 添加单元测试

```go
func TestFruitFeedingReminderStrategy(t *testing.T) {
    strategy := NewFruitFeedingReminderStrategy()

    // 测试 GetTemplateType
    assert.Equal(t, "fruit_feeding_reminder", strategy.GetTemplateType())

    // 测试 CanHandle
    record := &entity.FeedingRecord{
        Detail: entity.FeedingDetail{
            "type":      "fruit",
            "fruitName": "苹果",
        },
    }
    assert.True(t, strategy.CanHandle(record))

    // 测试 BuildMessageData
    lastFeedingTime := time.Now().Add(-2 * time.Hour)
    messageData := strategy.BuildMessageData(record, lastFeedingTime, 2.0)

    assert.NotNil(t, messageData)
    assert.Equal(t, "水果", messageData["phrase4"])
    assert.Equal(t, "苹果", messageData["character_string3"])
}
```

### 步骤 4: 在微信公众平台配置模板

在微信公众平台添加 `fruit_feeding_reminder` 订阅消息模板，配置对应字段。

## 常见问题

### Q: 如何修改现有策略的提示语？

**A**: 直接修改对应策略的 `BuildMessageData` 方法中的 `thing5` 字段。

### Q: 如何支持自定义字段？

**A**: 在 `BuildMessageData` 方法中从 `record.Detail` 中提取所需字段即可。

### Q: 如何设置默认策略？

**A**: 在 `FeedingReminderStrategyFactory.GetStrategy` 方法的最后返回默认策略。

### Q: 策略的执行顺序重要吗？

**A**: 重要！工厂会按照注册顺序依次调用 `CanHandle` 方法，第一个返回 `true` 的策略会被使用。

## 调试技巧

### 1. 查看选中的策略

在日志中查找 `🎯 [CheckFeedingReminders] 获取喂养提醒策略`，可以看到选中的模板类型。

### 2. 查看构造的消息数据

在日志中查找 `📦 [CheckFeedingReminders] 消息数据构造完成`，可以看到完整的消息数据。

### 3. 单元测试

运行特定策略的单元测试：

```bash
go test -v -run TestBreastFeedingReminderStrategy ./internal/application/service/
```

## 性能优化建议

1. **避免重复创建策略**: 策略在工厂初始化时创建，并复用
2. **提前获取策略**: 在循环外部获取策略，避免重复调用
3. **缓存模板类型**: 模板类型在策略创建时就确定，不需要重复计算

## 架构图

```
CheckFeedingReminders
       │
       ├─► 1. 获取宝宝列表
       ├─► 2. 查询最近喂养记录
       ├─► 3. 查询宝宝协作者
       │
       ├─► 4. 获取喂养提醒策略 ◄─────┐
       │                              │
       │   FeedingReminderStrategyFactory
       │                              │
       │   ┌──────────────────────────┴─────────────┐
       │   │                                        │
       │   ▼                                        ▼
       │   BreastFeedingReminderStrategy   BottleFeedingReminderStrategy
       │                                             │
       │                                             ▼
       │                                   FoodFeedingReminderStrategy
       │
       ├─► 5. 检查用户授权状态 (使用策略返回的模板类型)
       ├─► 6. 构造消息数据 (使用策略)
       └─► 7. 发送订阅消息
```

## 相关文件

- **策略接口和实现**: `internal/application/service/feeding_reminder_strategy.go`
- **单元测试**: `internal/application/service/feeding_reminder_strategy_test.go`
- **调度服务**: `internal/application/service/scheduler_service.go`
- **示例代码**: `examples/feeding_reminder_strategy_example.go`
- **重构文档**: `FEEDING_REMINDER_STRATEGY_REFACTOR.md`

---

**最后更新**: 2025-10-26
