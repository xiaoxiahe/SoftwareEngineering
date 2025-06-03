-- 删除触发器
DROP TRIGGER IF EXISTS trigger_pricing_config_check_overlap ON pricing_config;

-- 删除函数
DROP FUNCTION IF EXISTS check_pricing_time_overlap();

-- 删除索引
DROP INDEX IF EXISTS idx_pricing_config_price_type;
DROP INDEX IF EXISTS idx_pricing_config_effective_date;

-- 删除表
DROP TABLE IF EXISTS pricing_config;
