package service

import (
	"context"
	"fmt"
	"time"

	"github.com/go-co-op/gocron"
	"go.uber.org/zap"
	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/config"
)

// SchedulerService 定时任务服务
type SchedulerService struct {
	scheduler            *gocron.Scheduler
	vaccineRepo          repository.VaccineRecordRepository
	vaccineReminderRepo  repository.VaccineReminderRepository
	babyVaccinePlanRepo  repository.BabyVaccinePlanRepository
	feedingRecordRepo    repository.FeedingRecordRepository
	subscribeService     *SubscribeService
	strategyFactory      *FeedingReminderStrategyFactory
	logger               *zap.Logger
}

// NewSchedulerService 创建定时任务服务
func NewSchedulerService(
	vaccineRepo repository.VaccineRecordRepository,
	vaccineReminderRepo repository.VaccineReminderRepository,
	babyVaccinePlanRepo repository.BabyVaccinePlanRepository,
	feedingRecordRepo repository.FeedingRecordRepository,
	subscribeService *SubscribeService,
	cfg *config.Config,
	logger *zap.Logger,
) *SchedulerService {
	// 创建 gocron 调度器，使用本地时区
	scheduler := gocron.NewScheduler(time.Local)

	return &SchedulerService{
		scheduler:           scheduler,
		vaccineRepo:         vaccineRepo,
		vaccineReminderRepo: vaccineReminderRepo,
		babyVaccinePlanRepo: babyVaccinePlanRepo,
		feedingRecordRepo:   feedingRecordRepo,
		subscribeService:    subscribeService,
		strategyFactory:     NewFeedingReminderStrategyFactory(cfg),
		logger:              logger,
	}
}

// Start 启动定时任务
func (s *SchedulerService) Start() {
	// 启动调度器(用于一次性定时任务)
	s.scheduler.StartAsync()
	s.logger.Info("Scheduler service started (one-time task mode)")
}

// Stop 停止定时任务
func (s *SchedulerService) Stop() {
	s.scheduler.Stop()
	s.logger.Info("Scheduler service stopped")
}

// CheckVaccineReminders 检查疫苗提醒
func (s *SchedulerService) CheckVaccineReminders() error {
	ctx := context.Background()

	// 获取所有即将到期和已逾期的疫苗提醒
	reminders, err := s.vaccineReminderRepo.FindDueReminders(ctx)
	if err != nil {
		s.logger.Error("Failed to get due reminders", zap.Error(err))
		return err
	}

	s.logger.Info("Found vaccine reminders to process", zap.Int("count", len(reminders)))

	for _, reminder := range reminders {
		// 检查提醒状态
		if reminder.Status == "completed" || reminder.ReminderSent {
			continue
		}

		// 获取疫苗计划信息
		plan, err := s.babyVaccinePlanRepo.FindByID(ctx, reminder.PlanID)
		if err != nil {
			s.logger.Error("Failed to get vaccine plan",
				zap.String("planId", reminder.PlanID),
				zap.Error(err))
			continue
		}

		// 计算提醒状态
		now := time.Now()
		scheduledTime := time.Unix(reminder.ScheduledDate/1000, 0)
		daysUntilDue := int(scheduledTime.Sub(now).Hours() / 24)

		var status string
		var reminderMessage string

		if daysUntilDue < 0 {
			status = "overdue"
			reminderMessage = "已逾期"
		} else if daysUntilDue == 0 {
			status = "due"
			reminderMessage = "今天到期"
		} else if daysUntilDue <= 3 {
			status = "upcoming"
			reminderMessage = "即将到期"
		} else {
			status = "upcoming"
			reminderMessage = "提醒"
		}

		// 构造消息数据
		messageData := map[string]interface{}{
			"babyName":    reminder.BabyID, // TODO: 获取宝宝姓名
			"vaccineName": plan.VaccineType,
			"dueDate":     scheduledTime.Format("2006-01-02"),
			"location":    "请联系接种点",
			"doseNumber":  plan.DoseNumber,
		}

		// 直接发送订阅消息
		sendReq := &dto.SendMessageRequest{
			OpenID:     "", // TODO: 获取用户 OpenID
			TemplateID: "vaccine_reminder",
			Data:       messageData,
			Page:       "pages/vaccine/vaccine",
		}

		if err := s.subscribeService.SendSubscribeMessage(ctx, sendReq); err != nil {
			s.logger.Error("Failed to send vaccine reminder",
				zap.String("reminderId", reminder.ReminderID),
				zap.Error(err))
			continue
		}

		// 更新提醒状态
		reminder.Status = status
		reminder.ReminderSent = true
		if err := s.vaccineReminderRepo.Update(ctx, reminder); err != nil {
			s.logger.Error("Failed to update reminder status",
				zap.String("reminderId", reminder.ReminderID),
				zap.Error(err))
		}

		s.logger.Info("Vaccine reminder sent successfully",
			zap.String("reminderId", reminder.ReminderID),
			zap.String("status", status),
			zap.String("message", reminderMessage))
	}

	return nil
}


