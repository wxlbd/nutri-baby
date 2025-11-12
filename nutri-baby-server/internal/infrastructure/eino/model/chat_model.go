package model

import (
	"context"
	"strings"
	"time"

	"github.com/cloudwego/eino-ext/components/model/openai"
	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/config"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// ChatModelConfig 聊天模型配置
type ChatModelConfig struct {
	Provider    string  `yaml:"provider" json:"provider"`         // 模型提供商: openai, claude, ernie, etc.
	APIKey      string  `yaml:"api_key" json:"api_key"`           // API密钥
	BaseURL     string  `yaml:"base_url" json:"base_url"`         // 基础URL
	Model       string  `yaml:"model" json:"model"`               // 具体模型名称
	MaxTokens   int     `yaml:"max_tokens" json:"max_tokens"`     // 最大token数
	Temperature float32 `yaml:"temperature" json:"temperature"`   // 温度参数
	Timeout     int     `yaml:"timeout" json:"timeout"`           // 超时时间(秒)
	MaxRetries  int     `yaml:"max_retries" json:"max_retries"`   // 最大重试次数
	EnableCache bool    `yaml:"enable_cache" json:"enable_cache"` // 是否启用缓存
}

// NewChatModel 创建聊天模型实例
func NewChatModel(cfg *config.Config, logger *zap.Logger) (model.ChatModel, error) {
	// 使用配置文件中的AI配置
	aiConfig := cfg.AI

	// 默认使用mock模式进行开发测试
	if aiConfig.Provider == "mock" || aiConfig.Provider == "" {
		logger.Info("使用模拟AI模型进行开发测试")
		return NewMockChatModel(logger), nil
	}

	switch aiConfig.Provider {
	case "openai":
		return NewOpenAIChatModel(aiConfig.OpenAI, logger)
	case "claude":
		return NewClaudeChatModel(aiConfig.Claude, logger)
	case "ernie":
		return NewERNIEChatModel(aiConfig.ERNIE, logger)
	default:
		logger.Warn("未知的AI模型提供商，使用模拟模型", zap.String("provider", aiConfig.Provider))
		return NewMockChatModel(logger), nil
	}
}

// NewOpenAIChatModel 创建OpenAI聊天模型
func NewOpenAIChatModel(config config.OpenAIConfig, logger *zap.Logger) (model.ChatModel, error) {
	if config.APIKey == "" {
		return nil, errors.New(errors.ParamError, "OpenAI API密钥不能为空")
	}

	modelConfig := &openai.ChatModelConfig{
		APIKey:      config.APIKey,
		BaseURL:     config.BaseURL,
		Model:       config.Model,
		MaxTokens:   &config.MaxTokens,
		Temperature: float32Ptr(config.Temperature),
		Timeout:     time.Duration(30) * time.Second, // 默认30秒
	}

	// 默认配置
	if modelConfig.Model == "" {
		modelConfig.Model = "gpt-4"
	}
	if modelConfig.MaxTokens == nil || *modelConfig.MaxTokens == 0 {
		tokens := 4000
		modelConfig.MaxTokens = &tokens
	}
	if modelConfig.Temperature == nil {
		temp := float32(0.7)
		modelConfig.Temperature = &temp
	}
	if modelConfig.Timeout == 0 {
		modelConfig.Timeout = 30 * time.Second
	}

	chatModel, err := openai.NewChatModel(context.Background(), modelConfig)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "创建OpenAI聊天模型失败", err)
	}

	logger.Info("OpenAI聊天模型初始化成功",
		zap.String("model", modelConfig.Model),
		zap.Float32("temperature", *modelConfig.Temperature),
	)

	return chatModel, nil
}

// NewClaudeChatModel 创建Claude聊天模型
func NewClaudeChatModel(cfg config.ClaudeConfig, logger *zap.Logger) (model.ChatModel, error) {
	// 这里可以实现Claude模型的集成
	// 由于Eino框架可能还没有直接的Claude支持，可以先用OpenAI兼容模式
	logger.Info("Claude模型暂使用OpenAI兼容模式")

	// 使用Anthropic的OpenAI兼容API
	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://api.anthropic.com/v1"
	}
	if cfg.Model == "" {
		cfg.Model = "claude-3-sonnet-20240229"
	}

	// 转换配置
	openaiConfig := &config.OpenAIConfig{
		APIKey:      cfg.APIKey,
		BaseURL:     cfg.BaseURL,
		Model:       cfg.Model,
		MaxTokens:   cfg.MaxTokens,
		Temperature: cfg.Temperature,
	}

	return NewOpenAIChatModel(*openaiConfig, logger)
}

// NewERNIEChatModel 创建文心一言聊天模型
func NewERNIEChatModel(cfg config.ERNIEConfig, logger *zap.Logger) (model.ChatModel, error) {
	// 百度文心一言的OpenAI兼容API
	logger.Info("ERNIE模型暂使用OpenAI兼容模式")

	if cfg.BaseURL == "" {
		cfg.BaseURL = "https://aip.baidubce.com/rpc/2.0/ai_custom/v1/wenxinworkshop/chat"
	}
	if cfg.Model == "" {
		cfg.Model = "ernie-bot"
	}

	// 转换配置
	openaiConfig := &config.OpenAIConfig{
		APIKey:      cfg.APIKey,
		BaseURL:     cfg.BaseURL,
		Model:       cfg.Model,
		MaxTokens:   2000, // 默认token数
		Temperature: 0.7,  // 默认温度
	}

	return NewOpenAIChatModel(*openaiConfig, logger)
}

// MockChatModel 模拟聊天模型（用于开发和测试）
type MockChatModel struct {
	logger *zap.Logger
}

