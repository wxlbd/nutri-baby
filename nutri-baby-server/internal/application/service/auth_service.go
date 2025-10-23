package service

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"

	"github.com/golang-jwt/jwt/v5"

	"github.com/wxlbd/nutri-baby-server/internal/application/dto"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
	"github.com/wxlbd/nutri-baby-server/internal/domain/repository"
	"github.com/wxlbd/nutri-baby-server/internal/infrastructure/config"
	"github.com/wxlbd/nutri-baby-server/pkg/errors"
)

// AuthService 认证服务 (去家庭化架构)
type AuthService struct {
	userRepo repository.UserRepository
	cfg      *config.Config
}

// NewAuthService 创建认证服务
func NewAuthService(
	userRepo repository.UserRepository,
	cfg *config.Config,
) *AuthService {
	return &AuthService{
		userRepo: userRepo,
		cfg:      cfg,
	}
}

// WechatSession 微信登录会话
type WechatSession struct {
	OpenID     string `json:"openid"`
	SessionKey string `json:"session_key"`
	UnionID    string `json:"unionid"`
	ErrCode    int    `json:"errcode"`
	ErrMsg     string `json:"errmsg"`
}

// WechatLogin 微信小程序登录 (去家庭化架构)
func (s *AuthService) WechatLogin(ctx context.Context, req *dto.WechatLoginRequest) (*dto.LoginResponse, error) {
	// 调用微信API获取openid
	session, err := s.getWechatSession(req.Code)
	if err != nil {
		return nil, err
	}

	if session.ErrCode != 0 {
		return nil, errors.New(errors.Unauthorized, fmt.Sprintf("微信登录失败: %s", session.ErrMsg))
	}

	// 查找或创建用户
	user, err := s.userRepo.FindByOpenID(ctx, session.OpenID)
	if err != nil && !errors.Is(err, errors.ErrUserNotFound) {
		return nil, err
	}

	now := time.Now().UnixMilli()
	isNewUser := false

	if user == nil {
		// 创建新用户
		user = &entity.User{
			OpenID:        session.OpenID,
			NickName:      req.NickName,
			AvatarURL:     req.AvatarURL,
			CreateTime:    now,
			LastLoginTime: now,
			UpdateTime:    now,
		}

		if err := s.userRepo.Create(ctx, user); err != nil {
			return nil, err
		}

		isNewUser = true
	} else {
		// 更新用户信息
		user.NickName = req.NickName
		user.AvatarURL = req.AvatarURL
		user.LastLoginTime = now

		if err := s.userRepo.Update(ctx, user); err != nil {
			return nil, err
		}
	}

	// 生成Token
	token, err := s.generateToken(user.OpenID)
	if err != nil {
		return nil, err
	}

	return &dto.LoginResponse{
		Token: token,
		UserInfo: dto.UserInfoDTO{
			OpenID:        user.OpenID,
			NickName:      user.NickName,
			AvatarURL:     user.AvatarURL,
			DefaultBabyID: user.DefaultBabyID,
		},
		IsNewUser: isNewUser, // 前端根据此字段判断是否需要引导创建宝宝
	}, nil
}

// RefreshToken 刷新Token
func (s *AuthService) RefreshToken(ctx context.Context, openID string) (*dto.RefreshTokenResponse, error) {
	// 验证用户存在
	user, err := s.userRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return nil, err
	}

	// 生成新Token
	token, err := s.generateToken(user.OpenID)
	if err != nil {
		return nil, err
	}

	return &dto.RefreshTokenResponse{
		Token: token,
	}, nil
}

// GetUserInfo 获取用户信息
func (s *AuthService) GetUserInfo(ctx context.Context, openID string) (*dto.UserInfoDTO, error) {
	user, err := s.userRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return nil, err
	}

	return &dto.UserInfoDTO{
		OpenID:        user.OpenID,
		NickName:      user.NickName,
		AvatarURL:     user.AvatarURL,
		DefaultBabyID: user.DefaultBabyID,
		CreateTime:    user.CreateTime,
		LastLoginTime: user.LastLoginTime,
	}, nil
}

// SetDefaultBaby 设置默认宝宝
func (s *AuthService) SetDefaultBaby(ctx context.Context, openID string, req *dto.SetDefaultBabyRequest) error {
	// 验证用户存在
	_, err := s.userRepo.FindByOpenID(ctx, openID)
	if err != nil {
		return err
	}

	// 更新默认宝宝ID
	if err := s.userRepo.UpdateDefaultBabyID(ctx, openID, req.BabyID); err != nil {
		return err
	}

	return nil
}

// getWechatSession 获取微信会话
func (s *AuthService) getWechatSession(code string) (*WechatSession, error) {
	url := fmt.Sprintf(
		"https://api.weixin.qq.com/sns/jscode2session?appid=%s&secret=%s&js_code=%s&grant_type=authorization_code",
		s.cfg.Wechat.AppID,
		s.cfg.Wechat.AppSecret,
		code,
	)

	resp, err := http.Get(url)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "调用微信API失败", err)
	}
	defer resp.Body.Close()

	body, err := io.ReadAll(resp.Body)
	if err != nil {
		return nil, errors.Wrap(errors.InternalError, "读取微信API响应失败", err)
	}

	var session WechatSession
	if err := json.Unmarshal(body, &session); err != nil {
		return nil, errors.Wrap(errors.InternalError, "解析微信API响应失败", err)
	}

	return &session, nil
}

// generateToken 生成JWT Token
func (s *AuthService) generateToken(openID string) (string, error) {
	now := time.Now()
	claims := jwt.RegisteredClaims{
		Subject:   openID,
		ExpiresAt: jwt.NewNumericDate(now.Add(time.Hour * time.Duration(s.cfg.JWT.ExpireHours))),
		IssuedAt:  jwt.NewNumericDate(now),
	}

	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)
	tokenString, err := token.SignedString([]byte(s.cfg.JWT.Secret))
	if err != nil {
		return "", errors.Wrap(errors.InternalError, "生成Token失败", err)
	}

	return tokenString, nil
}
