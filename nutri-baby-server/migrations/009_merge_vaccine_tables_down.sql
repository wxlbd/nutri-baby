-- 009_merge_vaccine_tables_down.sql
-- 回滚脚本: 从 baby_vaccine_schedules 恢复到 baby_vaccine_plans + vaccine_records

-- =================================================================
-- 警告: 此脚本仅在迁移后发现问题时使用
-- 执行前请确保备份表 baby_vaccine_plans_backup_20250115 和
-- vaccine_records_backup_20250115 仍然存在
-- =================================================================

-- 第1步: 删除新表
DROP TABLE IF EXISTS baby_vaccine_schedules CASCADE;

-- 第2步: 恢复旧表
ALTER TABLE baby_vaccine_plans_backup_20250115 RENAME TO baby_vaccine_plans;
ALTER TABLE vaccine_records_backup_20250115 RENAME TO vaccine_records;

-- 第3步: 重建索引 (如果需要)
-- baby_vaccine_plans 索引
CREATE INDEX IF NOT EXISTS idx_baby_vaccine_plans_baby_id ON baby_vaccine_plans(baby_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_baby_vaccine_plans_template_id ON baby_vaccine_plans(template_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_baby_vaccine_plans_deleted_at ON baby_vaccine_plans(deleted_at);

-- vaccine_records 索引
CREATE INDEX IF NOT EXISTS idx_vaccine_records_baby_id ON vaccine_records(baby_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vaccine_records_plan_id ON vaccine_records(plan_id) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vaccine_records_vaccine_date ON vaccine_records(vaccine_date) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_vaccine_records_deleted_at ON vaccine_records(deleted_at);

-- 第4步: 验证回滚成功
DO $$
BEGIN
    RAISE NOTICE '========================================';
    RAISE NOTICE '回滚完成!';
    RAISE NOTICE '已恢复表: baby_vaccine_plans, vaccine_records';
    RAISE NOTICE '已删除表: baby_vaccine_schedules';
    RAISE NOTICE '========================================';
END $$;
