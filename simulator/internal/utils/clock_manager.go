package utils

import (
	"fmt"
	"sync"
	"time"
)

// ClockManager 时钟管理器，负责管理系统时钟和模拟时钟
type ClockManager struct {
	mu           sync.RWMutex
	systemClock  *SystemClock
	simClock     *SimulatedClock
	currentClock Clock
	useSimClock  bool
}

// NewClockManager 创建时钟管理器
func NewClockManager() *ClockManager {
	systemClock := NewSystemClock()
	return &ClockManager{
		systemClock:  systemClock,
		currentClock: systemClock,
		useSimClock:  false,
	}
}

// EnableSimulatedClock 启用模拟时钟
func (cm *ClockManager) EnableSimulatedClock(startTime time.Time) error {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	// 如果已经在使用模拟时钟，先停止
	if cm.useSimClock && cm.simClock != nil {
		cm.simClock.Stop()
	}

	// 创建新的模拟时钟
	cm.simClock = NewSimulatedClock(startTime)
	cm.simClock.Start()
	cm.currentClock = cm.simClock
	cm.useSimClock = true

	return nil
}

// DisableSimulatedClock 禁用模拟时钟，切换回系统时钟
func (cm *ClockManager) DisableSimulatedClock() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.useSimClock && cm.simClock != nil {
		cm.simClock.Stop()
	}

	cm.currentClock = cm.systemClock
	cm.useSimClock = false
}

// GetClock 获取当前使用的时钟
func (cm *ClockManager) GetClock() Clock {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.currentClock
}

// IsUsingSimulatedClock 检查是否正在使用模拟时钟
func (cm *ClockManager) IsUsingSimulatedClock() bool {
	cm.mu.RLock()
	defer cm.mu.RUnlock()
	return cm.useSimClock
}

// SetSimulatedTime 设置模拟时钟的当前时间
func (cm *ClockManager) SetSimulatedTime(t time.Time) error {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.useSimClock || cm.simClock == nil {
		return fmt.Errorf("模拟时钟未启用")
	}

	cm.simClock.SetTime(t)
	return nil
}

// GetSimulatedTime 获取模拟时钟的当前时间
func (cm *ClockManager) GetSimulatedTime() (time.Time, error) {
	cm.mu.RLock()
	defer cm.mu.RUnlock()

	if !cm.useSimClock || cm.simClock == nil {
		return time.Time{}, fmt.Errorf("模拟时钟未启用")
	}

	return cm.simClock.Now(), nil
}

// Stop 停止时钟管理器
func (cm *ClockManager) Stop() {
	cm.mu.Lock()
	defer cm.mu.Unlock()

	if cm.useSimClock && cm.simClock != nil {
		cm.simClock.Stop()
	}
}
