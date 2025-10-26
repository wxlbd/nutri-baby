package entity

import (
	"time"

	"gorm.io/gorm"
)

// SubscribeRecord 订阅记录实体 (一次性订阅消息机制)
type SubscribeRecord struct {
	ID           uint   `gorm:"primarykey" json:"id"`
	OpenID       string `gorm:"column:openid;size:64;not null;index:idx_openid_type" json:"openid"`
	TemplateID   string `gorm:"column:template_id;size:128;not null" json:"templateId"`
	TemplateType string `gorm:"column:template_type;size:32;not null;index:idx_openid_type" json:"templateType"`

	// 状态: available(可用), used(已使用), expired(已过期)
	Status string `gorm:"column:status;size:16;not null;default:'available';index" json:"status"`

	AuthorizeTime time.Time  `gorm:"column:authorize_time;not null;default:CURRENT_TIMESTAMP" json:"authorizeTime"` // 授权时间
	UsedTime      *time.Time `gorm:"column:used_time" json:"usedTime,omitempty"`                                    // 使用时间
	ExpireTime    *time.Time `gorm:"column:expire_time" json:"expireTime,omitempty"`                                // 过期时间(授权后7天)

	CreatedAt time.Time      `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
	UpdatedAt time.Time      `gorm:"column:updated_at;not null;default:CURRENT_TIMESTAMP" json:"updatedAt"`
	DeletedAt gorm.DeletedAt `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (SubscribeRecord) TableName() string {
	return "subscribe_records"
}

// IsAvailable 检查授权是否可用 (一次性订阅消息机制)
func (s *SubscribeRecord) IsAvailable() bool {
	// 检查状态
	if s.Status != "available" {
		return false
	}

	// 检查是否过期(微信一次性订阅消息有效期为7天)
	if s.ExpireTime != nil && time.Now().After(*s.ExpireTime) {
		return false
	}

	return true
}

// MarkAsUsed 标记为已使用
func (s *SubscribeRecord) MarkAsUsed() {
	s.Status = "used"
	now := time.Now()
	s.UsedTime = &now
}

// MarkAsExpired 标记为已过期
func (s *SubscribeRecord) MarkAsExpired() {
	s.Status = "expired"
}

// MessageSendLog 消息发送日志实体
type MessageSendLog struct {
	ID               uint       `gorm:"primarykey" json:"id"`
	OpenID           string     `gorm:"column:openid;type:varchar(64);not null;index" json:"openid"`
	TemplateID       string     `gorm:"column:template_id;type:varchar(128);not null" json:"templateId"`
	TemplateType     string     `gorm:"column:template_type;type:varchar(32);not null;index" json:"templateType"`
	Data             string     `gorm:"column:data;type:jsonb;not null" json:"data"` // JSONB存储
	Page             string     `gorm:"column:page;type:varchar(256)" json:"page,omitempty"`
	MiniprogramState string     `gorm:"column:miniprogram_state;size:32;default:'formal'" json:"miniprogramState"`
	SendStatus       string     `gorm:"column:send_status;type:varchar(16);not null;index" json:"sendStatus"` // success/failed/pending
	ErrCode          int        `json:"errcode,omitempty"`
	ErrMsg           string     `gorm:"column:err_msg;type:text" json:"errmsg,omitempty"`
	SendTime         *time.Time `gorm:"column:send_time;index" json:"sendTime,omitempty"`
	CreatedAt        time.Time  `gorm:"column:created_at;not null;default:CURRENT_TIMESTAMP" json:"createdAt"`
}

// TableName 指定表名
func (MessageSendLog) TableName() string {
	return "message_send_logs"
}

// MessageSendQueue 消息发送队列实体
type MessageSendQueue struct {
	ID            uint      `gorm:"primarykey" json:"id"`
	OpenID        string    `gorm:"column:openid;type:varchar(64);not null;index" json:"openid"`
	TemplateID    string    `gorm:"column:template_id;type:varchar(128);not null" json:"templateId"`
	TemplateType  string    `gorm:"column:template_type;type:varchar(32);not null" json:"templateType"`
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

// TableName 指定表名
func (MessageSendQueue) TableName() string {
	return "message_send_queue"
}

// CanRetry 判断是否可以重试
func (m *MessageSendQueue) CanRetry() bool {
	return m.RetryCount < m.MaxRetry
}

// IncrementRetry 增加重试次数
func (m *MessageSendQueue) IncrementRetry() {
	m.RetryCount++
}
