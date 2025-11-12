package chain

import (
	"context"
	"encoding/json"
	"fmt"
	"time"

	"github.com/cloudwego/eino/components/model"
	"github.com/cloudwego/eino/schema"
	"go.uber.org/zap"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// AnalysisChainBuilder 分析链构建器
type AnalysisChainBuilder struct {
	chatModel model.ChatModel
	logger    *zap.Logger
}

// NewAnalysisChainBuilder 创建分析链构建器
func NewAnalysisChainBuilder(chatModel model.ChatModel, logger *zap.Logger) *AnalysisChainBuilder {
	return &AnalysisChainBuilder{
		chatModel: chatModel,
		logger:    logger,
	}
}

// AnalysisData 分析数据
type AnalysisData struct {
	Baby           *entity.Baby              `json:"baby"`
	FeedingRecords []entity.FeedingRecord    `json:"feeding_records,omitempty"`
	SleepRecords   []entity.SleepRecord      `json:"sleep_records,omitempty"`
	GrowthRecords  []entity.GrowthRecord     `json:"growth_records,omitempty"`
	DiaperRecords  []entity.DiaperRecord     `json:"diaper_records,omitempty"`
	VaccineRecords []entity.BabyVaccineSchedule `json:"vaccine_records,omitempty"`
	AnalysisType   entity.AIAnalysisType     `json:"analysis_type"`
	StartDate      time.Time                 `json:"start_date"`
	EndDate        time.Time                 `json:"end_date"`
}

// AnalysisResult 分析结果
type AnalysisResult struct {
	Score       float64                `json:"score"`
	Insights    []entity.AIInsight     `json:"insights"`
	Alerts      []entity.AIAlert       `json:"alerts"`
	Patterns    []entity.AIPattern     `json:"patterns"`
	Predictions []entity.AIPrediction  `json:"predictions"`
	Metadata    map[string]interface{} `json:"metadata"`
}

// Analyze 执行AI分析
func (b *AnalysisChainBuilder) Analyze(ctx context.Context, analysis *entity.AIAnalysis, data map[string]interface{}) (*entity.AIAnalysisResult, error) {
	// 构建分析数据
	analysisData, err := b.buildAnalysisData(analysis, data)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "构建分析数据失败", err)
	}

	// 根据分析类型选择不同的分析链
	switch analysis.AnalysisType {
	case entity.AIAnalysisTypeFeeding:
		return b.analyzeFeeding(ctx, analysisData)
	case entity.AIAnalysisTypeSleep:
		return b.analyzeSleep(ctx, analysisData)
	case entity.AIAnalysisTypeGrowth:
		return b.analyzeGrowth(ctx, analysisData)
	case entity.AIAnalysisTypeHealth:
		return b.analyzeHealth(ctx, analysisData)
	case entity.AIAnalysisTypeBehavior:
		return b.analyzeBehavior(ctx, analysisData)
	default:
		return nil, errors.New(errors.ParamError, "不支持的分析类型")
	}
}

// GenerateDailyTips 生成每日建议
func (b *AnalysisChainBuilder) GenerateDailyTips(ctx context.Context, baby *entity.Baby, data map[string]interface{}) ([]entity.DailyTip, error) {
	// 构建提示消息
	messages := []*schema.Message{
		schema.SystemMessage(b.buildDailyTipsSystemPrompt()),
		schema.UserMessage(b.buildDailyTipsUserPrompt(baby, data)),
	}

	// 调用大模型
	response, err := b.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "调用大模型失败", err)
	}

	// 解析响应
	tips, err := b.parseDailyTipsResponse(response.Content)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "解析建议响应失败", err)
	}

	return tips, nil
}

