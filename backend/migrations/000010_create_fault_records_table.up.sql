-- 创建fault_records表
CREATE TABLE IF NOT EXISTS fault_records (
    id UUID PRIMARY KEY,
    pile_id VARCHAR(10) NOT NULL,
    fault_type VARCHAR(50) NOT NULL CHECK (fault_type IN ('hardware', 'software', 'power')),
    description TEXT,
    occurred_at TIMESTAMP NOT NULL,
    recovered_at TIMESTAMP,
    affected_sessions INTEGER DEFAULT 0 NOT NULL,
    status VARCHAR(20) NOT NULL CHECK (status IN ('active', 'resolved')),
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_fault_records_pile
        FOREIGN KEY (pile_id)
        REFERENCES charging_piles(id)
        ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_fault_records_pile_id ON fault_records(pile_id);
CREATE INDEX idx_fault_records_status ON fault_records(status);
CREATE INDEX idx_fault_records_occurred_at ON fault_records(occurred_at);

-- 创建故障发生时更新充电桩状态的触发器
CREATE OR REPLACE FUNCTION update_pile_status_on_fault()
RETURNS TRIGGER AS $$
BEGIN
    -- 故障发生时，将充电桩状态设为故障
    IF NEW.status = 'active' THEN
        UPDATE charging_piles SET status = 'fault' WHERE id = NEW.pile_id;
    -- 故障恢复时，将充电桩状态设为可用
    ELSIF NEW.status = 'resolved' THEN
        UPDATE charging_piles SET status = 'available' WHERE id = NEW.pile_id;
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_fault_records_status_change
AFTER INSERT OR UPDATE OF status ON fault_records
FOR EACH ROW
EXECUTE FUNCTION update_pile_status_on_fault();
