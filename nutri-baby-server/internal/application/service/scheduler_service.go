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
	babyRepo             repository.BabyRepository
	babyCollaboratorRepo repository.BabyCollaboratorRepository
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
	babyRepo repository.BabyRepository,
	babyCollaboratorRepo repository.BabyCollaboratorRepository,
	subscribeService *SubscribeService,
	cfg *config.Config,
	logger *zap.Logger,
) *SchedulerService {
	// 创建 gocron 调度器，使用本地时区
	scheduler := gocron.NewScheduler(time.Local)

	return &SchedulerService{
		scheduler:            scheduler,
		vaccineRepo:          vaccineRepo,
		vaccineReminderRepo:  vaccineReminderRepo,
		babyVaccinePlanRepo:  babyVaccinePlanRepo,
		feedingRecordRepo:    feedingRecordRepo,
		babyRepo:             babyRepo,
		babyCollaboratorRepo: babyCollaboratorRepo,
		subscribeService:     subscribeService,
		strategyFactory:      NewFeedingReminderStrategyFactory(cfg),
		logger:               logger,
	}
}

// Start 启动定时任务
func (s *SchedulerService) Start() {
	// 【测试模式】每1分钟检查喂养提醒 (生产环境改为: 每3分钟)
	_, err := s.scheduler.Every(1).Minute().Do(func() {
		s.logger.Info("Starting feeding reminder check...")
		if err := s.CheckFeedingReminders(); err != nil {
			s.logger.Error("Feeding reminder check failed", zap.Error(err))
		}
	})
	if err != nil {
		s.logger.Error("Failed to schedule feeding reminder check", zap.Error(err))
	}

	s.scheduler.StartAsync()
	s.logger.Info("Scheduler service started (TEST MODE: runs every 1 minute)")
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

// CheckFeedingReminders 检查喂养提醒
func (s *SchedulerService) CheckFeedingReminders() error {
	s.logger.Info("🔔 [CheckFeedingReminders] ===== START =====")
	s.logger.Info("⏰ [CheckFeedingReminders] 定时任务触发时间", zap.Time("triggerTime", time.Now()))

	ctx := context.Background()

	// 1. 获取所有宝宝
	s.logger.Info("🔍 [CheckFeedingReminders] STEP 1 - 获取所有宝宝列表")
	babies, err := s.babyRepo.FindAll(ctx)
	if err != nil {
		s.logger.Error("❌ [CheckFeedingReminders] 获取宝宝列表失败", zap.Error(err))
		return err
	}

	if len(babies) == 0 {
		s.logger.Info("ℹ️ [CheckFeedingReminders] 系统中没有宝宝,跳过检查")
		return nil
	}

	s.logger.Info("✅ [CheckFeedingReminders] 找到宝宝",
		zap.Int("babyCount", len(babies)),
		zap.Strings("babyIds", getBabyIDs(babies)),
	)

	now := time.Now()
	startTime := now.Add(-24 * time.Hour).UnixMilli() // 查询最近24小时
	endTime := now.UnixMilli()

	s.logger.Info("📅 [CheckFeedingReminders] 查询时间范围",
		zap.Time("startTime", time.UnixMilli(startTime)),
		zap.Time("endTime", time.UnixMilli(endTime)),
	)

	for i, baby := range babies {
		s.logger.Info("👶 [CheckFeedingReminders] 处理宝宝",
			zap.Int("index", i+1),
			zap.Int("total", len(babies)),
			zap.String("babyId", baby.BabyID),
			zap.String("babyName", baby.Name),
		)

		// 2. 获取该宝宝最近的喂养记录
		s.logger.Info("🔍 [CheckFeedingReminders] STEP 2 - 查询最近喂养记录",
			zap.String("babyId", baby.BabyID),
		)

		records, _, err := s.feedingRecordRepo.FindByBabyID(ctx, baby.BabyID, startTime, endTime, 1, 1)
		if err != nil {
			s.logger.Error("❌ [CheckFeedingReminders] 查询喂养记录失败",
				zap.String("babyId", baby.BabyID),
				zap.Error(err))
			continue
		}

		// 如果没有喂养记录，跳过
		if len(records) == 0 {
			s.logger.Info("ℹ️ [CheckFeedingReminders] 该宝宝没有喂养记录,跳过",
				zap.String("babyId", baby.BabyID),
				zap.String("babyName", baby.Name),
			)
			continue
		}

		lastFeeding := records[0]
		lastFeedingTime := time.UnixMilli(lastFeeding.Time)
		hoursSinceLastFeeding := now.Sub(lastFeedingTime).Hours()

		s.logger.Info("📊 [CheckFeedingReminders] 上次喂养时间分析",
			zap.String("babyId", baby.BabyID),
			zap.String("babyName", baby.Name),
			zap.Time("lastFeedingTime", lastFeedingTime),
			zap.Float64("hoursSinceLastFeeding", hoursSinceLastFeeding),
			zap.String("feedingType", getLastFeedingSide(lastFeeding)),
			zap.Any("record", lastFeeding),
		)

		// 如果距离上次喂养超过3小时，发送提醒
		// TODO: 改为用户自定义时间
		reminderThreshold := 0.0016 // 测试环境: ~1分钟, 生产环境应改为: 3.0 小时
		s.logger.Info("⚙️ [CheckFeedingReminders] 提醒阈值配置",
			zap.Float64("thresholdHours", reminderThreshold),
			zap.Bool("shouldRemind", hoursSinceLastFeeding >= reminderThreshold),
		)

		if hoursSinceLastFeeding >= reminderThreshold {
			s.logger.Info("⏰ [CheckFeedingReminders] 需要发送喂养提醒",
				zap.String("babyId", baby.BabyID),
				zap.String("babyName", baby.Name),
				zap.Float64("hoursSinceLastFeeding", hoursSinceLastFeeding))

			// 3. 获取宝宝的协作者（家庭成员）
			s.logger.Info("🔍 [CheckFeedingReminders] STEP 3 - 查询宝宝的协作者",
				zap.String("babyId", baby.BabyID),
			)

			collaborators, err := s.babyCollaboratorRepo.FindByBabyID(ctx, baby.BabyID)
			if err != nil {
				s.logger.Error("❌ [CheckFeedingReminders] 查询协作者失败",
					zap.String("babyId", baby.BabyID),
					zap.Error(err))
				continue
			}

			if len(collaborators) == 0 {
				s.logger.Warn("⚠️ [CheckFeedingReminders] 该宝宝没有协作者,无法发送提醒",
					zap.String("babyId", baby.BabyID),
					zap.String("babyName", baby.Name),
				)
				continue
			}

			s.logger.Info("✅ [CheckFeedingReminders] 找到协作者",
				zap.String("babyId", baby.BabyID),
				zap.Int("collaboratorCount", len(collaborators)),
			)

			// 4. 根据喂养类型获取策略
			strategy, err := s.strategyFactory.GetStrategy(lastFeeding)
			if err != nil {
				s.logger.Error("❌ [CheckFeedingReminders] 获取喂养提醒策略失败",
					zap.String("babyId", baby.BabyID),
					zap.Error(err))
				continue
			}
			templateType := strategy.GetTemplateType()

			s.logger.Info("🎯 [CheckFeedingReminders] 获取喂养提醒策略",
				zap.String("babyId", baby.BabyID),
				zap.String("templateType", templateType),
			)

			// 5. 检查每个协作者的授权状态并发送提醒
			for j, collaborator := range collaborators {
				s.logger.Info("👤 [CheckFeedingReminders] 处理协作者",
					zap.Int("collaboratorIndex", j+1),
					zap.Int("collaboratorTotal", len(collaborators)),
					zap.String("openid", collaborator.OpenID),
					zap.String("babyId", baby.BabyID),
				)

				// 检查用户是否有可用的授权
				s.logger.Info("🔍 [CheckFeedingReminders] STEP 5 - 检查用户授权状态",
					zap.String("openid", collaborator.OpenID),
					zap.String("templateType", templateType),
				)

				hasAuth, err := s.subscribeService.CheckAuthorizationStatus(ctx, collaborator.OpenID, templateType)
				if err != nil {
					s.logger.Error("❌ [CheckFeedingReminders] 检查授权状态失败",
						zap.String("openid", collaborator.OpenID),
						zap.Error(err))
					continue
				}

				if !hasAuth {
					s.logger.Warn("⚠️ [CheckFeedingReminders] 用户没有可用授权,跳过",
						zap.String("openid", collaborator.OpenID),
						zap.String("babyId", baby.BabyID))
					continue
				}

				s.logger.Info("✅ [CheckFeedingReminders] 用户有可用授权,准备发送提醒",
					zap.String("openid", collaborator.OpenID))

				// 6. 使用策略模式构造消息数据
				s.logger.Info("📦 [CheckFeedingReminders] STEP 6 - 使用策略模式构造消息数据",
					zap.String("openid", collaborator.OpenID),
				)

				// 使用之前获取的策略构造消息数据
				messageData := strategy.BuildMessageData(lastFeeding, lastFeedingTime, hoursSinceLastFeeding)

				s.logger.Info("📦 [CheckFeedingReminders] 消息数据构造完成",
					zap.String("openid", collaborator.OpenID),
					zap.String("templateType", templateType),
					zap.Any("messageData", messageData),
				)

				// 7. 直接发送订阅消息
				s.logger.Info("📤 [CheckFeedingReminders] STEP 7 - 发送订阅消息",
					zap.String("openid", collaborator.OpenID),
					zap.String("templateType", templateType),
					zap.String("page", "pages/record/feeding/feeding"),
				)

				sendReq := &dto.SendMessageRequest{
					OpenID:     collaborator.OpenID,
					TemplateID: strategy.GetTemplateID(),
					Data:       messageData,
					Page:       "pages/record/feeding/feeding",
				}

				if err := s.subscribeService.SendSubscribeMessage(ctx, sendReq); err != nil {
					s.logger.Error("❌ [CheckFeedingReminders] 发送喂养提醒失败",
						zap.String("babyId", baby.BabyID),
						zap.String("openid", collaborator.OpenID),
						zap.Error(err))
					continue
				}

				s.logger.Info("✅ [CheckFeedingReminders] 喂养提醒发送成功",
					zap.String("babyId", baby.BabyID),
					zap.String("babyName", baby.Name),
					zap.String("openid", collaborator.OpenID),
					zap.Float64("hoursSinceLastFeeding", hoursSinceLastFeeding))
			}

			// 8. 更新提醒标记 (循环结束后统一更新，避免多个协作者时重复更新)
			reminderTime := time.Now().UnixMilli()
			if err := s.feedingRecordRepo.UpdateReminderStatus(ctx, lastFeeding.RecordID, true, reminderTime); err != nil {
				s.logger.Error("❌ [CheckFeedingReminders] 更新提醒标记失败",
					zap.String("recordID", lastFeeding.RecordID),
					zap.Error(err))
			} else {
				s.logger.Info("✅ [CheckFeedingReminders] 提醒标记已更新",
					zap.String("recordID", lastFeeding.RecordID),
					zap.Int64("reminderTime", reminderTime))
			}
		} else {
			s.logger.Info("ℹ️ [CheckFeedingReminders] 距离上次喂养时间未达到提醒阈值,跳过",
				zap.String("babyId", baby.BabyID),
				zap.String("babyName", baby.Name),
				zap.Float64("hoursSinceLastFeeding", hoursSinceLastFeeding),
				zap.Float64("thresholdHours", reminderThreshold),
			)
		}
	}

	s.logger.Info("🏁 [CheckFeedingReminders] ===== END =====",
		zap.Time("endTime", time.Now()),
	)

	return nil
}

// getBabyIDs 获取宝宝ID列表
func getBabyIDs(babies []*entity.Baby) []string {
	ids := make([]string, len(babies))
	for i, baby := range babies {
		ids[i] = baby.BabyID
	}
	return ids
}

// formatDuration 格式化时长为人类可读格式
func formatDuration(hours float64) string {
	h := int(hours)
	if h < 1 {
		return "不到1小时"
	}
	if h == 1 {
		return "1小时"
	}
	return fmt.Sprintf("%d小时", h)
}

// getLastFeedingSide 获取上次喂养位置
func getLastFeedingSide(record *entity.FeedingRecord) string {
	// 从 FeedingDetail 中获取喂养类型和位置
	if side, ok := record.Detail["side"].(string); ok {
		switch side {
		case "left":
			return "左侧"
		case "right":
			return "右侧"
		case "both":
			return "两侧"
		}
	}

	// 如果没有 side 信息，检查是否是奶瓶喂养
	if feedType, ok := record.Detail["type"].(string); ok {
		switch feedType {
		case "bottle":
			return "奶瓶喂养"
		case "food":
			return "辅食"
		}
	}

	return "母乳喂养"
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
