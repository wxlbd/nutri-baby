package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// RecordService 记录服务 (去家庭化架构)
type RecordService struct {
	babyRepo          repository.BabyRepository
	feedingRecordRepo repository.FeedingRecordRepository
	sleepRecordRepo   repository.SleepRecordRepository
	diaperRecordRepo  repository.DiaperRecordRepository
	growthRecordRepo  repository.GrowthRecordRepository
	collaboratorRepo  repository.BabyCollaboratorRepository // 替换 familyMemberRepo
	schedulerService  *SchedulerService                     // 添加定时任务服务
	logger            *zap.Logger                           // 添加日志
}

// NewRecordService 创建记录服务
func NewRecordService(
	babyRepo repository.BabyRepository,
	feedingRecordRepo repository.FeedingRecordRepository,
	sleepRecordRepo repository.SleepRecordRepository,
	diaperRecordRepo repository.DiaperRecordRepository,
	growthRecordRepo repository.GrowthRecordRepository,
	collaboratorRepo repository.BabyCollaboratorRepository,
	schedulerService *SchedulerService,
	logger *zap.Logger,
) *RecordService {
	return &RecordService{
		babyRepo:          babyRepo,
		feedingRecordRepo: feedingRecordRepo,
		sleepRecordRepo:   sleepRecordRepo,
		diaperRecordRepo:  diaperRecordRepo,
		growthRecordRepo:  growthRecordRepo,
		collaboratorRepo:  collaboratorRepo,
		schedulerService:  schedulerService,
		logger:            logger,
	}
}

// CreateFeedingRecord 创建喂养记录
func (s *RecordService) CreateFeedingRecord(ctx context.Context, openID string, req *dto.CreateFeedingRecordRequest) (*dto.FeedingRecordDTO, error) {
	// 验证宝宝存在且用户有权限
	if err := s.checkBabyAccess(ctx, req.BabyID, openID); err != nil {
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
		Amount:      derefInt64(req.Amount),
		Duration:    derefInt(req.Duration),
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
		return nil, err
	}

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
		Note:        derefString(req.Note),
		FeedingTime: record.Time,
		CreateBy:    record.CreateBy,
		CreateTime:  record.CreateTime,
	}, nil
}

