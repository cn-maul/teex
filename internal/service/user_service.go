package service

import (
	"fmt"
	"log"

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
	// 检查注册开关
	enabled, err := repository.GetConfig("registration_enabled")
	if err != nil {
		log.Printf("WARNING: failed to read registration config: %v", err)
	}
	if enabled != "true" {
		return nil, fmt.Errorf("注册已关闭，请联系管理员创建账号")
	}

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

// GetRegistrationEnabled 查询注册开关状态
func GetRegistrationEnabled() bool {
	val, err := repository.GetConfig("registration_enabled")
	if err != nil {
		log.Printf("WARNING: failed to read registration config: %v", err)
	}
	return val == "true"
}

// SetRegistrationEnabled 设置注册开关
func SetRegistrationEnabled(enabled bool) error {
	val := "false"
	if enabled {
		val = "true"
	}
	return repository.SetConfig("registration_enabled", val)
}

// AdminCreateUser 管理员创建用户
func AdminCreateUser(username, password, nickname, role string) (*model.User, error) {
	if len(username) < 3 || len(username) > 50 {
		return nil, fmt.Errorf("用户名长度需在 3-50 之间")
	}
	if len(password) < 6 || len(password) > 72 {
		return nil, fmt.Errorf("密码长度需在 6-72 之间")
	}
	if role != "user" && role != "admin" {
		return nil, fmt.Errorf("角色只能是 user 或 admin")
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
	user := &model.User{
		Username: username,
		Password: hashed,
		Nickname: nickname,
		Role:     role,
	}
	if err := repository.CreateUser(user); err != nil {
		return nil, fmt.Errorf("创建用户失败")
	}
	return user, nil
}

// AdminUpdateUser 管理员更新用户（昵称、密码、角色）
func AdminUpdateUser(id uint, nickname, newPassword, role string) error {
	user, err := repository.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	updates := make(map[string]interface{})
	if nickname != "" {
		if len(nickname) > 50 {
			return fmt.Errorf("昵称长度不能超过 50 个字符")
		}
		updates["nickname"] = nickname
	}
	if newPassword != "" {
		if len(newPassword) < 6 || len(newPassword) > 72 {
			return fmt.Errorf("密码长度需在 6-72 之间")
		}
		hashed, err := util.HashPassword(newPassword)
		if err != nil {
			return fmt.Errorf("操作失败")
		}
		updates["password"] = hashed
	}
	if role != "" && role != user.Role {
		if role != "user" && role != "admin" {
			return fmt.Errorf("角色只能是 user 或 admin")
		}
		updates["role"] = role
	}
	if len(updates) == 0 {
		return fmt.Errorf("未提供需要更新的信息")
	}
	return repository.UpdateUserFields(id, updates)
}

// AdminDeleteUser 管理员删除用户（不能删除自己）
func AdminDeleteUser(id uint, currentAdminID uint) error {
	if id == currentAdminID {
		return fmt.Errorf("不能删除自己的账号")
	}
	_, err := repository.GetUserByID(id)
	if err != nil {
		return fmt.Errorf("用户不存在")
	}
	return repository.DeleteUser(id)
}
