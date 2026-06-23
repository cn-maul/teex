package handler

import (
	"log"
	"strings"

	"exam-quiz/internal/response"
	"exam-quiz/internal/service"
	"exam-quiz/internal/validator"

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

// GetRegistrationStatus 查询注册开关状态（所有登录用户可调用）
func GetRegistrationStatus(c *gin.Context) {
	enabled := service.GetRegistrationEnabled()
	response.OK(c, gin.H{"enabled": enabled})
}

// SetRegistrationStatus 设置注册开关（仅管理员）
func SetRegistrationStatus(c *gin.Context) {
	var req struct {
		Enabled bool `json:"enabled"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}
	if err := service.SetRegistrationEnabled(req.Enabled); err != nil {
		log.Printf("SetRegistrationStatus error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	msg := "注册已关闭"
	if req.Enabled {
		msg = "注册已开放"
	}
	response.OKWithMessage(c, gin.H{"enabled": req.Enabled}, msg)
}

// AdminCreateUserRequest 管理员创建用户请求
type AdminCreateUserRequest struct {
	Username string `json:"username" binding:"required"`
	Password string `json:"password" binding:"required"`
	Nickname string `json:"nickname"`
	Role     string `json:"role"`
}

// AdminCreateUser 管理员创建用户
func AdminCreateUser(c *gin.Context) {
	var req AdminCreateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}
	if req.Role == "" {
		req.Role = "user"
	}
	user, err := service.AdminCreateUser(req.Username, req.Password, req.Nickname, req.Role)
	if err != nil {
		log.Printf("AdminCreateUser error: %v", err)
		errMsg := err.Error()
		safeErrors := []string{"用户名长度", "密码长度", "用户名已存在", "角色只能是"}
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
	response.Created(c, user)
}

// AdminUpdateUserRequest 管理员更新用户请求
type AdminUpdateUserRequest struct {
	Nickname    string `json:"nickname"`
	NewPassword string `json:"new_password"`
	Role        string `json:"role"`
}

// AdminUpdateUser 管理员更新用户
func AdminUpdateUser(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的用户 ID")
		return
	}
	var req AdminUpdateUserRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}
	if req.Nickname == "" && req.NewPassword == "" && req.Role == "" {
		response.Error(c, 400, "请提供需要修改的信息")
		return
	}
	if err := service.AdminUpdateUser(id, req.Nickname, req.NewPassword, req.Role); err != nil {
		log.Printf("AdminUpdateUser error: %v", err)
		errMsg := err.Error()
		safeErrors := []string{"昵称长度", "密码长度", "用户不存在", "角色只能是", "未提供"}
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
	response.OKWithMessage(c, nil, "用户信息已更新")
}

// AdminDeleteUser 管理员删除用户
func AdminDeleteUser(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的用户 ID")
		return
	}
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	currentAdminID := userIDRaw.(uint)
	if err := service.AdminDeleteUser(id, currentAdminID); err != nil {
		log.Printf("AdminDeleteUser error: %v", err)
		response.Error(c, 400, err.Error())
		return
	}
	response.OKWithMessage(c, nil, "用户已删除")
}
