//go:build wireinject
// +build wireinject

package wire

import (
	"github.com/google/wire"
	"github.com/wxlbd/nutri-baby-server/internal/application/service"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/config"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/logger"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/persistence"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/wechat"
	"github.com/wxlbd/nutri-baby-server/internal/interface/http/handler"
	"github.com/wxlbd/nutri-baby-server/internal/interface/http/router"
)

// InitApp 初始化应用(Wire自动生成) (去家庭化架构)
func InitApp(cfg *config.Config) (*App, error) {
	wire.Build(
		// 基础设施层
		logger.NewLogger, // 日志系统
		persistence.NewDatabase,
		persistence.NewRedis, // Redis 客户端
		wechat.NewClient,     // 微信 SDK 客户端

		// 仓储层
		persistence.NewSubscriptionCacheRepository, // 订阅权限缓存管理器
		persistence.NewUserRepository,
		// persistence.NewFamilyRepository, // 已废弃：去家庭化架构
		// persistence.NewFamilyMemberRepository, // 已废弃：去家庭化架构
		// persistence.NewInvitationRepository, // 已废弃：去家庭化架构
		persistence.NewBabyRepository,
		persistence.NewBabyCollaboratorRepository, // 去家庭化架构：宝宝协作者仓储
		persistence.NewBabyInvitationRepository,   // 去家庭化架构：宝宝邀请仓储
		persistence.NewFeedingRecordRepository,
		persistence.NewSleepRecordRepository,
		persistence.NewDiaperRecordRepository,
		persistence.NewGrowthRecordRepository,
		persistence.NewVaccineRecordRepository,
		persistence.NewBabyVaccinePlanRepository, // 宝宝疫苗计划仓储
		persistence.NewVaccineReminderRepository, // 疫苗提醒仓储
		persistence.NewSubscribeRepository,       // 订阅消息仓储

		// 应用服务层
		service.NewWechatService,    // 微信服务
		service.NewSubscribeService, // 订阅消息服务
		service.NewAuthService,
		// service.NewFamilyService, // 已废弃：去家庭化架构
		service.NewBabyService,
		service.NewFeedingRecordService, // 喂养记录服务
		service.NewSleepRecordService,   // 睡眠记录服务
		service.NewDiaperRecordService,  // 尿布记录服务
		service.NewGrowthRecordService,  // 成长记录服务
		service.NewVaccineService,
		service.NewVaccinePlanService, // 疫苗计划管理服务
		service.NewSchedulerService,   // 定时任务服务
		// service.NewSyncService, // TODO: WebSocket同步未实现，暂时注释

		// HTTP处理器
		handler.NewAuthHandler,
		// handler.NewFamilyHandler, // 已废弃：去家庭化架构
		handler.NewBabyHandler,
		handler.NewRecordHandler,
		handler.NewVaccineHandler,
		handler.NewVaccinePlanHandler, // 疫苗计划管理处理器
		handler.NewSubscribeHandler,   // 订阅消息处理器
		handler.NewSyncHandler,

		// 路由
		router.NewRouter,

		// 应用
		NewApp,
	)
	return &App{}, nil
}
