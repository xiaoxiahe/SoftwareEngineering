-- 删除触发器
DROP TRIGGER IF EXISTS trigger_charging_piles_updated_at ON charging_piles;

-- 删除索引
DROP INDEX IF EXISTS idx_charging_piles_status;
DROP INDEX IF EXISTS idx_charging_piles_type;

-- 删除表
DROP TABLE IF EXISTS charging_piles;
