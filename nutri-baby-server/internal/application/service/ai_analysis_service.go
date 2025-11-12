package service

import (
	"context"
	"encoding/json"
	"fmt"
	"strconv"
	"strings"
	"time"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/eino/chain"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
	"go.uber.org/zap"
)

// CustomTime 自定义时间类型，支持多种日期格式
type CustomTime time.Time

// UnmarshalJSON 自定义JSON反序列化，支持多种日期格式
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := strings.Trim(string(b), `"`)

	// 尝试不同的日期格式
	layouts := []string{
		"2006-01-02T15:04:05Z07:00", // RFC3339
		"2006-01-02T15:04:05Z",      // RFC3339 without offset
		"2006-01-02T15:04:05",       // ISO 8601 without timezone
		"2006-01-02 15:04:05",       // 日期时间格式
		"2006-01-02",                // 日期格式（最常用）
	}

	var t time.Time
	var err error

	for _, layout := range layouts {
		t, err = time.Parse(layout, s)
		if err == nil {
			*ct = CustomTime(t)
			return nil
		}
	}

	return fmt.Errorf("unable to parse time %s, tried formats: %v", s, layouts)
}

// MarshalJSON 自定义JSON序列化
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	t := time.Time(ct)
	return json.Marshal(t.Format("2006-01-02T15:04:05Z07:00"))
}

// Time 返回 time.Time 类型
func (ct CustomTime) Time() time.Time {
	return time.Time(ct)
}

// Before 时间比较 - 是否在 u 之前
func (ct CustomTime) Before(u CustomTime) bool {
	return time.Time(ct).Before(time.Time(u))
}

// AIAnalysisService AI分析服务接口
type AIAnalysisService interface {
	// 创建AI分析任务
	CreateAnalysis(ctx context.Context, req *CreateAnalysisRequest) (*AnalysisResponse, error)

	// 获取AI分析结果
	GetAnalysisResult(ctx context.Context, analysisID string) (*AnalysisResponse, error)

	// 获取宝宝的最新AI分析结果
	GetLatestAnalysis(ctx context.Context, babyID string, analysisType entity.AIAnalysisType) (*AnalysisResponse, error)

	// 批量分析宝宝数据
	BatchAnalyze(ctx context.Context, babyID string, startDate, endDate time.Time) (*BatchAnalysisResponse, error)

	// 生成每日建议
	GenerateDailyTips(ctx context.Context, babyID string, date time.Time) (*DailyTipsResponse, error)

	// 获取每日建议
	GetDailyTips(ctx context.Context, babyID string, date time.Time) (*DailyTipsResponse, error)

	// 获取AI分析统计
	GetAnalysisStats(ctx context.Context, babyID string) (*AnalysisStatsResponse, error)

	// 处理待分析的AI任务
	ProcessPendingAnalyses(ctx context.Context) error
}

// CreateAnalysisRequest 创建分析请求
type CreateAnalysisRequest struct {
	BabyID       int64                  `json:"baby_id" binding:"required"`
	AnalysisType entity.AIAnalysisType  `json:"analysis_type" binding:"required"`
	StartDate    CustomTime             `json:"start_date" binding:"required"`
	EndDate      CustomTime             `json:"end_date" binding:"required"`
	Options      map[string]interface{} `json:"options,omitempty"`
}

// AnalysisResponse 分析响应
type AnalysisResponse struct {
	AnalysisID string                   `json:"analysis_id"`
	Status     entity.AIAnalysisStatus  `json:"status"`
	Result     *entity.AIAnalysisResult `json:"result,omitempty"`
	CreatedAt  time.Time                `json:"created_at"`
	Message    string                   `json:"message,omitempty"`
}

// BatchAnalysisResponse 批量分析响应
type BatchAnalysisResponse struct {
	Analyses       []*AnalysisResponse `json:"analyses"`
	TotalCount     int                 `json:"total_count"`
	CompletedCount int                 `json:"completed_count"`
	FailedCount    int                 `json:"failed_count"`
}

