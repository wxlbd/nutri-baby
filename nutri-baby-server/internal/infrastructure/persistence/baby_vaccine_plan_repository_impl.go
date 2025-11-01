package persistence

import (
	"context"
	"fmt"

	"github.com/google/uuid"
	"gorm.io/gorm"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
)

type babyVaccinePlanRepositoryImpl struct {
	db *gorm.DB
}

// NewBabyVaccinePlanRepository 创建宝宝疫苗计划仓储实现
func NewBabyVaccinePlanRepository(db *gorm.DB) repository.BabyVaccinePlanRepository {
	return &babyVaccinePlanRepositoryImpl{db: db}
}

func (r *babyVaccinePlanRepositoryImpl) Create(ctx context.Context, plan *entity.BabyVaccinePlan) error {
	return r.db.WithContext(ctx).Create(plan).Error
}

func (r *babyVaccinePlanRepositoryImpl) FindByID(ctx context.Context, planID string) (*entity.BabyVaccinePlan, error) {
	var plan entity.BabyVaccinePlan
	err := r.db.WithContext(ctx).Where("plan_id = ?", planID).First(&plan).Error
	if err != nil {
		return nil, err
	}
	return &plan, nil
}

func (r *babyVaccinePlanRepositoryImpl) FindByBabyID(ctx context.Context, babyID string) ([]*entity.BabyVaccinePlan, error) {
	var plans []*entity.BabyVaccinePlan
	err := r.db.WithContext(ctx).
		Where("baby_id = ?", babyID).
		Order("age_in_months ASC, dose_number ASC").
		Find(&plans).Error
	return plans, err
}

func (r *babyVaccinePlanRepositoryImpl) Update(ctx context.Context, plan *entity.BabyVaccinePlan) error {
	return r.db.WithContext(ctx).
		Model(&entity.BabyVaccinePlan{}).
		Where("plan_id = ?", plan.PlanID).
		Updates(map[string]interface{}{
			"vaccine_type":  plan.VaccineType,
			"vaccine_name":  plan.VaccineName,
			"description":   plan.Description,
			"age_in_months": plan.AgeInMonths,
			"dose_number":   plan.DoseNumber,
			"is_required":   plan.IsRequired,
			"reminder_days": plan.ReminderDays,
			"update_time":   plan.UpdateTime,
		}).Error
}

func (r *babyVaccinePlanRepositoryImpl) Delete(ctx context.Context, planID string) error {
	return r.db.WithContext(ctx).
		Where("plan_id = ?", planID).
		Delete(&entity.BabyVaccinePlan{}).Error
}

func (r *babyVaccinePlanRepositoryImpl) BatchCreate(ctx context.Context, plans []*entity.BabyVaccinePlan) error {
	if len(plans) == 0 {
		return nil
	}
	return r.db.WithContext(ctx).CreateInBatches(plans, 100).Error
}

func (r *babyVaccinePlanRepositoryImpl) InitializeFromTemplates(ctx context.Context, babyID, createBy string) error {
	// 1. 检查是否已初始化
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.BabyVaccinePlan{}).
		Where("baby_id = ?", babyID).
		Count(&count).Error
	if err != nil {
		return err
	}
	if count > 0 {
		return fmt.Errorf("宝宝 %s 已初始化疫苗计划", babyID)
	}

	// 2. 获取所有模板
	var templates []*entity.VaccinePlanTemplate
	err = r.db.WithContext(ctx).
		Order("sort_order ASC, age_in_months ASC").
		Find(&templates).Error
	if err != nil {
		return err
	}

	if len(templates) == 0 {
		return fmt.Errorf("未找到疫苗计划模板,请先初始化系统数据")
	}

	// 3. 基于模板创建宝宝的疫苗计划
	babyPlans := make([]*entity.BabyVaccinePlan, 0, len(templates))
	for _, tmpl := range templates {
		templateID := tmpl.TemplateID
		babyPlans = append(babyPlans, &entity.BabyVaccinePlan{
			PlanID:       uuid.NewString(),
			BabyID:       babyID,
			TemplateID:   &templateID,
			VaccineType:  tmpl.VaccineType,
			VaccineName:  tmpl.VaccineName,
			Description:  tmpl.Description,
			AgeInMonths:  tmpl.AgeInMonths,
			DoseNumber:   tmpl.DoseNumber,
			IsRequired:   tmpl.IsRequired,
			ReminderDays: tmpl.ReminderDays,
			IsCustom:     false,
			CreateBy:     createBy,
		})
	}

	// 4. 批量创建
	return r.BatchCreate(ctx, babyPlans)
}

func (r *babyVaccinePlanRepositoryImpl) CountByBabyID(ctx context.Context, babyID string) (int64, error) {
	var count int64
	err := r.db.WithContext(ctx).
		Model(&entity.BabyVaccinePlan{}).
		Where("baby_id = ?", babyID).
		Count(&count).Error
	return count, err
}
