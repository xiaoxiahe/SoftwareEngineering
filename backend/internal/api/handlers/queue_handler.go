package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"

	"github.com/google/uuid"
)

// QueueHandler 队列处理器
type QueueHandler struct {
	chargingRequestService *service.ChargingRequestService
	systemService          *service.SystemService
	userService            *service.UserService
}

// NewQueueHandler 创建队列处理器
func NewQueueHandler(chargingRequestService *service.ChargingRequestService, systemService *service.SystemService, userService *service.UserService) *QueueHandler {
	return &QueueHandler{
		chargingRequestService: chargingRequestService,
		systemService:          systemService,
		userService:            userService,
	}
}

// GetQueueStatus 获取排队状态
func (h *QueueHandler) GetQueueStatus(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	mode := r.URL.Query().Get("mode")
	if mode == "" {
		mode = "all"
	}

	if mode != "all" && mode != "fast" && mode != "slow" {
		http.Error(w, "模式参数无效", http.StatusBadRequest)
		return
	}
	// 获取排队状态
	queueStatus, err := h.chargingRequestService.GetQueueStatus()
	if err != nil {
		http.Error(w, "获取排队状态失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	data := make(map[string]any)

	if mode == "all" || mode == "fast" {
		data["fastChargingQueue"] = map[string]any{
			"waiting":        len(queueStatus.FastQueue),
			"availableSlots": queueStatus.AvailableSlots,
		}
	}

	if mode == "all" || mode == "slow" {
		data["slowChargingQueue"] = map[string]any{
			"waiting":        len(queueStatus.SlowQueue),
			"availableSlots": queueStatus.AvailableSlots,
		}
	}

	response := model.Response{
		Code:      200,
		Message:   "success",
		Data:      data,
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetUserQueuePosition 获取用户排队位置
func (h *QueueHandler) GetUserQueuePosition(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	// 获取路径参数中的用户ID
	userIDStr := r.PathValue("userId")

	// 解析为UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "无效的用户ID", http.StatusBadRequest)
		return
	}

	// 检查权限
	if userID != user.ID && user.UserType != model.UserTypeAdmin {
		http.Error(w, "权限不足", http.StatusForbidden)
		return
	}

	// 获取最新充电请求
	request, err := h.chargingRequestService.GetActiveRequestByUserID(userID)
	if err != nil {
		http.Error(w, "获取用户排队信息失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if request == nil || (request.Status != model.RequestStatusWaiting && request.Status != model.RequestStatusQueued && request.Status != model.RequestStatusCharging) {
		response := model.Response{
			Code:      404,
			Message:   "用户当前没有排队中的充电请求",
			Timestamp: model.NowTimestamp(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	} // 计算前面的车辆数量
	var carsAhead int
	if request.Status == model.RequestStatusWaiting {
		// 对于等候区的请求，计算同一充电模式下更早创建的请求数量
		waitingRequests, err := h.chargingRequestService.GetWaitingRequestsByMode(request.ChargingMode)
		if err != nil {
			http.Error(w, "获取等候区队列信息失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 计算比当前请求更早的请求数量
		for _, waitingReq := range waitingRequests {
			if waitingReq.CreatedAt.Before(request.CreatedAt) {
				carsAhead++
			}
		}
		// 获取系统配置中的充电区队列长度
		config, err := h.systemService.GetSchedulingConfig()
		if err != nil {
			http.Error(w, "获取系统配置失败: "+err.Error(), http.StatusInternalServerError)
			return
		}

		// 加上充电区队列长度
		carsAhead += config.ChargingQueueLen
	} else {
		// 对于已分配到充电桩队列的请求，使用队列位置计算
		carsAhead = request.QueuePosition - 1 // 假设队列位置从1开始，前面有position-1辆车
	}

	// 构建队列信息
	queueInfo := map[string]any{
		"queueNumber":       request.QueueNumber,
		"position":          request.QueuePosition,
		"estimatedWaitTime": request.EstimatedWaitTime,
		"carsAhead":         carsAhead,
	}

	response := model.Response{
		Code:      200,
		Message:   "success",
		Data:      queueInfo,
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetWaitingVehicles 获取等候区车辆信息
func (h *QueueHandler) GetWaitingVehicles(w http.ResponseWriter, r *http.Request) {
	// 获取快充等候区车辆
	fastRequests, err := h.chargingRequestService.GetWaitingRequestsByMode(model.ChargingModeFast)
	if err != nil {
		http.Error(w, "获取快充等候区车辆失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取慢充等候区车辆
	slowRequests, err := h.chargingRequestService.GetWaitingRequestsByMode(model.ChargingModeSlow)
	if err != nil {
		http.Error(w, "获取慢充等候区车辆失败: "+err.Error(), http.StatusInternalServerError)
		return
	}
	// 从服务中获取用户服务
	userService := h.userService

	// 处理快充车辆信息
	var fastVehicles []map[string]any
	for _, req := range fastRequests {
		// 获取用户信息以获取车牌号
		user, err := userService.GetUserByID(req.UserID)
		if err != nil {
			// 如果获取用户信息失败，跳过该车辆
			continue
		}

		vehicleInfo := map[string]any{
			"licensePlate":      user.LicensePlate,
			"requestType":       "快充",
			"requestedCapacity": req.RequestedCapacity,
			"queueNumber":       req.QueueNumber,
			"createdAt":         req.CreatedAt,
		}
		fastVehicles = append(fastVehicles, vehicleInfo)
	}

	// 处理慢充车辆信息
	var slowVehicles []map[string]any
	for _, req := range slowRequests {
		// 获取用户信息以获取车牌号
		user, err := userService.GetUserByID(req.UserID)
		if err != nil {
			// 如果获取用户信息失败，跳过该车辆
			continue
		}

		vehicleInfo := map[string]any{
			"licensePlate":      user.LicensePlate,
			"requestType":       "慢充",
			"requestedCapacity": req.RequestedCapacity,
			"queueNumber":       req.QueueNumber,
			"createdAt":         req.CreatedAt,
		}
		slowVehicles = append(slowVehicles, vehicleInfo)
	}

	// 合并所有车辆并按创建时间排序
	allVehicles := append(fastVehicles, slowVehicles...)

	// 按创建时间排序 (从早到晚)
	for i := 0; i < len(allVehicles)-1; i++ {
		for j := i + 1; j < len(allVehicles); j++ {
			if allVehicles[i]["createdAt"].(time.Time).After(allVehicles[j]["createdAt"].(time.Time)) {
				allVehicles[i], allVehicles[j] = allVehicles[j], allVehicles[i]
			}
		}
	}

	// 构建响应数据
	data := map[string]any{
		"waitingVehicles": allVehicles,
		"totalCount":      len(allVehicles),
		"fastCount":       len(fastVehicles),
		"slowCount":       len(slowVehicles),
	}

	response := model.Response{
		Code:      200,
		Message:   "success",
		Data:      data,
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
