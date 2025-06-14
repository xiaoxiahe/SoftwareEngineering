package simulator

import (
	"sync"
	"time"

	"simulator/internal/config"
	"simulator/internal/models"
	"simulator/internal/services"
	"simulator/internal/utils"
)

// PileSimulator 充电桩模拟器
type PileSimulator struct {
	pileService *services.PileService
	apiClient   *services.APIClient
	serverAPI   *services.ServerAPI
	config      *config.Config
	logger      *utils.Logger
	isRunning   bool
	stopCh      chan struct{}
	wg          sync.WaitGroup
}

// NewPileSimulator 创建新的充电桩模拟器
func NewPileSimulator(cfg *config.Config, logger *utils.Logger, clockManager *utils.ClockManager) *PileSimulator {
	// 获取当前时钟
	clock := clockManager.GetClock()

	// 创建API客户端，传入时钟对象
	apiClient := services.NewAPIClient(cfg, logger, clock)

	// 创建充电桩服务
	pileService := services.NewPileService(cfg, apiClient, logger, clock)

	// 创建服务器API
	serverAPI := services.NewServerAPI(cfg, pileService, logger)

	return &PileSimulator{
		pileService: pileService,
		apiClient:   apiClient,
		serverAPI:   serverAPI,
		config:      cfg,
		logger:      logger,
		stopCh:      make(chan struct{}),
	}
}

// Start 启动模拟器
func (s *PileSimulator) Start() error {
	if s.isRunning {
		return nil
	}

	s.logger.Info("启动充电桩模拟器")
	s.isRunning = true

	// 初始化充电桩
	s.pileService.InitializePiles()

	// 启动心跳检测
	s.pileService.StartHeartbeat()

	// 设置充电分配回调
	s.serverAPI.SetOnChargingAssign(s.pileService.AssignVehicle)

	// 启动服务器
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()
		port := 8090 // 默认端口
		if err := s.serverAPI.Start(port); err != nil {
			s.logger.Error("服务器启动失败: %v", err)
		}
	}()

	// 启动自动故障模拟
	if s.config.Fault.RandomFault {
		s.startRandomFaultSimulation()
	}

	return nil
}

// Stop 停止模拟器
func (s *PileSimulator) Stop() error {
	if !s.isRunning {
		return nil
	}

	s.logger.Info("停止充电桩模拟器")
	close(s.stopCh)

	// 停止服务器
	if err := s.serverAPI.Stop(); err != nil {
		s.logger.Error("停止服务器失败: %v", err)
	}

	// 等待所有协程退出
	s.wg.Wait()
	s.isRunning = false
	return nil
}

// startRandomFaultSimulation 启动随机故障模拟
func (s *PileSimulator) startRandomFaultSimulation() {
	s.wg.Add(1)
	go func() {
		defer s.wg.Done()

		// 检查故障概率，避免频繁检测
		if s.config.Fault.FaultChance <= 0 {
			return
		}

		ticker := time.NewTicker(30 * time.Minute) // 每30分钟检查一次
		defer ticker.Stop()

		for {
			select {
			case <-ticker.C:
				// 这里不需要做任何事情，PileService会自动处理随机故障
			case <-s.stopCh:
				return
			}
		}
	}()
}

// TriggerFault 手动触发故障
func (s *PileSimulator) TriggerFault(pileID string, faultType string, description string) error {
	// 转换故障类型
	var ft models.FaultType
	switch faultType {
	case "hardware":
		ft = models.FaultTypeHardware
	case "software":
		ft = models.FaultTypeSoftware
	case "power":
		ft = models.FaultTypePower
	default:
		ft = models.FaultTypeHardware
	}

	return s.pileService.TriggerFault(pileID, ft, description)
}

// RecoverFault 手动恢复故障
func (s *PileSimulator) RecoverFault(pileID string) error {
	return s.pileService.RecoverFault(pileID)
}

// GetPileStatus 获取充电桩状态
func (s *PileSimulator) GetPileStatus(pileID string) (models.PileStatus, *models.ChargingVehicle, error) {
	pile, err := s.pileService.GetPile(pileID)
	if err != nil {
		return "", nil, nil
	}

	status, vehicle := pile.GetStatus()
	return status, vehicle, nil
}

// GetAllPiles 获取所有充电桩
func (s *PileSimulator) GetAllPiles() []*models.Pile {
	return s.pileService.GetAllPiles()
}
