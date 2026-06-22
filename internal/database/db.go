package database

import (
	"encoding/json"
	"fmt"
	"os"
	"path/filepath"

	"exam-quiz/internal/model"

	"gorm.io/driver/sqlite"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
)

var DB *gorm.DB

// Init 初始化 SQLite 数据库
func Init(dbPath string) error {
	var err error

	// GORM 日志级别由环境变量 GORM_LOG 控制
	logLevel := logger.Silent
	if os.Getenv("GORM_LOG") == "true" {
		logLevel = logger.Info
	}

	DB, err = gorm.Open(sqlite.Open(dbPath), &gorm.Config{
		Logger: logger.Default.LogMode(logLevel),
	})
	if err != nil {
		return fmt.Errorf("failed to connect database: %w", err)
	}

	// 启用 WAL 模式，提升并发性能
	sqlDB, err := DB.DB()
	if err != nil {
		return fmt.Errorf("failed to get underlying db: %w", err)
	}
	_, err = sqlDB.Exec("PRAGMA journal_mode=WAL")
	if err != nil {
		return fmt.Errorf("failed to set WAL mode: %w", err)
	}

	// 配置连接池（SQLite 单写限制）
	sqlDB.SetMaxOpenConns(1)
	sqlDB.SetMaxIdleConns(1)
	sqlDB.SetConnMaxLifetime(0)

	// 自动建表
	err = DB.AutoMigrate(
		&model.ExamType{},
		&model.Module{},
		&model.Question{},
		&model.UserAnswer{},
		&model.ExamSession{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

// Seed 加载种子数据
func Seed(seedDir string) error {
	// 检查是否已有数据
	var count int64
	DB.Model(&model.ExamType{}).Count(&count)
	if count > 0 {
		fmt.Println("Database already seeded, skipping...")
		return nil
	}

	// 在事务中加载种子数据，确保原子性
	return DB.Transaction(func(tx *gorm.DB) error {
		// 加载考试类型和模块
		if err := seedExamTypesTx(tx, filepath.Join(seedDir, "exams.json")); err != nil {
			return fmt.Errorf("failed to seed exam types: %w", err)
		}

		// 加载示例题目
		if err := seedQuestionsTx(tx, filepath.Join(seedDir, "questions_sample.json")); err != nil {
			return fmt.Errorf("failed to seed questions: %w", err)
		}

		fmt.Println("Database seeded successfully!")
		return nil
	})
}

type ExamSeed struct {
	Name    string `json:"name"`
	Remark  string `json:"remark"`
	Modules []struct {
		Name string `json:"name"`
		Sort int    `json:"sort"`
	} `json:"modules"`
}

func seedExamTypesTx(tx *gorm.DB, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var exams []ExamSeed
	if err := json.Unmarshal(data, &exams); err != nil {
		return err
	}

	for _, exam := range exams {
		examType := model.ExamType{
			Name:   exam.Name,
			Remark: exam.Remark,
		}
		if err := tx.Create(&examType).Error; err != nil {
			return err
		}

		for _, mod := range exam.Modules {
			module := model.Module{
				Name:       mod.Name,
				ExamTypeID: examType.ID,
				Sort:       mod.Sort,
			}
			if err := tx.Create(&module).Error; err != nil {
				return err
			}
		}
	}

	return nil
}

type QuestionSeed struct {
	ModuleID   uint   `json:"module_id"`
	Type       string `json:"type"`
	Content    string `json:"content"`
	Options    string `json:"options"`
	Answer     string `json:"answer"`
	Analysis   string `json:"analysis"`
	Difficulty int    `json:"difficulty"`
	Tags       string `json:"tags"`
	Source     string `json:"source"`
}

func seedQuestionsTx(tx *gorm.DB, filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return err
	}

	var questions []QuestionSeed
	if err := json.Unmarshal(data, &questions); err != nil {
		return err
	}

	for _, q := range questions {
		question := model.Question{
			ModuleID:   q.ModuleID,
			Type:       q.Type,
			Content:    q.Content,
			Options:    q.Options,
			Answer:     q.Answer,
			Analysis:   q.Analysis,
			Difficulty: q.Difficulty,
			Tags:       q.Tags,
			Source:     q.Source,
		}
		if err := tx.Create(&question).Error; err != nil {
			return err
		}
	}

	return nil
}

// GetDB 获取数据库实例
func GetDB() *gorm.DB {
	return DB
}
