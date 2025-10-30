package service

import (
	"context"
	"errors"
	"time"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"go.uber.org/zap"
)

// FeedingService 喂养记录服务
type FeedingService struct {
	feedingRecordRepo repository.FeedingRecordRepository
	schedulerService  *SchedulerService
	logger            *zap.Logger
}

// NewFeedingService 创建喂养记录服务
func NewFeedingService(
	feedingRecordRepo repository.FeedingRecordRepository,
	schedulerService *SchedulerService,
	logger *zap.Logger,
) *FeedingService {
	return &FeedingService{
		feedingRecordRepo: feedingRecordRepo,
		schedulerService:  schedulerService,
		logger:            logger,
	}
}

// CreateRecord 创建喂养记录
func (s *FeedingService) CreateRecord(ctx context.Context, req *dto.CreateFeedingRecordRequest) (*dto.FeedingRecordResponse, error) {
	if req == nil {
		return nil, errors.New("request cannot be nil")
	}

	// 创建记录实体
	record := &entity.FeedingRecord{
		RecordID:       generateRecordID(), // 生成唯一ID
		BabyID:         req.BabyID,
		Time:           req.FeedingTime,
		FeedingType:    req.FeedingType,
		Amount:         *req.Amount,
		Duration:       *req.Duration,
		Detail:         entity.FeedingDetail(req.Detail),
		CreateBy:       extractUserID(ctx),
		CreateByName:   extractUserName(ctx),
		CreateByAvatar: extractUserAvatar(ctx),
		CreateTime:     time.Now().UnixMilli(),
		UpdateTime:     time.Now().UnixMilli(),
	}

	// 处理用户自定义的提醒间隔
	if req.ReminderInterval != nil && *req.ReminderInterval > 0 {
		record.ReminderInterval = req.ReminderInterval

		// 计算下次提醒时间: 喂养时间 + 间隔(分钟)
		nextReminderTime := req.FeedingTime + int64(*req.ReminderInterval*60*1000)
		record.NextReminderTime = &nextReminderTime

		s.logger.Info("设置喂养提醒",
			zap.String("babyID", req.BabyID),
			zap.Int("intervalMinutes", *req.ReminderInterval),
			zap.Int64("nextReminderTime", nextReminderTime))
	}

	// 保存记录到数据库
	err := s.feedingRecordRepo.Create(ctx, record)
	if err != nil {
		s.logger.Error("保存喂养记录失败",
			zap.String("babyID", req.BabyID),
			zap.Error(err))
		return nil, err
	}

	s.logger.Info("喂养记录创建成功",
		zap.String("recordID", record.RecordID),
		zap.String("babyID", record.BabyID),
		zap.String("feedingType", record.FeedingType))

	// 如果设置了提醒时间，添加定时任务到 gocron
	if record.NextReminderTime != nil {
		jobTag, err := s.schedulerService.AddFeedingReminderTask(ctx, record)
		if err != nil {
			// 定时任务添加失败不影响记录保存，仅记录警告日志
			s.logger.Warn("添加喂养提醒定时任务失败，用户将无法收到提醒",
				zap.String("recordID", record.RecordID),
				zap.Error(err))
			// 不返回错误，允许记录保存成功
		} else if jobTag != "" {
			s.logger.Info("喂养提醒定时任务已添加",
				zap.String("recordID", record.RecordID),
				zap.String("jobTag", jobTag))
		}
	}

	// 构建响应
	return &dto.FeedingRecordResponse{
		RecordID:         record.RecordID,
		BabyID:           record.BabyID,
		FeedingType:      record.FeedingType,
		Amount:           &record.Amount,
		Duration:         &record.Duration,
		Detail:           record.Detail,
		FeedingTime:      record.Time,
		ReminderInterval: record.ReminderInterval,
		NextReminderTime: record.NextReminderTime,
		CreateBy:         record.CreateBy,
		CreateByName:     record.CreateByName,
		CreateByAvatar:   record.CreateByAvatar,
		CreateTime:       record.CreateTime,
		UpdateTime:       record.UpdateTime,
	}, nil
}

// 辅助函数

func generateRecordID() string {
	// 实现ID生成逻辑，例如使用 UUID 或雪花算法
	// return uuid.New().String()
	return ""
}

func extractUserID(ctx context.Context) string {
	// 从context提取用户ID(OpenID)
	// 示例: return ctx.Value("openid").(string)
	return ""
}

func extractUserName(ctx context.Context) string {
	// 从context提取用户名
	return ""
}

func extractUserAvatar(ctx context.Context) string {
	// 从context提取用户头像
	return ""
}
