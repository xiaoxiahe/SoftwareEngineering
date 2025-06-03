package handlers

import (
	"encoding/json"
	"net/http"

	"github.com/google/uuid"

	"backend/internal/middleware"
	"backend/internal/model"
	"backend/internal/service"
)

// UserHandler 用户处理器
type UserHandler struct {
	userService *service.UserService
}

// NewUserHandler 创建用户处理器
func NewUserHandler(userService *service.UserService) *UserHandler {
	return &UserHandler{
		userService: userService,
	}
}

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username    string `json:"username"`
	Password    string `json:"password"`
	VehicleInfo struct {
		LicensePlate    string  `json:"licensePlate"`
		BatteryCapacity float64 `json:"batteryCapacity"`
	} `json:"vehicleInfo"`
}

// Register 用户注册
func (h *UserHandler) Register(w http.ResponseWriter, r *http.Request) {
	var req RegisterRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.Username == "" || req.Password == "" || req.VehicleInfo.LicensePlate == "" || req.VehicleInfo.BatteryCapacity <= 0 {
		http.Error(w, "请求参数错误", http.StatusBadRequest)
		return
	}

	// 创建用户请求对象
	createReq := &model.CreateUserRequest{
		Username:        req.Username,
		Password:        req.Password,
		LicensePlate:    req.VehicleInfo.LicensePlate,
		BatteryCapacity: req.VehicleInfo.BatteryCapacity,
	}

	// 创建用户
	user, err := h.userService.Register(createReq)
	if err != nil {
		http.Error(w, "注册失败: "+err.Error(), http.StatusBadRequest)
		return
	}

	response := model.Response{
		Code:    200,
		Message: "注册成功",
		Data: map[string]any{
			"userId":   user.ID.String(),
			"username": user.Username,
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username"`
	Password string `json:"password"`
}

// Login 用户登录
func (h *UserHandler) Login(w http.ResponseWriter, r *http.Request) {
	var req LoginRequest
	if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
		http.Error(w, "无效的请求参数", http.StatusBadRequest)
		return
	}

	// 验证参数
	if req.Username == "" || req.Password == "" {
		http.Error(w, "请求参数错误", http.StatusBadRequest)
		return
	}

	// 创建登录请求对象
	loginReq := &model.LoginRequest{
		Username: req.Username,
		Password: req.Password,
	}

	// 登录用户
	loginResp, err := h.userService.Login(loginReq)
	if err != nil {
		http.Error(w, "登录失败: "+err.Error(), http.StatusUnauthorized)
		return
	}

	response := model.Response{
		Code:    200,
		Message: "登录成功",
		Data: map[string]any{
			"token":     loginResp.Token,
			"userId":    loginResp.UserID.String(),
			"userType":  loginResp.UserType,
			"expiresIn": loginResp.ExpiresIn,
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}

// GetUserInfo 获取用户信息
func (h *UserHandler) GetUserInfo(w http.ResponseWriter, r *http.Request) {
	// 从上下文获取用户
	user, ok := middleware.GetUserFromContext(r.Context())
	if !ok {
		http.Error(w, "未授权访问", http.StatusUnauthorized)
		return
	}

	// 检查路径参数中的用户ID
	userIDStr := r.PathValue("userId")
	// 解析为UUID
	userID, err := uuid.Parse(userIDStr)
	if err != nil {
		http.Error(w, "无效的用户ID", http.StatusBadRequest)
		return
	}

	// 检查权限 (假设 UserType 是 User 对象中的字段，而不是 Role)
	if userID != user.ID && user.UserType != "admin" {
		http.Error(w, "权限不足", http.StatusForbidden)
		return
	}

	// 获取用户信息
	userInfo, err := h.userService.GetUserByID(userID)
	if err != nil {
		http.Error(w, "获取用户信息失败: "+err.Error(), http.StatusInternalServerError)
		return
	}

	response := model.Response{
		Code:    200,
		Message: "success",
		Data: map[string]any{
			"userId":   userInfo.ID.String(),
			"username": userInfo.Username,
			"vehicleInfo": map[string]any{
				"licensePlate":    userInfo.LicensePlate,
				"batteryCapacity": userInfo.BatteryCapacity,
			},
			"createdAt": userInfo.CreatedAt,
		},
		Timestamp: model.NowTimestamp(),
	}

	w.Header().Set("Content-Type", "application/json")
	w.WriteHeader(http.StatusOK)
	json.NewEncoder(w).Encode(response)
}
