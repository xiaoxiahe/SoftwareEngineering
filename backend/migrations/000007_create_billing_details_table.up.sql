-- 创建billing_details表
CREATE TABLE IF NOT EXISTS billing_details (
    id UUID PRIMARY KEY,
    session_id UUID NOT NULL,
    user_id UUID NOT NULL,
    pile_id VARCHAR(10) NOT NULL,
    charging_capacity DECIMAL(8,2) NOT NULL CHECK (charging_capacity > 0),
    charging_duration DECIMAL(8,4) NOT NULL,
    start_time TIMESTAMP NOT NULL,
    stop_time TIMESTAMP NOT NULL,
    unit_price DECIMAL(5,2) NOT NULL CHECK (unit_price > 0),
    price_type VARCHAR(20) NOT NULL CHECK (price_type IN ('peak', 'normal', 'valley')),
    charging_fee DECIMAL(10,2) NOT NULL CHECK (charging_fee >= 0),
    service_fee DECIMAL(10,2) NOT NULL CHECK (service_fee >= 0),
    total_fee DECIMAL(10,2) NOT NULL CHECK (total_fee >= 0),
    generated_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    
    CONSTRAINT fk_billing_details_session
        FOREIGN KEY (session_id)
        REFERENCES charging_sessions(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_billing_details_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE
);

-- 创建索引
CREATE INDEX idx_billing_details_user_id ON billing_details(user_id);
CREATE INDEX idx_billing_details_session_id ON billing_details(session_id);
CREATE INDEX idx_billing_details_generated_at ON billing_details(generated_at);
CREATE INDEX idx_billing_details_pile_id ON billing_details(pile_id);
CREATE INDEX idx_billing_details_user_generated_pile ON billing_details(user_id, generated_at, pile_id);
