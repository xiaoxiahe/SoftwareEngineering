package service

import (
	"database/sql"
	"fmt"
	"time"

	"backend/internal/config"
	"backend/internal/model"
	"backend/internal/repository"
)

// BootstrapService 启动引导服务，用于初始化系统配置和数据
type BootstrapService struct {
	systemRepo       *repository.SystemRepository
	chargingPileRepo *repository.ChargingPileRepository
	config           *config.Config
}

// NewBootstrapService 创建引导服务
func NewBootstrapService(
	systemRepo *repository.SystemRepository,
	chargingPileRepo *repository.ChargingPileRepository,
	config *config.Config,
) *BootstrapService {
	return &BootstrapService{
		systemRepo:       systemRepo,
		chargingPileRepo: chargingPileRepo,
		config:           config,
	}
}

// InitializeSystem 初始化系统配置
func (s *BootstrapService) InitializeSystem() error {
	// 同步系统配置
	if err := s.syncSystemConfig(); err != nil {
		return fmt.Errorf("同步系统配置失败: %w", err)
	}

	// 同步电价配置
	if err := s.syncPricingConfig(); err != nil {
		return fmt.Errorf("同步电价配置失败: %w", err)
	}

	// 初始化充电桩
	if err := s.ensureChargingPilesExist(); err != nil {
		return fmt.Errorf("确保充电桩存在失败: %w", err)
	}

	return nil
}

// syncSystemConfig 将配置文件中的系统配置同步到数据库
func (s *BootstrapService) syncSystemConfig() error {
	// 从配置对象获取系统配置项
	configItems := map[string]string{
		"fast_charging_pile_num":    fmt.Sprintf("%d", s.config.Charging.FastChargingPileNum),
		"trickle_charging_pile_num": fmt.Sprintf("%d", s.config.Charging.TrickleChargingPileNum),
		"waiting_area_size":         fmt.Sprintf("%d", s.config.Charging.WaitingAreaSize),
		"charging_queue_len":        fmt.Sprintf("%d", s.config.Charging.ChargingQueueLen),
		"fast_charging_power":       fmt.Sprintf("%.2f", s.config.Charging.FastChargingPower),
		"trickle_charging_power":    fmt.Sprintf("%.2f", s.config.Charging.TrickleChargingPower),
		"service_fee_per_unit":      fmt.Sprintf("%.2f", s.config.Charging.ServiceFeePerUnit),
	}

	// 对每个配置项，检查是否存在，不存在则创建，存在则更新
	for key, value := range configItems {
		config, err := s.systemRepo.GetConfig(key)
		if err != nil {
			if err == sql.ErrNoRows {
				// 配置项不存在，创建一个新的
				configType := "string"
				if key == "fault_rescheduling_policy" {
					configType = "string"
				} else {
					configType = "number"
				}

				description := getConfigDescription(key)

				if err := s.systemRepo.CreateConfig(key, value, configType, description); err != nil {
					return err
				}
			} else {
				return err
			}
		} else {
			// 配置项存在，检查是否需要更新
			if config.ConfigValue != value {
				if err := s.systemRepo.UpdateConfig(key, value); err != nil {
					return err
				}
			}
		}
	}

	return nil
}

// localTimeToUTC 将本地时间的小时分钟转换为UTC时间字符串
func (s *BootstrapService) localTimeToUTC(hour, minute int) string {
	// 使用今天的日期创建本地时间
	now := time.Now()
	localTime := time.Date(now.Year(), now.Month(), now.Day(), hour, minute, 0, 0, now.Location())

	// 转换为UTC时间
	utcTime := localTime.UTC()

	// 格式化为时间字符串
	return fmt.Sprintf("%02d:%02d:00", utcTime.Hour(), utcTime.Minute())
}

// syncPricingConfig 将配置文件中的电价配置同步到数据库
func (s *BootstrapService) syncPricingConfig() error {
	// 检查今天是否已经有生效的电价配置
	today := time.Now().UTC().Format("2006-01-02")

	// 使用仓库方法获取今天的电价配置
	// 这里应该增加一个新的方法来检查是否存在今天日期的配置项
	hasTodayPricing, err := s.systemRepo.HasPricingForDate(today)
	if err != nil {
		return err
	}

	// 如果今天已经有电价配置，不需要再创建
	if hasTodayPricing {
		return nil
	}

	// 从配置对象获取电价配置
	// 处理高峰时段电价
	for i := range len(s.config.Pricing.PeakStartTime) {
		startHour, startMin := s.config.Pricing.PeakStartTime[i][0], s.config.Pricing.PeakStartTime[i][1]
		endHour, endMin := s.config.Pricing.PeakEndTime[i][0], s.config.Pricing.PeakEndTime[i][1]

		// 将本地时间转换为UTC时间
		startTime := s.localTimeToUTC(startHour, startMin)
		endTime := s.localTimeToUTC(endHour, endMin)

		if err := s.systemRepo.CreatePricingConfig("peak", s.config.Pricing.PeakPrice, startTime, endTime, s.config.Pricing.ServiceFee); err != nil {
			return err
		}
	}
	// 处理平时电价
	for i := range len(s.config.Pricing.FlatStartTime) {
		startHour, startMin := s.config.Pricing.FlatStartTime[i][0], s.config.Pricing.FlatStartTime[i][1]
		endHour, endMin := s.config.Pricing.FlatEndTime[i][0], s.config.Pricing.FlatEndTime[i][1]

		// 将本地时间转换为UTC时间
		startTime := s.localTimeToUTC(startHour, startMin)
		endTime := s.localTimeToUTC(endHour, endMin)

		if err := s.systemRepo.CreatePricingConfig("normal", s.config.Pricing.NormalPrice, startTime, endTime, s.config.Pricing.ServiceFee); err != nil {
			return err
		}
	}

	// 处理谷时电价
	for i := range len(s.config.Pricing.ValleyStart) {
		startHour, startMin := s.config.Pricing.ValleyStart[i][0], s.config.Pricing.ValleyStart[i][1]
		endHour, endMin := s.config.Pricing.ValleyEnd[i][0], s.config.Pricing.ValleyEnd[i][1]

		// 将本地时间转换为UTC时间
		startTime := s.localTimeToUTC(startHour, startMin)
		endTime := s.localTimeToUTC(endHour, endMin)

		if err := s.systemRepo.CreatePricingConfig("valley", s.config.Pricing.ValleyPrice, startTime, endTime, s.config.Pricing.ServiceFee); err != nil {
			return err
		}
	}

	return nil
}

