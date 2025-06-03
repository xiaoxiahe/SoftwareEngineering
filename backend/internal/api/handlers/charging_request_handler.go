package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/repository"
	"backend/internal/service"

	"github.com/google/uuid"
)

// ChargingRequestHandler 充电请求处理器
type ChargingRequestHandler struct {
	chargingRequestService *service.ChargingRequestService
	sessionRepo            *repository.ChargingSessionRepository
}

// NewChargingRequestHandler 创建充电请求处理器
func NewChargingRequestHandler(chargingRequestService *service.ChargingRequestService, sessionRepo *repository.ChargingSessionRepository) *ChargingRequestHandler {
	return &ChargingRequestHandler{
		chargingRequestService: chargingRequestService,
		sessionRepo:            sessionRepo,
	}
}

// CreateRequestRequest 创建请求请求参数
type CreateRequestRequest struct {
	ChargingMode      string  `json:"chargingMode"` // 充电模式：fast|slow
	RequestedCapacity float64 `json:"requestedCapacity"`
	Urgency           string  `json:"urgency,omitempty"` // 紧急程度：normal|urgent
}

// CreateRequest 创建充电请求
func (h *ChargingRequestHandler) CreateRequest(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	var req CreateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.ChargingMode != "fast" && req.ChargingMode != "slow" {
		http.Error(w, "充电模式无效", http.StatusBadRequest)
		return
	}

	if req.RequestedCapacity <= 0 {
		http.Error(w, "请求充电量必须大于0", http.StatusBadRequest)
		return
	}

	// 创建充电请求对象
	createReq := &model.ChargingRequestCreate{
		ChargingMode:      model.ChargingMode(req.ChargingMode),
		RequestedCapacity: req.RequestedCapacity,
	}

	// 提交充电请求
	request, err := h.chargingRequestService.CreateRequest(user.ID, createReq)
	if err != nil {
		http.Error(w, "创建充电请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建队列信息，因为GetUserPosition方法不存在
	queueInfo := map[string]any{
		"queueNumber":       request.QueueNumber,
		"estimatedWaitTime": request.EstimatedWaitTime,
		"waitingPosition":   request.QueuePosition,
	}

	response := model.Response{
		Code:    200,
		Message: "充电请求提交成功",
		Data: map[string]any{
			"requestId":         request.ID.String(),
			"queueNumber":       queueInfo["queueNumber"],
			"estimatedWaitTime": queueInfo["estimatedWaitTime"],
			"waitingPosition":   queueInfo["waitingPosition"],
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// UpdateRequest 更新充电请求
func (h *ChargingRequestHandler) UpdateRequest(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	requestIDStr := r.PathValue("requestId")
	if requestIDStr == "" {
		http.Error(w, "充电请求ID不能为空", http.StatusBadRequest)
		return
	}

	// 将字符串ID转换为UUID
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		http.Error(w, "无效的充电请求ID", http.StatusBadRequest)
		return
	}

	var req CreateRequestRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 创建更新请求对象
	updateReq := &model.ChargingRequestUpdate{}

	// 判断充电模式
	if req.ChargingMode != "" {
		if req.ChargingMode != "fast" && req.ChargingMode != "slow" {
			http.Error(w, "充电模式无效", http.StatusBadRequest)
			return
		}
		updateReq.ChargingMode = model.ChargingMode(req.ChargingMode)
	}

	// 判断请求充电量
	if req.RequestedCapacity > 0 {
		updateReq.RequestedCapacity = req.RequestedCapacity
	}

	// 更新充电请求
	request, err := h.chargingRequestService.UpdateRequest(user.ID, requestID, updateReq)
	if err != nil {
		http.Error(w, "更新充电请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建队列信息，因为没有GetUserPosition方法
	queueInfo := map[string]any{
		"queueNumber":       request.QueueNumber,
		"estimatedWaitTime": request.EstimatedWaitTime,
		"waitingPosition":   request.QueuePosition,
	}

	response := model.Response{
		Code:    200,
		Message: "充电请求修改成功",
		Data: map[string]any{
			"requestId":         request.ID.String(),
			"queueNumber":       queueInfo["queueNumber"],
			"estimatedWaitTime": queueInfo["estimatedWaitTime"],
			"waitingPosition":   queueInfo["waitingPosition"],
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CancelRequest 取消充电请求
func (h *ChargingRequestHandler) CancelRequest(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	requestIDStr := r.PathValue("requestId")
	if requestIDStr == "" {
		http.Error(w, "充电请求ID不能为空", http.StatusBadRequest)
		return
	}

	// 将字符串ID转换为UUID
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		http.Error(w, "无效的充电请求ID", http.StatusBadRequest)
		return
	}

	// 取消充电请求
	err = h.chargingRequestService.CancelRequest(user.ID, requestID)
	if err != nil {
		http.Error(w, "取消充电请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取最新请求状态，因为CancelRequest不返回请求对象
	request, err := h.chargingRequestService.GetRequestByID(requestID)
	if err != nil {
		http.Error(w, "获取请求信息失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.Response{
		Code:    200,
		Message: "充电请求已取消",
		Data: map[string]any{
			"requestId":   request.ID.String(),
			"cancelledAt": request.UpdatedAt,
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetRequestByID 获取指定充电请求
func (h *ChargingRequestHandler) GetRequestByID(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	requestIDStr := r.PathValue("requestId")
	if requestIDStr == "" {
		http.Error(w, "充电请求ID不能为空", http.StatusBadRequest)
		return
	}

	// 将字符串ID转换为UUID
	requestID, err := uuid.Parse(requestIDStr)
	if err != nil {
		http.Error(w, "无效的充电请求ID", http.StatusBadRequest)
		return
	}

	// 获取充电请求
	request, err := h.chargingRequestService.GetRequestByID(requestID)
	if err != nil {
		http.Error(w, "获取充电请求失败: "+err.Error(), http.StatusNotFound)
		return
	}

	// 检查权限 (使用UserType而不是Role)
	if request.UserID != user.ID && user.UserType != model.UserTypeAdmin {
		http.Error(w, "无权访问该充电请求", http.StatusForbidden)
		return
	}

	// 构建响应数据
	requestData := map[string]any{
		"requestId":         request.ID.String(),
		"status":            request.Status,
		"requestedCapacity": request.RequestedCapacity,
		"queueNumber":       request.QueueNumber,
		"queuePosition":     request.QueuePosition,
		"estimatedWaitTime": request.EstimatedWaitTime,
	}

	if request.PileID != "" {
		requestData["chargingPileId"] = request.PileID
	}

	// 添加充电开始和结束时间信息 (如果需要这些字段，需要确保模型中有这些字段)
	// 假设这些字段在模型中不存在，我们跳过添加

	response := model.Response{
		Code:      200,
		Message:   "success",
		Data:      requestData,
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetRequests 获取充电请求列表
func (h *ChargingRequestHandler) GetRequests(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	// 获取分页参数
	page := 1
	pageSize := 10

	// 获取充电请求列表 - 使用GetUserRequests方法
	requests, total, err := h.chargingRequestService.GetUserRequests(user.ID, page, pageSize)
	if err != nil {
		http.Error(w, "获取充电请求列表失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建响应数据
	var requestsData []map[string]any
	for _, req := range requests {
		requestData := map[string]any{
			"requestId":         req.ID.String(),
			"status":            req.Status,
			"requestedCapacity": req.RequestedCapacity,
			"queueNumber":       req.QueueNumber,
			"createdAt":         req.CreatedAt,
			"updatedAt":         req.UpdatedAt,
		}
		requestsData = append(requestsData, requestData)
	}

	response := model.Response{
		Code:    200,
		Message: "success",
		Data: map[string]any{
			"requests": requestsData,
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetLatestRequest 获取最新充电请求
func (h *ChargingRequestHandler) GetLatestRequest(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	// 获取最新充电请求 - 使用GetActiveRequestByUserID方法
	request, err := h.chargingRequestService.GetActiveRequestByUserID(user.ID)
	if err != nil {
		http.Error(w, "获取最新充电请求失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	if request == nil {
		response := model.Response{
			Code:      404,
			Message:   "当前没有活跃的充电请求",
			Timestamp: model.NowTimestamp(),
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		json.NewEncoder(w).Encode(response)
		return
	}
	// 构建响应数据
	requestData := map[string]any{
		"requestId":         request.ID.String(),
		"status":            request.Status,
		"requestedCapacity": request.RequestedCapacity,
		"queueNumber":       request.QueueNumber,
		"queuePosition":     request.QueuePosition,
		"estimatedWaitTime": request.EstimatedWaitTime,
		"createdAt":         request.CreatedAt,
		"updatedAt":         request.UpdatedAt,
	}

	if request.PileID != "" {
		requestData["chargingPileId"] = request.PileID
	}

	// 如果状态是 "charging"，获取实际充电量
	if request.Status == "charging" {
		session, err := h.sessionRepo.GetByRequestID(request.ID)
		if err == nil && session != nil {
			requestData["actualCapacity"] = session.ActualCapacity
		}
	}

	response := model.Response{
		Code:      200,
		Message:   "success",
		Data:      requestData,
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
