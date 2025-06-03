package utils

import (
	"time"
)

// SimulationTimer 模拟时间管理器
type SimulationTimer struct {
	speedFactor float64 // 时间加速比例
}

// NewSimulationTimer 创建一个新的模拟时间管理器
func NewSimulationTimer(speedFactor float64) *SimulationTimer {
	if speedFactor <= 0 {
		speedFactor = 1.0
	}
	return &SimulationTimer{
		speedFactor: speedFactor,
	}
}

// SimTime 将实际时间转换为模拟时间
func (st *SimulationTimer) SimTime(realTime time.Duration) time.Duration {
	return time.Duration(float64(realTime) * st.speedFactor)
}

// RealTime 将模拟时间转换为实际时间
func (st *SimulationTimer) RealTime(simTime time.Duration) time.Duration {
	return time.Duration(float64(simTime) / st.speedFactor)
}

// NewTicker 创建一个模拟时间刻度器
func (st *SimulationTimer) NewTicker(d time.Duration) *time.Ticker {
	// 将模拟时间转换为实际时间
	realDuration := st.RealTime(d)
	return time.NewTicker(realDuration)
}

// Sleep 模拟时间睡眠
func (st *SimulationTimer) Sleep(d time.Duration) {
	// 将模拟时间转换为实际时间
	realDuration := st.RealTime(d)
	time.Sleep(realDuration)
}

// SetSpeedFactor 设置时间加速比例
func (st *SimulationTimer) SetSpeedFactor(speedFactor float64) {
	if speedFactor <= 0 {
		speedFactor = 1.0
	}
	st.speedFactor = speedFactor
}

// GetSpeedFactor 获取时间加速比例
func (st *SimulationTimer) GetSpeedFactor() float64 {
	return st.speedFactor
}
