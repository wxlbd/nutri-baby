package persistence

import (
	"context"

	"gorm.io/gorm"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// vaccinePlanRepositoryImpl 疫苗计划仓储实现
type vaccinePlanRepositoryImpl struct {
	db *gorm.DB
}

// NewVaccinePlanRepository 创建疫苗计划仓储
func NewVaccinePlanRepository(db *gorm.DB) repository.VaccinePlanRepository {
	return &vaccinePlanRepositoryImpl{db: db}
}

func (r *vaccinePlanRepositoryImpl) FindAll(ctx context.Context) ([]*entity.VaccinePlan, error) {
	var plans []*entity.VaccinePlan
	err := r.db.WithContext(ctx).
		Where("deleted_at IS NULL").
		Order("age_in_months ASC, dose_number ASC").
		Find(&plans).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine plans", err)
	}

	return plans, nil
}

func (r *vaccinePlanRepositoryImpl) FindByID(ctx context.Context, planID string) (*entity.VaccinePlan, error) {
	var plan entity.VaccinePlan
	err := r.db.WithContext(ctx).
		Where("plan_id = ? AND deleted_at IS NULL", planID).
		First(&plan).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine plan", err)
	}

	return &plan, nil
}

func (r *vaccinePlanRepositoryImpl) Create(ctx context.Context, plan *entity.VaccinePlan) error {
	if err := r.db.WithContext(ctx).Create(plan).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to create vaccine plan", err)
	}
	return nil
}

func (r *vaccinePlanRepositoryImpl) BatchCreate(ctx context.Context, plans []*entity.VaccinePlan) error {
	if err := r.db.WithContext(ctx).CreateInBatches(plans, 50).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to batch create vaccine plans", err)
	}
	return nil
}

// vaccineReminderRepositoryImpl 疫苗提醒仓储实现
type vaccineReminderRepositoryImpl struct {
	db *gorm.DB
}

// NewVaccineReminderRepository 创建疫苗提醒仓储
func NewVaccineReminderRepository(db *gorm.DB) repository.VaccineReminderRepository {
	return &vaccineReminderRepositoryImpl{db: db}
}

func (r *vaccineReminderRepositoryImpl) Create(ctx context.Context, reminder *entity.VaccineReminder) error {
	if err := r.db.WithContext(ctx).Create(reminder).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to create vaccine reminder", err)
	}
	return nil
}

func (r *vaccineReminderRepositoryImpl) FindByID(ctx context.Context, reminderID string) (*entity.VaccineReminder, error) {
	var reminder entity.VaccineReminder
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Preload("Baby").
		Where("reminder_id = ? AND deleted_at IS NULL", reminderID).
		First(&reminder).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, errors.ErrNotFound
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine reminder", err)
	}

	return &reminder, nil
}

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

func (r *vaccineReminderRepositoryImpl) FindByStatus(ctx context.Context, babyID, status string, limit int) ([]*entity.VaccineReminder, error) {
	var reminders []*entity.VaccineReminder
	query := r.db.WithContext(ctx).
		Preload("Plan").
		Where("deleted_at IS NULL")

	if babyID != "" {
		query = query.Where("baby_id = ?", babyID)
	}
	if status != "" {
		query = query.Where("status = ?", status)
	}

	query = query.Order("scheduled_date ASC")
	if limit > 0 {
		query = query.Limit(limit)
	}

	err := query.Find(&reminders).Error
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine reminders by status", err)
	}

	return reminders, nil
}

func (r *vaccineReminderRepositoryImpl) FindDueReminders(ctx context.Context) ([]*entity.VaccineReminder, error) {
	var reminders []*entity.VaccineReminder
	err := r.db.WithContext(ctx).
		Preload("Plan").
		Preload("Baby").
		Where("status IN (?, ?) AND reminder_sent = ? AND deleted_at IS NULL", "due", "overdue", false).
		Find(&reminders).Error

	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find due reminders", err)
	}

	return reminders, nil
}

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

func (r *vaccineReminderRepositoryImpl) UpdateStatus(ctx context.Context, reminderID, status string) error {
	err := r.db.WithContext(ctx).
		Model(&entity.VaccineReminder{}).
		Where("reminder_id = ?", reminderID).
		Update("status", status).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to update reminder status", err)
	}

	return nil
}

func (r *vaccineReminderRepositoryImpl) MarkSent(ctx context.Context, reminderID string) error {
	now := ctx.Value("current_time").(int64)
	err := r.db.WithContext(ctx).
		Model(&entity.VaccineReminder{}).
		Where("reminder_id = ?", reminderID).
		Updates(map[string]interface{}{
			"reminder_sent": true,
			"sent_time":     now,
		}).Error

	if err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to mark reminder sent", err)
	}

	return nil
}

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
		return nil, errors.Wrap(errors.DatabaseError, "failed to count by status", err)
	}

	counts := make(map[string]int64)
	for _, r := range results {
		counts[r.Status] = r.Count
	}

	return counts, nil
}

func (r *vaccineReminderRepositoryImpl) BatchCreate(ctx context.Context, reminders []*entity.VaccineReminder) error {
	if err := r.db.WithContext(ctx).CreateInBatches(reminders, 50).Error; err != nil {
		return errors.Wrap(errors.DatabaseError, "failed to batch create vaccine reminders", err)
	}
	return nil
}

func (r *vaccineReminderRepositoryImpl) FindByBabyAndPlan(ctx context.Context, babyID, planID string) (*entity.VaccineReminder, error) {
	var reminder entity.VaccineReminder
	err := r.db.WithContext(ctx).
		Where("baby_id = ? AND plan_id = ? AND deleted_at IS NULL", babyID, planID).
		First(&reminder).Error

	if errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, nil
	}
	if err != nil {
		return nil, errors.Wrap(errors.DatabaseError, "failed to find vaccine reminder", err)
	}

	return &reminder, nil
}
