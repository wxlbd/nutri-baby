package service

import (
	"github.com/silenceper/wechat/v2/miniprogram/subscribe"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/wechat"
	"go.uber.org/zap"
)

// WechatService 微信服务
type WechatService struct {
	wechatClient *wechat.Client
	logger       *zap.Logger
}

// NewWechatService 创建微信服务实例
func NewWechatService(wechatClient *wechat.Client, logger *zap.Logger) *WechatService {
	return &WechatService{
		wechatClient: wechatClient,
		logger:       logger,
	}
}

// SendSubscribeMessage 发送订阅消息
func (s *WechatService) SendSubscribeMessage(
	openid string,
	templateID string,
	data map[string]interface{},
	page string,
	miniprogramState string,
) error {
	s.logger.Info("🚀 [WechatService.SendSubscribeMessage] START - 开始发送微信订阅消息",
		zap.String("openid", openid),
		zap.String("templateID", templateID),
		zap.String("page", page),
		zap.String("miniprogramState", miniprogramState),
		zap.Any("data", data),
	)

	// 获取小程序订阅消息实例
	miniProgram := s.wechatClient.GetMiniProgram()
	subscribeService := miniProgram.GetSubscribe()

	// 格式化数据为 SDK 要求的格式
	formattedData := make(map[string]*subscribe.DataItem)
	for k, v := range data {
		formattedData[k] = &subscribe.DataItem{
			Value: v,
		}
	}

	// 构造消息
	msg := &subscribe.Message{
		ToUser:           openid,
		TemplateID:       templateID,
		Page:             page,
		Data:             formattedData,
		MiniprogramState: miniprogramState,
		Lang:             "zh_CN",
	}

	s.logger.Info("📦 [WechatService.SendSubscribeMessage] 发送消息",
		zap.Any("message", msg),
	)

	// 发送订阅消息
	err := subscribeService.Send(msg)
	if err != nil {
		s.logger.Error("❌ [WechatService.SendSubscribeMessage] 发送失败",
			zap.Error(err),
			zap.String("openid", openid),
			zap.String("templateID", templateID),
		)
		return err
	}

	s.logger.Info("✅ [WechatService.SendSubscribeMessage] 订阅消息发送成功",
		zap.String("openid", openid),
		zap.String("templateID", templateID),
	)

	return nil
}
