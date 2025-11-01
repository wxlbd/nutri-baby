package entity

import "time"

// BabyInvitation 宝宝邀请记录 (用于微信分享和二维码)
type BabyInvitation struct {
	InvitationID string  `gorm:"primaryKey;type:varchar(36)" json:"invitation_id"` // 邀请ID
	BabyID       string  `gorm:"type:varchar(36);index;not null" json:"baby_id"`   // 宝宝ID
	InviterID    string  `gorm:"type:varchar(128);not null" json:"inviter_id"`     // 邀请人openid
	Token        string  `gorm:"type:varchar(64);uniqueIndex;not null" json:"token"` // 临时token(用于验证)
	ShortCode    string  `gorm:"type:varchar(10);uniqueIndex;not null" json:"short_code"` // 6位短码(用于小程序码scene参数)
	InviteType   string  `gorm:"type:varchar(20);not null" json:"invite_type"`    // share=分享, qrcode=二维码
	Role         string  `gorm:"type:varchar(20);not null" json:"role"`           // admin, editor, viewer
	AccessType   string  `gorm:"type:varchar(20);not null" json:"access_type"`    // permanent, temporary
	ExpiresAt    *int64  `gorm:"type:bigint" json:"expires_at"`                   // 协作权限过期时间(毫秒)
	ValidUntil   int64   `gorm:"type:bigint;not null;index" json:"valid_until"`   // 邀请链接有效期(毫秒,默认7天)
	UsedBy       *string `gorm:"type:varchar(128)" json:"used_by"`                // 被谁使用(openid)
	UsedAt       *int64  `gorm:"type:bigint" json:"used_at"`                      // 使用时间
	CreateTime   int64   `gorm:"type:bigint;not null" json:"create_time"`         // 创建时间
	DeletedAt    *int64  `gorm:"type:bigint;index" json:"deleted_at"`             // 软删除
}

// TableName 指定表名
func (BabyInvitation) TableName() string {
	return "baby_invitations"
}

// IsExpired 检查邀请是否过期
func (i *BabyInvitation) IsExpired() bool {
	if i.ValidUntil == 0 {
		return false
	}
	return i.ValidUntil < time.Now().UnixMilli()
}

// IsUsed 检查邀请是否已被使用
func (i *BabyInvitation) IsUsed() bool {
	return i.UsedBy != nil && i.UsedAt != nil
}
