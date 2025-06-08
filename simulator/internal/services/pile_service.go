package services

import (
	"fmt"
	"math/rand"
	"sync"
	"time"

	"simulator/internal/config"
	"simulator/internal/models"
	"simulator/internal/utils"
)

// PileService 充电桩服务
type PileService struct {
	Piles     map[string]*models.Pile
	apiClient *APIClient
	config    *config.Config
	logger    *utils.Logger
	simTimer  *utils.SimulationTimer
	stopChans map[string]chan bool
	mu        sync.Mutex
}

// NewPileService 创建充电桩服务
func NewPileService(cfg *config.Config, apiClient *APIClient, logger *utils.Logger) *PileService {
	return &PileService{
		Piles:     make(map[string]*models.Pile),
		apiClient: apiClient,
		config:    cfg,
		logger:    logger,
		simTimer:  utils.NewSimulationTimer(cfg.Simulation.SpeedFactor),
		stopChans: make(map[string]chan bool),
	}
}

// InitializePiles 初始化充电桩
func (s *PileService) InitializePiles() {
	s.mu.Lock()
	defer s.mu.Unlock()

	// 初始化快充桩
	for i := 1; i <= s.config.Piles.Fast.Count; i++ {
		id := fmt.Sprintf("F%d", i)
		pile := models.NewPile(id, models.PileTypeFast, s.config.Piles.Fast.Power)
		s.Piles[id] = pile
		s.logger.Info("初始化快充桩: %s, 功率: %.1fkW", id, pile.Power)
	}

	// 初始化慢充桩
	for i := 1; i <= s.config.Piles.Trickle.Count; i++ {
		id := fmt.Sprintf("T%d", i)
		pile := models.NewPile(id, models.PileTypeTrickle, s.config.Piles.Trickle.Power)
		s.Piles[id] = pile
		s.logger.Info("初始化慢充桩: %s, 功率: %.1fkW", id, pile.Power)
	}
}

// AssignVehicle 分配车辆到充电桩
func (s *PileService) AssignVehicle(pileID, userID string, amount float64, chargingMode string) error {
	s.mu.Lock()

	pile, exists := s.Piles[pileID]
	if !exists {
		s.mu.Unlock()
		return fmt.Errorf("充电桩 %s 不存在", pileID)
	}

	if pile.Status != models.PileStatusAvailable {
		s.mu.Unlock()
		return fmt.Errorf("充电桩 %s 当前不可用", pileID)
	}

	// 创建充电车辆
	vehicle := &models.ChargingVehicle{
		UserID:            userID,
		StartTime:         time.Now().UTC(),
		RequestedCapacity: amount,
		CurrentCapacity:   0,
		ChargingMode:      chargingMode,
	}

	// 开始充电
	if !pile.StartCharging(vehicle) {
		s.mu.Unlock()
		return fmt.Errorf("充电桩 %s 启动充电失败", pileID)
	}

	s.logger.Info("用户 %s 开始在充电桩 %s 充电，请求电量: %.1fkWh", userID, pileID, amount)

	// 释放锁后再启动充电模拟，避免死锁
	s.mu.Unlock()

	// 启动充电过程模拟
	s.startChargingSimulation(pile)

	return nil
}

// startChargingSimulation 启动充电模拟
func (s *PileService) startChargingSimulation(pile *models.Pile) {
	s.mu.Lock()

	// 如果已经有充电模拟在运行，先停止它
	if stopCh, exists := s.stopChans[pile.ID]; exists {
		close(stopCh)
		delete(s.stopChans, pile.ID)
	}

	// 创建新的停止通道
	stopCh := make(chan bool)
	s.stopChans[pile.ID] = stopCh
	s.mu.Unlock()

	// 启动充电模拟协程
	go func(pile *models.Pile, stopCh chan bool) {
		ticker := s.simTimer.NewTicker(10 * time.Second) // 每10秒更新一次
		lastUpdateTime := time.Now().UTC()

		for {
			select {
			case <-ticker.C:
				now := time.Now().UTC()
				elapsed := now.Sub(lastUpdateTime)
				lastUpdateTime = now

				// 更新充电进度
				pile.UpdateChargingProgress(elapsed)

				// 上报充电进度
				if err := s.apiClient.UpdateChargingProgress(pile); err != nil {
					s.logger.Error("上报充电进度失败: %v", err)
				}

				// 检查是否充电完成
				if pile.IsChargingComplete() {
					s.completeCharging(pile)
					return
				}

				// 随机故障处理
				if s.config.Fault.RandomFault && rand.Float64() < s.config.Fault.FaultChance/100.0 {
					s.randomFault(pile)
					return
				}

			case <-stopCh:
				s.logger.Info("充电桩 %s 充电过程被中断", pile.ID)
				return
			}
		}
	}(pile, stopCh)
}

