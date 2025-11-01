package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/utils"
)

// FeedingRecordService 喂养记录服务
type FeedingRecordService struct {
	*BaseRecordService
	feedingRecordRepo repository.FeedingRecordRepository
	schedulerService  *SchedulerService
}

// NewFeedingRecordService 创建喂养记录服务
func NewFeedingRecordService(
	babyRepo repository.BabyRepository,
	collaboratorRepo repository.BabyCollaboratorRepository,
	feedingRecordRepo repository.FeedingRecordRepository,
	schedulerService *SchedulerService,
	logger *zap.Logger,
) *FeedingRecordService {
	return &FeedingRecordService{
		BaseRecordService: NewBaseRecordService(babyRepo, collaboratorRepo, logger),
		feedingRecordRepo: feedingRecordRepo,
		schedulerService:  schedulerService,
	}
}

// CreateFeedingRecord 创建喂养记录
func (s *FeedingRecordService) CreateFeedingRecord(ctx context.Context, openID string, req *dto.CreateFeedingRecordRequest) (*dto.FeedingRecordDTO, error) {
	// 验证宝宝存在且用户有权限
	if err := s.CheckBabyAccess(ctx, req.BabyID, openID); err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	feedingTime := req.FeedingTime
	if feedingTime == 0 {
		feedingTime = now
	}

	// 将请求参数转换为map存储到Detail中
	detail := entity.FeedingDetail{
		"feedingType": req.FeedingType,
		"amount":      req.Amount,
		"duration":    req.Duration,
		"note":        req.Note,
	}

	// 合并额外的detail信息 (从map中安全地提取值)
	if req.Detail != nil {
		if breastSide, ok := req.Detail["breastSide"].(string); ok && breastSide != "" {
			detail["breastSide"] = breastSide
		}
		if leftTime, ok := req.Detail["leftTime"].(float64); ok && leftTime > 0 {
			detail["leftTime"] = int64(leftTime)
		}
		if rightTime, ok := req.Detail["rightTime"].(float64); ok && rightTime > 0 {
			detail["rightTime"] = int64(rightTime)
		}
		if formulaType, ok := req.Detail["formulaType"].(string); ok && formulaType != "" {
			detail["formulaType"] = formulaType
		}
	}

	record := &entity.FeedingRecord{
		RecordID:    uuid.New().String(),
		FeedingType: req.FeedingType,
		Amount:      utils.DerefInt64(req.Amount),
		Duration:    utils.DerefInt(req.Duration),
		BabyID:      req.BabyID,
		Time:        feedingTime,
		Detail:      detail,
		CreateBy:    openID,
		CreateTime:  now,
		UpdateTime:  now,
	}

	// 处理用户自定义的提醒间隔
	if req.ReminderInterval != nil && *req.ReminderInterval > 0 {
		record.ReminderInterval = req.ReminderInterval

		// 计算下次提醒时间: 喂养时间 + 间隔(分钟)
		nextReminderTime := feedingTime + int64(*req.ReminderInterval*60*1000)
		record.NextReminderTime = &nextReminderTime

		s.logger.Info("设置喂养提醒",
			zap.String("babyID", req.BabyID),
			zap.Int("intervalMinutes", *req.ReminderInterval),
			zap.Int64("nextReminderTime", nextReminderTime))
	}

	if err := s.feedingRecordRepo.Create(ctx, record); err != nil {
		s.logger.Error("保存喂养记录失败",
			zap.String("babyID", req.BabyID),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("喂养记录创建成功",
		zap.String("recordID", record.RecordID),
		zap.String("babyID", record.BabyID),
		zap.String("feedingType", record.FeedingType))

	// 如果设置了提醒时间,添加定时任务到 gocron
	if record.NextReminderTime != nil && s.schedulerService != nil {
		jobTag, err := s.schedulerService.AddFeedingReminderTask(ctx, record)
		if err != nil {
			// 定时任务添加失败不影响记录保存,仅记录警告日志
			s.logger.Warn("添加喂养提醒定时任务失败,用户将无法收到提醒",
				zap.String("recordID", record.RecordID),
				zap.Error(err))
			// 不返回错误,允许记录保存成功
		} else if jobTag != "" {
			s.logger.Info("喂养提醒定时任务已添加",
				zap.String("recordID", record.RecordID),
				zap.String("jobTag", jobTag))
		}
	}

	// 将detail转换回FeedingDetail结构体
	feedingDetail := dto.FeedingDetail{}
	if breastSide, ok := detail["breastSide"].(string); ok {
		feedingDetail.BreastSide = breastSide
	}
	if leftTime, ok := detail["leftTime"].(int64); ok {
		feedingDetail.LeftTime = int(leftTime)
	}
	if rightTime, ok := detail["rightTime"].(int64); ok {
		feedingDetail.RightTime = int(rightTime)
	}
	if formulaType, ok := detail["formulaType"].(string); ok {
		feedingDetail.FormulaType = formulaType
	}

	return &dto.FeedingRecordDTO{
		RecordID:    record.RecordID,
		BabyID:      record.BabyID,
		FeedingType: req.FeedingType,
		Amount:      record.Amount,
		Duration:    record.Duration,
		Detail:      feedingDetail,
		Note:        utils.DerefString(req.Note),
		FeedingTime: record.Time,
		CreateBy:    record.CreateBy,
		CreateTime:  record.CreateTime,
	}, nil
}

// GetFeedingRecords 获取喂养记录列表
func (s *FeedingRecordService) GetFeedingRecords(ctx context.Context, openID string, query *dto.RecordListQuery) ([]dto.FeedingRecordDTO, int64, error) {
	// 验证宝宝访问权限
	if err := s.CheckBabyAccess(ctx, query.BabyID, openID); err != nil {
		return nil, 0, err
	}

	records, total, err := s.feedingRecordRepo.FindByBabyID(
		ctx,
		query.BabyID,
		query.StartTime,
		query.EndTime,
		query.Page,
		query.PageSize,
	)
	if err != nil {
		return nil, 0, err
	}

	result := make([]dto.FeedingRecordDTO, 0, len(records))
	for _, record := range records {
		note, _ := record.Detail["note"].(string)

		// 构建FeedingDetail
		detail := dto.FeedingDetail{}
		if v, ok := record.Detail["breastSide"].(string); ok {
			detail.BreastSide = v
		}
		if v, ok := record.Detail["leftTime"].(float64); ok {
			detail.LeftTime = int(v)
		}
		if v, ok := record.Detail["rightTime"].(float64); ok {
			detail.RightTime = int(v)
		}
		if v, ok := record.Detail["formulaType"].(string); ok {
			detail.FormulaType = v
		}

		result = append(result, dto.FeedingRecordDTO{
			RecordID:    record.RecordID,
			BabyID:      record.BabyID,
			FeedingType: record.FeedingType,
			Amount:      record.Amount,
			Duration:    record.Duration,
			Detail:      detail,
			Note:        note,
			FeedingTime: record.Time,
			CreateBy:    record.CreateBy,
			CreateTime:  record.CreateTime,
		})
	}

	return result, total, nil
}
