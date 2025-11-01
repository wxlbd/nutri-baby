package entity

import "time"

// Baby 宝宝实体 (去家庭化架构)
type Baby struct {
	BabyID      string     `gorm:"primaryKey;column:baby_id;type:varchar(64)" json:"babyId"`
	Name        string     `gorm:"column:name;type:varchar(64)" json:"name"`
	Nickname    string     `gorm:"column:nickname;type:varchar(64)" json:"nickname"`
	BirthDate   string     `gorm:"column:birth_date;type:varchar(10)" json:"birthDate"` // YYYY-MM-DD
	Gender      string     `gorm:"column:gender;type:varchar(16)" json:"gender"`        // male, female
	AvatarURL   string     `gorm:"column:avatar_url;type:varchar(512)" json:"avatarUrl"`
	Height      float64    `gorm:"column:height;type:decimal(10,2)" json:"height"`
	Weight      float64    `gorm:"column:weight;type:decimal(10,2)" json:"weight"`
	CreatorID   string     `gorm:"column:creator_id;type:varchar(64);index" json:"creatorId"`     // 创建者 openid
	FamilyGroup string     `gorm:"column:family_group;type:varchar(64);index" json:"familyGroup"` // 可选的家庭分组名称
	CreateTime  int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime  int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt   *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// 关联
	Collaborators []*BabyCollaborator `gorm:"foreignKey:BabyID;references:BabyID" json:"collaborators,omitempty"`
}

// TableName 指定表名
func (Baby) TableName() string {
	return "babies"
}

// BabyCollaborator 宝宝协作者实体 (替代 FamilyMember)
type BabyCollaborator struct {
	ID         int64      `gorm:"primaryKey;autoIncrement" json:"id"`
	BabyID     string     `gorm:"column:baby_id;type:varchar(64);index;uniqueIndex:idx_baby_user" json:"babyId"`
	OpenID     string     `gorm:"column:openid;type:varchar(64);index;uniqueIndex:idx_baby_user" json:"openid"`
	Role       string     `gorm:"column:role;type:varchar(16)" json:"role"`                                  // admin, editor, viewer
	AccessType string     `gorm:"column:access_type;type:varchar(16);default:'permanent'" json:"accessType"` // permanent, temporary
	ExpiresAt  *int64     `gorm:"column:expires_at" json:"expiresAt"`                                        // 临时权限过期时间(毫秒时间戳)
	JoinTime   int64      `gorm:"column:join_time;autoCreateTime:milli" json:"joinTime"`
	UpdateTime int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// 关联
	User *User `gorm:"foreignKey:OpenID;references:OpenID" json:"user,omitempty"`
	Baby *Baby `gorm:"foreignKey:BabyID;references:BabyID" json:"baby,omitempty"`
}

// TableName 指定表名
func (BabyCollaborator) TableName() string {
	return "baby_collaborators"
}

// IsExpired 检查临时权限是否过期
func (bc *BabyCollaborator) IsExpired() bool {
	if bc.AccessType != "temporary" || bc.ExpiresAt == nil {
		return false
	}
	return time.Now().UnixMilli() > *bc.ExpiresAt
}

// IsAdmin 检查是否为管理员
func (bc *BabyCollaborator) IsAdmin() bool {
	return bc.Role == "admin"
}

// CanEdit 检查是否有编辑权限
func (bc *BabyCollaborator) CanEdit() bool {
	return bc.Role == "admin" || bc.Role == "editor"
}
