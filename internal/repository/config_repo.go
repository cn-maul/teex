package repository

import (
	"exam-quiz/internal/database"
	"exam-quiz/internal/model"
	"gorm.io/gorm"
)

// GetConfig 获取系统配置值
func GetConfig(key string) (string, error) {
	var config model.SystemConfig
	err := database.DB.Where("key = ?", key).First(&config).Error
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			return "", nil
		}
		return "", err
	}
	return config.Value, nil
}

// SetConfig 设置系统配置值（不存在则创建）
func SetConfig(key, value string) error {
	return database.DB.Where("key = ?", key).Assign(model.SystemConfig{Value: value}).FirstOrCreate(&model.SystemConfig{}).Error
}
