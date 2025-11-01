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

// DiaperRecordService 尿布记录服务
type DiaperRecordService struct {
	*BaseRecordService
	diaperRecordRepo repository.DiaperRecordRepository
}

// NewDiaperRecordService 创建尿布记录服务
func NewDiaperRecordService(
	babyRepo repository.BabyRepository,
	collaboratorRepo repository.BabyCollaboratorRepository,
	diaperRecordRepo repository.DiaperRecordRepository,
	logger *zap.Logger,
) *DiaperRecordService {
	return &DiaperRecordService{
		BaseRecordService: NewBaseRecordService(babyRepo, collaboratorRepo, logger),
		diaperRecordRepo:  diaperRecordRepo,
	}
}

// CreateDiaperRecord 创建尿布记录
func (s *DiaperRecordService) CreateDiaperRecord(ctx context.Context, openID string, req *dto.CreateDiaperRecordRequest) (*dto.DiaperRecordDTO, error) {
	if err := s.CheckBabyAccess(ctx, req.BabyID, openID); err != nil {
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
func (s *DiaperRecordService) GetDiaperRecords(ctx context.Context, openID string, query *dto.RecordListQuery) ([]dto.DiaperRecordDTO, int64, error) {
	if err := s.CheckBabyAccess(ctx, query.BabyID, openID); err != nil {
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
