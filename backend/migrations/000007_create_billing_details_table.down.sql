-- 删除索引
DROP INDEX IF EXISTS idx_billing_details_user_id;
DROP INDEX IF EXISTS idx_billing_details_session_id;
DROP INDEX IF EXISTS idx_billing_details_generated_at;
DROP INDEX IF EXISTS idx_billing_details_pile_id;
DROP INDEX IF EXISTS idx_billing_details_user_generated_pile;

-- 删除表
DROP TABLE IF EXISTS billing_details;
