-- 创建charging_sessions表
CREATE TABLE IF NOT EXISTS charging_sessions (
    id UUID PRIMARY KEY,
    request_id UUID NOT NULL,
    user_id UUID NOT NULL,
    pile_id VARCHAR(10) NOT NULL,
    queue_number VARCHAR(20) NOT NULL,
    requested_capacity DECIMAL(8,2) NOT NULL,
    actual_capacity DECIMAL(8,2),
    start_time TIMESTAMP NOT NULL,
    end_time TIMESTAMP,
    duration DECIMAL(10,4),
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'completed', 'interrupted')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_charging_sessions_request
        FOREIGN KEY (request_id)
        REFERENCES charging_requests(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_charging_sessions_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_charging_sessions_pile
        FOREIGN KEY (pile_id)
        REFERENCES charging_piles(id)
        ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_charging_sessions_user_id ON charging_sessions(user_id);
CREATE INDEX idx_charging_sessions_pile_id ON charging_sessions(pile_id);
CREATE INDEX idx_charging_sessions_request_id ON charging_sessions(request_id);
CREATE INDEX idx_charging_sessions_start_time ON charging_sessions(start_time);
CREATE INDEX idx_charging_sessions_status ON charging_sessions(status);
CREATE INDEX idx_charging_sessions_pile_start_status ON charging_sessions(pile_id, start_time, status);

-- 创建更新充电桩统计数据的触发器
CREATE OR REPLACE FUNCTION update_charging_pile_stats()
RETURNS TRIGGER AS $$
BEGIN
    IF (TG_OP = 'INSERT') THEN
        -- 新增会话时更新充电桩统计
        UPDATE charging_piles
        SET 
            total_sessions = total_sessions + 1
        WHERE id = NEW.pile_id;
    ELSIF (TG_OP = 'UPDATE') THEN
        -- 会话完成时更新充电桩统计
        IF (NEW.status = 'completed' AND OLD.status != 'completed') THEN
            UPDATE charging_piles
            SET 
                total_duration = total_duration + NEW.duration,
                total_energy = total_energy + NEW.actual_capacity
            WHERE id = NEW.pile_id;
        END IF;
    END IF;
    
    RETURN NULL;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_after_charging_session_insert
AFTER INSERT ON charging_sessions
FOR EACH ROW
EXECUTE FUNCTION update_charging_pile_stats();

CREATE TRIGGER trigger_after_charging_session_update
AFTER UPDATE ON charging_sessions
FOR EACH ROW
EXECUTE FUNCTION update_charging_pile_stats();
