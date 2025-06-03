package service

import (
	"errors"
	"fmt"
	"sort"
	"time"

	"backend/internal/model"
	"backend/internal/repository"
)

// SystemService 系统服务
type SystemService struct {
	systemRepo       *repository.SystemRepository
	requestRepo      *repository.ChargingRequestRepository
	sessionRepo      *repository.ChargingSessionRepository
	billingRepo      *repository.BillingRepository
	queueRepo        *repository.QueueRepository
	pileRepo         *repository.ChargingPileRepository
	schedulingConfig *model.SchedulingConfig
}

// NewSystemService 创建系统服务
func NewSystemService(
	systemRepo *repository.SystemRepository,
	requestRepo *repository.ChargingRequestRepository,
	sessionRepo *repository.ChargingSessionRepository,
	billingRepo *repository.BillingRepository,
	queueRepo *repository.QueueRepository,
	pileRepo *repository.ChargingPileRepository,
) *SystemService {
	svc := &SystemService{
		systemRepo:  systemRepo,
		requestRepo: requestRepo,
		sessionRepo: sessionRepo,
		billingRepo: billingRepo,
		queueRepo:   queueRepo,
		pileRepo:    pileRepo,
	}

	// 加载调度配置
	config, err := systemRepo.GetSchedulingConfig()
	if err == nil {
		svc.schedulingConfig = config
	}

	return svc
}

// GetSchedulingConfig 获取调度配置
func (s *SystemService) GetSchedulingConfig() (*model.SchedulingConfig, error) {
	if s.schedulingConfig == nil {
		config, err := s.systemRepo.GetSchedulingConfig()
		if err != nil {
			return nil, err
		}
		s.schedulingConfig = config
	}
	return s.schedulingConfig, nil
}

// UpdateSchedulingConfig 更新调度配置
func (s *SystemService) UpdateSchedulingConfig(config *model.SchedulingConfig) error {
	// 验证配置参数
	if config.FastChargingPileNum <= 0 || config.SlowChargingPileNum <= 0 {
		return errors.New("充电桩数量必须大于0")
	}
	if config.WaitingAreaSize <= 0 {
		return errors.New("等候区大小必须大于0")
	}
	if config.ChargingQueueLen <= 0 {
		return errors.New("充电队列长度必须大于0")
	}
	if config.FastChargingPower <= 0 || config.SlowChargingPower <= 0 {
		return errors.New("充电功率必须大于0")
	}
	if config.Strategy != "shortest_completion_time" && config.Strategy != "first_come_first_served" {
		return errors.New("无效的调度策略")
	}

	// 更新数据库
	err := s.systemRepo.UpdateSchedulingConfig(config)
	if err != nil {
		return err
	}

	// 更新内存中的配置
	s.schedulingConfig = config
	return nil
}

// GetSystemConfigs 获取所有系统配置
func (s *SystemService) GetSystemConfigs() ([]*model.SystemConfig, error) {
	return s.systemRepo.GetAllConfigs()
}

// GetSystemConfig 获取系统配置
func (s *SystemService) GetSystemConfig(key string) (*model.SystemConfig, error) {
	config, err := s.systemRepo.GetConfig(key)
	if err != nil {
		return nil, errors.New("配置项不存在")
	}
	return config, nil
}

// GetFaultRecords 获取故障记录
func (s *SystemService) GetFaultRecords(startTime, endTime time.Time, pileID *string, page, pageSize int) ([]*model.FaultRecord, int, error) {
	// 参数验证
	if startTime.After(endTime) {
		return nil, 0, errors.New("开始时间不能晚于结束时间")
	}

	// 默认分页参数
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}

	return s.systemRepo.GetFaultRecords(startTime, endTime, pileID, page, pageSize)
}