// AddFeedingReminderTask 添加喂养提醒一次性定时任务
//
// 在创建喂养记录成功后调用此方法，将在指定的 nextReminderTime 时间自动执行提醒
//
// 参数:
//   - ctx: 上下文
//   - record: 喂养记录实体，必须包含有效的 NextReminderTime
//
// 返回:
//   - jobTag: gocron 任务的标签，可用于后续取消任务
//   - err: 错误信息
func (s *SchedulerService) AddFeedingReminderTask(ctx context.Context, record *entity.FeedingRecord) (string, error) {
	// 检查是否设置了下次提醒时间
	if record.NextReminderTime == nil {
		s.logger.Debug("未设置下次提醒时间，跳过任务添加",
			zap.String("recordID", record.RecordID))
		return "", nil
	}

	// 计算执行时间
	executeTime := time.UnixMilli(*record.NextReminderTime)
	now := time.Now()

	// 如果执行时间已经过期，不添加任务
	if executeTime.Before(now) {
		s.logger.Warn("下次提醒时间已过期，跳过任务添加",
			zap.String("recordID", record.RecordID),
			zap.Time("executeTime", executeTime),
			zap.Time("now", now))
		return "", nil
	}

	// 创建提醒回调函数
	reminderJob := func() {
		// 创建新的上下文用于后台任务
		taskCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
		defer cancel()

		if err := s.executeFeedingReminder(taskCtx, record); err != nil {
			s.logger.Error("执行喂养提醒失败",
				zap.String("recordID", record.RecordID),
				zap.Error(err))
		}
	}

	// 创建任务标签用于识别和取消
	jobTag := fmt.Sprintf("feeding_reminder_%s", record.RecordID)

	// 使用 gocron 的 At() 方法添加一次性任务
	// gocron 会在指定时间执行，然后自动移除该任务
	job, err := s.scheduler.At(executeTime).Tag(jobTag).Do(reminderJob)
	if err != nil {
		s.logger.Error("添加喂养提醒任务失败",
			zap.String("recordID", record.RecordID),
			zap.Error(err))
		return "", err
	}

	s.logger.Info("添加喂养提醒任务成功",
		zap.String("recordID", record.RecordID),
		zap.String("jobTag", jobTag),
		zap.String("jobName", job.GetName()),
		zap.Time("executeTime", executeTime),
		zap.Duration("delay", executeTime.Sub(now)))

	return jobTag, nil
}

// CancelFeedingReminderTask 取消喂养提醒任务
//
// 如果用户编辑了喂养记录或取消了提醒，可调用此方法取消已添加的任务
func (s *SchedulerService) CancelFeedingReminderTask(jobTag string) {
	err := s.scheduler.RemoveByTag(jobTag)
	if err != nil {
		s.logger.Warn("取消喂养提醒任务失败",
			zap.String("jobTag", jobTag),
			zap.Error(err))
	} else {
		s.logger.Info("喂养提醒任务已取消", zap.String("jobTag", jobTag))
	}
}

// executeFeedingReminder 执行喂养提醒逻辑
func (s *SchedulerService) executeFeedingReminder(ctx context.Context, record *entity.FeedingRecord) error {
	s.logger.Info("开始执行喂养提醒",
		zap.String("recordID", record.RecordID),
		zap.String("babyID", record.BabyID),
		zap.String("feedingType", record.FeedingType))

	// 1. 根据喂养类型获取模板类型
	templateType := s.getTemplateType(record.FeedingType)
	if templateType == "" {
		s.logger.Warn("不支持的喂养类型，无法发送提醒",
			zap.String("feedingType", record.FeedingType))
		return nil
	}

	// 2. 检查用户是否已授权此提醒
	hasAuth, err := s.subscribeService.CheckAuthorizationStatus(ctx, record.CreateBy, templateType)
	if err != nil {
		s.logger.Error("检查授权状态失败", zap.Error(err))
		return err
	}

	if !hasAuth {
		s.logger.Info("用户未授权此提醒，跳过发送",
			zap.String("templateType", templateType),
			zap.String("openID", record.CreateBy))
		return nil
	}

	// 3. 构建提醒消息数据
	strategy, err := s.strategyFactory.GetStrategy(record)
	if err != nil {
		s.logger.Error("获取提醒策略失败", zap.Error(err))
		return err
	}

	lastFeedingTime := time.UnixMilli(record.Time)
	hoursSince := time.Since(lastFeedingTime).Hours()
	messageData := strategy.BuildMessageData(record, lastFeedingTime, hoursSince)

	// 4. 发送微信订阅消息
	sendReq := &dto.SendMessageRequest{
		OpenID:     record.CreateBy,
		TemplateID: strategy.GetTemplateID(),
		Data:       messageData,
		Page:       "pages/record/feeding/feeding",
	}

	err = s.subscribeService.SendSubscribeMessage(ctx, sendReq)
	if err != nil {
		s.logger.Error("发送微信消息失败",
			zap.Error(err),
			zap.String("recordID", record.RecordID))
		return err
	}

	// 5. 标记提醒已发送
	now := time.Now().UnixMilli()
	record.ReminderSent = true
	record.ReminderTime = &now

	err = s.feedingRecordRepo.Update(ctx, record)
	if err != nil {
		s.logger.Error("更新记录状态失败", zap.Error(err))
		return err
	}

	s.logger.Info("喂养提醒发送成功",
		zap.String("recordID", record.RecordID),
		zap.String("templateType", templateType))

	return nil
}

// getTemplateType 根据喂养类型获取微信订阅消息模板类型
func (s *SchedulerService) getTemplateType(feedingType string) string {
	switch feedingType {
	case entity.FeedingTypeBreast:
		return "breast_feeding_reminder"
	case entity.FeedingTypeBottle:
		return "bottle_feeding_reminder"
	case entity.FeedingTypeFood:
		return "food_feeding_reminder"
	default:
		return ""
	}
}
