package service

import (
	"fmt"
	"log"
	"sort"
	"sync"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

// SchedulerService 调度服务
type SchedulerService struct {
	requestRepo      *repository.ChargingRequestRepository
	pileRepo         *repository.ChargingPileRepository
	queueRepo        *repository.QueueRepository
	sessionRepo      *repository.ChargingSessionRepository
	systemRepo       *repository.SystemRepository
	billingService   *BillingService
	simulatorClient  *ChargingDispatcherClient // 模拟器客户端
	waitingAreaLock  bool                      // 等候区锁定状态
	requestChan      chan uuid.UUID            // 请求调度通道
	stopChargingChan chan stopChargingReq      // 停止充电通道
	mutex            *sync.Mutex
}

// stopChargingReq 停止充电请求
type stopChargingReq struct {
	requestID uuid.UUID
	cancel    bool
}

// NewSchedulerService 创建调度服务
func NewSchedulerService(
	requestRepo *repository.ChargingRequestRepository,
	pileRepo *repository.ChargingPileRepository,
	queueRepo *repository.QueueRepository,
	sessionRepo *repository.ChargingSessionRepository,
	systemRepo *repository.SystemRepository,
) *SchedulerService {
	svc := &SchedulerService{
		requestRepo:      requestRepo,
		pileRepo:         pileRepo,
		queueRepo:        queueRepo,
		sessionRepo:      sessionRepo,
		systemRepo:       systemRepo,
		waitingAreaLock:  false,
		requestChan:      make(chan uuid.UUID, 100),
		stopChargingChan: make(chan stopChargingReq, 100),
		mutex:            &sync.Mutex{},
	}

	// 启动调度器
	go svc.schedulerLoop()
	go svc.stopChargingLoop()

	return svc
}

// SetBillingService 设置计费服务（避免循环依赖）
func (s *SchedulerService) SetBillingService(billingService *BillingService) {
	s.billingService = billingService
}

// SetSimulatorClient 设置模拟器客户端（用于向模拟器发送充电指令）
func (s *SchedulerService) SetSimulatorClient(client *ChargingDispatcherClient) {
	s.simulatorClient = client
}

// TryScheduleRequests 尝试调度请求
func (s *SchedulerService) TryScheduleRequests() {
	// 由于这个方法是非阻塞的，只是触发调度过程
	// 创建一个空UUID，表示不是针对特定请求的调度
	s.requestChan <- uuid.Nil
}

// StopCharging 停止充电
func (s *SchedulerService) StopCharging(requestID uuid.UUID, cancel bool) error {
	s.stopChargingChan <- stopChargingReq{
		requestID: requestID,
		cancel:    cancel,
	}
	return nil
}

// UpdateChargingProgress 更新充电进度
func (s *SchedulerService) UpdateChargingProgress(pileID, userID string, currentCapacity float64, remainingTime int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取充电桩的充电会话
	session, err := s.sessionRepo.GetActiveSessionByPileID(pileID)
	if err != nil {
		return fmt.Errorf("获取充电会话失败: %w", err)
	}

	// 验证充电会话是否存在
	if session == nil {
		return fmt.Errorf("充电桩 %s 没有活跃的充电会话", pileID)
	}

	// 验证用户ID是否匹配
	if session.UserID.String() != userID {
		return fmt.Errorf("用户ID不匹配: 会话用户=%s, 请求用户=%s", session.UserID, userID)
	}

	// 更新充电会话的充电量
	session.ActualCapacity = currentCapacity

	// 保存更新的会话
	err = s.sessionRepo.Update(session)
	if err != nil {
		return fmt.Errorf("更新充电会话失败: %w", err)
	}

	log.Printf("已更新充电进度: 充电桩=%s, 用户=%s, 当前电量=%.1fkWh, 剩余时间=%d秒",
		pileID, userID, currentCapacity, remainingTime)

	return nil
}

// 调度器主循环
func (s *SchedulerService) schedulerLoop() {
	for range s.requestChan {
		if s.waitingAreaLock {
			continue
		}

		// 执行调度
		s.executeSchedule()
	}
}

// 停止充电处理循环
func (s *SchedulerService) stopChargingLoop() {
	for req := range s.stopChargingChan {
		s.executeStopCharging(req.requestID, req.cancel)
	}
}

// executeSchedule 执行调度
func (s *SchedulerService) executeSchedule() {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取系统配置
	config, err := s.systemRepo.GetSchedulingConfig()
	if err != nil {
		log.Printf("获取系统配置失败: %v", err)
		return
	}

	// 调度流程
	if config.ExtendedSchedulingMode == model.ExtendedModeBatch {
		s.checkAndExecuteBatchScheduling(config)
	} else {
		s.executeNormalScheduling(config)
	}
}

// executeNormalScheduling 执行正常调度
func (s *SchedulerService) executeNormalScheduling(config *model.SchedulingConfig) {
	// 获取等候区请求
	fastRequests, err := s.requestRepo.GetWaitingRequestsByMode(model.ChargingModeFast)
	if err != nil {
		log.Printf("获取快充请求失败: %v", err)
		return
	}

	slowRequests, err := s.requestRepo.GetWaitingRequestsByMode(model.ChargingModeSlow)
	if err != nil {
		log.Printf("获取慢充请求失败: %v", err)
		return
	}
	// 根据配置的调度策略排序请求
	s.sortRequestsByStrategy(fastRequests, config.FaultRescheduling)
	s.sortRequestsByStrategy(slowRequests, config.FaultRescheduling)

	// 获取可用的充电桩
	fastPiles, err := s.pileRepo.GetAvailablePiles(model.PileTypeFast, config.ChargingQueueLen)
	if err != nil {
		log.Printf("获取快充桩失败: %v", err)
		return
	}

	slowPiles, err := s.pileRepo.GetAvailablePiles(model.PileTypeSlow, config.ChargingQueueLen)
	if err != nil {
		log.Printf("获取慢充桩失败: %v", err)
		return
	}

	// 快充调度 - 按完成时长最短策略
	for len(fastRequests) > 0 {
		req := fastRequests[0]
		bestPile := s.findBestPile(fastPiles, req.RequestedCapacity, config.ChargingQueueLen)
		if bestPile == nil {
			break // 没有可用充电桩
		}

		fastRequests = fastRequests[1:]
		s.scheduleRequestToPile(req.ID, bestPile.ID, bestPile.QueueLength+1)

		// 更新本地充电桩队列长度以便下次计算
		bestPile.QueueLength++
	}

	// 慢充调度 - 按完成时长最短策略
	for len(slowRequests) > 0 {
		req := slowRequests[0]
		bestPile := s.findBestPile(slowPiles, req.RequestedCapacity, config.ChargingQueueLen)
		if bestPile == nil {
			break // 没有可用充电桩
		}

		slowRequests = slowRequests[1:]
		s.scheduleRequestToPile(req.ID, bestPile.ID, bestPile.QueueLength+1)

		// 更新本地充电桩队列长度以便下次计算
		bestPile.QueueLength++
	}
}

// sortRequestsByStrategy 根据配置的调度策略对请求进行排序
func (s *SchedulerService) sortRequestsByStrategy(requests []*model.ChargingRequest, strategy model.FaultReschedulingStrategy) {
	if strategy == model.FaultStrategyPriority {
		// 优先级调度：高优先级优先，同优先级按队列号(时间)排序
		sort.Slice(requests, func(i, j int) bool {
			if requests[i].Priority != requests[j].Priority {
				return requests[i].Priority > requests[j].Priority
			}
			return requests[i].QueueNumber < requests[j].QueueNumber
		})
	} else {
		// 时间顺序调度：纯粹按队列号(时间)排序
		sort.Slice(requests, func(i, j int) bool {
			return requests[i].QueueNumber < requests[j].QueueNumber
		})
	}
}

// findBestPile 找到完成充电所需时长最短的充电桩
func (s *SchedulerService) findBestPile(piles []*model.ChargingPile, requestedCapacity float64, maxQueueLen int) *model.ChargingPile {
	var bestPile *model.ChargingPile
	var minCompletionTime float64 = -1

	for _, pile := range piles {
		// 检查是否有空位
		if pile.QueueLength >= maxQueueLen {
			continue
		}

		// 计算完成充电所需时长 = 等待时间 + 自己充电时间
		waitTime := s.calculateWaitTime(pile.ID)
		selfChargingTime := requestedCapacity / pile.Power * 3600 // 转换为秒
		completionTime := waitTime + selfChargingTime

		if minCompletionTime < 0 || completionTime < minCompletionTime {
			minCompletionTime = completionTime
			bestPile = pile
		}
	}

	return bestPile
}

// calculateWaitTime 计算等待时间（队列中所有车辆完成充电时间之和）
func (s *SchedulerService) calculateWaitTime(pileID string) float64 {
	// 获取当前充电桩队列中的所有请求
	requests, err := s.requestRepo.GetRequestsByPile(pileID)
	if err != nil {
		log.Printf("获取充电桩请求失败: %v", err)
		return 0
	}

	// 获取充电桩信息
	pile, err := s.pileRepo.GetByID(pileID)
	if err != nil {
		log.Printf("获取充电桩失败: %v", err)
		return 0
	}

	var totalWaitTime float64 = 0
	for _, req := range requests {
		// 计算每个请求的充电时间（秒）
		chargingTime := req.RequestedCapacity / pile.Power * 3600
		totalWaitTime += chargingTime
	}

	return totalWaitTime
}

// scheduleRequestToPile 将请求调度到特定充电桩
func (s *SchedulerService) scheduleRequestToPile(requestID uuid.UUID, pileID string, queuePosition int) {
	// 获取请求
	request, err := s.requestRepo.GetByID(requestID)
	if err != nil {
		log.Printf("获取请求失败: %v", err)
		return
	}

	// 获取充电桩
	pile, err := s.pileRepo.GetByID(pileID)
	if err != nil {
		log.Printf("获取充电桩失败: %v", err)
		return
	}

	// 计算预估等待时间
	waitTime := s.calculateEstimatedWaitTime(pileID, request.RequestedCapacity, pile.Power)

	// 更新请求状态
	err = s.requestRepo.AssignToPile(requestID, pileID, queuePosition, waitTime, model.RequestStatusQueued)
	if err != nil {
		log.Printf("更新请求状态失败: %v", err)
		return
	}

	// 更新充电桩队列长度
	err = s.pileRepo.UpdateQueueLength(pileID, pile.QueueLength+1)
	if err != nil {
		log.Printf("更新充电桩队列长度失败: %v", err)
		return
	}

	// 添加到队列
	queueItem := &model.QueueItem{
		PileID:            pileID,
		Position:          queuePosition,
		RequestID:         requestID,
		UserID:            request.UserID,
		QueueNumber:       request.QueueNumber,
		ChargingMode:      request.ChargingMode,
		RequestedCapacity: request.RequestedCapacity,
		EnterTime:         time.Now().UTC(),
	}

	err = s.queueRepo.AddToQueue(queueItem)
	if err != nil {
		log.Printf("添加到队列失败: %v", err)
		return
	}

	// 如果是第一个位置，开始充电
	if queuePosition == 1 {
		s.startCharging(requestID, pileID)
	}
}

// calculateEstimatedWaitTime 计算预估等待时间（秒）
func (s *SchedulerService) calculateEstimatedWaitTime(pileID string, requestedCapacity float64, pilePower float64) int {
	// 获取当前充电桩队列中的所有请求
	requests, err := s.requestRepo.GetRequestsByPile(pileID)
	if err != nil {
		log.Printf("获取充电桩请求失败: %v", err)
		return 0
	}

	// 计算前面所有请求的充电时间总和
	var totalWaitTime int
	for _, req := range requests {
		// 计算充电时间（秒）
		chargingTime := int(req.RequestedCapacity / pilePower * 3600)
		totalWaitTime += chargingTime
	}

	// 加上该请求自身的充电时间
	selfChargingTime := int(requestedCapacity / pilePower * 3600)

	return totalWaitTime + selfChargingTime
}

// startCharging 开始充电
func (s *SchedulerService) startCharging(requestID uuid.UUID, pileID string) {
	// 获取请求
	request, err := s.requestRepo.GetByID(requestID)
	if err != nil {
		log.Printf("获取请求失败: %v", err)
		return
	}

	// 更新请求状态
	err = s.requestRepo.UpdateRequestStatus(requestID, model.RequestStatusCharging)
	if err != nil {
		log.Printf("更新请求状态失败: %v", err)
		return
	}

	// 更新充电桩状态
	err = s.pileRepo.UpdateStatus(pileID, model.PileStatusOccupied)
	if err != nil {
		log.Printf("更新充电桩状态失败: %v", err)
		return
	}

	// 更新队列中的开始充电时间
	err = s.queueRepo.SetStartCharging(requestID)
	if err != nil {
		log.Printf("更新队列开始充电时间失败: %v", err)
		return
	}

	// 创建充电会话
	session := &model.ChargingSession{
		ID:                uuid.New(),
		RequestID:         requestID,
		UserID:            request.UserID,
		PileID:            pileID,
		QueueNumber:       request.QueueNumber,
		RequestedCapacity: request.RequestedCapacity,
		ActualCapacity:    0,
		StartTime:         time.Now().UTC(),
		Status:            model.SessionStatusActive,
		Duration:          0,
	}

	_, err = s.sessionRepo.Create(session)
	if err != nil {
		log.Printf("创建充电会话失败: %v", err)
		return
	}

	// 向模拟器发送充电指令
	if s.simulatorClient != nil {
		// 根据充电模式确定传递给模拟器的模式参数
		chargingMode := "trickle" // 默认为慢充
		if request.ChargingMode == model.ChargingModeFast {
			chargingMode = "fast"
		}

		// 发送充电指令到模拟器
		err = s.simulatorClient.AssignCharging(
			pileID,
			request.UserID.String(),
			request.RequestedCapacity,
			chargingMode,
		)
		if err != nil {
			log.Printf("向模拟器发送充电指令失败: %v", err)
			// 即使发送失败，也不中断充电流程，仅记录日志
		} else {
			log.Printf("成功向模拟器发送充电指令: 充电桩=%s, 用户=%s, 电量=%.1f, 模式=%s",
				pileID, request.UserID.String(), request.RequestedCapacity, chargingMode)
		}
	} else {
		log.Printf("模拟器客户端未配置，跳过发送充电指令")
	}
}

// executeStopCharging 执行停止充电
func (s *SchedulerService) executeStopCharging(requestID uuid.UUID, cancel bool) {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取请求
	request, err := s.requestRepo.GetByID(requestID)
	if err != nil {
		log.Printf("获取请求失败: %v", err)
		return
	}

	// 检查请求状态
	if request.Status != model.RequestStatusCharging {
		log.Printf("请求不处于充电状态: %v", requestID)
		return
	}
	pileID := request.PileID

	// 向模拟器发送停止充电指令
	if s.simulatorClient != nil {
		reason := "正常停止"
		if cancel {
			reason = "用户取消"
		}

		err = s.simulatorClient.StopCharging(pileID, request.UserID.String(), reason)
		if err != nil {
			log.Printf("向模拟器发送停止充电指令失败: %v", err)
			// 即使发送失败，也不中断停止充电流程，仅记录日志
		} else {
			log.Printf("成功向模拟器发送停止充电指令: 充电桩=%s, 用户=%s, 原因=%s",
				pileID, request.UserID.String(), reason)
		}
	} else {
		log.Printf("模拟器客户端未配置，跳过发送停止充电指令")
	}

	// 获取充电会话
	session, err := s.sessionRepo.GetByRequestID(requestID)
	if err != nil {
		log.Printf("获取充电会话失败: %v", err)
		return
	}

	// 更新会话状态
	now := time.Now().UTC()
	session.EndTime = &now
	session.Duration = now.Sub(session.StartTime).Seconds() // 获取充电桩信息（用于后续更新统计和队列）
	pile, err := s.pileRepo.GetByID(pileID)
	if err != nil {
		log.Printf("获取充电桩失败: %v", err)
		return
	}

	// 计算充电时长（小时），用于统计
	chargingHours := float64(session.Duration) / 3600

	// 保持当前的实际充电量，不重新计算
	// ActualCapacity 在充电过程中通过 UpdateChargingProgress 方法持续更新
	// 这里只需要确保不超过请求的充电量
	if session.ActualCapacity > session.RequestedCapacity {
		session.ActualCapacity = session.RequestedCapacity
	}

	// 根据充电量判断状态
	if session.ActualCapacity >= session.RequestedCapacity {
		session.Status = model.SessionStatusCompleted
	} else {
		session.Status = model.SessionStatusInterrupted
	}

	// 更新会话
	err = s.sessionRepo.Update(session)
	if err != nil {
		log.Printf("更新充电会话失败: %v", err)
		return
	}

	// 更新请求状态
	var status model.RequestStatus
	if cancel {
		status = model.RequestStatusCancelled
	} else {
		status = model.RequestStatusCompleted
	}

	err = s.requestRepo.UpdateRequestStatus(requestID, status)
	if err != nil {
		log.Printf("更新充电请求状态失败: %v", err)
		return
	}

	// 从队列中移除
	err = s.queueRepo.RemoveFromQueue(requestID)
	if err != nil {
		log.Printf("从队列中移除失败: %v", err)
		return
	}

	// 更新充电桩状态和队列长度
	err = s.pileRepo.UpdateStatus(pileID, model.PileStatusAvailable)
	if err != nil {
		log.Printf("更新充电桩状态失败: %v", err)
		return
	}

	err = s.pileRepo.UpdateQueueLength(pileID, pile.QueueLength-1)
	if err != nil {
		log.Printf("更新充电桩队列长度失败: %v", err)
		return
	}

	// 更新充电桩统计信息
	err = s.pileRepo.UpdateStats(pileID, 1, chargingHours, session.ActualCapacity)
	if err != nil {
		log.Printf("更新充电桩统计信息失败: %v", err)
		return
	}

	// 生成详单
	if s.billingService != nil {
		_, err = s.billingService.GenerateBill(session.ID)
		if err != nil {
			log.Printf("生成详单失败: %v", err)
		}
	}

	// 检查是否有下一个请求
	queueItems, err := s.queueRepo.GetQueueItemsByPile(pileID)
	if err != nil {
		log.Printf("获取队列项失败: %v", err)
		return
	}

	// 重新排序队列
	for i, item := range queueItems {
		// 更新位置
		newPosition := i + 1
		if item.Position != newPosition {
			err = s.queueRepo.UpdateQueuePosition(pileID, item.RequestID, newPosition)
			if err != nil {
				log.Printf("更新队列位置失败: %v", err)
				continue
			}

			err = s.requestRepo.AssignToPile(item.RequestID, pileID, newPosition, 0, model.RequestStatusQueued)
			if err != nil {
				log.Printf("更新请求队列位置失败: %v", err)
				continue
			}
		}
		// 如果新位置是1，开始充电
		if newPosition == 1 {
			s.startCharging(item.RequestID, pileID)
		}
	}

	// 在释放锁后触发调度，避免死锁
	defer func() {
		go s.TryScheduleRequests()
	}()
}

// HandlePileFault 处理充电桩故障
func (s *SchedulerService) HandlePileFault(pileID string, faultType string, description string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Printf("处理充电桩故障: %s, 类型: %s, 描述: %s", pileID, faultType, description)

	// 更新充电桩状态为故障
	err := s.pileRepo.UpdateStatus(pileID, model.PileStatusFault)
	if err != nil {
		return fmt.Errorf("更新充电桩状态失败: %w", err)
	}

	// 获取该充电桩的所有队列请求（包括正在充电和排队的）
	queuedRequests, err := s.requestRepo.GetRequestsByPile(pileID)
	if err != nil {
		log.Printf("获取充电桩队列失败: %v", err)
	} else {
		// 将故障队列中的请求移除，放回到等待区，并给予高优先级
		for _, req := range queuedRequests {
			// 设置高优先级
			req.Priority = model.PriorityFault
			err := s.requestRepo.UpdateRequestPriority(req.ID, req.Priority)
			if err != nil {
				log.Printf("更新请求优先级失败: %v", err)
			}

			// 将请求状态重置为等候
			err = s.requestRepo.UpdateRequestStatus(req.ID, model.RequestStatusWaiting)
			if err != nil {
				log.Printf("重置请求状态失败: %v", err)
				continue
			}

			// 清除充电桩分配
			err = s.requestRepo.AssignToPile(req.ID, "", 0, 0, model.RequestStatusWaiting)
			if err != nil {
				log.Printf("清除充电桩分配失败: %v", err)
				continue
			}

			// 从队列中移除
			err = s.queueRepo.RemoveFromQueue(req.ID)
			if err != nil {
				log.Printf("从队列移除失败: %v", err)
			}

			log.Printf("故障请求 %s 已移回等待区并设置高优先级", req.ID)
		}
	}

	// 获取该充电桩上正在充电的会话
	session, err := s.sessionRepo.GetActiveSessionByPileID(pileID)
	if err != nil {
		if err.Error() != "没有活跃的充电会话" {
			log.Printf("获取充电会话失败: %v", err)
		}
	} else if session != nil { // 停止当前充电会话
		now := time.Now().UTC()
		session.EndTime = &now
		session.Duration = now.Sub(session.StartTime).Seconds()
		session.Status = model.SessionStatusInterrupted

		// 保持当前的实际充电量，不重新计算
		// ActualCapacity 在充电过程中通过 UpdateChargingProgress 方法持续更新
		// 故障时保持当前的充电量即可
		if session.ActualCapacity > session.RequestedCapacity {
			session.ActualCapacity = session.RequestedCapacity
		}

		// 更新会话
		err = s.sessionRepo.Update(session)
		if err != nil {
			log.Printf("更新充电会话失败: %v", err)
		}

		// 更新请求状态为等候重新调度，并设置高优先级
		err = s.requestRepo.UpdateRequestStatus(session.RequestID, model.RequestStatusWaiting)
		if err != nil {
			log.Printf("更新请求状态失败: %v", err)
		} else {
			// 为正在充电的请求也设置高优先级
			err = s.requestRepo.UpdateRequestPriority(session.RequestID, model.PriorityFault)
			if err != nil {
				log.Printf("更新正在充电请求的优先级失败: %v", err)
			}
		}

		// 生成部分详单
		if s.billingService != nil {
			_, err = s.billingService.GenerateBill(session.ID)
			if err != nil {
				log.Printf("生成部分详单失败: %v", err)
			}
		}
	}
	// 重置充电桩队列长度
	err = s.pileRepo.UpdateQueueLength(pileID, 0)
	if err != nil {
		log.Printf("重置故障充电桩队列长度失败: %v", err)
	}

	// 在释放锁后触发故障调度，避免死锁
	// 使用 defer 确保在方法返回前执行
	defer func() {
		go s.TryScheduleRequests()
	}()

	return nil
}

// ExecuteBatchScheduling 执行批量调度总充电时长最短
func (s *SchedulerService) ExecuteBatchScheduling() error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Printf("执行批量调度")

	// 获取系统配置
	config, err := s.systemRepo.GetSchedulingConfig()
	if err != nil {
		return fmt.Errorf("获取系统配置失败: %w", err)
	}

	// 获取可用的快充桩和慢充桩
	availableFastPiles, err := s.pileRepo.GetAvailablePiles(model.PileTypeFast, config.ChargingQueueLen)
	if err != nil {
		return fmt.Errorf("获取可用快充桩失败: %w", err)
	}

	availableSlowPiles, err := s.pileRepo.GetAvailablePiles(model.PileTypeSlow, config.ChargingQueueLen)
	if err != nil {
		return fmt.Errorf("获取可用慢充桩失败: %w", err)
	}

	// 合并所有可用充电桩
	var availablePiles []*model.ChargingPile
	availablePiles = append(availablePiles, availableFastPiles...)
	availablePiles = append(availablePiles, availableSlowPiles...)

	// 基于可用充电桩计算总车位数
	totalSlots := len(availableFastPiles)*config.ChargingQueueLen +
		len(availableSlowPiles)*config.ChargingQueueLen

	// 获取所有等候区请求
	fastRequests, err := s.requestRepo.GetWaitingRequestsByMode(model.ChargingModeFast)
	if err != nil {
		return fmt.Errorf("获取快充请求失败: %w", err)
	}

	slowRequests, err := s.requestRepo.GetWaitingRequestsByMode(model.ChargingModeSlow)
	if err != nil {
		return fmt.Errorf("获取慢充请求失败: %w", err)
	}

	allRequests := append(fastRequests, slowRequests...)
	if len(allRequests) < totalSlots {
		return fmt.Errorf("等候区车辆数量不足: 需要%d辆，实际%d辆 (基于可用充电桩: 快充%d个, 慢充%d个)",
			totalSlots, len(allRequests), len(availableFastPiles), len(availableSlowPiles))
	}
	// 根据配置的调度策略排序，取前totalSlots个
	s.sortRequestsByStrategy(allRequests, config.FaultRescheduling)
	selectedRequests := allRequests[:totalSlots]

	// 计算最优分配方案（忽略充电模式限制）
	assignment := s.calculateGlobalOptimalAssignment(selectedRequests, availablePiles, config.ChargingQueueLen)

	// 执行分配
	for requestID, pileID := range assignment {
		pile, err := s.pileRepo.GetByID(pileID)
		if err != nil {
			log.Printf("获取充电桩失败: %v", err)
			continue
		}
		s.scheduleRequestToPile(requestID, pileID, pile.QueueLength+1)
	}

	return nil
}

