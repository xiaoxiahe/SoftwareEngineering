-- 创建charging_requests表
CREATE TABLE IF NOT EXISTS charging_requests (
    id UUID PRIMARY KEY,
    user_id UUID NOT NULL,
    charging_mode VARCHAR(10) NOT NULL CHECK (charging_mode IN ('fast', 'slow')),
    requested_capacity DECIMAL(8,2) NOT NULL CHECK (requested_capacity > 0),
    queue_number VARCHAR(20),
    pile_id VARCHAR(10),
    queue_position INTEGER,
    status VARCHAR(20) NOT NULL CHECK (status IN ('waiting', 'queued', 'charging', 'completed', 'cancelled')),
    estimated_wait_time INTEGER,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_charging_requests_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_charging_requests_pile
        FOREIGN KEY (pile_id)
        REFERENCES charging_piles(id)
        ON DELETE SET NULL
);

-- 创建索引
CREATE INDEX idx_charging_requests_user_id ON charging_requests(user_id);
CREATE INDEX idx_charging_requests_status ON charging_requests(status);
CREATE INDEX idx_charging_requests_created_at ON charging_requests(created_at);
CREATE INDEX idx_charging_requests_pile_id ON charging_requests(pile_id);
CREATE INDEX idx_charging_requests_status_mode_created_at ON charging_requests(status, charging_mode, created_at);

-- 使用之前创建的updated_at更新触发器
CREATE TRIGGER trigger_charging_requests_updated_at
BEFORE UPDATE ON charging_requests
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
