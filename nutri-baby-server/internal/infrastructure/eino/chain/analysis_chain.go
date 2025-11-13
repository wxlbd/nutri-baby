package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/eino/tools"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
	"go.uber.org/zap"
)

// AnalysisChainBuilder 分析链构建器
type AnalysisChainBuilder struct {
	chatModel         model.ToolCallingChatModel
	dataTools         *tools.DataQueryTools
	userFriendlyAgent *UserFriendlyAgent
	logger            *zap.Logger
}

// NewAnalysisChainBuilder 创建分析链构建器
func NewAnalysisChainBuilder(
	chatModel model.ToolCallingChatModel,
	dataTools *tools.DataQueryTools,
	logger *zap.Logger,
) *AnalysisChainBuilder {
	// 创建用户友好Agent
	userFriendlyAgent := NewUserFriendlyAgent(chatModel, logger)

	return &AnalysisChainBuilder{
		chatModel:         chatModel,
		dataTools:         dataTools,
		userFriendlyAgent: userFriendlyAgent,
		logger:            logger,
	}
}

// Analyze 执行AI分析
func (b *AnalysisChainBuilder) Analyze(ctx context.Context, analysis *entity.AIAnalysis) (*entity.AIAnalysisResult, error) {
	// 绑定数据查询工具
	toolBoundModel, err := b.chatModel.WithTools(b.dataTools.GetToolInfos())
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "绑定工具失败", err)
	}

	// 构建系统提示
	systemPrompt := b.buildSystemPrompt(analysis.AnalysisType)

	// 构建用户提示
	userPrompt := b.buildUserPrompt(analysis)

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	}

	// 开始对话循环，处理工具调用
	maxIterations := 10 // 防止无限循环
	for i := 0; i < maxIterations; i++ {
		response, err := toolBoundModel.Generate(ctx, messages)
		if err != nil {
			return nil, errors.Wrap(errors.InternalError, "AI分析失败", err)
		}

		messages = append(messages, response)

		// 检查是否有工具调用
		if len(response.ToolCalls) == 0 {
			// 没有工具调用，说明分析完成
			result, err := b.parseAnalysisResponse(response.Content, analysis.AnalysisType, analysis.BabyID)
			if err != nil {
				return nil, err
			}

			// 生成用户友好的分析结果
			if err := b.generateUserFriendlyResult(ctx, result, analysis.BabyID); err != nil {
				b.logger.Error("生成用户友好结果失败", zap.Error(err))
				// 不影响主要分析结果，继续返回
			}

			return result, nil
		}

		// 处理工具调用
		for _, toolCall := range response.ToolCalls {
			toolResult, err := b.executeToolCall(ctx, toolCall)
			if err != nil {
				b.logger.Error("工具调用失败",
					zap.String("tool_name", toolCall.Function.Name),
					zap.Error(err),
				)
				toolResult = fmt.Sprintf("工具调用失败: %v", err)
			}

			// 添加工具调用结果到消息历史
			messages = append(messages, &schema.Message{
				Role:       schema.Tool,
				Content:    toolResult,
				ToolCallID: toolCall.ID,
			})
		}
	}

	return nil, errors.New(errors.InternalError, "分析超时，达到最大迭代次数")
}

