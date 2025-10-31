-- 006_feeding_reminder_interval.sql
-- 添加喂养提醒间隔和下次提醒时间字段

ALTER TABLE feeding_records
ADD COLUMN reminder_interval INT,
ADD COLUMN next_reminder_time BIGINT;

-- 为下次提醒时间字段添加索引,用于定时扫描
CREATE INDEX idx_feeding_next_reminder ON feeding_records(next_reminder_time)
WHERE next_reminder_time IS NOT NULL AND reminder_sent = false;