// calculateGlobalOptimalAssignment 计算全局最优分配方案（忽略充电模式）
func (s *SchedulerService) calculateGlobalOptimalAssignment(requests []*model.ChargingRequest, piles []*model.ChargingPile, maxQueueLen int) map[uuid.UUID]string {
	assignment := make(map[uuid.UUID]string)

	// 简化版本：使用贪心算法为每个请求找到最佳充电桩
	for _, req := range requests {
		var bestPile *model.ChargingPile
		var minCompletionTime float64 = -1

		for _, pile := range piles {
			if pile.QueueLength >= maxQueueLen {
				continue
			}

			// 计算完成时间
			waitTime := s.calculateWaitTimeForPile(pile)
			selfChargingTime := req.RequestedCapacity / pile.Power * 3600
			completionTime := waitTime + selfChargingTime

			if minCompletionTime < 0 || completionTime < minCompletionTime {
				minCompletionTime = completionTime
				bestPile = pile
			}
		}

		if bestPile != nil {
			assignment[req.ID] = bestPile.ID
			bestPile.QueueLength++ // 模拟更新队列长度
		}
	}

	return assignment
}

// calculateWaitTimeForPile 计算充电桩的等待时间
func (s *SchedulerService) calculateWaitTimeForPile(pile *model.ChargingPile) float64 {
	// 获取该充电桩队列中的所有请求
	requests, err := s.requestRepo.GetRequestsByPile(pile.ID)
	if err != nil {
		return 0
	}

	var totalWaitTime float64 = 0
	for _, req := range requests {
		chargingTime := req.RequestedCapacity / pile.Power * 3600
		totalWaitTime += chargingTime
	}

	return totalWaitTime
}

