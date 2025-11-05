package router

import (
	"github.com/gin-gonic/gin"

	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/config"
	"github.com/wxlbd/nutri-baby-server/internal/interface/http/handler"
	"github.com/wxlbd/nutri-baby-server/internal/interface/middleware"
)

// NewRouter 创建并配置路由 (去家庭化架构)
func NewRouter(
	cfg *config.Config,
	authHandler *handler.AuthHandler,
	babyHandler *handler.BabyHandler,
	recordHandler *handler.RecordHandler,
	vaccineScheduleHandler *handler.VaccineScheduleHandler, // 新增
	subscribeHandler *handler.SubscribeHandler,
	syncHandler *handler.SyncHandler,
) *gin.Engine {
	// 设置Gin运行模式
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// 静态文件服务 (用于小程序码、头像等上传文件)
	r.Static("/uploads", "./uploads")

	// API v1 路由组
	v1 := r.Group("/v1")
	{
		// 认证相关路由（无需认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/wechat-login", authHandler.WechatLogin)
		}

		// 邀请相关（公开访问，无需认证）
		// 允许未登录用户通过短码查看邀请详情
		invitations := v1.Group("/invitations")
		{
			invitations.GET("/code/:shortCode", babyHandler.GetInvitationByShortCode)
		}

		// 需要认证的路由
		authRequired := v1.Group("")
		authRequired.Use(middleware.Auth(cfg))
		{
			// 认证相关（需要token）
			authRequired.POST("/auth/refresh-token", authHandler.RefreshToken)
			authRequired.GET("/auth/user-info", authHandler.GetUserInfo)
			authRequired.PUT("/auth/user-info", authHandler.UpdateUserInfo)
			authRequired.PUT("/auth/default-baby", authHandler.SetDefaultBaby)

			// 宝宝管理 (去家庭化架构)
			babies := authRequired.Group("/babies")
			{
				// 宝宝基础CRUD
				babies.POST("", babyHandler.CreateBaby)
				babies.GET("", babyHandler.GetUserBabies) // 获取用户可访问的宝宝列表
				babies.GET("/:babyId", babyHandler.GetBabyDetail)
				babies.PUT("/:babyId", babyHandler.UpdateBaby)
				babies.DELETE("/:babyId", babyHandler.DeleteBaby)

				// 协作者管理
				babies.GET("/:babyId/collaborators", babyHandler.GetCollaborators)
				babies.POST("/:babyId/collaborators/invite", babyHandler.InviteCollaborator)
				babies.POST("/join", babyHandler.JoinBaby) // 通过邀请码加入
				babies.DELETE("/:babyId/collaborators/:openid", babyHandler.RemoveCollaborator)
				babies.PUT("/:babyId/collaborators/:openid/role", babyHandler.UpdateCollaboratorRole)

				// 小程序码生成
				babies.GET("/:babyId/qrcode", babyHandler.GenerateInviteQRCode)

				// 疫苗接种日程(新接口)
				babies.GET("/:babyId/vaccine-schedules", vaccineScheduleHandler.GetVaccineSchedules)
				babies.POST("/:babyId/vaccine-schedules", vaccineScheduleHandler.CreateCustomSchedule)
				babies.PUT("/:babyId/vaccine-schedules/:scheduleId", vaccineScheduleHandler.UpdateVaccineSchedule)
				babies.PUT("/:babyId/vaccine-schedules/:scheduleId/info", vaccineScheduleHandler.UpdateScheduleInfo) // 更新基本信息
				babies.DELETE("/:babyId/vaccine-schedules/:scheduleId", vaccineScheduleHandler.DeleteSchedule)
				babies.GET("/:babyId/vaccine-schedule-statistics", vaccineScheduleHandler.GetStatistics)
				babies.GET("/:babyId/vaccine-reminders", vaccineScheduleHandler.GetReminders)
			}

			// 喂养记录
			feedingRecords := authRequired.Group("/feeding-records")
			{
				feedingRecords.POST("", recordHandler.CreateFeedingRecord)
				feedingRecords.GET("", recordHandler.GetFeedingRecords)
				feedingRecords.GET("/:id", recordHandler.GetFeedingRecordById)
				feedingRecords.PUT("/:id", recordHandler.UpdateFeedingRecord)
				feedingRecords.DELETE("/:id", recordHandler.DeleteFeedingRecord)
			}

			// 睡眠记录
			sleepRecords := authRequired.Group("/sleep-records")
			{
				sleepRecords.POST("", recordHandler.CreateSleepRecord)
				sleepRecords.GET("", recordHandler.GetSleepRecords)
				sleepRecords.GET("/:id", recordHandler.GetSleepRecordById)
				sleepRecords.PUT("/:id", recordHandler.UpdateSleepRecord)
				sleepRecords.DELETE("/:id", recordHandler.DeleteSleepRecord)
			}

			// 尿布记录
			diaperRecords := authRequired.Group("/diaper-records")
			{
				diaperRecords.POST("", recordHandler.CreateDiaperRecord)
				diaperRecords.GET("", recordHandler.GetDiaperRecords)
				diaperRecords.GET("/:id", recordHandler.GetDiaperRecordById)
				diaperRecords.PUT("/:id", recordHandler.UpdateDiaperRecord)
				diaperRecords.DELETE("/:id", recordHandler.DeleteDiaperRecord)
			}

			// 生长记录
			growthRecords := authRequired.Group("/growth-records")
			{
				growthRecords.POST("", recordHandler.CreateGrowthRecord)
				growthRecords.GET("", recordHandler.GetGrowthRecords)
				growthRecords.GET("/:id", recordHandler.GetGrowthRecordById)
				growthRecords.PUT("/:id", recordHandler.UpdateGrowthRecord)
				growthRecords.DELETE("/:id", recordHandler.DeleteGrowthRecord)
			}

			// 时间线聚合接口
			authRequired.GET("record/timeline", recordHandler.GetTimeline)

			// 订阅消息管理
			subscribe := authRequired.Group("/subscribe")
			{
				subscribe.POST("/auth", subscribeHandler.UploadAuth)
				subscribe.GET("/status", subscribeHandler.GetStatus)
				subscribe.DELETE("/cancel", subscribeHandler.Cancel)
				subscribe.GET("/logs", subscribeHandler.GetLogs)
			}

			// WebSocket同步
			authRequired.GET("/sync", syncHandler.HandleSync)
		}
	}

	return r
}
