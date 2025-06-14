package services

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"simulator/internal/config"
	"simulator/internal/models"
	"simulator/internal/utils"
)

// APIClient 后端API客户端
type APIClient struct {
	client  *http.Client
	baseURL string
	logger  *utils.Logger
	clock   utils.Clock // 添加时钟对象
}

// NewAPIClient 创建API客户端
func NewAPIClient(cfg *config.Config, logger *utils.Logger, clock utils.Clock) *APIClient {
	return &APIClient{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: cfg.BackendAPI.BaseURL,
		logger:  logger,
		clock:   clock,
	}
}

// 充电桩状态上报请求
type PileStatusRequest struct {
	PileID         string `json:"pileId"`
	Status         string `json:"status"` // charging|available|fault|maintenance|offline
	CurrentVehicle *struct {
		UserID            string    `json:"userId"`
		StartTime         time.Time `json:"startTime"`
		RequestedCapacity float64   `json:"requestedCapacity"`
		CurrentCapacity   float64   `json:"currentCapacity"`
	} `json:"currentVehicle,omitempty"`
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
}

// UpdateChargingProgress 上报充电进度
func (c *APIClient) UpdateChargingProgress(pile *models.Pile) error {
	status, vehicle := pile.GetStatus()

	if status != models.PileStatusCharging || vehicle == nil {
		return fmt.Errorf("充电桩 %s 当前未在充电", pile.ID)
	}

	// 准备请求数据
	req := ChargingProgressRequest{
		PileID:            pile.ID,
		UserID:            vehicle.UserID,
		StartTime:         vehicle.StartTime,
		CurrentCapacity:   vehicle.CurrentCapacity,
		RequestedCapacity: vehicle.RequestedCapacity,
		ChargingRate:      pile.Power,           // kW
		RemainingTime:     pile.RemainingTime(), // 秒
	}

	// 发送请求
	return c.sendRequest("POST", "/api/v1/simulator/charging-progress", req)
}

// FaultReportRequest 故障报告请求
type FaultReportRequest struct {
	PileID      string `json:"pileId"`
	FaultType   string `json:"faultType"` // hardware|software|power
	Description string `json:"description"`
}

// ReportFault 报告故障
func (c *APIClient) ReportFault(pile *models.Pile, faultType models.FaultType, description string) error {
	// 准备请求数据
	req := FaultReportRequest{
		PileID:      pile.ID,
		FaultType:   string(faultType),
		Description: description,
	}

	// 发送请求
	return c.sendRequest("POST", "/api/v1/simulator/fault-report", req)
}

// HeartbeatRequest 心跳请求
type HeartbeatRequest struct {
	PileIDs   []string  `json:"pileIds"`
	Timestamp time.Time `json:"timestamp"`
}

// SendHeartbeat 发送心跳
func (c *APIClient) SendHeartbeat(pileIDs []string) error {
	// 准备请求数据
	req := HeartbeatRequest{
		PileIDs:   pileIDs,
		Timestamp: time.Now().UTC(),
	}

	// 发送请求
	return c.sendRequest("POST", "/api/v1/simulator/heartbeat", req)
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

// CompleteCharging 上报充电完成
func (c *APIClient) CompleteCharging(pile *models.Pile, vehicle *models.ChargingVehicle) error {
	if vehicle == nil {
		return fmt.Errorf("充电车辆信息为空")
	}

	// 使用业务逻辑时钟计算充电完成时间和总时长
	endTime := c.clock.Now()
	chargingDuration := int(endTime.Sub(vehicle.StartTime).Seconds())

	// 准备请求数据
	req := ChargingCompleteRequest{
		PileID:            pile.ID,
		UserID:            vehicle.UserID,
		StartTime:         vehicle.StartTime,
		EndTime:           endTime,
		RequestedCapacity: vehicle.RequestedCapacity,
		ActualCapacity:    vehicle.CurrentCapacity,
		ChargingDuration:  chargingDuration,
	}

	// 发送请求
	return c.sendRequest("POST", "/api/v1/simulator/charging-complete", req)
}

// FaultRecoveryRequest 故障恢复请求
type FaultRecoveryRequest struct {
	PileID string `json:"pileId"`
}

// RecoverFault 上报故障恢复
func (c *APIClient) RecoverFault(pile *models.Pile) error {
	// 准备请求数据
	req := FaultRecoveryRequest{
		PileID: pile.ID,
	}

	// 发送请求
	return c.sendRequest("POST", "/api/v1/simulator/fault-recovery", req)
}

// 发送HTTP请求的通用方法
func (c *APIClient) sendRequest(method, path string, payload any) error {
	url := c.baseURL + path

	// 序列化请求体
	jsonData, err := json.Marshal(payload)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %w", err)
	}

	// 创建请求
	req, err := http.NewRequest(method, url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	req.Header.Set("Content-Type", "application/json")

	// 记录请求
	c.logger.Info("发送请求: %s %s", method, url)

	// 发送请求
	resp, err := c.client.Do(req)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	// 检查响应状态
	if resp.StatusCode < 200 || resp.StatusCode >= 300 {
		var errorResp struct {
			Message string `json:"message"`
		}
		if err := json.NewDecoder(resp.Body).Decode(&errorResp); err == nil && errorResp.Message != "" {
			return fmt.Errorf("服务器返回错误(状态码: %d): %s", resp.StatusCode, errorResp.Message)
		}
		return fmt.Errorf("服务器返回错误(状态码: %d)", resp.StatusCode)
	}

	return nil
}
