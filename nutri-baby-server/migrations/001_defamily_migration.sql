-- ====================================================
-- 去家庭化架构数据迁移脚本
-- 从 Family-Based 架构迁移到 Baby-Centered 架构
-- ====================================================

-- ====================================================
-- 阶段 1: 创建新表和字段
-- ====================================================

-- 1. 创建 baby_collaborators 表 (替代 family_members)
CREATE TABLE IF NOT EXISTS baby_collaborators (
    id BIGSERIAL PRIMARY KEY,
    baby_id VARCHAR(64) NOT NULL,
    openid VARCHAR(64) NOT NULL,
    role VARCHAR(16) NOT NULL,  -- admin, editor, viewer
    access_type VARCHAR(16) NOT NULL DEFAULT 'permanent',  -- permanent, temporary
    expires_at BIGINT,  -- 临时权限过期时间(毫秒时间戳)
    join_time BIGINT NOT NULL,
    update_time BIGINT NOT NULL,
    deleted_at TIMESTAMP,

    UNIQUE(baby_id, openid),
    CONSTRAINT chk_role CHECK (role IN ('admin', 'editor', 'viewer')),
    CONSTRAINT chk_access_type CHECK (access_type IN ('permanent', 'temporary'))
);

-- 创建索引
CREATE INDEX idx_baby_collaborators_baby_id ON baby_collaborators(baby_id);
CREATE INDEX idx_baby_collaborators_openid ON baby_collaborators(openid);
CREATE INDEX idx_baby_collaborators_deleted_at ON baby_collaborators(deleted_at);
CREATE UNIQUE INDEX idx_baby_collaborators_baby_user ON baby_collaborators(baby_id, openid) WHERE deleted_at IS NULL;

COMMENT ON TABLE baby_collaborators IS '宝宝协作者表(替代 family_members)';
COMMENT ON COLUMN baby_collaborators.baby_id IS '宝宝ID';
COMMENT ON COLUMN baby_collaborators.openid IS '用户openid';
COMMENT ON COLUMN baby_collaborators.role IS '角色:admin/editor/viewer';
COMMENT ON COLUMN baby_collaborators.access_type IS '访问类型:permanent/temporary';
COMMENT ON COLUMN baby_collaborators.expires_at IS '临时权限过期时间';

-- ====================================================

-- 2. 修改 babies 表
ALTER TABLE babies ADD COLUMN IF NOT EXISTS creator_id VARCHAR(64);
ALTER TABLE babies ADD COLUMN IF NOT EXISTS family_group VARCHAR(64);

-- 为新字段添加索引
CREATE INDEX IF NOT EXISTS idx_babies_creator_id ON babies(creator_id);
CREATE INDEX IF NOT EXISTS idx_babies_family_group ON babies(family_group);

COMMENT ON COLUMN babies.creator_id IS '宝宝创建者openid';
COMMENT ON COLUMN babies.family_group IS '可选的家庭分组名称';

-- ====================================================

-- 3. 为所有记录表添加创建者信息冗余字段
ALTER TABLE feeding_records ADD COLUMN IF NOT EXISTS create_by_name VARCHAR(64);
ALTER TABLE feeding_records ADD COLUMN IF NOT EXISTS create_by_avatar VARCHAR(512);

ALTER TABLE sleep_records ADD COLUMN IF NOT EXISTS create_by_name VARCHAR(64);
ALTER TABLE sleep_records ADD COLUMN IF NOT EXISTS create_by_avatar VARCHAR(512);

ALTER TABLE diaper_records ADD COLUMN IF NOT EXISTS create_by_name VARCHAR(64);
ALTER TABLE diaper_records ADD COLUMN IF NOT EXISTS create_by_avatar VARCHAR(512);

ALTER TABLE growth_records ADD COLUMN IF NOT EXISTS create_by_name VARCHAR(64);
ALTER TABLE growth_records ADD COLUMN IF NOT EXISTS create_by_avatar VARCHAR(512);

ALTER TABLE vaccine_records ADD COLUMN IF NOT EXISTS create_by_name VARCHAR(64);
ALTER TABLE vaccine_records ADD COLUMN IF NOT EXISTS create_by_avatar VARCHAR(512);

-- ====================================================
-- 阶段 2: 数据迁移
-- ====================================================

-- 1. 迁移 family_members → baby_collaborators
-- 为每个家庭的每个宝宝创建协作者记录
INSERT INTO baby_collaborators (baby_id, openid, role, access_type, join_time, update_time)
SELECT
    b.baby_id,
    fm.openid,
    fm.role,
    'permanent' as access_type,
    fm.join_time,
    EXTRACT(EPOCH FROM NOW()) * 1000 as update_time
FROM family_members fm
JOIN babies b ON b.family_id = fm.family_id
WHERE fm.deleted_at IS NULL
  AND b.deleted_at IS NULL
ON CONFLICT (baby_id, openid) DO NOTHING;

-- ====================================================

-- 2. 填充 babies.creator_id
-- 使用家庭的 creator_id 作为宝宝的 creator_id
UPDATE babies b
SET creator_id = f.creator_id
FROM families f
WHERE b.family_id = f.family_id
  AND b.creator_id IS NULL;

