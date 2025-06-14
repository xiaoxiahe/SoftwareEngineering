package services

import (
	"encoding/json"
	"fmt"
	"net/http"
	"sync"
	"time"

	"simulator/internal/config"
	"simulator/internal/utils"
)

// ServerAPI 模拟器服务器端API
type ServerAPI struct {
	server           *http.Server
	pileService      *PileService
	config           *config.Config
	logger           *utils.Logger
	handlers         map[string]http.HandlerFunc
	mu               sync.Mutex
	onChargingAssign func(pileID, userID string, amount float64, mode string) error
}

// NewServerAPI 创建模拟器服务器API
func NewServerAPI(cfg *config.Config, pileService *PileService, logger *utils.Logger) *ServerAPI {
	api := &ServerAPI{
		config:      cfg,
		pileService: pileService,
		logger:      logger,
		handlers:    make(map[string]http.HandlerFunc),
	}

	// 注册处理函数
	api.registerHandlers()

	return api
}

// registerHandlers 注册HTTP处理函数
func (api *ServerAPI) registerHandlers() {
	// 充电指令接收API
	api.handlers["/api/simulator/charging/assign"] = api.handleChargingAssign
	api.handlers["/api/simulator/charging/stop"] = api.handleChargingStop
	api.handlers["/api/simulator/status"] = api.handleStatus
}

// 充电分配请求结构
type ChargingAssignRequest struct {
	PileID            string  `json:"pileId"`            // 充电桩ID
	UserID            string  `json:"userId"`            // 用户ID
	RequestedCapacity float64 `json:"requestedCapacity"` // 请求充电量
	ChargingMode      string  `json:"chargingMode"`      // 充电模式
}

// 停止充电请求结构
type ChargingStopRequest struct {
	PileID string `json:"pileId"` // 充电桩ID
	UserID string `json:"userId"` // 用户ID
	Reason string `json:"reason"` // 停止原因
}

// 处理充电分配
func (api *ServerAPI) handleChargingAssign(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	var req ChargingAssignRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.PileID == "" || req.UserID == "" || req.RequestedCapacity <= 0 {
		http.Error(w, "参数不完整", http.StatusBadRequest)
		return
	}

	api.logger.Info("接收到充电分配请求: 充电桩=%s, 用户=%s, 电量=%.1f, 模式=%s",
		req.PileID, req.UserID, req.RequestedCapacity, req.ChargingMode)

	// 调用回调函数
	var err error
	if api.onChargingAssign != nil {
		err = api.onChargingAssign(req.PileID, req.UserID, req.RequestedCapacity, req.ChargingMode)
	} else {
		// 默认实现，直接调用充电桩服务
		err = api.pileService.AssignVehicle(req.PileID, req.UserID, req.RequestedCapacity, req.ChargingMode)
	}
	if err != nil {
		http.Error(w, "处理充电分配失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取充电桩以获取开始时间
	pile, err := api.pileService.GetPile(req.PileID)
	if err != nil {
		http.Error(w, "获取充电桩信息失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 获取充电桩状态和车辆信息以获取真实的开始时间
	_, vehicle := pile.GetStatus()
	var actualStartTime time.Time
	if vehicle != nil {
		actualStartTime = vehicle.StartTime
	} else {
		actualStartTime = time.Now().UTC() // 备用时间
	}

	// 构造响应
	response := struct {
		Code      int       `json:"code"`
		Message   string    `json:"message"`
		StartTime time.Time `json:"startTime"` // 添加实际开始时间
		Timestamp int64     `json:"timestamp"`
	}{
		Code:      200,
		Message:   "充电分配成功",
		StartTime: actualStartTime,
		Timestamp: time.Now().UTC().Unix(),
	}
	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	api.logger.Info("发送充电分配响应: %v", response)
}

// 处理停止充电
func (api *ServerAPI) handleChargingStop(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodPost {
		http.Error(w, "只支持POST方法", http.StatusMethodNotAllowed)
		return
	}

	var req ChargingStopRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.PileID == "" || req.UserID == "" {
		http.Error(w, "充电桩ID和用户ID不能为空", http.StatusBadRequest)
		return
	}

	api.logger.Info("接收到停止充电请求: 充电桩=%s, 用户=%s, 原因=%s",
		req.PileID, req.UserID, req.Reason)

	// 调用充电桩服务停止充电
	err := api.pileService.StopCharging(req.PileID, req.UserID, req.Reason)
	if err != nil {
		http.Error(w, "停止充电失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 构造响应
	response := struct {
		Code      int    `json:"code"`
		Message   string `json:"message"`
		Timestamp int64  `json:"timestamp"`
	}{
		Code:      200,
		Message:   "停止充电成功",
		Timestamp: time.Now().UTC().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
	api.logger.Info("发送停止充电响应: %v", response)
}

// 处理状态查询
func (api *ServerAPI) handleStatus(w http.ResponseWriter, r *http.Request) {
	if r.Method != http.MethodGet {
		http.Error(w, "只支持GET方法", http.StatusMethodNotAllowed)
		return
	}

	// 获取所有充电桩状态
	piles := api.pileService.GetAllPiles()
	status := make([]map[string]any, 0, len(piles))

	for _, pile := range piles {
		pileStatus, vehicle := pile.GetStatus()

		pileInfo := map[string]any{
			"id":     pile.ID,
			"type":   pile.Type,
			"status": pileStatus,
			"power":  pile.Power,
		}

		if vehicle != nil {
			pileInfo["currentVehicle"] = map[string]any{
				"userId":            vehicle.UserID,
				"startTime":         vehicle.StartTime,
				"requestedCapacity": vehicle.RequestedCapacity,
				"currentCapacity":   vehicle.CurrentCapacity,
				"remainingTime":     pile.RemainingTime(),
			}
		}

		status = append(status, pileInfo)
	}

	// 构造响应
	response := struct {
		Code      int              `json:"code"`
		Message   string           `json:"message"`
		Data      []map[string]any `json:"data"`
		Timestamp int64            `json:"timestamp"`
	}{
		Code:      200,
		Message:   "获取状态成功",
		Data:      status,
		Timestamp: time.Now().UTC().Unix(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// SetOnChargingAssign 设置充电分配回调函数
func (api *ServerAPI) SetOnChargingAssign(callback func(pileID, userID string, amount float64, mode string) error) {
	api.mu.Lock()
	defer api.mu.Unlock()
	api.onChargingAssign = callback
}

// Start 启动服务器
func (api *ServerAPI) Start(port int) error {
	mux := http.NewServeMux()

	// 注册路由处理器
	for pattern, handler := range api.handlers {
		mux.HandleFunc(pattern, handler)
	}

	// 创建服务器
	api.server = &http.Server{
		Addr:    fmt.Sprintf(":%d", port),
		Handler: mux,
	}

	// 启动服务器
	api.logger.Info("启动模拟器服务器，监听端口: %d", port)
	return api.server.ListenAndServe()
}

// Stop 停止服务器
func (api *ServerAPI) Stop() error {
	if api.server == nil {
		return nil
	}
	api.logger.Info("停止模拟器服务器")
	return api.server.Close()
}
