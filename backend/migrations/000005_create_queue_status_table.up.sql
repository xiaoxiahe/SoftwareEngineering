-- 创建queue_status表
CREATE TABLE IF NOT EXISTS queue_status (
    id UUID PRIMARY KEY,
    pile_id VARCHAR(10) NOT NULL,
    position INTEGER NOT NULL CHECK (position BETWEEN 1 AND 2),
    request_id UUID NOT NULL,
    user_id UUID NOT NULL,
    queue_number VARCHAR(20) NOT NULL,
    entered_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL,
    started_at TIMESTAMP,
    
    CONSTRAINT fk_queue_status_pile
        FOREIGN KEY (pile_id)
        REFERENCES charging_piles(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_queue_status_request
        FOREIGN KEY (request_id)
        REFERENCES charging_requests(id)
        ON DELETE CASCADE,
    CONSTRAINT fk_queue_status_user
        FOREIGN KEY (user_id)
        REFERENCES users(id)
        ON DELETE CASCADE,
    CONSTRAINT uq_queue_status_pile_position
        UNIQUE (pile_id, position)
);

-- 创建索引
CREATE INDEX idx_queue_status_pile_id ON queue_status(pile_id);
CREATE INDEX idx_queue_status_user_id ON queue_status(user_id);
CREATE INDEX idx_queue_status_request_id ON queue_status(request_id);
CREATE INDEX idx_queue_status_entered_at ON queue_status(entered_at);