// GenerateDailyTips 生成每日建议
func (b *AnalysisChainBuilder) GenerateDailyTips(ctx context.Context, baby *entity.Baby, date time.Time) ([]entity.DailyTip, error) {
	// 绑定数据查询工具
	toolBoundModel, err := b.chatModel.WithTools(b.dataTools.GetToolInfos())
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "绑定工具失败", err)
	}

	systemPrompt := b.buildDailyTipsSystemPrompt()
	userPrompt := b.buildDailyTipsUserPrompt(baby, date)

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	}

	// 对话循环处理工具调用
	maxIterations := 10
	for i := 0; i < maxIterations; i++ {
		response, err := toolBoundModel.Generate(ctx, messages)
		if err != nil {
			return nil, errors.Wrap(errors.InternalError, "生成每日建议失败", err)
		}

		messages = append(messages, response)

		// 检查是否有工具调用
		if len(response.ToolCalls) == 0 {
			// 没有工具调用，解析建议
			return b.parseDailyTipsResponse(response.Content)
		}

		// 处理工具调用
		for _, toolCall := range response.ToolCalls {
			toolResult, err := b.executeToolCall(ctx, toolCall)
			if err != nil {
				b.logger.Error("工具调用失败",
					zap.String("tool_name", toolCall.Function.Name),
					zap.Error(err),
				)
				toolResult = fmt.Sprintf("工具调用失败: %v", err)
			}

			messages = append(messages, &schema.Message{
				Role:       schema.Tool,
				Content:    toolResult,
				ToolCallID: toolCall.ID,
			})
		}
	}

	return nil, errors.New(errors.InternalError, "生成建议超时，达到最大迭代次数")
}

// executeToolCall 执行工具调用
func (b *AnalysisChainBuilder) executeToolCall(ctx context.Context, toolCall schema.ToolCall) (string, error) {
	// 解析工具参数
	var params map[string]interface{}
	if err := json.Unmarshal([]byte(toolCall.Function.Arguments), &params); err != nil {
		return "", fmt.Errorf("解析工具参数失败: %v", err)
	}

	// 执行工具
	return b.dataTools.ExecuteTool(ctx, toolCall.Function.Name, params)
}

// buildSystemPrompt 构建系统提示
func (b *AnalysisChainBuilder) buildSystemPrompt(analysisType entity.AIAnalysisType) string {
	basePrompt := `你是一个专业的婴幼儿护理专家，擅长分析宝宝的各项数据并提供专业建议。

你可以使用以下工具来获取宝宝的数据：
- get_baby_info: 获取宝宝基本信息
- get_feeding_data: 获取喂养记录
- get_sleep_data: 获取睡眠记录  
- get_growth_data: 获取成长记录
- get_diaper_data: 获取尿布记录
- get_vaccine_data: 获取疫苗记录

请根据分析类型，主动调用相关工具获取数据，然后进行专业分析。

**重要：最终必须只返回纯JSON格式的分析结果，不要包含任何解释文字或其他内容。**

JSON格式要求：
{
  "score": 0-100的评分,
  "insights": [洞察数组],
  "alerts": [警告数组],
  "patterns": [模式数组],
  "predictions": [预测数组]
}

**请确保响应只包含有效的JSON，不要添加任何前缀、后缀或解释文本。**

每个洞察包含：type(string), title(string), description(string), priority(string), category(string)
每个警告包含：level(string), type(string), title(string), description(string), suggestion(string), timestamp(time.Time)
每个模式包含：pattern_type(string), description(string), confidence(float64), frequency(string), time_range(TimeRange对象，包含start和end时间)
每个预测包含：prediction_type(string), value(string), confidence(float64), time_frame(string), reason(string)

注意：
- confidence字段必须是0-1之间的浮点数
- timestamp字段使用ISO 8601格式的时间字符串
- time_range对象格式：{"start": "2024-01-01T00:00:00Z", "end": "2024-01-02T00:00:00Z"}`

	switch analysisType {
	case entity.AIAnalysisTypeFeeding:
		return basePrompt + "\n\n专业领域：婴幼儿喂养营养分析。重点关注喂养规律、营养摄入、消化健康等方面。"
	case entity.AIAnalysisTypeSleep:
		return basePrompt + "\n\n专业领域：婴幼儿睡眠质量分析。重点关注睡眠时长、作息规律、睡眠质量等方面。"
	case entity.AIAnalysisTypeGrowth:
		return basePrompt + "\n\n专业领域：婴幼儿生长发育分析。重点关注身高体重增长、发育里程碑、WHO标准对比等方面。"
	case entity.AIAnalysisTypeHealth:
		return basePrompt + "\n\n专业领域：婴幼儿综合健康分析。需要综合多种数据进行整体健康评估。"
	case entity.AIAnalysisTypeBehavior:
		return basePrompt + "\n\n专业领域：婴幼儿行为模式分析。重点关注行为发展、习惯养成、个性特征等方面。"
	default:
		return basePrompt
	}
}

