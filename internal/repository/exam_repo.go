package repository

import (
	"errors"

	"exam-quiz/internal/apperr"
	"exam-quiz/internal/model"

	"gorm.io/gorm"
)

// ListExamTypes 获取所有考试类型
func ListExamTypes(db *gorm.DB) ([]model.ExamType, error) {
	var exams []model.ExamType
	err := db.Find(&exams).Error
	return exams, err
}

// GetExamType 获取单个考试类型
func GetExamType(db *gorm.DB, id uint) (*model.ExamType, error) {
	var exam model.ExamType
	err := db.First(&exam, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("考试类型不存在")
		}
		return nil, err
	}
	return &exam, nil
}

// GetExamTypeByName 按名称获取考试类型
func GetExamTypeByName(db *gorm.DB, name string) (*model.ExamType, error) {
	var exam model.ExamType
	err := db.Where("name = ?", name).First(&exam).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("考试类型不存在")
		}
		return nil, err
	}
	return &exam, nil
}

// GetModuleByNameAndExamID 按名称和考试类型获取模块
func GetModuleByNameAndExamID(db *gorm.DB, name string, examTypeID uint) (*model.Module, error) {
	var module model.Module
	err := db.Where("name = ? AND exam_type_id = ?", name, examTypeID).First(&module).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("模块不存在")
		}
		return nil, err
	}
	return &module, nil
}

// ListAllModules 获取所有模块
func ListAllModules(db *gorm.DB) ([]model.Module, error) {
	var modules []model.Module
	err := db.Order("exam_type_id ASC, sort ASC").Find(&modules).Error
	return modules, err
}

// GetModule 获取单个模块
func GetModule(db *gorm.DB, id uint) (*model.Module, error) {
	var module model.Module
	err := db.First(&module, id).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, apperr.NotFound("模块不存在")
		}
		return nil, err
	}
	return &module, nil
}

// CreateExamType 创建考试类型
func CreateExamType(db *gorm.DB, exam *model.ExamType) error {
	return db.Create(exam).Error
}

// CreateModule 创建模块
func CreateModule(db *gorm.DB, module *model.Module) error {
	return db.Create(module).Error
}

// UpdateExamType 更新考试类型
func UpdateExamType(db *gorm.DB, exam *model.ExamType) error {
	return db.Model(&model.ExamType{}).Where("id = ?", exam.ID).Updates(map[string]interface{}{
		"name":   exam.Name,
		"remark": exam.Remark,
	}).Error
}

// DeleteExamType 删除考试类型（级联删除关联数据）
func DeleteExamType(db *gorm.DB, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 获取该考试类型下所有模块 ID
		var moduleIDs []uint
		if err := tx.Model(&model.Module{}).Where("exam_type_id = ?", id).Pluck("id", &moduleIDs).Error; err != nil {
			return err
		}

		if len(moduleIDs) > 0 {
			// 删除 ExamSession
			if err := tx.Where("module_id IN ?", moduleIDs).Delete(&model.ExamSession{}).Error; err != nil {
				return err
			}
			// 删除 UserAnswer
			if err := tx.Where("question_id IN (SELECT id FROM questions WHERE module_id IN ?)", moduleIDs).Delete(&model.UserAnswer{}).Error; err != nil {
				return err
			}
			// 删除 Question
			if err := tx.Where("module_id IN ?", moduleIDs).Delete(&model.Question{}).Error; err != nil {
				return err
			}
			// 删除 Module
			if err := tx.Where("exam_type_id = ?", id).Delete(&model.Module{}).Error; err != nil {
				return err
			}
		}

		return tx.Delete(&model.ExamType{}, id).Error
	})
}

// UpdateModule 更新模块
func UpdateModule(db *gorm.DB, module *model.Module) error {
	return db.Model(&model.Module{}).Where("id = ?", module.ID).Updates(map[string]interface{}{
		"name":         module.Name,
		"exam_type_id": module.ExamTypeID,
		"sort":         module.Sort,
	}).Error
}

// DeleteModule 删除模块（级联删除关联数据）
func DeleteModule(db *gorm.DB, id uint) error {
	return db.Transaction(func(tx *gorm.DB) error {
		// 删除 ExamSession
		if err := tx.Where("module_id = ?", id).Delete(&model.ExamSession{}).Error; err != nil {
			return err
		}
		// 删除 UserAnswer
		if err := tx.Where("question_id IN (SELECT id FROM questions WHERE module_id = ?)", id).Delete(&model.UserAnswer{}).Error; err != nil {
			return err
		}
		// 删除 Question
		if err := tx.Where("module_id = ?", id).Delete(&model.Question{}).Error; err != nil {
			return err
		}
		// 删除 Module
		return tx.Delete(&model.Module{}, id).Error
	})
}

// GetModulesByExamIDWithStats 获取某考试类型下的模块列表（含题目数和未做题数，单次查询）
func GetModulesByExamIDWithStats(db *gorm.DB, examTypeID uint, userID uint) ([]model.ModuleWithStats, error) {
	var results []model.ModuleWithStats
	err := db.Raw(`
		SELECT m.*,
			COALESCE((SELECT COUNT(*) FROM questions q WHERE q.module_id = m.id), 0) AS question_count,
			COALESCE((SELECT COUNT(*) FROM questions q WHERE q.module_id = m.id
				AND NOT EXISTS (SELECT 1 FROM user_answers WHERE user_answers.question_id = q.id AND user_answers.user_id = ?)), 0) AS unanswered
		FROM modules m
		WHERE m.exam_type_id = ?
		ORDER BY m.sort ASC
	`, userID, examTypeID).Scan(&results).Error
	return results, err
}

// ExamModuleStatsRow is a single row from the aggregated stats query.
type ExamModuleStatsRow struct {
	ID             uint
	Name           string
	TotalQuestions int64
	TotalAnswered  int64
	CorrectCount   int64
}

// GetExamStatsAggregated returns per-module stats for an exam type in one query.
func GetExamStatsAggregated(db *gorm.DB, examTypeID uint, userID uint) ([]ExamModuleStatsRow, error) {
	var rows []ExamModuleStatsRow
	err := db.Raw(`
		SELECT
			m.id,
			m.name,
			COALESCE((SELECT COUNT(*) FROM questions q WHERE q.module_id = m.id), 0) AS total_questions,
			COALESCE(stats.total_answered, 0) AS total_answered,
			COALESCE(stats.correct_count, 0) AS correct_count
		FROM modules m
		LEFT JOIN (
			SELECT
				q.module_id,
				COUNT(*) AS total_answered,
				SUM(CASE WHEN ua.is_correct THEN 1 ELSE 0 END) AS correct_count
			FROM user_answers ua
			INNER JOIN questions q ON q.id = ua.question_id
			WHERE ua.user_id = ?
			  AND ua.id IN (
			      SELECT MAX(id) FROM user_answers WHERE user_id = ? GROUP BY question_id
			  )
			GROUP BY q.module_id
		) stats ON stats.module_id = m.id
		WHERE m.exam_type_id = ?
		ORDER BY m.sort ASC
	`, userID, userID, examTypeID).Scan(&rows).Error
	return rows, err
}
