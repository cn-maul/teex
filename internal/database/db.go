package database

import (
	"encoding/json"
	"fmt"
	"os"

	"exam-quiz/internal/model"

	"github.com/glebarez/sqlite"
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

	// 连接池配置：SQLite WAL 支持并发读 + 单写。
	// 纯 Go 驱动 (glebarez/sqlite) 并发能力有限，保守设为 3。
	sqlDB.SetMaxOpenConns(3)
	sqlDB.SetMaxIdleConns(3)
	sqlDB.SetConnMaxLifetime(0)

	// 自动建表
	err = DB.AutoMigrate(
		&model.User{},
		&model.ExamType{},
		&model.Module{},
		&model.Question{},
		&model.UserAnswer{},
		&model.ExamSession{},
		&model.SystemConfig{},
	)
	if err != nil {
		return fmt.Errorf("failed to migrate database: %w", err)
	}

	return nil
}

type ExamSeed struct {
	Name    string `json:"name"`
	Remark  string `json:"remark"`
	Modules []struct {
		Name string `json:"name"`
		Sort int    `json:"sort"`
	} `json:"modules"`
}

func seedExamTypesTx(tx *gorm.DB, data []byte) error {
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

func seedQuestionsTx(tx *gorm.DB, data []byte) error {
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

