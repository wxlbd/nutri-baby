package service

import (
	"context"
	"time"

	"github.com/google/uuid"

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
}

// NewRecordService 创建记录服务
func NewRecordService(
	babyRepo repository.BabyRepository,
	feedingRecordRepo repository.FeedingRecordRepository,
	sleepRecordRepo repository.SleepRecordRepository,
	diaperRecordRepo repository.DiaperRecordRepository,
	growthRecordRepo repository.GrowthRecordRepository,
	collaboratorRepo repository.BabyCollaboratorRepository,
) *RecordService {
	return &RecordService{
		babyRepo:          babyRepo,
		feedingRecordRepo: feedingRecordRepo,
		sleepRecordRepo:   sleepRecordRepo,
		diaperRecordRepo:  diaperRecordRepo,
		growthRecordRepo:  growthRecordRepo,
		collaboratorRepo:  collaboratorRepo,
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

	// 合并额外的detail信息
	if req.Detail.BreastSide != "" {
		detail["breastSide"] = req.Detail.BreastSide
	}
	if req.Detail.LeftTime > 0 {
		detail["leftTime"] = req.Detail.LeftTime
	}
	if req.Detail.RightTime > 0 {
		detail["rightTime"] = req.Detail.RightTime
	}
	if req.Detail.FormulaType != "" {
		detail["formulaType"] = req.Detail.FormulaType
	}

	record := &entity.FeedingRecord{
		RecordID:   uuid.New().String(),
		BabyID:     req.BabyID,
		Time:       feedingTime,
		Detail:     detail,
		CreateBy:   openID,
		CreateTime: now,
		UpdateTime: now,
	}

	if err := s.feedingRecordRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	return &dto.FeedingRecordDTO{
		RecordID:    record.RecordID,
		BabyID:      record.BabyID,
		FeedingType: req.FeedingType,
		Amount:      req.Amount,
		Duration:    req.Duration,
		Detail:      req.Detail,
		Note:        req.Note,
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
		// 从Detail map中提取各个字段
		feedingType, _ := record.Detail["feedingType"].(string)
		amount := 0
		if v, ok := record.Detail["amount"].(float64); ok {
			amount = int(v)
		}
		duration := 0
		if v, ok := record.Detail["duration"].(float64); ok {
			duration = int(v)
		}
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
			FeedingType: feedingType,
			Amount:      amount,
			Duration:    duration,
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
	recordTime := req.RecordTime
	if recordTime == 0 {
		recordTime = now
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

	var headCirc *float64
	if req.HeadCircum > 0 {
		hc := float64(req.HeadCircum)
		headCirc = &hc
	}

	var note *string
	if req.Note != "" {
		note = &req.Note
	}

	record := &entity.GrowthRecord{
		RecordID:          uuid.New().String(),
		BabyID:            req.BabyID,
		Time:              recordTime,
		Height:            height,
		Weight:            weight,
		HeadCircumference: headCirc,
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

	resultHeight := 0
	if record.Height != nil {
		resultHeight = int(*record.Height)
	}

	resultWeight := 0
	if record.Weight != nil {
		resultWeight = int(*record.Weight)
	}

	resultHeadCirc := 0
	if record.HeadCircumference != nil {
		resultHeadCirc = int(*record.HeadCircumference)
	}

	resultNote := ""
	if record.Note != nil {
		resultNote = *record.Note
	}

	return &dto.GrowthRecordDTO{
		RecordID:   record.RecordID,
		BabyID:     record.BabyID,
		Height:     resultHeight,
		Weight:     resultWeight,
		HeadCircum: resultHeadCirc,
		Note:       resultNote,
		RecordTime: record.Time,
		CreateBy:   record.CreateBy,
		CreateTime: record.CreateTime,
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
		height := 0
		if record.Height != nil {
			height = int(*record.Height)
		}

		weight := 0
		if record.Weight != nil {
			weight = int(*record.Weight)
		}

		headCirc := 0
		if record.HeadCircumference != nil {
			headCirc = int(*record.HeadCircumference)
		}

		note := ""
		if record.Note != nil {
			note = *record.Note
		}

		result = append(result, dto.GrowthRecordDTO{
			RecordID:   record.RecordID,
			BabyID:     record.BabyID,
			Height:     height,
			Weight:     weight,
			HeadCircum: headCirc,
			Note:       note,
			RecordTime: record.Time,
			CreateBy:   record.CreateBy,
			CreateTime: record.CreateTime,
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