// NewMockChatModel 创建模拟聊天模型
func NewMockChatModel(logger *zap.Logger) *MockChatModel {
	return &MockChatModel{
		logger: logger,
	}
}

// Generate 生成响应
func (m *MockChatModel) Generate(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.Message, error) {
	m.logger.Debug("MockChatModel.Generate 被调用", zap.Int("message_count", len(messages)))

	// 模拟延迟
	select {
	case <-time.After(500 * time.Millisecond):
		// 继续执行
	case <-ctx.Done():
		return nil, ctx.Err()
	}

	// 根据消息内容生成模拟响应
	lastMessage := messages[len(messages)-1]
	mockResponse := m.generateMockResponse(lastMessage.Content)

	return &schema.Message{
		Role:    schema.Assistant,
		Content: mockResponse,
	}, nil
}

// Stream 流式生成（如果要支持流式响应）
func (m *MockChatModel) Stream(ctx context.Context, messages []*schema.Message, opts ...model.Option) (*schema.StreamReader[*schema.Message], error) {
	// 返回一个模拟的流式读取器
	streamReader, streamWriter := schema.Pipe[*schema.Message](10)

	go func() {
		defer streamWriter.Close()

		// 模拟流式响应
		mockResponse := m.generateMockResponse(messages[len(messages)-1].Content)
		words := strings.Split(mockResponse, " ")

		for _, word := range words {
			select {
			case <-ctx.Done():
				return
			case <-time.After(100 * time.Millisecond):
				closed := streamWriter.Send(&schema.Message{
					Role:    schema.Assistant,
					Content: word + " ",
				}, nil)
				if closed {
					m.logger.Debug("流式发送通道已关闭")
					return
				}
			}
		}
	}()

	return streamReader, nil
}

// BindTools 绑定工具（模拟实现）
func (m *MockChatModel) BindTools(tools []*schema.ToolInfo) error {
	m.logger.Debug("MockChatModel.BindTools 被调用", zap.Int("tool_count", len(tools)))
	return nil
}

// generateMockResponse 生成模拟响应
func (m *MockChatModel) generateMockResponse(userInput string) string {
	// 分析用户输入，返回相应的模拟数据
	if strings.Contains(userInput, "喂养") || strings.Contains(userInput, "feeding") {
		return `{
			"score": 85,
			"insights": [
				{
					"type": "feeding",
					"title": "喂养规律良好",
					"description": "宝宝的喂养时间较为规律，建议继续保持",
					"priority": "medium",
					"category": "规律性"
				}
			],
			"alerts": [],
			"patterns": [
				{
					"pattern_type": "regular_feeding",
					"description": "每3-4小时喂养一次",
					"confidence": 0.9,
					"frequency": "daily"
				}
			],
			"predictions": []
		}`
	}

	if strings.Contains(userInput, "睡眠") || strings.Contains(userInput, "sleep") {
		return `{
			"score": 78,
			"insights": [
				{
					"type": "sleep",
					"title": "睡眠时长充足",
					"description": "宝宝每日睡眠时长符合月龄标准",
					"priority": "high",
					"category": "睡眠质量"
				}
			],
			"alerts": [
				{
					"level": "warning",
					"type": "sleep_interruption",
					"title": "夜间易醒",
					"description": "夜间睡眠中断次数较多",
					"suggestion": "建议检查睡眠环境，保持安静舒适"
				}
			],
			"patterns": [],
			"predictions": []
		}`
	}

	if strings.Contains(userInput, "成长") || strings.Contains(userInput, "growth") {
		return `{
			"score": 92,
			"insights": [
				{
					"type": "growth",
					"title": "生长发育良好",
					"description": "身高体重增长曲线正常，符合WHO标准",
					"priority": "high",
					"category": "发育评估"
				}
			],
			"alerts": [],
			"patterns": [],
			"predictions": [
				{
					"prediction_type": "height",
					"value": "75cm",
					"confidence": 0.85,
					"time_frame": "3个月后",
					"reason": "基于当前生长速度预测"
				}
			]
		}`
	}

	// 默认响应
	return `{
		"score": 80,
		"insights": [
			{
				"type": "general",
				"title": "整体状况良好",
				"description": "宝宝各项指标基本正常",
				"priority": "medium",
				"category": "综合评估"
			}
		],
		"alerts": [],
		"patterns": [],
		"predictions": []
	}`
}

// generateMockDailyTips 生成模拟每日建议
func (m *MockChatModel) generateMockDailyTips(babyInfo string, dataSummary string) string {
	return `[
		{
			"id": "tip_1",
			"icon": "🍼",
			"title": "喂养时间建议",
			"description": "建议在上午9-10点之间进行喂养，此时宝宝消化吸收效果最佳",
			"type": "feeding",
			"priority": "high",
			"action_url": "/pages/record/feeding/index"
		},
		{
			"id": "tip_2",
			"icon": "😴",
			"title": "午睡时间安排",
			"description": "建议午睡时间控制在1-2小时，避免影响夜间睡眠",
			"type": "sleep",
			"priority": "medium",
			"action_url": "/pages/record/sleep/index"
		},
		{
			"id": "tip_3",
			"icon": "🌡️",
			"title": "体温监测提醒",
			"description": "建议每天固定时间测量体温，关注宝宝健康状况",
			"type": "health",
			"priority": "low"
		}
	]`
}

// float32Ptr converts float64 to *float32
func float32Ptr(f float64) *float32 {
	result := float32(f)
	return &result
}

// Ensure MockChatModel implements model.ChatModel interface
var _ model.ChatModel = (*MockChatModel)(nil)
