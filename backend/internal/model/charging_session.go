package model

import (
	"time"

	"github.com/google/uuid"
)

// ChargingSession 充电会话
type ChargingSession struct {
	ID                uuid.UUID     `json:"id"`
	RequestID         uuid.UUID     `json:"requestId"`
	UserID            uuid.UUID     `json:"userId"`
	PileID            string        `json:"pileId"`
	QueueNumber       string        `json:"queueNumber"`
	RequestedCapacity float64       `json:"requestedCapacity"`
	ActualCapacity    float64       `json:"chargedCapacity"`
	StartTime         time.Time     `json:"startTime"`
	EndTime           *time.Time    `json:"endTime,omitempty"`
	Status            SessionStatus `json:"status"`
	Duration          float64       `json:"chargingDuration"` // 充电时长(秒)
	CreatedAt         time.Time     `json:"createdAt"`
}

// SessionStatus 充电会话状态
type SessionStatus string

const (
	SessionStatusActive      SessionStatus = "active"
	SessionStatusCompleted   SessionStatus = "completed"
	SessionStatusInterrupted SessionStatus = "interrupted"
)

// BillingDetail 充电详单
type BillingDetail struct {
	ID                uuid.UUID `json:"id"`
	SessionID         uuid.UUID `json:"sessionId"`
	UserID            uuid.UUID `json:"userId"`
	PileID            string    `json:"pileId"`
	ChargingCapacity  float64   `json:"chargingCapacity"`
	ChargingDuration  float64   `json:"chargingDuration"` // 小时
	StartTime         time.Time `json:"startTime"`
	EndTime           time.Time `json:"endTime"`
	UnitPrice         float64   `json:"unitPrice"`         // 电价单价
	PriceType         string    `json:"priceType"`         // 价格类型(peak/normal/valley)
	ChargingFee       float64   `json:"chargingFee"`       // 充电费用
	ServiceFee        float64   `json:"serviceFee"`        // 服务费用
	TotalFee          float64   `json:"totalFee"`          // 总费用
	PeakHours         float64   `json:"peakHours"`         // 峰时小时数
	NormalHours       float64   `json:"normalHours"`       // 平时小时数
	ValleyHours       float64   `json:"valleyHours"`       // 谷时小时数
	PeakElectricity   float64   `json:"peakElectricity"`   // 峰时电量
	NormalElectricity float64   `json:"normalElectricity"` // 平时电量
	ValleyElectricity float64   `json:"valleyElectricity"` // 谷时电量
	GeneratedAt       time.Time `json:"generatedAt"`
}

// BillingQuery 计费查询参数
type BillingQuery struct {
	StartDate time.Time  `form:"startDate"`
	EndDate   time.Time  `form:"endDate"`
	Page      int        `form:"page,default=1"`
	PageSize  int        `form:"pageSize,default=10"`
	UserID    *uuid.UUID `form:"userId"`
}

// CalculateFeeRequest 计算费用请求
type CalculateFeeRequest struct {
	Capacity  float64   `json:"capacity" binding:"required,gt=0"`
	StartTime time.Time `json:"startTime" binding:"required"`
}

// FeeCalculation 费用计算结果
type FeeCalculation struct {
	ChargingFee      float64 `json:"chargingFee"`
	ServiceFee       float64 `json:"serviceFee"`
	TotalFee         float64 `json:"totalFee"`
	ChargingDuration float64 `json:"chargingDuration"` // 小时
	PeakHours        float64 `json:"peakHours"`
	NormalHours      float64 `json:"normalHours"`
	ValleyHours      float64 `json:"valleyHours"`
}
