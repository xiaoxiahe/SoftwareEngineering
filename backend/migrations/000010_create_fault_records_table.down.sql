-- 删除触发器
DROP TRIGGER IF EXISTS trigger_fault_records_status_change ON fault_records;

-- 删除函数
DROP FUNCTION IF EXISTS update_pile_status_on_fault();

-- 删除索引
DROP INDEX IF EXISTS idx_fault_records_pile_id;
DROP INDEX IF EXISTS idx_fault_records_status;
DROP INDEX IF EXISTS idx_fault_records_occurred_at;

-- 删除表
DROP TABLE IF EXISTS fault_records;
