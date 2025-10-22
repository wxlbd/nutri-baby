package dto

// VaccineRecordDTO 疫苗接种记录DTO
type VaccineRecordDTO struct {
	RecordID    string  `json:"recordId"`
	BabyID      string  `json:"babyId"`
	PlanID      string  `json:"planId"`
	VaccineType string  `json:"vaccineType"`
	VaccineName string  `json:"vaccineName"`
	DoseNumber  int     `json:"doseNumber"`
	VaccineDate int64   `json:"vaccineDate"`
	Hospital    string  `json:"hospital"`
	BatchNumber *string `json:"batchNumber,omitempty"`
	Doctor      *string `json:"doctor,omitempty"`
	Reaction    *string `json:"reaction,omitempty"`
	Note        *string `json:"note,omitempty"`
	CreateBy    string  `json:"createBy"`
	CreateTime  int64   `json:"createTime"`
}

// CreateVaccineRecordRequest 创建疫苗记录请求
type CreateVaccineRecordRequest struct {
	PlanID      string  `json:"planId" binding:"required"`
	VaccineType string  `json:"vaccineType" binding:"required"`
	VaccineName string  `json:"vaccineName" binding:"required"`
	DoseNumber  int     `json:"doseNumber" binding:"required"`
	VaccineDate int64   `json:"vaccineDate" binding:"required"`
	Hospital    string  `json:"hospital" binding:"required"`
	BatchNumber *string `json:"batchNumber"`
	Doctor      *string `json:"doctor"`
	Reaction    *string `json:"reaction"`
	Note        *string `json:"note"`
}

// VaccineReminderDTO 疫苗提醒DTO
type VaccineReminderDTO struct {
	ReminderID    string `json:"reminderId"`
	BabyID        string `json:"babyId"`
	BabyName      string `json:"babyName,omitempty"`
	PlanID        string `json:"planId"`
	VaccineName   string `json:"vaccineName"`
	DoseNumber    int    `json:"doseNumber"`
	ScheduledDate int64  `json:"scheduledDate"`
	Status        string `json:"status"`
	DaysUntilDue  int    `json:"daysUntilDue"`
	ReminderSent  bool   `json:"reminderSent"`
	CreateTime    int64  `json:"createTime"`
}

// VaccineStatisticsDTO 疫苗统计DTO
type VaccineStatisticsDTO struct {
	Total         int64              `json:"total"`
	Completed     int64              `json:"completed"`
	Pending       int64              `json:"pending"`
	Overdue       int64              `json:"overdue"`
	Percentage    int                `json:"percentage"`
	NextVaccine   *NextVaccineDTO    `json:"nextVaccine,omitempty"`
	RecentRecords []VaccineRecordDTO `json:"recentRecords"`
}

// NextVaccineDTO 下一个待接种疫苗
type NextVaccineDTO struct {
	VaccineName   string `json:"vaccineName"`
	DoseNumber    int    `json:"doseNumber"`
	ScheduledDate int64  `json:"scheduledDate"`
	DaysUntilDue  int    `json:"daysUntilDue"`
}
