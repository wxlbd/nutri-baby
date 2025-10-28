-- ============================================
-- 喂养记录类型字段提取迁移
-- 文件: 006_feeding_type_field.sql
-- 日期: 2025-10-26
-- 说明: 将 detail 中的 feeding_type、amount、duration 提取为独立字段
--       便于查询、索引和统计
--       统一类型值: formula → bottle
-- ============================================

BEGIN;

-- 1. 添加新字段
ALTER TABLE feeding_records
ADD COLUMN IF NOT EXISTS feeding_type VARCHAR(16),
ADD COLUMN IF NOT EXISTS amount DECIMAL(10,2),
ADD COLUMN IF NOT EXISTS duration INTEGER;

-- 2. 从 detail 字段迁移数据并统一类型值
UPDATE feeding_records
SET
  feeding_type = CASE
    WHEN detail->>'feedingType' = 'formula' THEN 'bottle'
    WHEN detail->>'feedingType' = 'breast' THEN 'breast'
    WHEN detail->>'feedingType' = 'food' THEN 'food'
    WHEN detail->>'type' = 'formula' THEN 'bottle'
    WHEN detail->>'type' = 'breast' THEN 'breast'
    WHEN detail->>'type' = 'food' THEN 'food'
    ELSE COALESCE(detail->>'feedingType', detail->>'type', 'breast')
  END,
  amount = CASE
    WHEN detail->>'amount' IS NOT NULL AND detail->>'amount' != ''
    THEN CAST(detail->>'amount' AS DECIMAL)
    ELSE NULL
  END,
  duration = CASE
    WHEN detail->>'duration' IS NOT NULL AND detail->>'duration' != ''
    THEN CAST(detail->>'duration' AS INTEGER)
    ELSE NULL
  END
WHERE feeding_type IS NULL;

-- 3. 设置 NOT NULL 约束
ALTER TABLE feeding_records
ALTER COLUMN feeding_type SET NOT NULL;

-- 4. 添加检查约束，确保只能是有效的喂养类型
ALTER TABLE feeding_records
ADD CONSTRAINT IF NOT EXISTS chk_feeding_type
CHECK (feeding_type IN ('breast', 'bottle', 'food'));

-- 5. 创建复合索引（用于按类型查询和统计）
CREATE INDEX IF NOT EXISTS idx_feeding_records_type
ON feeding_records(baby_id, feeding_type, time DESC)
WHERE deleted_at IS NULL;

-- 6. 添加字段注释
COMMENT ON COLUMN feeding_records.feeding_type IS '喂养类型: breast/bottle/food';
COMMENT ON COLUMN feeding_records.amount IS '奶量(ml)，bottle类型时使用';
COMMENT ON COLUMN feeding_records.duration IS '时长(秒)，breast类型时使用';

COMMIT;

-- ============================================
-- 回滚脚本 (如需回滚，手动执行以下SQL)
-- ============================================

-- BEGIN;
-- DROP INDEX IF EXISTS idx_feeding_records_type;
-- ALTER TABLE feeding_records
-- DROP CONSTRAINT IF EXISTS chk_feeding_type,
-- DROP COLUMN IF EXISTS duration,
-- DROP COLUMN IF EXISTS amount,
-- DROP COLUMN IF EXISTS feeding_type;
-- COMMIT;

-- ============================================
-- 数据验证查询
-- ============================================

-- 验证数据迁移是否成功
-- SELECT
--     feeding_type,
--     amount,
--     duration,
--     detail->>'feedingType' AS old_feeding_type,
--     detail->>'amount' AS old_amount,
--     detail->>'duration' AS old_duration
-- FROM feeding_records
-- ORDER BY time DESC
-- LIMIT 10;

-- 查看各类型的统计
-- SELECT
--     feeding_type,
--     COUNT(*) AS count,
--     SUM(amount) AS total_amount,
--     SUM(duration) AS total_duration
-- FROM feeding_records
-- WHERE deleted_at IS NULL
-- GROUP BY feeding_type
-- ORDER BY feeding_type;
