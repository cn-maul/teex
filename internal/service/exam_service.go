package service

import (
	"fmt"

	"exam-quiz/internal/apperr"
	"exam-quiz/internal/database"
	"exam-quiz/internal/model"
	"exam-quiz/internal/repository"

	"gorm.io/gorm"
)

// GetExamTypes returns all exam types (without modules; the frontend fetches modules separately).
func GetExamTypes() ([]model.ExamType, error) {
	return repository.ListExamTypes(database.DB, )
}

// GetExamType returns a single exam type by ID.
func GetExamType(id uint) (*model.ExamType, error) {
	return repository.GetExamType(database.DB, id)
}

// GetModulesByExamID returns the module list for an exam type (with stats, single query).
func GetModulesByExamID(examTypeID uint, userID uint) ([]model.ModuleWithStats, error) {
	return repository.GetModulesByExamIDWithStats(database.DB, examTypeID, userID)
}

// CreateExamType creates a new exam type.
func CreateExamType(exam *model.ExamType) error {
	return repository.CreateExamType(database.DB, exam)
}

// UpdateExamType updates an existing exam type.
func UpdateExamType(exam *model.ExamType) error {
	return repository.UpdateExamType(database.DB, exam)
}

// DeleteExamType deletes an exam type and all related data.
func DeleteExamType(id uint) error {
	return repository.DeleteExamType(database.DB, id)
}

// CreateModule creates a new module.
func CreateModule(module *model.Module) error {
	return repository.CreateModule(database.DB, module)
}

// UpdateModule updates an existing module.
func UpdateModule(module *model.Module) error {
	return repository.UpdateModule(database.DB, module)
}

// DeleteModule deletes a module and all related data.
func DeleteModule(id uint) error {
	return repository.DeleteModule(database.DB, id)
}

// ValidateExamTypeExists checks that an exam type exists.
func ValidateExamTypeExists(id uint) error {
	_, err := repository.GetExamType(database.DB, id)
	return err
}

// ValidateModuleExists checks that a module exists.
func ValidateModuleExists(id uint) error {
	_, err := repository.GetModule(database.DB, id)
	return err
}

// CheckExamTypeNameUnique returns an error if the exam type name already exists.
func CheckExamTypeNameUnique(name string) error {
	existing, err := repository.GetExamTypeByName(database.DB, name)
	if err != nil {
		if apperr.IsNotFound(err) {
			return nil // not found = unique
		}
		return err
	}
	if existing != nil {
		return apperr.Conflict("该考试类型名称已存在")
	}
	return nil
}

// CheckModuleNameUnique returns an error if the module name already exists under the given exam type.
func CheckModuleNameUnique(name string, examTypeID uint) error {
	existing, err := repository.GetModuleByNameAndExamID(database.DB, name, examTypeID)
	if err != nil {
		if apperr.IsNotFound(err) {
			return nil // not found = unique
		}
		return err
	}
	if existing != nil {
		return apperr.Conflict("该考试类型下已存在同名模块")
	}
	return nil
}

// CheckExamTypeNameUniqueForUpdate checks name uniqueness excluding the current record.
func CheckExamTypeNameUniqueForUpdate(name string, excludeID uint) error {
	existing, err := repository.GetExamTypeByName(database.DB, name)
	if err != nil {
		if apperr.IsNotFound(err) {
			return nil
		}
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return apperr.Conflict("考试类型名称已存在")
	}
	return nil
}

// CheckModuleNameUniqueForUpdate checks module name uniqueness under an exam type, excluding the current record.
func CheckModuleNameUniqueForUpdate(name string, examTypeID uint, excludeID uint) error {
	existing, err := repository.GetModuleByNameAndExamID(database.DB, name, examTypeID)
	if err != nil {
		if apperr.IsNotFound(err) {
			return nil
		}
		return err
	}
	if existing != nil && existing.ID != excludeID {
		return apperr.Conflict("该考试类型下已存在同名模块")
	}
	return nil
}

// UpdateModuleWithValidation updates a module with field-default and name-uniqueness validation.
func UpdateModuleWithValidation(id uint, module *model.Module) error {
	// 读取现有模块，填充未提供的字段
	existing, err := repository.GetModule(database.DB, id)
	if err != nil {
		return err
	}
	if module.ExamTypeID == 0 {
		module.ExamTypeID = existing.ExamTypeID
	}
	// 校验 exam_type_id 存在
	if err := ValidateExamTypeExists(module.ExamTypeID); err != nil {
		return err
	}
	// 名称唯一性检查
	if module.Name != "" {
		if err := CheckModuleNameUniqueForUpdate(module.Name, module.ExamTypeID, id); err != nil {
			return err
		}
	}
	module.ID = id
	return repository.UpdateModule(database.DB, module)
}