// buildAnalysisData 构建分析数据
func (b *AnalysisChainBuilder) buildAnalysisData(analysis *entity.AIAnalysis, data map[string]interface{}) (*AnalysisData, error) {
	analysisData := &AnalysisData{
		AnalysisType: analysis.AnalysisType,
		StartDate:    analysis.StartDate,
		EndDate:      analysis.EndDate,
	}

	// 解析宝宝信息
	if baby, ok := data["baby"].(*entity.Baby); ok {
		analysisData.Baby = baby
	}

	// 解析各种记录数据
	if feedingRecords, ok := data["feeding_records"].([]entity.FeedingRecord); ok {
		analysisData.FeedingRecords = feedingRecords
	}

	if sleepRecords, ok := data["sleep_records"].([]entity.SleepRecord); ok {
		analysisData.SleepRecords = sleepRecords
	}

	if growthRecords, ok := data["growth_records"].([]entity.GrowthRecord); ok {
		analysisData.GrowthRecords = growthRecords
	}

	if diaperRecords, ok := data["diaper_records"].([]entity.DiaperRecord); ok {
		analysisData.DiaperRecords = diaperRecords
	}

	if vaccineRecords, ok := data["vaccine_records"].([]entity.BabyVaccineSchedule); ok {
		analysisData.VaccineRecords = vaccineRecords
	}

	return analysisData, nil
}

// analyzeFeeding 分析喂养数据
func (b *AnalysisChainBuilder) analyzeFeeding(ctx context.Context, data *AnalysisData) (*entity.AIAnalysisResult, error) {
	prompt := b.buildFeedingAnalysisPrompt(data)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的婴幼儿喂养专家，擅长分析宝宝的喂养数据并提供专业建议。"),
		schema.UserMessage(prompt),
	}

	response, err := b.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "喂养分析失败", err)
	}

	return b.parseAnalysisResponse(response.Content, entity.AIAnalysisTypeFeeding, data.Baby.ID)
}

// analyzeSleep 分析睡眠数据
func (b *AnalysisChainBuilder) analyzeSleep(ctx context.Context, data *AnalysisData) (*entity.AIAnalysisResult, error) {
	prompt := b.buildSleepAnalysisPrompt(data)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的婴幼儿睡眠专家，擅长分析宝宝的睡眠模式和质量。"),
		schema.UserMessage(prompt),
	}

	response, err := b.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "睡眠分析失败", err)
	}

	return b.parseAnalysisResponse(response.Content, entity.AIAnalysisTypeSleep, data.Baby.ID)
}

// analyzeGrowth 分析成长数据
func (b *AnalysisChainBuilder) analyzeGrowth(ctx context.Context, data *AnalysisData) (*entity.AIAnalysisResult, error) {
	prompt := b.buildGrowthAnalysisPrompt(data)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的儿科医生，擅长评估婴幼儿的生长发育情况。"),
		schema.UserMessage(prompt),
	}

	response, err := b.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "成长分析失败", err)
	}

	return b.parseAnalysisResponse(response.Content, entity.AIAnalysisTypeGrowth, data.Baby.ID)
}

// analyzeHealth 分析健康数据
func (b *AnalysisChainBuilder) analyzeHealth(ctx context.Context, data *AnalysisData) (*entity.AIAnalysisResult, error) {
	prompt := b.buildHealthAnalysisPrompt(data)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的儿科医生，擅长通过多维度数据评估宝宝的健康状况。"),
		schema.UserMessage(prompt),
	}

	response, err := b.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "健康分析失败", err)
	}

	return b.parseAnalysisResponse(response.Content, entity.AIAnalysisTypeHealth, data.Baby.ID)
}

// analyzeBehavior 分析行为数据
func (b *AnalysisChainBuilder) analyzeBehavior(ctx context.Context, data *AnalysisData) (*entity.AIAnalysisResult, error) {
	prompt := b.buildBehaviorAnalysisPrompt(data)

	messages := []*schema.Message{
		schema.SystemMessage("你是一个专业的儿童发展专家，擅长分析婴幼儿的行为模式和发展趋势。"),
		schema.UserMessage(prompt),
	}

	response, err := b.chatModel.Generate(ctx, messages)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "行为分析失败", err)
	}

	return b.parseAnalysisResponse(response.Content, entity.AIAnalysisTypeBehavior, data.Baby.ID)
}

