-- 删除触发器
DROP TRIGGER IF EXISTS trigger_charging_requests_updated_at ON charging_requests;

-- 删除索引
DROP INDEX IF EXISTS idx_charging_requests_user_id;
DROP INDEX IF EXISTS idx_charging_requests_status;
DROP INDEX IF EXISTS idx_charging_requests_created_at;
DROP INDEX IF EXISTS idx_charging_requests_pile_id;
DROP INDEX IF EXISTS idx_charging_requests_status_mode_created_at;

-- 删除表
DROP TABLE IF EXISTS charging_requests;
