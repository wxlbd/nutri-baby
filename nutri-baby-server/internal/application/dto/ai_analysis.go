package dto

import (
	"encoding/json"
	"time"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
)

// CreateAnalysisRequest 创建分析请求
type CreateAnalysisRequest struct {
	BabyID       int64                 `json:"baby_id" binding:"required"`
	AnalysisType entity.AIAnalysisType `json:"analysis_type" binding:"required"`
	StartDate    CustomTime            `json:"start_date" binding:"required"`
	EndDate      CustomTime            `json:"end_date" binding:"required"`
}

// AnalysisResponse 分析响应
type AnalysisResponse struct {
	AnalysisID int64                    `json:"analysis_id"` // 修改为int64以匹配前端number类型
	Status     entity.AIAnalysisStatus  `json:"status"`
	Result     *entity.AIAnalysisResult `json:"result,omitempty"`
	CreatedAt  time.Time                `json:"created_at"`
}

// AnalysisStatusResponse 分析状态响应（用于轮询）
type AnalysisStatusResponse struct {
	AnalysisID string                  `json:"analysis_id"`
	Status     entity.AIAnalysisStatus `json:"status"`
	Progress   int                     `json:"progress"` // 进度百分比 0-100
	Message    string                  `json:"message"`  // 状态描述
	UpdatedAt  time.Time               `json:"updated_at"`
}

// DailyTipsResponse 每日建议响应
type DailyTipsResponse struct {
	Tips        []entity.DailyTip `json:"tips"`
	GeneratedAt time.Time         `json:"generated_at"`
	ExpiredAt   time.Time         `json:"expired_at"`
}

// BatchAnalysisRequest 批量分析请求
type BatchAnalysisRequest struct {
	BabyID        int64                   `json:"baby_id" binding:"required"`
	AnalysisTypes []entity.AIAnalysisType `json:"analysis_types" binding:"required"`
	StartDate     CustomTime              `json:"start_date" binding:"required"`
	EndDate       CustomTime              `json:"end_date" binding:"required"`
}

// BatchAnalysisResponse 批量分析响应
type BatchAnalysisResponse struct {
	TotalCount     int                `json:"total_count"`
	Analyses       []AnalysisResponse `json:"analyses"`
	CompletedCount int                `json:"completed_count"` // 新增：已完成数量
	FailedCount    int                `json:"failed_count"`    // 新增：失败数量
}

// AnalysisStatsResponse 分析统计响应
type AnalysisStatsResponse struct {
	TotalAnalyses int                             `json:"total_analyses"`
	ByType        map[entity.AIAnalysisType]int   `json:"by_type"`
	ByStatus      map[entity.AIAnalysisStatus]int `json:"by_status"`
	AvgScore      float64                         `json:"avg_score"`
	Trends        []AnalysisTrend                 `json:"trends"`
}

// AnalysisTrend 分析趋势
type AnalysisTrend struct {
	Date  time.Time             `json:"date"`
	Score float64               `json:"score"`
	Type  entity.AIAnalysisType `json:"type"`
}

// CustomTime 自定义时间类型
type CustomTime struct {
	time.Time
}

// UnmarshalJSON 自定义JSON反序列化
func (ct *CustomTime) UnmarshalJSON(b []byte) error {
	s := string(b)
	s = s[1 : len(s)-1] // 移除引号
	t, err := time.Parse("2006-01-02", s)
	if err != nil {
		return err
	}
	ct.Time = t
	return nil
}

// MarshalJSON 自定义JSON序列化
func (ct CustomTime) MarshalJSON() ([]byte, error) {
	return json.Marshal(ct.Time.Format("2006-01-02"))
}