// buildFeedingAnalysisPrompt 构建喂养分析提示
func (b *AnalysisChainBuilder) buildFeedingAnalysisPrompt(data *AnalysisData) string {
	// 解析出生日期
	birthDate, err := time.Parse("2006-01-02", data.Baby.BirthDate)
	if err != nil {
		birthDate = time.Now().AddDate(-1, 0, 0) // 默认1岁
	}
	babyAge := calculateBabyAgeInMonths(birthDate, time.Now())

	prompt := fmt.Sprintf(`请分析以下宝宝的喂养数据：

宝宝信息：
- 月龄：%d个月
- 性别：%s
- 出生日期：%s

分析时间范围：%s 至 %s

喂养记录：
%s

请提供以下分析：
1. 喂养模式评估（规律性、适量性、多样性）
2. 营养摄入分析
3. 喂养时间建议
4. 异常情况识别
5. 总体评分（0-100分）

请以JSON格式返回分析结果，包含：score、insights、alerts、patterns、predictions字段。`,
		babyAge,
		data.Baby.Gender,
		data.Baby.BirthDate,
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02"),
		b.formatFeedingRecords(data.FeedingRecords),
	)

	return prompt
}

// buildSleepAnalysisPrompt 构建睡眠分析提示
func (b *AnalysisChainBuilder) buildSleepAnalysisPrompt(data *AnalysisData) string {
	// 解析出生日期
	birthDate, err := time.Parse("2006-01-02", data.Baby.BirthDate)
	if err != nil {
		birthDate = time.Now().AddDate(-1, 0, 0) // 默认1岁
	}
	babyAge := calculateBabyAgeInMonths(birthDate, time.Now())

	prompt := fmt.Sprintf(`请分析以下宝宝的睡眠数据：

宝宝信息：
- 月龄：%d个月
- 性别：%s

分析时间范围：%s 至 %s

睡眠记录：
%s

请提供以下分析：
1. 睡眠质量评估（连续性、时长、规律）
2. 作息规律分析
3. 睡眠问题识别
4. 改善建议
5. 总体评分（0-100分）

请以JSON格式返回分析结果，包含：score、insights、alerts、patterns、predictions字段。`,
		babyAge,
		data.Baby.Gender,
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02"),
		b.formatSleepRecords(data.SleepRecords),
	)

	return prompt
}

// buildGrowthAnalysisPrompt 构建成长分析提示
func (b *AnalysisChainBuilder) buildGrowthAnalysisPrompt(data *AnalysisData) string {
	// 解析出生日期
	birthDate, err := time.Parse("2006-01-02", data.Baby.BirthDate)
	if err != nil {
		birthDate = time.Now().AddDate(-1, 0, 0) // 默认1岁
	}
	babyAge := calculateBabyAgeInMonths(birthDate, time.Now())

	prompt := fmt.Sprintf(`请分析以下宝宝的成长发育数据：

宝宝信息：
- 月龄：%d个月
- 性别：%s

分析时间范围：%s 至 %s

成长记录：
%s

请提供以下分析：
1. 生长发育评估（身高、体重、头围）
2. 生长曲线分析
3. 与WHO标准对比
4. 发育里程碑评估
5. 总体评分（0-100分）

请以JSON格式返回分析结果，包含：score、insights、alerts、patterns、predictions字段。`,
		babyAge,
		data.Baby.Gender,
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02"),
		b.formatGrowthRecords(data.GrowthRecords),
	)

	return prompt
}

// buildHealthAnalysisPrompt 构建健康分析提示
func (b *AnalysisChainBuilder) buildHealthAnalysisPrompt(data *AnalysisData) string {
	// 解析出生日期
	birthDate, err := time.Parse("2006-01-02", data.Baby.BirthDate)
	if err != nil {
		birthDate = time.Now().AddDate(-1, 0, 0) // 默认1岁
	}
	babyAge := calculateBabyAgeInMonths(birthDate, time.Now())

	prompt := fmt.Sprintf(`请综合分析以下宝宝的健康数据：

宝宝信息：
- 月龄：%d个月
- 性别：%s

分析时间范围：%s 至 %s

数据汇总：
- 喂养记录：%d条
- 睡眠记录：%d条
- 排泄记录：%d条

请提供以下分析：
1. 整体健康状况评估
2. 消化健康分析
3. 行为模式识别
4. 潜在健康问题预警
5. 总体评分（0-100分）

请以JSON格式返回分析结果，包含：score、insights、alerts、patterns、predictions字段。`,
		babyAge,
		data.Baby.Gender,
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02"),
		len(data.FeedingRecords),
		len(data.SleepRecords),
		len(data.DiaperRecords),
	)

	return prompt
}

