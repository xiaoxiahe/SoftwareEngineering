package models

import (
	"sync"
	"time"
)

// 充电桩类型
type PileType string

const (
	PileTypeFast    PileType = "fast"    // 快充
	PileTypeTrickle PileType = "trickle" // 慢充
)

// 充电桩状态
type PileStatus string

const (
	PileStatusAvailable   PileStatus = "available"   // 空闲
	PileStatusCharging    PileStatus = "charging"    // 充电中
	PileStatusFault       PileStatus = "fault"       // 故障
	PileStatusMaintenance PileStatus = "maintenance" // 维护
	PileStatusOffline     PileStatus = "offline"     // 离线
)

// 故障类型
type FaultType string

const (
	FaultTypeHardware FaultType = "hardware" // 硬件故障
	FaultTypeSoftware FaultType = "software" // 软件故障
	FaultTypePower    FaultType = "power"    // 电力故障
)

// Pile 充电桩模型
type Pile struct {
	ID           string     `json:"id"`     // 充电桩ID
	Type         PileType   `json:"type"`   // 充电桩类型
	Status       PileStatus `json:"status"` // 当前状态
	Power        float64    `json:"power"`  // 充电功率(kW)
	CurrentFault *Fault     `json:"fault"`  // 当前故障信息

	CurrentVehicle *ChargingVehicle `json:"currentVehicle"` // 当前正在充电的车辆

	// 统计数据
	TotalChargingSessions int     `json:"totalChargingSessions"` // 总充电次数
	TotalChargingTime     int64   `json:"totalChargingTime"`     // 总充电时间(秒)
	TotalChargingAmount   float64 `json:"totalChargingAmount"`   // 总充电量(kWh)

	mu sync.Mutex // 互斥锁，保护并发操作
}

// ChargingVehicle 充电中的车辆
type ChargingVehicle struct {
	UserID            string    `json:"userId"`            // 用户ID
	StartTime         time.Time `json:"startTime"`         // 开始充电时间
	RequestedCapacity float64   `json:"requestedCapacity"` // 请求充电量(kWh)
	CurrentCapacity   float64   `json:"currentCapacity"`   // 当前已充电量(kWh)
	ChargingMode      string    `json:"chargingMode"`      // 充电模式
}

// Fault 故障信息
type Fault struct {
	Type        FaultType `json:"type"`        // 故障类型
	Description string    `json:"description"` // 故障描述
	StartTime   time.Time `json:"startTime"`   // 故障开始时间
	EndTime     time.Time `json:"endTime"`     // 预计恢复时间
}

// NewPile 创建新的充电桩
func NewPile(id string, pileType PileType, power float64) *Pile {
	return &Pile{
		ID:     id,
		Type:   pileType,
		Status: PileStatusAvailable,
		Power:  power,
	}
}

// StartCharging 开始充电
func (p *Pile) StartCharging(vehicle *ChargingVehicle) bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PileStatusAvailable {
		return false
	}

	p.CurrentVehicle = vehicle
	p.Status = PileStatusCharging
	return true
}

// StopCharging 停止充电
func (p *Pile) StopCharging() *ChargingVehicle {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PileStatusCharging || p.CurrentVehicle == nil {
		return nil
	}

	vehicle := p.CurrentVehicle
	chargingTime := time.Since(vehicle.StartTime).Seconds()

	// 更新统计数据
	p.TotalChargingSessions++
	p.TotalChargingTime += int64(chargingTime)
	p.TotalChargingAmount += vehicle.CurrentCapacity

	// 清空当前充电记录
	completedVehicle := p.CurrentVehicle
	p.CurrentVehicle = nil
	p.Status = PileStatusAvailable

	return completedVehicle
}

// UpdateChargingProgress 更新充电进度
func (p *Pile) UpdateChargingProgress(elapsed time.Duration) float64 {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PileStatusCharging || p.CurrentVehicle == nil {
		return 0
	}

	// 计算本次更新充电量
	hoursFraction := elapsed.Hours()
	additionalCharge := p.Power * hoursFraction // kW * h = kWh

	// 更新充电量，不超过请求量
	p.CurrentVehicle.CurrentCapacity += additionalCharge
	if p.CurrentVehicle.CurrentCapacity > p.CurrentVehicle.RequestedCapacity {
		p.CurrentVehicle.CurrentCapacity = p.CurrentVehicle.RequestedCapacity
	}

	return p.CurrentVehicle.CurrentCapacity
}

// RemainingTime 计算剩余充电时间(秒)
func (p *Pile) RemainingTime() int {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PileStatusCharging || p.CurrentVehicle == nil {
		return 0
	}

	remainingCapacity := p.CurrentVehicle.RequestedCapacity - p.CurrentVehicle.CurrentCapacity
	if remainingCapacity <= 0 {
		return 0
	}

	// 剩余时间 = 剩余电量 / 充电功率 (小时) * 3600 (转换为秒)
	remainingTimeSeconds := int((remainingCapacity / p.Power) * 3600)
	return remainingTimeSeconds
}

// ReportFault 报告故障
func (p *Pile) ReportFault(faultType FaultType, description string) {
	p.mu.Lock()
	defer p.mu.Unlock()

	now := time.Now().UTC()
	p.Status = PileStatusFault
	p.CurrentFault = &Fault{
		Type:        faultType,
		Description: description,
		StartTime:   now,
	}
}

// RecoverFromFault 从故障恢复
func (p *Pile) RecoverFromFault() {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PileStatusFault {
		return
	}

	p.Status = PileStatusAvailable
	p.CurrentFault = nil
}

// GetStatus 获取充电桩状态信息
func (p *Pile) GetStatus() (PileStatus, *ChargingVehicle) {
	p.mu.Lock()
	defer p.mu.Unlock()

	var vehicle *ChargingVehicle
	if p.CurrentVehicle != nil {
		// 创建副本，避免外部修改
		vehicle = &ChargingVehicle{
			UserID:            p.CurrentVehicle.UserID,
			StartTime:         p.CurrentVehicle.StartTime,
			RequestedCapacity: p.CurrentVehicle.RequestedCapacity,
			CurrentCapacity:   p.CurrentVehicle.CurrentCapacity,
			ChargingMode:      p.CurrentVehicle.ChargingMode,
		}
	}
	return p.Status, vehicle
}

// IsChargingComplete 检查充电是否完成
func (p *Pile) IsChargingComplete() bool {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.Status != PileStatusCharging || p.CurrentVehicle == nil {
		return false
	}

	return p.CurrentVehicle.CurrentCapacity >= p.CurrentVehicle.RequestedCapacity
}
