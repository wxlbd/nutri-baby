package repository

import (
	"context"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
)

// VaccinePlanTemplateRepository 疫苗计划模板仓储接口(国家免疫规划标准)
type VaccinePlanTemplateRepository interface {
	// FindAll 查找所有模板
	FindAll(ctx context.Context) ([]*entity.VaccinePlanTemplate, error)
	// FindByID 根据ID查找模板
	FindByID(ctx context.Context, templateID string) (*entity.VaccinePlanTemplate, error)
	// Create 创建模板(系统初始化)
	Create(ctx context.Context, template *entity.VaccinePlanTemplate) error
	// BatchCreate 批量创建模板
	BatchCreate(ctx context.Context, templates []*entity.VaccinePlanTemplate) error
}

// BabyVaccinePlanRepository 宝宝疫苗计划仓储接口(可自定义)
type BabyVaccinePlanRepository interface {
	// Create 创建宝宝疫苗计划
	Create(ctx context.Context, plan *entity.BabyVaccinePlan) error
	// FindByID 根据ID查找计划
	FindByID(ctx context.Context, planID string) (*entity.BabyVaccinePlan, error)
	// FindByBabyID 查找宝宝的所有疫苗计划
	FindByBabyID(ctx context.Context, babyID string) ([]*entity.BabyVaccinePlan, error)
	// Update 更新计划
	Update(ctx context.Context, plan *entity.BabyVaccinePlan) error
	// Delete 删除计划(软删除)
	Delete(ctx context.Context, planID string) error
	// BatchCreate 批量创建计划(从模板初始化)
	BatchCreate(ctx context.Context, plans []*entity.BabyVaccinePlan) error
	// InitializeFromTemplates 从模板为宝宝初始化疫苗计划
	InitializeFromTemplates(ctx context.Context, babyID, createBy string) error
	// CountByBabyID 统计宝宝的疫苗计划数量
	CountByBabyID(ctx context.Context, babyID string) (int64, error)
}

// VaccineRecordRepository 疫苗接种记录仓储接口
type VaccineRecordRepository interface {
	// Create 创建记录
	Create(ctx context.Context, record *entity.VaccineRecord) error
	// FindByID 根据ID查找记录
	FindByID(ctx context.Context, recordID string) (*entity.VaccineRecord, error)
	// FindByBabyID 查找宝宝的疫苗记录(分页)
	FindByBabyID(ctx context.Context, babyID string, startTime, endTime int64, vaccineType string, page, pageSize int) ([]*entity.VaccineRecord, int64, error)
	// FindByBabyAndPlan 查找宝宝特定计划的接种记录
	FindByBabyAndPlan(ctx context.Context, babyID, planID string) (*entity.VaccineRecord, error)
	// Update 更新记录
	Update(ctx context.Context, record *entity.VaccineRecord) error
	// Delete 删除记录
	Delete(ctx context.Context, recordID string) error
	// FindUpdatedAfter 查找指定时间后更新的记录(用于同步)
	FindUpdatedAfter(ctx context.Context, familyID string, timestamp int64) ([]*entity.VaccineRecord, error)
	// CountCompleted 统计已完成接种数量
	CountCompleted(ctx context.Context, babyID string) (int64, error)
}

// VaccineReminderRepository 疫苗提醒仓储接口
type VaccineReminderRepository interface {
	// Create 创建提醒
	Create(ctx context.Context, reminder *entity.VaccineReminder) error
	// FindByID 根据ID查找提醒
	FindByID(ctx context.Context, reminderID string) (*entity.VaccineReminder, error)
	// FindByBabyID 查找宝宝的所有提醒
	FindByBabyID(ctx context.Context, babyID string) ([]*entity.VaccineReminder, error)
	// FindByStatus 根据状态查找提醒
	FindByStatus(ctx context.Context, babyID, status string, limit int) ([]*entity.VaccineReminder, error)
	// FindDueReminders 查找即将到期和已逾期的提醒(用于推送)
	FindDueReminders(ctx context.Context) ([]*entity.VaccineReminder, error)
	// Update 更新提醒
	Update(ctx context.Context, reminder *entity.VaccineReminder) error
	// UpdateStatus 更新提醒状态
	UpdateStatus(ctx context.Context, reminderID, status string) error
	// MarkSent 标记提醒已发送
	MarkSent(ctx context.Context, reminderID string) error
	// CountByStatus 统计各状态的提醒数量
	CountByStatus(ctx context.Context, babyID string) (map[string]int64, error)
	// BatchCreate 批量创建提醒(初始化宝宝疫苗计划时使用)
	BatchCreate(ctx context.Context, reminders []*entity.VaccineReminder) error
	// FindByBabyAndPlan 查找宝宝特定计划的提醒
	FindByBabyAndPlan(ctx context.Context, babyID, planID string) (*entity.VaccineReminder, error)
}
