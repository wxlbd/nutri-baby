-- =====================================================
-- 订阅消息诊断脚本
-- 用途：快速检查订阅消息系统的状态
-- 使用方法：psql -h localhost -U wxl -d nutri_baby -f diagnose_subscribe.sql
-- =====================================================

\echo '========================================='
\echo '  订阅消息系统诊断报告'
\echo '========================================='
\echo ''

-- 1. 检查表是否存在
\echo '>>> 1. 检查数据库表是否存在'
SELECT
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'subscribe_records')
        THEN '✓ subscribe_records 表存在'
        ELSE '✗ subscribe_records 表不存在'
    END AS check_result
UNION ALL
SELECT
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'message_send_queue')
        THEN '✓ message_send_queue 表存在'
        ELSE '✗ message_send_queue 表不存在'
    END
UNION ALL
SELECT
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'message_send_logs')
        THEN '✓ message_send_logs 表存在'
        ELSE '✗ message_send_logs 表不存在'
    END;

\echo ''

-- 2. 检查订阅记录
\echo '>>> 2. 订阅记录统计'
SELECT
    template_type AS "消息类型",
    COUNT(*) AS "总记录数",
    SUM(CASE WHEN status = 'active' THEN 1 ELSE 0 END) AS "有效订阅",
    SUM(CASE WHEN status = 'inactive' THEN 1 ELSE 0 END) AS "已取消",
    SUM(CASE WHEN status = 'expired' THEN 1 ELSE 0 END) AS "已过期"
FROM subscribe_records
WHERE deleted_at IS NULL
GROUP BY template_type
ORDER BY template_type;

\echo ''

-- 3. 检查是否有空的 template_id
\echo '>>> 3. 检查 template_id 字段完整性'
SELECT
    COUNT(*) AS "空 template_id 记录数"
FROM subscribe_records
WHERE template_id IS NULL OR template_id = ''
  AND deleted_at IS NULL;

\echo ''

-- 4. 显示最近的订阅记录
\echo '>>> 4. 最近5条订阅记录'
SELECT
    openid,
    template_id AS tid,
    template_type AS type,
    status,
    subscribe_time AS sub_time,
    expire_time AS exp_time
FROM subscribe_records
WHERE deleted_at IS NULL
ORDER BY created_at DESC
LIMIT 5;

\echo ''

-- 5. 检查消息队列状态
\echo '>>> 5. 消息队列状态统计'
SELECT
    status AS "队列状态",
    COUNT(*) AS "消息数量",
    MIN(scheduled_time) AS "最早计划时间",
    MAX(scheduled_time) AS "最晚计划时间"
FROM message_send_queue
GROUP BY status
ORDER BY status;

\echo ''

-- 6. 检查待发送消息
\echo '>>> 6. 待发送消息（应该被定时任务处理）'
SELECT
    id,
    openid,
    template_type AS type,
    scheduled_time AS sche_time,
    retry_count AS retry,
    status,
    EXTRACT(EPOCH FROM (NOW() - scheduled_time))/60 AS "延迟分钟数"
FROM message_send_queue
WHERE status = 'pending'
  AND scheduled_time <= NOW()
ORDER BY scheduled_time ASC
LIMIT 10;

\echo ''

-- 7. 检查发送日志
\echo '>>> 7. 消息发送日志统计（最近24小时）'
SELECT
    send_status AS "发送状态",
    COUNT(*) AS "消息数量",
    COUNT(DISTINCT openid) AS "涉及用户数"
FROM message_send_logs
WHERE created_at >= NOW() - INTERVAL '24 hours'
GROUP BY send_status
ORDER BY send_status;

\echo ''

-- 8. 显示最近的发送日志
\echo '>>> 8. 最近10条发送日志'
SELECT
    openid,
    template_type AS type,
    send_status AS status,
    errcode,
    LEFT(errmsg, 30) AS error_msg,
    send_time
FROM message_send_logs
ORDER BY created_at DESC
LIMIT 10;

\echo ''

-- 9. 检查失败的消息
\echo '>>> 9. 失败的消息详情（最近10条）'
SELECT
    id,
    openid,
    template_type AS type,
    errcode,
    errmsg AS error_message,
    created_at
FROM message_send_logs
WHERE send_status = 'failed'
ORDER BY created_at DESC
LIMIT 10;

\echo ''

-- 10. 检查用户的完整订阅情况
\echo '>>> 10. 特定用户的订阅和消息情况（需要指定 openid）'
\echo '示例：SELECT * FROM user_subscribe_summary WHERE openid = '\''om8hB12mqHOp1BiTf3KZ_ew8eWH4'\'';'

-- 创建临时视图方便查询
CREATE OR REPLACE TEMP VIEW user_subscribe_summary AS
SELECT
    sr.openid,
    sr.template_type,
    sr.template_id,
    sr.status AS subscribe_status,
    sr.subscribe_time,
    sr.expire_time,
    (SELECT COUNT(*) FROM message_send_queue WHERE openid = sr.openid AND template_type = sr.template_type) AS queued_messages,
    (SELECT COUNT(*) FROM message_send_logs WHERE openid = sr.openid AND template_type = sr.template_type) AS sent_messages,
    (SELECT COUNT(*) FROM message_send_logs WHERE openid = sr.openid AND template_type = sr.template_type AND send_status = 'success') AS success_count,
    (SELECT COUNT(*) FROM message_send_logs WHERE openid = sr.openid AND template_type = sr.template_type AND send_status = 'failed') AS failed_count
FROM subscribe_records sr
WHERE sr.deleted_at IS NULL
ORDER BY sr.created_at DESC;

\echo ''
\echo '========================================='
\echo '  诊断完成！'
\echo '========================================='
\echo ''
\echo '使用提示：'
\echo '1. 查看特定用户订阅情况：'
\echo '   SELECT * FROM user_subscribe_summary WHERE openid = '\''你的openid'\'';'
\echo ''
\echo '2. 手动触发消息（测试）：'
\echo '   INSERT INTO message_send_queue (openid, template_id, template_type, data, page, scheduled_time, status)'
\echo '   VALUES ('\''你的openid'\'', '\''模板ID'\'', '\''breast_feeding_reminder'\'', '\''{"test":"data"}'\'', '\''pages/index/index'\'', NOW(), '\''pending'\'');'
\echo ''
