package handler

import (
	"fmt"
	"strconv"
	"time"

	"github.com/gin-gonic/gin"
	"go.uber.org/zap"

	"github.com/wxlbd/nutri-baby-server/internal/application/service"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
	"github.com/wxlbd/nutri-baby-server/pkg/response"
)

// AIAnalysisHandler AI分析处理器
type AIAnalysisHandler struct {
	aiAnalysisService service.AIAnalysisService
	logger            *zap.Logger
}

// NewAIAnalysisHandler 创建AI分析处理器
func NewAIAnalysisHandler(aiAnalysisService service.AIAnalysisService, logger *zap.Logger) *AIAnalysisHandler {
	return &AIAnalysisHandler{
		aiAnalysisService: aiAnalysisService,
		logger:            logger,
	}
}

// CreateAnalysis 创建AI分析任务
// @Summary 创建AI分析任务
// @Description 创建指定类型的AI分析任务
// @Tags AI分析
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param request body service.CreateAnalysisRequest true "分析请求"
// @Success 200 {object} response.Response{data=service.AnalysisResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/analysis [post]
func (h *AIAnalysisHandler) CreateAnalysis(c *gin.Context) {
	var req service.CreateAnalysisRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	// 验证权限
	if err := h.checkPermission(c, req.BabyID); err != nil {
		response.Error(c, err)
		return
	}

	result, err := h.aiAnalysisService.CreateAnalysis(c.Request.Context(), &req)
	if err != nil {
		h.logger.Error("创建AI分析任务失败",
			zap.Error(err),
			zap.Int64("baby_id", req.BabyID),
			zap.String("analysis_type", string(req.AnalysisType)),
		)
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetAnalysisResult 获取AI分析结果
// @Summary 获取AI分析结果
// @Description 获取指定分析ID的分析结果
// @Tags AI分析
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param id path int true "分析ID"
// @Success 200 {object} response.Response{data=service.AnalysisResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/analysis/{id} [get]
func (h *AIAnalysisHandler) GetAnalysisResult(c *gin.Context) {
	analysisID, err := strconv.ParseInt(c.Param("id"), 10, 64)
	if err != nil {
		response.ErrorWithMessage(c, 1001, "无效的分析ID")
		return
	}

	result, err := h.aiAnalysisService.GetAnalysisResult(c.Request.Context(), strconv.FormatInt(analysisID, 10))
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			response.Error(c, errors.ErrNotFound)
			return
		}
		h.logger.Error("获取AI分析结果失败",
			zap.Error(err),
			zap.Int64("analysis_id", analysisID),
		)
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetLatestAnalysis 获取最新AI分析结果
// @Summary 获取最新AI分析结果
// @Description 获取指定宝宝和类型的最新AI分析结果
// @Tags AI分析
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param baby_id query int true "宝宝ID"
// @Param analysis_type query string true "分析类型" Enums(feeding,sleep,growth,health,behavior)
// @Success 200 {object} response.Response{data=service.AnalysisResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 404 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/analysis/latest [get]
func (h *AIAnalysisHandler) GetLatestAnalysis(c *gin.Context) {
	babyID, err := strconv.ParseInt(c.Query("baby_id"), 10, 64)
	if err != nil {
		response.ErrorWithMessage(c, 1001, "无效的宝宝ID")
		return
	}

	analysisType := entity.AIAnalysisType(c.Query("analysis_type"))
	if analysisType == "" {
		response.ErrorWithMessage(c, 1001, "分析类型不能为空")
		return
	}

	// 验证权限
	if err := h.checkPermission(c, babyID); err != nil {
		response.Error(c, errors.ErrPermissionDenied)
		return
	}

	result, err := h.aiAnalysisService.GetLatestAnalysis(c.Request.Context(), strconv.FormatInt(babyID, 10), analysisType)
	if err != nil {
		if errors.Is(err, errors.ErrNotFound) {
			response.Error(c, errors.ErrNotFound)
			return
		}
		h.logger.Error("获取最新AI分析结果失败",
			zap.Error(err),
			zap.Int64("baby_id", babyID),
			zap.String("analysis_type", string(analysisType)),
		)
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// BatchAnalyze 批量分析
// @Summary 批量AI分析
// @Description 对指定时间范围内的宝宝数据进行批量AI分析
// @Tags AI分析
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param baby_id query int true "宝宝ID"
// @Param start_date query string true "开始日期 (YYYY-MM-DD)"
// @Param end_date query string true "结束日期 (YYYY-MM-DD)"
// @Success 200 {object} response.Response{data=service.BatchAnalysisResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/analysis/batch [post]
func (h *AIAnalysisHandler) BatchAnalyze(c *gin.Context) {
	babyID, err := strconv.ParseInt(c.Query("baby_id"), 10, 64)
	if err != nil {
		response.ErrorWithMessage(c, 1001, "无效的宝宝ID")
		return
	}

	startDate, err := time.Parse("2006-01-02", c.Query("start_date"))
	if err != nil {
		response.ErrorWithMessage(c, 1001, "无效的开始日期格式")
		return
	}

	endDate, err := time.Parse("2006-01-02", c.Query("end_date"))
	if err != nil {
		response.ErrorWithMessage(c, 1001, "无效的结束日期格式")
		return
	}

	// 验证权限
	if err := h.checkPermission(c, babyID); err != nil {
		response.Error(c, errors.ErrPermissionDenied)
		return
	}

	result, err := h.aiAnalysisService.BatchAnalyze(c.Request.Context(), strconv.FormatInt(babyID, 10), startDate, endDate)
	if err != nil {
		h.logger.Error("批量AI分析失败",
			zap.Error(err),
			zap.Int64("baby_id", babyID),
			zap.Time("start_date", startDate),
			zap.Time("end_date", endDate),
		)
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GenerateDailyTips 生成每日建议
// @Summary 生成每日建议
// @Description 为指定宝宝生成当日的AI育儿建议
// @Tags AI分析
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param baby_id query int true "宝宝ID"
// @Param date query string false "日期 (YYYY-MM-DD)，默认为今天"
// @Success 200 {object} response.Response{data=service.DailyTipsResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/daily-tips [post]
func (h *AIAnalysisHandler) GenerateDailyTips(c *gin.Context) {
	babyID, err := strconv.ParseInt(c.Query("baby_id"), 10, 64)
	if err != nil {
		response.ErrorWithMessage(c, 1001, "无效的宝宝ID")
		return
	}

	dateStr := c.Query("date")
	var date time.Time
	if dateStr != "" {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			response.ErrorWithMessage(c, 1001, "无效的日期格式")
			return
		}
	} else {
		date = time.Now()
	}

	// 验证权限
	if err := h.checkPermission(c, babyID); err != nil {
		response.Error(c, errors.ErrPermissionDenied)
		return
	}

	result, err := h.aiAnalysisService.GenerateDailyTips(c.Request.Context(), strconv.FormatInt(babyID, 10), date)
	if err != nil {
		h.logger.Error("生成每日建议失败",
			zap.Error(err),
			zap.Int64("baby_id", babyID),
			zap.Time("date", date),
		)
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetDailyTips 获取每日建议
// @Summary 获取每日建议
// @Description 获取指定宝宝和日期的AI育儿建议
// @Tags AI分析
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param baby_id query int true "宝宝ID"
// @Param date query string false "日期 (YYYY-MM-DD)，默认为今天"
// @Success 200 {object} response.Response{data=service.DailyTipsResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/daily-tips [get]
func (h *AIAnalysisHandler) GetDailyTips(c *gin.Context) {
	babyID, err := strconv.ParseInt(c.Query("baby_id"), 10, 64)
	if err != nil {
		response.ErrorWithMessage(c, 1001, "无效的宝宝ID")
		return
	}

	dateStr := c.Query("date")
	var date time.Time
	if dateStr != "" {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			response.ErrorWithMessage(c, 1001, "无效的日期格式")
			return
		}
	} else {
		date = time.Now()
	}

	// 验证权限
	if err := h.checkPermission(c, babyID); err != nil {
		response.Error(c, errors.ErrPermissionDenied)
		return
	}

	result, err := h.aiAnalysisService.GetDailyTips(c.Request.Context(), strconv.FormatInt(babyID, 10), date)
	if err != nil {
		h.logger.Error("获取每日建议失败",
			zap.Error(err),
			zap.Int64("baby_id", babyID),
			zap.Time("date", date),
		)
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetAnalysisStats 获取AI分析统计
// @Summary 获取AI分析统计
// @Description 获取指定宝宝的AI分析历史统计信息
// @Tags AI分析
// @Accept json
// @Produce json
// @Param Authorization header string true "Bearer Token"
// @Param baby_id query int true "宝宝ID"
// @Success 200 {object} response.Response{data=service.AnalysisStatsResponse}
// @Failure 400 {object} response.Response
// @Failure 401 {object} response.Response
// @Failure 500 {object} response.Response
// @Router /api/ai/analysis/stats [get]
func (h *AIAnalysisHandler) GetAnalysisStats(c *gin.Context) {
	babyID, err := strconv.ParseInt(c.Query("baby_id"), 10, 64)
	if err != nil {
		response.ErrorWithMessage(c, 1001, "无效的宝宝ID")
		return
	}

	// 验证权限
	if err := h.checkPermission(c, babyID); err != nil {
		response.Error(c, errors.ErrPermissionDenied)
		return
	}

	result, err := h.aiAnalysisService.GetAnalysisStats(c.Request.Context(), strconv.FormatInt(babyID, 10))
	if err != nil {
		h.logger.Error("获取AI分析统计失败",
			zap.Error(err),
			zap.Int64("baby_id", babyID),
		)
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// checkPermission 检查权限
func (h *AIAnalysisHandler) checkPermission(c *gin.Context, babyID int64) error {
	// 从上下文中获取当前用户的openid (由auth中间件设置)
	openid, exists := c.Get("openid")
	if !exists {
		return errors.ErrUnauthorized
	}

	openidStr, ok := openid.(string)
	if !ok || openidStr == "" {
		return errors.ErrUnauthorized
	}

	// 检查用户是否有权限访问该宝宝的数据
	// 这里需要调用权限检查服务，暂时直接返回nil
	// TODO: 实现完整的权限检查逻辑
	_ = openidStr
	_ = babyID
	return nil
}

// RegisterAIAnalysisRoutes 注册AI分析路由
func RegisterAIAnalysisRoutes(router *gin.RouterGroup, handler *AIAnalysisHandler) {
	aiGroup := router.Group("/ai")
	{
		aiGroup.POST("/analysis", handler.CreateAnalysis)
		aiGroup.GET("/analysis/:id", handler.GetAnalysisResult)
		aiGroup.GET("/analysis/latest", handler.GetLatestAnalysis)
		aiGroup.POST("/analysis/batch", handler.BatchAnalyze)
		aiGroup.GET("/analysis/stats", handler.GetAnalysisStats)

		aiGroup.POST("/daily-tips", handler.GenerateDailyTips)
		aiGroup.GET("/daily-tips", handler.GetDailyTips)
	}
}

// RegisterAIAnalysisHandlers 注册AI分析处理器到Wire
func RegisterAIAnalysisHandlers(router *gin.RouterGroup, aiAnalysisService service.AIAnalysisService, logger *zap.Logger) {
	handler := NewAIAnalysisHandler(aiAnalysisService, logger)
	RegisterAIAnalysisRoutes(router, handler)
}

// ProcessPendingAnalysesJob 处理待分析的AI任务（后台任务）
func ProcessPendingAnalysesJob(aiAnalysisService service.AIAnalysisService, logger *zap.Logger) gin.HandlerFunc {
	return func(c *gin.Context) {
		if err := aiAnalysisService.ProcessPendingAnalyses(c.Request.Context()); err != nil {
			logger.Error("处理待分析AI任务失败", zap.Error(err))
			response.Error(c, err)
			return
		}
		response.Success(c, nil)
	}
}

// RegisterBackgroundJobs 注册后台任务路由
func RegisterBackgroundJobs(router *gin.RouterGroup, aiAnalysisService service.AIAnalysisService, logger *zap.Logger) {
	router.POST("/jobs/process-pending-analyses", ProcessPendingAnalysesJob(aiAnalysisService, logger))
}

// ConvertToAnalysisResponse 转换分析响应格式
func ConvertToAnalysisResponse(analysis interface{}) *service.AnalysisResponse {
	// 实现转换逻辑
	return &service.AnalysisResponse{}
}

// ConvertToBatchAnalysisResponse 转换批量分析响应格式
func ConvertToBatchAnalysisResponse(analyses interface{}) *service.BatchAnalysisResponse {
	// 实现转换逻辑
	return &service.BatchAnalysisResponse{}
}

// ConvertToDailyTipsResponse 转换每日建议响应格式
func ConvertToDailyTipsResponse(tips interface{}) *service.DailyTipsResponse {
	// 实现转换逻辑
	return &service.DailyTipsResponse{}
}

// ConvertToAnalysisStatsResponse 转换分析统计响应格式
func ConvertToAnalysisStatsResponse(stats interface{}) *service.AnalysisStatsResponse {
	// 实现转换逻辑
	return &service.AnalysisStatsResponse{}
}

// ValidateAnalysisRequest 验证分析请求
func ValidateAnalysisRequest(req interface{}) error {
	// 实现验证逻辑
	return nil
}

// ValidateAnalysisType 验证分析类型
func ValidateAnalysisType(analysisType string) bool {
	validTypes := []string{
		string(entity.AIAnalysisTypeFeeding),
		string(entity.AIAnalysisTypeSleep),
		string(entity.AIAnalysisTypeGrowth),
		string(entity.AIAnalysisTypeHealth),
		string(entity.AIAnalysisTypeBehavior),
	}

	for _, validType := range validTypes {
		if analysisType == validType {
			return true
		}
	}
	return false
}

// ParseAnalysisType 解析分析类型
func ParseAnalysisType(analysisType string) (entity.AIAnalysisType, error) {
	if !ValidateAnalysisType(analysisType) {
		return "", errors.New(errors.ParamError, "无效的分析类型")
	}
	return entity.AIAnalysisType(analysisType), nil
}

// GenerateAnalysisID 生成分析ID
func GenerateAnalysisID() int64 {
	// 使用雪花ID生成器
	return time.Now().UnixNano()
}

// FormatAnalysisDate 格式化分析日期
func FormatAnalysisDate(date time.Time) string {
	return date.Format("2006-01-02")
}

// ParseAnalysisDate 解析分析日期
func ParseAnalysisDate(dateStr string) (time.Time, error) {
	return time.Parse("2006-01-02", dateStr)
}

// ValidateDateRange 验证日期范围
func ValidateDateRange(startDate, endDate time.Time) error {
	if endDate.Before(startDate) {
		return errors.New(errors.ParamError, "结束日期不能早于开始日期")
	}
	if startDate.After(time.Now()) {
		return errors.New(errors.ParamError, "开始日期不能晚于今天")
	}
	if endDate.After(time.Now()) {
		return errors.New(errors.ParamError, "结束日期不能晚于今天")
	}
	return nil
}

// CalculateAnalysisDuration 计算分析时长
func CalculateAnalysisDuration(startDate, endDate time.Time) int {
	return int(endDate.Sub(startDate).Hours() / 24)
}

// GetDefaultAnalysisDateRange 获取默认分析日期范围
func GetDefaultAnalysisDateRange() (time.Time, time.Time) {
	endDate := time.Now()
	startDate := endDate.AddDate(0, 0, -7) // 默认分析最近7天
	return startDate, endDate
}

// GetAnalysisCacheKey 获取分析缓存键
func GetAnalysisCacheKey(babyID int64, analysisType entity.AIAnalysisType, startDate, endDate time.Time) string {
	return fmt.Sprintf("ai_analysis:%d:%s:%s:%s", babyID, analysisType, startDate.Format("2006-01-02"), endDate.Format("2006-01-02"))
}

// GetDailyTipsCacheKey 获取每日建议缓存键
func GetDailyTipsCacheKey(babyID int64, date time.Time) string {
	return fmt.Sprintf("daily_tips:%d:%s", babyID, date.Format("2006-01-02"))
}

// IsAnalysisExpired 判断分析是否过期
func IsAnalysisExpired(createdAt time.Time, expiryHours int) bool {
	return time.Since(createdAt) > time.Duration(expiryHours)*time.Hour
}

// GetAnalysisExpiryHours 获取分析过期时间（小时）
func GetAnalysisExpiryHours() int {
	return 24 // 分析结果24小时后过期
}

// GetDailyTipsExpiryHours 获取每日建议过期时间（小时）
func GetDailyTipsExpiryHours() int {
	return 24 // 每日建议24小时后过期
}