// completeCharging 完成充电过程
func (s *PileService) completeCharging(pile *models.Pile) {
	// 停止充电模拟
	s.mu.Lock()
	if stopCh, exists := s.stopChans[pile.ID]; exists {
		close(stopCh)
		delete(s.stopChans, pile.ID)
	}
	s.mu.Unlock()

	// 停止充电
	vehicle := pile.StopCharging()
	if vehicle == nil {
		s.logger.Error("充电桩 %s 没有正在充电的车辆", pile.ID)
		return
	}

	s.logger.Info("用户 %s 在充电桩 %s 充电完成，充电量: %.1fkWh",
		vehicle.UserID, pile.ID, vehicle.CurrentCapacity)

	// 上报充电完成
	if err := s.apiClient.CompleteCharging(pile, vehicle); err != nil {
		s.logger.Error("上报充电完成失败: %v", err)
	}
}

// randomFault 随机故障模拟
func (s *PileService) randomFault(pile *models.Pile) {
	// 停止充电模拟
	s.mu.Lock()
	if stopCh, exists := s.stopChans[pile.ID]; exists {
		close(stopCh)
		delete(s.stopChans, pile.ID)
	}
	s.mu.Unlock()

	// 故障类型
	faultTypes := []models.FaultType{
		models.FaultTypeHardware,
		models.FaultTypeSoftware,
		models.FaultTypePower,
	}

	faultType := faultTypes[rand.Intn(len(faultTypes))]
	description := fmt.Sprintf("%s故障 - 随机生成", faultType)

	// 故障持续时间
	minTime := s.config.Fault.MinFaultTime
	maxTime := s.config.Fault.MaxFaultTime
	faultDuration := time.Duration(rand.Intn(maxTime-minTime+1)+minTime) * time.Minute

	// 报告故障
	pile.ReportFault(faultType, description, faultDuration)

	s.logger.Warning("充电桩 %s 发生%s故障，预计恢复时间: %d分钟后",
		pile.ID, faultType, int(faultDuration.Minutes()))

	// 上报故障
	if err := s.apiClient.ReportFault(pile, faultType, description); err != nil {
		s.logger.Error("上报故障失败: %v", err)
	}

	go func(pile *models.Pile, duration time.Duration) {
		time.Sleep(duration)

		s.mu.Lock()
		defer s.mu.Unlock()

		pile.RecoverFromFault()
		s.logger.Info("充电桩 %s 故障已恢复", pile.ID)

		// 上报故障恢复
		if err := s.apiClient.RecoverFault(pile); err != nil {
			s.logger.Error("上报故障恢复失败: %v", err)
		}
	}(pile, faultDuration)
}

// TriggerFault 手动触发故障
func (s *PileService) TriggerFault(pileID string, faultType models.FaultType, description string, duration time.Duration) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pile, exists := s.Piles[pileID]
	if !exists {
		return fmt.Errorf("充电桩 %s 不存在", pileID)
	}

	// 停止充电模拟
	if stopCh, exists := s.stopChans[pileID]; exists {
		close(stopCh)
		delete(s.stopChans, pileID)
	}

	// 报告故障
	pile.ReportFault(faultType, description, duration)

	s.logger.Warning("充电桩 %s 手动触发%s故障，预计恢复时间: %d分钟后",
		pileID, faultType, int(duration.Minutes()))

	// 上报故障
	if err := s.apiClient.ReportFault(pile, faultType, description); err != nil {
		s.logger.Error("上报故障失败: %v", err)
	}
	// 启动故障恢复定时器
	go func(pile *models.Pile, duration time.Duration) {
		time.Sleep(duration)

		s.mu.Lock()
		defer s.mu.Unlock()

		pile.RecoverFromFault()
		s.logger.Info("充电桩 %s 故障已恢复", pile.ID)

		// 上报故障恢复
		if err := s.apiClient.RecoverFault(pile); err != nil {
			s.logger.Error("上报故障恢复失败: %v", err)
		}
	}(pile, duration)

	return nil
}

