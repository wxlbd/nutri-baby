package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// diaperRecordRepositoryImpl 尿布记录仓储实现
type diaperRecordRepositoryImpl struct {
	db *gorm.DB
}

// NewDiaperRecordRepository 创建尿布记录仓储
func NewDiaperRecordRepository(db *gorm.DB) repository.DiaperRecordRepository {
	return &diaperRecordRepositoryImpl{db: db}
}

func (r *diaperRecordRepositoryImpl) Create(ctx context.Context, record *entity.DiaperRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to create diaper record", err)
	}
	return nil
}

func (r *diaperRecordRepositoryImpl) FindByID(ctx context.Context, recordID string) (*entity.DiaperRecord, error) {
	var record entity.DiaperRecord
	err := r.db.WithContext(ctx).
		Where("record_id = ? AND deleted_at IS NULL", recordID).
		First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.ErrRecordNotFound
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find diaper record", err)
	}

	return &record, nil
}

func (r *diaperRecordRepositoryImpl) FindByBabyID(
	ctx context.Context,
	babyID string,
	startTime, endTime int64,
	page, pageSize int,
) ([]*entity.DiaperRecord, int64, error) {
	var records []*entity.DiaperRecord
	var total int64

	query := r.db.WithContext(ctx).
		Model(&entity.DiaperRecord{}).
		Where("baby_id = ? AND deleted_at IS NULL", babyID)

	if startTime > 0 {
		query = query.Where("change_time >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("change_time <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.DatabaseError, "failed to count diaper records", err)
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("change_time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&records).Error

	if err != nil {
		return nil, 0, errors.Wrap(errors.DatabaseError, "failed to find diaper records", err)
	}

	return records, total, nil
}

func (r *diaperRecordRepositoryImpl) Update(ctx context.Context, record *entity.DiaperRecord) error {
	err := r.db.WithContext(ctx).
		Model(&entity.DiaperRecord{}).
		Where("record_id = ? AND deleted_at IS NULL", record.RecordID).
		Updates(record).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to update diaper record", err)
	}

	return nil
}

func (r *diaperRecordRepositoryImpl) Delete(ctx context.Context, recordID string) error {
	err := r.db.WithContext(ctx).
		Where("record_id = ?", recordID).
		Delete(&entity.DiaperRecord{}).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to delete diaper record", err)
	}

	return nil
}

func (r *diaperRecordRepositoryImpl) FindUpdatedAfter(
	ctx context.Context,
	familyID string,
	timestamp int64,
) ([]*entity.DiaperRecord, error) {
	var records []*entity.DiaperRecord

	err := r.db.WithContext(ctx).
		Joins("JOIN babies ON babies.baby_id = diaper_records.baby_id").
		Where("babies.family_id = ? AND diaper_records.update_time > ?", familyID, timestamp).
		Find(&records).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find updated diaper records", err)
	}

	return records, nil
}
