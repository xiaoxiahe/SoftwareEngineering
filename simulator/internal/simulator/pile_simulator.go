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
func NewPileSimulator(cfg *config.Config, logger *utils.Logger) *PileSimulator {
	// 创建API客户端
	apiClient := services.NewAPIClient(cfg, logger)

	// 创建充电桩服务
	pileService := services.NewPileService(cfg, apiClient, logger)

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

// SimulateChargingRequest 模拟充电请求
func (s *PileSimulator) SimulateChargingRequest(userID string, amount float64, mode string) {
	// 随机选择一个可用的充电桩
	piles := s.pileService.GetAllPiles()
	var availablePiles []*models.Pile

	for _, pile := range piles {
		status, _ := pile.GetStatus()
		if status == models.PileStatusAvailable {
			// 检查充电桩类型与请求的充电模式是否匹配
			if (mode == string(models.ChargingModeFast) && pile.Type == models.PileTypeFast) ||
				(mode == string(models.ChargingModeTrickle) && pile.Type == models.PileTypeTrickle) {
				availablePiles = append(availablePiles, pile)
			}
		}
	}

	if len(availablePiles) == 0 {
		s.logger.Warning("没有可用的充电桩满足请求: 用户=%s, 电量=%.1f, 模式=%s", userID, amount, mode)
		return
	}

	// 随机选择一个充电桩
	selectedIndex := utils.RandomInt(0, len(availablePiles)-1)
	selectedPile := availablePiles[selectedIndex]

	// 分配车辆到充电桩
	if err := s.pileService.AssignVehicle(selectedPile.ID, userID, amount, mode); err != nil {
		s.logger.Error("分配车辆到充电桩失败: %v", err)
	}
}

// TriggerFault 手动触发故障
func (s *PileSimulator) TriggerFault(pileID string, faultType string, description string, durationMinutes int) error {
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

	// 故障持续时间
	duration := time.Duration(durationMinutes) * time.Minute

	return s.pileService.TriggerFault(pileID, ft, description, duration)
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