// CompleteCharging 处理充电完成
func (s *SchedulerService) CompleteCharging(pileID, userID string, startTime, endTime time.Time, requestedCapacity, actualCapacity float64, chargingDuration int) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	// 获取充电桩的活跃充电会话
	session, err := s.sessionRepo.GetActiveSessionByPileID(pileID)
	if err != nil {
		return fmt.Errorf("获取充电会话失败: %w", err)
	}

	// 验证充电会话是否存在
	if session == nil {
		return fmt.Errorf("充电桩 %s 没有活跃的充电会话", pileID)
	}

	// 验证用户ID是否匹配
	if session.UserID.String() != userID {
		return fmt.Errorf("用户ID不匹配: 会话用户=%s, 请求用户=%s", session.UserID, userID)
	}

	// 更新充电会话信息
	session.EndTime = &endTime
	session.ActualCapacity = actualCapacity
	session.Duration = float64(chargingDuration)
	session.Status = model.SessionStatusCompleted

	// 保存更新的会话
	err = s.sessionRepo.Update(session)
	if err != nil {
		return fmt.Errorf("更新充电会话失败: %w", err)
	}

	// 更新请求状态为已完成
	err = s.requestRepo.UpdateRequestStatus(session.RequestID, model.RequestStatusCompleted)
	if err != nil {
		return fmt.Errorf("更新充电请求状态失败: %w", err)
	}

	// 从队列中移除
	err = s.queueRepo.RemoveFromQueue(session.RequestID)
	if err != nil {
		return fmt.Errorf("从队列中移除失败: %w", err)
	}

	// 获取充电桩信息
	pile, err := s.pileRepo.GetByID(pileID)
	if err != nil {
		return fmt.Errorf("获取充电桩失败: %w", err)
	}

	// 更新充电桩状态和队列长度
	err = s.pileRepo.UpdateStatus(pileID, model.PileStatusAvailable)
	if err != nil {
		return fmt.Errorf("更新充电桩状态失败: %w", err)
	}

	err = s.pileRepo.UpdateQueueLength(pileID, pile.QueueLength-1)
	if err != nil {
		return fmt.Errorf("更新充电桩队列长度失败: %w", err)
	}

	// 更新充电桩统计信息
	chargingHours := float64(chargingDuration) / 3600
	err = s.pileRepo.UpdateStats(pileID, 1, chargingHours, actualCapacity)
	if err != nil {
		return fmt.Errorf("更新充电桩统计信息失败: %w", err)
	}

	// 生成详单
	if s.billingService != nil {
		_, err = s.billingService.GenerateBill(session.ID)
		if err != nil {
			log.Printf("生成详单失败: %v", err)
		}
	}

	// 重新排序队列并尝试开始下一个充电
	queueItems, err := s.queueRepo.GetQueueItemsByPile(pileID)
	if err != nil {
		log.Printf("获取队列项失败: %v", err)
	} else {
		// 重新排序队列
		for i, item := range queueItems {
			newPosition := i + 1
			if item.Position != newPosition {
				err = s.queueRepo.UpdateQueuePosition(pileID, item.RequestID, newPosition)
				if err != nil {
					log.Printf("更新队列位置失败: %v", err)
					continue
				}

				err = s.requestRepo.AssignToPile(item.RequestID, pileID, newPosition, 0, model.RequestStatusQueued)
				if err != nil {
					log.Printf("更新请求队列位置失败: %v", err)
					continue
				}
			}

			// 如果新位置是1，开始充电
			if newPosition == 1 {
				s.startCharging(item.RequestID, pileID)
			}
		}
	}

	// 在持有锁的情况下，启动新的goroutine来触发调度
	go s.TryScheduleRequests()

	log.Printf("充电完成处理成功: 充电桩=%s, 用户=%s, 实际充电量=%.1fkWh, 充电时长=%d秒",
		pileID, userID, actualCapacity, chargingDuration)

	return nil
}

