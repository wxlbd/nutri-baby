package repository

import (
	"context"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"time"
)

// SubscribeRepository 订阅消息仓储接口
type SubscribeRepository interface {
	// ==================== 一次性订阅授权管理 ====================

	// CreateSubscribeRecord 创建订阅授权记录(每次授权创建新记录)
	CreateSubscribeRecord(ctx context.Context, record *entity.SubscribeRecord) error

	// GetAvailableSubscribeRecord 获取用户可用的授权记录(按授权时间倒序,取最新的一条)
	GetAvailableSubscribeRecord(ctx context.Context, openid, templateType string) (*entity.SubscribeRecord, error)

	// GetSubscribeRecord 根据openid和模板类型获取最新的订阅记录
	GetSubscribeRecord(ctx context.Context, openid, templateType string) (*entity.SubscribeRecord, error)

	// ListUserSubscriptions 获取用户的所有订阅记录(包括已使用和已过期)
	ListUserSubscriptions(ctx context.Context, openid string) ([]*entity.SubscribeRecord, error)

	// UpdateSubscribeRecord 更新订阅记录
	UpdateSubscribeRecord(ctx context.Context, record *entity.SubscribeRecord) error

	// CountAvailableAuthorizations 统计用户可用的授权数量
	CountAvailableAuthorizations(ctx context.Context, openid, templateType string) (int64, error)

	// DeleteExpiredRecords 清理过期的授权记录(定期任务使用)
	DeleteExpiredRecords(ctx context.Context, beforeTime time.Time) (int64, error)

	// ==================== 消息发送队列管理(暂不使用) ====================

	// AddToSendQueue 将消息加入发送队列
	AddToSendQueue(ctx context.Context, queue *entity.MessageSendQueue) error

	// GetPendingMessages 获取待发送的消息(按计划时间排序)
	GetPendingMessages(ctx context.Context, limit int) ([]*entity.MessageSendQueue, error)

	// UpdateQueueStatus 更新队列消息状态
	UpdateQueueStatus(ctx context.Context, id uint, status string, errorMsg string) error

	// IncrementRetryCount 增加重试次数
	IncrementRetryCount(ctx context.Context, id uint) error

	// DeleteQueueMessage 删除队列消息
	DeleteQueueMessage(ctx context.Context, id uint) error

	// ==================== 消息发送日志 ====================

	// CreateSendLog 创建发送日志
	CreateSendLog(ctx context.Context, log *entity.MessageSendLog) error

	// GetSendLogs 获取发送日志(分页)
	GetSendLogs(ctx context.Context, openid string, offset, limit int) ([]*entity.MessageSendLog, int64, error)

	// GetSendLogsByTemplateType 根据模板类型获取发送日志
	GetSendLogsByTemplateType(ctx context.Context, templateType string, offset, limit int) ([]*entity.MessageSendLog, int64, error)

	// GetRecentFailedLogs 获取最近失败的发送日志
	GetRecentFailedLogs(ctx context.Context, hours int, limit int) ([]*entity.MessageSendLog, error)
}
