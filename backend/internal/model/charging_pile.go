package model

import (
	"time"
)

// PileType 表示充电桩类型
type PileType string

const (
	PileTypeFast PileType = "fast"
	PileTypeSlow PileType = "slow"
)

// PileStatus 表示充电桩状态
type PileStatus string

const (
	PileStatusAvailable   PileStatus = "available"
	PileStatusOccupied    PileStatus = "occupied"
	PileStatusFault       PileStatus = "fault"
	PileStatusMaintenance PileStatus = "maintenance"
	PileStatusOffline     PileStatus = "offline"
)

// ChargingPile 充电桩模型
type ChargingPile struct {
	ID            string     `json:"id"`            // A, B, C, D, E
	PileType      PileType   `json:"pileType"`      // fast/slow
	Power         float64    `json:"power"`         // 充电功率(度/小时)
	Status        PileStatus `json:"status"`        // 充电桩状态
	QueueLength   int        `json:"queueLength"`   // 队列长度
	TotalSessions int        `json:"totalSessions"` // 累计充电次数
	TotalDuration float64    `json:"totalDuration"` // 累计充电时长(小时)
	TotalEnergy   float64    `json:"totalEnergy"`   // 累计充电电量(度)
	CreatedAt     time.Time  `json:"createdAt"`
	UpdatedAt     time.Time  `json:"updatedAt"`
}

// ChargingPileControlRequest 充电桩控制请求
type ChargingPileControlRequest struct {
	Action string `json:"action" binding:"required,oneof=start stop maintenance"` // start/stop/maintenance
	Reason string `json:"reason"`                                                 // 原因(可选)
}

// ChargingPileStatus 充电桩状态响应
type ChargingPileStatus struct {
	ID                string      `json:"id"`
	PileType          PileType    `json:"pileType"`
	Power             float64     `json:"power"`
	Status            PileStatus  `json:"status"`
	CurrentVehicle    *QueueItem  `json:"currentVehicle,omitempty"` // 当前充电的车辆
	QueueItems        []QueueItem `json:"queueItems"`               // 排队的车辆
	TotalSessions     int         `json:"totalSessions"`
	TotalDuration     float64     `json:"totalDuration"`
	TotalEnergy       float64     `json:"totalEnergy"`
	AvailableQueuePos int         `json:"availableQueuePos"` // 可用队列位置
}
