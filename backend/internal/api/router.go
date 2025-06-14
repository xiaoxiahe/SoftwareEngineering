package api

import (
	"net/http"

	"backend/internal/api/handlers"
	"backend/internal/config"
	"backend/internal/middleware"
	"backend/internal/service"
)

// SetupRouter 初始化路由
func SetupRouter(services *service.Services, cfg *config.Config) http.Handler {
	// 创建路由
	mux := http.NewServeMux()

	// 中间件
	auth := middleware.NewAuthMiddleware(services.User)
	admin := middleware.NewAdminMiddleware() // 创建处理器
	userHandler := handlers.NewUserHandler(services.User)
	chargingRequestHandler := handlers.NewChargingRequestHandler(services.ChargingRequest, services.ChargingSessionRepo)
	chargingPileHandler := handlers.NewChargingPileHandler(services.ChargingPile, services.Billing)
	queueHandler := handlers.NewQueueHandler(services.ChargingRequest, services.System, services.User)
	billingHandler := handlers.NewBillingHandler(services.Billing)
	systemHandler := handlers.NewSystemHandler(services.System)
	simulatorHandler := handlers.NewSimulatorHandler(services.ChargingPile, services.Scheduler)

	// === 公共接口 ===

	// 健康检查
	mux.HandleFunc("GET /health", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write([]byte("OK"))
	})

	// === 用户认证接口 ===

	// 用户注册
	mux.HandleFunc("POST /api/v1/auth/register", userHandler.Register)

	// 用户登录
	mux.HandleFunc("POST /api/v1/auth/login", userHandler.Login)

	// === 用户接口(需认证) ===

	// 获取用户信息
	mux.HandleFunc("GET /api/v1/users/{userId}", auth(userHandler.GetUserInfo))

	// === 充电请求接口 ===

	// 提交充电请求
	mux.HandleFunc("POST /api/v1/charging/requests", auth(chargingRequestHandler.CreateRequest))

	// 获取充电请求
	mux.HandleFunc("GET /api/v1/charging/requests", auth(chargingRequestHandler.GetRequests))

	// 获取最新充电请求
	mux.HandleFunc("GET /api/v1/charging/requests/latest", auth(chargingRequestHandler.GetLatestRequest))

	// 获取指定充电请求
	mux.HandleFunc("GET /api/v1/charging/requests/{requestId}", auth(chargingRequestHandler.GetRequestByID))

	// 修改充电请求
	mux.HandleFunc("PUT /api/v1/charging/requests/{requestId}", auth(chargingRequestHandler.UpdateRequest))

	// 取消充电请求
	mux.HandleFunc("DELETE /api/v1/charging/requests/{requestId}", auth(chargingRequestHandler.CancelRequest))

	// === 排队系统接口 ===

	// 查询排队状态
	mux.HandleFunc("GET /api/v1/queue/status", auth(queueHandler.GetQueueStatus))

	// 查询用户排队位置
	mux.HandleFunc("GET /api/v1/queue/position/{userId}", auth(queueHandler.GetUserQueuePosition))

	// 查询等候区车辆信息
	mux.HandleFunc("GET /api/v1/queue/waiting-vehicles", queueHandler.GetWaitingVehicles)

	// === 充电桩接口 ===

	// 查询所有充电桩状态
	mux.HandleFunc("GET /api/v1/charging-piles", chargingPileHandler.GetAllPiles)

	// === 计费接口 ===

	// 查询充电详单列表
	mux.HandleFunc("GET /api/v1/billing/details", auth(billingHandler.GetBillingDetails))

	// 查询单个详单
	mux.HandleFunc("GET /api/v1/billing/details/{detailId}", auth(billingHandler.GetBillingDetailByID))

	// 计算预估充电费用
	mux.HandleFunc("POST /api/v1/billing/calculate", auth(billingHandler.CalculateChargingFee))

	// === 管理员接口 ===

	// 控制充电桩
	mux.HandleFunc("POST /api/v1/admin/charging-piles/{pileId}/control", auth(admin(chargingPileHandler.ControlPile)))

	// 获取充电桩等候车辆信息
	mux.HandleFunc("GET /api/v1/admin/charging-piles/queue-vehicles", auth(admin(chargingPileHandler.GetQueueVehicles)))

	// 充电桩使用报表
	mux.HandleFunc("GET /api/v1/admin/reports/charging-piles", auth(admin(systemHandler.GetPileUsageReport)))

	// 系统运营统计
	mux.HandleFunc("GET /api/v1/admin/reports/operations", auth(admin(systemHandler.GetOperationStats)))

	// === 模拟器接口 ===

	// 充电进度更新
	mux.HandleFunc("POST /api/v1/simulator/charging-progress", simulatorHandler.UpdateChargingProgress)

	// 充电完成上报
	mux.HandleFunc("POST /api/v1/simulator/charging-complete", simulatorHandler.CompleteCharging)

	// 故障报告
	mux.HandleFunc("POST /api/v1/simulator/fault-report", simulatorHandler.ReportFault)

	// 故障恢复
	mux.HandleFunc("POST /api/v1/simulator/fault-recovery", simulatorHandler.RecoverFault)

	// 模拟器心跳检测
	mux.HandleFunc("POST /api/v1/simulator/heartbeat", simulatorHandler.Heartbeat)

	// 应用CORS中间件
	corsHandler := middleware.CORSMiddleware(mux)
	return corsHandler
}
