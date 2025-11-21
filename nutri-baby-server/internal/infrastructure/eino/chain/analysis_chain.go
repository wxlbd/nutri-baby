package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"
	"sync"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/eino/cache"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/eino/prompts"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/eino/tools"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
	"go.uber.org/zap"
)

const (
	// 最大迭代次数 - 防止无限循环
	maxIterations = 10
	// 最大消息历史数 - 控制上下文窗口大小
	maxMessageHistory = 20
	// 最小保留消息数 - 至少保留 system + user + 最近几轮
	minMessageHistory = 6
	// 早停机制 - 只要没有工具调用，就视为分析完成（因为我们强制要求返回JSON）
	earlyStopThreshold = 1
)

// AnalysisChainBuilder 分析链构建器
type AnalysisChainBuilder struct {
	chatModel      model.ToolCallingChatModel
	dataTools      *tools.DataQueryTools
	batchDataTools *tools.BatchDataTools
	dataCache      *cache.AnalysisDataCache
	logger         *zap.Logger
	enableParallel bool // 是否启用并行工具调用
}

// NewAnalysisChainBuilder 创建分析链构建器
func NewAnalysisChainBuilder(
	chatModel model.ToolCallingChatModel,
	dataTools *tools.DataQueryTools,
	batchDataTools *tools.BatchDataTools,
	logger *zap.Logger,
) *AnalysisChainBuilder {
	// 创建数据缓存（5分钟TTL，最多缓存100个宝宝的数据）
	dataCache := cache.NewAnalysisDataCache(5*time.Minute, 100)

	return &AnalysisChainBuilder{
		chatModel:      chatModel,
		dataTools:      dataTools,
		batchDataTools: batchDataTools,
		dataCache:      dataCache,
		logger:         logger,
		enableParallel: true, // 默认启用并行优化
	}
}

// getDataTypesForAnalysis 根据分析类型获取需要的数据类型
func (b *AnalysisChainBuilder) getDataTypesForAnalysis(analysisType entity.AIAnalysisType) []string {
	switch analysisType {
	case entity.AIAnalysisTypeFeeding:
		return []string{"baby_info", "feeding", "diaper"}
	case entity.AIAnalysisTypeSleep:
		return []string{"baby_info", "sleep"}
	case entity.AIAnalysisTypeGrowth:
		return []string{"baby_info", "growth"}
	case entity.AIAnalysisTypeHealth:
		return []string{"baby_info", "feeding", "sleep", "growth", "diaper"}
	case entity.AIAnalysisTypeBehavior:
		return []string{"baby_info", "feeding", "sleep", "diaper"}
	default:
		return []string{"baby_info"}
	}
}

// preloadBatchData 预加载批量数据
func (b *AnalysisChainBuilder) preloadBatchData(
	ctx context.Context,
	babyID int64,
	startDate, endDate time.Time,
	analysisType entity.AIAnalysisType,
) (string, error) {
	dataTypes := b.getDataTypesForAnalysis(analysisType)

	params := map[string]interface{}{
		"baby_id":    float64(babyID),
		"start_date": startDate.Format("2006-01-02"),
		"end_date":   endDate.Format("2006-01-02"),
		"data_types": dataTypes,
	}

	// 使用 BatchDataTools 执行批量查询
	result, err := b.batchDataTools.Execute(ctx, params)
	if err != nil {
		return "", err
	}

	b.logger.Debug("批量数据预加载完成",
		zap.Int64("baby_id", babyID),
		zap.Strings("data_types", dataTypes),
		zap.Int("result_size", len(result)),
	)

	return result, nil
}

