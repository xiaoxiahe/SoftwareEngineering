package model

import (
	"time"

	"github.com/google/uuid"
)

// UserType 表示用户类型
type UserType string

const (
	UserTypeUser  UserType = "user"
	UserTypeAdmin UserType = "admin"
)

// User 表示系统中的用户
type User struct {
	ID              uuid.UUID `json:"id"`
	Username        string    `json:"username"`
	PasswordHash    string    `json:"-"` // 不包含在JSON响应中
	UserType        UserType  `json:"userType"`
	LicensePlate    string    `json:"licensePlate"`
	BatteryCapacity float64   `json:"batteryCapacity"`
	CreatedAt       time.Time `json:"createdAt"`
	UpdatedAt       time.Time `json:"updatedAt"`
}

// CreateUserRequest 创建用户的请求体
type CreateUserRequest struct {
	Username        string  `json:"username" binding:"required"`
	Password        string  `json:"password" binding:"required"`
	LicensePlate    string  `json:"licensePlate" binding:"required"`
	BatteryCapacity float64 `json:"batteryCapacity" binding:"required,gt=0"`
}

// LoginRequest 登录请求体
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// LoginResponse 登录响应体
type LoginResponse struct {
	Token     string    `json:"token"`
	UserID    uuid.UUID `json:"userId"`
	UserType  UserType  `json:"userType"`
	ExpiresIn int       `json:"expiresIn"`
}

// UserInfo 用户信息响应
type UserInfo struct {
	ID              uuid.UUID `json:"id"`
	Username        string    `json:"username"`
	UserType        UserType  `json:"userType"`
	LicensePlate    string    `json:"licensePlate"`
	BatteryCapacity float64   `json:"batteryCapacity"`
	CreatedAt       time.Time `json:"createdAt"`
}
