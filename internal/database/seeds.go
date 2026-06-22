package database

import (
	_ "embed"
	"fmt"
	"log"

	"exam-quiz/internal/model"
	"exam-quiz/internal/util"

	"gorm.io/gorm"
)

//go:embed seeddata/exams.json
var examsSeedData []byte

//go:embed seeddata/questions_sample.json
var questionsSeedData []byte

// Seed 从嵌入的种子数据中初始化考试类型、模块、题目
func Seed() error {
	var count int64
	DB.Model(&model.ExamType{}).Count(&count)
	if count > 0 {
		// 数据库已初始化，但仍需确保管理员账户存在
		ensureAdmin()
		fmt.Println("Database already seeded, skipping...")
		return nil
	}

	err := DB.Transaction(func(tx *gorm.DB) error {
		if err := seedExamTypesTx(tx, examsSeedData); err != nil {
			return fmt.Errorf("failed to seed exam types: %w", err)
		}
		if err := seedQuestionsTx(tx, questionsSeedData); err != nil {
			return fmt.Errorf("failed to seed questions: %w", err)
		}
		fmt.Println("Database seeded successfully!")
		return nil
	})
	if err != nil {
		return err
	}

	ensureAdmin()
	return nil
}

func ensureAdmin() {
	var count int64
	DB.Model(&model.User{}).Count(&count)
	if count == 0 {
		hashedPassword, err := util.HashPassword("admin123")
		if err != nil {
			log.Printf("warning: failed to hash admin password: %v", err)
			return
		}
		admin := model.User{
			Username: "admin",
			Password: hashedPassword,
			Nickname: "管理员",
			Role:     "admin",
		}
		if err := DB.Create(&admin).Error; err != nil {
			log.Printf("warning: failed to create admin user: %v", err)
		} else {
			fmt.Println("Default admin user created (admin/admin123)")
		}
	}
}