// DailyTipsResponse 每日建议响应
type DailyTipsResponse struct {
	Tips        []entity.DailyTip `json:"tips"`
	GeneratedAt time.Time         `json:"generated_at"`
	ExpiredAt   time.Time         `json:"expired_at"`
}

// AnalysisStatsResponse 分析统计响应
type AnalysisStatsResponse struct {
	TotalAnalyses      int64               `json:"total_analyses"`
	CompletedAnalyses  int64               `json:"completed_analyses"`
	FailedAnalyses     int64               `json:"failed_analyses"`
	AverageScore       *float64            `json:"average_score"`
	AnalysisTypeCounts map[string]int64    `json:"analysis_type_counts"`
	RecentAnalyses     []*AnalysisResponse `json:"recent_analyses"`
}

// aiAnalysisServiceImpl AI分析服务实现
type aiAnalysisServiceImpl struct {
	aiAnalysisRepo repository.AIAnalysisRepository
	dailyTipsRepo  repository.DailyTipsRepository
	babyRepo       repository.BabyRepository
	feedingRepo    repository.FeedingRecordRepository
	sleepRepo      repository.SleepRecordRepository
	diaperRepo     repository.DiaperRecordRepository
	growthRepo     repository.GrowthRecordRepository
	vaccineRepo    repository.BabyVaccineScheduleRepository
	chainBuilder   *chain.AnalysisChainBuilder
	logger         *zap.Logger
}

// NewAIAnalysisService 创建AI分析服务实例
func NewAIAnalysisService(
	aiAnalysisRepo repository.AIAnalysisRepository,
	dailyTipsRepo repository.DailyTipsRepository,
	babyRepo repository.BabyRepository,
	feedingRepo repository.FeedingRecordRepository,
	sleepRepo repository.SleepRecordRepository,
	diaperRepo repository.DiaperRecordRepository,
	growthRepo repository.GrowthRecordRepository,
	vaccineRepo repository.BabyVaccineScheduleRepository,
	chainBuilder *chain.AnalysisChainBuilder,
	logger *zap.Logger,
) AIAnalysisService {
	return &aiAnalysisServiceImpl{
		aiAnalysisRepo: aiAnalysisRepo,
		dailyTipsRepo:  dailyTipsRepo,
		babyRepo:       babyRepo,
		feedingRepo:    feedingRepo,
		sleepRepo:      sleepRepo,
		diaperRepo:     diaperRepo,
		growthRepo:     growthRepo,
		vaccineRepo:    vaccineRepo,
		chainBuilder:   chainBuilder,
		logger:         logger,
	}
}

// CreateAnalysis 创建AI分析任务
func (s *aiAnalysisServiceImpl) CreateAnalysis(ctx context.Context, req *CreateAnalysisRequest) (*AnalysisResponse, error) {
	// 验证宝宝是否存在
	_, err := s.babyRepo.FindByID(ctx, req.BabyID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "获取宝宝信息失败", err)
	}

	// 验证时间范围
	if req.EndDate.Before(req.StartDate) {
		return nil, errors.New(errors.ParamError, "结束日期不能早于开始日期")
	}

	// 创建分析记录
	analysis := &entity.AIAnalysis{
		BabyID:       req.BabyID,
		AnalysisType: req.AnalysisType,
		Status:       entity.AIAnalysisStatusPending,
		StartDate:    req.StartDate.Time(),
		EndDate:      req.EndDate.Time(),
	}

	if err := s.aiAnalysisRepo.Create(ctx, analysis); err != nil {
		return nil, errors.Wrap(errors.InternalError, "创建分析任务失败", err)
	}

	s.logger.Info("创建AI分析任务",
		zap.Int64("analysis_id", analysis.ID),
		zap.Int64("baby_id", req.BabyID),
		zap.String("analysis_type", string(req.AnalysisType)),
		zap.Time("start_date", req.StartDate.Time()),
		zap.Time("end_date", req.EndDate.Time()),
	)

	return &AnalysisResponse{
		AnalysisID: strconv.FormatInt(analysis.ID, 10),
		Status:     analysis.Status,
		CreatedAt:  analysis.CreatedAt,
		Message:    "AI分析任务已创建，正在处理中...",
	}, nil
}

