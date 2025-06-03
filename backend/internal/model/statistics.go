package model

import "time"

// BillingStatistics 计费统计数据
type BillingStatistics struct {
	StartTime        time.Time `json:"startTime"`        // 开始时间
	EndTime          time.Time `json:"endTime"`          // 结束时间
	PileID           string    `json:"pileId,omitempty"` // 充电桩ID（可选）
	Count            int       `json:"count"`            // 充电次数
	TotalDuration    float64   `json:"totalDuration"`    // 总充电时长（小时）
	TotalCapacity    float64   `json:"totalCapacity"`    // 总充电电量（度）
	TotalChargingFee float64   `json:"totalChargingFee"` // 总充电费用（元）
	TotalServiceFee  float64   `json:"totalServiceFee"`  // 总服务费用（元）
	TotalFee         float64   `json:"totalFee"`         // 总费用（元）
}

// DailyStatistics 每日统计
type DailyStatistics struct {
	Date             string  `json:"date"`             // 日期（YYYY-MM-DD）
	Count            int     `json:"count"`            // 充电次数
	TotalDuration    float64 `json:"totalDuration"`    // 总充电时长（小时）
	TotalCapacity    float64 `json:"totalCapacity"`    // 总充电电量（度）
	TotalChargingFee float64 `json:"totalChargingFee"` // 总充电费用（元）
	TotalServiceFee  float64 `json:"totalServiceFee"`  // 总服务费用（元）
	TotalFee         float64 `json:"totalFee"`         // 总费用（元）
}

// PileUsageStatistics 充电桩使用统计
type PileUsageStatistics struct {
	PileID           string  `json:"pileID"`           // 充电桩ID
	Count            int     `json:"count"`            // 充电次数
	TotalDuration    float64 `json:"totalDuration"`    // 总充电时长（小时）
	TotalCapacity    float64 `json:"totalCapacity"`    // 总充电电量（度）
	TotalChargingFee float64 `json:"totalChargingFee"` // 总充电费用（元）
	TotalServiceFee  float64 `json:"totalServiceFee"`  // 总服务费用（元）
	TotalFee         float64 `json:"totalFee"`         // 总费用（元）
}

// QueueStatistics 队列统计
type QueueStatistics struct {
	AvgWaitTime    float64 `json:"avgWaitTime"`    // 平均等待时间（分钟）
	MaxWaitTime    float64 `json:"maxWaitTime"`    // 最长等待时间（分钟）
	TotalRequests  int     `json:"totalRequests"`  // 总请求数
	CancelledCount int     `json:"cancelledCount"` // 取消请求数
	CompletedCount int     `json:"completedCount"` // 完成请求数
}
