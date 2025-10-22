package service

import (
	"context"
	"fmt"

	"github.com/google/uuid"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	errs "github.com/wxlbd/nutri-baby-server/pkg/errors"
)

type VaccinePlanService struct {
	babyVaccinePlanRepo repository.BabyVaccinePlanRepository
	vaccineRecordRepo   repository.VaccineRecordRepository
	vaccineReminderRepo repository.VaccineReminderRepository
	babyRepo            repository.BabyRepository
}

// NewVaccinePlanService 创建疫苗计划服务
func NewVaccinePlanService(
	babyVaccinePlanRepo repository.BabyVaccinePlanRepository,
	vaccineRecordRepo repository.VaccineRecordRepository,
	vaccineReminderRepo repository.VaccineReminderRepository,
	babyRepo repository.BabyRepository,
) *VaccinePlanService {
	return &VaccinePlanService{
		babyVaccinePlanRepo: babyVaccinePlanRepo,
		vaccineRecordRepo:   vaccineRecordRepo,
		vaccineReminderRepo: vaccineReminderRepo,
		babyRepo:            babyRepo,
	}
}

// InitializePlansForBaby 为宝宝初始化疫苗计划(从模板)
func (s *VaccinePlanService) InitializePlansForBaby(ctx context.Context, babyID, openID string, force bool) (*dto.InitializeVaccinePlansResponse, error) {
	// 1. 验证宝宝是否存在
	baby, err := s.babyRepo.FindByID(ctx, babyID)
	if err != nil {
		return nil, errs.ErrNotFound
	}

	// 2. 如果强制初始化,先删除已有计划
	if force {
		// TODO: 实现批量删除逻辑
	}

	// 3. 检查是否已初始化
	count, err := s.babyVaccinePlanRepo.CountByBabyID(ctx, babyID)
	if err != nil {
		return nil, err
	}
	if count > 0 && !force {
		// 已初始化,直接返回现有计划
		plans, err := s.babyVaccinePlanRepo.FindByBabyID(ctx, babyID)
		if err != nil {
			return nil, err
		}
		return &dto.InitializeVaccinePlansResponse{
			TotalPlans: len(plans),
			Plans:      s.toDTOList(plans),
			Message:    fmt.Sprintf("宝宝 %s 已有 %d 条疫苗计划", baby.Name, len(plans)),
		}, nil
	}

	// 4. 从模板初始化
	err = s.babyVaccinePlanRepo.InitializeFromTemplates(ctx, babyID, openID)
	if err != nil {
		return nil, err
	}

	// 5. 获取初始化后的计划
	plans, err := s.babyVaccinePlanRepo.FindByBabyID(ctx, babyID)
	if err != nil {
		return nil, err
	}

	// 6. 为宝宝生成疫苗提醒
	err = s.generateRemindersForBaby(ctx, baby, plans)
	if err != nil {
		// 提醒生成失败不影响计划初始化
		fmt.Printf("Warning: 生成疫苗提醒失败: %v\n", err)
	}

	return &dto.InitializeVaccinePlansResponse{
		TotalPlans: len(plans),
		Plans:      s.toDTOList(plans),
		Message:    fmt.Sprintf("已为宝宝 %s 初始化 %d 条疫苗计划", baby.Name, len(plans)),
	}, nil
}

// GetPlansForBaby 获取宝宝的疫苗计划列表
func (s *VaccinePlanService) GetPlansForBaby(ctx context.Context, babyID string) (*dto.GetVaccinePlansResponse, error) {
	// 1. 获取计划列表
	plans, err := s.babyVaccinePlanRepo.FindByBabyID(ctx, babyID)
	if err != nil {
		return nil, err
	}

	// 2. 统计完成数量
	completed, err := s.vaccineRecordRepo.CountCompleted(ctx, babyID)
	if err != nil {
		return nil, err
	}

	total := len(plans)
	percentage := 0
	if total > 0 {
		percentage = int(float64(completed) / float64(total) * 100)
	}

	return &dto.GetVaccinePlansResponse{
		Plans:      s.toDTOList(plans),
		Total:      total,
		Completed:  int(completed),
		Percentage: percentage,
	}, nil
}

// CreatePlan 创建自定义疫苗计划
func (s *VaccinePlanService) CreatePlan(ctx context.Context, babyID, openID string, req *dto.CreateBabyVaccinePlanRequest) (*dto.VaccinePlanDTO, error) {
	// 1. 验证宝宝是否存在
	_, err := s.babyRepo.FindByID(ctx, babyID)
	if err != nil {
		return nil, errs.ErrNotFound.WithMessage("宝宝不存在")
	}

	// 2. 创建计划
	plan := &entity.BabyVaccinePlan{
		PlanID:       uuid.NewString(),
		BabyID:       babyID,
		VaccineType:  req.VaccineType,
		VaccineName:  req.VaccineName,
		Description:  req.Description,
		AgeInMonths:  req.AgeInMonths,
		DoseNumber:   req.DoseNumber,
		IsRequired:   req.IsRequired,
		ReminderDays: req.ReminderDays,
		IsCustom:     true,
		CreateBy:     openID,
	}

	err = s.babyVaccinePlanRepo.Create(ctx, plan)
	if err != nil {
		return nil, err
	}

	return s.toDTO(plan), nil
}

