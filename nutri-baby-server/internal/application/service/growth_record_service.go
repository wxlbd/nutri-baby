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
