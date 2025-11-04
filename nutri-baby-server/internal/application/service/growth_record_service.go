package service

import (
	"context"
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
)

// GrowthRecordService 成长记录服务
type GrowthRecordService struct {
	*BaseRecordService
	growthRecordRepo repository.GrowthRecordRepository
}

// NewGrowthRecordService 创建成长记录服务
func NewGrowthRecordService(
	babyRepo repository.BabyRepository,
	collaboratorRepo repository.BabyCollaboratorRepository,
	growthRecordRepo repository.GrowthRecordRepository,
	logger *zap.Logger,
) *GrowthRecordService {
	return &GrowthRecordService{
		BaseRecordService: NewBaseRecordService(babyRepo, collaboratorRepo, logger),
		growthRecordRepo:  growthRecordRepo,
	}
}

// CreateGrowthRecord 创建生长记录
func (s *GrowthRecordService) CreateGrowthRecord(ctx context.Context, openID string, req *dto.CreateGrowthRecordRequest) (*dto.GrowthRecordDTO, error) {
	if err := s.CheckBabyAccess(ctx, req.BabyID, openID); err != nil {
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
func (s *GrowthRecordService) GetGrowthRecords(ctx context.Context, openID string, query *dto.RecordListQuery) ([]dto.GrowthRecordDTO, int64, error) {
	if err := s.CheckBabyAccess(ctx, query.BabyID, openID); err != nil {
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

// GetGrowthRecordById 根据ID获取单条生长记录
func (s *GrowthRecordService) GetGrowthRecordById(ctx context.Context, openID, recordID string) (*dto.GrowthRecordDTO, error) {
	// 先获取记录
	record, err := s.growthRecordRepo.FindByID(ctx, recordID)
	if err != nil {
		s.logger.Error("获取生长记录失败",
			zap.String("recordID", recordID),
			zap.Error(err))
		return nil, err
	}

	// 验证用户是否有权限访问该宝宝的记录
	if err := s.CheckBabyAccess(ctx, record.BabyID, openID); err != nil {
		return nil, err
	}

	note := ""
	if record.Note != nil {
		note = *record.Note
	}

	return &dto.GrowthRecordDTO{
		RecordID:          record.RecordID,
		BabyID:            record.BabyID,
		Height:            record.Height,
		Weight:            record.Weight,
		HeadCircumference: record.HeadCircumference,
		Note:              note,
		MeasureTime:       record.Time,
		CreateBy:          record.CreateBy,
		CreateTime:        record.CreateTime,
	}, nil
}

// UpdateGrowthRecord 更新生长记录
func (s *GrowthRecordService) UpdateGrowthRecord(ctx context.Context, openID, recordID string, req *dto.UpdateGrowthRecordRequest) (*dto.GrowthRecordDTO, error) {
	// 先获取记录
	record, err := s.growthRecordRepo.FindByID(ctx, recordID)
	if err != nil {
		s.logger.Error("获取生长记录失败",
			zap.String("recordID", recordID),
			zap.Error(err))
		return nil, err
	}

	// 验证权限
	if err := s.CheckBabyAccess(ctx, record.BabyID, openID); err != nil {
		return nil, err
	}

	// 更新字段 (只更新非nil字段)
	now := time.Now().UnixMilli()
	updated := false

	if req.Height != nil {
		record.Height = req.Height
		updated = true
	}

	if req.Weight != nil {
		record.Weight = req.Weight
		updated = true
	}

	if req.HeadCircumference != nil {
		record.HeadCircumference = req.HeadCircumference
		updated = true
	}

	if req.MeasureTime != nil && *req.MeasureTime != record.Time {
		record.Time = *req.MeasureTime
		updated = true
	}

	if req.Note != nil {
		record.Note = req.Note
		updated = true
	}

	// 如果没有更新任何字段,直接返回
	if !updated {
		s.logger.Info("没有更新任何字段",
			zap.String("recordID", recordID))
		return s.GetGrowthRecordById(ctx, openID, recordID)
	}

	// 更新时间戳
	record.UpdateTime = now

	// 保存更新
	if err := s.growthRecordRepo.Update(ctx, record); err != nil {
		s.logger.Error("更新生长记录失败",
			zap.String("recordID", recordID),
			zap.Error(err))
		return nil, err
	}

	// 如果更新了身高体重,同时更新宝宝档案中的身高体重
	if req.Height != nil || req.Weight != nil {
		baby, err := s.babyRepo.FindByID(ctx, record.BabyID)
		if err == nil {
			if req.Height != nil {
				baby.Height = *req.Height
			}
			if req.Weight != nil {
				baby.Weight = *req.Weight
			}
			baby.UpdateTime = now
			_ = s.babyRepo.Update(ctx, baby)
		}
	}

	s.logger.Info("生长记录更新成功",
		zap.String("recordID", record.RecordID),
		zap.String("babyID", record.BabyID))

	// 返回更新后的记录
	return s.GetGrowthRecordById(ctx, openID, recordID)
}

// DeleteGrowthRecord 删除生长记录
func (s *GrowthRecordService) DeleteGrowthRecord(ctx context.Context, openID, recordID string) error {
	// 先获取记录
	record, err := s.growthRecordRepo.FindByID(ctx, recordID)
	if err != nil {
		s.logger.Error("获取生长记录失败",
			zap.String("recordID", recordID),
			zap.Error(err))
		return err
	}

	// 验证权限
	if err := s.CheckBabyAccess(ctx, record.BabyID, openID); err != nil {
		return err
	}

	// 删除记录 (软删除)
	if err := s.growthRecordRepo.Delete(ctx, recordID); err != nil {
		s.logger.Error("删除生长记录失败",
			zap.String("recordID", recordID),
			zap.Error(err))
		return err
	}

	s.logger.Info("生长记录删除成功",
		zap.String("recordID", recordID),
		zap.String("babyID", record.BabyID))

	return nil
}
