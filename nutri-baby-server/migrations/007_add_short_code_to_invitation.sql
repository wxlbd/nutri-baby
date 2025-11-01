-- 为 baby_invitations 表添加 short_code 字段
-- Migration: 007_add_short_code_to_invitation
-- Created: 2025-10-31
-- Description: 添加6位短码字段用于小程序码scene参数，避免32字符限制

-- 添加 short_code 字段
ALTER TABLE baby_invitations
ADD COLUMN short_code VARCHAR(10) DEFAULT '' NOT NULL;

-- 创建唯一索引
CREATE UNIQUE INDEX idx_baby_invitations_short_code ON baby_invitations(short_code) WHERE deleted_at IS NULL;

-- 为现有记录生成短码（如果有数据的话）
-- 注意：这里使用简单的序号方式，实际生产环境建议使用随机生成
UPDATE baby_invitations
SET short_code = LPAD(CAST(EXTRACT(EPOCH FROM CURRENT_TIMESTAMP)::BIGINT % 999999 AS TEXT), 6, '0')
WHERE short_code = '' OR short_code IS NULL;

-- 添加注释
COMMENT ON COLUMN baby_invitations.short_code IS '6位短码(用于小程序码scene参数,避免32字符限制)';