// RecoverFault 手动恢复故障
func (s *PileService) RecoverFault(pileID string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pile, exists := s.Piles[pileID]
	if !exists {
		return fmt.Errorf("充电桩 %s 不存在", pileID)
	}

	if pile.Status != models.PileStatusFault {
		return fmt.Errorf("充电桩 %s 当前不是故障状态", pileID)
	}
	pile.RecoverFromFault()
	s.logger.Info("充电桩 %s 故障已手动恢复", pileID)

	// 上报故障恢复
	if err := s.apiClient.RecoverFault(pile); err != nil {
		s.logger.Error("上报故障恢复失败: %v", err)
	}

	return nil
}

// StartHeartbeat 开始心跳
func (s *PileService) StartHeartbeat() {
	interval := time.Duration(s.config.BackendAPI.HeartbeatInterval) * time.Second
	go func() {
		ticker := time.NewTicker(interval)
		for range ticker.C {
			s.sendHeartbeat()
		}
	}()
	s.logger.Info("启动心跳检测，间隔: %d秒", s.config.BackendAPI.HeartbeatInterval)
}

// sendHeartbeat 发送心跳
func (s *PileService) sendHeartbeat() {
	s.mu.Lock()
	pileIDs := make([]string, 0, len(s.Piles))
	for id := range s.Piles {
		pileIDs = append(pileIDs, id)
	}
	s.mu.Unlock()

	if err := s.apiClient.SendHeartbeat(pileIDs); err != nil {
		s.logger.Error("发送心跳失败: %v", err)
	}
}

// StopCharging 停止指定充电桩的充电
func (s *PileService) StopCharging(pileID, userID string, reason string) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	pile, exists := s.Piles[pileID]
	if !exists {
		return fmt.Errorf("充电桩 %s 不存在", pileID)
	}

	status, vehicle := pile.GetStatus()
	if status != models.PileStatusCharging || vehicle == nil {
		return fmt.Errorf("充电桩 %s 当前没有在充电", pileID)
	}

	// 验证用户ID
	if vehicle.UserID != userID {
		return fmt.Errorf("用户ID不匹配，当前充电用户: %s, 请求用户: %s", vehicle.UserID, userID)
	}

	// 停止充电模拟
	if stopCh, exists := s.stopChans[pileID]; exists {
		close(stopCh)
		delete(s.stopChans, pileID)
	}

	// 停止充电
	stoppedVehicle := pile.StopCharging()
	if stoppedVehicle == nil {
		return fmt.Errorf("停止充电失败")
	}

	s.logger.Info("用户 %s 在充电桩 %s 的充电被停止，原因: %s，已充电量: %.1fkWh",
		userID, pileID, reason, stoppedVehicle.CurrentCapacity)

	return nil
}

// GetPile 获取充电桩
func (s *PileService) GetPile(pileID string) (*models.Pile, error) {
	s.mu.Lock()
	defer s.mu.Unlock()

	pile, exists := s.Piles[pileID]
	if !exists {
		return nil, fmt.Errorf("充电桩 %s 不存在", pileID)
	}

	return pile, nil
}

// GetAllPiles 获取所有充电桩
func (s *PileService) GetAllPiles() []*models.Pile {
	s.mu.Lock()
	defer s.mu.Unlock()

	piles := make([]*models.Pile, 0, len(s.Piles))
	for _, pile := range s.Piles {
		piles = append(piles, pile)
	}

	return piles
}
