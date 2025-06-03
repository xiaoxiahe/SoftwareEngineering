-- 删除函数
DROP FUNCTION IF EXISTS clean_expired_sessions();

-- 删除索引
DROP INDEX IF EXISTS idx_user_sessions_user_id;
DROP INDEX IF EXISTS idx_user_sessions_expires_at;

-- 删除表
DROP TABLE IF EXISTS user_sessions;
