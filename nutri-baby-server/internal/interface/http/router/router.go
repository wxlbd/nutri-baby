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
	vaccineHandler *handler.VaccineHandler,
	vaccinePlanHandler *handler.VaccinePlanHandler,
	syncHandler *handler.SyncHandler,
) *gin.Engine {
	// 设置Gin运行模式
	gin.SetMode(cfg.Server.Mode)

	r := gin.New()

	// 全局中间件
	r.Use(middleware.CORS())
	r.Use(middleware.Logger())
	r.Use(gin.Recovery())

	// API v1 路由组
	v1 := r.Group("/v1")
	{
		// 认证相关路由（无需认证）
		auth := v1.Group("/auth")
		{
			auth.POST("/wechat-login", authHandler.WechatLogin)
		}

		// 需要认证的路由
		authRequired := v1.Group("")
		authRequired.Use(middleware.Auth(cfg))
		{
			// 认证相关（需要token）
			authRequired.POST("/auth/refresh-token", authHandler.RefreshToken)
			authRequired.GET("/auth/user-info", authHandler.GetUserInfo)

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

				// 疫苗计划管理
				babies.POST("/:babyId/vaccine-plans/initialize", vaccinePlanHandler.InitializePlans)
				babies.GET("/:babyId/vaccine-plans", vaccinePlanHandler.GetPlans)
				babies.POST("/:babyId/vaccine-plans", vaccinePlanHandler.CreatePlan)

				// 疫苗接种记录和提醒
				babies.POST("/:babyId/vaccine-records", vaccineHandler.CreateVaccineRecord)
				babies.GET("/:babyId/vaccine-reminders", vaccineHandler.GetVaccineReminders)
				babies.GET("/:babyId/vaccine-statistics", vaccineHandler.GetVaccineStatistics)
			}

			// 疫苗计划单个操作（不依赖babyId路径参数）
			vaccinePlans := authRequired.Group("/vaccine-plans")
			{
				vaccinePlans.PUT("/:planId", vaccinePlanHandler.UpdatePlan)
				vaccinePlans.DELETE("/:planId", vaccinePlanHandler.DeletePlan)
			}

			// 喂养记录
			feedingRecords := authRequired.Group("/feeding-records")
			{
				feedingRecords.POST("", recordHandler.CreateFeedingRecord)
				feedingRecords.GET("", recordHandler.GetFeedingRecords)
			}

			// 睡眠记录
			sleepRecords := authRequired.Group("/sleep-records")
			{
				sleepRecords.POST("", recordHandler.CreateSleepRecord)
				sleepRecords.GET("", recordHandler.GetSleepRecords)
			}

			// 尿布记录
			diaperRecords := authRequired.Group("/diaper-records")
			{
				diaperRecords.POST("", recordHandler.CreateDiaperRecord)
				diaperRecords.GET("", recordHandler.GetDiaperRecords)
			}

			// 生长记录
			growthRecords := authRequired.Group("/growth-records")
			{
				growthRecords.POST("", recordHandler.CreateGrowthRecord)
				growthRecords.GET("", recordHandler.GetGrowthRecords)
			}

			// WebSocket同步
			authRequired.GET("/sync", syncHandler.HandleSync)
		}
	}

	return r
}
