package handlers

import (
	"encoding/json"
	"net/http"
	"strconv"
	"time"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"

	"github.com/google/uuid"
)

// BillingHandler 计费处理器
type BillingHandler struct {
	billingService *service.BillingService
}

// NewBillingHandler 创建计费处理器
func NewBillingHandler(billingService *service.BillingService) *BillingHandler {
	return &BillingHandler{
		billingService: billingService,
	}
}

// GetBillingDetails 获取充电详单列表
func (h *BillingHandler) GetBillingDetails(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	// 获取查询参数
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")
	pageStr := r.URL.Query().Get("page")
	pageSizeStr := r.URL.Query().Get("pageSize")

	// 解析参数（我们暂时不使用日期参数，但仍然验证其格式）
	if startDateStr != "" {
		_, err := time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "开始日期格式错误", http.StatusBadRequest)
			return
		}
	}

	if endDateStr != "" {
		_, err := time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "结束日期格式错误", http.StatusBadRequest)
			return
		}
	}

	// 解析分页参数
	page := 1
	if pageStr != "" {
		var err error
		page, err = strconv.Atoi(pageStr)
		if err != nil || page < 1 {
			page = 1
		}
	}

	pageSize := 10
	if pageSizeStr != "" {
		var err error
		pageSize, err = strconv.Atoi(pageSizeStr)
		if err != nil || pageSize < 1 {
			pageSize = 10
		}
	}

	// 获取充电详单
	details, total, err := h.billingService.GetUserBills(user.ID, page, pageSize)
	if err != nil {
		http.Error(w, "获取充电详单失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 构建响应数据
	var detailsData []map[string]any
	for _, detail := range details {
		detailData := map[string]any{
			"detailId":         detail.ID.String(),
			"pileId":           detail.PileID,
			"chargingCapacity": detail.ChargingCapacity,
			"chargingDuration": detail.ChargingDuration,
			"startTime":        detail.StartTime,
			"endTime":          detail.EndTime,
			"chargingFee":      detail.ChargingFee,
			"serviceFee":       detail.ServiceFee,
			"totalFee":         detail.TotalFee,
		}
		detailsData = append(detailsData, detailData)
	}

	response := model.Response{
		Code:    200,
		Message: "success",
		Data: map[string]any{
			"total":    total,
			"page":     page,
			"pageSize": pageSize,
			"details":  detailsData,
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetBillingDetailByID 获取单个详单
func (h *BillingHandler) GetBillingDetailByID(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	// 获取路径参数
	detailID := r.PathValue("detailId")
	if detailID == "" {
		http.Error(w, "详单ID不能为空", http.StatusBadRequest)
		return
	}

	// 将detailID解析为UUID
	detailUUID, err := uuid.Parse(detailID)
	if err != nil {
		http.Error(w, "无效的详单ID", http.StatusBadRequest)
		return
	}

	// 获取详单信息
	detail, err := h.billingService.GetBillByID(detailUUID)
	if err != nil {
		http.Error(w, "获取详单失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	// 检查权限
	if detail.UserID != user.ID && user.UserType != model.UserTypeAdmin {
		http.Error(w, "权限不足", http.StatusForbidden)
		return
	}

	// 构建响应数据
	detailData := map[string]any{
		"detailId":         detail.ID.String(),
		"userId":           detail.UserID.String(),
		"pileId":           detail.PileID,
		"chargingCapacity": detail.ChargingCapacity,
		"chargingDuration": detail.ChargingDuration,
		"startTime":        detail.StartTime,
		"endTime":          detail.EndTime,
		"unitPrice":        detail.UnitPrice,
		"serviceFeeRate":   0.8, // 服务费率固定为0.8元/度
		"chargingFee":      detail.ChargingFee,
		"serviceFee":       detail.ServiceFee,
		"totalFee":         detail.TotalFee,
		"priceType":        detail.PriceType,
	}

	response := model.Response{
		Code:      200,
		Message:   "success",
		Data:      detailData,
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// CalculateRequest 计算预估充电费用请求
type CalculateRequest struct {
	Capacity     float64   `json:"capacity"`
	ChargingMode string    `json:"chargingMode"` // fast|slow
	StartTime    time.Time `json:"startTime"`
}

// CalculateChargingFee 计算预估充电费用
func (h *BillingHandler) CalculateChargingFee(w http.ResponseWriter, r *http.Request) {
	var req CalculateRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.Capacity <= 0 {
		http.Error(w, "充电量必须大于0", http.StatusBadRequest)
		return
	}

	if req.ChargingMode != "fast" && req.ChargingMode != "slow" {
		http.Error(w, "无效的充电模式", http.StatusBadRequest)
		return
	}

	pileType := model.PileTypeSlow
	if req.ChargingMode == "fast" {
		pileType = model.PileTypeFast
	} // 从系统配置获取电价信息
	priceRate, err := h.billingService.GetCurrentPricing(req.StartTime)
	if err != nil {
		http.Error(w, "获取电价配置失败", http.StatusInternalServerError)
		return
	}

	// 计算充电时长（以小时为单位）
	// 假设充电功率为服务配置中的值
	chargingPower := 7.0 // 慢充桩充电功率7kW/h
	if pileType == model.PileTypeFast {
		chargingPower = 30.0 // 快充桩充电功率30kW/h
	}

	// 估算充电时长
	chargingDuration := req.Capacity / chargingPower

	// 计算充电费用
	chargingFee := req.Capacity * priceRate.ElectricFee

	// 计算服务费（从配置获取）
	serviceFee := req.Capacity * priceRate.ServiceFee

	// 总费用
	totalFee := chargingFee + serviceFee
	response := model.Response{
		Code:    200,
		Message: "success",
		Data: map[string]any{
			"capacity":         req.Capacity,
			"chargingDuration": chargingDuration,
			"unitPrice":        priceRate.ElectricFee,
			"priceType":        priceRate.Period,
			"chargingFee":      chargingFee,
			"serviceFee":       serviceFee,
			"totalFee":         totalFee,
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