// trimMessageHistory 修剪消息历史，保持在合理范围内
func (b *AnalysisChainBuilder) trimMessageHistory(messages []*schema.Message) []*schema.Message {
	if len(messages) <= maxMessageHistory {
		return messages
	}

	// 保留 system prompt (第1条) 和 user prompt (第2条)
	// 移除中间的旧消息，保留最新的消息
	keepFromStart := 2
	keepFromEnd := maxMessageHistory - keepFromStart

	if keepFromEnd < minMessageHistory-keepFromStart {
		keepFromEnd = minMessageHistory - keepFromStart
	}

	trimmed := make([]*schema.Message, 0, keepFromStart+keepFromEnd)
	trimmed = append(trimmed, messages[:keepFromStart]...)
	trimmed = append(trimmed, messages[len(messages)-keepFromEnd:]...)

	b.logger.Debug("消息历史已修剪",
		zap.Int("original_count", len(messages)),
		zap.Int("trimmed_count", len(trimmed)),
		zap.Int("removed_count", len(messages)-len(trimmed)),
	)

	return trimmed
}

// Analyze 执行AI分析
func (b *AnalysisChainBuilder) Analyze(ctx context.Context, analysis *entity.AIAnalysis) (*entity.AIAnalysisResult, error) {
	// 预加载批量数据
	batchData, err := b.preloadBatchData(ctx, analysis.BabyID, analysis.StartDate, analysis.EndDate, analysis.AnalysisType)
	if err != nil {
		b.logger.Warn("批量数据预加载失败，将使用按需加载",
			zap.Error(err),
		)
		// 预加载失败不影响后续流程，AI 可以按需调用工具
		batchData = ""
	}

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

	// 如果批量数据预加载成功，添加到初始消息中
	if batchData != "" {
		messages = append(messages, schema.UserMessage(fmt.Sprintf("以下是已预加载的数据，可以直接使用：\n%s", batchData)))
	}

	// 开始对话循环，处理工具调用
	consecutiveNoToolCalls := 0 // 连续无工具调用的次数
	for i := 0; i < maxIterations; i++ {
		b.logger.Debug("AI分析迭代",
			zap.Int("iteration", i+1),
			zap.Int("message_count", len(messages)),
		)

		response, err := toolBoundModel.Generate(ctx, messages)
		if err != nil {
			return nil, errors.Wrap(errors.InternalError, "AI分析失败", err)
		}

		messages = append(messages, response)

		// 检查是否有工具调用
		if len(response.ToolCalls) == 0 {
			consecutiveNoToolCalls++
			b.logger.Debug("无工具调用",
				zap.Int("consecutive_count", consecutiveNoToolCalls),
			)

			// 早停机制：连续 N 次没有工具调用，说明分析完成
			if consecutiveNoToolCalls >= earlyStopThreshold {
				b.logger.Debug("触发早停机制",
					zap.Int("iteration", i+1),
					zap.Int("consecutive_no_tool_calls", consecutiveNoToolCalls),
				)
				result, err := b.parseAnalysisResponse(response.Content, analysis.AnalysisType, analysis.BabyID)
				if err != nil {
					return nil, err
				}

				b.logger.Info("AI分析完成",
					zap.Int64("baby_id", analysis.BabyID),
					zap.String("analysis_type", string(analysis.AnalysisType)),
					zap.Int("iterations", i+1),
					zap.Int("final_message_count", len(messages)),
				)

				return result, nil
			}
		} else {
			// 有工具调用，重置计数器
			consecutiveNoToolCalls = 0

			// 处理工具调用（支持并行执行）
			if b.enableParallel && len(response.ToolCalls) > 1 {
				// 并行执行多个工具调用
				toolResults := b.executeToolCallsParallel(ctx, response.ToolCalls)
				messages = append(messages, toolResults...)
			} else {
				// 串行执行工具调用
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
		}

		// 修剪消息历史，控制上下文窗口大小
		messages = b.trimMessageHistory(messages)
	}

	return nil, errors.New(errors.InternalError, "分析超时，达到最大迭代次数")
}

// GenerateDailyTips 生成每日建议
func (b *AnalysisChainBuilder) GenerateDailyTips(ctx context.Context, baby *entity.Baby, date time.Time) ([]entity.DailyTip, error) {
	// 预加载批量数据（每日建议通常需要综合数据）
	// 使用 AIAnalysisTypeHealth 获取最全面的数据
	endDate := date.Add(24 * time.Hour)
	startDate := date.Add(-7 * 24 * time.Hour) // 获取过去7天的数据
	batchData, err := b.preloadBatchData(ctx, baby.ID, startDate, endDate, entity.AIAnalysisTypeHealth)
	if err != nil {
		b.logger.Warn("批量数据预加载失败，将使用按需加载",
			zap.Error(err),
		)
		batchData = ""
	}

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

	// 如果批量数据预加载成功，添加到初始消息中
	if batchData != "" {
		messages = append(messages, schema.UserMessage(fmt.Sprintf("以下是已预加载的最近7天数据，可以直接使用：\n%s", batchData)))
	}

	// 对话循环处理工具调用
	consecutiveNoToolCalls := 0
	for i := 0; i < maxIterations; i++ {
		b.logger.Debug("每日建议生成迭代",
			zap.Int("iteration", i+1),
			zap.Int("message_count", len(messages)),
		)

		response, err := toolBoundModel.Generate(ctx, messages)
		if err != nil {
			return nil, errors.Wrap(errors.InternalError, "生成每日建议失败", err)
		}

		messages = append(messages, response)

		// 检查是否有工具调用
		if len(response.ToolCalls) == 0 {
			consecutiveNoToolCalls++
			if consecutiveNoToolCalls >= earlyStopThreshold {
				b.logger.Info("每日建议生成完成",
					zap.Int64("baby_id", baby.ID),
					zap.Int("iterations", i+1),
				)
				// 没有工具调用，解析建议
				return b.parseDailyTipsResponse(response.Content)
			}
		} else {
			consecutiveNoToolCalls = 0
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

		// 修剪消息历史
		messages = b.trimMessageHistory(messages)
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

// executeToolCallsParallel 并行执行多个工具调用
func (b *AnalysisChainBuilder) executeToolCallsParallel(ctx context.Context, toolCalls []schema.ToolCall) []*schema.Message {
	var wg sync.WaitGroup
	results := make([]*schema.Message, len(toolCalls))

	for i, toolCall := range toolCalls {
		wg.Add(1)
		go func(index int, tc schema.ToolCall) {
			defer wg.Done()

			toolResult, err := b.executeToolCall(ctx, tc)
			if err != nil {
				b.logger.Error("并行工具调用失败",
					zap.String("tool_name", tc.Function.Name),
					zap.Error(err),
				)
				toolResult = fmt.Sprintf("工具调用失败: %v", err)
			}

			results[index] = &schema.Message{
				Role:       schema.Tool,
				Content:    toolResult,
				ToolCallID: tc.ID,
			}
		}(i, toolCall)
	}

	wg.Wait()

	b.logger.Debug("并行工具调用完成",
		zap.Int("tool_count", len(toolCalls)),
	)

	return results
}

// buildSystemPrompt 构建系统提示
func (b *AnalysisChainBuilder) buildSystemPrompt(analysisType entity.AIAnalysisType) string {

	switch analysisType {
	case entity.AIAnalysisTypeFeeding:
		return prompts.AnalysisSystem + "\n\n专业领域：婴幼儿喂养营养分析。重点关注喂养规律、营养摄入、消化健康等方面。"
	case entity.AIAnalysisTypeSleep:
		return prompts.AnalysisSystem + "\n\n专业领域：婴幼儿睡眠质量分析。重点关注睡眠时长、作息规律、睡眠质量等方面。"
	case entity.AIAnalysisTypeGrowth:
		return prompts.AnalysisSystem + "\n\n专业领域：婴幼儿生长发育分析。重点关注身高体重增长、发育里程碑、WHO标准对比等方面。"
	case entity.AIAnalysisTypeHealth:
		return prompts.AnalysisSystem + "\n\n专业领域：婴幼儿综合健康分析。需要综合多种数据进行整体健康评估。"
	case entity.AIAnalysisTypeBehavior:
		return prompts.AnalysisSystem + "\n\n专业领域：婴幼儿行为模式分析。重点关注行为发展、习惯养成、个性特征等方面。"
	default:
		return prompts.AnalysisSystem
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
	return prompts.DailyTipsSystem
}

// buildDailyTipsUserPrompt 构建每日建议用户提示
func (b *AnalysisChainBuilder) buildDailyTipsUserPrompt(baby *entity.Baby, date time.Time) string {
	return fmt.Sprintf(`请分析宝宝（ID: %d）在 %s 的各项数据，生成今日的个性化育儿建议。
	请检查上下文中是否已有足够数据，若不足请调用工具查询最近 7 天的详细记录（喂养、睡眠、成长等）。
	记住：保持温暖亲切的语气，仅返回 JSON 数组。`,
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
		OverallScore     float64                        `json:"overall_score"`
		OverallSummary   string                         `json:"overall_summary"`
		KeyHighlights    []entity.UserFriendlyHighlight `json:"key_highlights"`
		ImprovementAreas []entity.UserFriendlyImprovement `json:"improvement_areas"`
		NextStepActions  []entity.UserFriendlyAction    `json:"next_step_actions"`
		EncouragingWords string                         `json:"encouraging_words"`
	}

	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		b.logger.Error("JSON解析失败",
			zap.String("json", jsonContent),
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.InternalError, "解析分析响应失败", err)
	}

	// 将新的简化结构转换为现有的AIAnalysisResult结构
	return &entity.AIAnalysisResult{
		BabyID:       babyID,
		AnalysisType: analysisType,
		Score:        result.OverallScore,
		// 简化结构不再直接返回insights/alerts/patterns/predictions数组
		// 这些信息现在包含在user_friendly结构中
		UserFriendly: &entity.UserFriendlyResult{
			OverallSummary:   result.OverallSummary,
			ScoreExplanation: "", // 新结构中没有评分说明，可以留空或生成
			KeyHighlights:    result.KeyHighlights,
			ImprovementAreas: result.ImprovementAreas,
			NextStepActions:  result.NextStepActions,
			EncouragingWords: result.EncouragingWords,
		},
	}, nil
}

// parseDailyTipsResponse 解析每日建议响应
func (b *AnalysisChainBuilder) parseDailyTipsResponse(content string) ([]entity.DailyTip, error) {
	// 清理响应内容，移除可能的代码块标记和其他格式化字符
	cleanContent := b.cleanJSONResponse(content)

	var tips []entity.DailyTip
	if err := json.Unmarshal([]byte(cleanContent), &tips); err != nil {
		// 记录原始内容以便调试
		b.logger.Error("JSON解析失败",
			zap.String("original_content", content),
			zap.String("cleaned_content", cleanContent),
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.InternalError, "解析建议响应失败", err)
	}
	return tips, nil
}

// cleanJSONResponse 清理AI响应中的格式化字符，提取纯JSON
func (b *AnalysisChainBuilder) cleanJSONResponse(content string) string {
	// 去除前后空白
	content = strings.TrimSpace(content)

	// 移除常见的代码块标记
	content = strings.TrimPrefix(content, "```json")
	content = strings.TrimPrefix(content, "```")
	content = strings.TrimSuffix(content, "```")

	// 移除可能的前缀文本（如"以下是建议："等）
	lines := strings.Split(content, "\n")
	var jsonLines []string
	jsonStarted := false

	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// 检查是否是JSON开始
		if strings.HasPrefix(line, "[") || strings.HasPrefix(line, "{") {
			jsonStarted = true
		}

		// 如果JSON已经开始，收集所有行
		if jsonStarted {
			jsonLines = append(jsonLines, line)
		}
	}

	if len(jsonLines) > 0 {
		content = strings.Join(jsonLines, "\n")
	}

	// 再次去除前后空白
	content = strings.TrimSpace(content)

	return content
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
