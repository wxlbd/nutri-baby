package service

import (
	"context"
	"encoding/json"
	"time"

	"go.uber.org/zap"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	errs "github.com/wxlbd/nutri-baby-server/pkg/errors"
)

type SubscribeService struct {
	subscribeRepo repository.SubscribeRepository
	wechatService *WechatService
	logger        *zap.Logger
}

func NewSubscribeService(
	subscribeRepo repository.SubscribeRepository,
	wechatService *WechatService,
	logger *zap.Logger,
) *SubscribeService {
	return &SubscribeService{
		subscribeRepo: subscribeRepo,
		wechatService: wechatService,
		logger:        logger,
	}
}

// SaveSubscribeAuth 保存用户授权记录(一次性订阅消息机制)
func (s *SubscribeService) SaveSubscribeAuth(ctx context.Context, openid string, records []dto.SubscribeAuthDTO) (*dto.SubscribeAuthResponse, error) {
	successCount := 0
	failedCount := 0

	for _, r := range records {
		// 只保存用户同意的记录
		if r.Status != "accept" {
			continue
		}

		// 计算过期时间(微信一次性订阅消息有效期为7天)
		authorizeTime := time.Now()
		expireTime := authorizeTime.Add(7 * 24 * time.Hour)

		record := &entity.SubscribeRecord{
			OpenID:        openid,
			TemplateID:    r.TemplateID,
			TemplateType:  r.TemplateType,
			Status:        "available",
			AuthorizeTime: authorizeTime,
			ExpireTime:    &expireTime,
		}

		// 每次授权创建新记录(一次性消息机制)
		if err := s.subscribeRepo.CreateSubscribeRecord(ctx, record); err != nil {
			s.logger.Error("Failed to save subscribe record",
				zap.String("openid", openid),
				zap.String("templateType", r.TemplateType),
				zap.Error(err),
			)
			failedCount++
		} else {
			successCount++
			s.logger.Info("Subscribe authorization saved",
				zap.String("openid", openid),
				zap.String("templateType", r.TemplateType),
				zap.Time("expireTime", expireTime))
		}
	}

	return &dto.SubscribeAuthResponse{
		SuccessCount: successCount,
		FailedCount:  failedCount,
	}, nil
}

// GetUserSubscriptions 获取用户订阅状态
func (s *SubscribeService) GetUserSubscriptions(ctx context.Context, openid string) (*dto.SubscribeStatusResponse, error) {
	records, err := s.subscribeRepo.ListUserSubscriptions(ctx, openid)
	if err != nil {
		s.logger.Error("Failed to get user subscriptions",
			zap.String("openid", openid),
			zap.Error(err),
		)
		return nil, errs.ErrInternal
	}

	subscriptions := make([]dto.SubscriptionItem, 0, len(records))
	for _, record := range records {
		item := dto.SubscriptionItem{
			TemplateType:  record.TemplateType,
			TemplateID:    record.TemplateID,
			Status:        record.Status,
			SubscribeTime: record.AuthorizeTime.Unix(),
		}
		if record.ExpireTime != nil {
			item.ExpireTime = record.ExpireTime.Unix()
		}
		subscriptions = append(subscriptions, item)
	}

	return &dto.SubscribeStatusResponse{
		Subscriptions: subscriptions,
	}, nil
}

// CheckAuthorizationStatus 检查用户是否有可用的授权
func (s *SubscribeService) CheckAuthorizationStatus(ctx context.Context, openid, templateType string) (bool, error) {
	// TODO: 根据用户openid 和 模板ID 查询是否有可用的授权记录
	// count, err := s.subscribeRepo.CountAvailableAuthorizations(ctx, openid, templateType)
	// if err != nil {
	// 	s.logger.Error("Failed to count available authorizations",
	// 		zap.String("openid", openid),
	// 		zap.String("templateType", templateType),
	// 		zap.Error(err),
	// 	)
	// 	return false, errs.ErrInternal
	// }

	return true, nil
}