// GetPileUsageReport 获取充电桩使用报表
func (s *SystemService) GetPileUsageReport(startDate, endDate time.Time, period string) ([]model.PileUsageStatistics, error) {
	// 参数验证
	if startDate.After(endDate) {
		return nil, errors.New("开始时间不能晚于结束时间")
	}

	// 获取所有充电桩
	piles, err := s.pileRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("获取充电桩信息失败: %v", err)
	}

	var stats []model.PileUsageStatistics

	// 计算每个充电桩的统计数据
	for _, pile := range piles {
		// 获取充电桩的会话数据
		sessions, err := s.sessionRepo.GetSessionsByPileID(pile.ID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("获取充电桩 %s 的会话数据失败: %v", pile.ID, err)
		}
		// 计算充电次数、总时长、总电量
		count := len(sessions)
		var totalDuration, totalCapacity float64
		var totalFee, totalChargingFee, totalServiceFee float64

		// 获取该充电桩对应的所有详单
		for _, session := range sessions {
			if session.Status == model.SessionStatusCompleted || session.Status == model.SessionStatusInterrupted {
				totalDuration += session.Duration
				totalCapacity += session.ActualCapacity

				// 获取账单详情
				billing, err := s.billingRepo.GetBySessionID(session.ID)
				if err == nil && billing != nil {
					totalFee += billing.TotalFee
					totalChargingFee += billing.ChargingFee
					totalServiceFee += billing.ServiceFee
				}
			}
		}
		// 创建并添加统计数据
		stat := model.PileUsageStatistics{
			PileID:           pile.ID,
			Count:            count,
			TotalDuration:    totalDuration,
			TotalCapacity:    totalCapacity,
			TotalChargingFee: totalChargingFee,
			TotalServiceFee:  totalServiceFee,
			TotalFee:         totalFee,
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// GetOperationStats 获取系统运营统计
func (s *SystemService) GetOperationStats(date time.Time, period string) (map[string]any, error) {
	// 验证参数
	if period != "day" && period != "week" && period != "month" {
		return nil, errors.New("无效的统计周期")
	}

	// 计算时间范围
	var startDate, endDate time.Time
	endDate = time.Date(date.Year(), date.Month(), date.Day(), 23, 59, 59, 999999999, date.Location())

	switch period {
	case "day":
		startDate = time.Date(date.Year(), date.Month(), date.Day(), 0, 0, 0, 0, date.Location())
	case "week":
		// 获取本周开始（周一）
		daysToMonday := int(date.Weekday())
		if daysToMonday == 0 {
			daysToMonday = 7
		}
		startDate = time.Date(date.Year(), date.Month(), date.Day()-daysToMonday+1, 0, 0, 0, 0, date.Location())
	case "month":
		startDate = time.Date(date.Year(), date.Month(), 1, 0, 0, 0, 0, date.Location())
	}

	// 获取所有充电桩
	piles, err := s.pileRepo.GetAll()
	if err != nil {
		return nil, fmt.Errorf("获取充电桩信息失败: %v", err)
	}

	// 统计数据
	var totalSessions int
	var totalDuration float64
	var totalCapacity float64
	var totalRevenue float64

	// 按小时统计的会话数（用于计算高峰时段）
	hourlyStats := make(map[int]int)

	// 按桩类型统计使用时间和总时间（用于计算利用率）
	fastPileUsage := struct {
		totalUsage float64
		totalTime  float64
		count      int
	}{}

	slowPileUsage := struct {
		totalUsage float64
		totalTime  float64
		count      int
	}{}

	// 遍历所有充电桩，获取其会话数据
	for _, pile := range piles {
		// 统计桩类型
		if pile.PileType == model.PileTypeFast {
			fastPileUsage.count++
		} else if pile.PileType == model.PileTypeSlow {
			slowPileUsage.count++
		}

		// 获取充电桩的会话数据
		sessions, err := s.sessionRepo.GetSessionsByPileID(pile.ID, startDate, endDate)
		if err != nil {
			return nil, fmt.Errorf("获取充电桩 %s 的会话数据失败: %v", pile.ID, err)
		}

		// 统计数据
		for _, session := range sessions {
			if session.Status == model.SessionStatusCompleted || session.Status == model.SessionStatusInterrupted {
				totalSessions++
				totalDuration += session.Duration
				totalCapacity += session.ActualCapacity

				// 记录会话开始的小时，用于计算高峰时段
				hour := session.StartTime.Hour()
				hourlyStats[hour]++

				// 统计桩类型使用情况
				if pile.PileType == model.PileTypeFast {
					fastPileUsage.totalUsage += session.Duration
				} else if pile.PileType == model.PileTypeSlow {
					slowPileUsage.totalUsage += session.Duration
				}

				// 获取账单详情计算总收入
				billing, err := s.billingRepo.GetBySessionID(session.ID)
				if err == nil && billing != nil {
					totalRevenue += billing.TotalFee
				}
			}
		}
	}

	// 计算时间段总小时数
	totalHours := endDate.Sub(startDate).Hours()

	// 计算每种桩类型的总可用时间
	fastPileUsage.totalTime = float64(fastPileUsage.count) * totalHours
	slowPileUsage.totalTime = float64(slowPileUsage.count) * totalHours

	// 查找前三个高峰时段
	type hourStat struct {
		hour  int
		count int
	}
	var hourStats []hourStat
	for hour, count := range hourlyStats {
		hourStats = append(hourStats, hourStat{hour, count})
	}

	// 按数量降序排序
	sort.Slice(hourStats, func(i, j int) bool {
		return hourStats[i].count > hourStats[j].count
	})

	// 取前3个高峰时段（如果有的话）
	peakHours := []map[string]any{}
	for i := 0; i < len(hourStats) && i < 3; i++ {
		peakHours = append(peakHours, map[string]any{
			"hour":  hourStats[i].hour,
			"count": hourStats[i].count,
		})
	}

	// 返回统计结果
	stats := map[string]any{
		"chargingSessions": totalSessions,
		"totalDuration":    totalDuration,
		"totalCapacity":    totalCapacity,
		"totalRevenue":     totalRevenue,
		"peakHours":        peakHours,
	}

	return stats, nil
}
