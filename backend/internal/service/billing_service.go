package service

import (
	"errors"
	"math"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/google/uuid"
)

// BillingService 计费服务
type BillingService struct {
	billingRepo *repository.BillingRepository
	sessionRepo *repository.ChargingSessionRepository
	systemRepo  *repository.SystemRepository
	pileRepo    *repository.ChargingPileRepository
}

// SessionFeeCalculation 会话费用计算结果
type SessionFeeCalculation struct {
	TotalChargingFee  float64                     `json:"totalChargingFee"`  // 总充电费用
	TotalServiceFee   float64                     `json:"totalServiceFee"`   // 总服务费用
	TotalFee          float64                     `json:"totalFee"`          // 总费用
	PeakHours         float64                     `json:"peakHours"`         // 峰时小时数
	NormalHours       float64                     `json:"normalHours"`       // 平时小时数
	ValleyHours       float64                     `json:"valleyHours"`       // 谷时小时数
	PeakElectricity   float64                     `json:"peakElectricity"`   // 峰时电量
	NormalElectricity float64                     `json:"normalElectricity"` // 平时电量
	ValleyElectricity float64                     `json:"valleyElectricity"` // 谷时电量
	MainPriceType     string                      `json:"mainPriceType"`     // 主要电价类型
	MainUnitPrice     float64                     `json:"mainUnitPrice"`     // 主要单价
	SegmentBillings   []*model.TimeSegmentBilling `json:"segmentBillings"`   // 分时段计费详情
}

// NewBillingService 创建计费服务
func NewBillingService(
	billingRepo *repository.BillingRepository,
	sessionRepo *repository.ChargingSessionRepository,
	systemRepo *repository.SystemRepository,
	pileRepo *repository.ChargingPileRepository,
) *BillingService {
	return &BillingService{
		billingRepo: billingRepo,
		sessionRepo: sessionRepo,
		systemRepo:  systemRepo,
		pileRepo:    pileRepo,
	}
}

// GenerateBill 生成账单（支持跨时段计费）
func (s *BillingService) GenerateBill(sessionID uuid.UUID) (*model.BillingDetail, error) {
	// 获取充电会话
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, errors.New("充电会话不存在")
	}

	// 检查是否已经生成过账单
	existingBill, _ := s.billingRepo.GetBySessionID(sessionID)
	if existingBill != nil {
		return existingBill, nil
	}

	// 使用可复用的费用计算函数
	feeCalculation, err := s.CalculateSessionFeeBySession(session)
	if err != nil {
		return nil, err
	}

	// 确定充电结束时间
	var endTime time.Time
	if session.EndTime != nil {
		endTime = *session.EndTime
	} else {
		endTime = time.Now().UTC()
	}

	// 计算总充电时长（小时）
	chargingDuration := endTime.Sub(session.StartTime).Hours()
	chargingCapacity := session.ActualCapacity

	// 四舍五入到小数点后2位
	chargingCapacity = math.Round(chargingCapacity*100) / 100
	chargingDuration = math.Round(chargingDuration*100) / 100

	// 创建账单
	bill := &model.BillingDetail{
		ID:                uuid.New(),
		SessionID:         sessionID,
		UserID:            session.UserID,
		PileID:            session.PileID,
		ChargingCapacity:  chargingCapacity,
		ChargingDuration:  chargingDuration,
		StartTime:         session.StartTime,
		EndTime:           endTime,
		UnitPrice:         feeCalculation.MainUnitPrice,
		PriceType:         feeCalculation.MainPriceType,
		ChargingFee:       feeCalculation.TotalChargingFee,
		ServiceFee:        feeCalculation.TotalServiceFee,
		TotalFee:          feeCalculation.TotalFee,
		PeakHours:         feeCalculation.PeakHours,
		NormalHours:       feeCalculation.NormalHours,
		ValleyHours:       feeCalculation.ValleyHours,
		PeakElectricity:   feeCalculation.PeakElectricity,
		NormalElectricity: feeCalculation.NormalElectricity,
		ValleyElectricity: feeCalculation.ValleyElectricity,
	}

	// 保存到数据库
	return s.billingRepo.CreateBillingDetail(bill)
}

