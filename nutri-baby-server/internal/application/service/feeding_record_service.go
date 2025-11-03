package service

import (
	"context"
	"encoding/json"
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

	// 将 map[string]any 转换为 FeedingDetail 结构体
	var feedingDetail dto.FeedingDetail
	if req.Detail != nil {
		// 使用 JSON 序列化/反序列化进行类型安全转换
		detailBytes, err := json.Marshal(req.Detail)
		if err != nil {
			s.logger.Error("Detail序列化失败", zap.Error(err))
			return nil, err
		}
		if err := json.Unmarshal(detailBytes, &feedingDetail); err != nil {
			s.logger.Error("Detail反序列化失败", zap.Error(err))
			return nil, err
		}
	}

	// 确保 Type 字段与 FeedingType 一致
	feedingDetail.Type = req.FeedingType

	// 将 FeedingDetail 转换为 map 存储到数据库
	detailMap := make(entity.FeedingDetail)
	detailBytes, _ := json.Marshal(feedingDetail)
	_ = json.Unmarshal(detailBytes, &detailMap)

	record := &entity.FeedingRecord{
		RecordID:           uuid.New().String(),
		FeedingType:        req.FeedingType,
		Amount:             utils.DerefInt64(req.Amount),
		Duration:           utils.DerefInt(req.Duration),
		BabyID:             req.BabyID,
		Time:               feedingTime,
		Detail:             detailMap,
		CreateBy:           openID,
		CreateTime:         now,
		UpdateTime:         now,
		ActualCompleteTime: req.ActualCompleteTime, // 记录实际完成时间
	}

	// 处理用户自定义的提醒间隔
	if req.ReminderInterval != nil && *req.ReminderInterval > 0 {
		record.ReminderInterval = req.ReminderInterval

		// 计算下次提醒时间: 使用实际完成时间(如果有)，否则使用喂养时间
		// 这样可以确保即使用户延迟记录,提醒时间也是准确的
		baseTime := feedingTime
		if req.ActualCompleteTime != nil {
			baseTime = *req.ActualCompleteTime
		}
		nextReminderTime := baseTime + int64(*req.ReminderInterval*60*1000)
		record.NextReminderTime = &nextReminderTime

		s.logger.Info("设置喂养提醒",
			zap.String("babyID", req.BabyID),
			zap.Int("intervalMinutes", *req.ReminderInterval),
			zap.Int64("baseTime", baseTime),
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

	return &dto.FeedingRecordDTO{
		RecordID:           record.RecordID,
		BabyID:             record.BabyID,
		FeedingType:        req.FeedingType,
		Amount:             record.Amount,
		Duration:           record.Duration,
		Detail:             feedingDetail, // 使用强类型结构体
		Note:               utils.DerefString(req.Note),
		FeedingTime:        record.Time,
		ActualCompleteTime: record.ActualCompleteTime,
		CreateBy:           record.CreateBy,
		CreateTime:         record.CreateTime,
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
		// 将数据库的 map 转换为 FeedingDetail 结构体
		var feedingDetail dto.FeedingDetail
		if record.Detail != nil {
			detailBytes, err := json.Marshal(record.Detail)
			if err != nil {
				s.logger.Warn("Detail序列化失败,使用空Detail",
					zap.String("recordID", record.RecordID),
					zap.Error(err))
				feedingDetail = dto.FeedingDetail{Type: record.FeedingType}
			} else {
				if err := json.Unmarshal(detailBytes, &feedingDetail); err != nil {
					s.logger.Warn("Detail反序列化失败,使用空Detail",
						zap.String("recordID", record.RecordID),
						zap.Error(err))
					feedingDetail = dto.FeedingDetail{Type: record.FeedingType}
				}
			}
		} else {
			feedingDetail = dto.FeedingDetail{Type: record.FeedingType}
		}

		// 从 detail.Note 中提取 note 字段(向后兼容)
		note := ""
		if feedingDetail.Note != nil {
			note = *feedingDetail.Note
		}

		result = append(result, dto.FeedingRecordDTO{
			RecordID:           record.RecordID,
			BabyID:             record.BabyID,
			FeedingType:        record.FeedingType,
			Amount:             record.Amount,
			Duration:           record.Duration,
			Detail:             feedingDetail, // 使用强类型结构体
			Note:               note,
			FeedingTime:        record.Time,
			ActualCompleteTime: record.ActualCompleteTime,
			CreateBy:           record.CreateBy,
			CreateTime:         record.CreateTime,
		})
	}

	return result, total, nil
}