// SendSubscribeMessage 发送订阅消息(一次性消息机制)
func (s *SubscribeService) SendSubscribeMessage(
	ctx context.Context,
	req *dto.SendMessageRequest,
) error {
	s.logger.Info("📤 [SendSubscribeMessage] START - 开始发送订阅消息",
		zap.String("openid", req.OpenID),
		zap.String("page", req.Page),
		zap.Any("data", req.Data),
	)

	// 1. 查找可用的授权记录(按授权时间倒序,取最新的一条)
	s.logger.Info("🔍 [SendSubscribeMessage] STEP 1 - 查询可用授权记录",
		zap.String("openid", req.OpenID),
		zap.String("templateID", req.TemplateID),
	)

	//record, err := s.subscribeRepo.GetAvailableSubscribeRecord(ctx, req.OpenID, req.TemplateType)
	//if err != nil {
	//	s.logger.Error("❌ [SendSubscribeMessage] 查询授权记录失败",
	//		zap.String("openid", req.OpenID),
	//		zap.String("templateType", req.TemplateType),
	//		zap.Error(err),
	//	)
	//	return errs.New(4001, "查询授权记录失败")
	//}
	//
	//if record == nil {
	//	s.logger.Warn("⚠️ [SendSubscribeMessage] 未找到可用授权记录",
	//		zap.String("openid", req.OpenID),
	//		zap.String("templateType", req.TemplateType),
	//	)
	//	return errs.New(4001, "用户未授权或授权已使用")
	//}
	//
	//s.logger.Info("✅ [SendSubscribeMessage] 找到可用授权记录",
	//	zap.String("openid", req.OpenID),
	//	zap.String("templateType", req.TemplateType),
	//	zap.String("templateID", record.TemplateID),
	//	zap.String("status", record.Status),
	//	zap.Time("authorizeTime", record.AuthorizeTime),
	//	zap.Timep("expireTime", record.ExpireTime),
	//)
	//
	//// 2. 检查授权是否可用
	//s.logger.Info("🔍 [SendSubscribeMessage] STEP 2 - 检查授权是否可用",
	//	zap.String("status", record.Status),
	//)
	//
	//if !record.IsAvailable() {
	//	s.logger.Warn("⚠️ [SendSubscribeMessage] 授权不可用",
	//		zap.String("openid", req.OpenID),
	//		zap.String("templateType", req.TemplateType),
	//		zap.String("status", record.Status),
	//	)
	//	return errs.New(4002, "授权已失效")
	//}

	s.logger.Info("✅ [SendSubscribeMessage] 授权可用,准备调用微信API")

	// 3. 调用微信API发送
	s.logger.Info("📞 [SendSubscribeMessage] STEP 3 - 调用微信API发送订阅消息",
		zap.String("openid", req.OpenID),
		//zap.String("templateID", record.TemplateID),
		zap.String("page", req.Page),
		zap.Any("data", req.Data),
	)

	err := s.wechatService.SendSubscribeMessage(
		req.OpenID,
		req.TemplateID,
		req.Data,
		req.Page,
		"developer",
	)

	// 4. 标记授权为已使用(无论发送成功或失败,授权都会被消耗)
	s.logger.Info("🔄 [SendSubscribeMessage] STEP 4 - 标记授权为已使用",
		zap.String("openid", req.OpenID),
		zap.String("templateID", req.TemplateID),
	)

	//record.MarkAsUsed()
	//updateErr := s.subscribeRepo.UpdateSubscribeRecord(ctx, record)
	//if updateErr != nil {
	//	s.logger.Error("❌ [SendSubscribeMessage] 更新授权状态失败",
	//		zap.String("openid", req.OpenID),
	//		zap.Error(updateErr),
	//	)
	//} else {
	//	s.logger.Info("✅ [SendSubscribeMessage] 授权状态已更新为已使用")
	//}

	// 5. 记录发送日志
	s.logger.Info("📝 [SendSubscribeMessage] STEP 5 - 保存发送日志")

	dataJSON, _ := json.Marshal(req.Data)
	log := &entity.MessageSendLog{
		OpenID:           req.OpenID,
		TemplateID:       req.TemplateID,
		Data:             string(dataJSON),
		Page:             req.Page,
		MiniprogramState: "formal",
	}

	now := time.Now()
	if err != nil {
		log.SendStatus = "failed"
		log.ErrMsg = err.Error()
		s.logger.Error("❌ [SendSubscribeMessage] 发送订阅消息失败",
			zap.String("openid", req.OpenID),
			zap.String("templateID", req.TemplateID),
			//zap.String("templateID", record.TemplateID),
			zap.Error(err),
		)
	} else {
		log.SendStatus = "success"
		log.SendTime = &now
		s.logger.Info("✅ [SendSubscribeMessage] 订阅消息发送成功",
			zap.String("openid", req.OpenID),
			zap.String("templateID", req.TemplateID),
			//zap.String("templateID", record.TemplateID),
		)
	}

	logErr := s.subscribeRepo.CreateSendLog(ctx, log)
	if logErr != nil {
		s.logger.Error("❌ [SendSubscribeMessage] 保存发送日志失败",
			zap.Error(logErr),
		)
	} else {
		s.logger.Info("✅ [SendSubscribeMessage] 发送日志已保存")
	}

	s.logger.Info("🏁 [SendSubscribeMessage] END - 订阅消息发送流程结束",
		zap.String("openid", req.OpenID),
		zap.String("templateID", req.TemplateID),
		zap.Bool("success", err == nil),
	)

	return err
}

// GetMessageLogs 获取消息发送日志
func (s *SubscribeService) GetMessageLogs(ctx context.Context, openid string, offset, limit int) (*dto.MessageLogsResponse, error) {
	logs, total, err := s.subscribeRepo.GetSendLogs(ctx, openid, offset, limit)
	if err != nil {
		s.logger.Error("Failed to get message logs",
			zap.String("openid", openid),
			zap.Error(err),
		)
		return nil, errs.ErrInternal
	}

	items := make([]dto.MessageLogItem, 0, len(logs))
	for _, log := range logs {
		item := dto.MessageLogItem{
			ID:         log.ID,
			SendStatus: log.SendStatus,
			ErrMsg:     log.ErrMsg,
			CreatedAt:  log.CreatedAt.Unix(),
		}
		if log.SendTime != nil {
			item.SendTime = log.SendTime.Unix()
		}
		items = append(items, item)
	}

	return &dto.MessageLogsResponse{
		Logs:  items,
		Total: total,
	}, nil
}

// CleanExpiredRecords 清理过期的授权记录(定时任务调用)
func (s *SubscribeService) CleanExpiredRecords(ctx context.Context) error {
	// 清理7天前过期的记录
	beforeTime := time.Now().Add(-7 * 24 * time.Hour)

	count, err := s.subscribeRepo.DeleteExpiredRecords(ctx, beforeTime)
	if err != nil {
		s.logger.Error("Failed to clean expired records", zap.Error(err))
		return err
	}

	if count > 0 {
		s.logger.Info("Cleaned expired authorization records", zap.Int64("count", count))
	}

	return nil
}
