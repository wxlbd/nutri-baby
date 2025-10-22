package dto

// CreateFamilyRequest 创建家庭请求
type CreateFamilyRequest struct {
	FamilyName string `json:"familyName" binding:"required"`
}

// UpdateFamilyRequest 更新家庭请求
type UpdateFamilyRequest struct {
	FamilyName string `json:"familyName" binding:"required"`
}

// CreateInvitationRequest 创建邀请请求
type CreateInvitationRequest struct {
	FamilyID string `json:"familyId" binding:"required"`
	Role     string `json:"role" binding:"required,oneof=admin member"`
}

// InvitationDTO 邀请DTO
type InvitationDTO struct {
	InvitationCode string `json:"invitationCode"`
	FamilyID       string `json:"familyId"`
	FamilyName     string `json:"familyName"`
	Role           string `json:"role"`
	CreateBy       string `json:"createBy"`
	CreateTime     int64  `json:"createTime"`
	ExpireTime     int64  `json:"expireTime"`
}

// JoinFamilyRequest 加入家庭请求
type JoinFamilyRequest struct {
	InvitationCode string `json:"invitationCode" binding:"required"`
}

// FamilyMemberDTO 家庭成员DTO
type FamilyMemberDTO struct {
	MemberID string       `json:"memberId"`
	FamilyID string       `json:"familyId"`
	UserID   string       `json:"userId"`
	UserInfo UserInfoDTO  `json:"userInfo"`
	Role     string       `json:"role"`
	JoinTime int64        `json:"joinTime"`
}

// FamilyDetailDTO 家庭详情DTO
type FamilyDetailDTO struct {
	FamilyID   string            `json:"familyId"`
	FamilyName string            `json:"familyName"`
	CreateBy   string            `json:"createBy"`
	CreateTime int64             `json:"createTime"`
	Members    []FamilyMemberDTO `json:"members"`
}