// DeleteResult holds cascade-deletion impact counts.
type DeleteResult struct {
	AffectedModules   int64 `json:"affected_modules,omitempty"`
	AffectedQuestions int64 `json:"affected_questions"`
	AffectedAnswers   int64 `json:"affected_answers"`
}

// DeleteExamTypeWithStats counts affected records then deletes an exam type.
func DeleteExamTypeWithStats(id uint) (*DeleteResult, error) {
	modules, questions, answers, _ := repository.CountAffectedByExamType(database.DB, id)
	if err := repository.DeleteExamType(database.DB, id); err != nil {
		return nil, err
	}
	return &DeleteResult{
		AffectedModules:   modules,
		AffectedQuestions: questions,
		AffectedAnswers:   answers,
	}, nil
}

// DeleteModuleWithStats counts affected records then deletes a module.
func DeleteModuleWithStats(id uint) (*DeleteResult, error) {
	questions, answers, _ := repository.CountAffectedByModule(database.DB, id)
	if err := repository.DeleteModule(database.DB, id); err != nil {
		return nil, err
	}
	return &DeleteResult{
		AffectedQuestions: questions,
		AffectedAnswers:   answers,
	}, nil
}

// GetModule returns a single module by ID.
func GetModule(id uint) (*model.Module, error) {
	return repository.GetModule(database.DB, id)
}

// GetExamTypeByName returns an exam type by name.
func GetExamTypeByName(name string) (*model.ExamType, error) {
	return repository.GetExamTypeByName(database.DB, name)
}

// GetModuleByNameAndExamID returns a module by name and exam type ID.
func GetModuleByNameAndExamID(name string, examTypeID uint) (*model.Module, error) {
	return repository.GetModuleByNameAndExamID(database.DB, name, examTypeID)
}

// CountAffectedByExamType counts the cascade-deletion impact for an exam type.
func CountAffectedByExamType(examTypeID uint) (modules int64, questions int64, answers int64, err error) {
	return repository.CountAffectedByExamType(database.DB, examTypeID)
}

// CountAffectedByModule counts the cascade-deletion impact for a module.
func CountAffectedByModule(moduleID uint) (questions int64, answers int64, err error) {
	return repository.CountAffectedByModule(database.DB, moduleID)
}

// FullImportData is the top-level structure for full data import.
type FullImportData struct {
	ExamTypes []FullImportExamType `json:"exam_types"`
}

// FullImportExamType is the exam type structure used for import.
type FullImportExamType struct {
	model.ExamType
	Modules []FullImportModule `json:"modules,omitzero"`
}

// FullImportModule is the module structure used for import.
type FullImportModule struct {
	model.Module
	Questions []FullImportQuestion `json:"questions,omitzero"`
}

// FullImportQuestion is the question structure used for import (matched by module name).
type FullImportQuestion struct {
	ModuleName string `json:"module_name"`
	model.Question
}

// FullImportResult holds the counts of items created during import.
type FullImportResult struct {
	ExamTypesCreated int `json:"exam_types_created"`
	ModulesCreated   int `json:"modules_created"`
	QuestionsCreated int `json:"questions_created"`
}

// ExportAllData exports all exam types, modules, and questions.
func ExportAllData() (map[string]interface{}, error) {
	exams, err := repository.ListExamTypes(database.DB, )
	if err != nil {
		return nil, err
	}
	// Fetch all modules in a single query, grouped by exam_type_id
	allModules, err := repository.ListAllModules(database.DB, )
	if err != nil {
		return nil, err
	}
	moduleMap := make(map[uint][]model.Module)
	for _, m := range allModules {
		moduleMap[m.ExamTypeID] = append(moduleMap[m.ExamTypeID], m)
	}
	for i := range exams {
		exams[i].Modules = moduleMap[exams[i].ID]
	}

	questions, err := repository.ListAllQuestions(database.DB, )
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"exam_types": exams,
		"questions":  questions,
	}, nil
}