// GetBillByID 通过ID获取账单
func (s *BillingService) GetBillByID(billID uuid.UUID) (*model.BillingDetail, error) {
	bill, err := s.billingRepo.GetByID(billID)
	if err != nil {
		return nil, errors.New("账单不存在")
	}
	return bill, nil
}

// GetUserBills 获取用户账单
func (s *BillingService) GetUserBills(userID uuid.UUID, startDate, endDate *time.Time, page, pageSize int) ([]*model.BillingDetail, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.billingRepo.GetUserBillingDetails(userID, startDate, endDate, page, pageSize)
}

// GetBillingStatistics 获取计费统计
func (s *BillingService) GetBillingStatistics(startTime, endTime time.Time, pileID *string) (*model.BillingStatistics, error) {
	// 参数验证
	if startTime.After(endTime) {
		return nil, errors.New("开始时间不能晚于结束时间")
	}

	// 从仓储层获取原始数据
	stats, err := s.billingRepo.GetBillingStatistics(startTime, endTime, pileID)
	if err != nil {
		return nil, err
	}

	return stats, nil
}

// CalculateSessionFee 根据sessionID计算充电费用（可复用的核心计算逻辑）
func (s *BillingService) CalculateSessionFee(sessionID uuid.UUID) (*SessionFeeCalculation, error) {
	// 获取充电会话
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, errors.New("充电会话不存在")
	}

	return s.CalculateSessionFeeBySession(session)
}

// CalculateSessionFeeBySession 根据会话对象计算充电费用
func (s *BillingService) CalculateSessionFeeBySession(session *model.ChargingSession) (*SessionFeeCalculation, error) {
	// 确定充电结束时间
	var endTime time.Time
	if session.EndTime != nil {
		endTime = *session.EndTime
	} else {
		endTime = time.Now().UTC()
	}

	chargingCapacity := session.ActualCapacity

	// 获取所有电价配置
	priceRates, err := s.billingRepo.GetAllPricingConfigWithRates()
	if err != nil {
		return nil, err
	}

	// 计算跨时段费用
	segmentBillings := s.calculateCrossTimeSegmentBilling(session.StartTime, endTime, chargingCapacity, priceRates)

	// 汇总各时段费用
	var totalChargingFee, totalServiceFee float64
	var peakHours, normalHours, valleyHours float64
	var peakElectricity, normalElectricity, valleyElectricity float64

	for _, segment := range segmentBillings {
		totalChargingFee += segment.ElectricCost
		totalServiceFee += segment.ServiceCost

		switch segment.Period {
		case "peak":
			peakHours += segment.Duration
			peakElectricity += segment.Capacity
		case "normal":
			normalHours += segment.Duration
			normalElectricity += segment.Capacity
		case "valley":
			valleyHours += segment.Duration
			valleyElectricity += segment.Capacity
		}
	}

	totalFee := totalChargingFee + totalServiceFee

	// 四舍五入到小数点后2位
	totalChargingFee = math.Round(totalChargingFee*100) / 100
	totalServiceFee = math.Round(totalServiceFee*100) / 100
	totalFee = math.Round(totalFee*100) / 100
	peakHours = math.Round(peakHours*100) / 100
	normalHours = math.Round(normalHours*100) / 100
	valleyHours = math.Round(valleyHours*100) / 100
	peakElectricity = math.Round(peakElectricity*100) / 100
	normalElectricity = math.Round(normalElectricity*100) / 100
	valleyElectricity = math.Round(valleyElectricity*100) / 100

	// 获取主要电价类型（占用时间最长的时段）
	var mainPriceType string
	var mainUnitPrice float64
	maxHours := math.Max(peakHours, math.Max(normalHours, valleyHours))

	if maxHours == peakHours {
		mainPriceType = "peak"
		for _, rate := range priceRates {
			if rate.Period == "peak" {
				mainUnitPrice = rate.ElectricFee
				break
			}
		}
	} else if maxHours == normalHours {
		mainPriceType = "normal"
		for _, rate := range priceRates {
			if rate.Period == "normal" {
				mainUnitPrice = rate.ElectricFee
				break
			}
		}
	} else {
		mainPriceType = "valley"
		for _, rate := range priceRates {
			if rate.Period == "valley" {
				mainUnitPrice = rate.ElectricFee
				break
			}
		}
	}

	return &SessionFeeCalculation{
		TotalChargingFee:  totalChargingFee,
		TotalServiceFee:   totalServiceFee,
		TotalFee:          totalFee,
		PeakHours:         peakHours,
		NormalHours:       normalHours,
		ValleyHours:       valleyHours,
		PeakElectricity:   peakElectricity,
		NormalElectricity: normalElectricity,
		ValleyElectricity: valleyElectricity,
		MainPriceType:     mainPriceType,
		MainUnitPrice:     mainUnitPrice,
		SegmentBillings:   segmentBillings,
	}, nil
}

