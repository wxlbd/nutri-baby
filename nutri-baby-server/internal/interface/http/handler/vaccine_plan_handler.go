package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/application/service"
	errs "github.com/wxlbd/nutri-baby-server/pkg/errors"
	"github.com/wxlbd/nutri-baby-server/pkg/response"
)

type VaccinePlanHandler struct {
	vaccinePlanService *service.VaccinePlanService
}

// NewVaccinePlanHandler 创建疫苗计划处理器
func NewVaccinePlanHandler(vaccinePlanService *service.VaccinePlanService) *VaccinePlanHandler {
	return &VaccinePlanHandler{
		vaccinePlanService: vaccinePlanService,
	}
}

// InitializePlans 初始化宝宝的疫苗计划
// POST /api/v1/babies/:babyId/vaccine-plans/initialize
func (h *VaccinePlanHandler) InitializePlans(c *gin.Context) {
	babyID := c.Param("babyId")
	if babyID == "" {
		response.Error(c, errs.ErrInvalidParam.WithMessage("babyId不能为空"))
		return
	}

	// 获取用户openID
	openID, exists := c.Get("openid")
	if !exists {
		response.Error(c, errs.ErrUnauthorized)
		return
	}

	var req dto.InitializeVaccinePlansRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		// 允许空请求体
		req.Force = false
	}

	result, err := h.vaccinePlanService.InitializePlansForBaby(c.Request.Context(), babyID, openID.(string), req.Force)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// GetPlans 获取宝宝的疫苗计划列表
// GET /api/v1/babies/:babyId/vaccine-plans
func (h *VaccinePlanHandler) GetPlans(c *gin.Context) {
	babyID := c.Param("babyId")
	if babyID == "" {
		response.Error(c, errs.ErrInvalidParam.WithMessage("babyId不能为空"))
		return
	}

	result, err := h.vaccinePlanService.GetPlansForBaby(c.Request.Context(), babyID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// CreatePlan 创建自定义疫苗计划
// POST /api/v1/babies/:babyId/vaccine-plans
func (h *VaccinePlanHandler) CreatePlan(c *gin.Context) {
	babyID := c.Param("babyId")
	if babyID == "" {
		response.Error(c, errs.ErrInvalidParam.WithMessage("babyId不能为空"))
		return
	}

	// 获取用户openID
	openID, exists := c.Get("openid")
	if !exists {
		response.Error(c, errs.ErrUnauthorized)
		return
	}

	var req dto.CreateBabyVaccinePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrInvalidParam.WithMessage(err.Error()))
		return
	}

	result, err := h.vaccinePlanService.CreatePlan(c.Request.Context(), babyID, openID.(string), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// UpdatePlan 更新疫苗计划
// PUT /api/v1/vaccine-plans/:planId
func (h *VaccinePlanHandler) UpdatePlan(c *gin.Context) {
	planID := c.Param("planId")
	if planID == "" {
		response.Error(c, errs.ErrInvalidParam.WithMessage("planId不能为空"))
		return
	}

	var req dto.UpdateBabyVaccinePlanRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, errs.ErrInvalidParam.WithMessage(err.Error()))
		return
	}

	result, err := h.vaccinePlanService.UpdatePlan(c.Request.Context(), planID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, result)
}

// DeletePlan 删除疫苗计划
// DELETE /api/v1/vaccine-plans/:planId
func (h *VaccinePlanHandler) DeletePlan(c *gin.Context) {
	planID := c.Param("planId")
	if planID == "" {
		response.Error(c, errs.ErrInvalidParam.WithMessage("planId不能为空"))
		return
	}

	err := h.vaccinePlanService.DeletePlan(c.Request.Context(), planID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, map[string]interface{}{
		"message": "删除成功",
	})
}
