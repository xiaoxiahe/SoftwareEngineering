-- 删除触发器
DROP TRIGGER IF EXISTS trigger_users_updated_at ON users;

-- 删除函数
DROP FUNCTION IF EXISTS update_updated_at_column();

-- 删除索引
DROP INDEX IF EXISTS idx_users_username;
DROP INDEX IF EXISTS idx_users_user_type;

-- 删除表
DROP TABLE IF EXISTS users;
