package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"regexp"
	"strings"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// UserFriendlyAgent 用户友好Agent，负责将专业分析转换为通俗易懂的内容
type UserFriendlyAgent struct {
	chatModel model.ToolCallingChatModel
	logger    *zap.Logger
}

// NewUserFriendlyAgent 创建用户友好Agent
func NewUserFriendlyAgent(chatModel model.ToolCallingChatModel, logger *zap.Logger) *UserFriendlyAgent {
	return &UserFriendlyAgent{
		chatModel: chatModel,
		logger:    logger,
	}
}


// GenerateUserFriendlyAnalysis 生成用户友好的分析结果
func (a *UserFriendlyAgent) GenerateUserFriendlyAnalysis(ctx context.Context, analysisResult *entity.AIAnalysisResult, baby *entity.Baby) (*entity.UserFriendlyResult, error) {
	// 构建系统提示
	systemPrompt := a.buildSystemPrompt(baby)
	
	// 构建用户提示
	userPrompt := a.buildUserPrompt(analysisResult)

	messages := []*schema.Message{
		schema.SystemMessage(systemPrompt),
		schema.UserMessage(userPrompt),
	}

	// 调用AI生成用户友好的分析
	response, err := a.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "生成用户友好分析失败", err)
	}

	// 解析响应
	return a.parseUserFriendlyResponse(response.Content)
}

// buildSystemPrompt 构建系统提示
func (a *UserFriendlyAgent) buildSystemPrompt(baby *entity.Baby) string {
	return fmt.Sprintf(`你是一个温暖、专业的育儿顾问，擅长将复杂的数据分析转换为父母容易理解和实施的建议。

宝宝信息：
- 姓名：%s
- 性别：%s
- 出生日期：%s

你的任务是将专业的AI分析结果转换为：
1. 温暖、鼓励的语言风格
2. 通俗易懂的表达方式
3. 具体可操作的建议
4. 积极正面的态度

**重要：最终必须只返回纯JSON格式，不要包含任何解释文字。**

JSON格式要求：
{
  "overall_summary": "总体评价，用温暖的语言概括宝宝的整体情况",
  "score_explanation": "评分说明，用通俗的语言解释评分含义",
  "key_highlights": [
    {
      "title": "亮点标题",
      "description": "亮点描述，突出宝宝的优秀表现",
      "icon": "建议的图标名称"
    }
  ],
  "improvement_areas": [
    {
      "area": "改进领域",
      "issue": "问题描述，用温和的语言",
      "suggestion": "具体建议，可操作性强",
      "priority": "优先级",
      "difficulty": "实施难度"
    }
  ],
  "next_step_actions": [
    {
      "action": "具体行动",
      "timeline": "时间安排",
      "benefit": "预期收益",
      "how_to": "具体做法"
    }
  ],
  "encouraging_words": "鼓励话语，给父母信心和支持"
}

**请确保响应只包含有效的JSON，语言温暖友好，建议具体可行。**`, 
		baby.Name, baby.Gender, baby.BirthDate)
}

// buildUserPrompt 构建用户提示
func (a *UserFriendlyAgent) buildUserPrompt(analysisResult *entity.AIAnalysisResult) string {
	// 将分析结果转换为JSON字符串
	resultJSON, _ := json.Marshal(analysisResult)
	
	return fmt.Sprintf(`请将以下专业的AI分析结果转换为用户友好的格式：

专业分析结果：
%s

请重点关注：
1. 将专业术语转换为日常用语
2. 突出积极的方面，给父母信心
3. 将警告转换为温和的建议
4. 提供具体可操作的改进方案
5. 用鼓励的语言结束

分析类型：%s
评分：%.1f分`, string(resultJSON), analysisResult.AnalysisType, analysisResult.Score)
}

// parseUserFriendlyResponse 解析用户友好响应
func (a *UserFriendlyAgent) parseUserFriendlyResponse(content string) (*entity.UserFriendlyResult, error) {
	// 记录原始响应用于调试
	a.logger.Debug("用户友好Agent原始响应", zap.String("content", content))
	
	// 清理和提取JSON
	jsonContent := a.extractJSON(content)
	if jsonContent == "" {
		a.logger.Error("无法从用户友好响应中提取JSON", zap.String("content", content))
		return nil, errors.New(errors.InternalError, "用户友好响应中未找到有效的JSON格式")
	}
	
	a.logger.Debug("用户友好Agent提取的JSON", zap.String("json", jsonContent))

	var result entity.UserFriendlyResult
	if err := json.Unmarshal([]byte(jsonContent), &result); err != nil {
		a.logger.Error("用户友好JSON解析失败", 
			zap.String("json", jsonContent), 
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.InternalError, "解析用户友好响应失败", err)
	}

	return &result, nil
}

// extractJSON 从响应中提取JSON内容
func (a *UserFriendlyAgent) extractJSON(content string) string {
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
