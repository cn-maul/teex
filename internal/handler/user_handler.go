package handler

import (
	"log"
	"strings"

	"exam-quiz/internal/response"
	"exam-quiz/internal/service"

	"github.com/gin-gonic/gin"
)

// RegisterRequest 注册请求
type RegisterRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
}

// LoginRequest 登录请求
type LoginRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
}

// Register 用户注册
func Register(c *gin.Context) {
	var req RegisterRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	result, err := service.Register(req.Username, req.Password, req.Nickname)
	if err != nil {
		log.Printf("Register error: %v", err)
		errMsg := err.Error()
		safeErrors := []string{"用户名长度", "密码长度", "用户名已存在"}
		safe := false
		for _, se := range safeErrors {
			if strings.Contains(errMsg, se) {
				safe = true
				break
			}
		}
		if !safe {
			errMsg = "操作失败，请稍后重试"
		}
		response.Error(c, 400, errMsg)
		return
	}

	response.OKWithMessage(c, gin.H{
		"token": result.Token,
		"user":  result.User,
	}, "注册成功")
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	result, err := service.Login(req.Username, req.Password)
	if err != nil {
		log.Printf("Login error: %v", err)
		response.Error(c, 401, err.Error())
		return
	}

	response.OK(c, gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}

// GetProfile 获取当前用户信息
func GetProfile(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	user, err := service.GetProfile(userID)
	if err != nil {
		response.Error(c, 500, "获取用户信息失败")
		return
	}
	response.OK(c, user)
}

// ListUsers 获取用户列表（管理员）
func ListUsers(c *gin.Context) {
	users, err := service.ListUsers()
	if err != nil {
		log.Printf("ListUsers error: %v", err)
		response.Error(c, 500, "操作失败")
		return
	}
	response.OK(c, users)
}

// UpdateProfileRequest 修改昵称请求
type UpdateProfileRequest struct {
	Nickname string `json:"nickname" binding:"required"`
}

// UpdateProfile 修改当前用户昵称
func UpdateProfile(c *gin.Context) {
	var req UpdateProfileRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	if err := service.UpdateProfile(userID, req.Nickname); err != nil {
		log.Printf("UpdateProfile error: %v", err)
		response.Error(c, 400, err.Error())
		return
	}
	response.OKWithMessage(c, nil, "昵称修改成功")
}

// ChangePasswordRequest 修改密码请求
type ChangePasswordRequest struct {
	OldPassword string `json:"old_password" binding:"required"`
	NewPassword string `json:"new_password" binding:"required"`
}

// ChangePassword 修改密码
func ChangePassword(c *gin.Context) {
	var req ChangePasswordRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	if err := service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		log.Printf("ChangePassword error: %v", err)
		response.Error(c, 400, err.Error())
		return
	}
	response.OKWithMessage(c, nil, "密码修改成功")
}
