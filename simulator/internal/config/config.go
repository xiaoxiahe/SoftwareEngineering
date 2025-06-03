package config

import (
	"encoding/json"
	"os"
)

// Config 模拟器配置结构
type Config struct {
	// 充电桩配置
	Piles struct {
		Fast struct {
			Count int     `json:"count"` // 快充数量
			Power float64 `json:"power"` // 快充功率(kW)
		} `json:"fast"`
		Trickle struct {
			Count int     `json:"count"` // 慢充数量
			Power float64 `json:"power"` // 慢充功率(kW)
		} `json:"trickle"`
	} `json:"piles"`

	// 故障模拟配置
	Fault struct {
		RandomFault  bool    `json:"randomFault"`  // 是否启用随机故障
		FaultChance  float64 `json:"faultChance"`  // 故障概率 (0-1)
		MaxFaultTime int     `json:"maxFaultTime"` // 最大故障持续时间(分钟)
		MinFaultTime int     `json:"minFaultTime"` // 最小故障持续时间(分钟)
	} `json:"fault"`

	// 后端API设置
	BackendAPI struct {
		BaseURL           string `json:"baseURL"`           // API基础URL
		StatusInterval    int    `json:"statusInterval"`    // 状态上报间隔(秒)
		ProgressInterval  int    `json:"progressInterval"`  // 进度上报间隔(秒)
		HeartbeatInterval int    `json:"heartbeatInterval"` // 心跳间隔(秒)
	} `json:"backendAPI"`

	// 模拟设置
	Simulation struct {
		SpeedFactor float64 `json:"speedFactor"` // 模拟加速比例
		LogLevel    string  `json:"logLevel"`    // 日志级别
	} `json:"simulation"`
}

// LoadConfig 从指定路径加载配置
func LoadConfig(path string) (*Config, error) {
	file, err := os.Open(path)
	if err != nil {
		return nil, err
	}
	defer file.Close()

	config := &Config{}
	decoder := json.NewDecoder(file)
	if err := decoder.Decode(config); err != nil {
		return nil, err
	}

	return config, nil
}

// SaveConfig 保存配置到指定路径
func SaveConfig(config *Config, path string) error {
	file, err := os.Create(path)
	if err != nil {
		return err
	}
	defer file.Close()

	encoder := json.NewEncoder(file)
	encoder.SetIndent("", "  ")
	if err := encoder.Encode(config); err != nil {
		return err
	}

	return nil
}