// buildUserPrompt 构建用户提示
func (b *AnalysisChainBuilder) buildUserPrompt(analysis *entity.AIAnalysis) string {
	return fmt.Sprintf(`请对宝宝ID %d 在 %s 至 %s 期间的 %s 数据进行专业分析。

请先获取宝宝的基本信息，然后根据分析类型获取相关数据，最后提供专业的分析报告。`,
		analysis.BabyID,
		analysis.StartDate.Format("2006-01-02"),
		analysis.EndDate.Format("2006-01-02"),
		b.getAnalysisTypeName(analysis.AnalysisType),
	)
}

// buildDailyTipsSystemPrompt 构建每日建议系统提示
func (b *AnalysisChainBuilder) buildDailyTipsSystemPrompt() string {
	return `你是一个专业的育儿专家，擅长根据宝宝的日常数据提供个性化的育儿建议。

你可以使用工具获取宝宝的各项数据，然后基于这些数据生成实用的育儿建议。

请生成3-5条实用、具体的育儿建议，以JSON数组格式返回：
[
  {
    "id": "唯一标识",
    "icon": "表情符号",
    "title": "建议标题",
    "description": "详细描述",
    "type": "类型(feeding/sleep/growth/health/behavior)",
    "priority": "优先级(high/medium/low)",
    "action_url": "相关页面链接(可选)"
  }
]

建议应该：
1. 基于实际数据，具有针对性
2. 实用性强，易于执行
3. 考虑宝宝的月龄和发展阶段
4. 包含具体的行动建议
5. 使用友好的语气`
}

// buildDailyTipsUserPrompt 构建每日建议用户提示
func (b *AnalysisChainBuilder) buildDailyTipsUserPrompt(baby *entity.Baby, date time.Time) string {
	return fmt.Sprintf(`请为宝宝ID %d 生成 %s 的个性化育儿建议。

请先获取宝宝的基本信息，然后获取最近7天的相关数据（喂养、睡眠、成长等），基于这些数据生成针对性的建议。`,
		baby.ID,
		date.Format("2006-01-02"),
	)
}

// getAnalysisTypeName 获取分析类型名称
func (b *AnalysisChainBuilder) getAnalysisTypeName(analysisType entity.AIAnalysisType) string {
	switch analysisType {
	case entity.AIAnalysisTypeFeeding:
		return "喂养"
	case entity.AIAnalysisTypeSleep:
		return "睡眠"
	case entity.AIAnalysisTypeGrowth:
		return "成长"
	case entity.AIAnalysisTypeHealth:
		return "健康"
	case entity.AIAnalysisTypeBehavior:
		return "行为"
	default:
		return "综合"
	}
}

// parseAnalysisResponse 解析分析响应
func (b *AnalysisChainBuilder) parseAnalysisResponse(content string, analysisType entity.AIAnalysisType, babyID int64) (*entity.AIAnalysisResult, error) {
	// 记录原始响应用于调试
	b.logger.Debug("原始AI响应", zap.String("content", content))

	// 清理和提取JSON
	jsonContent := b.extractJSON(content)
	if jsonContent == "" {
		b.logger.Error("无法从响应中提取JSON", zap.String("content", content))
		return nil, errors.New(errors.InternalError, "响应中未找到有效的JSON格式")
	}

	b.logger.Debug("提取的JSON", zap.String("json", jsonContent))

	var result struct {
		Score       float64               `json:"score"`
		Insights    []entity.AIInsight    `json:"insights"`
		Alerts      []entity.AIAlert      `json:"alerts"`
		Patterns    []entity.AIPattern    `json:"patterns"`
		Predictions []entity.AIPrediction `json:"predictions"`
	}

	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		b.logger.Error("JSON解析失败",
			zap.String("json", jsonContent),
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.InternalError, "解析分析响应失败", err)
	}

	return &entity.AIAnalysisResult{
		BabyID:       babyID,
		AnalysisType: analysisType,
		Score:        result.Score,
		Insights:     result.Insights,
		Alerts:       result.Alerts,
		Patterns:     result.Patterns,
		Predictions:  result.Predictions,
	}, nil
}