// ImportFullData imports complete data (exam types + modules + questions).
// Validates all data first, then creates; any creation failure returns immediately.
func ImportFullData(data FullImportData) (*FullImportResult, error) {
	result := &FullImportResult{}

	// Step 1: Validate and map module names to their questions
	type moduleImport struct {
		Name      string
		Sort      int
		Questions []model.Question
	}
	type examImport struct {
		Name    string
		Remark  string
		Modules []moduleImport
	}
	var imports []examImport
	for _, exam := range data.ExamTypes {
		if exam.Name == "" {
			return nil, fmt.Errorf("exam type name cannot be empty")
		}
		ei := examImport{Name: exam.Name, Remark: exam.Remark}
		for _, mod := range exam.Modules {
			mi := moduleImport{Name: mod.Name, Sort: mod.Sort}
			for _, q := range mod.Questions {
				newQ := q.Question
				newQ.ModuleID = 0 // will be set after module creation
				mi.Questions = append(mi.Questions, newQ)
			}
			ei.Modules = append(ei.Modules, mi)
		}
		imports = append(imports, ei)
	}

	// Step 2: Create everything inside a transaction, fail fast on any error
	err := database.DB.Transaction(func(tx *gorm.DB) error {
		for _, ei := range imports {
			newExam := model.ExamType{
				Name:   ei.Name,
				Remark: ei.Remark,
			}
			if err := tx.Create(&newExam).Error; err != nil {
				// Exam type with this name may already exist — look it up and reuse
				existing := &model.ExamType{}
				if lookupErr := tx.Where("name = ?", ei.Name).First(existing).Error; lookupErr != nil {
					return fmt.Errorf("failed to create exam type %q: %w", ei.Name, err)
				}
				newExam = *existing
			} else {
				result.ExamTypesCreated++
			}

			for _, mi := range ei.Modules {
				newMod := model.Module{
					Name:       mi.Name,
					ExamTypeID: newExam.ID,
					Sort:       mi.Sort,
				}
				if err := tx.Create(&newMod).Error; err != nil {
					existing := &model.Module{}
					if lookupErr := tx.Where("name = ? AND exam_type_id = ?", mi.Name, newExam.ID).First(existing).Error; lookupErr != nil {
						return fmt.Errorf("failed to create module %q: %w", mi.Name, err)
					}
					newMod = *existing
				} else {
					result.ModulesCreated++
				}

				if len(mi.Questions) > 0 {
					for i := range mi.Questions {
						mi.Questions[i].ModuleID = newMod.ID
					}
					if err := tx.Create(&mi.Questions).Error; err != nil {
						return fmt.Errorf("failed to create questions for module %q: %w", mi.Name, err)
					}
					result.QuestionsCreated += len(mi.Questions)
				}
			}
		}
		return nil
	})
	if err != nil {
		return result, err
	}

	return result, nil
}

// GetModuleStats returns module-level statistics.
func GetModuleStats(moduleID uint, userID uint) (*repository.StatsResult, error) {
	return repository.GetModuleStats(database.DB, moduleID, userID)
}

// GetOverallStats returns global statistics.
func GetOverallStats(userID uint) (*repository.StatsResult, error) {
	return repository.GetOverallStats(database.DB, userID)
}

// GetRecentAnswers returns recent answer records for a user.
func GetRecentAnswers(limit int, userID uint) ([]model.UserAnswer, error) {
	return repository.GetRecentAnswers(database.DB, limit, userID)
}

// ClearAllRecords clears all answer records for a user.
func ClearAllRecords(userID uint) error {
	return repository.ClearAllRecords(database.DB, userID)
}

// ExamModuleStats holds a module with its full statistics.
type ExamModuleStats struct {
	ID             uint    `json:"id"`
	Name           string  `json:"name"`
	TotalQuestions  int64   `json:"total_questions"`
	TotalAnswered  int64   `json:"total_answered"`
	CorrectCount   int64   `json:"correct_count"`
	Accuracy       float64 `json:"accuracy"`
	Unanswered     int64   `json:"unanswered"`
}

// GetExamStats returns per-module stats for an entire exam type in a single call.
func GetExamStats(examTypeID uint, userID uint) ([]ExamModuleStats, error) {
	rows, err := repository.GetExamStatsAggregated(database.DB, examTypeID, userID)
	if err != nil {
		return nil, err
	}

	results := make([]ExamModuleStats, len(rows))
	for i, row := range rows {
		var accuracy float64
		if row.TotalAnswered > 0 {
			accuracy = float64(row.CorrectCount) / float64(row.TotalAnswered) * 100
		}
		results[i] = ExamModuleStats{
			ID:             row.ID,
			Name:           row.Name,
			TotalQuestions:  row.TotalQuestions,
			TotalAnswered:  row.TotalAnswered,
			CorrectCount:   row.CorrectCount,
			Accuracy:       accuracy,
			Unanswered:     row.TotalQuestions - row.TotalAnswered,
		}
	}
	return results, nil
}
