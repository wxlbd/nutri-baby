-- ============================================
-- 喂养记录实际完成时间字段迁移
-- 文件: 008_feeding_actual_complete_time.sql
-- 日期: 2025-11-03
-- 说明: 为 feeding_records 表添加实际完成时间字段，用于准确计算提醒时间
--       解决用户延迟记录导致提醒时间偏晚的问题
-- ============================================

BEGIN;

-- 1. 添加实际完成时间字段
ALTER TABLE feeding_records
ADD COLUMN IF NOT EXISTS actual_complete_time BIGINT;

-- 2. 为现有记录设置默认值（使用 time 字段作为回退值）
-- 这保证了向后兼容性：旧记录依然使用 time 字段计算提醒
UPDATE feeding_records
SET actual_complete_time = time
WHERE actual_complete_time IS NULL AND time IS NOT NULL;

-- 3. 添加字段注释
COMMENT ON COLUMN feeding_records.actual_complete_time IS '实际喂养完成时间戳(毫秒)，用于计算下次提醒时间。如果为NULL则使用time字段';

-- 4. 创建索引优化基于完成时间的查询
CREATE INDEX IF NOT EXISTS idx_feeding_actual_complete_time
ON feeding_records(baby_id, actual_complete_time DESC)
WHERE deleted_at IS NULL AND actual_complete_time IS NOT NULL;

COMMIT;

-- ============================================
-- 回滚脚本 (如需回滚，手动执行以下SQL)
-- ============================================

-- BEGIN;
-- DROP INDEX IF EXISTS idx_feeding_actual_complete_time;
-- ALTER TABLE feeding_records
-- DROP COLUMN IF EXISTS actual_complete_time;
-- COMMIT;
