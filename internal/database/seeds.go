package database

import (
	_ "embed"
	"fmt"

	"exam-quiz/internal/model"

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
		fmt.Println("Database already seeded, skipping...")
		return nil
	}

	return DB.Transaction(func(tx *gorm.DB) error {
		if err := seedExamTypesTx(tx, examsSeedData); err != nil {
			return fmt.Errorf("failed to seed exam types: %w", err)
		}
		if err := seedQuestionsTx(tx, questionsSeedData); err != nil {
			return fmt.Errorf("failed to seed questions: %w", err)
		}
		fmt.Println("Database seeded successfully!")
		return nil
	})
}
