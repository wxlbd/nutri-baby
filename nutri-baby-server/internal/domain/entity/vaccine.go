package entity

import "time"

// VaccinePlanTemplate 疫苗计划模板(国家免疫规划标准)
type VaccinePlanTemplate struct {
	TemplateID   string     `gorm:"primaryKey;column:template_id;type:varchar(64)" json:"templateId"`
	VaccineType  string     `gorm:"column:vaccine_type;type:varchar(32);index" json:"vaccineType"`
	VaccineName  string     `gorm:"column:vaccine_name;type:varchar(64)" json:"vaccineName"`
	Description  string     `gorm:"column:description;type:text" json:"description"`
	AgeInMonths  int        `gorm:"column:age_in_months" json:"ageInMonths"`
	DoseNumber   int        `gorm:"column:dose_number" json:"doseNumber"`
	IsRequired   bool       `gorm:"column:is_required" json:"isRequired"`
	ReminderDays int        `gorm:"column:reminder_days;default:7" json:"reminderDays"`
	SortOrder    int        `gorm:"column:sort_order;default:0" json:"sortOrder"` // 排序
	CreateTime   int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime   int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (VaccinePlanTemplate) TableName() string {
	return "vaccine_plan_templates"
}

// BabyVaccinePlan 宝宝疫苗计划(基于模板,可自定义)
type BabyVaccinePlan struct {
	PlanID       string     `gorm:"primaryKey;column:plan_id;type:varchar(64)" json:"planId"`
	BabyID       string     `gorm:"column:baby_id;type:varchar(64);index" json:"babyId"`
	TemplateID   *string    `gorm:"column:template_id;type:varchar(64)" json:"templateId"` // 可选:来自哪个模板
	VaccineType  string     `gorm:"column:vaccine_type;type:varchar(32)" json:"vaccineType"`
	VaccineName  string     `gorm:"column:vaccine_name;type:varchar(64)" json:"vaccineName"`
	Description  string     `gorm:"column:description;type:text" json:"description"`
	AgeInMonths  int        `gorm:"column:age_in_months" json:"ageInMonths"`
	DoseNumber   int        `gorm:"column:dose_number" json:"doseNumber"`
	IsRequired   bool       `gorm:"column:is_required" json:"isRequired"`
	ReminderDays int        `gorm:"column:reminder_days;default:7" json:"reminderDays"`
	IsCustom     bool       `gorm:"column:is_custom;default:false" json:"isCustom"` // 是否用户自定义
	CreateBy     string     `gorm:"column:create_by;type:varchar(64)" json:"createBy"`
	CreateTime   int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime   int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// 关联
	Template *VaccinePlanTemplate `gorm:"foreignKey:TemplateID;references:TemplateID" json:"template,omitempty"`
	Baby     *Baby                `gorm:"foreignKey:BabyID;references:BabyID" json:"baby,omitempty"`
}

// TableName 指定表名
func (BabyVaccinePlan) TableName() string {
	return "baby_vaccine_plans"
}

// VaccinePlan 疫苗计划实体(系统预设) - @deprecated 使用 VaccinePlanTemplate 替代
type VaccinePlan struct {
	PlanID       string     `gorm:"primaryKey;column:plan_id;type:varchar(64)" json:"planId"`
	VaccineType  string     `gorm:"column:vaccine_type;type:varchar(32)" json:"vaccineType"`
	VaccineName  string     `gorm:"column:vaccine_name;type:varchar(64)" json:"vaccineName"`
	Description  string     `gorm:"column:description;type:text" json:"description"`
	AgeInMonths  int        `gorm:"column:age_in_months" json:"ageInMonths"`
	DoseNumber   int        `gorm:"column:dose_number" json:"doseNumber"`
	IsRequired   bool       `gorm:"column:is_required" json:"isRequired"`
	ReminderDays int        `gorm:"column:reminder_days" json:"reminderDays"`
	CreateTime   int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime   int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt    *time.Time `gorm:"column:deleted_at;index" json:"-"`
}

// TableName 指定表名
func (VaccinePlan) TableName() string {
	return "vaccine_plans"
}

// VaccineRecord 疫苗接种记录实体
type VaccineRecord struct {
	RecordID       string     `gorm:"primaryKey;column:record_id;type:varchar(64)" json:"recordId"`
	BabyID         string     `gorm:"column:baby_id;type:varchar(64);index" json:"babyId"`
	PlanID         string     `gorm:"column:plan_id;type:varchar(64);index" json:"planId"`
	VaccineType    string     `gorm:"column:vaccine_type;type:varchar(32)" json:"vaccineType"`
	VaccineName    string     `gorm:"column:vaccine_name;type:varchar(64)" json:"vaccineName"`
	DoseNumber     int        `gorm:"column:dose_number" json:"doseNumber"`
	VaccineDate    int64      `gorm:"column:vaccine_date;index" json:"vaccineDate"`
	Hospital       string     `gorm:"column:hospital;type:varchar(128)" json:"hospital"`
	BatchNumber    *string    `gorm:"column:batch_number;type:varchar(64)" json:"batchNumber"`
	Doctor         *string    `gorm:"column:doctor;type:varchar(64)" json:"doctor"`
	Reaction       *string    `gorm:"column:reaction;type:text" json:"reaction"`
	Note           *string    `gorm:"column:note;type:text" json:"note"`
	CreateBy       string     `gorm:"column:create_by;type:varchar(64)" json:"createBy"`
	CreateByName   string     `gorm:"column:create_by_name;type:varchar(64)" json:"createByName"`     // 冗余:创建者昵称
	CreateByAvatar string     `gorm:"column:create_by_avatar;type:varchar(512)" json:"createByAvatar"` // 冗余:创建者头像
	CreateTime     int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime     int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// 关联
	Plan *VaccinePlan `gorm:"foreignKey:PlanID;references:PlanID" json:"plan,omitempty"`
}

// TableName 指定表名
func (VaccineRecord) TableName() string {
	return "vaccine_records"
}

// VaccineReminder 疫苗提醒实体
type VaccineReminder struct {
	ReminderID    string    `gorm:"primaryKey;column:reminder_id;type:varchar(64)" json:"reminderId"`
	BabyID        string    `gorm:"column:baby_id;type:varchar(64);index" json:"babyId"`
	PlanID        string    `gorm:"column:plan_id;type:varchar(64);index" json:"planId"`
	VaccineName   string    `gorm:"column:vaccine_name;type:varchar(64)" json:"vaccineName"`
	DoseNumber    int       `gorm:"column:dose_number" json:"doseNumber"`
	ScheduledDate int64     `gorm:"column:scheduled_date;index" json:"scheduledDate"`
	Status        string    `gorm:"column:status;type:varchar(16);index" json:"status"` // upcoming, due, overdue, completed
	ReminderSent  bool      `gorm:"column:reminder_sent" json:"reminderSent"`
	SentTime      *int64    `gorm:"column:sent_time" json:"sentTime"`
	CreateTime    int64     `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime    int64     `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// 关联
	Plan *VaccinePlan `gorm:"foreignKey:PlanID;references:PlanID" json:"plan,omitempty"`
	Baby *Baby        `gorm:"foreignKey:BabyID;references:BabyID" json:"baby,omitempty"`
}

// TableName 指定表名
func (VaccineReminder) TableName() string {
	return "vaccine_reminders"
}

// DaysUntilDue 计算距离预定日期的天数
func (v *VaccineReminder) DaysUntilDue() int {
	now := time.Now().UnixMilli()
	diff := v.ScheduledDate - now
	return int(diff / (24 * 60 * 60 * 1000))
}

// UpdateStatus 更新提醒状态
func (v *VaccineReminder) UpdateStatus() {
	days := v.DaysUntilDue()

	if v.Status == "completed" {
		return
	}

	if days > 7 {
		v.Status = "upcoming"
	} else if days >= 0 {
		v.Status = "due"
	} else {
		v.Status = "overdue"
	}
}
