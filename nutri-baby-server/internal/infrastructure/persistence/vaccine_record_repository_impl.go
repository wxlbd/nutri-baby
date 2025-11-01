package persistence

import (
	"context"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
	"gorm.io/gorm"
)

// vaccineRecordRepositoryImpl 疫苗接种记录仓储实现
type vaccineRecordRepositoryImpl struct {
	db *gorm.DB
}

// NewVaccineRecordRepository 创建疫苗接种记录仓储
func NewVaccineRecordRepository(db *gorm.DB) repository.VaccineRecordRepository {
	return &vaccineRecordRepositoryImpl{db: db}
}

// Create 创建记录
func (r *vaccineRecordRepositoryImpl) Create(ctx context.Context, record *entity.VaccineRecord) error {
	if err := r.db.WithContext(ctx).Create(record).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to create vaccine record", err)
	}
	return nil
}

// FindByID 根据ID查找记录
func (r *vaccineRecordRepositoryImpl) FindByID(ctx context.Context, recordID string) (*entity.VaccineRecord, error) {
	var record entity.VaccineRecord
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("record_id = ? AND deleted_at IS NULL", recordID).
		First(&record).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.ErrRecordNotFound
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine record", err)
	}

	return &record, nil
}

// FindByBabyID 查找宝宝的疫苗记录(分页)
func (r *vaccineRecordRepositoryImpl) FindByBabyID(
	ctx context.Context,
	babyID string,
	startTime, endTime int64,
	vaccineType string,
	page, pageSize int,
) ([]*entity.VaccineRecord, int64, error) {
	var records []*entity.VaccineRecord
	var total int64

	query := r.db.WithContext(ctx).
		Model(&entity.VaccineRecord{}).
		Where("baby_id = ? AND deleted_at IS NULL", babyID)

	// 时间范围过滤
	if startTime > 0 {
		query = query.Where("vaccine_date >= ?", startTime)
	}
	if endTime > 0 {
		query = query.Where("vaccine_date <= ?", endTime)
	}

	// 疫苗类型过滤
	if vaccineType != "" {
		query = query.Where("vaccine_type = ?", vaccineType)
	}

	// 统计总数
	if err := query.Count(&total).Error; err != nil {
		return nil, 0, errors.Wrap(errors.DatabaseError, "failed to count vaccine records", err)
	}

	// 分页查询
	offset := (page - 1) * pageSize
	err := query.
		Preload("Plan").
		Order("vaccine_date DESC").
		Limit(pageSize).
		Offset(offset).
		Find(&records).Error

	if err != nil {
		return nil, 0, errors.Wrap(errors.DatabaseError, "failed to find vaccine records", err)
	}

	return records, total, nil
}

// FindByBabyAndPlan 查找宝宝特定计划的接种记录
func (r *vaccineRecordRepositoryImpl) FindByBabyAndPlan(
	ctx context.Context,
	babyID, planID string,
) (*entity.VaccineRecord, error) {
	var record entity.VaccineRecord
	err := r.db.WithContext(ctx).
		Where("baby_id = ? AND plan_id = ? AND deleted_at IS NULL", babyID, planID).
		First(&record).Error

	if err == gorm.ErrRecordNotFound {
		return nil, nil // 没有找到不算错误
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine record", err)
	}

	return &record, nil
}

// Update 更新记录
func (r *vaccineRecordRepositoryImpl) Update(ctx context.Context, record *entity.VaccineRecord) error {
	err := r.db.WithContext(ctx).
		Model(&entity.VaccineRecord{}).
		Where("record_id = ? AND deleted_at IS NULL", record.RecordID).
		Updates(record).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to update vaccine record", err)
	}

	return nil
}

// Delete 删除记录(软删除)
func (r *vaccineRecordRepositoryImpl) Delete(ctx context.Context, recordID string) error {
	err := r.db.WithContext(ctx).
		Where("record_id = ?", recordID).
		Delete(&entity.VaccineRecord{}).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to delete vaccine record", err)
	}

	return nil
}

// FindUpdatedAfter 查找指定时间后更新的记录(用于同步)
func (r *vaccineRecordRepositoryImpl) FindUpdatedAfter(
	ctx context.Context,
	familyID string,
	timestamp int64,
) ([]*entity.VaccineRecord, error) {
	var records []*entity.VaccineRecord

	// 通过baby关联查询family的记录
	err := r.db.WithContext(ctx).
		Joins("JOIN babies ON babies.baby_id = vaccine_records.baby_id").
		Where("babies.family_id = ? AND vaccine_records.update_time > ?", familyID, timestamp).
		Preload("Plan").
		Find(&records).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find updated vaccine records", err)
	}

	return records, nil
}

// CountCompleted 统计已完成接种数量
func (r *vaccineRecordRepositoryImpl) CountCompleted(ctx context.Context, babyID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.VaccineRecord{}).
		Where("baby_id = ? AND deleted_at IS NULL", babyID).
		Count(&count).Error

	if err != nil {
		return 0, errors.Wrap(errors.DatabaseError, "failed to count completed vaccines", err)
	}

	return count, nil
}