// buildBehaviorAnalysisPrompt 构建行为分析提示
func (b *AnalysisChainBuilder) buildBehaviorAnalysisPrompt(data *AnalysisData) string {
	// 解析出生日期
	birthDate, err := time.Parse("2006-01-02", data.Baby.BirthDate)
	if err != nil {
		birthDate = time.Now().AddDate(-1, 0, 0) // 默认1岁
	}
	babyAge := calculateBabyAgeInMonths(birthDate, time.Now())

	prompt := fmt.Sprintf(`请分析以下宝宝的行为模式：

宝宝信息：
- 月龄：%d个月
- 性别：%s

分析时间范围：%s 至 %s

行为数据：
%s

请提供以下分析：
1. 行为模式识别
2. 发展里程碑评估
3. 个性化特征分析
4. 行为趋势预测
5. 总体评分（0-100分）

请以JSON格式返回分析结果，包含：score、insights、alerts、patterns、predictions字段。`,
		babyAge,
		data.Baby.Gender,
		data.StartDate.Format("2006-01-02"),
		data.EndDate.Format("2006-01-02"),
		b.formatBehaviorData(data),
	)

	return prompt
}

// buildDailyTipsSystemPrompt 构建每日建议系统提示
func (b *AnalysisChainBuilder) buildDailyTipsSystemPrompt() string {
	return `你是一个专业的育儿专家，擅长根据宝宝的日常数据提供个性化的育儿建议。
请基于提供的数据，生成3-5条实用、具体的育儿建议。
建议应该：
1. 基于实际数据，具有针对性
2. 实用性强，易于执行
3. 考虑宝宝的月龄和发展阶段
4. 包含具体的行动建议
5. 使用友好的语气

请以JSON数组格式返回，每个建议包含：id、icon、title、description、type、priority、action_url字段。`
}

// buildDailyTipsUserPrompt 构建每日建议用户提示
func (b *AnalysisChainBuilder) buildDailyTipsUserPrompt(baby *entity.Baby, data map[string]interface{}) string {
	// 解析出生日期
	birthDate, err := time.Parse("2006-01-02", baby.BirthDate)
	if err != nil {
		birthDate = time.Now().AddDate(-1, 0, 0) // 默认1岁
	}
	babyAge := calculateBabyAgeInMonths(birthDate, time.Now())

	prompt := fmt.Sprintf(`请为以下宝宝生成今日育儿建议：

宝宝信息：
- 月龄：%d个月
- 性别：%s
- 出生日期：%s

最近数据概况：
%s

请基于以上信息，生成个性化的育儿建议。`,
		babyAge,
		baby.Gender,
		baby.BirthDate,
		b.formatRecentData(data),
	)

	return prompt
}

// formatFeedingRecords 格式化喂养记录
func (b *AnalysisChainBuilder) formatFeedingRecords(records []entity.FeedingRecord) string {
	if len(records) == 0 {
		return "暂无喂养记录"
	}

	result := ""
	for i, record := range records {
		if i >= 5 { // 只显示最近5条
			break
		}
		result += fmt.Sprintf("- %s: %s喂养，奶量%dml，时长%d分钟\n",
			time.Unix(record.Time/1000, 0).Format("01-02 15:04"),
			record.FeedingType,
			record.Amount,
			record.Duration/60,
		)
	}
	return result
}

// formatSleepRecords 格式化睡眠记录
func (b *AnalysisChainBuilder) formatSleepRecords(records []entity.SleepRecord) string {
	if len(records) == 0 {
		return "暂无睡眠记录"
	}

	result := ""
	for i, record := range records {
		if i >= 5 { // 只显示最近5条
			break
		}
		endTime := ""
		if record.EndTime != nil {
			endTime = time.Unix(*record.EndTime/1000, 0).Format("01-02 15:04")
		}
		duration := 0
		if record.Duration != nil {
			duration = *record.Duration
		}
		result += fmt.Sprintf("- %s至%s: 睡眠%d小时%d分钟\n",
			time.Unix(record.StartTime/1000, 0).Format("01-02 15:04"),
			endTime,
			duration/3600,
			(duration%3600)/60,
		)
	}
	return result
}

