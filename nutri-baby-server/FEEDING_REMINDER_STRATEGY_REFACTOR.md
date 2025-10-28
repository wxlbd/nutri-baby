# 喂养提醒策略模式重构总结

## 重构目标

将 `CheckFeedingReminders` 方法中硬编码的消息模板逻辑重构为基于策略模式的可扩展架构，以支持不同喂养类型使用不同的消息模板。

## 架构设计

### 1. 策略接口定义 (`FeedingReminderStrategy`)

```go
type FeedingReminderStrategy interface {
    GetTemplateType() string
    BuildMessageData(record *entity.FeedingRecord, lastFeedingTime time.Time, hoursSinceLastFeeding float64) map[string]interface{}
    CanHandle(record *entity.FeedingRecord) bool
}
```

### 2. 策略实现

#### 2.1 母乳喂养策略 (`BreastFeedingReminderStrategy`)
- **模板类型**: `breast_feeding_reminder`
- **支持字段**:
  - 喂养侧: 左侧/右侧/两侧
  - 喂养时长: 分钟数
  - 温馨提示: "该喂奶啦，注意观察宝宝的饥饿信号"

#### 2.2 奶瓶喂养策略 (`BottleFeedingReminderStrategy`)
- **模板类型**: `bottle_feeding_reminder`
- **支持字段**:
  - 奶瓶类型: 配方奶/母乳
  - 喂养量: ml
  - 温馨提示: "该喂奶啦，记得准备好奶瓶哦"

#### 2.3 辅食喂养策略 (`FoodFeedingReminderStrategy`)
- **模板类型**: `food_feeding_reminder`
- **支持字段**:
  - 辅食名称
  - 温馨提示: "该给宝宝准备辅食啦，注意观察过敏反应"

### 3. 策略工厂 (`FeedingReminderStrategyFactory`)

负责根据喂养记录的类型选择合适的策略实现：

```go
func (f *FeedingReminderStrategyFactory) GetStrategy(record *entity.FeedingRecord) FeedingReminderStrategy {
    for _, strategy := range f.strategies {
        if strategy.CanHandle(record) {
            return strategy
        }
    }
    // 默认返回母乳喂养策略（向后兼容）
    return NewBreastFeedingReminderStrategy()
}
```

## 重构内容

### 1. 新增文件

- `feeding_reminder_strategy.go`: 策略接口和实现
- `feeding_reminder_strategy_test.go`: 单元测试

### 2. 修改文件

#### `scheduler_service.go`

**添加字段**:
```go
type SchedulerService struct {
    // ...
    strategyFactory *FeedingReminderStrategyFactory
    // ...
}
```

**构造函数更新**:
```go
func NewSchedulerService(...) *SchedulerService {
    return &SchedulerService{
        // ...
        strategyFactory: NewFeedingReminderStrategyFactory(),
        // ...
    }
}
```

**CheckFeedingReminders 方法重构**:

**重构前**:
```go
// 硬编码的消息数据
messageData := map[string]interface{}{
    "time1":             lastFeedingTime.Format("2006-01-02 15:04"),
    "thing2":            time.Now().Sub(lastFeedingTime).Seconds(),
    "character_string3": fmt.Sprintf("%dml", lastFeeding.Detail["amount"]),
    "phrase4":           "奶粉",
    "thing5":            "该喂奶啦，注意观察宝宝的饥饿信号",
}

sendReq := &dto.SendMessageRequest{
    OpenID:       collaborator.OpenID,
    TemplateType: "breast_feeding_reminder", // 硬编码
    Data:         messageData,
    Page:         "pages/record/feeding/feeding",
}
```

**重构后**:
```go
// 4. 根据喂养类型获取策略
strategy := s.strategyFactory.GetStrategy(lastFeeding)
templateType := strategy.GetTemplateType()

// 5. 检查每个协作者的授权状态（使用动态模板类型）
hasAuth, err := s.subscribeService.CheckAuthorizationStatus(ctx, collaborator.OpenID, templateType)

// 6. 使用策略模式构造消息数据
messageData := strategy.BuildMessageData(lastFeeding, lastFeedingTime, hoursSinceLastFeeding)

// 7. 发送订阅消息
sendReq := &dto.SendMessageRequest{
    OpenID:       collaborator.OpenID,
    TemplateType: templateType, // 动态模板类型
    Data:         messageData,
    Page:         "pages/record/feeding/feeding",
}
```

## 设计优势

### 1. 符合 SOLID 原则

- **单一职责原则 (SRP)**: 每个策略只负责一种喂养类型的消息构造
- **开闭原则 (OCP)**: 新增喂养类型无需修改现有代码，只需添加新策略
- **里氏替换原则 (LSP)**: 所有策略都实现相同接口，可相互替换
- **接口隔离原则 (ISP)**: 接口方法精简，只包含必要操作
- **依赖倒置原则 (DIP)**: 依赖抽象接口而非具体实现