// GetAnalysisResult 获取AI分析结果
func (s *aiAnalysisServiceImpl) GetAnalysisResult(ctx context.Context, analysisID string) (*AnalysisResponse, error) {
	// 转换ID类型
	id, err := strconv.ParseInt(analysisID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(errors.ParamError, "无效的分析ID", err)
	}

	analysis, err := s.aiAnalysisRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "获取分析记录失败", err)
	}

	response := &AnalysisResponse{
		AnalysisID: strconv.FormatInt(analysis.ID, 10),
		Status:     analysis.Status,
		CreatedAt:  analysis.CreatedAt,
	}

	// 如果分析已完成，解析结果
	if analysis.Status == entity.AIAnalysisStatusCompleted && analysis.Result != "" {
		var result entity.AIAnalysisResult
		if err := json.Unmarshal([]byte(analysis.Result), &result); err != nil {
			s.logger.Error("解析分析结果失败",
				zap.String("analysis_id", analysisID),
				zap.Error(err),
			)
			return response, nil
		}
		response.Result = &result
	}

	return response, nil
}

// GetLatestAnalysis 获取宝宝的最新AI分析结果
func (s *aiAnalysisServiceImpl) GetLatestAnalysis(ctx context.Context, babyID string, analysisType entity.AIAnalysisType) (*AnalysisResponse, error) {
	// 转换ID类型
	id, err := strconv.ParseInt(babyID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(errors.ParamError, "无效的宝宝ID", err)
	}

	analysis, err := s.aiAnalysisRepo.GetLatestByBabyIDAndType(ctx, id, analysisType)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			return nil, errors.New(errors.NotFound, "未找到对应的AI分析结果")
		}
		return nil, errors.Wrap(errors.InternalError, "获取最新分析结果失败", err)
	}

	response := &AnalysisResponse{
		AnalysisID: strconv.FormatInt(analysis.ID, 10),
		Status:     analysis.Status,
		CreatedAt:  analysis.CreatedAt,
	}

	// 解析结果
	if analysis.Status == entity.AIAnalysisStatusCompleted && analysis.Result != "" {
		var result entity.AIAnalysisResult
		if err := json.Unmarshal([]byte(analysis.Result), &result); err != nil {
			s.logger.Error("解析分析结果失败",
				zap.Int64("analysis_id", analysis.ID),
				zap.Error(err),
			)
			return response, nil
		}
		response.Result = &result
	}

	return response, nil
}

// BatchAnalyze 批量分析宝宝数据
func (s *aiAnalysisServiceImpl) BatchAnalyze(ctx context.Context, babyID string, startDate, endDate time.Time) (*BatchAnalysisResponse, error) {
	// 转换ID类型
	id, err := strconv.ParseInt(babyID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(errors.ParamError, "无效的宝宝ID", err)
	}

	analysisTypes := []entity.AIAnalysisType{
		entity.AIAnalysisTypeFeeding,
		entity.AIAnalysisTypeSleep,
		entity.AIAnalysisTypeGrowth,
		entity.AIAnalysisTypeHealth,
	}

	var analyses []*AnalysisResponse
	var completedCount, failedCount int

	for _, analysisType := range analysisTypes {
		// 创建分析任务
		req := &CreateAnalysisRequest{
			BabyID:       id, // 使用转换后的ID
			AnalysisType: analysisType,
			StartDate:    CustomTime(startDate),
			EndDate:      CustomTime(endDate),
		}

		analysis, err := s.CreateAnalysis(ctx, req)
		if err != nil {
			s.logger.Error("创建分析任务失败",
				zap.String("baby_id", babyID),
				zap.String("analysis_type", string(analysisType)),
				zap.Error(err),
			)
			failedCount++
			continue
		}

		analyses = append(analyses, analysis)

		// 立即处理分析任务
		analysisID, _ := strconv.ParseInt(analysis.AnalysisID, 10, 64)
		if err := s.processAnalysis(ctx, analysisID); err != nil {
			s.logger.Error("处理分析任务失败",
				zap.String("analysis_id", analysis.AnalysisID),
				zap.Error(err),
			)
			failedCount++
		} else {
			completedCount++
		}
	}

	return &BatchAnalysisResponse{
		Analyses:       analyses,
		TotalCount:     len(analyses),
		CompletedCount: completedCount,
		FailedCount:    failedCount,
	}, nil
}

