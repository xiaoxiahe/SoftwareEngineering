package model

import (
	"time"

	"github.com/google/uuid"
)

// ChargingMode 充电模式
type ChargingMode string

const (
	ChargingModeFast ChargingMode = "fast"
	ChargingModeSlow ChargingMode = "slow"
)

// RequestStatus 充电请求状态
type RequestStatus string

const (
	RequestStatusWaiting   RequestStatus = "waiting"   // 等候区等待
	RequestStatusQueued    RequestStatus = "queued"    // 进入充电桩队列
	RequestStatusCharging  RequestStatus = "charging"  // 正在充电
	RequestStatusCompleted RequestStatus = "completed" // 充电完成
	RequestStatusCancelled RequestStatus = "cancelled" // 已取消
)

// ChargingRequest 充电请求
type ChargingRequest struct {
	ID                uuid.UUID     `json:"id"`
	UserID            uuid.UUID     `json:"userId"`
	ChargingMode      ChargingMode  `json:"chargingMode"`      // fast/slow
	RequestedCapacity float64       `json:"requestedCapacity"` // 请求充电量(度)
	QueueNumber       string        `json:"queueNumber"`       // F1, T1 等
	PileID            string        `json:"pileId,omitempty"`  // 分配的充电桩ID
	QueuePosition     int           `json:"queuePosition"`     // 队列位置
	Status            RequestStatus `json:"status"`            // waiting/queued/charging/completed/cancelled
	EstimatedWaitTime int           `json:"estimatedWaitTime"` // 预估等待时间(秒)
	CreatedAt         time.Time     `json:"createdAt"`
	UpdatedAt         time.Time     `json:"updatedAt"`
}

// ChargingRequestCreate 创建充电请求
type ChargingRequestCreate struct {
	ChargingMode      ChargingMode `json:"chargingMode" binding:"required,oneof=fast slow"`
	RequestedCapacity float64      `json:"requestedCapacity" binding:"required,gt=0"`
}

// ChargingRequestUpdate 更新充电请求
type ChargingRequestUpdate struct {
	ChargingMode      ChargingMode `json:"chargingMode,omitempty" binding:"omitempty,oneof=fast slow"`
	RequestedCapacity float64      `json:"requestedCapacity,omitempty" binding:"omitempty,gt=0"`
}

// QueueItem 队列项
type QueueItem struct {
	UserID            uuid.UUID    `json:"userId"`
	QueueNumber       string       `json:"queueNumber"`
	RequestID         uuid.UUID    `json:"requestId"`
	PileID            string       `json:"pileId"` // 充电桩ID
	ChargingMode      ChargingMode `json:"chargingMode"`
	RequestedCapacity float64      `json:"requestedCapacity"`
	Position          int          `json:"position"` // 队列位置
	WaitTime          int          `json:"waitTime"` // 等待时间(秒)
	EnterTime         time.Time    `json:"enterTime"`
}

// QueueStatus 排队状态
type QueueStatus struct {
	FastQueue      []QueueItem `json:"fastQueue"`      // 快充队列
	SlowQueue      []QueueItem `json:"slowQueue"`      // 慢充队列
	AvailableSlots int         `json:"availableSlots"` // 等候区可用车位
}

// UserPosition 用户在队列中的位置
type UserPosition struct {
	UserID        uuid.UUID     `json:"userId"`
	QueueNumber   string        `json:"queueNumber"`
	ChargingMode  ChargingMode  `json:"chargingMode"`
	Position      int           `json:"position"`      // 队列位置
	WaitingTime   int           `json:"waitingTime"`   // 等待时间(秒)
	AssignedPile  string        `json:"assignedPile"`  // 分配的充电桩ID(如果已分配)
	QueuePosition int           `json:"queuePosition"` // 充电桩队列位置(如果已分配)
	Status        RequestStatus `json:"status"`        // 请求状态
}
