package repository

import (
	"exam-quiz/internal/database"
	"exam-quiz/internal/model"

	"gorm.io/gorm"
)

// ListExamTypes 获取所有考试类型
func ListExamTypes() ([]model.ExamType, error) {
	var exams []model.ExamType
	err := database.DB.Find(&exams).Error
	return exams, err
}

// GetExamType 获取单个考试类型
func GetExamType(id uint) (*model.ExamType, error) {
	var exam model.ExamType
	err := database.DB.First(&exam, id).Error
	return &exam, err
}

// GetExamTypeByName 按名称获取考试类型
func GetExamTypeByName(name string) (*model.ExamType, error) {
	var exam model.ExamType
	err := database.DB.Where("name = ?", name).First(&exam).Error
	if err != nil {
		return nil, err
	}
	return &exam, nil
}

// GetModuleByNameAndExamID 按名称和考试类型获取模块
func GetModuleByNameAndExamID(name string, examTypeID uint) (*model.Module, error) {
	var module model.Module
	err := database.DB.Where("name = ? AND exam_type_id = ?", name, examTypeID).First(&module).Error
	if err != nil {
		return nil, err
	}
	return &module, nil
}

// GetModulesByExamID 获取某考试类型下的所有模块
func GetModulesByExamID(examTypeID uint) ([]model.Module, error) {
	var modules []model.Module
	err := database.DB.Where("exam_type_id = ?", examTypeID).
		Order("sort ASC").
		Find(&modules).Error
	return modules, err
}

// GetModule 获取单个模块
func GetModule(id uint) (*model.Module, error) {
	var module model.Module
	err := database.DB.First(&module, id).Error
	return &module, err
}

// CreateExamType 创建考试类型
func CreateExamType(exam *model.ExamType) error {
	return database.DB.Create(exam).Error
}

// CreateModule 创建模块
func CreateModule(module *model.Module) error {
	return database.DB.Create(module).Error
}

// UpdateExamType 更新考试类型
func UpdateExamType(exam *model.ExamType) error {
	return database.DB.Model(&model.ExamType{}).Where("id = ?", exam.ID).Updates(map[string]interface{}{
		"name":   exam.Name,
		"remark": exam.Remark,
	}).Error
}

// DeleteExamType 删除考试类型（级联删除关联数据）
func DeleteExamType(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
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
			// 删除 Favorite
			if err := tx.Where("question_id IN (SELECT id FROM questions WHERE module_id IN ?)", moduleIDs).Delete(&model.Favorite{}).Error; err != nil {
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
func UpdateModule(module *model.Module) error {
	return database.DB.Model(&model.Module{}).Where("id = ?", module.ID).Updates(map[string]interface{}{
		"name":         module.Name,
		"exam_type_id": module.ExamTypeID,
		"sort":         module.Sort,
	}).Error
}

// DeleteModule 删除模块（级联删除关联数据）
func DeleteModule(id uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		// 删除 ExamSession
		if err := tx.Where("module_id = ?", id).Delete(&model.ExamSession{}).Error; err != nil {
			return err
		}
		// 删除 Favorite
		if err := tx.Where("question_id IN (SELECT id FROM questions WHERE module_id = ?)", id).Delete(&model.Favorite{}).Error; err != nil {
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
