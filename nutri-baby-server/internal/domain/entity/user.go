package entity

import "time"

// User 用户实体
type User struct {
	OpenID        string     `gorm:"primaryKey;column:openid;type:varchar(64)" json:"openid"`
	NickName      string     `gorm:"column:nick_name;type:varchar(64)" json:"nickName"`
	AvatarURL     string     `gorm:"column:avatar_url;type:varchar(512)" json:"avatarUrl"`
	CreateTime    int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	LastLoginTime int64      `gorm:"column:last_login_time" json:"lastLoginTime"`
	UpdateTime    int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (User) TableName() string {
	return "users"
}

// Invitation 邀请码实体
type Invitation struct {
	InvitationCode string     `gorm:"primaryKey;column:invitation_code;type:varchar(16)" json:"invitationCode"`
	FamilyID       string     `gorm:"column:family_id;type:varchar(64);index" json:"familyId"`
	CreatorID      string     `gorm:"column:creator_id;type:varchar(64)" json:"creatorId"`
	ExpiresAt      int64      `gorm:"column:expires_at;index" json:"expiresAt"`
	CreateTime     int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (Invitation) TableName() string {
	return "invitations"
}

// IsExpired 检查邀请码是否过期
func (i *Invitation) IsExpired() bool {
	return time.Now().UnixMilli() > i.ExpiresAt
}
