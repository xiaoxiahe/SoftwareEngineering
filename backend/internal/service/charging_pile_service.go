package service

import (
	"errors"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
)

// ChargingPileService 充电桩服务
type ChargingPileService struct {
	pileRepo  *repository.ChargingPileRepository
	sysRepo   *repository.SystemRepository
	userRepo  *repository.UserRepository
	queueRepo *repository.QueueRepository
}

// NewChargingPileService 创建充电桩服务
func NewChargingPileService(
	pileRepo *repository.ChargingPileRepository,
	sysRepo *repository.SystemRepository,
	userRepo *repository.UserRepository,
	queueRepo *repository.QueueRepository,
) *ChargingPileService {
	return &ChargingPileService{
		pileRepo:  pileRepo,
		sysRepo:   sysRepo,
		userRepo:  userRepo,
		queueRepo: queueRepo,
	}
}

// GetAllPiles 获取所有充电桩
func (s *ChargingPileService) GetAllPiles() ([]*model.ChargingPile, error) {
	return s.pileRepo.GetAll()
}

// GetPileByID 根据ID获取充电桩
func (s *ChargingPileService) GetPileByID(id string) (*model.ChargingPile, error) {
	pile, err := s.pileRepo.GetByID(id)
	if err != nil {
		return nil, errors.New("充电桩不存在或系统错误")
	}
	return pile, nil
}

// GetPilesByType 根据类型获取充电桩
func (s *ChargingPileService) GetPilesByType(pileType model.PileType) ([]*model.ChargingPile, error) {
	if pileType != model.PileTypeFast && pileType != model.PileTypeSlow {
		return nil, errors.New("无效的充电桩类型")
	}

	return s.pileRepo.GetByType(pileType)
}

// UpdatePileStatus 更新充电桩状态
func (s *ChargingPileService) UpdatePileStatus(id string, status model.PileStatus) error {
	// 验证充电桩存在
	_, err := s.pileRepo.GetByID(id)
	if err != nil {
		return errors.New("充电桩不存在")
	}

	// 验证状态值合法
	if !isValidPileStatus(status) {
		return errors.New("无效的充电桩状态")
	}

	return s.pileRepo.UpdateStatus(id, status)
}

// ReportPileFault 报告充电桩故障
func (s *ChargingPileService) ReportPileFault(id string, faultType string, description string) error {
	// 验证充电桩存在
	pile, err := s.pileRepo.GetByID(id)
	if err != nil {
		return errors.New("充电桩不存在")
	}

	// 检查当前是否已经处于故障状态
	if pile.Status == model.PileStatusFault {
		return errors.New("充电桩已处于故障状态")
	}

	// 更新充电桩状态为故障
	err = s.pileRepo.UpdateStatus(id, model.PileStatusFault)
	if err != nil {
		return err
	}

	// 创建故障记录
	faultRecord := &model.FaultRecord{
		PileID:      id,
		FaultType:   model.FaultType(faultType),
		Description: description,
		OccurredAt:  time.Now().UTC(),
		Status:      "active",
	}

	return s.sysRepo.CreateFaultRecord(faultRecord)
}

// RepairPile 维修充电桩
func (s *ChargingPileService) RepairPile(id string) error {
	// 验证充电桩存在
	pile, err := s.pileRepo.GetByID(id)
	if err != nil {
		return errors.New("充电桩不存在")
	}

	// 检查当前是否为故障状态
	if pile.Status != model.PileStatusFault {
		return errors.New("充电桩不处于故障状态")
	}

	// 获取活跃故障记录
	faultRecord, err := s.sysRepo.GetActiveFaultByPileID(id)
	if err != nil {
		return err
	}

	if faultRecord == nil {
		return errors.New("未找到该充电桩的活跃故障记录")
	}

	// 更新充电桩状态为空闲
	err = s.pileRepo.UpdateStatus(id, model.PileStatusAvailable)
	if err != nil {
		return err
	}

	// 更新故障记录
	now := time.Now().UTC()
	return s.sysRepo.UpdateFaultRecord(faultRecord.ID, now, 0) // 影响的会话数可以从调度服务获取
}

// UpdateQueueLength 更新充电桩队列长度
func (s *ChargingPileService) UpdateQueueLength(id string, queueLength int) error {
	// 验证充电桩存在
	_, err := s.pileRepo.GetByID(id)
	if err != nil {
		return errors.New("充电桩不存在")
	}

	// 验证队列长度非负
	if queueLength < 0 {
		return errors.New("队列长度不能为负数")
	}

	return s.pileRepo.UpdateQueueLength(id, queueLength)
}

// UpdatePileStats 更新充电桩统计信息
func (s *ChargingPileService) UpdatePileStats(id string, sessionCount int, duration, energy float64) error {
	// 验证充电桩存在
	_, err := s.pileRepo.GetByID(id)
	if err != nil {
		return errors.New("充电桩不存在")
	}

	// 验证参数合法性
	if sessionCount < 0 || duration < 0 || energy < 0 {
		return errors.New("统计参数不能为负数")
	}

	return s.pileRepo.UpdateStats(id, sessionCount, duration, energy)
}

// GetAvailablePiles 获取可用的充电桩
func (s *ChargingPileService) GetAvailablePiles(pileType model.PileType, maxQueueLength int) ([]*model.ChargingPile, error) {
	if pileType != model.PileTypeFast && pileType != model.PileTypeSlow {
		return nil, errors.New("无效的充电桩类型")
	}

	if maxQueueLength < 0 {
		return nil, errors.New("最大队列长度不能为负数")
	}

	return s.pileRepo.GetAvailablePiles(pileType, maxQueueLength)
}

// GetUserRepository 获取用户仓库
func (s *ChargingPileService) GetUserRepository() *repository.UserRepository {
	return s.userRepo
}

// GetQueueRepository 获取队列仓库
func (s *ChargingPileService) GetQueueRepository() *repository.QueueRepository {
	return s.queueRepo
}

// 辅助函数: 验证充电桩状态是否合法
func isValidPileStatus(status model.PileStatus) bool {
	validStatuses := []model.PileStatus{
		model.PileStatusAvailable,
		model.PileStatusOccupied,
		model.PileStatusFault,
		model.PileStatusMaintenance,
		model.PileStatusOffline,
	}

	for _, s := range validStatuses {
		if s == status {
			return true
		}
	}

	return false
}
