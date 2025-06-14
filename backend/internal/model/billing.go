package model

import (
	"time"

	"github.com/google/uuid"
)

// PriceRate 电价费率
type PriceRate struct {
	Period      string  `json:"period"`      // peak/normal/valley
	ElectricFee float64 `json:"electricFee"` // 电费(元/度)
	ServiceFee  float64 `json:"serviceFee"`  // 服务费(元/度)
	StartHour   int     `json:"startHour"`   // 开始小时
	EndHour     int     `json:"endHour"`     // 结束小时
}

// TimePeriod 时间段
type TimePeriod struct {
	StartHour int `json:"startHour"`
	EndHour   int `json:"endHour"`
}

// PricePeriod 价格时段
type PricePeriod struct {
	Period    string     `json:"period"` // peak/normal/valley
	TimeRange TimePeriod `json:"timeRange"`
}

// ChargingBill 充电账单
type ChargingBill struct {
	ID               uuid.UUID  `json:"id"`
	UserID           uuid.UUID  `json:"userId"`
	SessionID        uuid.UUID  `json:"sessionId"`
	ChargingCapacity float64    `json:"chargingCapacity"` // 充电电量(度)
	ChargingFee      float64    `json:"chargingFee"`      // 充电电费(元)
	ServiceFee       float64    `json:"serviceFee"`       // 服务费(元)
	TotalAmount      float64    `json:"totalAmount"`      // 总金额(元)
	StartTime        time.Time  `json:"startTime"`        // 开始充电时间
	EndTime          time.Time  `json:"endTime"`          // 结束充电时间
	ChargingDuration float64    `json:"chargingDuration"` // 充电时长(小时)
	PileID           string     `json:"pileId"`           // 充电桩ID
	ChargingMode     string     `json:"chargingMode"`     // 充电模式(快充/慢充)
	PeakCharge       float64    `json:"peakCharge"`       // 峰时充电量(度)
	NormalCharge     float64    `json:"normalCharge"`     // 平时充电量(度)
	ValleyCharge     float64    `json:"valleyCharge"`     // 谷时充电量(度)
	PeakFee          float64    `json:"peakFee"`          // 峰时费用(元)
	NormalFee        float64    `json:"normalFee"`        // 平时费用(元)
	ValleyFee        float64    `json:"valleyFee"`        // 谷时费用(元)
	CreatedAt        time.Time  `json:"createdAt"`        // 创建时间
	PaidAt           *time.Time `json:"paidAt,omitempty"` // 支付时间
	Status           string     `json:"status"`           // 支付状态(unpaid/paid)
}

// BillRequest 账单查询请求
type BillRequest struct {
	UserID    uuid.UUID `form:"userId"`
	StartDate time.Time `form:"startDate"`
	EndDate   time.Time `form:"endDate"`
	Status    string    `form:"status"`
	Page      int       `form:"page,default=1"`
	PageSize  int       `form:"pageSize,default=10"`
}

// DailySummary 每日收入汇总
type DailySummary struct {
	Date        string  `json:"date"`
	ChargingFee float64 `json:"chargingFee"`
	ServiceFee  float64 `json:"serviceFee"`
	TotalAmount float64 `json:"totalAmount"`
	Sessions    int     `json:"sessions"`
}

// RevenueStatistics 收入统计
type RevenueStatistics struct {
	PeriodType         string         `json:"periodType"`         // daily/weekly/monthly
	TotalChargingFee   float64        `json:"totalChargingFee"`   // 总充电费(元)
	TotalServiceFee    float64        `json:"totalServiceFee"`    // 总服务费(元)
	TotalRevenue       float64        `json:"totalRevenue"`       // 总收入(元)
	TotalSessions      int            `json:"totalSessions"`      // 总充电会话数
	TotalChargingHours float64        `json:"totalChargingHours"` // 总充电时长(小时)
	FastRevenue        float64        `json:"fastRevenue"`        // 快充收入(元)
	SlowRevenue        float64        `json:"slowRevenue"`        // 慢充收入(元)
	DailySummaries     []DailySummary `json:"dailySummaries"`     // 每日汇总
}

// TimeSegmentBilling 分时段计费详情
type TimeSegmentBilling struct {
	Period       string    `json:"period"`       // peak/normal/valley
	StartTime    time.Time `json:"startTime"`    // 该时段开始时间
	EndTime      time.Time `json:"endTime"`      // 该时段结束时间
	Duration     float64   `json:"duration"`     // 该时段充电时长(小时)
	Capacity     float64   `json:"capacity"`     // 该时段充电电量(度)
	ElectricFee  float64   `json:"electricFee"`  // 该时段电费单价
	ServiceFee   float64   `json:"serviceFee"`   // 该时段服务费单价
	ElectricCost float64   `json:"electricCost"` // 该时段电费
	ServiceCost  float64   `json:"serviceCost"`  // 该时段服务费
	TotalCost    float64   `json:"totalCost"`    // 该时段总费用
}
