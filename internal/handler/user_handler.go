package handler

import (
	"log"
	"net/http"

	"exam-quiz/internal/model"
	"exam-quiz/internal/repository"
	"exam-quiz/internal/util"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	if len(req.Username) < 3 || len(req.Username) > 50 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名长度需在 3-50 之间"})
		return
	}
	if len(req.Password) < 6 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "密码长度不能少于 6 位"})
		return
	}

	// 检查用户名是否已存在
	if existing, _ := repository.GetUserByUsername(req.Username); existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "用户名已存在"})
		return
	}

	hashedPassword, err := util.HashPassword(req.Password)
	if err != nil {
		log.Printf("Register hash error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	nickname := req.Nickname
	if nickname == "" {
		nickname = req.Username
	}

	user := &model.User{
		Username: req.Username,
		Password: hashedPassword,
		Nickname: nickname,
		Role:     "user",
	}
	if err := repository.CreateUser(user); err != nil {
		log.Printf("Register error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册失败，请稍后重试"})
		return
	}

	// 自动登录，返回 token
	token, err := util.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		log.Printf("Register token error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "注册成功但自动登录失败"})
		return
	}

	c.JSON(http.StatusCreated, gin.H{
		"data": gin.H{
			"token": token,
			"user":  user,
		},
		"message": "注册成功",
	})
}

// Login 用户登录
func Login(c *gin.Context) {
	var req LoginRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	user, err := repository.GetUserByUsername(req.Username)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
			return
		}
		log.Printf("Login error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	if !util.CheckPassword(req.Password, user.Password) {
		c.JSON(http.StatusUnauthorized, gin.H{"error": "用户名或密码错误"})
		return
	}

	token, err := util.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		log.Printf("Login token error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "登录失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data": gin.H{
			"token": token,
			"user":  user,
		},
	})
}

// GetProfile 获取当前用户信息
func GetProfile(c *gin.Context) {
	userID, _ := c.Get("user_id")
	user, err := repository.GetUserByID(userID.(uint))
	if err != nil {
		c.JSON(http.StatusInternalServerError, gin.H{"error": "获取用户信息失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": user})
}

// ListUsers 获取用户列表（管理员）
func ListUsers(c *gin.Context) {
	users, err := repository.ListUsers()
	if err != nil {
		log.Printf("ListUsers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": users})
}
