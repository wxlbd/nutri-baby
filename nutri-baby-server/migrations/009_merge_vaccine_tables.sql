-- 009_merge_vaccine_tables.sql
-- 合并 baby_vaccine_plans 和 vaccine_records 表为 baby_vaccine_schedules 表
-- 目标: 消除冗余字段,简化查询逻辑,提升性能

-- =================================================================
-- 第1步: 创建新表 baby_vaccine_schedules
-- =================================================================

CREATE TABLE IF NOT EXISTS baby_vaccine_schedules (
    -- 主键 (复用原 plan_id,保持与 vaccine_reminders 的外键关系)
    schedule_id VARCHAR(64) PRIMARY KEY,

    -- 宝宝和计划基础信息
    baby_id VARCHAR(64) NOT NULL,
    template_id VARCHAR(64),  -- 来源模板ID(可选,用于跟踪是否来自标准模板)

    -- 疫苗基本信息
    vaccine_type VARCHAR(32) NOT NULL,
    vaccine_name VARCHAR(64) NOT NULL,
    description TEXT,
    age_in_months INTEGER NOT NULL,
    dose_number INTEGER NOT NULL,
    is_required BOOLEAN DEFAULT true,
    reminder_days INTEGER DEFAULT 7,
    is_custom BOOLEAN DEFAULT false,

    -- 接种状态 (关键字段)
    vaccination_status VARCHAR(16) NOT NULL DEFAULT 'pending',
    -- 状态值: 'pending'(未接种), 'completed'(已完成), 'skipped'(跳过/不接种)

    -- 接种记录信息 (仅在 status='completed' 时有值)
    vaccine_date BIGINT,          -- 实际接种日期(毫秒时间戳)
    hospital VARCHAR(128),        -- 接种医院
    batch_number VARCHAR(64),     -- 疫苗批号
    doctor VARCHAR(64),           -- 接种医生
    reaction TEXT,                -- 不良反应
    note TEXT,                    -- 备注
    completed_by VARCHAR(64),     -- 记录接种的用户openid
    completed_by_name VARCHAR(64),      -- 记录者昵称(冗余,避免JOIN)
    completed_by_avatar VARCHAR(512),   -- 记录者头像(冗余,避免JOIN)
    completed_time BIGINT,        -- 接种记录创建时间(毫秒时间戳)

    -- 审计字段
    create_by VARCHAR(64) NOT NULL,
    create_time BIGINT NOT NULL,
    update_time BIGINT NOT NULL,
    deleted_at TIMESTAMP,

    -- 索引
    CONSTRAINT fk_baby FOREIGN KEY (baby_id) REFERENCES babies(baby_id) ON DELETE CASCADE,
    CONSTRAINT fk_template FOREIGN KEY (template_id) REFERENCES vaccine_plan_templates(template_id) ON DELETE SET NULL
);

-- 创建索引以优化查询性能
CREATE INDEX idx_vaccine_schedules_baby_id ON baby_vaccine_schedules(baby_id) WHERE deleted_at IS NULL;
CREATE INDEX idx_vaccine_schedules_baby_status ON baby_vaccine_schedules(baby_id, vaccination_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_vaccine_schedules_status ON baby_vaccine_schedules(vaccination_status) WHERE deleted_at IS NULL;
CREATE INDEX idx_vaccine_schedules_vaccine_date ON baby_vaccine_schedules(baby_id, vaccine_date DESC) WHERE deleted_at IS NULL;
CREATE INDEX idx_vaccine_schedules_deleted_at ON baby_vaccine_schedules(deleted_at);
CREATE INDEX idx_vaccine_schedules_template_id ON baby_vaccine_schedules(template_id) WHERE deleted_at IS NULL;

-- =================================================================
-- 第2步: 备份旧表 (重命名,保留2-4周后删除)
-- =================================================================

-- 检查旧表是否存在,如果存在则备份
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'baby_vaccine_plans') THEN
        ALTER TABLE baby_vaccine_plans RENAME TO baby_vaccine_plans_backup_20250115;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'vaccine_records') THEN
        ALTER TABLE vaccine_records RENAME TO vaccine_records_backup_20250115;
    END IF;