// checkAndExecuteBatchScheduling 检查并执行批量调度
func (s *SchedulerService) checkAndExecuteBatchScheduling(config *model.SchedulingConfig) {
	// 获取可用的快充桩和慢充桩
	availableFastPiles, err := s.pileRepo.GetAvailablePiles(model.PileTypeFast, config.ChargingQueueLen)
	if err != nil {
		log.Printf("获取可用快充桩失败: %v", err)
		return
	}

	availableSlowPiles, err := s.pileRepo.GetAvailablePiles(model.PileTypeSlow, config.ChargingQueueLen)
	if err != nil {
		log.Printf("获取可用慢充桩失败: %v", err)
		return
	}

	// 基于可用充电桩计算总车位数
	totalSlots := len(availableFastPiles)*config.ChargingQueueLen +
		len(availableSlowPiles)*config.ChargingQueueLen

	// 获取所有等候区请求
	fastRequests, err := s.requestRepo.GetWaitingRequestsByMode(model.ChargingModeFast)
	if err != nil {
		log.Printf("获取快充请求失败: %v", err)
		return
	}

	slowRequests, err := s.requestRepo.GetWaitingRequestsByMode(model.ChargingModeSlow)
	if err != nil {
		log.Printf("获取慢充请求失败: %v", err)
		return
	}

	allRequests := append(fastRequests, slowRequests...)
	// 只有当等候区车辆数量达到充电区总车位数时才执行批量调度
	if len(allRequests) >= totalSlots {
		err := s.ExecuteBatchScheduling()
		if err != nil {
			log.Printf("批量调度失败: %v", err)
		}
	} else {
		// 车辆数量不足，等待而不执行调度
		log.Printf("批量调度等待中: 当前等候区车辆数量=%d，需要达到可用总车位数=%d (可用快充桩:%d, 可用慢充桩:%d)",
			len(allRequests), totalSlots, len(availableFastPiles), len(availableSlowPiles))
	}
}
