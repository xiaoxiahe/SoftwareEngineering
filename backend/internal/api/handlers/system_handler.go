package handlers

import (
	"encoding/json"
	"net/http"
	"time"

	"backend/internal/model"
	"backend/internal/service"
)

// SystemHandler 系统处理器
type SystemHandler struct {
	systemService *service.SystemService
}

// NewSystemHandler 创建系统处理器
func NewSystemHandler(systemService *service.SystemService) *SystemHandler {
	return &SystemHandler{
		systemService: systemService,
	}
}

// GetPileUsageReport 获取充电桩使用报表
func (h *SystemHandler) GetPileUsageReport(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	periodStr := r.URL.Query().Get("period")
	startDateStr := r.URL.Query().Get("startDate")
	endDateStr := r.URL.Query().Get("endDate")

	if periodStr == "" {
		periodStr = "day"
	}

	if periodStr != "day" && periodStr != "week" && periodStr != "month" {
		http.Error(w, "周期参数无效", http.StatusBadRequest)
		return
	}

	// 解析日期参数
	var startDate, endDate time.Time
	var err error

	if startDateStr == "" {
		// 默认开始日期为30天前
		startDate = time.Now().UTC().AddDate(0, 0, -30)
	} else {
		startDate, err = time.Parse("2006-01-02", startDateStr)
		if err != nil {
			http.Error(w, "开始日期格式错误", http.StatusBadRequest)
			return
		}
	}

	if endDateStr == "" {
		// 默认结束日期为今天
		now := time.Now().UTC()
		endDate = time.Date(now.Year(), now.Month(), now.Day(), 23, 59, 59, 999999999, now.Location())
	} else {
		endDate, err = time.Parse("2006-01-02", endDateStr)
		if err != nil {
			http.Error(w, "结束日期格式错误", http.StatusBadRequest)
			return
		}
		// 设置为当天23:59:59
		endDate = time.Date(endDate.Year(), endDate.Month(), endDate.Day(), 23, 59, 59, 999999999, endDate.Location())
	}

	// 调用系统服务获取充电桩使用统计
	stats, err := h.systemService.GetPileUsageReport(startDate, endDate, periodStr)
	if err != nil {
		http.Error(w, "获取充电桩使用报表失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.Response{
		Code:      200,
		Message:   "success",
		Data:      stats,
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetOperationStats 获取系统运营统计
func (h *SystemHandler) GetOperationStats(w http.ResponseWriter, r *http.Request) {
	// 获取查询参数
	periodStr := r.URL.Query().Get("period")
	dateStr := r.URL.Query().Get("date")

	if periodStr == "" {
		periodStr = "day"
	}

	if periodStr != "day" && periodStr != "week" && periodStr != "month" {
		http.Error(w, "周期参数无效", http.StatusBadRequest)
		return
	}

	// 解析日期参数
	var date time.Time
	var err error

	if dateStr == "" {
		// 默认日期为今天
		date = time.Now().UTC()
	} else {
		date, err = time.Parse("2006-01-02", dateStr)
		if err != nil {
			http.Error(w, "日期格式错误", http.StatusBadRequest)
			return
		}
	}

	// 调用系统服务获取系统运营统计
	stats, err := h.systemService.GetOperationStats(date, periodStr)
	if err != nil {
		http.Error(w, "获取系统运营统计失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.Response{
		Code:      200,
		Message:   "success",
		Data:      stats,
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
