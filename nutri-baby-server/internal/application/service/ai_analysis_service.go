package service

import (
	"context"
	"encoding/json"
	"strconv"
	"time"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/eino/chain"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
	"go.uber.org/zap"
)

// 请求和响应类型定义

// AIAnalysisService AI分析服务接口
type AIAnalysisService interface {
	// CreateAnalysis 创建分析任务
	CreateAnalysis(ctx context.Context, req *dto.CreateAnalysisRequest) (*dto.AnalysisResponse, error)

	// GenerateDailyTips 生成每日建议
	GenerateDailyTips(ctx context.Context, babyID string, date time.Time) (*dto.DailyTipsResponse, error)

	// ProcessPendingAnalyses 处理待分析的任务
	ProcessPendingAnalyses(ctx context.Context) error

	// GetAnalysisResult 获取分析结果
	GetAnalysisResult(ctx context.Context, analysisID string) (*dto.AnalysisResponse, error)

	// GetAnalysisStatus 获取分析状态（用于轮询）
	GetAnalysisStatus(ctx context.Context, analysisID string) (*dto.AnalysisStatusResponse, error)

	// GetLatestAnalysis 获取最新分析
	GetLatestAnalysis(ctx context.Context, babyID string, analysisType entity.AIAnalysisType) (*dto.AnalysisResponse, error)

	// BatchAnalyze 批量分析
	BatchAnalyze(ctx context.Context, req *dto.BatchAnalysisRequest) (*dto.BatchAnalysisResponse, error)

	// GetDailyTips 获取每日建议
	GetDailyTips(ctx context.Context, babyID string, date time.Time) (*dto.DailyTipsResponse, error)

	// GetAnalysisStats 获取分析统计
	GetAnalysisStats(ctx context.Context, babyID string, days int) (*dto.AnalysisStatsResponse, error)
}

// aiAnalysisServiceImpl AI分析服务实现
type aiAnalysisServiceImpl struct {
	aiAnalysisRepo repository.AIAnalysisRepository
	dailyTipsRepo  repository.DailyTipsRepository
	babyRepo       repository.BabyRepository
	chainBuilder   *chain.AnalysisChainBuilder
	logger         *zap.Logger
}

// NewAIAnalysisService 创建AI分析服务实例
func NewAIAnalysisService(
	aiAnalysisRepo repository.AIAnalysisRepository,
	dailyTipsRepo repository.DailyTipsRepository,
	babyRepo repository.BabyRepository,
	chainBuilder *chain.AnalysisChainBuilder,
	logger *zap.Logger,
) AIAnalysisService {
	return &aiAnalysisServiceImpl{
		aiAnalysisRepo: aiAnalysisRepo,
		dailyTipsRepo:  dailyTipsRepo,
		babyRepo:       babyRepo,
		chainBuilder:   chainBuilder,
		logger:         logger,
	}
}

// CreateAnalysis 创建分析任务（异步模式）
func (s *aiAnalysisServiceImpl) CreateAnalysis(ctx context.Context, req *dto.CreateAnalysisRequest) (*dto.AnalysisResponse, error) {
	// 验证宝宝是否存在
	_, err := s.babyRepo.FindByID(ctx, req.BabyID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "获取宝宝信息失败", err)
	}

	// 验证时间范围
	if req.EndDate.Before(req.StartDate.Time) {
		return nil, errors.New(errors.ParamError, "结束日期不能早于开始日期")
	}

	// 创建分析记录
	analysis := &entity.AIAnalysis{
		BabyID:       req.BabyID,
		AnalysisType: req.AnalysisType,
		Status:       entity.AIAnalysisStatusPending,
		StartDate:    req.StartDate.Time,
		EndDate:      req.EndDate.Time,
	}

	if err := s.aiAnalysisRepo.Create(ctx, analysis); err != nil {
		return nil, errors.Wrap(errors.InternalError, "创建分析任务失败", err)
	}

	// 异步处理分析任务
	go func() {
		// 使用新的context，避免原context被取消
		bgCtx := context.Background()

		if err := s.processAnalysis(bgCtx, analysis.ID); err != nil {
			s.logger.Error("AI分析失败",
				zap.Int64("analysis_id", analysis.ID),
				zap.Error(err),
			)
			// 更新状态为失败
			s.aiAnalysisRepo.UpdateStatus(bgCtx, analysis.ID, entity.AIAnalysisStatusFailed)
		}
	}()

	// 立即返回任务ID和pending状态
	return &dto.AnalysisResponse{
		AnalysisID: analysis.ID, // 直接使用int64类型
		Status:     entity.AIAnalysisStatusPending,
		CreatedAt:  analysis.CreatedAt,
	}, nil
}

