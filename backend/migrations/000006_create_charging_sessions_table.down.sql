-- 删除触发器
DROP TRIGGER IF EXISTS trigger_after_charging_session_insert ON charging_sessions;
DROP TRIGGER IF EXISTS trigger_after_charging_session_update ON charging_sessions;

-- 删除函数
DROP FUNCTION IF EXISTS update_charging_pile_stats();

-- 删除索引
DROP INDEX IF EXISTS idx_charging_sessions_user_id;
DROP INDEX IF EXISTS idx_charging_sessions_pile_id;
DROP INDEX IF EXISTS idx_charging_sessions_request_id;
DROP INDEX IF EXISTS idx_charging_sessions_start_time;
DROP INDEX IF EXISTS idx_charging_sessions_status;
DROP INDEX IF EXISTS idx_charging_sessions_pile_start_status;

-- 删除表
DROP TABLE IF EXISTS charging_sessions;
