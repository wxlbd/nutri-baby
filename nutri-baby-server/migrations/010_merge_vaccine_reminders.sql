-- 010_merge_vaccine_reminders.sql
-- 合并 vaccine_reminders 表的功能到 baby_vaccine_schedules 表
-- 目标: 消除 1:1 关系冗余,简化提醒逻辑,提升性能

-- =================================================================
-- 第1步: 添加提醒相关字段到 baby_vaccine_schedules 表
-- =================================================================

-- 添加提醒相关字段
ALTER TABLE baby_vaccine_schedules
ADD COLUMN IF NOT EXISTS scheduled_date BIGINT,           -- 计划接种日期(毫秒时间戳)
ADD COLUMN IF NOT EXISTS reminder_sent BOOLEAN DEFAULT false,  -- 是否已发送提醒
ADD COLUMN IF NOT EXISTS reminder_sent_at BIGINT;         -- 提醒发送时间(毫秒时间戳)

-- 创建索引以优化提醒查询
CREATE INDEX IF NOT EXISTS idx_vaccine_schedules_scheduled_date
ON baby_vaccine_schedules(baby_id, scheduled_date)
WHERE deleted_at IS NULL AND vaccination_status = 'pending';

-- =================================================================
-- 第2步: 从 vaccine_reminders 迁移数据
-- =================================================================

-- 如果 vaccine_reminders 表存在,迁移其数据
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vaccine_reminders') THEN
        -- 更新 baby_vaccine_schedules 的提醒字段
        UPDATE baby_vaccine_schedules s
        SET
            scheduled_date = r.scheduled_date,
            reminder_sent = r.reminder_sent,
            reminder_sent_at = r.sent_time
        FROM vaccine_reminders r
        WHERE s.schedule_id = r.plan_id
          AND r.deleted_at IS NULL
          AND s.deleted_at IS NULL;

        RAISE NOTICE '✓ 已从 vaccine_reminders 迁移提醒数据';
    ELSE
        RAISE NOTICE 'vaccine_reminders 表不存在,跳过数据迁移';
    END IF;
END $$;

-- =================================================================
-- 第3步: 为没有 scheduled_date 的记录计算并设置该字段
-- =================================================================

-- 根据宝宝出生日期 + age_in_months 计算 scheduled_date
UPDATE baby_vaccine_schedules s
SET scheduled_date = EXTRACT(EPOCH FROM (
    b.birth_date::date + (s.age_in_months || ' months')::interval
)) * 1000
FROM babies b
WHERE s.baby_id = b.baby_id
  AND s.scheduled_date IS NULL
  AND s.deleted_at IS NULL
  AND b.deleted_at IS NULL;

-- =================================================================
-- 第4步: 添加非空约束和注释
-- =================================================================

-- 设置 scheduled_date 为非空(所有记录都应该有计划日期)
ALTER TABLE baby_vaccine_schedules
ALTER COLUMN scheduled_date SET NOT NULL;

-- 添加字段注释
COMMENT ON COLUMN baby_vaccine_schedules.scheduled_date IS '计划接种日期(毫秒时间戳),基于出生日期+月龄计算';
COMMENT ON COLUMN baby_vaccine_schedules.reminder_sent IS '是否已发送微信订阅消息提醒';
COMMENT ON COLUMN baby_vaccine_schedules.reminder_sent_at IS '提醒发送时间(毫秒时间戳)';

-- =================================================================
-- 第5步: 备份 vaccine_reminders 表 (重命名,保留2-4周后删除)
-- =================================================================

DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vaccine_reminders') THEN
        ALTER TABLE vaccine_reminders RENAME TO vaccine_reminders_backup_20250115;
        RAISE NOTICE '✓ vaccine_reminders 表已备份为 vaccine_reminders_backup_20250115';
    END IF;
END $$;

-- =================================================================
-- 第6步: 数据验证
-- =================================================================

DO $$
DECLARE
    total_schedules INTEGER;
    schedules_with_date INTEGER;
    schedules_without_date INTEGER;
    old_reminders_count INTEGER := 0;
