package service

import (
	"exam-quiz/internal/model"
	"exam-quiz/internal/repository"
)

// GetExamTypes 获取所有考试类型（含模块列表）
func GetExamTypes() ([]model.ExamType, error) {
	exams, err := repository.ListExamTypes()
	if err != nil {
		return nil, err
	}

	// 为每个考试类型加载模块
	for i := range exams {
		modules, err := repository.GetModulesByExamID(exams[i].ID)
		if err != nil {
			return nil, err
		}
		exams[i].Modules = modules
	}

	return exams, nil
}

// GetExamType 获取单个考试类型
func GetExamType(id uint) (*model.ExamType, error) {
	return repository.GetExamType(id)
}

// GetModulesByExamID 获取某考试类型下的模块列表（含题目数）
func GetModulesByExamID(examTypeID uint) ([]model.ModuleWithStats, error) {
	modules, err := repository.GetModulesByExamID(examTypeID)
	if err != nil {
		return nil, err
	}

	var result []model.ModuleWithStats
	for _, mod := range modules {
		questionCount, err := repository.CountQuestionsByModule(mod.ID)
		if err != nil {
			return nil, err
		}
		unanswered, err := repository.CountUnansweredByModule(mod.ID)
		if err != nil {
			return nil, err
		}
		result = append(result, model.ModuleWithStats{
			Module:        mod,
			QuestionCount: questionCount,
			Unanswered:    unanswered,
		})
	}

	return result, nil
}

// CreateExamType 创建考试类型
func CreateExamType(exam *model.ExamType) error {
	return repository.CreateExamType(exam)
}

// UpdateExamType 更新考试类型
func UpdateExamType(exam *model.ExamType) error {
	return repository.UpdateExamType(exam)
}

// DeleteExamType 删除考试类型
func DeleteExamType(id uint) error {
	return repository.DeleteExamType(id)
}

// CreateModule 创建模块
func CreateModule(module *model.Module) error {
	return repository.CreateModule(module)
}

// UpdateModule 更新模块
func UpdateModule(module *model.Module) error {
	return repository.UpdateModule(module)
}

// DeleteModule 删除模块
func DeleteModule(id uint) error {
	return repository.DeleteModule(id)
}

// FullImportData 完整导入数据结构
type FullImportData struct {
	ExamTypes []model.ExamType `json:"exam_types"`
}

// FullImportResult 导入结果
type FullImportResult struct {
	ExamTypesCreated int `json:"exam_types_created"`
	ModulesCreated   int `json:"modules_created"`
	QuestionsCreated int `json:"questions_created"`
}

// ExportAllData 导出所有数据
func ExportAllData() (map[string]interface{}, error) {
	exams, err := repository.ListExamTypes()
	if err != nil {
		return nil, err
	}
	for i := range exams {
		modules, err := repository.GetModulesByExamID(exams[i].ID)
		if err != nil {
			return nil, err
		}
		exams[i].Modules = modules
	}

	questions, _, err := repository.ListQuestions(repository.QuestionFilter{Size: 10000})
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"exam_types": exams,
		"questions":  questions,
	}, nil
}

// ImportFullData 导入完整数据
func ImportFullData(data FullImportData) (*FullImportResult, error) {
	result := &FullImportResult{}

	for _, exam := range data.ExamTypes {
		// 创建考试类型
		newExam := model.ExamType{
			Name:   exam.Name,
			Remark: exam.Remark,
		}
		if err := repository.CreateExamType(&newExam); err != nil {
			// 跳过已存在的
			continue
		}
		result.ExamTypesCreated++

		// 创建模块
		for _, mod := range exam.Modules {
			newMod := model.Module{
				Name:       mod.Name,
				ExamTypeID: newExam.ID,
				Sort:       mod.Sort,
			}
			if err := repository.CreateModule(&newMod); err != nil {
				continue
			}
			result.ModulesCreated++
		}
	}

	return result, nil
}