END $$;

-- =================================================================
-- 第3步: 数据迁移
-- =================================================================

-- 3.1 迁移未接种的计划 (status='pending')
-- 从 baby_vaccine_plans_backup 中找出没有接种记录的计划
INSERT INTO baby_vaccine_schedules (
    schedule_id, baby_id, template_id, vaccine_type, vaccine_name,
    description, age_in_months, dose_number, is_required, reminder_days,
    is_custom, vaccination_status,
    create_by, create_time, update_time, deleted_at
)
SELECT
    p.plan_id,
    p.baby_id,
    p.template_id,
    p.vaccine_type,
    p.vaccine_name,
    p.description,
    p.age_in_months,
    p.dose_number,
    p.is_required,
    p.reminder_days,
    p.is_custom,
    'pending',  -- 未接种状态
    p.create_by,
    p.create_time,
    p.update_time,
    p.deleted_at
FROM baby_vaccine_plans_backup_20250115 p
WHERE NOT EXISTS (
    SELECT 1 FROM vaccine_records_backup_20250115 r
    WHERE r.plan_id = p.plan_id AND r.deleted_at IS NULL
);

-- 3.2 迁移已接种的计划 (status='completed')
-- 合并 baby_vaccine_plans + vaccine_records
INSERT INTO baby_vaccine_schedules (
    schedule_id, baby_id, template_id, vaccine_type, vaccine_name,
    description, age_in_months, dose_number, is_required, reminder_days,
    is_custom, vaccination_status,
    vaccine_date, hospital, batch_number, doctor, reaction, note,
    completed_by, completed_by_name, completed_by_avatar, completed_time,
    create_by, create_time, update_time, deleted_at
)
SELECT
    p.plan_id,
    p.baby_id,
    p.template_id,
    p.vaccine_type,
    p.vaccine_name,
    p.description,
    p.age_in_months,
    p.dose_number,
    p.is_required,
    p.reminder_days,
    p.is_custom,
    'completed',  -- 已完成状态
    r.vaccine_date,
    r.hospital,
    r.batch_number,
    r.doctor,
    r.reaction,
    r.note,
    r.create_by,
    r.create_by_name,
    r.create_by_avatar,
    r.create_time,
    p.create_by,
    p.create_time,
    GREATEST(p.update_time, r.update_time),  -- 取最新更新时间
    COALESCE(p.deleted_at, r.deleted_at)     -- 任一删除则标记删除
FROM baby_vaccine_plans_backup_20250115 p
INNER JOIN vaccine_records_backup_20250115 r ON p.plan_id = r.plan_id
WHERE EXISTS (
    SELECT 1 FROM vaccine_records_backup_20250115 r2
    WHERE r2.plan_id = p.plan_id AND r2.deleted_at IS NULL
);

-- =================================================================
-- 第4步: 更新 vaccine_reminders 外键引用
-- =================================================================

-- 由于 schedule_id 复用了原 plan_id,外键关系自动维护,无需额外操作
-- 但为了明确,我们检查并确认外键约束

-- 检查 vaccine_reminders 表中是否有孤立的 plan_id
DO $$
DECLARE
    orphan_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO orphan_count
    FROM vaccine_reminders vr
    WHERE vr.deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM baby_vaccine_schedules s
          WHERE s.schedule_id = vr.plan_id AND s.deleted_at IS NULL
      );

    IF orphan_count > 0 THEN
        RAISE WARNING '发现 % 条孤立的疫苗提醒记录 (plan_id不存在于新表)', orphan_count;
    ELSE
        RAISE NOTICE '所有疫苗提醒记录的 plan_id 均存在于新表,外键关系完整';
    END IF;
END $$;

-- =================================================================
-- 第5步: 数据验证
-- =================================================================