BEGIN
    -- 统计新表数据
    SELECT COUNT(*) INTO total_schedules
    FROM baby_vaccine_schedules
    WHERE deleted_at IS NULL;

    SELECT COUNT(*) INTO schedules_with_date
    FROM baby_vaccine_schedules
    WHERE deleted_at IS NULL AND scheduled_date IS NOT NULL;

    SELECT COUNT(*) INTO schedules_without_date
    FROM baby_vaccine_schedules
    WHERE deleted_at IS NULL AND scheduled_date IS NULL;

    -- 统计旧表数据(如果存在)
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vaccine_reminders_backup_20250115') THEN
        SELECT COUNT(*) INTO old_reminders_count
        FROM vaccine_reminders_backup_20250115
        WHERE deleted_at IS NULL;
    END IF;

    -- 输出统计结果
    RAISE NOTICE '========================================';
    RAISE NOTICE '数据迁移验证结果:';
    RAISE NOTICE 'baby_vaccine_schedules 总记录数: %', total_schedules;
    RAISE NOTICE '  ├─ 有 scheduled_date: %', schedules_with_date;
    RAISE NOTICE '  └─ 无 scheduled_date: %', schedules_without_date;

    IF old_reminders_count > 0 THEN
        RAISE NOTICE 'vaccine_reminders_backup 记录数: %', old_reminders_count;
    END IF;

    RAISE NOTICE '========================================';

    -- 数据一致性检查
    IF schedules_without_date > 0 THEN
        RAISE WARNING '警告: 有 % 条记录缺少 scheduled_date!', schedules_without_date;
    ELSE
        RAISE NOTICE '✓ 所有记录都有 scheduled_date';
    END IF;
END $$;

-- =================================================================
-- 第7步: 检查提醒状态计算逻辑
-- =================================================================

-- 创建视图用于查看提醒状态(实时计算)
CREATE OR REPLACE VIEW vaccine_schedule_reminders AS
SELECT
    schedule_id,
    baby_id,
    vaccine_name,
    dose_number,
    scheduled_date,
    vaccination_status,
    reminder_sent,
    reminder_sent_at,
    -- 实时计算提醒状态
    CASE
        WHEN vaccination_status <> 'pending' THEN 'completed'
        WHEN (scheduled_date - EXTRACT(EPOCH FROM NOW()) * 1000) / (24 * 60 * 60 * 1000) > 7 THEN 'upcoming'
        WHEN (scheduled_date - EXTRACT(EPOCH FROM NOW()) * 1000) / (24 * 60 * 60 * 1000) >= 0 THEN 'due'
        ELSE 'overdue'
    END as reminder_status,
    -- 实时计算距离应接种天数
    ((scheduled_date - EXTRACT(EPOCH FROM NOW()) * 1000) / (24 * 60 * 60 * 1000))::INTEGER as days_until_due
FROM baby_vaccine_schedules
WHERE deleted_at IS NULL;

COMMENT ON VIEW vaccine_schedule_reminders IS '疫苗提醒视图,实时计算提醒状态和距离应接种天数';

-- =================================================================
-- 迁移完成提示
-- =================================================================

DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE '数据库迁移 010_merge_vaccine_reminders 完成!';
    RAISE NOTICE '';
    RAISE NOTICE '变更说明:';
    RAISE NOTICE '1. 已将 vaccine_reminders 的功能合并到 baby_vaccine_schedules';
    RAISE NOTICE '2. 添加了 scheduled_date, reminder_sent, reminder_sent_at 字段';
    RAISE NOTICE '3. 提醒状态现在通过实时计算获得(不再存储)';
    RAISE NOTICE '4. 创建了 vaccine_schedule_reminders 视图用于查询提醒';
    RAISE NOTICE '';
    RAISE NOTICE '后续步骤:';
    RAISE NOTICE '1. 更新后端代码,移除 VaccineReminder 相关代码';
    RAISE NOTICE '2. 测试提醒功能是否正常';
    RAISE NOTICE '3. 确认无误后,2-4周后执行:';
    RAISE NOTICE '   DROP TABLE vaccine_reminders_backup_20250115;';
    RAISE NOTICE '========================================';
END $$;
