-- =====================================================
-- 订阅消息功能数据库迁移脚本
-- 版本: 003
-- 创建时间: 2025-10-24
-- =====================================================

-- 1. 创建订阅记录表
CREATE TABLE IF NOT EXISTS subscribe_records (
    id BIGSERIAL PRIMARY KEY,
    openid VARCHAR(64) NOT NULL,
    template_id VARCHAR(128) NOT NULL,
    template_type VARCHAR(32) NOT NULL,
    status VARCHAR(16) NOT NULL DEFAULT 'active',
    subscribe_time TIMESTAMP NOT NULL DEFAULT NOW(),
    expire_time TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    deleted_at TIMESTAMP,

    CONSTRAINT uq_openid_template UNIQUE(openid, template_id)
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_subscribe_records_openid ON subscribe_records(openid);
CREATE INDEX IF NOT EXISTS idx_subscribe_records_template_type ON subscribe_records(template_type);
CREATE INDEX IF NOT EXISTS idx_subscribe_records_status ON subscribe_records(status);
CREATE INDEX IF NOT EXISTS idx_subscribe_records_deleted_at ON subscribe_records(deleted_at);

-- 添加表注释
COMMENT ON TABLE subscribe_records IS '订阅消息授权记录表';
COMMENT ON COLUMN subscribe_records.openid IS '用户微信openid';
COMMENT ON COLUMN subscribe_records.template_id IS '微信模板ID';
COMMENT ON COLUMN subscribe_records.template_type IS '模板类型: vaccine_reminder, breast_feeding_reminder, bottle_feeding_reminder, pump_reminder, feeding_duration_alert';
COMMENT ON COLUMN subscribe_records.status IS '状态: active-有效, inactive-已取消, expired-已过期';
COMMENT ON COLUMN subscribe_records.subscribe_time IS '订阅授权时间';
COMMENT ON COLUMN subscribe_records.expire_time IS '过期时间(微信订阅消息通常30天有效期)';

-- 2. 创建消息发送记录表
CREATE TABLE IF NOT EXISTS message_send_logs (
    id BIGSERIAL PRIMARY KEY,
    openid VARCHAR(64) NOT NULL,
    template_id VARCHAR(128) NOT NULL,
    template_type VARCHAR(32) NOT NULL,
    data JSONB NOT NULL,
    page VARCHAR(256),
    miniprogram_state VARCHAR(32) DEFAULT 'formal',
    send_status VARCHAR(16) NOT NULL,
    errcode INTEGER,
    errmsg TEXT,
    send_time TIMESTAMP,
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_message_logs_openid ON message_send_logs(openid);
CREATE INDEX IF NOT EXISTS idx_message_logs_template_type ON message_send_logs(template_type);
CREATE INDEX IF NOT EXISTS idx_message_logs_send_status ON message_send_logs(send_status);
CREATE INDEX IF NOT EXISTS idx_message_logs_send_time ON message_send_logs(send_time);
CREATE INDEX IF NOT EXISTS idx_message_logs_created_at ON message_send_logs(created_at);

-- 添加表注释
COMMENT ON TABLE message_send_logs IS '订阅消息发送记录表';
COMMENT ON COLUMN message_send_logs.data IS '消息数据(JSON格式,包含模板字段)';
COMMENT ON COLUMN message_send_logs.page IS '点击消息跳转的小程序页面路径';
COMMENT ON COLUMN message_send_logs.miniprogram_state IS '小程序状态: developer-开发版, trial-体验版, formal-正式版';
COMMENT ON COLUMN message_send_logs.send_status IS '发送状态: success-成功, failed-失败, pending-待发送';
COMMENT ON COLUMN message_send_logs.errcode IS '微信API返回的错误码';
COMMENT ON COLUMN message_send_logs.errmsg IS '微信API返回的错误信息';
COMMENT ON COLUMN message_send_logs.send_time IS '实际发送时间';

-- 3. 创建消息发送队列表
CREATE TABLE IF NOT EXISTS message_send_queue (
    id BIGSERIAL PRIMARY KEY,
    openid VARCHAR(64) NOT NULL,
    template_id VARCHAR(128) NOT NULL,
    template_type VARCHAR(32) NOT NULL,
    data JSONB NOT NULL,
    page VARCHAR(256),
    scheduled_time TIMESTAMP NOT NULL,
    retry_count INTEGER NOT NULL DEFAULT 0,
    max_retry INTEGER NOT NULL DEFAULT 3,
    status VARCHAR(16) NOT NULL DEFAULT 'pending',
    error_msg TEXT,
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- 创建索引
CREATE INDEX IF NOT EXISTS idx_message_queue_openid ON message_send_queue(openid);
CREATE INDEX IF NOT EXISTS idx_message_queue_scheduled_time ON message_send_queue(scheduled_time);
CREATE INDEX IF NOT EXISTS idx_message_queue_status ON message_send_queue(status);
CREATE INDEX IF NOT EXISTS idx_message_queue_created_at ON message_send_queue(created_at);

-- 添加表注释
COMMENT ON TABLE message_send_queue IS '订阅消息发送队列表(用于异步处理和重试)';
COMMENT ON COLUMN message_send_queue.scheduled_time IS '计划发送时间';
COMMENT ON COLUMN message_send_queue.retry_count IS '已重试次数';
COMMENT ON COLUMN message_send_queue.max_retry IS '最大重试次数';
COMMENT ON COLUMN message_send_queue.status IS '状态: pending-待发送, processing-处理中, sent-已发送, failed-失败';
COMMENT ON COLUMN message_send_queue.error_msg IS '错误信息(失败时记录)';

-- 插入初始数据(可选)
-- INSERT INTO ... 如果需要插入初始配置数据

-- 完成提示
SELECT 'Subscribe message tables created successfully!' AS status;
