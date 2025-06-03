-- 删除触发器
DROP TRIGGER IF EXISTS trigger_system_config_updated_at ON system_config;

-- 删除索引
DROP INDEX IF EXISTS idx_system_config_key;

-- 删除表
DROP TABLE IF EXISTS system_config;