// GenerateDailyTips 生成每日建议
func (s *aiAnalysisServiceImpl) GenerateDailyTips(ctx context.Context, babyID string, date time.Time) (*DailyTipsResponse, error) {
	// 转换ID类型
	id, err := strconv.ParseInt(babyID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(errors.ParamError, "无效的宝宝ID", err)
	}

	// 检查是否已存在当日建议
	existingTips, err := s.dailyTipsRepo.GetByBabyIDAndDate(ctx, id, date)
	if err == nil && existingTips != nil {
		// 返回现有建议
		return &DailyTipsResponse{
			Tips:        existingTips.Tips,
			GeneratedAt: existingTips.CreatedAt,
			ExpiredAt:   existingTips.ExpiredAt,
		}, nil
	}

	// 获取宝宝信息
	baby, err := s.babyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "获取宝宝信息失败", err)
	}

	// 获取最近7天的数据用于分析
	startDate := date.AddDate(0, 0, -7)
	endDate := date

	// 收集数据
	data := make(map[string]interface{})

	// 转换时间为时间戳
	startTime := startDate.Unix() * 1000
	endTime := endDate.Unix() * 1000

	// 喂养数据
	feedingRecords, _, err := s.feedingRepo.FindByBabyID(ctx, id, startTime, endTime, 1, 100)
	if err == nil {
		data["feeding_records"] = feedingRecords
	}

	// 睡眠数据
	sleepRecords, _, err := s.sleepRepo.FindByBabyID(ctx, id, startTime, endTime, 1, 100)
	if err == nil {
		data["sleep_records"] = sleepRecords
	}

	// 成长数据
	growthRecords, _, err := s.growthRepo.FindByBabyID(ctx, id, startTime, endTime, 1, 100)
	if err == nil {
		data["growth_records"] = growthRecords
	}

	// 使用Eino链生成建议
	tips, err := s.chainBuilder.GenerateDailyTips(ctx, baby, data)
	if err != nil {
		s.logger.Error("生成每日建议失败",
			zap.String("baby_id", babyID),
			zap.Error(err),
		)
		return nil, errors.Wrap(errors.InternalError, "生成每日建议失败", err)
	}

	// 保存建议
	dailyTips := &entity.DailyTips{
		BabyID:    id,
		Date:      date,
		Tips:      tips,
		ExpiredAt: date.AddDate(0, 0, 1), // 24小时后过期
	}

	if err := s.dailyTipsRepo.Create(ctx, dailyTips); err != nil {
		return nil, errors.Wrap(errors.InternalError, "保存每日建议失败", err)
	}

	return &DailyTipsResponse{
		Tips:        tips,
		GeneratedAt: dailyTips.CreatedAt,
		ExpiredAt:   dailyTips.ExpiredAt,
	}, nil
}

// GetDailyTips 获取每日建议
func (s *aiAnalysisServiceImpl) GetDailyTips(ctx context.Context, babyID string, date time.Time) (*DailyTipsResponse, error) {
	// 转换ID类型
	id, err := strconv.ParseInt(babyID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(errors.ParamError, "无效的宝宝ID", err)
	}

	tips, err := s.dailyTipsRepo.GetByBabyIDAndDate(ctx, id, date)
	if err != nil {
		// 如果不存在，尝试生成新的建议
		return s.GenerateDailyTips(ctx, babyID, date)
	}

	return &DailyTipsResponse{
		Tips:        tips.Tips,
		GeneratedAt: tips.CreatedAt,
		ExpiredAt:   tips.ExpiredAt,
	}, nil
}

