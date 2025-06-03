package service

import (
	"bytes"
	"encoding/json"
	"fmt"
	"net/http"
	"time"
)

// ChargingDispatcherClient 充电派发客户端
// 用于从模拟器向充电桩发送充电指令
type ChargingDispatcherClient struct {
	client  *http.Client
	baseURL string
}

// NewChargingDispatcherClient 创建充电派发客户端
func NewChargingDispatcherClient(simulatorBaseURL string) *ChargingDispatcherClient {
	return &ChargingDispatcherClient{
		client:  &http.Client{Timeout: 10 * time.Second},
		baseURL: simulatorBaseURL, // 例如："http://localhost:8090"
	}
}

// ChargingAssignRequest 充电分配请求
type ChargingAssignRequest struct {
	PileID            string  `json:"pileId"`            // 充电桩ID
	UserID            string  `json:"userId"`            // 用户ID
	RequestedCapacity float64 `json:"requestedCapacity"` // 请求充电量
	ChargingMode      string  `json:"chargingMode"`      // 充电模式(fast/trickle)
}

// ChargingStopRequest 停止充电请求
type ChargingStopRequest struct {
	PileID string `json:"pileId"` // 充电桩ID
	UserID string `json:"userId"` // 用户ID
	Reason string `json:"reason"` // 停止原因
}

// AssignCharging 分配充电
// 后端调用此方法向模拟器发送充电指令
func (c *ChargingDispatcherClient) AssignCharging(pileID, userID string, capacity float64, mode string) error {
	req := ChargingAssignRequest{
		PileID:            pileID,
		UserID:            userID,
		RequestedCapacity: capacity,
		ChargingMode:      mode,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %w", err)
	}

	url := c.baseURL + "/api/simulator/charging/assign"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

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

// StopCharging 停止充电
// 后端调用此方法向模拟器发送停止充电指令
func (c *ChargingDispatcherClient) StopCharging(pileID, userID, reason string) error {
	req := ChargingStopRequest{
		PileID: pileID,
		UserID: userID,
		Reason: reason,
	}

	jsonData, err := json.Marshal(req)
	if err != nil {
		return fmt.Errorf("序列化请求数据失败: %w", err)
	}

	url := c.baseURL + "/api/simulator/charging/stop"
	httpReq, err := http.NewRequest("POST", url, bytes.NewBuffer(jsonData))
	if err != nil {
		return fmt.Errorf("创建HTTP请求失败: %w", err)
	}
	httpReq.Header.Set("Content-Type", "application/json")

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

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

// GetSimulatorStatus 获取模拟器状态
func (c *ChargingDispatcherClient) GetSimulatorStatus() ([]map[string]any, error) {
	url := c.baseURL + "/api/simulator/status"
	httpReq, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return nil, fmt.Errorf("创建HTTP请求失败: %w", err)
	}

	resp, err := c.client.Do(httpReq)
	if err != nil {
		return nil, fmt.Errorf("发送HTTP请求失败: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		return nil, fmt.Errorf("服务器返回错误(状态码: %d)", resp.StatusCode)
	}

	var response struct {
		Code    int              `json:"code"`
		Message string           `json:"message"`
		Data    []map[string]any `json:"data"`
	}

	if err := json.NewDecoder(resp.Body).Decode(&response); err != nil {
		return nil, fmt.Errorf("解析响应失败: %w", err)
	}

	return response.Data, nil
}
