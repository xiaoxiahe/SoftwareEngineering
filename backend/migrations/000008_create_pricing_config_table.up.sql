-- 创建pricing_config表
CREATE TABLE IF NOT EXISTS pricing_config (
    id SERIAL PRIMARY KEY,
    price_type VARCHAR(20) NOT NULL CHECK (price_type IN ('peak', 'normal', 'valley')),
    unit_price DECIMAL(5,2) NOT NULL CHECK (unit_price > 0),
    start_time TIME NOT NULL,
    end_time TIME NOT NULL,
    service_fee_rate DECIMAL(5,2) NOT NULL CHECK (service_fee_rate >= 0),
    effective_date DATE NOT NULL DEFAULT CURRENT_DATE,
    created_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP NOT NULL
);

-- 创建索引
CREATE INDEX idx_pricing_config_price_type ON pricing_config(price_type);
CREATE INDEX idx_pricing_config_effective_date ON pricing_config(effective_date);

-- 检查时间段不重叠的约束
CREATE OR REPLACE FUNCTION check_pricing_time_overlap()
RETURNS TRIGGER AS $$
BEGIN
    -- 检查同一类型的时间段是否重叠
    IF EXISTS (
        SELECT 1 FROM pricing_config
        WHERE
            price_type = NEW.price_type AND
            id != NEW.id AND
            effective_date = NEW.effective_date AND
            (
                (NEW.start_time < end_time AND NEW.end_time > start_time) OR
                (NEW.start_time = start_time AND NEW.end_time = end_time)
            )
    ) THEN
        RAISE EXCEPTION 'Time periods for the same price type cannot overlap';
    END IF;
    
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

CREATE TRIGGER trigger_pricing_config_check_overlap
BEFORE INSERT OR UPDATE ON pricing_config
FOR EACH ROW
EXECUTE FUNCTION check_pricing_time_overlap();
