-- 创建charging_piles表
CREATE TABLE IF NOT EXISTS charging_piles (
    id VARCHAR(10) PRIMARY KEY,
    pile_type VARCHAR(10) NOT NULL CHECK (pile_type IN ('fast', 'slow')),
    power DECIMAL(5,2) NOT NULL CHECK (power > 0),
    status VARCHAR(20) NOT NULL CHECK (status IN ('available', 'occupied', 'fault', 'maintenance', 'offline')),
    queue_length INTEGER NOT NULL CHECK (queue_length >= 0),
    total_sessions INTEGER DEFAULT 0 NOT NULL,
    total_duration DECIMAL(10,2) DEFAULT 0 NOT NULL,
    total_energy DECIMAL(12,2) DEFAULT 0 NOT NULL,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    updated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 创建索引
CREATE INDEX idx_charging_piles_status ON charging_piles(status);
CREATE INDEX idx_charging_piles_type ON charging_piles(pile_type);

-- 使用之前创建的updated_at更新触发器
CREATE TRIGGER trigger_charging_piles_updated_at
BEFORE UPDATE ON charging_piles
FOR EACH ROW
EXECUTE FUNCTION update_updated_at_column();
