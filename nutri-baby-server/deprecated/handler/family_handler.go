package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/application/service"
	"github.com/wxlbd/nutri-baby-server/pkg/response"
)

// FamilyHandler 家庭处理器
type FamilyHandler struct {
	familyService *service.FamilyService
}

// NewFamilyHandler 创建家庭处理器
func NewFamilyHandler(familyService *service.FamilyService) *FamilyHandler {
	return &FamilyHandler{familyService: familyService}
}

// CreateFamily 创建家庭
// @Router /families [post]
func (h *FamilyHandler) CreateFamily(c *gin.Context) {
	var req dto.CreateFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	openID := c.GetString("openid")

	family, err := h.familyService.CreateFamily(c.Request.Context(), openID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, family)
}

// GetFamilyList 获取家庭列表
// @Router /families [get]
func (h *FamilyHandler) GetFamilyList(c *gin.Context) {
	openID := c.GetString("openid")

	families, err := h.familyService.GetFamilyList(c.Request.Context(), openID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, families)
}

// GetFamilyDetail 获取家庭详情
// @Router /families/:familyId [get]
func (h *FamilyHandler) GetFamilyDetail(c *gin.Context) {
	familyID := c.Param("familyId")
	openID := c.GetString("openid")

	family, err := h.familyService.GetFamilyDetail(c.Request.Context(), familyID, openID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, family)
}

// UpdateFamily 更新家庭
// @Router /families/:familyId [put]
func (h *FamilyHandler) UpdateFamily(c *gin.Context) {
	familyID := c.Param("familyId")
	openID := c.GetString("openid")

	var req dto.UpdateFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	if err := h.familyService.UpdateFamily(c.Request.Context(), familyID, openID, &req); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// DeleteFamily 删除家庭
// @Router /families/:familyId [delete]
func (h *FamilyHandler) DeleteFamily(c *gin.Context) {
	familyID := c.Param("familyId")
	openID := c.GetString("openid")

	if err := h.familyService.DeleteFamily(c.Request.Context(), familyID, openID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// CreateInvitation 创建邀请
// @Router /families/invitations [post]
func (h *FamilyHandler) CreateInvitation(c *gin.Context) {
	var req dto.CreateInvitationRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	openID := c.GetString("openid")

	invitation, err := h.familyService.CreateInvitation(c.Request.Context(), openID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, invitation)
}

// JoinFamily 加入家庭
// @Router /families/join [post]
func (h *FamilyHandler) JoinFamily(c *gin.Context) {
	var req dto.JoinFamilyRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	openID := c.GetString("openid")

	family, err := h.familyService.JoinFamily(c.Request.Context(), openID, &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, family)
}

// RemoveMember 移除成员
// @Router /families/:familyId/members/:memberId [delete]
func (h *FamilyHandler) RemoveMember(c *gin.Context) {
	familyID := c.Param("familyId")
	memberID := c.Param("memberId")
	openID := c.GetString("openid")

	if err := h.familyService.RemoveMember(c.Request.Context(), familyID, memberID, openID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}

// LeaveFamily 离开家庭
// @Router /families/:familyId/leave [post]
func (h *FamilyHandler) LeaveFamily(c *gin.Context) {
	familyID := c.Param("familyId")
	openID := c.GetString("openid")

	if err := h.familyService.LeaveFamily(c.Request.Context(), familyID, openID); err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, nil)
}
