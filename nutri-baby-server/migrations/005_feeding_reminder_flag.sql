-- ============================================
-- 喂养记录提醒标记字段迁移
-- 文件: 005_feeding_reminder_flag.sql
-- 日期: 2025-10-26
-- 说明: 为 feeding_records 表添加提醒标记字段，防止重复发送提醒
-- ============================================

BEGIN;

-- 1. 添加提醒标记字段
ALTER TABLE feeding_records
ADD COLUMN IF NOT EXISTS reminder_sent BOOLEAN DEFAULT FALSE,
ADD COLUMN IF NOT EXISTS reminder_time BIGINT;

-- 2. 为现有记录设置默认值
UPDATE feeding_records
SET reminder_sent = FALSE
WHERE reminder_sent IS NULL;

-- 3. 创建索引优化查询性能
-- 组合索引：用于定时任务查询未提醒的最近喂养记录
CREATE INDEX IF NOT EXISTS idx_feeding_records_reminder
ON feeding_records(baby_id, reminder_sent, time DESC)
WHERE deleted_at IS NULL;

-- 4. 添加字段注释
COMMENT ON COLUMN feeding_records.reminder_sent IS '是否已发送提醒';
COMMENT ON COLUMN feeding_records.reminder_time IS '提醒发送时间戳(毫秒)';

COMMIT;

-- ============================================
-- 回滚脚本 (如需回滚，手动执行以下SQL)
-- ============================================

-- BEGIN;
-- DROP INDEX IF EXISTS idx_feeding_records_reminder;
-- ALTER TABLE feeding_records
-- DROP COLUMN IF EXISTS reminder_time,
-- DROP COLUMN IF EXISTS reminder_sent;
-- COMMIT;
