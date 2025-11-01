package service

import (
	"context"
	"time"

	"github.com/google/uuid"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// VaccineService 疫苗服务
type VaccineService struct {
	vaccinePlanRepo     repository.BabyVaccinePlanRepository
	vaccineRecordRepo   repository.VaccineRecordRepository
	vaccineReminderRepo repository.VaccineReminderRepository
	babyRepo            repository.BabyRepository
}

// NewVaccineService 创建疫苗服务
func NewVaccineService(
	vaccinePlanRepo repository.BabyVaccinePlanRepository,
	vaccineRecordRepo repository.VaccineRecordRepository,
	vaccineReminderRepo repository.VaccineReminderRepository,
	babyRepo repository.BabyRepository,
) *VaccineService {
	return &VaccineService{
		vaccinePlanRepo:     vaccinePlanRepo,
		vaccineRecordRepo:   vaccineRecordRepo,
		vaccineReminderRepo: vaccineReminderRepo,
		babyRepo:            babyRepo,
	}
}

// GetVaccinePlans 获取宝宝的疫苗计划
func (s *VaccineService) GetVaccinePlans(ctx context.Context, babyID string) ([]*dto.VaccinePlanDTO, error) {
	// 查找宝宝
	baby, err := s.babyRepo.FindByID(ctx, babyID)
	if err != nil {
		return nil, err
	}

	// 获取所有疫苗计划
	plans, err := s.vaccinePlanRepo.FindByBabyID(ctx, babyID)
	if err != nil {
		return nil, err
	}

	// 获取宝宝的接种记录
	records, _, err := s.vaccineRecordRepo.FindByBabyID(ctx, babyID, 0, 0, "", 1, 1000)
	if err != nil {
		return nil, err
	}

	// 创建已完成的计划ID映射
	completedMap := make(map[string]bool)
	for _, record := range records {
		completedMap[record.PlanID] = true
	}

	// 转换为DTO
	birthDate, _ := time.Parse("2006-01-02", baby.BirthDate)
	result := make([]*dto.VaccinePlanDTO, 0, len(plans))

	for _, plan := range plans {
		// 计算预定日期
		scheduledDate := birthDate.AddDate(0, plan.AgeInMonths, 0).UnixMilli()

		// 确定状态
		status := "pending"
		if completedMap[plan.PlanID] {
			status = "completed"
		} else if scheduledDate < time.Now().UnixMilli() {
			status = "overdue"
		}

		result = append(result, &dto.VaccinePlanDTO{
			PlanID:        plan.PlanID,
			VaccineType:   plan.VaccineType,
			VaccineName:   plan.VaccineName,
			Description:   plan.Description,
			AgeInMonths:   plan.AgeInMonths,
			DoseNumber:    plan.DoseNumber,
			IsRequired:    plan.IsRequired,
			ReminderDays:  plan.ReminderDays,
			ScheduledDate: scheduledDate,
			Status:        status,
		})
	}

	return result, nil
}

// CreateVaccineRecord 创建疫苗接种记录
func (s *VaccineService) CreateVaccineRecord(
	ctx context.Context,
	babyID, createBy string,
	req *dto.CreateVaccineRecordRequest,
) (*dto.VaccineRecordDTO, error) {
	// 验证宝宝存在
	_, err := s.babyRepo.FindByID(ctx, babyID)
	if err != nil {
		return nil, err
	}

	// 创建记录
	record := &entity.VaccineRecord{
		RecordID:    uuid.New().String(),
		BabyID:      babyID,
		PlanID:      req.PlanID,
		VaccineType: req.VaccineType,
		VaccineName: req.VaccineName,
		DoseNumber:  req.DoseNumber,
		VaccineDate: req.VaccineDate,
		Hospital:    req.Hospital,
		BatchNumber: req.BatchNumber,
		Doctor:      req.Doctor,
		Reaction:    req.Reaction,
		Note:        req.Note,
		CreateBy:    createBy,
	}

	if err := s.vaccineRecordRepo.Create(ctx, record); err != nil {
		return nil, err
	}

	// 更新提醒状态为completed
	reminder, err := s.vaccineReminderRepo.FindByBabyAndPlan(ctx, babyID, req.PlanID)
	if err == nil && reminder != nil {
		_ = s.vaccineReminderRepo.UpdateStatus(ctx, reminder.ReminderID, "completed")
	}

	return &dto.VaccineRecordDTO{
		RecordID:    record.RecordID,
		BabyID:      record.BabyID,
		PlanID:      record.PlanID,
		VaccineType: record.VaccineType,
		VaccineName: record.VaccineName,
		DoseNumber:  record.DoseNumber,
		VaccineDate: record.VaccineDate,
		Hospital:    record.Hospital,
		BatchNumber: record.BatchNumber,
		Doctor:      record.Doctor,
		Reaction:    record.Reaction,
		Note:        record.Note,
		CreateBy:    record.CreateBy,
		CreateTime:  record.CreateTime,
	}, nil
}

// GetVaccineReminders 获取疫苗提醒列表
func (s *VaccineService) GetVaccineReminders(
	ctx context.Context,
	babyID, status string,
	limit int,
) ([]*dto.VaccineReminderDTO, error) {
	// 如果指定了特定状态，先只查询该状态的记录
	var reminders []*entity.VaccineReminder
	var err error

	if status != "" {
		reminders, err = s.vaccineReminderRepo.FindByStatus(ctx, babyID, status, limit)
		if err != nil {
			return nil, err
		}
	} else {
		// 如果没有指定状态，获取所有未完成的提醒
		reminders, err = s.vaccineReminderRepo.FindByBabyID(ctx, babyID)
		if err != nil {
			return nil, err
		}
		// 过滤掉已完成的提醒
		filtered := make([]*entity.VaccineReminder, 0, len(reminders))
		for _, r := range reminders {
			if r.Status != entity.ReminderStatusCompleted {
				filtered = append(filtered, r)
			}
		}
		reminders = filtered
	}

	result := make([]*dto.VaccineReminderDTO, 0, len(reminders))
	for _, reminder := range reminders {
		// 实时更新提醒状态
		oldStatus := reminder.Status
		reminder.UpdateStatus()

		// 如果状态发生了变化且提醒未完成，则保存到数据库
		if oldStatus != reminder.Status && reminder.Status != entity.ReminderStatusCompleted {
			_ = s.vaccineReminderRepo.UpdateStatus(ctx, reminder.ReminderID, reminder.Status)
		}

		// 只返回与请求状态匹配的提醒
		if status != "" && reminder.Status != status {
			continue
		}

		result = append(result, &dto.VaccineReminderDTO{
			ReminderID:    reminder.ReminderID,
			BabyID:        reminder.BabyID,
			PlanID:        reminder.PlanID,
			VaccineName:   reminder.VaccineName,
			DoseNumber:    reminder.DoseNumber,
			ScheduledDate: reminder.ScheduledDate,
			Status:        reminder.Status,
			DaysUntilDue:  reminder.DaysUntilDue(),
			ReminderSent:  reminder.ReminderSent,
			CreateTime:    reminder.CreateTime,
		})
	}

	return result, nil
}

// GetVaccineStatistics 获取疫苗接种统计
func (s *VaccineService) GetVaccineStatistics(
	ctx context.Context,
	babyID string,
) (*dto.VaccineStatisticsDTO, error) {
	// 获取状态统计
	statusCounts, err := s.vaccineReminderRepo.CountByStatus(ctx, babyID)
	if err != nil {
		return nil, err
	}

	total := int64(0)
	for _, count := range statusCounts {
		total += count
	}

	completed := statusCounts["completed"]
	pending := total - completed
	overdue := statusCounts["overdue"]

	percentage := 0
	if total > 0 {
		percentage = int(completed * 100 / total)
	}

	// 获取下一个待接种疫苗
	upcomingReminders, err := s.vaccineReminderRepo.FindByStatus(ctx, babyID, "due", 1)
	var nextVaccine *dto.NextVaccineDTO
	if err == nil && len(upcomingReminders) > 0 {
		r := upcomingReminders[0]
		nextVaccine = &dto.NextVaccineDTO{
			VaccineName:   r.VaccineName,
			DoseNumber:    r.DoseNumber,
			ScheduledDate: r.ScheduledDate,
			DaysUntilDue:  r.DaysUntilDue(),
		}
	}

	// 获取最近接种记录
	recentRecords, _, err := s.vaccineRecordRepo.FindByBabyID(ctx, babyID, 0, 0, "", 1, 5)
	if err != nil {
		return nil, err
	}

	recordDTOs := make([]dto.VaccineRecordDTO, 0, len(recentRecords))
	for _, r := range recentRecords {
		recordDTOs = append(recordDTOs, dto.VaccineRecordDTO{
			VaccineName: r.VaccineName,
			VaccineDate: r.VaccineDate,
			Hospital:    r.Hospital,
		})
	}

	return &dto.VaccineStatisticsDTO{
		Total:         total,
		Completed:     completed,
		Pending:       pending,
		Overdue:       overdue,
		Percentage:    percentage,
		NextVaccine:   nextVaccine,
		RecentRecords: recordDTOs,
	}, nil
}

// InitializeVaccineReminders 初始化宝宝的疫苗提醒
func (s *VaccineService) InitializeVaccineReminders(ctx context.Context, babyID string) error {
	baby, err := s.babyRepo.FindByID(ctx, babyID)
	if err != nil {
		return err
	}

	plans, err := s.vaccinePlanRepo.FindByBabyID(ctx, babyID)
	if err != nil {
		return err
	}

	birthDate, err := time.Parse("2006-01-02", baby.BirthDate)
	if err != nil {
		return errors.Wrap(errors.ParamError, "invalid birth date", err)
	}

	reminders := make([]*entity.VaccineReminder, 0, len(plans))
	for _, plan := range plans {
		// 检查是否已存在提醒
		existing, _ := s.vaccineReminderRepo.FindByBabyAndPlan(ctx, babyID, plan.PlanID)
		if existing != nil {
			continue
		}

		// 计算预定日期
		scheduledDate := birthDate.AddDate(0, plan.AgeInMonths, 0)

		reminder := &entity.VaccineReminder{
			ReminderID:    uuid.New().String(),
			BabyID:        babyID,
			PlanID:        plan.PlanID,
			VaccineName:   plan.VaccineName,
			DoseNumber:    plan.DoseNumber,
			ScheduledDate: scheduledDate.UnixMilli(),
			Status:        "upcoming",
			ReminderSent:  false,
		}

		// 更新状态
		reminder.UpdateStatus()

		reminders = append(reminders, reminder)
	}

	if len(reminders) > 0 {
		return s.vaccineReminderRepo.BatchCreate(ctx, reminders)
	}

	return nil
}

// GetVaccineRecords 获取疫苗接种记录
func (s *VaccineService) GetVaccineRecords(ctx context.Context, babyID string, page, pageSize int) ([]*dto.VaccineRecordDTO, error) {
	records, _, err := s.vaccineRecordRepo.FindByBabyID(ctx, babyID, 0, 0, "", page, pageSize)
	if err != nil {
		return nil, err
	}

	result := make([]*dto.VaccineRecordDTO, 0, len(records))
	for _, record := range records {
		result = append(result, &dto.VaccineRecordDTO{
			RecordID:    record.RecordID,
			BabyID:      record.BabyID,
			PlanID:      record.PlanID,
			VaccineType: record.VaccineType,
			VaccineName: record.VaccineName,
			DoseNumber:  record.DoseNumber,
			VaccineDate: record.VaccineDate,
			Hospital:    record.Hospital,
			BatchNumber: record.BatchNumber,
			Doctor:      record.Doctor,
			Reaction:    record.Reaction,
			Note:        record.Note,
			CreateBy:    record.CreateBy,
			CreateTime:  record.CreateTime,
		})
	}
	return result, nil
}
