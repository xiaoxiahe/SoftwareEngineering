package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"

	"github.com/google/uuid"
)

// QueueHandler 队列处理器
type QueueHandler struct {
	chargingRequestService *service.ChargingRequestService
	systemService          *service.SystemService
}

// NewQueueHandler 创建队列处理器
func NewQueueHandler(chargingRequestService *service.ChargingRequestService, systemService *service.SystemService) *QueueHandler {
	return &QueueHandler{
		chargingRequestService: chargingRequestService,
		systemService:          systemService,
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