// UpdatePlan 更新疫苗计划
func (s *VaccinePlanService) UpdatePlan(ctx context.Context, planID string, req *dto.UpdateBabyVaccinePlanRequest) (*dto.VaccinePlanDTO, error) {
	// 1. 查找计划
	plan, err := s.babyVaccinePlanRepo.FindByID(ctx, planID)
	if err != nil {
		return nil, errs.ErrNotFound.WithMessage("疫苗计划不存在")
	}

	// 2. 更新字段
	if req.VaccineName != "" {
		plan.VaccineName = req.VaccineName
	}
	if req.Description != "" {
		plan.Description = req.Description
	}
	if req.AgeInMonths > 0 {
		plan.AgeInMonths = req.AgeInMonths
	}
	if req.DoseNumber > 0 {
		plan.DoseNumber = req.DoseNumber
	}
	plan.IsRequired = req.IsRequired
	if req.ReminderDays >= 0 {
		plan.ReminderDays = req.ReminderDays
	}

	// 3. 保存更新
	err = s.babyVaccinePlanRepo.Update(ctx, plan)
	if err != nil {
		return nil, err
	}

	return s.toDTO(plan), nil
}

// DeletePlan 删除疫苗计划
func (s *VaccinePlanService) DeletePlan(ctx context.Context, planID string) error {
	// 1. 查找计划
	plan, err := s.babyVaccinePlanRepo.FindByID(ctx, planID)
	if err != nil {
		return errs.ErrNotFound.WithMessage("疫苗计划不存在")
	}

	// 2. 检查是否已有接种记录
	record, err := s.vaccineRecordRepo.FindByBabyAndPlan(ctx, plan.BabyID, planID)
	if err == nil && record != nil {
		return errs.ErrConflict.WithMessage("该疫苗计划已有接种记录,无法删除")
	}

	// 3. 删除计划
	return s.babyVaccinePlanRepo.Delete(ctx, planID)
}

// generateRemindersForBaby 为宝宝生成疫苗提醒
func (s *VaccinePlanService) generateRemindersForBaby(ctx context.Context, baby *entity.Baby, plans []*entity.BabyVaccinePlan) error {
	birthDate := baby.BirthDate
	reminders := make([]*entity.VaccineReminder, 0, len(plans))

	for _, plan := range plans {
		// 计算预定接种日期
		scheduledDate := birthDate + int64(plan.AgeInMonths)*30*24*60*60*1000

		// 检查是否已有记录
		record, _ := s.vaccineRecordRepo.FindByBabyAndPlan(ctx, baby.BabyID, plan.PlanID)
		if record != nil {
			continue // 已接种,跳过
		}

		reminder := &entity.VaccineReminder{
			ReminderID:    uuid.NewString(),
			BabyID:        baby.BabyID,
			PlanID:        plan.PlanID,
			VaccineName:   plan.VaccineName,
			DoseNumber:    plan.DoseNumber,
			ScheduledDate: scheduledDate,
			ReminderSent:  false,
		}
		reminder.UpdateStatus()
		reminders = append(reminders, reminder)
	}

	if len(reminders) > 0 {
		return s.vaccineReminderRepo.BatchCreate(ctx, reminders)
	}
	return nil
}

// toDTO 转换为DTO
func (s *VaccinePlanService) toDTO(plan *entity.BabyVaccinePlan) *dto.VaccinePlanDTO {
	templateID := ""
	if plan.TemplateID != nil {
		templateID = *plan.TemplateID
	}

	return &dto.VaccinePlanDTO{
		PlanID:       plan.PlanID,
		VaccineType:  plan.VaccineType,
		VaccineName:  plan.VaccineName,
		Description:  plan.Description,
		AgeInMonths:  plan.AgeInMonths,
		DoseNumber:   plan.DoseNumber,
		IsRequired:   plan.IsRequired,
		ReminderDays: plan.ReminderDays,
		IsCustom:     plan.IsCustom,
		TemplateID:   templateID,
	}
}

// toDTOList 转换为DTO列表
func (s *VaccinePlanService) toDTOList(plans []*entity.BabyVaccinePlan) []dto.VaccinePlanDTO {
	result := make([]dto.VaccinePlanDTO, 0, len(plans))
	for _, plan := range plans {
		result = append(result, *s.toDTO(plan))
	}
	return result
}
