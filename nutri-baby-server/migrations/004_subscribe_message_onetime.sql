-- ============================================
-- 订阅消息表结构迁移(一次性订阅消息机制)
-- 文件: 004_subscribe_message_onetime.sql
-- 日期: 2025-01-XX
-- ============================================

BEGIN;

-- 1. 删除旧的订阅记录表(如果存在)
DROP TABLE IF EXISTS subscribe_records CASCADE;

-- 2. 创建一次性订阅授权记录表
CREATE TABLE IF NOT EXISTS subscribe_records (
    id SERIAL PRIMARY KEY,
    openid VARCHAR(64) NOT NULL,
    template_id VARCHAR(128) NOT NULL,
    template_type VARCHAR(32) NOT NULL,

    -- 状态: available(可用), used(已使用), expired(已过期)
    status VARCHAR(16) NOT NULL DEFAULT 'available',

    -- 时间字段
    authorize_time TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP, -- 授权时间
    used_time TIMESTAMP, -- 使用时间
    expire_time TIMESTAMP, -- 过期时间(授权后7天)

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    deleted_at TIMESTAMP,

    -- 索引
    CONSTRAINT idx_openid_type UNIQUE (openid, template_type, authorize_time)
);

-- 创建索引
CREATE INDEX idx_subscribe_records_openid ON subscribe_records(openid);
CREATE INDEX idx_subscribe_records_template_type ON subscribe_records(template_type);
CREATE INDEX idx_subscribe_records_status ON subscribe_records(status);
CREATE INDEX idx_subscribe_records_expire_time ON subscribe_records(expire_time);
CREATE INDEX idx_subscribe_records_deleted_at ON subscribe_records(deleted_at);

-- 为常见查询创建组合索引
CREATE INDEX idx_subscribe_records_available ON subscribe_records(openid, template_type, status, expire_time)
WHERE status = 'available' AND deleted_at IS NULL;

-- 3. 创建消息发送日志表(如果不存在)
CREATE TABLE IF NOT EXISTS message_send_logs (
    id SERIAL PRIMARY KEY,
    openid VARCHAR(64) NOT NULL,
    template_id VARCHAR(128) NOT NULL,
    template_type VARCHAR(32) NOT NULL,
    data JSONB NOT NULL, -- 消息数据
    page VARCHAR(256), -- 小程序页面路径
    miniprogram_state VARCHAR(32) DEFAULT 'formal', -- 小程序状态: developer/trial/formal

    send_status VARCHAR(16) NOT NULL, -- 发送状态: success/failed/pending
    errcode INTEGER, -- 微信错误码
    err_msg TEXT, -- 错误信息
    send_time TIMESTAMP, -- 实际发送时间

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- 索引
    INDEX idx_message_logs_openid (openid),
    INDEX idx_message_logs_template_type (template_type),
    INDEX idx_message_logs_send_status (send_status),
    INDEX idx_message_logs_send_time (send_time),
    INDEX idx_message_logs_created_at (created_at)
);

-- 4. 创建消息发送队列表(暂不使用,保留供将来扩展)
CREATE TABLE IF NOT EXISTS message_send_queue (
    id SERIAL PRIMARY KEY,
    openid VARCHAR(64) NOT NULL,
    template_id VARCHAR(128) NOT NULL,
    template_type VARCHAR(32) NOT NULL,
    data JSONB NOT NULL,
    page VARCHAR(256),

    scheduled_time TIMESTAMP NOT NULL, -- 计划发送时间
    retry_count INTEGER NOT NULL DEFAULT 0, -- 重试次数
    max_retry INTEGER NOT NULL DEFAULT 3, -- 最大重试次数

    status VARCHAR(16) NOT NULL DEFAULT 'pending', -- pending/processing/sent/failed
    error_msg TEXT, -- 错误信息

    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,

    -- 索引
    INDEX idx_queue_status (status),
    INDEX idx_queue_scheduled_time (scheduled_time)
);

-- 5. 添加表注释
COMMENT ON TABLE subscribe_records IS '一次性订阅消息授权记录表';
COMMENT ON COLUMN subscribe_records.status IS '状态: available(可用), used(已使用), expired(已过期)';
COMMENT ON COLUMN subscribe_records.authorize_time IS '用户授权时间';
COMMENT ON COLUMN subscribe_records.used_time IS '消息发送时间(授权消耗时间)';
COMMENT ON COLUMN subscribe_records.expire_time IS '授权过期时间(授权后7天)';

COMMENT ON TABLE message_send_logs IS '订阅消息发送日志表';
COMMENT ON COLUMN message_send_logs.send_status IS '发送状态: success(成功), failed(失败), pending(待发送)';

COMMENT ON TABLE message_send_queue IS '消息发送队列表(暂不使用)';

COMMIT;

-- ============================================
-- 数据清理定时任务SQL(可选,用于定期清理过期数据)
-- ============================================

-- 清理7天前过期的授权记录
-- DELETE FROM subscribe_records
-- WHERE expire_time < NOW() - INTERVAL '7 days'
-- OR (status = 'used' AND used_time < NOW() - INTERVAL '7 days');

-- 清理30天前的发送日志
-- DELETE FROM message_send_logs
-- WHERE created_at < NOW() - INTERVAL '30 days';

-- 清理7天前的队列记录
-- DELETE FROM message_send_queue
-- WHERE created_at < NOW() - INTERVAL '7 days'
-- AND status IN ('sent', 'failed');
