package dto

// WechatLoginRequest 微信登录请求
type WechatLoginRequest struct {
	Code      string `json:"code" binding:"required"`
	NickName  string `json:"nickName"`
	AvatarURL string `json:"avatarUrl"`
}

// LoginResponse 登录响应 (去家庭化架构)
type LoginResponse struct {
	Token     string      `json:"token"`
	UserInfo  UserInfoDTO `json:"userInfo"`
	IsNewUser bool        `json:"isNewUser"` // 是否为新用户,前端根据此字段引导创建宝宝
}

// UserInfoDTO 用户信息DTO
type UserInfoDTO struct {
	OpenID        string `json:"openid"`
	NickName      string `json:"nickName"`
	AvatarURL     string `json:"avatarUrl"`
	DefaultBabyID string `json:"defaultBabyId"`
	CreateTime    int64  `json:"createTime"`
	LastLoginTime int64  `json:"lastLoginTime"`
}

// FamilyDTO 家庭DTO
type FamilyDTO struct {
	FamilyID    string `json:"familyId"`
	FamilyName  string `json:"familyName"`
	Description string `json:"description,omitempty"`
	Role        string `json:"role"` // admin, member
	JoinTime    int64  `json:"joinTime"`
}

// RefreshTokenResponse 刷新Token响应
type RefreshTokenResponse struct {
	Token     string `json:"token"`
	ExpiresIn int    `json:"expiresIn"`
}

// SetDefaultBabyRequest 设置默认宝宝请求
type SetDefaultBabyRequest struct {
	BabyID string `json:"babyId" binding:"required"`
}
