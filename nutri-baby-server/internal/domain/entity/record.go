package entity

import (
	"database/sql/driver"
	"encoding/json"
	"time"
)

// 喂养类型常量 (与前端保持一致)
const (
	FeedingTypeBreast = "breast" // 母乳喂养
	FeedingTypeBottle = "bottle" // 奶瓶喂养
	FeedingTypeFood   = "food"   // 辅食
)

// FeedingRecord 喂养记录实体
type FeedingRecord struct {
	RecordID       string        `gorm:"primaryKey;column:record_id;type:varchar(64)" json:"recordId"`
	BabyID         string        `gorm:"column:baby_id;type:varchar(64);index" json:"babyId"`
	Time           int64         `gorm:"column:time;index" json:"time"`
	FeedingType    string        `gorm:"column:feeding_type;type:varchar(16);not null" json:"feedingType"` // 喂养类型: breast/bottle/food
	Amount         int64         `gorm:"column:amount" json:"amount,omitempty"`                            // 奶量(ml)，bottle类型时使用
	Duration       int           `gorm:"column:duration" json:"duration,omitempty"`                        // 时长(秒)，breast类型时使用
	Detail         FeedingDetail `gorm:"column:detail;type:jsonb" json:"detail"`                           // 完整详情(向后兼容)
	CreateBy       string        `gorm:"column:create_by;type:varchar(64)" json:"createBy"`
	CreateByName   string        `gorm:"column:create_by_name;type:varchar(64)" json:"createByName"`      // 冗余:创建者昵称
	CreateByAvatar string        `gorm:"column:create_by_avatar;type:varchar(512)" json:"createByAvatar"` // 冗余:创建者头像

	// 提醒相关字段
	ReminderInterval *int   `gorm:"column:reminder_interval" json:"reminderInterval,omitempty"` // 提醒间隔(分钟)
	NextReminderTime *int64 `gorm:"column:next_reminder_time" json:"nextReminderTime,omitempty"` // 下次提醒时间戳(毫秒)
	ReminderSent     bool   `gorm:"column:reminder_sent;default:false;index" json:"reminderSent"` // 是否已发送提醒
	ReminderTime     *int64 `gorm:"column:reminder_time" json:"reminderTime,omitempty"`           // 提醒发送时间戳(毫秒)

	CreateTime int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (FeedingRecord) TableName() string {
	return "feeding_records"
}

// FeedingDetail 喂养详情(使用interface{}存储不同类型)
type FeedingDetail map[string]any

// Scan 实现sql.Scanner接口
func (f *FeedingDetail) Scan(value any) error {
	bytes, ok := value.([]byte)
	if !ok {
		return nil
	}
	return json.Unmarshal(bytes, f)
}

// Value 实现driver.Valuer接口
func (f FeedingDetail) Value() (driver.Value, error) {
	return json.Marshal(f)
}

// SleepRecord 睡眠记录实体
type SleepRecord struct {
	RecordID       string     `gorm:"primaryKey;column:record_id;type:varchar(64)" json:"recordId"`
	BabyID         string     `gorm:"column:baby_id;type:varchar(64);index" json:"babyId"`
	StartTime      int64      `gorm:"column:start_time;index" json:"startTime"`
	EndTime        *int64     `gorm:"column:end_time" json:"endTime"`
	Duration       *int       `gorm:"column:duration" json:"duration"`          // 时长(秒)
	Type           string     `gorm:"column:type;type:varchar(16)" json:"type"` // nap, night
	CreateBy       string     `gorm:"column:create_by;type:varchar(64)" json:"createBy"`
	CreateByName   string     `gorm:"column:create_by_name;type:varchar(64)" json:"createByName"`      // 冗余:创建者昵称
	CreateByAvatar string     `gorm:"column:create_by_avatar;type:varchar(512)" json:"createByAvatar"` // 冗余:创建者头像
	CreateTime     int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime     int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (SleepRecord) TableName() string {
	return "sleep_records"
}

// DiaperRecord 换尿布记录实体
type DiaperRecord struct {
	RecordID       string     `gorm:"primaryKey;column:record_id;type:varchar(64)" json:"recordId"`
	BabyID         string     `gorm:"column:baby_id;type:varchar(64);index" json:"babyId"`
	Time           int64      `gorm:"column:time;index" json:"time"`
	Type           string     `gorm:"column:type;type:varchar(16)" json:"type"` // wet, dirty, both
	PoopColor      *string    `gorm:"column:poop_color;type:varchar(16)" json:"poopColor"`
	PoopTexture    *string    `gorm:"column:poop_texture;type:varchar(16)" json:"poopTexture"`
	Note           *string    `gorm:"column:note;type:text" json:"note"`
	CreateBy       string     `gorm:"column:create_by;type:varchar(64)" json:"createBy"`
	CreateByName   string     `gorm:"column:create_by_name;type:varchar(64)" json:"createByName"`      // 冗余:创建者昵称
	CreateByAvatar string     `gorm:"column:create_by_avatar;type:varchar(512)" json:"createByAvatar"` // 冗余:创建者头像
	CreateTime     int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime     int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (DiaperRecord) TableName() string {
	return "diaper_records"
}

// GrowthRecord 成长记录实体
type GrowthRecord struct {
	RecordID          string     `gorm:"primaryKey;column:record_id;type:varchar(64)" json:"recordId"`
	BabyID            string     `gorm:"column:baby_id;type:varchar(64);index" json:"babyId"`
	Time              int64      `gorm:"column:time;index" json:"time"`
	Height            *float64   `gorm:"column:height" json:"height"`                        // cm
	Weight            *float64   `gorm:"column:weight" json:"weight"`                        // kg
	HeadCircumference *float64   `gorm:"column:head_circumference" json:"headCircumference"` // cm
	Note              *string    `gorm:"column:note;type:text" json:"note"`
	CreateBy          string     `gorm:"column:create_by;type:varchar(64)" json:"createBy"`
	CreateByName      string     `gorm:"column:create_by_name;type:varchar(64)" json:"createByName"`      // 冗余:创建者昵称
	CreateByAvatar    string     `gorm:"column:create_by_avatar;type:varchar(512)" json:"createByAvatar"` // 冗余:创建者头像
	CreateTime        int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime        int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt         *time.Time `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (GrowthRecord) TableName() string {
	return "growth_records"
}