// formatGrowthRecords 格式化成长记录
func (b *AnalysisChainBuilder) formatGrowthRecords(records []entity.GrowthRecord) string {
	if len(records) == 0 {
		return "暂无成长记录"
	}

	result := ""
	for i, record := range records {
		if i >= 3 { // 只显示最近3条
			break
		}
		result += fmt.Sprintf("- %s: 身高%.1fcm，体重%.1fkg，头围%.1fcm\n",
			time.Unix(record.Time/1000, 0).Format("2006-01-02"),
			record.Height,
			record.Weight,
			record.HeadCircumference,
		)
	}
	return result
}

// formatBehaviorData 格式化行为数据
func (b *AnalysisChainBuilder) formatBehaviorData(data *AnalysisData) string {
	result := fmt.Sprintf("喂养记录：%d条\n", len(data.FeedingRecords))
	result += fmt.Sprintf("睡眠记录：%d条\n", len(data.SleepRecords))
	result += fmt.Sprintf("成长记录：%d条\n", len(data.GrowthRecords))
	return result
}

// formatRecentData 格式化近期数据
func (b *AnalysisChainBuilder) formatRecentData(data map[string]interface{}) string {
	result := ""

	if feedingRecords, ok := data["feeding_records"].([]entity.FeedingRecord); ok {
		result += fmt.Sprintf("- 最近喂养：%d次\n", len(feedingRecords))
	}

	if sleepRecords, ok := data["sleep_records"].([]entity.SleepRecord); ok {
		totalSleep := 0
		for _, record := range sleepRecords {
			if record.Duration != nil {
				totalSleep += *record.Duration
			}
		}
		result += fmt.Sprintf("- 总睡眠时长：%d小时\n", totalSleep/3600)
	}

	if growthRecords, ok := data["growth_records"].([]entity.GrowthRecord); ok && len(growthRecords) > 0 {
		latest := growthRecords[len(growthRecords)-1]
		result += fmt.Sprintf("- 最新成长数据：身高%.1fcm，体重%.1fkg\n", latest.Height, latest.Weight)
	}

	return result
}

// parseAnalysisResponse 解析分析响应
func (b *AnalysisChainBuilder) parseAnalysisResponse(content string, analysisType entity.AIAnalysisType, babyID int64) (*entity.AIAnalysisResult, error) {
	var result AnalysisResult
	if err := json.Unmarshal([]byte(content), &result); err != nil {
		return nil, errors.Wrap(errors.InternalError, "解析分析响应失败", err)
	}

	// 转换格式
	analysisResult := &entity.AIAnalysisResult{
		BabyID:       babyID,
		AnalysisType: analysisType,
		Score:        result.Score,
		Insights:     result.Insights,
		Alerts:       result.Alerts,
		Patterns:     result.Patterns,
		Predictions:  result.Predictions,
		Metadata:     result.Metadata,
	}

	return analysisResult, nil
}

// parseDailyTipsResponse 解析每日建议响应
func (b *AnalysisChainBuilder) parseDailyTipsResponse(content string) ([]entity.DailyTip, error) {
	var tips []entity.DailyTip
	if err := json.Unmarshal([]byte(content), &tips); err != nil {
		return nil, errors.Wrap(errors.InternalError, "解析建议响应失败", err)
	}
	return tips, nil
}

// calculateBabyAgeInMonths 计算宝宝月龄
func calculateBabyAgeInMonths(birthDate, currentDate time.Time) int {
	years := currentDate.Year() - birthDate.Year()
	months := int(currentDate.Month()) - int(birthDate.Month())

	totalMonths := years*12 + months
	if currentDate.Day() < birthDate.Day() {
		totalMonths--
	}

	if totalMonths < 0 {
		totalMonths = 0
	}

	return totalMonths
}