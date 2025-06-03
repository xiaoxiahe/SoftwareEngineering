-- 创建system_config表
CREATE TABLE IF NOT EXISTS system_config (
    id SERIAL PRIMARY KEY,
    config_key VARCHAR(100) UNIQUE NOT NULL,
    config_value TEXT NOT NULL,
    config_type VARCHAR(20) NOT NULL CHECK (config_type IN ('string', 'number', 'boolean', 'json')),
    description TEXT,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 创建索引
CREATE INDEX idx_system_config_key ON system_config(config_key);

-- 使用之前创建的updated_at更新触发器
CREATE TRIGGER trigger_system_config_updated_at
BEFORE UPDATE ON system_config
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
