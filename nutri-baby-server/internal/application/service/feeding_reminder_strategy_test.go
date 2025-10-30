package service

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
)

func TestBreastFeedingReminderStrategy(t *testing.T) {
	strategy := NewBreastFeedingReminderStrategy()

	// 测试 GetTemplateType
	assert.Equal(t, "breast_feeding_reminder", strategy.GetTemplateType())

	// 测试 CanHandle - 母乳喂养
	record := &entity.FeedingRecord{
		Detail: entity.FeedingDetail{
			"type": "breast",
			"side": "left",
			"duration": float64(600), // 10分钟
		},
	}
	assert.True(t, strategy.CanHandle(record))

	// 测试 BuildMessageData
	lastFeedingTime := time.Now().Add(-2 * time.Hour)
	hoursSinceLastFeeding := 2.0
	messageData := strategy.BuildMessageData(record, lastFeedingTime, hoursSinceLastFeeding)

	assert.NotNil(t, messageData)
	assert.Contains(t, messageData, "time1")
	assert.Contains(t, messageData, "thing2")
	assert.Contains(t, messageData, "character_string3")
	assert.Contains(t, messageData, "phrase4")
	assert.Contains(t, messageData, "thing5")
	assert.Equal(t, "左侧", messageData["phrase4"])
	assert.Equal(t, "10分钟", messageData["character_string3"])
}

func TestBottleFeedingReminderStrategy(t *testing.T) {
	strategy := NewBottleFeedingReminderStrategy()

	// 测试 GetTemplateType
	assert.Equal(t, "bottle_feeding_reminder", strategy.GetTemplateType())

	// 测试 CanHandle - 奶瓶喂养
	record := &entity.FeedingRecord{
		Detail: entity.FeedingDetail{
			"type":       "bottle",
			"bottleType": "formula",
			"amount":     float64(120),
		},
	}
	assert.True(t, strategy.CanHandle(record))

	// 测试 BuildMessageData
	lastFeedingTime := time.Now().Add(-3 * time.Hour)
	hoursSinceLastFeeding := 3.0
	messageData := strategy.BuildMessageData(record, lastFeedingTime, hoursSinceLastFeeding)

	assert.NotNil(t, messageData)
	assert.Contains(t, messageData, "time1")
	assert.Contains(t, messageData, "thing2")
	assert.Contains(t, messageData, "character_string3")
	assert.Contains(t, messageData, "phrase4")
	assert.Contains(t, messageData, "thing5")
	assert.Equal(t, "配方奶", messageData["phrase4"])
	assert.Equal(t, "120ml", messageData["character_string3"])
}

func TestFoodFeedingReminderStrategy(t *testing.T) {
	strategy := NewFoodFeedingReminderStrategy()

	// 测试 GetTemplateType
	assert.Equal(t, "food_feeding_reminder", strategy.GetTemplateType())

	// 测试 CanHandle - 辅食
	record := &entity.FeedingRecord{
		Detail: entity.FeedingDetail{
			"type":     "food",
			"foodName": "米糊",
		},
	}
	assert.True(t, strategy.CanHandle(record))

	// 测试 BuildMessageData
	lastFeedingTime := time.Now().Add(-4 * time.Hour)
	hoursSinceLastFeeding := 4.0
	messageData := strategy.BuildMessageData(record, lastFeedingTime, hoursSinceLastFeeding)

	assert.NotNil(t, messageData)
	assert.Contains(t, messageData, "time1")
	assert.Contains(t, messageData, "thing2")
	assert.Contains(t, messageData, "character_string3")
	assert.Contains(t, messageData, "phrase4")
	assert.Contains(t, messageData, "thing5")
	assert.Equal(t, "辅食", messageData["phrase4"])
	assert.Equal(t, "米糊", messageData["character_string3"])
}

func TestFeedingReminderStrategyFactory(t *testing.T) {
	factory := NewFeedingReminderStrategyFactory()

	// 测试母乳喂养策略选择
	breastRecord := &entity.FeedingRecord{
		Detail: entity.FeedingDetail{
			"type": "breast",
		},
	}
	strategy,_ := factory.GetStrategy(breastRecord)
	assert.Equal(t, "breast_feeding_reminder", strategy.GetTemplateType())

	// 测试奶瓶喂养策略选择
	bottleRecord := &entity.FeedingRecord{
		Detail: entity.FeedingDetail{
			"type": "bottle",
		},
	}
	strategy,_ = factory.GetStrategy(bottleRecord)
	assert.Equal(t, "bottle_feeding_reminder", strategy.GetTemplateType())

	// 测试辅食策略选择
	foodRecord := &entity.FeedingRecord{
		Detail: entity.FeedingDetail{
			"type": "food",
		},
	}
	strategy,_ = factory.GetStrategy(foodRecord)
	assert.Equal(t, "food_feeding_reminder", strategy.GetTemplateType())

	// 测试未知类型（应返回默认母乳策略）
	unknownRecord := &entity.FeedingRecord{
		Detail: entity.FeedingDetail{
			"type": "unknown",
		},
	}
	strategy,_ = factory.GetStrategy(unknownRecord)
	assert.Equal(t, "breast_feeding_reminder", strategy.GetTemplateType())
}

func TestFormatTimeSince(t *testing.T) {
	tests := []struct {
		name     string
		hours    float64
		expected string
	}{
		{"30分钟", 0.5, "约30分钟"},
		{"45分钟", 0.75, "约45分钟"},
		{"1小时", 1.0, "约1小时"},
		{"2小时", 2.0, "约2小时"},
		{"3小时", 3.0, "约3小时"},
		{"24小时", 24.0, "约24小时"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := formatTimeSince(tt.hours)
			assert.Equal(t, tt.expected, result)
		})
	}
}

func TestBreastFeedingDifferentSides(t *testing.T) {
	strategy := NewBreastFeedingReminderStrategy()

	testCases := []struct {
		side     string
		expected string
	}{
		{"left", "左侧"},
		{"right", "右侧"},
		{"both", "两侧"},
	}

	for _, tc := range testCases {
		t.Run(tc.side, func(t *testing.T) {
			record := &entity.FeedingRecord{
				Detail: entity.FeedingDetail{
					"type":     "breast",
					"side":     tc.side,
					"duration": float64(600),
				},
			}

			lastFeedingTime := time.Now().Add(-2 * time.Hour)
			messageData := strategy.BuildMessageData(record, lastFeedingTime, 2.0)

			assert.Equal(t, tc.expected, messageData["phrase4"])
		})
	}
}

func TestBottleFeedingDifferentTypes(t *testing.T) {
	strategy := NewBottleFeedingReminderStrategy()

	testCases := []struct {
		bottleType string
		expected   string
	}{
		{"formula", "配方奶"},
		{"breast-milk", "母乳"},
	}

	for _, tc := range testCases {
		t.Run(tc.bottleType, func(t *testing.T) {
			record := &entity.FeedingRecord{
				Detail: entity.FeedingDetail{
					"type":       "bottle",
					"bottleType": tc.bottleType,
					"amount":     float64(120),
				},
			}

			lastFeedingTime := time.Now().Add(-3 * time.Hour)
			messageData := strategy.BuildMessageData(record, lastFeedingTime, 3.0)

			assert.Equal(t, tc.expected, messageData["phrase4"])
		})
	}
}
