package handlers

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"backend/internal/model"
	"backend/internal/service"
)

// SimulatorHandler 模拟器处理器
type SimulatorHandler struct {
	chargingPileService *service.ChargingPileService
	schedulerService    *service.SchedulerService
}

// NewSimulatorHandler 创建模拟器处理器
func NewSimulatorHandler(chargingPileService *service.ChargingPileService, schedulerService *service.SchedulerService) *SimulatorHandler {
	return &SimulatorHandler{
		chargingPileService: chargingPileService,
		schedulerService:    schedulerService,
	}
}

// PileStatusUpdateRequest 充电桩状态更新请求
type PileStatusUpdateRequest struct {
	PileID         string `json:"pileId"`
	Status         string `json:"status"` // charging|available|fault|maintenance|offline
	CurrentVehicle *struct {
		UserID            string    `json:"userId"`
		StartTime         time.Time `json:"startTime"`
		RequestedCapacity float64   `json:"requestedCapacity"`
		CurrentCapacity   float64   `json:"currentCapacity"`
	} `json:"currentVehicle,omitempty"`
	Queue []struct {
		UserID            string  `json:"userId"`
		QueueNumber       string  `json:"queueNumber"`
		RequestedCapacity float64 `json:"requestedCapacity"`
	} `json:"queue,omitempty"`
}

// ChargingProgressRequest 充电进度更新请求
type ChargingProgressRequest struct {
	PileID            string    `json:"pileId"`
	UserID            string    `json:"userId"`
	StartTime         time.Time `json:"startTime"`
	CurrentCapacity   float64   `json:"currentCapacity"`
	RequestedCapacity float64   `json:"requestedCapacity"`
	ChargingRate      float64   `json:"chargingRate"`  // 度/小时
	RemainingTime     int       `json:"remainingTime"` // 秒
	ReportTime        time.Time `json:"reportTime"`    // 上报时间
}

// UpdateChargingProgress 更新充电进度
func (h *SimulatorHandler) UpdateChargingProgress(w http.ResponseWriter, r *http.Request) {
	var req ChargingProgressRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.PileID == "" || req.UserID == "" {
		http.Error(w, "充电桩ID和用户ID不能为空", http.StatusBadRequest)
		return
	}

	err := h.schedulerService.UpdateChargingProgress(req.PileID, req.UserID, req.CurrentCapacity, req.RemainingTime, req.StartTime, req.ReportTime)

	if err != nil {
		http.Error(w, "更新充电进度失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.Response{
		Code:      200,
		Message:   "充电进度已更新",
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// FaultReportSimRequest 故障报告请求（模拟器版本）
type FaultReportSimRequest struct {
	PileID      string `json:"pileId"`
	FaultType   string `json:"faultType"` // hardware|software|power
	Description string `json:"description"`
}

// ReportFault 报告故障（模拟器）
func (h *SimulatorHandler) ReportFault(w http.ResponseWriter, r *http.Request) {
	var req FaultReportSimRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.PileID == "" || req.FaultType == "" {
		http.Error(w, "充电桩ID和故障类型不能为空", http.StatusBadRequest)
		return
	}

	// 报告故障到充电桩服务
	err := h.chargingPileService.ReportPileFault(req.PileID, req.FaultType, req.Description)
	if err != nil {
		http.Error(w, "报告故障失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 调用调度服务处理故障
	err = h.schedulerService.HandlePileFault(req.PileID, req.FaultType, req.Description)
	if err != nil {
		http.Error(w, "处理故障调度失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 创建一个故障记录ID
	faultID := "fault-" + time.Now().UTC().Format("20060102150405")

	fmt.Printf("故障报告成功，故障ID: %s\n", faultID)

	response := model.Response{
		Code:    200,
		Message: "故障已报告，正在重新调度",
		Data: map[string]any{
			"faultId": faultID,
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	PileIDs   []string  `json:"pileIds"`
	Timestamp time.Time `json:"timestamp"`
}

// Heartbeat 心跳处理
func (h *SimulatorHandler) Heartbeat(w http.ResponseWriter, r *http.Request) {
	var req HeartbeatRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	response := model.Response{
		Code:      200,
		Message:   "心跳成功",
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// FaultRecoveryRequest 故障恢复请求
type FaultRecoveryRequest struct {
	PileID string `json:"pileId"`
}

// RecoverFault 故障恢复（模拟器）
func (h *SimulatorHandler) RecoverFault(w http.ResponseWriter, r *http.Request) {
	var req FaultRecoveryRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.PileID == "" {
		http.Error(w, "充电桩ID不能为空", http.StatusBadRequest)
		return
	}

	// 调用调度服务处理故障恢复
	err := h.schedulerService.HandlePileRecovery(req.PileID)
	if err != nil {
		http.Error(w, "处理故障恢复失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.Response{
		Code:    200,
		Message: "故障已恢复，正在重新调度",
		Data: map[string]any{
			"pileId": req.PileID,
			"status": "recovered",
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ChargingCompleteRequest 充电完成请求
type ChargingCompleteRequest struct {
	PileID            string    `json:"pileId"`
	UserID            string    `json:"userId"`
	StartTime         time.Time `json:"startTime"`
	EndTime           time.Time `json:"endTime"`
	RequestedCapacity float64   `json:"requestedCapacity"`
	ActualCapacity    float64   `json:"actualCapacity"`
	ChargingDuration  int       `json:"chargingDuration"` // 秒
}

// CompleteCharging 处理充电完成请求
func (h *SimulatorHandler) CompleteCharging(w http.ResponseWriter, r *http.Request) {
	var req ChargingCompleteRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.PileID == "" || req.UserID == "" {
		http.Error(w, "充电桩ID和用户ID不能为空", http.StatusBadRequest)
		return
	}

	// 调用调度服务处理充电完成
	err := h.schedulerService.CompleteCharging(req.PileID, req.UserID, req.StartTime, req.EndTime, req.RequestedCapacity, req.ActualCapacity, req.ChargingDuration)
	if err != nil {
		http.Error(w, "处理充电完成失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.Response{
		Code:    200,
		Message: "充电完成处理成功",
		Data: map[string]any{
			"pileId":           req.PileID,
			"userId":           req.UserID,
			"actualCapacity":   req.ActualCapacity,
			"chargingDuration": req.ChargingDuration,
			"status":           "completed",
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
