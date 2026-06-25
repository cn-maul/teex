package repository

import (
	"errors"

	"exam-quiz/internal/apperr"
	"exam-quiz/internal/model"

	"gorm.io/gorm"
)

// CreateUser 创建用户
func CreateUser(db *gorm.DB, user *model.User) error {
	return db.Create(user).Error
}

// GetUserByUsername 按用户名查找用户
func GetUserByUsername(db *gorm.DB, username string) (*model.User, error) {
	var user model.User
	err := db.Where("username = ?", username).First(&user).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// GetUserByID 按 ID 查找用户
func GetUserByID(db *gorm.DB, id uint) (*model.User, error) {
	var user model.User
	err := db.First(&user, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("用户不存在")
		}
		return nil, err
	}
	return &user, nil
}

// ListUsers 列出所有用户（管理员用）
func ListUsers(db *gorm.DB) ([]model.User, error) {
	var users []model.User
	err := db.Order("id ASC").Find(&users).Error
	return users, err
}

// DeleteUser 删除用户（事务中先清理关联数据再删除用户）
func DeleteUser(db *gorm.DB, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
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
func UpdateUser(db *gorm.DB, user *model.User) error {
	return db.Model(&model.User{}).Where("id = ?", user.ID).Updates(map[string]interface{}{
		"nickname": user.Nickname,
	}).Error
}

// UpdatePassword 更新用户密码
func UpdatePassword(db *gorm.DB, userID uint, hashedPassword string) error {
	return db.Model(&model.User{}).Where("id = ?", userID).Update("password", hashedPassword).Error
}

// UpdateUserFields 按字段批量更新用户（单次 SQL，原子操作）
func UpdateUserFields(db *gorm.DB, id uint, updates map[string]interface{}) error {
	return db.Model(&model.User{}).Where("id = ?", id).Updates(updates).Error
}
