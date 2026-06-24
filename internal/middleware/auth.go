package middleware

import (
	"net/http"
	"strings"

	"exam-quiz/internal/repository"
	"exam-quiz/internal/response"
	"exam-quiz/internal/util"

	"github.com/gin-gonic/gin"
)

// AuthRequired JWT 认证中间件
func AuthRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		authHeader := c.GetHeader("Authorization")
		if authHeader == "" {
			response.Error(c, http.StatusUnauthorized, "未登录")
			c.Abort()
			return
		}

		// 解析 Bearer token
		parts := strings.SplitN(authHeader, " ", 2)
		if len(parts) != 2 || parts[0] != "Bearer" {
			response.Error(c, http.StatusUnauthorized, "认证格式错误")
			c.Abort()
			return
		}

		claims, err := util.ParseToken(parts[1])
		if err != nil {
			response.Error(c, http.StatusUnauthorized, "登录已过期，请重新登录")
			c.Abort()
			return
		}

		// 将用户信息存入上下文
		c.Set("user_id", claims.UserID)
		c.Set("username", claims.Username)
		c.Set("role", claims.Role)
		c.Next()
	}
}

// UserOnly 中间件拒绝管理员访问（仅普通用户可用）
func UserOnly() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleRaw, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusForbidden, "需要登录")
			c.Abort()
			return
		}
		role, _ := roleRaw.(string)
		if role == "admin" {
			response.Error(c, http.StatusForbidden, "管理员不能参与答题")
			c.Abort()
			return
		}
		c.Next()
	}
}

// AdminRequired 管理员权限中间件（需在 AuthRequired 之后使用）
func AdminRequired() gin.HandlerFunc {
	return func(c *gin.Context) {
		roleRaw, exists := c.Get("role")
		if !exists {
			response.Error(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}
		role, ok := roleRaw.(string)
		if !ok || role != "admin" {
			response.Error(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}

		// Fast-path passed. Now verify against the database in case the
		// user's role was changed (or the account deleted) after the JWT
		// was issued.
		userIDRaw, exists := c.Get("user_id")
		if !exists {
			response.Error(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}
		userID, ok := userIDRaw.(uint)
		if !ok {
			response.Error(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}
		user, err := repository.GetUserByID(userID)
		if err != nil || user.Role != "admin" {
			response.Error(c, http.StatusForbidden, "需要管理员权限")
			c.Abort()
			return
		}

		c.Next()
	}
}
