-- 为充电请求表添加优先级字段
ALTER TABLE charging_requests 
ADD COLUMN priority INTEGER DEFAULT 0 CHECK (priority >= 0);

-- 创建优先级索引
CREATE INDEX idx_charging_requests_priority ON charging_requests(priority);

-- 创建复合索引用于按优先级和创建时间排序
CREATE INDEX idx_charging_requests_priority_created_at ON charging_requests(priority DESC, created_at);

-- 创建按模式、状态、优先级排序的索引
CREATE INDEX idx_charging_requests_mode_status_priority ON charging_requests(charging_mode, status, priority DESC, created_at);