// CalculateMultipleSessionsFee 计算多个会话的累积费用
func (s *BillingService) CalculateMultipleSessionsFee(sessions []*model.ChargingSession) (float64, float64, error) {
	var totalChargingFee, totalServiceFee float64

	for _, session := range sessions {
		feeCalculation, err := s.CalculateSessionFeeBySession(session)
		if err != nil {
			// 如果单个会话计算失败，可以选择跳过或返回错误
			// 这里选择跳过，继续计算其他会话
			continue
		}
		totalChargingFee += feeCalculation.TotalChargingFee
		totalServiceFee += feeCalculation.TotalServiceFee
	}

	// 四舍五入到小数点后2位
	totalChargingFee = math.Round(totalChargingFee*100) / 100
	totalServiceFee = math.Round(totalServiceFee*100) / 100

	return totalChargingFee, totalServiceFee, nil
}

// CalculateSessionsFeeByRequestID 根据requestID计算所有相关会话的累积费用
func (s *BillingService) CalculateSessionsFeeByRequestID(requestID uuid.UUID) (float64, float64, error) {
	// 获取该requestID的所有会话
	sessions, err := s.sessionRepo.GetAllByRequestID(requestID)
	if err != nil {
		return 0, 0, err
	}

	return s.CalculateMultipleSessionsFee(sessions)
}

// CalculateSessionsFeeByIDs 根据会话ID列表计算累积费用
func (s *BillingService) CalculateSessionsFeeByIDs(sessionIDs []uuid.UUID) (float64, float64, error) {
	var sessions []*model.ChargingSession

	for _, sessionID := range sessionIDs {
		session, err := s.sessionRepo.GetByID(sessionID)
		if err != nil {
			// 跳过获取失败的会话
			continue
		}
		sessions = append(sessions, session)
	}

	return s.CalculateMultipleSessionsFee(sessions)
}

// GetCurrentPricing 获取当前时间的电价配置
func (s *BillingService) GetCurrentPricing(startTime time.Time) (*model.PriceRate, error) {
	return s.billingRepo.GetCurrentPricing(startTime)
}

