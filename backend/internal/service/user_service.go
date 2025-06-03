package service

import (
	"errors"
	"time"

	"backend/internal/model"
	"backend/internal/repository"

	"github.com/golang-jwt/jwt"
	"github.com/google/uuid"
)

const (
	jwtExpirationTime = 24 * time.Hour
	jwtSecretKey      = "your-secret-key-here" // 实际项目中应从配置文件读取
)

// UserService 用户服务
type UserService struct {
	userRepo *repository.UserRepository
}

// NewUserService 创建用户服务
func NewUserService(userRepo *repository.UserRepository) *UserService {
	return &UserService{
		userRepo: userRepo,
	}
}

// Register 用户注册
func (s *UserService) Register(req *model.CreateUserRequest) (*model.User, error) {
	return s.userRepo.Create(req)
}

// Login 用户登录
func (s *UserService) Login(req *model.LoginRequest) (*model.LoginResponse, error) {
	// 验证用户名和密码
	user, err := s.userRepo.VerifyPassword(req.Username, req.Password)
	if err != nil {
		return nil, err
	}

	// 生成JWT令牌
	token, err := s.generateJWT(user)
	if err != nil {
		return nil, err
	}

	// 设置过期时间
	expiresAt := time.Now().UTC().Add(jwtExpirationTime)

	// 保存会话到数据库
	err = s.userRepo.SaveSession(user.ID, token, expiresAt)
	if err != nil {
		return nil, err
	}

	// 返回登录响应
	return &model.LoginResponse{
		Token:     token,
		UserID:    user.ID,
		UserType:  user.UserType,
		ExpiresIn: int(jwtExpirationTime.Seconds()),
	}, nil
}

// Logout 用户登出
func (s *UserService) Logout(token string) error {
	return s.userRepo.InvalidateSession(token)
}

// GetUserByID 根据ID获取用户
func (s *UserService) GetUserByID(id uuid.UUID) (*model.User, error) {
	return s.userRepo.GetByID(id)
}

// UpdateUser 更新用户信息
func (s *UserService) UpdateUser(user *model.User) error {
	return s.userRepo.Update(user)
}

// GetUsers 获取用户列表
func (s *UserService) GetUsers(page, pageSize int) ([]*model.User, int, error) {
	return s.userRepo.List(page, pageSize)
}

// ValidateToken 验证JWT令牌
func (s *UserService) ValidateToken(tokenString string) (*uuid.UUID, error) {
	// 解析令牌
	token, err := jwt.Parse(tokenString, func(token *jwt.Token) (any, error) {
		// 验证签名方法
		if _, ok := token.Method.(*jwt.SigningMethodHMAC); !ok {
			return nil, errors.New("非法的签名方法")
		}
		return []byte(jwtSecretKey), nil
	})

	if err != nil {
		return nil, err
	}

	// 验证令牌有效性
	if !token.Valid {
		return nil, errors.New("无效的令牌")
	}

	// 从令牌中提取用户ID
	claims, ok := token.Claims.(jwt.MapClaims)
	if !ok {
		return nil, errors.New("无效的令牌声明")
	}

	userIDStr, ok := claims["user_id"].(string)
	if !ok {
		return nil, errors.New("令牌中没有用户ID")
	}

	// 将用户ID字符串转换为UUID
	_, err = uuid.Parse(userIDStr)
	if err != nil {
		return nil, err
	}

	// 从数据库验证会话有效性
	return s.userRepo.GetSessionByToken(tokenString)
}

// generateJWT 生成JWT令牌
func (s *UserService) generateJWT(user *model.User) (string, error) {
	// 设置JWT声明
	claims := jwt.MapClaims{
		"user_id":   user.ID.String(),
		"username":  user.Username,
		"user_type": user.UserType,
		"exp":       time.Now().UTC().Add(jwtExpirationTime).Unix(),
	}

	// 创建令牌
	token := jwt.NewWithClaims(jwt.SigningMethodHS256, claims)

	// 签名令牌
	return token.SignedString([]byte(jwtSecretKey))
}