-- 验证脚本: 确保数据迁移完整无误
DO $$
DECLARE
    old_plans_count INTEGER;
    old_records_count INTEGER;
    new_schedules_count INTEGER;
    new_pending_count INTEGER;
    new_completed_count INTEGER;
BEGIN
    -- 统计旧表数据
    SELECT COUNT(*) INTO old_plans_count
    FROM baby_vaccine_plans_backup_20250115
    WHERE deleted_at IS NULL;

    SELECT COUNT(*) INTO old_records_count
    FROM vaccine_records_backup_20250115
    WHERE deleted_at IS NULL;

    -- 统计新表数据
    SELECT COUNT(*) INTO new_schedules_count
    FROM baby_vaccine_schedules
    WHERE deleted_at IS NULL;

    SELECT COUNT(*) INTO new_pending_count
    FROM baby_vaccine_schedules
    WHERE deleted_at IS NULL AND vaccination_status = 'pending';

    SELECT COUNT(*) INTO new_completed_count
    FROM baby_vaccine_schedules
    WHERE deleted_at IS NULL AND vaccination_status = 'completed';

    -- 输出统计结果
    RAISE NOTICE '========================================';
    RAISE NOTICE '数据迁移验证结果:';
    RAISE NOTICE '旧表 baby_vaccine_plans 记录数: %', old_plans_count;
    RAISE NOTICE '旧表 vaccine_records 记录数: %', old_records_count;
    RAISE NOTICE '新表 baby_vaccine_schedules 总记录数: %', new_schedules_count;
    RAISE NOTICE '  ├─ pending (未接种): %', new_pending_count;
    RAISE NOTICE '  └─ completed (已完成): %', new_completed_count;
    RAISE NOTICE '========================================';

    -- 数据一致性检查
    IF new_schedules_count <> old_plans_count THEN
        RAISE WARNING '警告: 新表记录数 (%) 与旧表计划数 (%) 不一致!', new_schedules_count, old_plans_count;
    END IF;

    IF new_completed_count <> old_records_count THEN
        RAISE WARNING '警告: 新表已完成记录数 (%) 与旧表接种记录数 (%) 不一致!', new_completed_count, old_records_count;
    END IF;

    IF new_schedules_count = old_plans_count AND new_completed_count = old_records_count THEN
        RAISE NOTICE '✓ 数据迁移成功,记录数完全一致!';
    END IF;
END $$;

-- =================================================================
-- 第6步: 额外的完整性检查
-- =================================================================

-- 检查是否有疫苗接种记录没有对应的计划
DO $$
DECLARE
    orphan_records_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO orphan_records_count
    FROM vaccine_records_backup_20250115 r
    WHERE r.deleted_at IS NULL
      AND NOT EXISTS (
          SELECT 1 FROM baby_vaccine_plans_backup_20250115 p
          WHERE p.plan_id = r.plan_id
      );

    IF orphan_records_count > 0 THEN
        RAISE WARNING '发现 % 条孤立的接种记录 (plan_id不存在于计划表)', orphan_records_count;
        -- 可选: 将这些孤立记录也迁移过来,创建新的 schedule
    ELSE
        RAISE NOTICE '✓ 所有接种记录均有对应的计划';
    END IF;
END $$;

-- =================================================================
-- 迁移完成提示
-- =================================================================

DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE '数据库迁移 009_merge_vaccine_tables 完成!';
    RAISE NOTICE '';
    RAISE NOTICE '后续步骤:';
    RAISE NOTICE '1. 验证新表数据是否完整';
    RAISE NOTICE '2. 更新后端代码以使用新表';
    RAISE NOTICE '3. 测试所有疫苗相关功能';
    RAISE NOTICE '4. 确认无误后,2-4周后执行:';
    RAISE NOTICE '   DROP TABLE baby_vaccine_plans_backup_20250115;';
    RAISE NOTICE '   DROP TABLE vaccine_records_backup_20250115;';
    RAISE NOTICE '========================================';
END $$;