// GenerateDailyTips 生成每日建议
func (s *aiAnalysisServiceImpl) GenerateDailyTips(ctx context.Context, babyID string, date time.Time) (*dto.DailyTipsResponse, error) {
	// 转换ID类型
	id, err := strconv.ParseInt(babyID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(errors.ParamError, "无效的宝宝ID", err)
	}

	// 优先检查是否已存在当日建议，如果存在直接返回
	existingTips, err := s.dailyTipsRepo.GetByBabyIDAndDate(ctx, id, date)
	if err == nil && existingTips != nil {
		s.logger.Info("返回已存在的每日建议",
			zap.String("baby_id", babyID),
			zap.String("date", date.Format("2006-01-02")),
		)
		return &dto.DailyTipsResponse{
			Tips:        existingTips.Tips,
			GeneratedAt: existingTips.CreatedAt,
			ExpiredAt:   existingTips.ExpiredAt,
		}, nil
	}

	// 记录查询错误但继续生成（可能是记录不存在）
	if err != nil {
		s.logger.Debug("查询已存在建议时出错，将生成新建议",
			zap.String("baby_id", babyID),
			zap.Error(err),
		)
	}

	// 获取宝宝信息
	baby, err := s.babyRepo.FindByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "获取宝宝信息失败", err)
	}

	s.logger.Info("开始生成新的每日建议",
		zap.String("baby_id", babyID),
		zap.String("date", date.Format("2006-01-02")),
	)
	// 使用分析链生成建议
	tips, err := s.chainBuilder.GenerateDailyTips(ctx, baby, date)
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

	s.logger.Info("成功生成并保存每日建议",
		zap.String("baby_id", babyID),
		zap.Int("tips_count", len(tips)),
	)

	return &dto.DailyTipsResponse{
		Tips:        tips,
		GeneratedAt: dailyTips.CreatedAt,
		ExpiredAt:   dailyTips.ExpiredAt,
	}, nil
}

// ProcessPendingAnalyses 处理待分析的任务
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
			// 更新状态为失败
			s.aiAnalysisRepo.UpdateStatus(ctx, analysis.ID, entity.AIAnalysisStatusFailed)
			continue
		}
	}

	return nil
}

// GetAnalysisResult 获取分析结果
func (s *aiAnalysisServiceImpl) GetAnalysisResult(ctx context.Context, analysisID string) (*dto.AnalysisResponse, error) {
	id, err := strconv.ParseInt(analysisID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(errors.ParamError, "无效的分析ID", err)
	}

	analysis, err := s.aiAnalysisRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "获取分析记录失败", err)
	}

	response := &dto.AnalysisResponse{
		AnalysisID: id, // 使用int64类型的id
		Status:     analysis.Status,
		CreatedAt:  analysis.CreatedAt,
	}

	// 如果分析已完成，解析结果
	if analysis.Status == entity.AIAnalysisStatusCompleted && analysis.Result != "" {
		var result entity.AIAnalysisResult
		if err := json.Unmarshal([]byte(analysis.Result), &result); err != nil {
			s.logger.Error("解析分析结果失败",
				zap.Int64("analysis_id", id),
				zap.Error(err),
			)
			return response, nil
		}
		response.Result = &result
	}

	return response, nil
}

// GetAnalysisStatus 获取分析状态（用于轮询）
func (s *aiAnalysisServiceImpl) GetAnalysisStatus(ctx context.Context, analysisID string) (*dto.AnalysisStatusResponse, error) {
	id, err := strconv.ParseInt(analysisID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(errors.ParamError, "无效的分析ID", err)
	}

	analysis, err := s.aiAnalysisRepo.GetByID(ctx, id)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "获取分析记录失败", err)
	}

	// 根据状态计算进度和消息
	var progress int
	var message string

	switch analysis.Status {
	case entity.AIAnalysisStatusPending:
		progress = 10
		message = "分析任务已创建，等待处理"
	case entity.AIAnalysisStatusAnalyzing:
		progress = 50
		message = "AI正在分析数据中..."
	case entity.AIAnalysisStatusCompleted:
		progress = 100
		message = "分析完成"
	case entity.AIAnalysisStatusFailed:
		progress = 0
		message = "分析失败，请重试"
	default:
		progress = 0
		message = "未知状态"
	}

	return &dto.AnalysisStatusResponse{
		AnalysisID: analysisID,
		Status:     analysis.Status,
		Progress:   progress,
		Message:    message,
		UpdatedAt:  analysis.UpdatedAt,
	}, nil
}

