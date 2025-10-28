package handler

import (
	"strconv"

	"github.com/gin-gonic/gin"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/application/service"
	"github.com/wxlbd/nutri-baby-server/pkg/response"
)

// VaccineHandler 疫苗处理器
type VaccineHandler struct {
	vaccineService *service.VaccineService
}

// NewVaccineHandler 创建疫苗处理器
func NewVaccineHandler(vaccineService *service.VaccineService) *VaccineHandler {
	return &VaccineHandler{vaccineService: vaccineService}
}

// GetVaccinePlans 获取疫苗计划
// @Router /babies/:babyId/vaccine-plans [get]
func (h *VaccineHandler) GetVaccinePlans(c *gin.Context) {
	babyID := c.Param("babyId")

	plans, err := h.vaccineService.GetVaccinePlans(c.Request.Context(), babyID)
	if err != nil {
		response.Error(c, err)
		return
	}

	// 统计完成度
	total := len(plans)
	completed := 0
	for _, p := range plans {
		if p.Status == "completed" {
			completed++
		}
	}

	percentage := 0
	if total > 0 {
		percentage = completed * 100 / total
	}

	response.Success(c, gin.H{
		"plans":      plans,
		"total":      total,
		"completed":  completed,
		"percentage": percentage,
	})
}

// CreateVaccineRecord 创建疫苗接种记录
// @Router /babies/:babyId/vaccine-records [post]
func (h *VaccineHandler) CreateVaccineRecord(c *gin.Context) {
	babyID := c.Param("babyId")

	var req dto.CreateVaccineRecordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	// 从context获取用户信息
	createBy := c.GetString("openid")

	record, err := h.vaccineService.CreateVaccineRecord(c.Request.Context(), babyID, createBy, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, record)
}

// GetVaccineReminders 获取疫苗提醒列表
// @Router /babies/:babyId/vaccine-reminders [get]
func (h *VaccineHandler) GetVaccineReminders(c *gin.Context) {
	babyID := c.Param("babyId")
	status := c.Query("status")
	limit := 10
	if l := c.Query("limit"); l != "" {
		if n, err := c.GetQuery("limit"); err {
			_ = n
		}
	}

	reminders, err := h.vaccineService.GetVaccineReminders(c.Request.Context(), babyID, status, limit)
	if err != nil {
		response.Error(c, err)
		return
	}

	// 统计各状态数量
	statusCount := map[string]int{
		"upcoming": 0,
		"due":      0,
		"overdue":  0,
	}

	for _, r := range reminders {
		statusCount[r.Status]++
	}

	response.Success(c, gin.H{
		"reminders": reminders,
		"total":     len(reminders),
		"upcoming":  statusCount["upcoming"],
		"due":       statusCount["due"],
		"overdue":   statusCount["overdue"],
	})
}

// GetVaccineStatistics 获取疫苗接种统计
// @Router /babies/:babyId/vaccine-statistics [get]
func (h *VaccineHandler) GetVaccineStatistics(c *gin.Context) {
	babyID := c.Param("babyId")

	stats, err := h.vaccineService.GetVaccineStatistics(c.Request.Context(), babyID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, stats)
}

// GetVaccineRecords 获取疫苗接种记录
// @Router /babies/:babyId/vaccine-records [get]
func (h *VaccineHandler) GetVaccineRecords(c *gin.Context) {
	babyID := c.Param("babyId")
	page, err := strconv.Atoi(c.DefaultQuery("page", "1"))
	if err != nil {
		response.Error(c, err)
		return
	}
	pageSize, err := strconv.Atoi(c.DefaultQuery("pageSize", "10"))
	if err != nil {
		response.Error(c, err)
		return
	}
	records, err := h.vaccineService.GetVaccineRecords(c.Request.Context(), babyID, page, pageSize)
	if err != nil {
		response.Error(c, err)
		return
	}
	response.Success(c, records)
}
