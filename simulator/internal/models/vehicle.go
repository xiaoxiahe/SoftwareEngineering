package models

import (
	"time"
)

// ChargingMode 充电模式
type ChargingMode string

const (
	ChargingModeFast    ChargingMode = "fast"    // 快充
	ChargingModeTrickle ChargingMode = "trickle" // 慢充
)

// ChargingRequest 充电请求
type ChargingRequest struct {
	UserID       string       `json:"userId"`       // 用户ID
	Amount       float64      `json:"amount"`       // 请求充电量(kWh)
	ChargingMode ChargingMode `json:"chargingMode"` // 充电模式
	RequestTime  time.Time    `json:"requestTime"`  // 请求时间
}

// ChargingResponse 充电响应
type ChargingResponse struct {
	RequestID     string    `json:"requestId"`    // 请求ID
	UserID        string    `json:"userId"`       // 用户ID
	PileID        string    `json:"pileId"`       // 分配的充电桩ID
	QueueNumber   string    `json:"queueNumber"`  // 排队号
	EstimatedWait int       `json:"waitTime"`     // 预计等待时间(分钟)
	ResponseTime  time.Time `json:"responseTime"` // 响应时间
}

// ChargingProgress 充电进度
type ChargingProgress struct {
	UserID            string    `json:"userId"`            // 用户ID
	PileID            string    `json:"pileId"`            // 充电桩ID
	StartTime         time.Time `json:"startTime"`         // 开始时间
	RequestedCapacity float64   `json:"requestedCapacity"` // 请求电量
	CurrentCapacity   float64   `json:"currentCapacity"`   // 当前电量
	ChargingRate      float64   `json:"chargingRate"`      // 充电速率(kWh/h)
	RemainingTime     int       `json:"remainingTime"`     // 剩余时间(秒)
}

// ChargingCompletion 充电完成信息
type ChargingCompletion struct {
	UserID       string    `json:"userId"`       // 用户ID
	PileID       string    `json:"pileId"`       // 充电桩ID
	StartTime    time.Time `json:"startTime"`    // 开始时间
	EndTime      time.Time `json:"endTime"`      // 结束时间
	ChargingMode string    `json:"chargingMode"` // 充电模式
	Amount       float64   `json:"amount"`       // 充电量
}

// Vehicle 车辆信息
type Vehicle struct {
	UserID          string           `json:"userId"`          // 用户ID，表示车辆所有者
	BatteryCapacity float64          `json:"batteryCapacity"` // 电池容量(kWh)
	ChargingRequest *ChargingRequest // 当前充电请求
	ChargingStatus  string           `json:"status"`         // 状态: idle, waiting, charging, complete
	AssignedPileID  string           `json:"assignedPileId"` // 分配的充电桩ID
	QueueNumber     string           `json:"queueNumber"`    // 排队号
}

// NewVehicle 创建新车辆
func NewVehicle(userID string, batteryCapacity float64) *Vehicle {
	return &Vehicle{
		UserID:          userID,
		BatteryCapacity: batteryCapacity,
		ChargingStatus:  "idle",
	}
}

// RequestCharging 发起充电请求
func (v *Vehicle) RequestCharging(amount float64, mode ChargingMode) *ChargingRequest {
	v.ChargingRequest = &ChargingRequest{
		UserID:       v.UserID,
		Amount:       amount,
		ChargingMode: mode,
		RequestTime:  time.Now().UTC(),
	}
	v.ChargingStatus = "waiting"
	return v.ChargingRequest
}

// AssignCharger 分配充电桩
func (v *Vehicle) AssignCharger(pileID string, queueNumber string) {
	v.AssignedPileID = pileID
	v.QueueNumber = queueNumber
}

// StartCharging 开始充电
func (v *Vehicle) StartCharging() {
	v.ChargingStatus = "charging"
}

// CompleteCharging 完成充电
func (v *Vehicle) CompleteCharging() {
	v.ChargingStatus = "complete"
	v.ChargingRequest = nil
	v.AssignedPileID = ""
	v.QueueNumber = ""
}

// CancelCharging 取消充电请求
func (v *Vehicle) CancelCharging() {
	v.ChargingStatus = "idle"
	v.ChargingRequest = nil
	v.AssignedPileID = ""
	v.QueueNumber = ""
}
