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

// SleepRecordService 睡眠记录服务
type SleepRecordService struct {
	*BaseRecordService
	sleepRecordRepo repository.SleepRecordRepository
}

// NewSleepRecordService 创建睡眠记录服务
func NewSleepRecordService(
	babyRepo repository.BabyRepository,
	collaboratorRepo repository.BabyCollaboratorRepository,
	sleepRecordRepo repository.SleepRecordRepository,
	logger *zap.Logger,
) *SleepRecordService {
	return &SleepRecordService{
		BaseRecordService: NewBaseRecordService(babyRepo, collaboratorRepo, logger),
		sleepRecordRepo:   sleepRecordRepo,
	}
}

// CreateSleepRecord 创建睡眠记录
func (s *SleepRecordService) CreateSleepRecord(ctx context.Context, openID string, req *dto.CreateSleepRecordRequest) (*dto.SleepRecordDTO, error) {
	if err := s.CheckBabyAccess(ctx, req.BabyID, openID); err != nil {
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
func (s *SleepRecordService) GetSleepRecords(ctx context.Context, openID string, query *dto.RecordListQuery) ([]dto.SleepRecordDTO, int64, error) {
	if err := s.CheckBabyAccess(ctx, query.BabyID, openID); err != nil {
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