// parseDailyTipsResponse 解析每日建议响应
func (b *AnalysisChainBuilder) parseDailyTipsResponse(content string) ([]entity.DailyTip, error) {
	var tips []entity.DailyTip
	if err := json.Unmarshal([]byte(content), &tips); err != nil {
		return nil, errors.Wrap(errors.InternalError, "解析建议响应失败", err)
	}
	return tips, nil
}

// extractJSON 从响应中提取JSON内容
func (b *AnalysisChainBuilder) extractJSON(content string) string {
	// 去除前后空白
	content = strings.TrimSpace(content)

	// 如果内容以 { 开始，尝试找到完整的JSON
	if strings.HasPrefix(content, "{") {
		// 找到第一个 { 和最后一个 } 之间的内容
		braceCount := 0
		start := -1
		end := -1

		for i, char := range content {
			if char == '{' {
				if start == -1 {
					start = i
				}
				braceCount++
			} else if char == '}' {
				braceCount--
				if braceCount == 0 && start != -1 {
					end = i + 1
					break
				}
			}
		}

		if start != -1 && end != -1 {
			return content[start:end]
		}
	}

	// 尝试使用正则表达式提取JSON
	re := regexp.MustCompile(`\{[^{}]*(?:\{[^{}]*\}[^{}]*)*\}`)
	matches := re.FindAllString(content, -1)

	// 返回最长的匹配项（可能是最完整的JSON）
	var longest string
	for _, match := range matches {
		if len(match) > len(longest) {
			longest = match
		}
	}

	if longest != "" {
		return longest
	}

	// 如果都失败了，返回原内容让JSON解析器处理
	return content
}

// generateUserFriendlyResult 生成用户友好的分析结果
func (b *AnalysisChainBuilder) generateUserFriendlyResult(ctx context.Context, result *entity.AIAnalysisResult, babyID int64) error {
	// 获取宝宝信息
	baby, err := b.getBabyInfo(ctx, babyID)
	if err != nil {
		return errors.Wrap(errors.InternalError, "获取宝宝信息失败", err)
	}

	// 使用用户友好Agent生成结果
	userFriendlyResult, err := b.userFriendlyAgent.GenerateUserFriendlyAnalysis(ctx, result, baby)
	if err != nil {
		return errors.Wrap(errors.InternalError, "生成用户友好分析失败", err)
	}

	// 将用户友好结果添加到原始结果中
	result.UserFriendly = userFriendlyResult
	return nil
}

// getBabyInfo 获取宝宝信息（通过数据工具）
func (b *AnalysisChainBuilder) getBabyInfo(ctx context.Context, babyID int64) (*entity.Baby, error) {
	// 调用数据查询工具获取宝宝信息
	params := map[string]interface{}{
		"baby_id": float64(babyID),
	}

	resultStr, err := b.dataTools.ExecuteTool(ctx, "get_baby_info", params)
	if err != nil {
		return nil, err
	}

	// 解析结果
	var result struct {
		Type string       `json:"type"`
		Baby *entity.Baby `json:"baby"`
	}

	if err := json.Unmarshal([]byte(resultStr), &result); err != nil {
		return nil, errors.Wrap(errors.InternalError, "解析宝宝信息失败", err)
	}

	return result.Baby, nil
}
