package entity

import "time"

// 疫苗提醒状态常量
const (
	ReminderStatusUpcoming  = "upcoming"  // 即将到期
	ReminderStatusDue       = "due"       // 应接种
	ReminderStatusOverdue   = "overdue"   // 已逾期
	ReminderStatusCompleted = "completed" // 已完成
)

// 疫苗接种状态常量
const (
	VaccinationStatusPending   = "pending"   // 未接种
	VaccinationStatusCompleted = "completed" // 已完成
	VaccinationStatusSkipped   = "skipped"   // 跳过/不接种
)

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
	CreateByName   string     `gorm:"column:create_by_name;type:varchar(64)" json:"createByName"`      // 冗余:创建者昵称
	CreateByAvatar string     `gorm:"column:create_by_avatar;type:varchar(512)" json:"createByAvatar"` // 冗余:创建者头像
	CreateTime     int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime     int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt      *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// 关联
	Plan *BabyVaccinePlan `gorm:"foreignKey:PlanID;references:PlanID" json:"plan,omitempty"`
}

// TableName 指定表名
func (VaccineRecord) TableName() string {
	return "vaccine_records"
}

// VaccineReminder 疫苗提醒实体
type VaccineReminder struct {
	ReminderID    string     `gorm:"primaryKey;column:reminder_id;type:varchar(64)" json:"reminderId"`
	BabyID        string     `gorm:"column:baby_id;type:varchar(64);index" json:"babyId"`
	PlanID        string     `gorm:"column:plan_id;type:varchar(64);index" json:"planId"`
	VaccineName   string     `gorm:"column:vaccine_name;type:varchar(64)" json:"vaccineName"`
	DoseNumber    int        `gorm:"column:dose_number" json:"doseNumber"`
	ScheduledDate int64      `gorm:"column:scheduled_date;index" json:"scheduledDate"`
	Status        string     `gorm:"column:status;type:varchar(16);index" json:"status"` // upcoming, due, overdue, completed
	ReminderSent  bool       `gorm:"column:reminder_sent" json:"reminderSent"`
	SentTime      *int64     `gorm:"column:sent_time" json:"sentTime"`
	CreateTime    int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime    int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt     *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// 关联
	Plan *BabyVaccinePlan `gorm:"foreignKey:PlanID;references:PlanID" json:"plan,omitempty"`
	Baby *Baby            `gorm:"foreignKey:BabyID;references:BabyID" json:"baby,omitempty"`
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

	if v.Status == ReminderStatusCompleted {
		return
	}

	if days > 7 {
		v.Status = ReminderStatusUpcoming
	} else if days >= 0 {
		v.Status = ReminderStatusDue
	} else {
		v.Status = ReminderStatusOverdue
	}
}

