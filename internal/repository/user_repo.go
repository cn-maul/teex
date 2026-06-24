package repository

import (
	"exam-quiz/internal/database"
	"exam-quiz/internal/model"

	"gorm.io/gorm"
)

// CreateUser 创建用户
func CreateUser(user *model.User) error {
	return database.DB.Create(user).Error
}

// GetUserByUsername 按用户名查找用户
func GetUserByUsername(username string) (*model.User, error) {
	var user model.User
	err := database.DB.Where("username = ?", username).First(&user).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// GetUserByID 按 ID 查找用户
func GetUserByID(id uint) (*model.User, error) {
	var user model.User
	err := database.DB.First(&user, id).Error
	if err != nil {
		return nil, err
	}
	return &user, nil
}

// ListUsers 列出所有用户（管理员用）
func ListUsers() ([]model.User, error) {
	var users []model.User
	err := database.DB.Order("id ASC").Find(&users).Error
	return users, err
}

// DeleteUser 删除用户（事务中先清理关联数据再删除用户）
func DeleteUser(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 1. 删除该用户的所有答题记录
		if err := tx.Where("user_id = ?", id).Delete(&model.UserAnswer{}).Error; err != nil {
			return err
		}
		// 2. 删除该用户的所有考试场次
		if err := tx.Where("user_id = ?", id).Delete(&model.ExamSession{}).Error; err != nil {
			return err
		}
		// 3. 删除用户
		if err := tx.Delete(&model.User{}, id).Error; err != nil {
			return err
		}
		return nil
	})
}

// UpdateUser 更新用户信息
func UpdateUser(user *model.User) error {
	return database.DB.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"nickname": user.Nickname,
	}).Error
}

// UpdatePassword 更新用户密码
func UpdatePassword(userID uint, hashedPassword string) error {
	return database.DB.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

// UpdateUserFields 按字段批量更新用户（单次 SQL，原子操作）
func UpdateUserFields(id uint, updates map[string]interface{}) error {
	return database.DB.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}
