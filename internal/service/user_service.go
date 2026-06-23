package service

import (
	"fmt"

	"exam-quiz/internal/model"
	"exam-quiz/internal/repository"
	"exam-quiz/internal/util"
)

// AuthResult holds the authentication token and user information returned after login/register.
type AuthResult struct {
	Token string      `json:"token"`
	User  *model.User `json:"user"`
}

// Register creates a new user account and returns an auth token.
func Register(username, password, nickname string) (*AuthResult, error) {
	if len(username) < 3 || len(username) > 50 {
		return nil, fmt.Errorf("用户名长度需在 3-50 之间")
	}
	if len(password) < 6 || len(password) > 72 {
		return nil, fmt.Errorf("密码长度需在 6-72 之间")
	}
	if existing, _ := repository.GetUserByUsername(username); existing != nil {
		return nil, fmt.Errorf("用户名已存在")
	}
	hashed, err := util.HashPassword(password)
	if err != nil {
		return nil, fmt.Errorf("操作失败")
	}
	if nickname == "" {
		nickname = username
	}
	user := &model.User{Username: username, Password: hashed, Nickname: nickname, Role: "user"}
	if err := repository.CreateUser(user); err != nil {
		return nil, fmt.Errorf("注册失败")
	}
	token, err := util.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("注册成功但自动登录失败")
	}
	return &AuthResult{Token: token, User: user}, nil
}

// Login authenticates a user and returns an auth token.
func Login(username, password string) (*AuthResult, error) {
	user, err := repository.GetUserByUsername(username)
	if err != nil {
		return nil, fmt.Errorf("用户名或密码错误")
	}
	if !util.CheckPassword(password, user.Password) {
		return nil, fmt.Errorf("用户名或密码错误")
	}
	token, err := util.GenerateToken(user.ID, user.Username, user.Role)
	if err != nil {
		return nil, fmt.Errorf("登录失败")
	}
	return &AuthResult{Token: token, User: user}, nil
}

// GetProfile retrieves a user by ID.
func GetProfile(userID uint) (*model.User, error) {
	return repository.GetUserByID(userID)
}

// UpdateProfile updates a user's nickname.
func UpdateProfile(userID uint, nickname string) error {
	if len(nickname) > 50 {
		return fmt.Errorf("昵称长度不能超过 50 个字符")
	}
	user := &model.User{ID: userID, Nickname: nickname}
	return repository.UpdateUser(user)
}

// ChangePassword verifies the old password and sets a new one.
func ChangePassword(userID uint, oldPassword, newPassword string) error {
	if len(newPassword) < 6 || len(newPassword) > 72 {
		return fmt.Errorf("新密码长度需在 6-72 之间")
	}
	user, err := repository.GetUserByID(userID)
	if err != nil {
		return fmt.Errorf("获取用户信息失败")
	}
	if !util.CheckPassword(oldPassword, user.Password) {
		return fmt.Errorf("原密码错误")
	}
	hashed, err := util.HashPassword(newPassword)
	if err != nil {
		return fmt.Errorf("操作失败")
	}
	return repository.UpdatePassword(userID, hashed)
}

// ListUsers returns all users (admin only).
func ListUsers() ([]model.User, error) {
	return repository.ListUsers()
}