// BabyVaccineSchedule 宝宝疫苗接种日程(合并计划和记录)
type BabyVaccineSchedule struct {
	// 主键 (复用原 plan_id,保持与 vaccine_reminders 的外键关系)
	ScheduleID string `gorm:"primaryKey;column:schedule_id;type:varchar(64)" json:"scheduleId"`

	// 宝宝和计划基础信息
	BabyID     string  `gorm:"column:baby_id;type:varchar(64);index;not null" json:"babyId"`
	TemplateID *string `gorm:"column:template_id;type:varchar(64)" json:"templateId"` // 来源模板ID(可选)

	// 疫苗基本信息
	VaccineType  string `gorm:"column:vaccine_type;type:varchar(32);not null" json:"vaccineType"` // 疫苗类型
	VaccineName  string `gorm:"column:vaccine_name;type:varchar(64);not null" json:"vaccineName"` // 疫苗名称
	Description  string `gorm:"column:description;type:text" json:"description"`                  // 描述
	AgeInMonths  int    `gorm:"column:age_in_months;not null" json:"ageInMonths"`                 // 接种年龄(月)
	DoseNumber   int    `gorm:"column:dose_number;not null" json:"doseNumber"`                    // 接种次数
	IsRequired   bool   `gorm:"column:is_required;default:true" json:"isRequired"`                // 是否必接种
	ReminderDays int    `gorm:"column:reminder_days;default:7" json:"reminderDays"`               // 提前多少天提醒
	IsCustom     bool   `gorm:"column:is_custom;default:false" json:"isCustom"`                   // 是否用户自定义

	// 接种状态 (关键字段)
	VaccinationStatus string `gorm:"column:vaccination_status;type:varchar(16);not null;default:'pending';index" json:"vaccinationStatus"`
	// 状态值: pending(未接种), completed(已完成), skipped(跳过/不接种)

	// 接种记录信息 (仅在 status='completed' 时有值)
	VaccineDate       *int64  `gorm:"column:vaccine_date" json:"vaccineDate"`            // 实际接种日期(毫秒时间戳)
	Hospital          *string `gorm:"column:hospital;type:varchar(128)" json:"hospital"` // 接种医院
	BatchNumber       *string `gorm:"column:batch_number;type:varchar(64)" json:"batchNumber"`
	Doctor            *string `gorm:"column:doctor;type:varchar(64)" json:"doctor"`
	Reaction          *string `gorm:"column:reaction;type:text" json:"reaction"`
	Note              *string `gorm:"column:note;type:text" json:"note"`
	CompletedBy       *string `gorm:"column:completed_by;type:varchar(64)" json:"completedBy"`          // 记录接种的用户openid
	CompletedByName   *string `gorm:"column:completed_by_name;type:varchar(64)" json:"completedByName"` // 记录者昵称
	CompletedByAvatar *string `gorm:"column:completed_by_avatar;type:varchar(512)" json:"completedByAvatar"`
	CompletedTime     *int64  `gorm:"column:completed_time" json:"completedTime"` // 接种记录创建时间

	// 提醒相关字段 (合并自 VaccineReminder)
	ScheduledDate  int64  `gorm:"column:scheduled_date;index" json:"scheduledDate"`        // 计划接种日期(毫秒时间戳)
	ReminderSent   bool   `gorm:"column:reminder_sent;default:false" json:"reminderSent"`  // 是否已发送提醒
	ReminderSentAt *int64 `gorm:"column:reminder_sent_at" json:"reminderSentAt,omitempty"` // 提醒发送时间

	// 审计字段
	CreateBy   string     `gorm:"column:create_by;type:varchar(64);not null" json:"createBy"`
	CreateTime int64      `gorm:"column:create_time;autoCreateTime:milli" json:"createTime"`
	UpdateTime int64      `gorm:"column:update_time;autoUpdateTime:milli" json:"updateTime"`
	DeletedAt  *time.Time `gorm:"column:deleted_at;index" json:"-"`

	// 关联
	Template *VaccinePlanTemplate `gorm:"foreignKey:TemplateID;references:TemplateID" json:"template,omitempty"`
	Baby     *Baby                `gorm:"foreignKey:BabyID;references:BabyID" json:"baby,omitempty"`
}

// TableName 指定表名
func (BabyVaccineSchedule) TableName() string {
	return "baby_vaccine_schedules"
}

// IsCompleted 判断是否已完成接种
func (s *BabyVaccineSchedule) IsCompleted() bool {
	return s.VaccinationStatus == VaccinationStatusCompleted
}

// IsPending 判断是否待接种
func (s *BabyVaccineSchedule) IsPending() bool {
	return s.VaccinationStatus == VaccinationStatusPending
}

// IsSkipped 判断是否已跳过
func (s *BabyVaccineSchedule) IsSkipped() bool {
	return s.VaccinationStatus == VaccinationStatusSkipped
}

// MarkAsCompleted 标记为已完成
func (s *BabyVaccineSchedule) MarkAsCompleted(vaccineDate int64, hospital string, completedBy, completedByName, completedByAvatar string) {
	s.VaccinationStatus = VaccinationStatusCompleted
	s.VaccineDate = &vaccineDate
	s.Hospital = &hospital
	s.CompletedBy = &completedBy
	s.CompletedByName = &completedByName
	s.CompletedByAvatar = &completedByAvatar
	now := time.Now().UnixMilli()
	s.CompletedTime = &now
}

// MarkAsSkipped 标记为跳过
func (s *BabyVaccineSchedule) MarkAsSkipped() {
	s.VaccinationStatus = VaccinationStatusSkipped
}

// GetReminderStatus 获取提醒状态(实时计算)
func (s *BabyVaccineSchedule) GetReminderStatus() string {
	if s.VaccinationStatus != VaccinationStatusPending {
		return ReminderStatusCompleted
	}
	days := s.DaysUntilDue()
	if days > 7 {
		return ReminderStatusUpcoming
	} else if days >= 0 {
		return ReminderStatusDue
	}
	return ReminderStatusOverdue
}

// DaysUntilDue 计算距离应接种日期的天数
func (s *BabyVaccineSchedule) DaysUntilDue() int {
	now := time.Now().UnixMilli()
	diff := s.ScheduledDate - now
	return int(diff / (24 * 60 * 60 * 1000))
}

// MarkReminderSent 标记提醒已发送
func (s *BabyVaccineSchedule) MarkReminderSent() {
	s.ReminderSent = true
	now := time.Now().UnixMilli()
	s.ReminderSentAt = &now
}