-- ====================================================

-- 3. 填充 babies.family_group
-- 使用 family_name 作为 family_group
UPDATE babies b
SET family_group = f.family_name
FROM families f
WHERE b.family_id = f.family_id
  AND b.family_group IS NULL;

-- ====================================================

-- 4. 填充记录表的创建者信息冗余
-- feeding_records
UPDATE feeding_records fr
SET
    create_by_name = u.nick_name,
    create_by_avatar = u.avatar_url
FROM users u
WHERE fr.create_by = u.openid
  AND fr.create_by_name IS NULL;

-- sleep_records
UPDATE sleep_records sr
SET
    create_by_name = u.nick_name,
    create_by_avatar = u.avatar_url
FROM users u
WHERE sr.create_by = u.openid
  AND sr.create_by_name IS NULL;

-- diaper_records
UPDATE diaper_records dr
SET
    create_by_name = u.nick_name,
    create_by_avatar = u.avatar_url
FROM users u
WHERE dr.create_by = u.openid
  AND dr.create_by_name IS NULL;

-- growth_records
UPDATE growth_records gr
SET
    create_by_name = u.nick_name,
    create_by_avatar = u.avatar_url
FROM users u
WHERE gr.create_by = u.openid
  AND gr.create_by_name IS NULL;

-- vaccine_records
UPDATE vaccine_records vr
SET
    create_by_name = u.nick_name,
    create_by_avatar = u.avatar_url
FROM users u
WHERE vr.create_by = u.openid
  AND vr.create_by_name IS NULL;

-- ====================================================
-- 阶段 3: 数据验证
-- ====================================================

-- 验证 baby_collaborators 数据
SELECT
    '验证: baby_collaborators 记录数' as check_point,
    COUNT(*) as count
FROM baby_collaborators;

-- 验证每个宝宝至少有一个协作者
SELECT
    '验证: 每个宝宝至少有一个协作者' as check_point,
    COUNT(*) as babies_without_collaborators
FROM babies b
LEFT JOIN baby_collaborators bc ON b.baby_id = bc.baby_id AND bc.deleted_at IS NULL
WHERE b.deleted_at IS NULL
  AND bc.id IS NULL;

-- 验证 babies.creator_id 已填充
SELECT
    '验证: babies.creator_id 已填充' as check_point,
    COUNT(*) as babies_without_creator
FROM babies
WHERE deleted_at IS NULL
  AND creator_id IS NULL;

-- 验证记录表创建者信息已填充
SELECT
    '验证: feeding_records 创建者信息' as check_point,
    COUNT(*) as records_without_creator_name
FROM feeding_records
WHERE deleted_at IS NULL
  AND create_by_name IS NULL;

-- ====================================================
-- 阶段 4: (可选) 备份和清理旧表
-- ====================================================

-- 备份旧表(重命名)
-- ALTER TABLE families RENAME TO families_backup_20250101;
-- ALTER TABLE family_members RENAME TO family_members_backup_20250101;

-- 如果确认数据迁移成功且不需要回滚,可以删除旧表
-- DROP TABLE IF EXISTS families_backup_20250101;
-- DROP TABLE IF EXISTS family_members_backup_20250101;

-- ====================================================
-- 完成迁移
-- ====================================================

-- 显示迁移统计
SELECT
    'Migration Summary' as summary,
    (SELECT COUNT(*) FROM baby_collaborators) as total_collaborators,
    (SELECT COUNT(*) FROM babies WHERE creator_id IS NOT NULL) as babies_with_creator,
    (SELECT COUNT(*) FROM babies WHERE family_group IS NOT NULL) as babies_with_group,
    (SELECT COUNT(*) FROM feeding_records WHERE create_by_name IS NOT NULL) as feeding_records_enriched;

-- ====================================================
-- 回滚脚本 (如果需要)
-- ====================================================

/*
-- 回滚步骤 1: 删除新增的表和字段
DROP TABLE IF EXISTS baby_collaborators;

ALTER TABLE babies DROP COLUMN IF EXISTS creator_id;
ALTER TABLE babies DROP COLUMN IF EXISTS family_group;

ALTER TABLE feeding_records DROP COLUMN IF EXISTS create_by_name;
ALTER TABLE feeding_records DROP COLUMN IF EXISTS create_by_avatar;

ALTER TABLE sleep_records DROP COLUMN IF EXISTS create_by_name;
ALTER TABLE sleep_records DROP COLUMN IF EXISTS create_by_avatar;

ALTER TABLE diaper_records DROP COLUMN IF EXISTS create_by_name;
ALTER TABLE diaper_records DROP COLUMN IF EXISTS create_by_avatar;

ALTER TABLE growth_records DROP COLUMN IF EXISTS create_by_name;
ALTER TABLE growth_records DROP COLUMN IF EXISTS create_by_avatar;

ALTER TABLE vaccine_records DROP COLUMN IF EXISTS create_by_name;
ALTER TABLE vaccine_records DROP COLUMN IF EXISTS create_by_avatar;

-- 回滚步骤 2: 恢复备份的表
-- ALTER TABLE families_backup_20250101 RENAME TO families;
-- ALTER TABLE family_members_backup_20250101 RENAME TO family_members;
*/
