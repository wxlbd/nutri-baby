package persistence

import (
	"context"
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"gorm.io/driver/sqlite"
	"gorm.io/gorm"

	"github.com/wxlbd/nutri-baby-server/internal/domain/entity"
)

// setupTestDB 设置测试数据库
func setupTestDB(t *testing.T) *gorm.DB {
	db, err := gorm.Open(sqlite.Open(":memory:"), &gorm.Config{})
	require.NoError(t, err)

	// 自动迁移表结构
	err = db.AutoMigrate(&entity.FeedingRecord{})
	require.NoError(t, err)

	return db
}

func TestFeedingRecordRepository_UpdateReminderStatus(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFeedingRecordRepository(db)
	ctx := context.Background()

	// 创建测试数据
	record := &entity.FeedingRecord{
		RecordID:    "test-record-1",
		BabyID:      "test-baby-1",
		FeedingType: entity.FeedingTypeBreast,
		Time:        time.Now().UnixMilli(),
		Detail: entity.FeedingDetail{
			"type":     "breast",
			"side":     "left",
			"duration": float64(600),
		},
		ReminderSent: false,
	}

	err := repo.Create(ctx, record)
	require.NoError(t, err)

	// 测试更新提醒状态
	reminderTime := time.Now().UnixMilli()
	err = repo.UpdateReminderStatus(ctx, record.RecordID, true, reminderTime)
	assert.NoError(t, err)

	// 验证更新结果
	updated, err := repo.FindByID(ctx, record.RecordID)
	require.NoError(t, err)
	assert.True(t, updated.ReminderSent)
	assert.NotNil(t, updated.ReminderTime)
	assert.Equal(t, reminderTime, *updated.ReminderTime)
}

func TestFeedingRecordRepository_FindByBabyID_WithReminderFilter(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFeedingRecordRepository(db)
	ctx := context.Background()

	now := time.Now()
	babyID := "test-baby-1"

	// 创建测试数据：一条已提醒，一条未提醒
	records := []*entity.FeedingRecord{
		{
			RecordID:     "record-1",
			BabyID:       babyID,
			FeedingType:  entity.FeedingTypeBreast,
			Time:         now.Add(-2 * time.Hour).UnixMilli(),
			Detail:       entity.FeedingDetail{"type": "breast"},
			ReminderSent: true,
		},
		{
			RecordID:     "record-2",
			BabyID:       babyID,
			FeedingType:  entity.FeedingTypeBreast,
			Time:         now.Add(-1 * time.Hour).UnixMilli(),
			Detail:       entity.FeedingDetail{"type": "breast"},
			ReminderSent: false,
		},
	}

	for _, r := range records {
		err := repo.Create(ctx, r)
		require.NoError(t, err)
	}

	// 测试：查询最近一条记录(page=1, pageSize=1)，应该过滤已提醒的记录
	startTime := now.Add(-24 * time.Hour).UnixMilli()
	endTime := now.UnixMilli()

	results, total, err := repo.FindByBabyID(ctx, babyID, startTime, endTime, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(1), total) // 只有一条未提醒的记录
	assert.Len(t, results, 1)
	assert.Equal(t, "record-2", results[0].RecordID)
	assert.False(t, results[0].ReminderSent)
}

func TestFeedingRecordRepository_PreventDuplicateReminder(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFeedingRecordRepository(db)
	ctx := context.Background()

	babyID := "test-baby-1"
	now := time.Now()

	// 创建一条喂养记录
	record := &entity.FeedingRecord{
		RecordID:    "test-record",
		BabyID:      babyID,
		FeedingType: entity.FeedingTypeBreast,
		Time:        now.Add(-3 * time.Hour).UnixMilli(),
		Detail: entity.FeedingDetail{
			"type": "breast",
			"side": "left",
		},
		ReminderSent: false,
	}

	err := repo.Create(ctx, record)
	require.NoError(t, err)

	// 第一次查询：应该能查到未提醒的记录
	startTime := now.Add(-24 * time.Hour).UnixMilli()
	endTime := now.UnixMilli()

	results1, _, err := repo.FindByBabyID(ctx, babyID, startTime, endTime, 1, 1)
	require.NoError(t, err)
	assert.Len(t, results1, 1)
	assert.Equal(t, record.RecordID, results1[0].RecordID)

	// 模拟发送提醒并更新状态
	reminderTime := now.UnixMilli()
	err = repo.UpdateReminderStatus(ctx, record.RecordID, true, reminderTime)
	require.NoError(t, err)

	// 第二次查询：不应该再查到该记录（已提醒）
	results2, total2, err := repo.FindByBabyID(ctx, babyID, startTime, endTime, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(0), total2) // 没有未提醒的记录了
	assert.Len(t, results2, 0)
}

func TestFeedingRecordRepository_MultipleRecords(t *testing.T) {
	db := setupTestDB(t)
	repo := NewFeedingRecordRepository(db)
	ctx := context.Background()

	babyID := "test-baby-1"
	now := time.Now()

	// 创建多条记录：3条未提醒，2条已提醒
	records := []*entity.FeedingRecord{
		{
			RecordID:     "record-1",
			BabyID:       babyID,
			FeedingType:  entity.FeedingTypeBreast,
			Time:         now.Add(-5 * time.Hour).UnixMilli(),
			Detail:       entity.FeedingDetail{"type": "breast"},
			ReminderSent: true,
		},
		{
			RecordID:     "record-2",
			BabyID:       babyID,
			FeedingType:  entity.FeedingTypeBottle,
			Time:         now.Add(-4 * time.Hour).UnixMilli(),
			Detail:       entity.FeedingDetail{"type": "bottle"},
			ReminderSent: false,
		},
		{
			RecordID:     "record-3",
			BabyID:       babyID,
			FeedingType:  entity.FeedingTypeBreast,
			Time:         now.Add(-3 * time.Hour).UnixMilli(),
			Detail:       entity.FeedingDetail{"type": "breast"},
			ReminderSent: false,
		},
		{
			RecordID:     "record-4",
			BabyID:       babyID,
			FeedingType:  entity.FeedingTypeFood,
			Time:         now.Add(-2 * time.Hour).UnixMilli(),
			Detail:       entity.FeedingDetail{"type": "food"},
			ReminderSent: true,
		},
		{
			RecordID:     "record-5",
			BabyID:       babyID,
			FeedingType:  entity.FeedingTypeBreast,
			Time:         now.Add(-1 * time.Hour).UnixMilli(),
			Detail:       entity.FeedingDetail{"type": "breast"},
			ReminderSent: false,
		},
	}

	for _, r := range records {
		err := repo.Create(ctx, r)
		require.NoError(t, err)
	}

	// 查询最近一条未提醒的记录
	startTime := now.Add(-24 * time.Hour).UnixMilli()
	endTime := now.UnixMilli()

	results, total, err := repo.FindByBabyID(ctx, babyID, startTime, endTime, 1, 1)
	require.NoError(t, err)
	assert.Equal(t, int64(3), total) // 3条未提醒的记录
	assert.Len(t, results, 1)
	assert.Equal(t, "record-5", results[0].RecordID) // 最近的未提醒记录
	assert.False(t, results[0].ReminderSent)
}