// GetLatestAnalysis 获取最新分析
func (s *aiAnalysisServiceImpl) GetLatestAnalysis(ctx context.Context, babyID string, analysisType entity.AIAnalysisType) (*dto.AnalysisResponse, error) {
	id, err := strconv.ParseInt(babyID, 10, 64)
	if err != nil {
		return nil, errors.Wrap(errors.ParamError, "无效的宝宝ID", err)
	}

	analysis, err := s.aiAnalysisRepo.GetLatestByBabyIDAndType(ctx, id, analysisType)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "获取最新分析失败", err)
	}

	return s.GetAnalysisResult(ctx, strconv.FormatInt(analysis.ID, 10))
}

// BatchAnalyze 批量分析（异步模式）
func (s *aiAnalysisServiceImpl) BatchAnalyze(ctx context.Context, req *dto.BatchAnalysisRequest) (*dto.BatchAnalysisResponse, error) {
	// 验证宝宝是否存在
	_, err := s.babyRepo.FindByID(ctx, req.BabyID)
	if err != nil {
		return nil, errors.Wrap(errors.NotFound, "获取宝宝信息失败", err)
	}

	// 验证时间范围
	if req.EndDate.Before(req.StartDate.Time) {
		return nil, errors.New(errors.ParamError, "结束日期不能早于开始日期")
	}

	var results []dto.AnalysisResponse

	// 如果没有指定分析类型，使用所有类型
	analysisTypes := req.AnalysisTypes
	if len(analysisTypes) == 0 {
		analysisTypes = []entity.AIAnalysisType{
			entity.AIAnalysisTypeFeeding,
			entity.AIAnalysisTypeSleep,
			entity.AIAnalysisTypeGrowth,
			entity.AIAnalysisTypeHealth,
			entity.AIAnalysisTypeBehavior,
		}
	}

	for _, analysisType := range analysisTypes {
		// 创建分析记录
		analysis := &entity.AIAnalysis{
			BabyID:       req.BabyID,
			AnalysisType: analysisType,
			Status:       entity.AIAnalysisStatusPending,
			StartDate:    req.StartDate.Time,
			EndDate:      req.EndDate.Time,
		}

		if err := s.aiAnalysisRepo.Create(ctx, analysis); err != nil {
			s.logger.Error("创建批量分析任务失败",
				zap.Int64("baby_id", req.BabyID),
				zap.String("analysis_type", string(analysisType)),
				zap.Error(err),
			)
			continue
		}

		// 异步处理分析任务
		analysisID := analysis.ID
		go func(id int64, aType entity.AIAnalysisType) {
			bgCtx := context.Background()

			if err := s.processAnalysis(bgCtx, id); err != nil {
				s.logger.Error("批量AI分析失败",
					zap.Int64("analysis_id", id),
					zap.String("analysis_type", string(aType)),
					zap.Error(err),
				)
				s.aiAnalysisRepo.UpdateStatus(bgCtx, id, entity.AIAnalysisStatusFailed)
			}
		}(analysisID, analysisType)

		// 添加到结果列表
		results = append(results, dto.AnalysisResponse{
			AnalysisID: analysis.ID, // 使用int64类型
			Status:     entity.AIAnalysisStatusPending,
			CreatedAt:  analysis.CreatedAt,
		})
	}

	return &dto.BatchAnalysisResponse{
		TotalCount:     len(results),
		Analyses:       results,
		CompletedCount: 0, // 初始时都是pending状态
		FailedCount:    0,
	}, nil
}

// GetDailyTips 获取每日建议
func (s *aiAnalysisServiceImpl) GetDailyTips(ctx context.Context, babyID string, date time.Time) (*dto.DailyTipsResponse, error) {
	return s.GenerateDailyTips(ctx, babyID, date)
}

// GetAnalysisStats 获取分析统计
func (s *aiAnalysisServiceImpl) GetAnalysisStats(ctx context.Context, babyID string, days int) (*dto.AnalysisStatsResponse, error) {
	// 简化实现，返回基本统计信息
	stats := &dto.AnalysisStatsResponse{
		TotalAnalyses: 0,
		ByType:        make(map[entity.AIAnalysisType]int),
		ByStatus:      make(map[entity.AIAnalysisStatus]int),
		AvgScore:      0.0,
		Trends:        []dto.AnalysisTrend{},
	}

	// TODO: 实现完整的统计功能，需要扩展 repository 接口
	s.logger.Info("GetAnalysisStats 调用",
		zap.String("baby_id", babyID),
		zap.Int("days", days),
	)

	return stats, nil
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

	// 使用分析链进行分析
	result, err := s.chainBuilder.Analyze(ctx, analysis)
	if err != nil {
		return errors.Wrap(errors.InternalError, "AI分析失败", err)
	}

	// 序列化结果
	resultJSON, err := json.Marshal(result)
	if err != nil {
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