// GetAnalysisStats 获取AI分析统计
func (s *aiAnalysisServiceImpl) GetAnalysisStats(ctx context.Context, babyID string) (*AnalysisStatsResponse, error) {
	// 转换ID类型
	id, err := strconv.ParseInt(babyID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(errors.ParamError, "无效的宝宝ID", err)
	}

	stats, err := s.aiAnalysisRepo.GetAnalysisStats(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "获取分析统计失败", err)
	}

	// 转换响应格式
	var recentAnalyses []*AnalysisResponse
	for _, analysis := range stats.RecentAnalyses {
		response := &AnalysisResponse{
			AnalysisID: strconv.FormatInt(analysis.ID, 10),
			Status:     analysis.Status,
			CreatedAt:  analysis.CreatedAt,
		}

		if analysis.Status == entity.AIAnalysisStatusCompleted && analysis.Result != "" {
			var result entity.AIAnalysisResult
			if err := json.Unmarshal([]byte(analysis.Result), &result); err == nil {
				response.Result = &result
			}
		}

		recentAnalyses = append(recentAnalyses, response)
	}

	return &AnalysisStatsResponse{
		TotalAnalyses:      stats.TotalAnalyses,
		CompletedAnalyses:  stats.CompletedAnalyses,
		FailedAnalyses:     stats.FailedAnalyses,
		AverageScore:       stats.AverageScore,
		AnalysisTypeCounts: stats.AnalysisTypeCounts,
		RecentAnalyses:     recentAnalyses,
	}, nil
}

// ProcessPendingAnalyses 处理待分析的AI任务
func (s *aiAnalysisServiceImpl) ProcessPendingAnalyses(ctx context.Context) error {
	pendingAnalyses, err := s.aiAnalysisRepo.GetPendingAnalyses(ctx, 10)
	if err != nil {
		return errors.Wrap(errors.InternalError, "获取待分析任务失败", err)
	}

	for _, analysis := range pendingAnalyses {
		if err := s.processAnalysis(ctx, analysis.ID); err != nil {
			s.logger.Error("处理分析任务失败",
				zap.Int64("analysis_id", analysis.ID),
				zap.Error(err),
			)
			// 继续处理其他任务
			continue
		}
	}

	return nil
}

// processAnalysis 处理单个分析任务
func (s *aiAnalysisServiceImpl) processAnalysis(ctx context.Context, analysisID int64) error {
	// 更新状态为分析中
	if err := s.aiAnalysisRepo.UpdateStatus(ctx, analysisID, entity.AIAnalysisStatusAnalyzing); err != nil {
		return errors.Wrap(errors.InternalError, "更新分析状态失败", err)
	}

	// 获取分析记录
	analysis, err := s.aiAnalysisRepo.GetByID(ctx, analysisID)
	if err != nil {
		return errors.Wrap(errors.InternalError, "获取分析记录失败", err)
	}

	// 收集分析所需数据
	data, err := s.collectAnalysisData(ctx, analysis)
	if err != nil {
		s.updateAnalysisStatus(ctx, analysisID, entity.AIAnalysisStatusFailed, fmt.Sprintf("数据收集失败: %v", err))
		return errors.Wrap(errors.InternalError, "收集分析数据失败", err)
	}

	// 使用Eino链进行分析
	result, err := s.chainBuilder.Analyze(ctx, analysis, data)
	if err != nil {
		s.updateAnalysisStatus(ctx, analysisID, entity.AIAnalysisStatusFailed, fmt.Sprintf("分析失败: %v", err))
		return errors.Wrap(errors.InternalError, "AI分析失败", err)
	}

	// 序列化结果
	resultJSON, err := json.Marshal(result)
	if err != nil {
		s.updateAnalysisStatus(ctx, analysisID, entity.AIAnalysisStatusFailed, fmt.Sprintf("结果序列化失败: %v", err))
		return errors.Wrap(errors.ParamError, "序列化分析结果失败", err)
	}

	// 更新分析结果
	if err := s.aiAnalysisRepo.UpdateResult(ctx, analysisID, string(resultJSON), entity.AIAnalysisStatusCompleted); err != nil {
		return errors.Wrap(errors.InternalError, "更新分析结果失败", err)
	}

	s.logger.Info("AI分析任务完成",
		zap.Int64("analysis_id", analysisID),
		zap.String("analysis_type", string(analysis.AnalysisType)),
		zap.Float64("score", result.Score),
	)

	return nil
}

