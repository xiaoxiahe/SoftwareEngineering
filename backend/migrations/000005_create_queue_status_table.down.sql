-- 删除索引
DROP INDEX IF EXISTS idx_queue_status_pile_id;
DROP INDEX IF EXISTS idx_queue_status_user_id;
DROP INDEX IF EXISTS idx_queue_status_request_id;
DROP INDEX IF EXISTS idx_queue_status_entered_at;

-- 删除表
DROP TABLE IF EXISTS queue_status;
