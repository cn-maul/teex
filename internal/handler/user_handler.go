package handler

import (
	"log/slog"

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
		slog.Error("register failed", "error", err)
		response.HandleError(c, err)
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
		slog.Error("login failed", "error", err)
		response.HandleError(c, err)
		return
	}

	response.OK(c, gin.H{
		"token": result.Token,
		"user":  result.User,
	})
}

// GetProfile 获取当前用户信息
func GetProfile(c *gin.Context) {
	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	user, err := service.GetProfile(userID)
	if err != nil {
		slog.Error("get profile failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, user)
}

// ListUsers 获取用户列表（管理员）
func ListUsers(c *gin.Context) {
	users, err := service.ListUsers()
	if err != nil {
		slog.Error("list users failed", "error", err)
		response.HandleError(c, err)
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

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	if err := service.UpdateProfile(userID, req.Nickname); err != nil {
		slog.Error("update profile failed", "error", err)
		response.HandleError(c, err)
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

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	if err := service.ChangePassword(userID, req.OldPassword, req.NewPassword); err != nil {
		slog.Error("change password failed", "error", err)
		response.HandleError(c, err)
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
		slog.Error("set registration status failed", "error", err)
		response.HandleError(c, err)
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
	user, err := service.AdminCreateUser(req.Username, req.Password, req.Nickname, req.Role)
	if err != nil {
		slog.Error("admin create user failed", "error", err)
		response.HandleError(c, err)
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
	if err := service.AdminUpdateUser(id, req.Nickname, req.NewPassword, req.Role); err != nil {
		slog.Error("admin update user failed", "error", err)
		response.HandleError(c, err)
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
	currentAdminID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	if err := service.AdminDeleteUser(id, currentAdminID); err != nil {
		slog.Error("admin delete user failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OKWithMessage(c, nil, "用户已删除")
}
