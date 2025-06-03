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

// GenerateBill 生成账单
func (s *BillingService) GenerateBill(sessionID uuid.UUID) (*model.BillingDetail, error) { // 获取充电会话
	session, err := s.sessionRepo.GetByID(sessionID)
	if err != nil {
		return nil, errors.New("充电会话不存在")
	}

	// 检查是否已经生成过账单
	existingBill, _ := s.billingRepo.GetBySessionID(sessionID)
	if existingBill != nil {
		return existingBill, nil
	}
	// 计算充电时长（小时）
	var chargingDuration float64
	if session.EndTime != nil {
		chargingDuration = session.EndTime.Sub(session.StartTime).Hours()
	} else {
		// 如果还没结束，使用当前时间
		chargingDuration = time.Now().UTC().Sub(session.StartTime).Hours()
	}

	// 使用充电会话中的实际充电量（ActualCapacity）
	chargingCapacity := session.ActualCapacity
	// 从系统配置获取电价信息
	priceRate, err := s.billingRepo.GetCurrentPricing(session.StartTime)
	if err != nil {
		return nil, err
	}

	// 根据电价和电量计算充电费用
	chargingFee := chargingCapacity * priceRate.ElectricFee
	// 计算服务费（从配置获取）
	serviceFee := chargingCapacity * priceRate.ServiceFee

	// 计算总费用
	totalFee := chargingFee + serviceFee

	// 四舍五入到小数点后2位
	chargingCapacity = math.Round(chargingCapacity*100) / 100
	chargingDuration = math.Round(chargingDuration*100) / 100
	chargingFee = math.Round(chargingFee*100) / 100
	serviceFee = math.Round(serviceFee*100) / 100
	totalFee = math.Round(totalFee*100) / 100
	// 创建账单
	bill := &model.BillingDetail{
		ID:               uuid.New(),
		SessionID:        sessionID,
		UserID:           session.UserID,
		PileID:           session.PileID,
		ChargingCapacity: chargingCapacity,
		ChargingDuration: chargingDuration,
		StartTime:        session.StartTime,
		EndTime:          time.Now().UTC(), // 使用当前时间作为结束时间
		UnitPrice:        priceRate.ElectricFee,
		PriceType:        priceRate.Period,
		ChargingFee:      chargingFee,
		ServiceFee:       serviceFee,
		TotalFee:         totalFee,
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
func (s *BillingService) GetUserBills(userID uuid.UUID, page, pageSize int) ([]*model.BillingDetail, int, error) {
	if page <= 0 {
		page = 1
	}
	if pageSize <= 0 {
		pageSize = 10
	}
	return s.billingRepo.GetUserBillingDetails(userID, page, pageSize)
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

// GetCurrentPricing 获取当前时间的电价配置
func (s *BillingService) GetCurrentPricing(startTime time.Time) (*model.PriceRate, error) {
	return s.billingRepo.GetCurrentPricing(startTime)
}
