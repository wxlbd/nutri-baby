package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/application/service"
	"github.com/wxlbd/nutri-baby-server/pkg/response"
)

// RecordHandler 记录处理器
type RecordHandler struct {
	recordService *service.RecordService
}

// NewRecordHandler 创建记录处理器
func NewRecordHandler(recordService *service.RecordService) *RecordHandler {
	return &RecordHandler{recordService: recordService}
}

// CreateFeedingRecord 创建喂养记录
// @Router /feeding-records [post]
func (h *RecordHandler) CreateFeedingRecord(c *gin.Context) {
	var req dto.CreateFeedingRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	openID := c.GetString("openid")

	record, err := h.recordService.CreateFeedingRecord(c.Request.Context(), openID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, record)
}

// GetFeedingRecords 获取喂养记录列表
// @Router /feeding-records [get]
func (h *RecordHandler) GetFeedingRecords(c *gin.Context) {
	query := h.parseRecordQuery(c)
	openID := c.GetString("openid")

	records, total, err := h.recordService.GetFeedingRecords(c.Request.Context(), openID, query)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"records": records,
		"total":   total,
		"page":    query.Page,
		"pageSize": query.PageSize,
	})
}

// CreateSleepRecord 创建睡眠记录
// @Router /sleep-records [post]
func (h *RecordHandler) CreateSleepRecord(c *gin.Context) {
	var req dto.CreateSleepRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	openID := c.GetString("openid")

	record, err := h.recordService.CreateSleepRecord(c.Request.Context(), openID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, record)
}

// GetSleepRecords 获取睡眠记录列表
// @Router /sleep-records [get]
func (h *RecordHandler) GetSleepRecords(c *gin.Context) {
	query := h.parseRecordQuery(c)
	openID := c.GetString("openid")

	records, total, err := h.recordService.GetSleepRecords(c.Request.Context(), openID, query)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"records": records,
		"total":   total,
		"page":    query.Page,
		"pageSize": query.PageSize,
	})
}

// CreateDiaperRecord 创建尿布记录
// @Router /diaper-records [post]
func (h *RecordHandler) CreateDiaperRecord(c *gin.Context) {
	var req dto.CreateDiaperRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	openID := c.GetString("openid")

	record, err := h.recordService.CreateDiaperRecord(c.Request.Context(), openID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, record)
}

// GetDiaperRecords 获取尿布记录列表
// @Router /diaper-records [get]
func (h *RecordHandler) GetDiaperRecords(c *gin.Context) {
	query := h.parseRecordQuery(c)
	openID := c.GetString("openid")

	records, total, err := h.recordService.GetDiaperRecords(c.Request.Context(), openID, query)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"records": records,
		"total":   total,
		"page":    query.Page,
		"pageSize": query.PageSize,
	})
}

// CreateGrowthRecord 创建生长记录
// @Router /growth-records [post]
func (h *RecordHandler) CreateGrowthRecord(c *gin.Context) {
	var req dto.CreateGrowthRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	openID := c.GetString("openid")

	record, err := h.recordService.CreateGrowthRecord(c.Request.Context(), openID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, record)
}

// GetGrowthRecords 获取生长记录列表
// @Router /growth-records [get]
func (h *RecordHandler) GetGrowthRecords(c *gin.Context) {
	query := h.parseRecordQuery(c)
	openID := c.GetString("openid")

	records, total, err := h.recordService.GetGrowthRecords(c.Request.Context(), openID, query)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, gin.H{
		"records": records,
		"total":   total,
		"page":    query.Page,
		"pageSize": query.PageSize,
	})
}

// parseRecordQuery 解析记录查询参数
func (h *RecordHandler) parseRecordQuery(c *gin.Context) *dto.RecordListQuery {
	query := &dto.RecordListQuery{
		BabyID:    c.Query("babyId"),
		StartTime: parseInt64(c.Query("startTime")),
		EndTime:   parseInt64(c.Query("endTime")),
		Page:      parseInt(c.Query("page"), 1),
		PageSize:  parseInt(c.Query("pageSize"), 20),
	}

	return query
}

// parseInt 解析整数
func parseInt(s string, defaultValue int) int {
	if s == "" {
		return defaultValue
	}
	val, err := strconv.Atoi(s)
	if err != nil {
		return defaultValue
	}
	return val
}

// parseInt64 解析int64
func parseInt64(s string) int64 {
	if s == "" {
		return 0
	}
	val, err := strconv.ParseInt(s, 10, 64)
	if err != nil {
		return 0
	}
	return val
}
