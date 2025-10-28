package dto

// FeedingDetail 喂养详情
type FeedingDetail struct {
	// 母乳喂养相关
	BreastSide string `json:"breastSide,omitempty"` // left, right, both
	LeftTime   int    `json:"leftTime,omitempty"`   // 左侧时长(秒)
	RightTime  int    `json:"rightTime,omitempty"`  // 右侧时长(秒)
	Duration   int    `json:"duration,omitempty"`   // 总时长(秒)

	// 奶瓶喂养相关
	BottleType string  `json:"bottleType,omitempty"` // formula, breast-milk
	Unit       string  `json:"unit,omitempty"`       // ml, oz
	Remaining  float64 `json:"remaining,omitempty"`  // 剩余量

	// 辅食相关
	FoodName string `json:"foodName,omitempty"` // 辅食名称

	// 通用
	Note        string `json:"note,omitempty"`        // 备注
	FormulaType string `json:"formulaType,omitempty"` // 兼容旧数据
}

// CreateFeedingRecordRequest 创建喂养记录请求
type CreateFeedingRecordRequest struct {
	BabyID      string        `json:"babyId" binding:"required"`
	FeedingType string        `json:"feedingType" binding:"required,oneof=breast bottle food"`
	Amount      int64         `json:"amount"`   // ml
	Duration    int           `json:"duration"` // 秒
	Detail      FeedingDetail `json:"detail"`
	Note        string        `json:"note"`
	FeedingTime int64         `json:"feedingTime"` // 毫秒时间戳
}

// FeedingRecordDTO 喂养记录DTO
type FeedingRecordDTO struct {
	RecordID    string        `json:"recordId"`
	BabyID      string        `json:"babyId"`
	FeedingType string        `json:"feedingType"`
	Amount      int64         `json:"amount"`
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
	Duration  int    `json:"duration"` // 秒
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
	BabyID            string  `json:"babyId" binding:"required"`
	Height            float64 `json:"height"`            // cm
	Weight            float64 `json:"weight"`            // kg
	HeadCircumference float64 `json:"headCircumference"` // cm
	Note              string  `json:"note"`
	MeasureTime       int64   `json:"measureTime"` // 毫秒时间戳
}

// GrowthRecordDTO 生长记录DTO
type GrowthRecordDTO struct {
	RecordID          string   `json:"recordId"`
	BabyID            string   `json:"babyId"`
	Height            *float64 `json:"height,omitempty"`            // cm, 仅当有值时返回
	Weight            *float64 `json:"weight,omitempty"`            // kg, 仅当有值时返回
	HeadCircumference *float64 `json:"headCircumference,omitempty"` // cm, 仅当有值时返回
	Note              string   `json:"note,omitempty"`
	MeasureTime       int64    `json:"measureTime"`
	CreateBy          string   `json:"createBy"`
	CreateTime        int64    `json:"createTime"`
}

// RecordListQuery 记录列表查询参数
type RecordListQuery struct {
	BabyID    string `form:"babyId"`
	StartTime int64  `form:"startTime"`
	EndTime   int64  `form:"endTime"`
	Page      int    `form:"page"`
	PageSize  int    `form:"pageSize"`
}
