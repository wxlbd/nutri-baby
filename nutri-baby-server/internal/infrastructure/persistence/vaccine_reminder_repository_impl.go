package persistence

import (
	"context"
	"time"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
	"gorm.io/gorm"
)

// vaccineReminderRepositoryImpl 疫苗提醒仓储实现
type vaccineReminderRepositoryImpl struct {
	db *gorm.DB
}

// NewVaccineReminderRepository 创建疫苗提醒仓储
func NewVaccineReminderRepository(db *gorm.DB) repository.VaccineReminderRepository {
	return &vaccineReminderRepositoryImpl{db: db}
}

// Create 创建提醒
func (r *vaccineReminderRepositoryImpl) Create(ctx context.Context, reminder *entity.VaccineReminder) error {
	if err := r.db.WithContext(ctx).Create(reminder).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to create vaccine reminder", err)
	}
	return nil
}

// FindByID 根据ID查找提醒
func (r *vaccineReminderRepositoryImpl) FindByID(ctx context.Context, reminderID string) (*entity.VaccineReminder, error) {
	var reminder entity.VaccineReminder
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("reminder_id = ? AND deleted_at IS NULL", reminderID).
		First(&reminder).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.ErrRecordNotFound
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine reminder", err)
	}

	return &reminder, nil
}

// FindByBabyID 查找宝宝的所有提醒
func (r *vaccineReminderRepositoryImpl) FindByBabyID(ctx context.Context, babyID string) ([]*entity.VaccineReminder, error) {
	var reminders []*entity.VaccineReminder
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("baby_id = ? AND deleted_at IS NULL", babyID).
		Order("scheduled_date ASC").
		Find(&reminders).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine reminders", err)
	}

	return reminders, nil
}

// FindByStatus 根据状态查找提醒
func (r *vaccineReminderRepositoryImpl) FindByStatus(
	ctx context.Context,
	babyID, status string,
	limit int,
) ([]*entity.VaccineReminder, error) {
	var reminders []*entity.VaccineReminder
	query := r.db.WithContext(ctx).
		Preload("Plan").
		Where("baby_id = ? AND status = ? AND deleted_at IS NULL", babyID, status).
		Order("scheduled_date ASC")

	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&reminders).Error
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine reminders by status", err)
	}

	return reminders, nil
}

// FindDueReminders 查找即将到期和已逾期的提醒(用于推送)
func (r *vaccineReminderRepositoryImpl) FindDueReminders(ctx context.Context) ([]*entity.VaccineReminder, error) {
	var reminders []*entity.VaccineReminder

	// 查找状态为 upcoming、due 或 overdue 且未发送的提醒
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Where("status IN (?, ?, ?) AND reminder_sent = ? AND deleted_at IS NULL",
			entity.ReminderStatusUpcoming,
			entity.ReminderStatusDue,
			entity.ReminderStatusOverdue,
			false).
		Order("scheduled_date ASC").
		Find(&reminders).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find due vaccine reminders", err)
	}

	return reminders, nil
}

// Update 更新提醒
func (r *vaccineReminderRepositoryImpl) Update(ctx context.Context, reminder *entity.VaccineReminder) error {
	err := r.db.WithContext(ctx).
		Model(&entity.VaccineReminder{}).
		Where("reminder_id = ? AND deleted_at IS NULL", reminder.ReminderID).
		Updates(reminder).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to update vaccine reminder", err)
	}

	return nil
}

// UpdateStatus 更新提醒状态
func (r *vaccineReminderRepositoryImpl) UpdateStatus(ctx context.Context, reminderID, status string) error {
	err := r.db.WithContext(ctx).
		Model(&entity.VaccineReminder{}).
		Where("reminder_id = ? AND deleted_at IS NULL", reminderID).
		Update("status", status).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to update vaccine reminder status", err)
	}

	return nil
}

// MarkSent 标记提醒已发送
func (r *vaccineReminderRepositoryImpl) MarkSent(ctx context.Context, reminderID string) error {
	now := time.Now().UnixMilli()
	err := r.db.WithContext(ctx).
		Model(&entity.VaccineReminder{}).
		Where("reminder_id = ? AND deleted_at IS NULL", reminderID).
		Updates(map[string]interface{}{
			"reminder_sent": true,
			"sent_time":     now,
		}).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to mark vaccine reminder as sent", err)
	}

	return nil
}

// CountByStatus 统计各状态的提醒数量
func (r *vaccineReminderRepositoryImpl) CountByStatus(ctx context.Context, babyID string) (map[string]int64, error) {
	type StatusCount struct {
		Status string
		Count  int64
	}

	var results []StatusCount
	err := r.db.WithContext(ctx).
		Model(&entity.VaccineReminder{}).
		Select("status, COUNT(*) as count").
		Where("baby_id = ? AND deleted_at IS NULL", babyID).
		Group("status").
		Scan(&results).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to count vaccine reminders by status", err)
	}

	// 转换为 map
	counts := make(map[string]int64)
	for _, result := range results {
		counts[result.Status] = result.Count
	}

	return counts, nil
}

// BatchCreate 批量创建提醒(初始化宝宝疫苗计划时使用)
func (r *vaccineReminderRepositoryImpl) BatchCreate(ctx context.Context, reminders []*entity.VaccineReminder) error {
	if len(reminders) == 0 {
		return nil
	}

	if err := r.db.WithContext(ctx).CreateInBatches(reminders, 100).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to batch create vaccine reminders", err)
	}

	return nil
}

// FindByBabyAndPlan 查找宝宝特定计划的提醒
func (r *vaccineReminderRepositoryImpl) FindByBabyAndPlan(
	ctx context.Context,
	babyID, planID string,
) (*entity.VaccineReminder, error) {
	var reminder entity.VaccineReminder
	err := r.db.WithContext(ctx).
		Where("baby_id = ? AND plan_id = ? AND deleted_at IS NULL", babyID, planID).
		First(&reminder).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil // 没有找到不算错误
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine reminder", err)
	}

	return &reminder, nil
}