// collectAnalysisData 收集分析所需数据
func (s *aiAnalysisServiceImpl) collectAnalysisData(ctx context.Context, analysis *entity.AIAnalysis) (map[string]interface{}, error) {
	data := make(map[string]interface{})

	// 获取宝宝信息
	baby, err := s.babyRepo.FindByID(ctx, analysis.BabyID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "获取宝宝信息失败", err)
	}
	data["baby"] = baby

	// 根据分析类型收集相应数据
	switch analysis.AnalysisType {
	case entity.AIAnalysisTypeFeeding:
		// 转换时间为时间戳
		startTime := analysis.StartDate.Unix() * 1000
		endTime := analysis.EndDate.Unix() * 1000
		feedingRecords, _, err := s.feedingRepo.FindByBabyID(ctx, analysis.BabyID, startTime, endTime, 1, 1000)
		if err != nil {
			return nil, errors.Wrap(errors.DatabaseError, "获取喂养记录失败", err)
		}
		data["feeding_records"] = feedingRecords

	case entity.AIAnalysisTypeSleep:
		// 转换时间为时间戳
		startTime := analysis.StartDate.Unix() * 1000
		endTime := analysis.EndDate.Unix() * 1000
		sleepRecords, _, err := s.sleepRepo.FindByBabyID(ctx, analysis.BabyID, startTime, endTime, 1, 1000)
		if err != nil {
			return nil, errors.Wrap(errors.DatabaseError, "获取睡眠记录失败", err)
		}
		data["sleep_records"] = sleepRecords

	case entity.AIAnalysisTypeGrowth:
		// 转换时间为时间戳
		startTime := analysis.StartDate.Unix() * 1000
		endTime := analysis.EndDate.Unix() * 1000
		growthRecords, _, err := s.growthRepo.FindByBabyID(ctx, analysis.BabyID, startTime, endTime, 1, 1000)
		if err != nil {
			return nil, errors.Wrap(errors.DatabaseError, "获取成长记录失败", err)
		}
		data["growth_records"] = growthRecords

	case entity.AIAnalysisTypeHealth:
		// 健康分析需要多种数据
		startTime := analysis.StartDate.Unix() * 1000
		endTime := analysis.EndDate.Unix() * 1000
		feedingRecords, _, _ := s.feedingRepo.FindByBabyID(ctx, analysis.BabyID, startTime, endTime, 1, 1000)
		sleepRecords, _, _ := s.sleepRepo.FindByBabyID(ctx, analysis.BabyID, startTime, endTime, 1, 1000)
		diaperRecords, _, _ := s.diaperRepo.FindByBabyID(ctx, analysis.BabyID, startTime, endTime, 1, 1000)

		data["feeding_records"] = feedingRecords
		data["sleep_records"] = sleepRecords
		data["diaper_records"] = diaperRecords
	}

	// 添加成长数据（所有分析类型都需要）
	startTime := analysis.StartDate.Unix() * 1000
	endTime := analysis.EndDate.Unix() * 1000
	growthRecords, _, err := s.growthRepo.FindByBabyID(ctx, analysis.BabyID, startTime, endTime, 1, 1000)
	if err == nil {
		data["growth_records"] = growthRecords
	}

	// 添加疫苗数据
	vaccineRecords, err := s.vaccineRepo.FindByBabyID(ctx, analysis.BabyID, 1, 1000)
	if err == nil {
		data["vaccine_records"] = vaccineRecords
	}

	return data, nil
}

// updateAnalysisStatus 更新分析状态
func (s *aiAnalysisServiceImpl) updateAnalysisStatus(ctx context.Context, analysisID int64, status entity.AIAnalysisStatus, message string) {
	if err := s.aiAnalysisRepo.UpdateStatus(ctx, analysisID, status); err != nil {
		s.logger.Error("更新分析状态失败",
			zap.Int64("analysis_id", analysisID),
			zap.String("status", string(status)),
			zap.Error(err),
		)
	}
}
