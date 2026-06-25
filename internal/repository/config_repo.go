package repository

import (
	"errors"

	"exam-quiz/internal/model"
	"gorm.io/gorm"
)

// GetConfig 获取系统配置值
func GetConfig(db *gorm.DB, key string) (string, error) {
	var config model.SystemConfig
	err := db.Where("key = ?", key).First(&config).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return "", nil
		}
		return "", err
	}
	return config.Value, nil
}

// SetConfig 设置系统配置值（不存在则创建）
func SetConfig(db *gorm.DB, key, value string) error {
	return db.Where("key = ?", key).Assign(model.SystemConfig{Value: value}).FirstOrCreate(&model.SystemConfig{}).Error
}