### 2. 提高可扩展性

- 新增喂养类型只需实现 `FeedingReminderStrategy` 接口
- 无需修改 `CheckFeedingReminders` 核心逻辑
- 工厂自动识别并选择合适的策略

### 3. 改进代码可读性

- 消息构造逻辑集中在各自策略中
- 主流程更清晰，专注于业务逻辑
- 减少了条件判断和硬编码

### 4. 便于测试

- 每个策略可独立测试
- 工厂选择逻辑可单独验证
- 测试覆盖率更高

## 测试覆盖

### 单元测试覆盖内容

1. **策略功能测试**
   - 母乳喂养策略 (不同侧面: 左侧/右侧/两侧)
   - 奶瓶喂养策略 (不同类型: 配方奶/母乳)
   - 辅食喂养策略

2. **工厂选择测试**
   - 正确识别不同喂养类型
   - 未知类型回退到默认策略

3. **工具函数测试**
   - 时间格式化 (`formatTimeSince`)

### 测试结果

```
=== RUN   TestBreastFeedingReminderStrategy
--- PASS: TestBreastFeedingReminderStrategy (0.00s)
=== RUN   TestBottleFeedingReminderStrategy
--- PASS: TestBottleFeedingReminderStrategy (0.00s)
=== RUN   TestFoodFeedingReminderStrategy
--- PASS: TestFoodFeedingReminderStrategy (0.00s)
=== RUN   TestFeedingReminderStrategyFactory
--- PASS: TestFeedingReminderStrategyFactory (0.00s)
=== RUN   TestFormatTimeSince
--- PASS: TestFormatTimeSince (0.00s)
=== RUN   TestBreastFeedingDifferentSides
--- PASS: TestBreastFeedingDifferentSides (0.00s)
=== RUN   TestBottleFeedingDifferentTypes
--- PASS: TestBottleFeedingDifferentTypes (0.00s)
PASS
```

## 微信订阅消息模板字段映射

### 母乳喂养提醒 (`breast_feeding_reminder`)
| 字段 | 说明 | 示例值 |
|-----|------|--------|
| time1 | 上次时间 | "2025-10-26 14:30" |
| thing2 | 距离上次 | "约2小时" |
| character_string3 | 喂养时长 | "10分钟" |
| phrase4 | 喂养位置 | "左侧" / "右侧" / "两侧" |
| thing5 | 温馨提示 | "该喂奶啦，注意观察宝宝的饥饿信号" |

### 奶瓶喂养提醒 (`bottle_feeding_reminder`)
| 字段 | 说明 | 示例值 |
|-----|------|--------|
| time1 | 上次时间 | "2025-10-26 14:30" |
| thing2 | 距离上次 | "约3小时" |
| character_string3 | 喂养量 | "120ml" |
| phrase4 | 喂养类型 | "配方奶" / "母乳" |
| thing5 | 温馨提示 | "该喂奶啦，记得准备好奶瓶哦" |

### 辅食喂养提醒 (`food_feeding_reminder`)
| 字段 | 说明 | 示例值 |
|-----|------|--------|
| time1 | 上次时间 | "2025-10-26 14:30" |
| thing2 | 距离上次 | "约4小时" |
| character_string3 | 食物名称 | "米糊" |
| phrase4 | 喂养类型 | "辅食" |
| thing5 | 温馨提示 | "该给宝宝准备辅食啦，注意观察过敏反应" |

## 向后兼容性

- 未知喂养类型默认使用母乳喂养策略
- 保持原有日志记录格式
- 不影响现有数据库结构

## 未来扩展方向

1. **用户自定义提醒模板**: 允许用户自定义提醒消息内容
2. **提醒阈值配置**: 支持不同喂养类型设置不同的提醒间隔
3. **多语言支持**: 策略可根据用户语言返回不同文案
4. **提醒优先级**: 根据宝宝年龄和喂养类型调整提醒优先级

## 文件清单

### 新增文件
- `internal/application/service/feeding_reminder_strategy.go` (195 行)
- `internal/application/service/feeding_reminder_strategy_test.go` (195 行)

### 修改文件
- `internal/application/service/scheduler_service.go`:
  - 添加 `strategyFactory` 字段
  - 重构 `CheckFeedingReminders` 方法 (约 30 行修改)

## 编译和测试状态

✅ 编译通过
✅ 所有单元测试通过
✅ 代码符合 Go 规范

---

**重构日期**: 2025-10-26
**作者**: Claude Code
**版本**: v1.0
