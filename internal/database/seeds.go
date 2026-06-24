package database

import (
	_ "embed"
	"crypto/rand"
	"fmt"
	"log/slog"
	"math/big"
	"os"

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
		ensureRegistrationConfig()
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
	ensureRegistrationConfig()
	return nil
}

func ensureAdmin() {
	var count int64
	DB.Model(&model.User{}).Count(&count)
	if count == 0 {
		password := os.Getenv("ADMIN_PASSWORD")
		if password == "" {
			password = "admin123"
			fmt.Println("╔══════════════════════════════════════════════════════════╗")
			fmt.Println("║  Default admin account created.                         ║")
			fmt.Println("║  Username: admin  /  Password: admin123                 ║")
			fmt.Println("║  Please change the password after first login!          ║")
			fmt.Println("║  Set ADMIN_PASSWORD env var to use a custom password.   ║")
			fmt.Println("╚══════════════════════════════════════════════════════════╝")
		}

		hashedPassword, err := util.HashPassword(password)
		if err != nil {
			slog.Warn("failed to hash admin password", "error", err)
			return
		}
		admin := model.User{
			Username: "admin",
			Password: hashedPassword,
			Nickname: "管理员",
			Role:     "admin",
		}
		if err := DB.Create(&admin).Error; err != nil {
			slog.Warn("failed to create admin user", "error", err)
		} else {
			slog.Warn("default admin user created, please change the password after first login")
		}
	}
}

// generateRandomPassword creates a cryptographically random password of the given length
// using uppercase, lowercase, digits, and a safe special character set.
func generateRandomPassword(length int) (string, error) {
	const charset = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789!@#$%&*"
	result := make([]byte, length)
	for i := range result {
		n, err := rand.Int(rand.Reader, big.NewInt(int64(len(charset))))
		if err != nil {
			return "", fmt.Errorf("failed to generate random bytes: %w", err)
		}
		result[i] = charset[n.Int64()]
	}
	return string(result), nil
}

func ensureRegistrationConfig() {
	var count int64
	DB.Model(&model.SystemConfig{}).Where("key = ?", "registration_enabled").Count(&count)
	if count == 0 {
		DB.Create(&model.SystemConfig{Key: "registration_enabled", Value: "false"})
		fmt.Println("Registration disabled by default (set via admin settings).")
	}
}