// ensureChargingPilesExist 确保充电桩存在，如果不存在则创建
func (s *BootstrapService) ensureChargingPilesExist() error {
	// 定义充电桩ID前缀
	fastPilePrefix := "F" // 快充桩前缀
	slowPilePrefix := "T" // 慢充桩前缀

	// 获取系统中配置的充电桩数量
	fastPileCount := s.config.Charging.FastChargingPileNum
	slowPileCount := s.config.Charging.TrickleChargingPileNum

	// 获取现有充电桩
	existingPiles, err := s.chargingPileRepo.GetAll()
	if err != nil {
		return err
	}

	// 记录现有桩ID
	existingPileIDs := make(map[string]bool)
	for _, pile := range existingPiles {
		existingPileIDs[pile.ID] = true
	}

	// 创建快充桩（如果不存在）
	for i := 1; i <= fastPileCount; i++ {
		pileID := fmt.Sprintf("%s%d", fastPilePrefix, i)
		if !existingPileIDs[pileID] {
			pile := &model.ChargingPile{
				ID:            pileID,
				PileType:      model.PileTypeFast,
				Power:         s.config.Charging.FastChargingPower,
				Status:        model.PileStatusAvailable,
				QueueLength:   0,
				TotalSessions: 0,
				TotalDuration: 0,
				TotalEnergy:   0,
				CreatedAt:     time.Now().UTC(),
				UpdatedAt:     time.Now().UTC(),
			}
			if err := s.chargingPileRepo.Create(pile); err != nil {
				return err
			}
		}
	}

	// 创建慢充桩（如果不存在）
	for i := 1; i <= slowPileCount; i++ {
		pileID := fmt.Sprintf("%s%d", slowPilePrefix, i)
		if !existingPileIDs[pileID] {
			pile := &model.ChargingPile{
				ID:            pileID,
				PileType:      model.PileTypeSlow,
				Power:         s.config.Charging.TrickleChargingPower,
				Status:        model.PileStatusAvailable,
				QueueLength:   0,
				TotalSessions: 0,
				TotalDuration: 0,
				TotalEnergy:   0,
				CreatedAt:     time.Now().UTC(),
				UpdatedAt:     time.Now().UTC(),
			}
			if err := s.chargingPileRepo.Create(pile); err != nil {
				return err
			}
		}
	}

	return nil
}

// FindOrCreatePileByID 查找或创建充电桩
func (s *BootstrapService) FindOrCreatePileByID(pileID string, status model.PileStatus) (*model.ChargingPile, error) {
	// 尝试查找现有充电桩
	pile, err := s.chargingPileRepo.GetByID(pileID)
	if err == nil {
		// 充电桩已存在，更新状态
		if pile.Status != status {
			if err := s.chargingPileRepo.UpdateStatus(pileID, status); err != nil {
				return nil, err
			}
			pile.Status = status
		}
		return pile, nil
	}

	// 充电桩不存在，根据ID前缀判断类型
	var pileType model.PileType
	var power float64

	if pileID[0] == 'F' || pileID[0] == 'A' || pileID[0] == 'B' {
		pileType = model.PileTypeFast
		power = s.config.Charging.FastChargingPower
	} else {
		pileType = model.PileTypeSlow
		power = s.config.Charging.TrickleChargingPower
	}

	// 创建新充电桩
	newPile := &model.ChargingPile{
		ID:            pileID,
		PileType:      pileType,
		Power:         power,
		Status:        status,
		QueueLength:   0,
		TotalSessions: 0,
		TotalDuration: 0,
		TotalEnergy:   0,
		CreatedAt:     time.Now().UTC(),
		UpdatedAt:     time.Now().UTC(),
	}

	// 保存到数据库
	if err := s.chargingPileRepo.Create(newPile); err != nil {
		return nil, err
	}

	return newPile, nil
}

// 配置项描述帮助函数
func getConfigDescription(key string) string {
	descriptions := map[string]string{
		"fast_charging_pile_num":    "快充电桩数量",
		"trickle_charging_pile_num": "慢充电桩数量",
		"waiting_area_size":         "等候区容量",
		"charging_queue_len":        "充电桩队列长度",
		"fast_charging_power":       "快充功率(度/小时)",
		"trickle_charging_power":    "慢充功率(度/小时)",
		"service_fee_per_unit":      "服务费率(元/度)",
	}

	if desc, ok := descriptions[key]; ok {
		return desc
	}
	return key
}
