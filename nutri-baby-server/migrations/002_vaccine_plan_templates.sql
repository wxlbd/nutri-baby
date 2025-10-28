-- 002_vaccine_plan_templates.sql
-- 初始化国家免疫规划疫苗计划模板数据

-- 清空现有模板数据（如果存在）
TRUNCATE TABLE vaccine_plan_templates RESTART IDENTITY CASCADE;

-- 插入国家免疫规划疫苗模板（24条）
-- 注意: create_time 和 update_time 使用毫秒时间戳(BIGINT)
INSERT INTO vaccine_plan_templates (template_id, vaccine_type, vaccine_name, description, age_in_months, dose_number, is_required, reminder_days, sort_order, create_time, update_time)
VALUES
    -- 乙肝疫苗 (HepB)
    (gen_random_uuid()::text, 'HepB', '乙肝疫苗', '出生24小时内接种', 0, 1, true, 3, 1, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'HepB', '乙肝疫苗', '满1个月接种', 1, 2, true, 7, 2, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'HepB', '乙肝疫苗', '满6个月接种', 6, 3, true, 7, 3, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),

    -- 卡介苗 (BCG)
    (gen_random_uuid()::text, 'BCG', '卡介苗', '出生后尽快接种', 0, 1, true, 3, 4, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),

    -- 脊灰疫苗 (OPV)
    (gen_random_uuid()::text, 'OPV', '脊灰疫苗', '满2个月接种', 2, 1, true, 7, 5, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'OPV', '脊灰疫苗', '满3个月接种', 3, 2, true, 7, 6, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'OPV', '脊灰疫苗', '满4个月接种', 4, 3, true, 7, 7, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'OPV', '脊灰疫苗', '满18个月接种', 18, 4, true, 7, 8, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),

    -- 百白破疫苗 (DTaP)
    (gen_random_uuid()::text, 'DTaP', '百白破疫苗', '满3个月接种', 3, 1, true, 7, 9, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'DTaP', '百白破疫苗', '满4个月接种', 4, 2, true, 7, 10, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'DTaP', '百白破疫苗', '满5个月接种', 5, 3, true, 7, 11, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'DTaP', '百白破疫苗', '满18个月接种', 18, 4, true, 7, 12, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),

    -- 麻风疫苗 (MR)
    (gen_random_uuid()::text, 'MR', '麻风疫苗', '满8个月接种', 8, 1, true, 7, 13, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),

    -- 麻腮风疫苗 (MMR)
    (gen_random_uuid()::text, 'MMR', '麻腮风疫苗', '满18个月接种', 18, 1, true, 7, 14, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),

    -- 乙脑疫苗 (JE)
    (gen_random_uuid()::text, 'JE', '乙脑疫苗', '满8个月接种', 8, 1, true, 7, 15, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'JE', '乙脑疫苗', '满2岁接种', 24, 2, true, 7, 16, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),

    -- 流脑AC疫苗 (MeningAC)
    (gen_random_uuid()::text, 'MeningAC', '流脑AC疫苗', '满6个月接种', 6, 1, true, 7, 17, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'MeningAC', '流脑AC疫苗', '满9个月接种', 9, 2, true, 7, 18, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'MeningAC', '流脑AC疫苗', '满3岁接种', 36, 3, true, 7, 19, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),

    -- 甲肝疫苗 (HepA)
    (gen_random_uuid()::text, 'HepA', '甲肝疫苗', '满18个月接种', 18, 1, true, 7, 20, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),

    -- 常见自费疫苗（非必打，供用户参考）
    -- 肺炎疫苗 (PCV)
    (gen_random_uuid()::text, 'PCV', '肺炎13价疫苗', '满2个月接种（自费）', 2, 1, false, 7, 21, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'PCV', '肺炎13价疫苗', '满4个月接种（自费）', 4, 2, false, 7, 22, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),
    (gen_random_uuid()::text, 'PCV', '肺炎13价疫苗', '满6个月接种（自费）', 6, 3, false, 7, 23, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000),

    -- 轮状病毒疫苗 (Rota)
    (gen_random_uuid()::text, 'Rota', '轮状病毒疫苗', '满2个月接种（自费）', 2, 1, false, 7, 24, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000, EXTRACT(EPOCH FROM CURRENT_TIMESTAMP) * 1000);

-- 验证插入数据
SELECT COUNT(*) as total_templates FROM vaccine_plan_templates;

-- 按疫苗类型统计
SELECT vaccine_type, COUNT(*) as count,
       SUM(CASE WHEN is_required THEN 1 ELSE 0 END) as required_count
FROM vaccine_plan_templates
GROUP BY vaccine_type
ORDER BY MIN(sort_order);
