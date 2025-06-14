package config

import (
	"encoding/json"
	"os"
)

// Config 应用程序配置结构
type Config struct {
	Server   ServerConfig   `json:"server"`
	Database DatabaseConfig `json:"database"`
	Auth     AuthConfig     `json:"auth"`
	Charging ChargingConfig `json:"charging"`
	Pricing  PricingConfig  `json:"pricing"`
}

// ServerConfig 服务器配置
type ServerConfig struct {
	Port            int `json:"port"`
	ReadTimeout     int `json:"readTimeout"`
	WriteTimeout    int `json:"writeTimeout"`
	IdleTimeout     int `json:"idleTimeout"`
	ShutdownTimeout int `json:"shutdownTimeout"`
}

// DatabaseConfig 数据库配置
type DatabaseConfig struct {
	Host     string `json:"host"`
	Port     int    `json:"port"`
	User     string `json:"user"`
	Password string `json:"password"`
	DBName   string `json:"dbName"`
	SSLMode  string `json:"sslMode"`
}

// AuthConfig 认证配置
type AuthConfig struct {
	JWTSecret        string `json:"jwtSecret"`
	JWTExpirationMin int    `json:"jwtExpirationMin"`
}

// ChargingConfig 充电系统配置
type ChargingConfig struct {
	FastChargingPileNum    int     `json:"fastChargingPileNum"`
	TrickleChargingPileNum int     `json:"trickleChargingPileNum"`
	WaitingAreaSize        int     `json:"waitingAreaSize"`
	ChargingQueueLen       int     `json:"chargingQueueLen"`
	FastChargingPower      float64 `json:"fastChargingPower"`
	TrickleChargingPower   float64 `json:"trickleChargingPower"`
	ServiceFeePerUnit      float64 `json:"serviceFeePerUnit"`
	ExtendedSchedulingMode string  `json:"extendedSchedulingMode"` // "disabled", "batch", "singleOptimal"
}

// PricingConfig 计价配置
type PricingConfig struct {
	PeakPrice     float64 `json:"peakPrice"`
	NormalPrice   float64 `json:"normalPrice"`
	ValleyPrice   float64 `json:"valleyPrice"`
	ServiceFee    float64 `json:"serviceFee"`
	PeakStartTime [][]int `json:"peakStartTime"` // 格式为 [小时, 分钟]
	PeakEndTime   [][]int `json:"peakEndTime"`   // 格式为 [小时, 分钟]
	FlatStartTime [][]int `json:"flatStartTime"` // 格式为 [小时, 分钟]
	FlatEndTime   [][]int `json:"flatEndTime"`   // 格式为 [小时, 分钟]
	ValleyStart   [][]int `json:"valleyStart"`   // 格式为 [小时, 分钟]
	ValleyEnd     [][]int `json:"valleyEnd"`     // 格式为 [小时, 分钟]
}

// Load 从文件加载配置
func Load(configPath string) (*Config, error) {
	data, err := os.ReadFile(configPath)
	if err != nil {
		return nil, err
	}

	var config Config
	err = json.Unmarshal(data, &config)
	if err != nil {
		return nil, err
	}

	return &config, nil
}
