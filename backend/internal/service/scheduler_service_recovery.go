package service

import (
	"backend/internal/model"
	"fmt"
	"log"
	"time"
)

// HandlePileRecovery 处理充电桩恢复
func (s *SchedulerService) HandlePileRecovery(pileID string) error {
	s.mutex.Lock()
	defer s.mutex.Unlock()

	log.Printf("处理充电桩恢复: %s", pileID)

	// 更新充电桩状态为可用
	err := s.pileRepo.UpdateStatus(pileID, model.PileStatusAvailable)
	if err != nil {
		return fmt.Errorf("更新充电桩状态失败: %w", err)
	}

	// 重置充电桩队列长度
	err = s.pileRepo.UpdateQueueLength(pileID, 0)
	if err != nil {
		log.Printf("重置充电桩队列长度失败: %v", err)
	}

	// 获取恢复充电桩的类型
	recoveredPile, err := s.pileRepo.GetByID(pileID)
	if err != nil {
		return fmt.Errorf("获取充电桩信息失败: %w", err)
	}
	// 检查其他同类型充电桩中是否有车辆排队
	hasQueuedVehicles, err := s.hasQueuedVehiclesInSameTypePiles(recoveredPile.PileType)
	if err != nil {
		log.Printf("检查同类型充电桩队列失败: %v", err)
		hasQueuedVehicles = false
	}

	if hasQueuedVehicles {
		// 暂停等候区叫号服务
		s.waitingAreaLock = true
		log.Printf("充电桩 %s 恢复，发现其他同类型充电桩有排队车辆，暂停等候区叫号服务", pileID)

		// 执行故障恢复重调度
		// 在释放锁后触发调度，避免死锁
		defer func() {
			go s.executeRecoveryRescheduling(recoveredPile.PileType)
		}()
	} else {
		// 没有排队车辆，直接恢复等候区叫号服务
		log.Printf("充电桩 %s 恢复，其他同类型充电桩无排队车辆，直接恢复正常调度", pileID)
		defer func() {
			go s.resumeWaitingAreaService()
		}()
	}

	return nil
}

// hasQueuedVehiclesInSameTypePiles 检查同类型充电桩中是否有车辆排队
func (s *SchedulerService) hasQueuedVehiclesInSameTypePiles(pileType model.PileType) (bool, error) {
	// 获取所有同类型的充电桩
	piles, err := s.pileRepo.GetByType(pileType)
	if err != nil {
		return false, err
	}

	// 检查每个充电桩是否有排队车辆
	for _, pile := range piles {
		if pile.QueueLength > 0 {
			return true, nil
		}

		// 另外检查数据库中是否有分配给该充电桩的等待请求
		requests, err := s.requestRepo.GetRequestsByPile(pile.ID)
		if err != nil {
			log.Printf("检查充电桩 %s 队列失败: %v", pile.ID, err)
			continue
		}
		// 统计排队中的请求（不包括正在充电的）
		waitingCount := 0
		for _, req := range requests {
			if req.Status == model.RequestStatusQueued {
				waitingCount++
			}
		}

		if waitingCount > 0 {
			return true, nil
		}
	}

	return false, nil
}

// executeRecoveryRescheduling 执行故障恢复重调度
func (s *SchedulerService) executeRecoveryRescheduling(pileType model.PileType) {
	log.Printf("开始执行故障恢复重调度，充电桩类型: %s", pileType)

	// 获取所有同类型充电桩中排队的车辆
	allQueuedRequests, err := s.collectQueuedRequestsFromSameTypePiles(pileType)
	if err != nil {
		log.Printf("收集同类型充电桩排队请求失败: %v", err)
		s.resumeWaitingAreaService()
		return
	}

	if len(allQueuedRequests) == 0 {
		log.Printf("没有找到需要重新调度的排队车辆")
		s.resumeWaitingAreaService()
		return
	}

	// 按照排队号码排序
	s.sortRequests(allQueuedRequests)

	log.Printf("找到 %d 个需要重新调度的排队车辆", len(allQueuedRequests))

	// 将所有车辆从原充电桩队列中移除
	for _, req := range allQueuedRequests {
		err := s.queueRepo.RemoveFromQueueAndDecrementPile(req.ID, req.PileID)
		if err != nil {
			log.Printf("从队列移除请求 %s 失败: %v", req.ID, err)
		}

		// 清除原充电桩分配
		err = s.requestRepo.AssignToPile(req.ID, "", 0, 0, model.RequestStatusWaiting)
		if err != nil {
			log.Printf("清除请求 %s 充电桩分配失败: %v", req.ID, err)
		}
	}

	// 重新调度所有车辆
	// 获取系统配置
	config, err := s.systemRepo.GetSchedulingConfig()
	if err != nil {
		log.Printf("获取系统配置失败: %v", err)
		s.resumeWaitingAreaService()
		return
	}

	// 获取可用的同类型充电桩
	availablePiles, err := s.pileRepo.GetAvailablePiles(pileType, config.ChargingQueueLen)
	if err != nil {
		log.Printf("获取可用充电桩失败: %v", err)
		s.resumeWaitingAreaService()
		return
	}

	// 重新调度每个车辆到最佳充电桩
	for _, req := range allQueuedRequests {
		bestPile := s.findBestPile(availablePiles, req.RequestedCapacity, config.ChargingQueueLen)
		if bestPile != nil {
			s.scheduleRequestToPile(req.ID, bestPile.ID, bestPile.QueueLength+1)
			// 更新本地充电桩队列长度以便下次计算
			bestPile.QueueLength++
		} else {
			log.Printf("无法为请求 %s 找到合适的充电桩", req.ID)
		}
		// 稍微延迟，避免过度并发
		time.Sleep(10 * time.Millisecond)
	}

	log.Printf("故障恢复重调度完成")

	// 恢复等候区叫号服务
	s.resumeWaitingAreaService()
}

// collectQueuedRequestsFromSameTypePiles 收集同类型充电桩中的排队请求
func (s *SchedulerService) collectQueuedRequestsFromSameTypePiles(pileType model.PileType) ([]*model.ChargingRequest, error) {
	// 获取所有同类型的充电桩
	piles, err := s.pileRepo.GetByType(pileType)
	if err != nil {
		return nil, err
	}

	var allQueuedRequests []*model.ChargingRequest

	// 收集每个充电桩的排队请求
	for _, pile := range piles {
		requests, err := s.requestRepo.GetRequestsByPile(pile.ID)
		if err != nil {
			log.Printf("获取充电桩 %s 队列失败: %v", pile.ID, err)
			continue
		}
		// 只收集排队中的请求（不包括正在充电的）
		for _, req := range requests {
			if req.Status == model.RequestStatusQueued {
				allQueuedRequests = append(allQueuedRequests, req)
			}
		}
	}

	return allQueuedRequests, nil
}

// resumeWaitingAreaService 恢复等候区叫号服务
func (s *SchedulerService) resumeWaitingAreaService() {
	s.waitingAreaLock = false

	log.Printf("等候区叫号服务已恢复")

	// 触发一次调度，处理等候区中的车辆
	go s.TryScheduleRequests()
}
