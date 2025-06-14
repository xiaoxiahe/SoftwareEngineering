package service

import (
	"database/sql"

	"backend/internal/config"
	"backend/internal/repository"
)

// Services 包含所有服务
type Services struct {
	User                *UserService
	ChargingPile        *ChargingPileService
	ChargingRequest     *ChargingRequestService
	Scheduler           *SchedulerService
	Billing             *BillingService
	System              *SystemService
	Bootstrap           *BootstrapService
	ChargingSessionRepo *repository.ChargingSessionRepository
}

// NewServices 创建服务集合
func NewServices(db *sql.DB, cfg *config.Config) *Services {
	// 创建仓库
	userRepo := repository.NewUserRepository(db)
	chargingPileRepo := repository.NewChargingPileRepository(db)
	chargingRequestRepo := repository.NewChargingRequestRepository(db)
	chargingSessionRepo := repository.NewChargingSessionRepository(db)
	queueRepo := repository.NewQueueRepository(db)
	billingRepo := repository.NewBillingRepository(db)
	systemRepo := repository.NewSystemRepository(db)
	// 创建服务
	userService := NewUserService(userRepo)
	chargingRequestService := NewChargingRequestService(chargingRequestRepo, queueRepo, chargingPileRepo, systemRepo)
	billingService := NewBillingService(billingRepo, chargingSessionRepo, systemRepo, chargingPileRepo)
	systemService := NewSystemService(systemRepo, chargingRequestRepo, chargingSessionRepo, billingRepo, queueRepo, chargingPileRepo)
	schedulerService := NewSchedulerService(chargingRequestRepo, chargingPileRepo, queueRepo, chargingSessionRepo, systemRepo)
	chargingPileService := NewChargingPileService(chargingPileRepo, systemRepo, userRepo, queueRepo, chargingSessionRepo, billingRepo)
	bootstrapService := NewBootstrapService(systemRepo, chargingPileRepo, cfg)
	// 设置计费服务（避免循环依赖）
	schedulerService.SetBillingService(billingService)

	// 创建模拟器客户端并设置到调度器
	simulatorClient := NewChargingDispatcherClient("http://localhost:8090") // 模拟器的地址
	schedulerService.SetSimulatorClient(simulatorClient)
	// 设置调度服务到充电请求服务（避免循环依赖）
	chargingRequestService.SetSchedulerService(schedulerService)
	return &Services{
		User:                userService,
		ChargingPile:        chargingPileService,
		ChargingRequest:     chargingRequestService,
		Scheduler:           schedulerService,
		Billing:             billingService,
		System:              systemService,
		Bootstrap:           bootstrapService,
		ChargingSessionRepo: chargingSessionRepo,
	}
}
