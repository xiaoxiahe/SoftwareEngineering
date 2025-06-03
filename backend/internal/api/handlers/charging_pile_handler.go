package handlers

import (
	"encoding/json"
	"net/http"

	"backend/internal/model"
	"backend/internal/service"
)

// ChargingPileHandler 充电桩处理器
type ChargingPileHandler struct {
	chargingPileService *service.ChargingPileService
}

// NewChargingPileHandler 创建充电桩处理器
func NewChargingPileHandler(chargingPileService *service.ChargingPileService) *ChargingPileHandler {
	return &ChargingPileHandler{
		chargingPileService: chargingPileService,
	}
}

// GetAllPiles 获取所有充电桩
func (h *ChargingPileHandler) GetAllPiles(w http.ResponseWriter, r *http.Request) {
	// 获取所有充电桩
	allPiles, err := h.chargingPileService.GetAllPiles()
	if err != nil {
		http.Error(w, "获取充电桩数据失败:"+err.Error(), http.StatusInternalServerError)
		return
	}

	// 将充电桩按类型分类
	var fastPiles []*model.ChargingPile
	var slowPiles []*model.ChargingPile

	for _, pile := range allPiles {
		if pile.PileType == model.PileTypeFast {
			fastPiles = append(fastPiles, pile)
		} else {
			slowPiles = append(slowPiles, pile)
		}
	}

	// 构建响应数据
	var fastPilesData []map[string]any
	for _, pile := range fastPiles {
		// 处理快充桩数据
		fastPilesData = append(fastPilesData, map[string]any{
			"pileId": pile.ID,
			"status": pile.Status,
			"power":  pile.Power,
		})
	}

	var slowPilesData []map[string]any
	for _, pile := range slowPiles {
		// 处理慢充桩数据
		slowPilesData = append(slowPilesData, map[string]any{
			"pileId": pile.ID,
			"status": pile.Status,
			"power":  pile.Power,
		})
	}

	response := model.Response{
		Code:    200,
		Message: "success",
		Data: map[string]any{
			"fastChargingPiles": fastPilesData,
			"slowChargingPiles": slowPilesData,
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// ControlPileRequest 控制充电桩请求
type ControlPileRequest struct {
	Action string `json:"action"` // start|stop|maintenance
	Reason string `json:"reason"`
}

// ControlPile 控制充电桩（管理员）
func (h *ChargingPileHandler) ControlPile(w http.ResponseWriter, r *http.Request) {
	// 获取路径参数
	pileID := r.PathValue("pileId")
	if pileID == "" {
		http.Error(w, "充电桩ID不能为空", http.StatusBadRequest)
		return
	}

	var req ControlPileRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.Action != "start" && req.Action != "stop" && req.Action != "maintenance" {
		http.Error(w, "无效的操作类型", http.StatusBadRequest)
		return
	}
	// 执行操作
	var newStatus model.PileStatus
	switch req.Action {
	case "start":
		newStatus = model.PileStatusAvailable
	case "stop":
		newStatus = model.PileStatusOffline
	case "maintenance":
		newStatus = model.PileStatusMaintenance
	}

	// 更新充电桩状态
	err := h.chargingPileService.UpdatePileStatus(pileID, newStatus)
	if err != nil {
		http.Error(w, "更新充电桩状态失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.Response{
		Code:    200,
		Message: "充电桩状态已更新",
		Data: map[string]any{
			"pileId":    pileID,
			"status":    newStatus,
			"updatedAt": model.NowTimestamp(),
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetQueueVehicles 获取充电桩等候服务的车辆信息
func (h *ChargingPileHandler) GetQueueVehicles(w http.ResponseWriter, r *http.Request) {
	var allPiles []*model.ChargingPile
	var err error

	// 获取pileId参数，如果有的话
	pileId := r.URL.Query().Get("pileId")

	// 根据是否有pileId参数，获取对应充电桩或所有充电桩
	if pileId != "" {
		// 获取单个充电桩
		pile, err := h.chargingPileService.GetPileByID(pileId)
		if err != nil {
			http.Error(w, "获取充电桩数据失败: "+err.Error(), http.StatusNotFound)
			return
		}
		allPiles = []*model.ChargingPile{pile}
	} else {
		// 获取所有充电桩
		allPiles, err = h.chargingPileService.GetAllPiles()
		if err != nil {
			http.Error(w, "获取充电桩数据失败: "+err.Error(), http.StatusInternalServerError)
			return
		}
	}

	// 用于存储所有充电桩的等候车辆信息
	var pileQueueData []map[string]any

	// 从服务中获取用户仓库和队列仓库
	userRepo := h.chargingPileService.GetUserRepository()
	queueRepo := h.chargingPileService.GetQueueRepository()

	// 遍历所有充电桩，获取其队列信息
	for _, pile := range allPiles {
		queueItems, err := queueRepo.GetQueueItemsByPile(pile.ID)
		if err != nil {
			// 跳过出错的充电桩
			continue
		}

		var pileType string
		if pile.PileType == model.PileTypeFast {
			pileType = "fast"
		} else {
			pileType = "slow"
		}

		// 整理该充电桩的基本信息
		pileInfo := map[string]any{
			"pileId": pile.ID,
			"type":   pileType,
			"status": pile.Status,
			"power":  pile.Power,
		}

		// 整理等候车辆信息
		var queueVehicles []map[string]any
		for _, item := range queueItems {
			// 获取用户信息，主要是为了获取车辆电池容量
			user, err := userRepo.GetByID(item.UserID)
			if err != nil {
				// 如果获取用户信息失败，使用默认值或跳过
				continue
			}

			// 整理车辆信息
			vehicleInfo := map[string]any{
				"userId":            item.UserID.String(),
				"batteryCapacity":   user.BatteryCapacity,
				"requestedCapacity": item.RequestedCapacity,
				"queueTime":         item.WaitTime, // 排队时长（秒）
				"queuePosition":     item.Position,
				"queueNumber":       item.QueueNumber,
			}

			queueVehicles = append(queueVehicles, vehicleInfo)
		}

		pileInfo["queueVehicles"] = queueVehicles
		pileQueueData = append(pileQueueData, pileInfo)
	}

	// 构建响应
	response := model.Response{
		Code:      200,
		Message:   "success",
		Data:      map[string]any{"piles": pileQueueData},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
