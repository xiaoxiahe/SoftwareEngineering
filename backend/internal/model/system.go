package model

import (
	"time"

	"github.com/google/uuid"
)

// FaultType 故障类型
type FaultType string

const (
	FaultTypeHardware FaultType = "hardware"
	FaultTypeSoftware FaultType = "software"
	FaultTypePower    FaultType = "power"
)

// FaultStatus 故障状态
type FaultStatus string

const (
	FaultStatusActive   FaultStatus = "active"
	FaultStatusResolved FaultStatus = "resolved"
)

// FaultRecord 故障记录
type FaultRecord struct {
	ID               uuid.UUID   `json:"id"`
	PileID           string      `json:"pileId"`
	FaultType        FaultType   `json:"faultType"`
	Description      string      `json:"description"`
	OccurredAt       time.Time   `json:"occurredAt"`
	RecoveredAt      *time.Time  `json:"recoveredAt,omitempty"`
	AffectedSessions int         `json:"affectedSessions"`
	Status           FaultStatus `json:"status"`
	CreatedAt        time.Time   `json:"createdAt"`
}

// FaultReportRequest 故障报告请求
type FaultReportRequest struct {
	FaultType FaultType `json:"faultType" binding:"required,oneof=hardware software power"`
}

// FaultRecoveryRequest 故障恢复请求
type FaultRecoveryRequest struct {
	FaultID       uuid.UUID `json:"faultId" binding:"required"`
	RecoveryNotes string    `json:"recoveryNotes"`
}

// SystemConfig 系统配置
type SystemConfig struct {
	ID          int       `json:"id"`
	ConfigKey   string    `json:"configKey"`
	ConfigValue string    `json:"configValue"`
	ConfigType  string    `json:"configType"` // string/number/boolean/json
	Description string    `json:"description"`
	UpdatedAt   time.Time `json:"updatedAt"`
}

// ExtendedSchedulingMode 扩展调度模式
type ExtendedSchedulingMode string

const (
	ExtendedModeDisabled      ExtendedSchedulingMode = "disabled"      // 禁用扩展调度
	ExtendedModeBatch         ExtendedSchedulingMode = "batch"         // 批量调度总充电时长最短
	ExtendedModeSingleOptimal ExtendedSchedulingMode = "singleOptimal" // 单次调度总充电时长最短
)

// SchedulingConfig 调度配置
type SchedulingConfig struct {
	Strategy               string                 `json:"strategy"` // shortest_completion_time, etc.
	FastChargingPileNum    int                    `json:"fastChargingPileNum"`
	SlowChargingPileNum    int                    `json:"slowChargingPileNum"`
	WaitingAreaSize        int                    `json:"waitingAreaSize"`
	ChargingQueueLen       int                    `json:"chargingQueueLen"`
	FastChargingPower      float64                `json:"fastChargingPower"`      // 度/小时
	SlowChargingPower      float64                `json:"slowChargingPower"`      // 度/小时
	ExtendedSchedulingMode ExtendedSchedulingMode `json:"extendedSchedulingMode"` // 扩展调度模式
}

// StatisticsReport 统计报表
type StatisticsReport struct {
	Period       string       `json:"period"`       // day, week, month
	Date         string       `json:"date"`         // YYYY-MM-DD, YYYY-WW, YYYY-MM
	PileReports  []PileReport `json:"pileReports"`  // 充电桩报表
	SystemReport SystemReport `json:"systemReport"` // 系统报表
}

// PileReport 充电桩报表
type PileReport struct {
	PileID        string  `json:"pileId"`
	TotalSessions int     `json:"totalSessions"` // 充电次数
	TotalDuration float64 `json:"totalDuration"` // 小时
	TotalEnergy   float64 `json:"totalEnergy"`   // 度
	ChargingFee   float64 `json:"chargingFee"`   // 元
	ServiceFee    float64 `json:"serviceFee"`    // 元
	TotalFee      float64 `json:"totalFee"`      // 元
}

// SystemReport 系统报表
type SystemReport struct {
	TotalRequests     int     `json:"totalRequests"`     // 总请求数
	CompletedRequests int     `json:"completedRequests"` // 完成的请求数
	CancelledRequests int     `json:"cancelledRequests"` // 取消的请求数
	AvgWaitingTime    float64 `json:"avgWaitingTime"`    // 平均等待时间(分钟)
	TotalRevenue      float64 `json:"totalRevenue"`      // 总收入(元)
	PeakTimeUsage     float64 `json:"peakTimeUsage"`     // 峰时使用率(%)
}
