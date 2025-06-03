package middleware

import (
	"context"
	"net/http"
	"strings"

	"backend/internal/model"
	"backend/internal/service"
)

// UserContextKey 用户上下文键
type UserContextKey string

// UserKey 用户键
const UserKey UserContextKey = "user"

// NewAuthMiddleware 创建认证中间件
func NewAuthMiddleware(userService *service.UserService) func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 从头部获取认证令牌
			authHeader := r.Header.Get("Authorization")
			if authHeader == "" {
				http.Error(w, "认证失败：未提供令牌", http.StatusUnauthorized)
				return
			}

			// 提取令牌
			tokenParts := strings.Split(authHeader, " ")
			if len(tokenParts) != 2 || tokenParts[0] != "Bearer" {
				http.Error(w, "认证失败：令牌格式错误", http.StatusUnauthorized)
				return
			}
			token := tokenParts[1]

			// 验证令牌
			userID, err := userService.ValidateToken(token)
			if err != nil {
				http.Error(w, "认证失败："+err.Error(), http.StatusUnauthorized)
				return
			}

			// 根据ID获取用户信息
			user, err := userService.GetUserByID(*userID)
			if err != nil {
				http.Error(w, "认证失败：获取用户信息失败", http.StatusUnauthorized)
				return
			}

			// 将用户信息添加到请求上下文
			ctx := context.WithValue(r.Context(), UserKey, user)
			next.ServeHTTP(w, r.WithContext(ctx))
		}
	}
}

// GetUserFromContext 从上下文获取用户
func GetUserFromContext(ctx context.Context) (*model.User, bool) {
	user, ok := ctx.Value(UserKey).(*model.User)
	return user, ok
}

// NewAdminMiddleware 创建管理员中间件
func NewAdminMiddleware() func(http.HandlerFunc) http.HandlerFunc {
	return func(next http.HandlerFunc) http.HandlerFunc {
		return func(w http.ResponseWriter, r *http.Request) {
			// 从上下文获取用户
			user, ok := GetUserFromContext(r.Context())
			if !ok {
				http.Error(w, "未授权访问", http.StatusUnauthorized)
				return
			}

			// 检查是否为管理员
			if user.UserType != "admin" {
				http.Error(w, "权限不足", http.StatusForbidden)
				return
			}

			next.ServeHTTP(w, r)
		}
	}
}