// GetFeedingRecords 获取喂养记录列表
func (s *RecordService) GetFeedingRecords(ctx context.Context, openID string, query *dto.RecordListQuery) ([]dto.FeedingRecordDTO, int64, error) {
	// 验证宝宝访问权限
	if err := s.checkBabyAccess(ctx, query.BabyID, openID); err != nil {
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

// CreateSleepRecord 创建睡眠记录
func (s *RecordService) CreateSleepRecord(ctx context.Context, openID string, req *dto.CreateSleepRecordRequest) (*dto.SleepRecordDTO, error) {
	if err := s.checkBabyAccess(ctx, req.BabyID, openID); err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()

	var endTime *int64
	if req.EndTime > 0 {
		endTime = &req.EndTime
	}

	var duration *int
	if req.Duration > 0 {
		duration = &req.Duration
	}

	sleepType := "nap"
	if req.Quality == "night" {
		sleepType = "night"
	}

	record := &entity.SleepRecord{
		RecordID:   uuid.New().String(),
		BabyID:     req.BabyID,
		StartTime:  req.StartTime,
		EndTime:    endTime,
		Duration:   duration,
		Type:       sleepType,
		CreateBy:   openID,
		CreateTime: now,
		UpdateTime: now,
	}

	if err := s.sleepRecordRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	resultEndTime := int64(0)
	if record.EndTime != nil {
		resultEndTime = *record.EndTime
	}

	resultDuration := 0
	if record.Duration != nil {
		resultDuration = *record.Duration
	}

	return &dto.SleepRecordDTO{
		RecordID:   record.RecordID,
		BabyID:     record.BabyID,
		StartTime:  record.StartTime,
		EndTime:    resultEndTime,
		Duration:   resultDuration,
		Quality:    record.Type,
		Note:       "",
		CreateBy:   record.CreateBy,
		CreateTime: record.CreateTime,
	}, nil
}

// GetSleepRecords 获取睡眠记录列表
func (s *RecordService) GetSleepRecords(ctx context.Context, openID string, query *dto.RecordListQuery) ([]dto.SleepRecordDTO, int64, error) {
	if err := s.checkBabyAccess(ctx, query.BabyID, openID); err != nil {
		return nil, 0, err
	}

	records, total, err := s.sleepRecordRepo.FindByBabyID(
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

	result := make([]dto.SleepRecordDTO, 0, len(records))
	for _, record := range records {
		endTime := int64(0)
		if record.EndTime != nil {
			endTime = *record.EndTime
		}

		duration := 0
		if record.Duration != nil {
			duration = *record.Duration
		}

		result = append(result, dto.SleepRecordDTO{
			RecordID:   record.RecordID,
			BabyID:     record.BabyID,
			StartTime:  record.StartTime,
			EndTime:    endTime,
			Duration:   duration,
			Quality:    record.Type,
			Note:       "",
			CreateBy:   record.CreateBy,
			CreateTime: record.CreateTime,
		})
	}

	return result, total, nil
}

// CreateDiaperRecord 创建尿布记录
func (s *RecordService) CreateDiaperRecord(ctx context.Context, openID string, req *dto.CreateDiaperRecordRequest) (*dto.DiaperRecordDTO, error) {
	if err := s.checkBabyAccess(ctx, req.BabyID, openID); err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	changeTime := req.ChangeTime
	if changeTime == 0 {
		changeTime = now
	}

	var note *string
	if req.Note != "" {
		note = &req.Note
	}

	record := &entity.DiaperRecord{
		RecordID:   uuid.New().String(),
		BabyID:     req.BabyID,
		Time:       changeTime,
		Type:       req.DiaperType,
		Note:       note,
		CreateBy:   openID,
		CreateTime: now,
		UpdateTime: now,
	}

	if err := s.diaperRecordRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	resultNote := ""
	if record.Note != nil {
		resultNote = *record.Note
	}

	return &dto.DiaperRecordDTO{
		RecordID:   record.RecordID,
		BabyID:     record.BabyID,
		DiaperType: record.Type,
		Note:       resultNote,
		ChangeTime: record.Time,
		CreateBy:   record.CreateBy,
		CreateTime: record.CreateTime,
	}, nil
}

// GetDiaperRecords 获取尿布记录列表
func (s *RecordService) GetDiaperRecords(ctx context.Context, openID string, query *dto.RecordListQuery) ([]dto.DiaperRecordDTO, int64, error) {
	if err := s.checkBabyAccess(ctx, query.BabyID, openID); err != nil {
		return nil, 0, err
	}

	records, total, err := s.diaperRecordRepo.FindByBabyID(
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

	result := make([]dto.DiaperRecordDTO, 0, len(records))
	for _, record := range records {
		note := ""
		if record.Note != nil {
			note = *record.Note
		}

		result = append(result, dto.DiaperRecordDTO{
			RecordID:   record.RecordID,
			BabyID:     record.BabyID,
			DiaperType: record.Type,
			Note:       note,
			ChangeTime: record.Time,
			CreateBy:   record.CreateBy,
			CreateTime: record.CreateTime,
		})
	}

	return result, total, nil
}

// CreateGrowthRecord 创建生长记录
func (s *RecordService) CreateGrowthRecord(ctx context.Context, openID string, req *dto.CreateGrowthRecordRequest) (*dto.GrowthRecordDTO, error) {
	if err := s.checkBabyAccess(ctx, req.BabyID, openID); err != nil {
		return nil, err
	}

	now := time.Now().UnixMilli()
	measureTime := req.MeasureTime
	if measureTime == 0 {
		measureTime = now
	}

	var height *float64
	if req.Height > 0 {
		h := float64(req.Height)
		height = &h
	}

	var weight *float64
	if req.Weight > 0 {
		w := float64(req.Weight)
		weight = &w
	}

	var headCircumference *float64
	if req.HeadCircumference > 0 {
		hc := req.HeadCircumference
		headCircumference = &hc
	}

	var note *string
	if req.Note != "" {
		note = &req.Note
	}

	record := &entity.GrowthRecord{
		RecordID:          uuid.New().String(),
		BabyID:            req.BabyID,
		Time:              measureTime,
		Height:            height,
		Weight:            weight,
		HeadCircumference: headCircumference,
		Note:              note,
		CreateBy:          openID,
		CreateTime:        now,
		UpdateTime:        now,
	}

	if err := s.growthRecordRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	// 同时更新宝宝的身高体重
	baby, err := s.babyRepo.FindByID(ctx, req.BabyID)
	if err == nil && height != nil && weight != nil {
		baby.Height = *height
		baby.Weight = *weight
		baby.UpdateTime = now
		_ = s.babyRepo.Update(ctx, baby)
	}

	resultNote := ""
	if record.Note != nil {
		resultNote = *record.Note
	}

	return &dto.GrowthRecordDTO{
		RecordID:          record.RecordID,
		BabyID:            record.BabyID,
		Height:            record.Height,
		Weight:            record.Weight,
		HeadCircumference: record.HeadCircumference,
		Note:              resultNote,
		MeasureTime:       record.Time,
		CreateBy:          record.CreateBy,
		CreateTime:        record.CreateTime,
	}, nil
}

// GetGrowthRecords 获取生长记录列表
func (s *RecordService) GetGrowthRecords(ctx context.Context, openID string, query *dto.RecordListQuery) ([]dto.GrowthRecordDTO, int64, error) {
	if err := s.checkBabyAccess(ctx, query.BabyID, openID); err != nil {
		return nil, 0, err
	}

	records, total, err := s.growthRecordRepo.FindByBabyID(
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

	result := make([]dto.GrowthRecordDTO, 0, len(records))
	for _, record := range records {
		note := ""
		if record.Note != nil {
			note = *record.Note
		}

		result = append(result, dto.GrowthRecordDTO{
			RecordID:          record.RecordID,
			BabyID:            record.BabyID,
			Height:            record.Height,
			Weight:            record.Weight,
			HeadCircumference: record.HeadCircumference,
			Note:              note,
			MeasureTime:       record.Time,
			CreateBy:          record.CreateBy,
			CreateTime:        record.CreateTime,
		})
	}

	return result, total, nil
}

// checkBabyAccess 检查用户对宝宝的访问权限 (去家庭化架构)
func (s *RecordService) checkBabyAccess(ctx context.Context, babyID, openID string) error {
	// 检查用户是否为宝宝的协作者
	isCollaborator, err := s.collaboratorRepo.IsCollaborator(ctx, babyID, openID)
	if err != nil {
		return err
	}

	if !isCollaborator {
		return errors.New(errors.PermissionDenied, "您没有权限访问该宝宝的记录")
	}

	return nil
}

// derefInt64 解引用int64指针，如果为nil返回0
func derefInt64(val *int64) int64 {
	if val == nil {
		return 0
	}
	return *val
}

// derefInt 解引用int指针，如果为nil返回0
func derefInt(val *int) int {
	if val == nil {
		return 0
	}
	return *val
}

// derefString 解引用string指针，如果为nil返回空字符串
func derefString(val *string) string {
	if val == nil {
		return ""
	}
	return *val
}
