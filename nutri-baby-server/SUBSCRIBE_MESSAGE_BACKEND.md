# 订阅消息后端实现方案

本文档说明如何实现订阅消息的后端服务,包括授权记录管理、消息发送和定时任务。

## 📋 目录

1. [数据库设计](#数据库设计)
2. [API 接口设计](#api-接口设计)
3. [后端服务实现](#后端服务实现)
4. [微信API对接](#微信api对接)
5. [定时任务设计](#定时任务设计)
6. [前端对接改造](#前端对接改造)

---

## 1. 数据库设计

### 1.1 订阅记录表 (subscribe_records)

存储用户的订阅授权记录。

```sql
CREATE TABLE subscribe_records (
    id BIGSERIAL PRIMARY KEY,
    openid VARCHAR(64) NOT NULL,                    -- 用户openid
    template_id VARCHAR(128) NOT NULL,              -- 微信模板ID
    template_type VARCHAR(32) NOT NULL,             -- 模板类型(vaccine_reminder等)
    status VARCHAR(16) NOT NULL DEFAULT 'active',   -- 状态: active/inactive/expired
    subscribe_time TIMESTAMP NOT NULL DEFAULT NOW(),-- 订阅时间
    expire_time TIMESTAMP,                          -- 过期时间(微信订阅有效期)
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,                           -- 软删除

    UNIQUE(openid, template_id),
    INDEX idx_openid (openid),
    INDEX idx_template_type (template_type),
    INDEX idx_status (status)
);

COMMENT ON TABLE subscribe_records IS '订阅消息授权记录表';
COMMENT ON COLUMN subscribe_records.template_type IS '模板类型: vaccine_reminder, breast_feeding_reminder等';
COMMENT ON COLUMN subscribe_records.status IS '状态: active-有效, inactive-已取消, expired-已过期';
```

### 1.2 消息发送记录表 (message_send_logs)

存储消息发送历史,用于追踪和调试。

```sql
CREATE TABLE message_send_logs (
    id BIGSERIAL PRIMARY KEY,
    openid VARCHAR(64) NOT NULL,                    -- 接收用户openid
    template_id VARCHAR(128) NOT NULL,              -- 微信模板ID
    template_type VARCHAR(32) NOT NULL,             -- 模板类型
    data JSONB NOT NULL,                            -- 消息数据(模板字段)
    page VARCHAR(256),                              -- 跳转页面路径
    miniprogram_state VARCHAR(32) DEFAULT 'formal', -- 小程序状态: developer/trial/formal
    send_status VARCHAR(16) NOT NULL,               -- 发送状态: success/failed/pending
    errcode INTEGER,                                -- 微信错误码
    errmsg TEXT,                                    -- 错误信息
    send_time TIMESTAMP,                            -- 实际发送时间
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),

    INDEX idx_openid (openid),
    INDEX idx_template_type (template_type),
    INDEX idx_send_status (send_status),
    INDEX idx_send_time (send_time)
);

COMMENT ON TABLE message_send_logs IS '订阅消息发送记录表';
COMMENT ON COLUMN message_send_logs.send_status IS '发送状态: success-成功, failed-失败, pending-待发送';
```

### 1.3 消息发送队列表 (message_send_queue)

待发送的消息队列,用于异步处理和重试。

```sql
CREATE TABLE message_send_queue (
    id BIGSERIAL PRIMARY KEY,
    openid VARCHAR(64) NOT NULL,                    -- 接收用户openid
    template_id VARCHAR(128) NOT NULL,              -- 微信模板ID
    template_type VARCHAR(32) NOT NULL,             -- 模板类型
    data JSONB NOT NULL,                            -- 消息数据
    page VARCHAR(256),                              -- 跳转页面路径
    scheduled_time TIMESTAMP NOT NULL,              -- 计划发送时间
    retry_count INTEGER NOT NULL DEFAULT 0,         -- 重试次数
    max_retry INTEGER NOT NULL DEFAULT 3,           -- 最大重试次数
    status VARCHAR(16) NOT NULL DEFAULT 'pending',  -- 状态: pending/processing/sent/failed
    error_msg TEXT,                                 -- 错误信息
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),

    INDEX idx_openid (openid),
    INDEX idx_scheduled_time (scheduled_time),
    INDEX idx_status (status)
);

COMMENT ON TABLE message_send_queue IS '订阅消息发送队列表';
COMMENT ON COLUMN message_send_queue.status IS '状态: pending-待发送, processing-处理中, sent-已发送, failed-失败';
```

---

## 2. API 接口设计

### 2.1 上传订阅授权记录

用户授权后,前端调用此接口上传授权结果。

**接口**: `POST /api/v1/subscribe/auth`

**请求 Headers**:
```
Authorization: Bearer {token}
```

**请求 Body**:
```json
{
  "records": [
    {
      "templateId": "J6RbROH-yhNdgj2FPwrz4FnzzpITH2KcHV5h9qjcVbI",
      "templateType": "vaccine_reminder",
      "status": "accept"  // accept 或 reject
    }
  ]
}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "successCount": 1,
    "failedCount": 0
  }
}
```

### 2.2 获取用户订阅状态

前端查询用户当前的订阅状态。

**接口**: `GET /api/v1/subscribe/status`

**请求 Headers**:
```
Authorization: Bearer {token}
```

**响应**:
```json
{
  "code": 0,
  "message": "success",
  "data": {
    "subscriptions": [
      {
        "templateType": "vaccine_reminder",
        "status": "active",
        "subscribeTime": 1698825600000,
        "expireTime": 1701417600000
      }
    ]
  }
}
```

### 2.3 取消订阅

用户主动取消订阅。

**接口**: `DELETE /api/v1/subscribe/cancel`

**请求 Headers**:
```
Authorization: Bearer {token}
```

**请求 Body**:
```json
{
  "templateType": "vaccine_reminder"
}
```

**响应**:
```json
{
  "code": 0,
  "message": "订阅已取消"
}
```

### 2.4 (内部接口) 发送订阅消息

供后端服务调用,不对外暴露。

**方法**: `SendSubscribeMessage(openid, templateType, data, page)`

---

## 3. 后端服务实现

### 3.1 实体定义 (Entity)

**文件**: `internal/domain/entity/subscribe.go`

```go
package entity

import (
    "time"
    "gorm.io/gorm"
)

// SubscribeRecord 订阅记录实体
type SubscribeRecord struct {
    ID            uint           `gorm:"primarykey" json:"id"`
    OpenID        string         `gorm:"size:64;not null;index" json:"openid"`
    TemplateID    string         `gorm:"size:128;not null" json:"templateId"`
    TemplateType  string         `gorm:"size:32;not null;index" json:"templateType"`
    Status        string         `gorm:"size:16;not null;default:'active';index" json:"status"` // active/inactive/expired
    SubscribeTime time.Time      `gorm:"not null;default:CURRENT_TIMESTAMP" json:"subscribeTime"`
    ExpireTime    *time.Time     `json:"expireTime,omitempty"`
    CreatedAt     time.Time      `json:"createdAt"`
    UpdatedAt     time.Time      `json:"updatedAt"`
    DeletedAt     gorm.DeletedAt `gorm:"index" json:"-"`
}

func (SubscribeRecord) TableName() string {
    return "subscribe_records"
}

// MessageSendLog 消息发送日志实体
type MessageSendLog struct {
    ID               uint      `gorm:"primarykey" json:"id"`
    OpenID           string    `gorm:"size:64;not null;index" json:"openid"`
    TemplateID       string    `gorm:"size:128;not null" json:"templateId"`
    TemplateType     string    `gorm:"size:32;not null;index" json:"templateType"`
    Data             string    `gorm:"type:jsonb;not null" json:"data"` // JSONB存储
    Page             string    `gorm:"size:256" json:"page,omitempty"`
    MiniprogramState string    `gorm:"size:32;default:'formal'" json:"miniprogramState"`
    SendStatus       string    `gorm:"size:16;not null;index" json:"sendStatus"` // success/failed/pending
    ErrCode          int       `json:"errcode,omitempty"`
    ErrMsg           string    `gorm:"type:text" json:"errmsg,omitempty"`
    SendTime         *time.Time `gorm:"index" json:"sendTime,omitempty"`
    CreatedAt        time.Time `json:"createdAt"`
}

func (MessageSendLog) TableName() string {
    return "message_send_logs"
}

// MessageSendQueue 消息发送队列实体
type MessageSendQueue struct {
    ID            uint      `gorm:"primarykey" json:"id"`
    OpenID        string    `gorm:"size:64;not null;index" json:"openid"`
    TemplateID    string    `gorm:"size:128;not null" json:"templateId"`
    TemplateType  string    `gorm:"size:32;not null" json:"templateType"`
    Data          string    `gorm:"type:jsonb;not null" json:"data"`
    Page          string    `gorm:"size:256" json:"page,omitempty"`
    ScheduledTime time.Time `gorm:"not null;index" json:"scheduledTime"`
    RetryCount    int       `gorm:"not null;default:0" json:"retryCount"`
    MaxRetry      int       `gorm:"not null;default:3" json:"maxRetry"`
    Status        string    `gorm:"size:16;not null;default:'pending';index" json:"status"` // pending/processing/sent/failed
    ErrorMsg      string    `gorm:"type:text" json:"errorMsg,omitempty"`
    CreatedAt     time.Time `json:"createdAt"`
    UpdatedAt     time.Time `json:"updatedAt"`
}

func (MessageSendQueue) TableName() string {
    return "message_send_queue"
}
```

### 3.2 仓储接口 (Repository)

**文件**: `internal/domain/repository/subscribe_repository.go`

```go
package repository

import (
    "context"
    "nutri-baby-server/internal/domain/entity"
)

type SubscribeRepository interface {
    // 订阅记录管理
    CreateSubscribeRecord(ctx context.Context, record *entity.SubscribeRecord) error
    GetSubscribeRecord(ctx context.Context, openid, templateType string) (*entity.SubscribeRecord, error)
    ListUserSubscriptions(ctx context.Context, openid string) ([]*entity.SubscribeRecord, error)
    UpdateSubscribeStatus(ctx context.Context, openid, templateType, status string) error
    DeleteSubscribeRecord(ctx context.Context, openid, templateType string) error

    // 消息发送队列管理
    AddToSendQueue(ctx context.Context, queue *entity.MessageSendQueue) error
    GetPendingMessages(ctx context.Context, limit int) ([]*entity.MessageSendQueue, error)
    UpdateQueueStatus(ctx context.Context, id uint, status string, errorMsg string) error
    IncrementRetryCount(ctx context.Context, id uint) error

    // 消息发送日志
    CreateSendLog(ctx context.Context, log *entity.MessageSendLog) error
    GetSendLogs(ctx context.Context, openid string, offset, limit int) ([]*entity.MessageSendLog, int64, error)
}
```

### 3.3 应用服务 (Service)

**文件**: `internal/application/service/subscribe_service.go`

```go
package service

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "nutri-baby-server/internal/domain/entity"
    "nutri-baby-server/internal/domain/repository"
    "nutri-baby-server/pkg/errors"
)

type SubscribeService struct {
    subscribeRepo repository.SubscribeRepository
    wechatService *WechatService
}

func NewSubscribeService(
    subscribeRepo repository.SubscribeRepository,
    wechatService *WechatService,
) *SubscribeService {
    return &SubscribeService{
        subscribeRepo: subscribeRepo,
        wechatService: wechatService,
    }
}

// SaveSubscribeAuth 保存用户授权记录
func (s *SubscribeService) SaveSubscribeAuth(ctx context.Context, openid string, records []SubscribeAuthDTO) error {
    for _, r := range records {
        if r.Status != "accept" {
            continue // 只保存用户同意的记录
        }

        record := &entity.SubscribeRecord{
            OpenID:        openid,
            TemplateID:    r.TemplateID,
            TemplateType:  r.TemplateType,
            Status:        "active",
            SubscribeTime: time.Now(),
            ExpireTime:    calculateExpireTime(), // 微信订阅消息有效期通常为30天
        }

        // 先尝试查询,如果存在则更新
        existing, err := s.subscribeRepo.GetSubscribeRecord(ctx, openid, r.TemplateType)
        if err == nil && existing != nil {
            // 更新已有记录
            if err := s.subscribeRepo.UpdateSubscribeStatus(ctx, openid, r.TemplateType, "active"); err != nil {
                return err
            }
        } else {
            // 创建新记录
            if err := s.subscribeRepo.CreateSubscribeRecord(ctx, record); err != nil {
                return err
            }
        }
    }

    return nil
}

// GetUserSubscriptions 获取用户订阅状态
func (s *SubscribeService) GetUserSubscriptions(ctx context.Context, openid string) ([]*entity.SubscribeRecord, error) {
    return s.subscribeRepo.ListUserSubscriptions(ctx, openid)
}

// CancelSubscription 取消订阅
func (s *SubscribeService) CancelSubscription(ctx context.Context, openid, templateType string) error {
    return s.subscribeRepo.UpdateSubscribeStatus(ctx, openid, templateType, "inactive")
}

// QueueSubscribeMessage 将消息加入发送队列
func (s *SubscribeService) QueueSubscribeMessage(
    ctx context.Context,
    openid string,
    templateType string,
    data map[string]interface{},
    page string,
    scheduledTime time.Time,
) error {
    // 1. 检查用户是否订阅
    record, err := s.subscribeRepo.GetSubscribeRecord(ctx, openid, templateType)
    if err != nil || record == nil || record.Status != "active" {
        return errors.ErrSubscriptionNotFound
    }

    // 2. 检查是否过期
    if record.ExpireTime != nil && time.Now().After(*record.ExpireTime) {
        s.subscribeRepo.UpdateSubscribeStatus(ctx, openid, templateType, "expired")
        return errors.ErrSubscriptionExpired
    }

    // 3. 序列化数据
    dataJSON, err := json.Marshal(data)
    if err != nil {
        return err
    }

    // 4. 加入队列
    queue := &entity.MessageSendQueue{
        OpenID:        openid,
        TemplateID:    record.TemplateID,
        TemplateType:  templateType,
        Data:          string(dataJSON),
        Page:          page,
        ScheduledTime: scheduledTime,
        Status:        "pending",
    }

    return s.subscribeRepo.AddToSendQueue(ctx, queue)
}

// SendSubscribeMessage 立即发送订阅消息
func (s *SubscribeService) SendSubscribeMessage(
    ctx context.Context,
    openid string,
    templateType string,
    data map[string]interface{},
    page string,
) error {
    // 1. 检查订阅状态
    record, err := s.subscribeRepo.GetSubscribeRecord(ctx, openid, templateType)
    if err != nil || record == nil || record.Status != "active" {
        return errors.ErrSubscriptionNotFound
    }

    // 2. 调用微信API发送
    result, err := s.wechatService.SendSubscribeMessage(
        openid,
        record.TemplateID,
        data,
        page,
        "formal",
    )

    // 3. 记录发送日志
    dataJSON, _ := json.Marshal(data)
    log := &entity.MessageSendLog{
        OpenID:           openid,
        TemplateID:       record.TemplateID,
        TemplateType:     templateType,
        Data:             string(dataJSON),
        Page:             page,
        MiniprogramState: "formal",
    }

    now := time.Now()
    if err != nil {
        log.SendStatus = "failed"
        log.ErrMsg = err.Error()
    } else {
        log.SendStatus = "success"
        log.ErrCode = result.ErrCode
        log.ErrMsg = result.ErrMsg
        log.SendTime = &now
    }

    s.subscribeRepo.CreateSendLog(ctx, log)

    return err
}

// calculateExpireTime 计算订阅过期时间(30天后)
func calculateExpireTime() *time.Time {
    expireTime := time.Now().Add(30 * 24 * time.Hour)
    return &expireTime
}
```

### 3.4 DTO 定义

**文件**: `internal/application/dto/subscribe_dto.go`

```go
package dto

// SubscribeAuthDTO 订阅授权请求
type SubscribeAuthDTO struct {
    TemplateID   string `json:"templateId" binding:"required"`
    TemplateType string `json:"templateType" binding:"required"`
    Status       string `json:"status" binding:"required,oneof=accept reject"`
}

// SubscribeAuthRequest 批量上传授权记录请求
type SubscribeAuthRequest struct {
    Records []SubscribeAuthDTO `json:"records" binding:"required,min=1"`
}

// SubscribeStatusResponse 订阅状态响应
type SubscribeStatusResponse struct {
    Subscriptions []SubscriptionItem `json:"subscriptions"`
}

type SubscriptionItem struct {
    TemplateType  string `json:"templateType"`
    Status        string `json:"status"`
    SubscribeTime int64  `json:"subscribeTime"`
    ExpireTime    int64  `json:"expireTime,omitempty"`
}

// CancelSubscriptionRequest 取消订阅请求
type CancelSubscriptionRequest struct {
    TemplateType string `json:"templateType" binding:"required"`
}
```

---

## 4. 微信API对接

### 4.1 微信服务实现

**文件**: `internal/application/service/wechat_service.go` (扩展)

```go
// SendSubscribeMessage 发送订阅消息
func (s *WechatService) SendSubscribeMessage(
    openid string,
    templateID string,
    data map[string]interface{},
    page string,
    miniprogramState string,
) (*WechatSubscribeMessageResult, error) {
    // 1. 获取 access_token
    accessToken, err := s.getAccessToken()
    if err != nil {
        return nil, err
    }

    // 2. 构建请求体
    requestBody := map[string]interface{}{
        "touser":            openid,
        "template_id":       templateID,
        "page":              page,
        "miniprogram_state": miniprogramState,
        "lang":              "zh_CN",
        "data":              formatTemplateData(data),
    }

    // 3. 调用微信API
    url := fmt.Sprintf("https://api.weixin.qq.com/cgi-bin/message/subscribe/send?access_token=%s", accessToken)

    resp, err := httpPost(url, requestBody)
    if err != nil {
        return nil, err
    }

    var result WechatSubscribeMessageResult
    if err := json.Unmarshal(resp, &result); err != nil {
        return nil, err
    }

    if result.ErrCode != 0 {
        return &result, fmt.Errorf("wechat api error: %d - %s", result.ErrCode, result.ErrMsg)
    }

    return &result, nil
}

// formatTemplateData 格式化模板数据为微信要求的格式
// 输入: {"name": "张三", "time": "2025-10-24"}
// 输出: {"name": {"value": "张三"}, "time": {"value": "2025-10-24"}}
func formatTemplateData(data map[string]interface{}) map[string]interface{} {
    formatted := make(map[string]interface{})
    for k, v := range data {
        formatted[k] = map[string]interface{}{
            "value": v,
        }
    }
    return formatted
}

type WechatSubscribeMessageResult struct {
    ErrCode int    `json:"errcode"`
    ErrMsg  string `json:"errmsg"`
}
```

---

## 5. 定时任务设计

### 5.1 疫苗提醒定时任务

**文件**: `internal/application/job/vaccine_reminder_job.go`

```go
package job

import (
    "context"
    "fmt"
    "time"

    "nutri-baby-server/internal/application/service"
    "nutri-baby-server/internal/domain/repository"
)

type VaccineReminderJob struct {
    vaccineRepo   repository.VaccineRepository
    subscribeServ *service.SubscribeService
}

func NewVaccineReminderJob(
    vaccineRepo repository.VaccineRepository,
    subscribeServ *service.SubscribeService,
) *VaccineReminderJob {
    return &VaccineReminderJob{
        vaccineRepo:   vaccineRepo,
        subscribeServ: subscribeServ,
    }
}

// Run 每天凌晨1点执行
func (j *VaccineReminderJob) Run() {
    ctx := context.Background()

    // 1. 查询3天后需要接种的疫苗提醒
    targetDate := time.Now().Add(3 * 24 * time.Hour)
    reminders, err := j.vaccineRepo.GetRemindersByDate(ctx, targetDate)
    if err != nil {
        fmt.Printf("Failed to get vaccine reminders: %v\n", err)
        return
    }

    // 2. 遍历提醒,发送订阅消息
    for _, reminder := range reminders {
        baby, err := j.vaccineRepo.GetBabyByID(ctx, reminder.BabyID)
        if err != nil {
            continue
        }

        plan, err := j.vaccineRepo.GetPlanByID(ctx, reminder.PlanID)
        if err != nil {
            continue
        }

        // 3. 构建消息数据
        data := map[string]interface{}{
            "thing1": baby.Name,                              // 宝宝名称
            "thing2": plan.VaccineName,                       // 疫苗名称
            "time3":  reminder.ScheduledDate.Format("2006-01-02 15:04"), // 接种时间
            "thing4": "请携带预防接种证前往",                    // 接种地址/提示
            "number5": plan.DoseNumber,                       // 接种针数
        }

        // 4. 发送给宝宝的所有关联用户
        collaborators, err := j.vaccineRepo.GetBabyCollaborators(ctx, baby.BabyID)
        if err != nil {
            continue
        }

        for _, collaborator := range collaborators {
            page := fmt.Sprintf("/pages/vaccine/vaccine?babyId=%s", baby.BabyID)

            // 加入发送队列
            j.subscribeServ.QueueSubscribeMessage(
                ctx,
                collaborator.OpenID,
                "vaccine_reminder",
                data,
                page,
                time.Now(),
            )
        }

        // 5. 标记提醒已发送
        j.vaccineRepo.MarkReminderSent(ctx, reminder.ID)
    }
}
```

### 5.2 消息队列处理器

**文件**: `internal/application/job/message_queue_processor.go`

```go
package job

import (
    "context"
    "encoding/json"
    "fmt"
    "time"

    "nutri-baby-server/internal/application/service"
    "nutri-baby-server/internal/domain/repository"
)

type MessageQueueProcessor struct {
    subscribeRepo repository.SubscribeRepository
    subscribeServ *service.SubscribeService
}

func NewMessageQueueProcessor(
    subscribeRepo repository.SubscribeRepository,
    subscribeServ *service.SubscribeService,
) *MessageQueueProcessor {
    return &MessageQueueProcessor{
        subscribeRepo: subscribeRepo,
        subscribeServ: subscribeServ,
    }
}

// Run 每分钟执行一次,处理待发送的消息
func (j *MessageQueueProcessor) Run() {
    ctx := context.Background()

    // 1. 获取待发送的消息(限制100条)
    messages, err := j.subscribeRepo.GetPendingMessages(ctx, 100)
    if err != nil {
        fmt.Printf("Failed to get pending messages: %v\n", err)
        return
    }

    // 2. 逐条发送
    for _, msg := range messages {
        // 检查是否到达发送时间
        if time.Now().Before(msg.ScheduledTime) {
            continue
        }

        // 更新状态为处理中
        j.subscribeRepo.UpdateQueueStatus(ctx, msg.ID, "processing", "")

        // 解析数据
        var data map[string]interface{}
        if err := json.Unmarshal([]byte(msg.Data), &data); err != nil {
            j.subscribeRepo.UpdateQueueStatus(ctx, msg.ID, "failed", fmt.Sprintf("Invalid data: %v", err))
            continue
        }

        // 发送消息
        err := j.subscribeServ.SendSubscribeMessage(
            ctx,
            msg.OpenID,
            msg.TemplateType,
            data,
            msg.Page,
        )

        if err != nil {
            // 发送失败,检查是否需要重试
            if msg.RetryCount < msg.MaxRetry {
                j.subscribeRepo.IncrementRetryCount(ctx, msg.ID)
                j.subscribeRepo.UpdateQueueStatus(ctx, msg.ID, "pending", fmt.Sprintf("Retry %d: %v", msg.RetryCount+1, err))
            } else {
                j.subscribeRepo.UpdateQueueStatus(ctx, msg.ID, "failed", fmt.Sprintf("Max retry exceeded: %v", err))
            }
        } else {
            // 发送成功
            j.subscribeRepo.UpdateQueueStatus(ctx, msg.ID, "sent", "")
        }
    }
}
```

### 5.3 注册定时任务

**文件**: `cmd/server/main.go` (扩展)

```go
import (
    "github.com/robfig/cron/v3"
    "nutri-baby-server/internal/application/job"
)

func main() {
    // ... 初始化应用 ...

    // 创建定时任务调度器
    c := cron.New()

    // 注册疫苗提醒任务(每天凌晨1点执行)
    vaccineJob := job.NewVaccineReminderJob(/* 依赖注入 */)
    c.AddFunc("0 1 * * *", vaccineJob.Run)

    // 注册消息队列处理器(每分钟执行)
    queueProcessor := job.NewMessageQueueProcessor(/* 依赖注入 */)
    c.AddFunc("* * * * *", queueProcessor.Run)

    // 启动定时任务
    c.Start()
    defer c.Stop()

    // ... 启动HTTP服务器 ...
}
```

---

## 6. 前端对接改造

### 6.1 前端Store改造

修改 `src/store/subscribe.ts`,在授权成功后调用后端API。

```typescript
/**
 * 请求订阅消息授权(改造版)
 */
export async function requestSubscribeMessage(
  types: SubscribeMessageType[]
): Promise<Map<SubscribeMessageType, 'accept' | 'reject'>> {
  initializeIfNeeded()

  const templateIds = types.map((type) => {
    const config = getTemplateConfig(type)
    if (!config) {
      throw new Error(`未找到模板配置: ${type}`)
    }
    return config.templateId
  })

  return new Promise((resolve) => {
    uni.requestSubscribeMessage({
      tmplIds: templateIds,
      success: async (res) => {
        console.log('[Subscribe] requestSubscribeMessage success:', res)

        const results = new Map<SubscribeMessageType, 'accept' | 'reject'>()
        const acceptedRecords: Array<{ templateId: string; templateType: string; status: string }> = []

        types.forEach((type, index) => {
          const templateId = templateIds[index]
          const status = res[templateId]

          let authStatus: 'accept' | 'reject' = 'reject'
          if (status === 'accept') {
            authStatus = 'accept'

            // 收集授权成功的记录
            acceptedRecords.push({
              templateId,
              templateType: type,
              status: 'accept'
            })
          }

          results.set(type, authStatus)
          updateAuthRecord(type, authStatus)
        })

        // ⭐ 新增:上传授权记录到后端
        if (acceptedRecords.length > 0) {
          try {
            await uploadSubscribeAuth(acceptedRecords)
          } catch (error) {
            console.error('[Subscribe] 上传授权记录失败:', error)
            // 上传失败不影响前端流程,仅打印日志
          }
        }

        resolve(results)
      },
      fail: (err) => {
        console.error('[Subscribe] requestSubscribeMessage fail:', err)

        const results = new Map<SubscribeMessageType, 'accept' | 'reject'>()
        types.forEach((type) => {
          results.set(type, 'reject')
          updateAuthRecord(type, 'reject')
        })

        resolve(results)
      },
    })
  })
}

/**
 * 上传订阅授权记录到后端
 */
async function uploadSubscribeAuth(records: Array<{ templateId: string; templateType: string; status: string }>) {
  const { post } = await import('@/utils/request')

  const response = await post('/api/v1/subscribe/auth', {
    records
  })

  if (response.code !== 0) {
    throw new Error(response.message || '上传授权记录失败')
  }

  console.log('[Subscribe] 授权记录上传成功:', response.data)
}
```

---

## 7. 部署和测试

### 7.1 环境变量配置

**文件**: `config/config.yaml`

```yaml
wechat:
  app_id: "your_wechat_app_id"
  app_secret: "your_wechat_app_secret"

subscribe:
  enabled: true                # 是否启用订阅消息功能
  queue_batch_size: 100       # 队列处理批次大小
  max_retry: 3                # 最大重试次数
  miniprogram_state: "formal" # 小程序状态: developer/trial/formal
```

### 7.2 数据库迁移

**文件**: `migrations/003_subscribe_message.sql`

```sql
-- 创建订阅相关表
-- (将上述1.数据库设计中的SQL放在这里)
```

运行迁移:
```bash
make migrate-up
```

### 7.3 测试流程

1. **前端测试授权**:
   - 触发订阅引导
   - 同意授权
   - 检查浏览器控制台是否有"授权记录上传成功"日志

2. **后端测试接口**:
   ```bash
   # 查询订阅状态
   curl -H "Authorization: Bearer {token}" \
     http://localhost:8080/api/v1/subscribe/status
   ```

3. **测试定时任务**:
   - 手动触发疫苗提醒任务
   - 检查消息发送日志表

4. **微信消息接收**:
   - 在微信小程序中查看订阅消息

---

## 8. 常见问题

### Q1: 微信订阅消息有效期多久?
**A**: 通常为30天,具体以微信官方文档为准。过期后需要用户重新授权。

### Q2: 发送失败如何处理?
**A**: 系统会自动重试3次,3次失败后标记为失败状态,可在管理后台查看失败原因。

### Q3: 如何处理用户取消订阅?
**A**: 用户在微信端取消订阅后,下次发送会返回错误码,系统应自动更新订阅状态为inactive。

### Q4: 订阅消息发送频率限制?
**A**: 微信对订阅消息有频率限制,建议合理控制发送频率,避免被限流。

---

## 9. 总结

本方案提供了完整的订阅消息后端实现,包括:

✅ 数据库设计(3张表)
✅ API接口设计(3个接口)
✅ 后端服务实现(Entity/Repository/Service)
✅ 微信API对接(发送订阅消息)
✅ 定时任务设计(疫苗提醒/队列处理)
✅ 前端对接改造(上传授权记录)

后续可根据实际需求进行扩展和优化。
