package dto

// FeedingDetail 喂养详情
type FeedingDetail struct {
	BreastSide  string `json:"breastSide,omitempty"`  // left, right, both
	LeftTime    int    `json:"leftTime,omitempty"`    // 左侧时长(分钟)
	RightTime   int    `json:"rightTime,omitempty"`   // 右侧时长(分钟)
	FormulaType string `json:"formulaType,omitempty"` // 奶粉类型
}

// CreateFeedingRecordRequest 创建喂养记录请求
type CreateFeedingRecordRequest struct {
	BabyID      string        `json:"babyId" binding:"required"`
	FeedingType string        `json:"feedingType" binding:"required,oneof=breast formula mixed"`
	Amount      int           `json:"amount"`      // ml
	Duration    int           `json:"duration"`    // 分钟
	Detail      FeedingDetail `json:"detail"`
	Note        string        `json:"note"`
	FeedingTime int64         `json:"feedingTime"` // 毫秒时间戳
}

// FeedingRecordDTO 喂养记录DTO
type FeedingRecordDTO struct {
	RecordID    string        `json:"recordId"`
	BabyID      string        `json:"babyId"`
	FeedingType string        `json:"feedingType"`
	Amount      int           `json:"amount"`
	Duration    int           `json:"duration"`
	Detail      FeedingDetail `json:"detail"`
	Note        string        `json:"note"`
	FeedingTime int64         `json:"feedingTime"`
	CreateBy    string        `json:"createBy"`
	CreateTime  int64         `json:"createTime"`
}

// CreateSleepRecordRequest 创建睡眠记录请求
type CreateSleepRecordRequest struct {
	BabyID    string `json:"babyId" binding:"required"`
	StartTime int64  `json:"startTime" binding:"required"`
	EndTime   int64  `json:"endTime"`
	Duration  int    `json:"duration"` // 分钟
	Quality   string `json:"quality" binding:"omitempty,oneof=good fair poor"`
	Note      string `json:"note"`
}

// SleepRecordDTO 睡眠记录DTO
type SleepRecordDTO struct {
	RecordID   string `json:"recordId"`
	BabyID     string `json:"babyId"`
	StartTime  int64  `json:"startTime"`
	EndTime    int64  `json:"endTime"`
	Duration   int    `json:"duration"`
	Quality    string `json:"quality"`
	Note       string `json:"note"`
	CreateBy   string `json:"createBy"`
	CreateTime int64  `json:"createTime"`
}

// CreateDiaperRecordRequest 创建尿布记录请求
type CreateDiaperRecordRequest struct {
	BabyID     string `json:"babyId" binding:"required"`
	DiaperType string `json:"diaperType" binding:"required,oneof=pee poop both"`
	Note       string `json:"note"`
	ChangeTime int64  `json:"changeTime"` // 毫秒时间戳
}

// DiaperRecordDTO 尿布记录DTO
type DiaperRecordDTO struct {
	RecordID   string `json:"recordId"`
	BabyID     string `json:"babyId"`
	DiaperType string `json:"diaperType"`
	Note       string `json:"note"`
	ChangeTime int64  `json:"changeTime"`
	CreateBy   string `json:"createBy"`
	CreateTime int64  `json:"createTime"`
}

// CreateGrowthRecordRequest 创建生长记录请求
type CreateGrowthRecordRequest struct {
	BabyID      string `json:"babyId" binding:"required"`
	Height      int    `json:"height" binding:"required"` // cm
	Weight      int    `json:"weight" binding:"required"` // g
	HeadCircum  int    `json:"headCircum"`                // cm
	Note        string `json:"note"`
	RecordTime  int64  `json:"recordTime"` // 毫秒时间戳
}

// GrowthRecordDTO 生长记录DTO
type GrowthRecordDTO struct {
	RecordID   string `json:"recordId"`
	BabyID     string `json:"babyId"`
	Height     int    `json:"height"`
	Weight     int    `json:"weight"`
	HeadCircum int    `json:"headCircum"`
	Note       string `json:"note"`
	RecordTime int64  `json:"recordTime"`
	CreateBy   string `json:"createBy"`
	CreateTime int64  `json:"createTime"`
}

// RecordListQuery 记录列表查询参数
type RecordListQuery struct {
	BabyID    string `form:"babyId"`
	StartTime int64  `form:"startTime"`
	EndTime   int64  `form:"endTime"`
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
}