// calculateCrossTimeSegmentBilling 计算跨时段计费
func (s *BillingService) calculateCrossTimeSegmentBilling(
	startTime, endTime time.Time,
	totalCapacity float64,
	priceRates []*model.PriceRate,
) []*model.TimeSegmentBilling {

	// 按时间顺序对电价配置进行排序
	sortedRates := make([]*model.PriceRate, len(priceRates))
	copy(sortedRates, priceRates)

	// 简单排序，按开始小时排序
	for i := 0; i < len(sortedRates)-1; i++ {
		for j := i + 1; j < len(sortedRates); j++ {
			if sortedRates[i].StartHour > sortedRates[j].StartHour {
				sortedRates[i], sortedRates[j] = sortedRates[j], sortedRates[i]
			}
		}
	}

	var segments []*model.TimeSegmentBilling
	totalDuration := endTime.Sub(startTime).Hours()

	if totalDuration <= 0 {
		return segments
	}

	currentTime := startTime
	remainingCapacity := totalCapacity

	for currentTime.Before(endTime) {
		// 确定当前时间所属的电价时段
		currentRate := s.getPriceRateForTime(currentTime, sortedRates)
		if currentRate == nil {
			// 如果找不到匹配的电价，跳到下一小时
			currentTime = currentTime.Add(time.Hour)
			continue
		}

		// 计算在当前电价时段内的结束时间
		segmentEndTime := s.getSegmentEndTime(currentTime, endTime, currentRate)

		// 计算该时段的持续时间
		segmentDuration := segmentEndTime.Sub(currentTime).Hours()

		// 按比例计算该时段的充电量（假设充电速度均匀）
		segmentCapacity := (segmentDuration / totalDuration) * totalCapacity

		// 计算该时段的费用
		electricCost := segmentCapacity * currentRate.ElectricFee
		serviceCost := segmentCapacity * currentRate.ServiceFee
		totalCost := electricCost + serviceCost

		// 创建时段计费记录
		segment := &model.TimeSegmentBilling{
			Period:       currentRate.Period,
			StartTime:    currentTime,
			EndTime:      segmentEndTime,
			Duration:     segmentDuration,
			Capacity:     segmentCapacity,
			ElectricFee:  currentRate.ElectricFee,
			ServiceFee:   currentRate.ServiceFee,
			ElectricCost: electricCost,
			ServiceCost:  serviceCost,
			TotalCost:    totalCost,
		}

		segments = append(segments, segment)

		// 移动到下一个时段
		currentTime = segmentEndTime
		remainingCapacity -= segmentCapacity
	}

	return segments
}

// getPriceRateForTime 获取指定时间的电价配置
func (s *BillingService) getPriceRateForTime(t time.Time, priceRates []*model.PriceRate) *model.PriceRate {
	hour := t.Hour()

	for _, rate := range priceRates {
		// 处理跨天的情况，例如 22:00-06:00
		if rate.StartHour <= rate.EndHour {
			// 同一天内的时段，例如 06:00-22:00
			if hour >= rate.StartHour && hour < rate.EndHour {
				return rate
			}
		} else {
			// 跨天的时段，例如 22:00-06:00
			if hour >= rate.StartHour || hour < rate.EndHour {
				return rate
			}
		}
	}

	// 如果没有找到匹配的，返回默认的normal时段
	for _, rate := range priceRates {
		if rate.Period == "normal" {
			return rate
		}
	}

	return nil
}

// getSegmentEndTime 获取当前时段的结束时间
func (s *BillingService) getSegmentEndTime(currentTime, endTime time.Time, currentRate *model.PriceRate) time.Time {
	// 计算当前电价时段的结束时间
	var segmentEndHour int
	if currentRate.StartHour <= currentRate.EndHour {
		// 同一天内的时段
		segmentEndHour = currentRate.EndHour
	} else {
		// 跨天的时段
		if currentTime.Hour() >= currentRate.StartHour {
			// 当前在跨天时段的前半部分，结束时间是次日的EndHour
			segmentEndHour = currentRate.EndHour
		} else {
			// 当前在跨天时段的后半部分，结束时间是今日的EndHour
			segmentEndHour = currentRate.EndHour
		}
	}

	// 构造时段结束时间
	year, month, day := currentTime.Date()
	var segmentEnd time.Time

	if currentRate.StartHour > currentRate.EndHour && currentTime.Hour() >= currentRate.StartHour {
		// 跨天时段的前半部分，结束时间在次日
		segmentEnd = time.Date(year, month, day+1, segmentEndHour, 0, 0, 0, currentTime.Location())
	} else {
		// 同一天的时段或跨天时段的后半部分
		segmentEnd = time.Date(year, month, day, segmentEndHour, 0, 0, 0, currentTime.Location())
	}

	// 如果计算出的时段结束时间超过了充电结束时间，使用充电结束时间
	if segmentEnd.After(endTime) {
		return endTime
	}

	return segmentEnd
}
