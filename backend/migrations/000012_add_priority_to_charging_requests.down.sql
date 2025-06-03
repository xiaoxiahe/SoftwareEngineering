-- 移除为充电请求表添加的优先级字段
DROP INDEX IF EXISTS idx_charging_requests_mode_status_priority;
DROP INDEX IF EXISTS idx_charging_requests_priority_created_at;
DROP INDEX IF EXISTS idx_charging_requests_priority;
ALTER TABLE charging_requests DROP COLUMN IF EXISTS priority;
