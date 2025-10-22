package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// feedingRecordRepositoryImpl 喂养记录仓储实现
type feedingRecordRepositoryImpl struct {
	db *gorm.DB
}

// NewFeedingRecordRepository 创建喂养记录仓储
func NewFeedingRecordRepository(db *gorm.DB) repository.FeedingRecordRepository {
	return &feedingRecordRepositoryImpl{db: db}
}

func (r *feedingRecordRepositoryImpl) Create(ctx context.Context, record *entity.FeedingRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to create feeding record", err)
	}
	return nil
}

func (r *feedingRecordRepositoryImpl) FindByID(ctx context.Context, recordID string) (*entity.FeedingRecord, error) {
	var record entity.FeedingRecord
	err := r.db.WithContext(ctx).
		Where("record_id = ? AND deleted_at IS NULL", recordID).
		First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.ErrRecordNotFound
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find feeding record", err)
	}

	return &record, nil
}

func (r *feedingRecordRepositoryImpl) FindByBabyID(
	ctx context.Context,
	babyID string,
	startTime, endTime int64,
	page, pageSize int,
) ([]*entity.FeedingRecord, int64, error) {
	var records []*entity.FeedingRecord
	var total int64

	query := r.db.WithContext(ctx).
		Model(&entity.FeedingRecord{}).
		Where("baby_id = ? AND deleted_at IS NULL", babyID)

	if startTime > 0 {
		query = query.Where("time >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("time <= ?", endTime)
	}

	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.DatabaseError, "failed to count feeding records", err)
	}

	offset := (page - 1) * pageSize
	err := query.
		Order("time DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&records).Error

	if err != nil {
		return nil, 0, errors.Wrap(errors.DatabaseError, "failed to find feeding records", err)
	}

	return records, total, nil
}

func (r *feedingRecordRepositoryImpl) Update(ctx context.Context, record *entity.FeedingRecord) error {
	err := r.db.WithContext(ctx).
		Model(&entity.FeedingRecord{}).
		Where("record_id = ? AND deleted_at IS NULL", record.RecordID).
		Updates(record).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to update feeding record", err)
	}

	return nil
}

func (r *feedingRecordRepositoryImpl) Delete(ctx context.Context, recordID string) error {
	err := r.db.WithContext(ctx).
		Where("record_id = ?", recordID).
		Delete(&entity.FeedingRecord{}).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to delete feeding record", err)
	}

	return nil
}

func (r *feedingRecordRepositoryImpl) FindUpdatedAfter(
	ctx context.Context,
	familyID string,
	timestamp int64,
) ([]*entity.FeedingRecord, error) {
	var records []*entity.FeedingRecord

	err := r.db.WithContext(ctx).
		Joins("JOIN babies ON babies.baby_id = feeding_records.baby_id").
		Where("babies.family_id = ? AND feeding_records.update_time > ?", familyID, timestamp).
		Find(&records).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find updated feeding records", err)
	}

	return records, nil
}
