package handler

import (
	"github.com/gin-gonic/gin"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/application/service"
	"github.com/wxlbd/nutri-baby-server/pkg/response"
)

// AuthHandler 认证处理器
type AuthHandler struct {
	authService *service.AuthService
}

// NewAuthHandler 创建认证处理器
func NewAuthHandler(authService *service.AuthService) *AuthHandler {
	return &AuthHandler{authService: authService}
}

// WechatLogin 微信小程序登录
// @Router /auth/wechat-login [post]
func (h *AuthHandler) WechatLogin(c *gin.Context) {
	var req dto.WechatLoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.ErrorWithMessage(c, 1001, "参数错误: "+err.Error())
		return
	}

	resp, err := h.authService.WechatLogin(c.Request.Context(), &req)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// RefreshToken 刷新Token
// @Router /auth/refresh-token [post]
func (h *AuthHandler) RefreshToken(c *gin.Context) {
	// 从context获取当前用户openid (由Auth中间件设置)
	openID := c.GetString("openid")

	resp, err := h.authService.RefreshToken(c.Request.Context(), openID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, resp)
}

// GetUserInfo 获取用户信息
// @Router /auth/user-info [get]
func (h *AuthHandler) GetUserInfo(c *gin.Context) {
	// 从context获取当前用户openid
	openID := c.GetString("openid")

	userInfo, err := h.authService.GetUserInfo(c.Request.Context(), openID)
	if err != nil {
		response.Error(c, err)
		return
	}

	response.Success(c, userInfo)
}
