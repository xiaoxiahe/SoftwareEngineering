package service

import (
	"errors"
	"fmt"
	"sync"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

// ChargingRequestService 充电请求服务
type ChargingRequestService struct {
	requestRepo     *repository.ChargingRequestRepository
	queueRepo       *repository.QueueRepository
	pileRepo        *repository.ChargingPileRepository
	systemRepo      *repository.SystemRepository
	schedulerSvc    *SchedulerService
	fastQueueNumber int // 快充队列号计数器
	slowQueueNumber int // 慢充队列号计数器
	mutex           *sync.Mutex
}

// NewChargingRequestService 创建充电请求服务
func NewChargingRequestService(
	requestRepo *repository.ChargingRequestRepository,
	queueRepo *repository.QueueRepository,
	pileRepo *repository.ChargingPileRepository,
	systemRepo *repository.SystemRepository,
) *ChargingRequestService {
	svc := &ChargingRequestService{
		requestRepo:     requestRepo,
		queueRepo:       queueRepo,
		pileRepo:        pileRepo,
		systemRepo:      systemRepo,
		fastQueueNumber: 0,
		slowQueueNumber: 0,
		mutex:           &sync.Mutex{},
	}

	// 初始化队列号
	svc.initQueueNumbers()

	return svc
}

// SetSchedulerService 设置调度服务（避免循环依赖）
func (s *ChargingRequestService) SetSchedulerService(schedulerSvc *SchedulerService) {
	s.schedulerSvc = schedulerSvc
}

// initQueueNumbers 初始化队列号
func (s *ChargingRequestService) initQueueNumbers() {
	// 从数据库获取最大队列号
	fastRequests, err := s.requestRepo.GetWaitingRequestsByMode(model.ChargingModeFast)
	if err == nil && len(fastRequests) > 0 {
		for _, req := range fastRequests {
			if req.QueueNumber != "" {
				// 从队列号（如 F10）中提取数字部分
				var num int
				_, err := fmt.Sscanf(req.QueueNumber, "F%d", &num)
				if err == nil && num > s.fastQueueNumber {
					s.fastQueueNumber = num
				}
			}
		}
	}

	slowRequests, err := s.requestRepo.GetWaitingRequestsByMode(model.ChargingModeSlow)
	if err == nil && len(slowRequests) > 0 {
		for _, req := range slowRequests {
			if req.QueueNumber != "" {
				// 从队列号（如 T10）中提取数字部分
				var num int
				_, err := fmt.Sscanf(req.QueueNumber, "T%d", &num)
				if err == nil && num > s.slowQueueNumber {
					s.slowQueueNumber = num
				}
			}
		}
	}
}

// CreateRequest 创建充电请求
func (s *ChargingRequestService) CreateRequest(userID uuid.UUID, req *model.ChargingRequestCreate) (*model.ChargingRequest, error) {
	// 检查用户是否已有活跃请求
	activeReq, err := s.requestRepo.GetActiveRequestByUserID(userID)
	if err == nil && activeReq != nil {
		return nil, errors.New("用户已有活跃的充电请求")
	}

	// 检查等候区是否已满
	count, err := s.requestRepo.CountWaitingRequests()
	if err != nil {
		return nil, err
	}

	// 获取系统配置中的等候区容量
	config, err := s.systemRepo.GetSchedulingConfig()
	if err != nil {
		return nil, err
	}

	if count >= config.WaitingAreaSize {
		return nil, errors.New("等候区已满，请稍后再试")
	}

	// 生成队列号
	queueNumber := s.generateQueueNumber(req.ChargingMode)

	// 创建充电请求
	chargingReq := &model.ChargingRequest{
		ID:                uuid.New(),
		UserID:            userID,
		ChargingMode:      req.ChargingMode,
		RequestedCapacity: req.RequestedCapacity,
		QueueNumber:       queueNumber,
		Status:            model.RequestStatusWaiting,
		CreatedAt:         time.Now().UTC(),
		UpdatedAt:         time.Now().UTC(),
	}

	// 保存到数据库
	createdReq, err := s.requestRepo.Create(chargingReq)
	if err != nil {
		return nil, err
	}

	// 尝试进行调度
	go s.schedulerSvc.TryScheduleRequests()

	return createdReq, nil
}

// UpdateRequest 更新充电请求
func (s *ChargingRequestService) UpdateRequest(userID uuid.UUID, requestID uuid.UUID, req *model.ChargingRequestUpdate) (*model.ChargingRequest, error) {
	// 获取请求
	currentReq, err := s.requestRepo.GetByID(requestID)
	if err != nil {
		return nil, err
	}

	// 检查请求所有者
	if currentReq.UserID != userID {
		return nil, errors.New("无权修改该请求")
	}

	// 检查请求状态
	if currentReq.Status != model.RequestStatusWaiting {
		return nil, errors.New("只有等候区的请求可以修改")
	}

	// 更新充电模式
	if req.ChargingMode != "" && req.ChargingMode != currentReq.ChargingMode {
		// 生成新队列号
		queueNumber := s.generateQueueNumber(req.ChargingMode)
		currentReq.ChargingMode = req.ChargingMode
		currentReq.QueueNumber = queueNumber
	}

	// 更新请求充电量
	if req.RequestedCapacity > 0 {
		currentReq.RequestedCapacity = req.RequestedCapacity
	}

	currentReq.UpdatedAt = time.Now().UTC()

	// 保存到数据库
	err = s.requestRepo.UpdateRequest(currentReq)
	if err != nil {
		return nil, err
	}

	// 尝试进行调度
	go s.schedulerSvc.TryScheduleRequests()

	return currentReq, nil
}

// CancelRequest 取消充电请求
func (s *ChargingRequestService) CancelRequest(userID uuid.UUID, requestID uuid.UUID) error {
	// 获取请求
	req, err := s.requestRepo.GetByID(requestID)
	if err != nil {
		return err
	}

	// 检查请求所有者
	if req.UserID != userID {
		return errors.New("无权取消该请求")
	}

	// 根据请求状态执行不同处理
	switch req.Status {
	case model.RequestStatusWaiting:
		// 等候区：直接标记为已取消
		return s.requestRepo.UpdateRequestStatus(requestID, model.RequestStatusCancelled)

	case model.RequestStatusQueued:
		// 充电区排队中：从队列移除，并触发重新调度
		err := s.queueRepo.RemoveFromQueue(requestID)
		if err != nil {
			return err
		}

		err = s.requestRepo.UpdateRequestStatus(requestID, model.RequestStatusCancelled)
		if err != nil {
			return err
		}

		// 更新充电桩队列长度
		if req.PileID != "" {
			pile, err := s.pileRepo.GetByID(req.PileID)
			if err == nil && pile != nil {
				s.pileRepo.UpdateQueueLength(req.PileID, pile.QueueLength-1)
			}
		}

		// 触发重新调度
		go s.schedulerSvc.TryScheduleRequests()
		return nil

	case model.RequestStatusCharging:
		// 充电中：需要停止充电并生成详单
		// 由 BillingService 处理，这里只负责取消请求
		// 调用调度服务的StopCharging方法
		return s.schedulerSvc.StopCharging(requestID, true)

	case model.RequestStatusCompleted, model.RequestStatusCancelled:
		return errors.New("请求已完成或已取消")

	default:
		return errors.New("未知的请求状态")
	}
}

// GetActiveRequestByUserID 获取用户的活跃请求
func (s *ChargingRequestService) GetActiveRequestByUserID(userID uuid.UUID) (*model.ChargingRequest, error) {
	return s.requestRepo.GetActiveRequestByUserID(userID)
}

// GetUserRequests 获取用户的充电请求历史
func (s *ChargingRequestService) GetUserRequests(userID uuid.UUID, page, pageSize int) ([]*model.ChargingRequest, int, error) {
	return s.requestRepo.GetUserRequests(userID, page, pageSize)
}

// GetRequestByID 根据ID获取充电请求
func (s *ChargingRequestService) GetRequestByID(requestID uuid.UUID) (*model.ChargingRequest, error) {
	return s.requestRepo.GetByID(requestID)
}

// GetUserPosition 获取用户在队列中的位置
func (s *ChargingRequestService) GetUserPosition(userID uuid.UUID) (*model.UserPosition, error) {
	return s.queueRepo.GetUserPosition(userID)
}

// generateQueueNumber 生成队列号
func (s *ChargingRequestService) generateQueueNumber(mode model.ChargingMode) string {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	if mode == model.ChargingModeFast {
		s.fastQueueNumber++
		return fmt.Sprintf("F%d", s.fastQueueNumber)
	} else {
		s.slowQueueNumber++
		return fmt.Sprintf("T%d", s.slowQueueNumber)
	}
}

// GetQueueStatus 获取排队状态
func (s *ChargingRequestService) GetQueueStatus() (*model.QueueStatus, error) {
	return s.queueRepo.GetQueueStatus()
}

// GetWaitingRequestsByMode 获取特定模式的等待请求
func (s *ChargingRequestService) GetWaitingRequestsByMode(mode model.ChargingMode) ([]*model.ChargingRequest, error) {
	return s.requestRepo.GetWaitingRequestsByMode(mode)
}
