package service

import (
	"fmt"
	"sync"
	"time"

	"exam-quiz/internal/model"
	"exam-quiz/internal/repository"
)

// GetExamTypes 获取所有考试类型（不含模块列表，前端通过 GetExamModules 获取）
func GetExamTypes() ([]model.ExamType, error) {
	return repository.ListExamTypes()
}

// GetExamType 获取单个考试类型
func GetExamType(id uint) (*model.ExamType, error) {
	return repository.GetExamType(id)
}

// GetModulesByExamID 获取某考试类型下的模块列表（含题目数，单次查询）
func GetModulesByExamID(examTypeID uint, userID uint) ([]model.ModuleWithStats, error) {
	return repository.GetModulesByExamIDWithStats(examTypeID, userID)
}

// CreateExamType 创建考试类型
func CreateExamType(exam *model.ExamType) error {
	defer InvalidateAllStatsCache()
	return repository.CreateExamType(exam)
}

// UpdateExamType 更新考试类型
func UpdateExamType(exam *model.ExamType) error {
	defer InvalidateAllStatsCache()
	return repository.UpdateExamType(exam)
}

// DeleteExamType 删除考试类型
func DeleteExamType(id uint) error {
	defer InvalidateAllStatsCache()
	return repository.DeleteExamType(id)
}

// CreateModule 创建模块
func CreateModule(module *model.Module) error {
	defer InvalidateAllStatsCache()
	return repository.CreateModule(module)
}

// UpdateModule 更新模块
func UpdateModule(module *model.Module) error {
	defer InvalidateAllStatsCache()
	return repository.UpdateModule(module)
}

// DeleteModule 删除模块
func DeleteModule(id uint) error {
	defer InvalidateAllStatsCache()
	return repository.DeleteModule(id)
}

// FullImportData 完整导入数据结构
type FullImportData struct {
	ExamTypes []FullImportExamType `json:"exam_types"`
}

// FullImportExamType 导入用的考试类型
type FullImportExamType struct {
	model.ExamType
	Modules []FullImportModule `json:"modules,omitempty"`
}

// FullImportModule 导入用的模块
type FullImportModule struct {
	model.Module
	Questions []FullImportQuestion `json:"questions,omitempty"`
}

// FullImportQuestion 导入用的题目
type FullImportQuestion struct {
	ModuleName string `json:"module_name"` // 通过模块名匹配
	model.Question
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
	// 一次查出所有模块，按 exam_type_id 分组
	allModules, err := repository.ListAllModules()
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

	questions, err := repository.ListAllQuestions()
	if err != nil {
		return nil, err
	}

	return map[string]interface{}{
		"exam_types": exams,
		"questions":  questions,
	}, nil
}

// ImportFullData 导入完整数据（考试类型 + 模块 + 题目）
func ImportFullData(data FullImportData) (*FullImportResult, error) {
	result := &FullImportResult{}

	for _, exam := range data.ExamTypes {
		// 创建考试类型
		newExam := model.ExamType{
			Name:   exam.Name,
			Remark: exam.Remark,
		}
		if err := repository.CreateExamType(&newExam); err != nil {
			// 尝试查找已存在的
			existing, lookupErr := repository.GetExamTypeByName(exam.Name)
			if lookupErr == nil {
				newExam = *existing
			} else {
				continue
			}
		} else {
			result.ExamTypesCreated++
		}

		// 创建模块和题目
		for _, mod := range exam.Modules {
			newMod := model.Module{
				Name:       mod.Name,
				ExamTypeID: newExam.ID,
				Sort:       mod.Sort,
			}
			if err := repository.CreateModule(&newMod); err != nil {
				// 尝试查找已存在的
				existing, lookupErr := repository.GetModuleByNameAndExamID(mod.Name, newExam.ID)
				if lookupErr == nil {
					newMod = *existing
				} else {
					continue
				}
			} else {
				result.ModulesCreated++
			}

			// 创建题目（批量插入）
			var newQuestions []model.Question
			for _, q := range mod.Questions {
				newQ := q.Question
				newQ.ModuleID = newMod.ID // 更新为新模块的 ID
				newQuestions = append(newQuestions, newQ)
			}
			if len(newQuestions) > 0 {
				if err := repository.BatchCreateQuestions(newQuestions); err == nil {
					result.QuestionsCreated += len(newQuestions)
				}
			}
		}
	}

	return result, nil
}

// ====== 缓存层 ======

var (
	statsCache sync.Map
)

type cacheEntry struct {
	data      interface{}
	expiresAt time.Time
}

const cacheTTL = 30 * time.Second

func getFromCache(key string) (interface{}, bool) {
	if val, ok := statsCache.Load(key); ok {
		entry := val.(cacheEntry)
		if time.Now().Before(entry.expiresAt) {
			return entry.data, true
		}
		statsCache.Delete(key)
	}
	return nil, false
}

func setCache(key string, data interface{}) {
	statsCache.Store(key, cacheEntry{
		data:      data,
		expiresAt: time.Now().Add(cacheTTL),
	})
}

// InvalidateStatsCache 清除指定用户的统计缓存（答题提交后调用）
func InvalidateStatsCache(moduleID uint, userID uint) {
	statsCache.Delete(fmt.Sprintf("overall_stats:%d", userID))
	statsCache.Delete(fmt.Sprintf("module_stats:%d:%d", moduleID, userID))
}

// InvalidateAllStatsCache 清除全部统计缓存（管理员修改题目/考试/模块后调用）
func InvalidateAllStatsCache() {
	statsCache.Range(func(key, _ interface{}) bool {
		statsCache.Delete(key)
		return true
	})
}

// GetModuleStats 获取模块统计（带缓存）
func GetModuleStats(moduleID uint, userID uint) (*repository.StatsResult, error) {
	cacheKey := fmt.Sprintf("module_stats:%d:%d", moduleID, userID)
	if cached, ok := getFromCache(cacheKey); ok {
		return cached.(*repository.StatsResult), nil
	}

	stats, err := repository.GetModuleStats(moduleID, userID)
	if err != nil {
		return nil, err
	}

	setCache(cacheKey, stats)
	return stats, nil
}

// GetOverallStats 获取全局统计（带缓存）
func GetOverallStats(userID uint) (*repository.StatsResult, error) {
	cacheKey := fmt.Sprintf("overall_stats:%d", userID)
	if cached, ok := getFromCache(cacheKey); ok {
		return cached.(*repository.StatsResult), nil
	}

	stats, err := repository.GetOverallStats(userID)
	if err != nil {
		return nil, err
	}

	setCache(cacheKey, stats)
	return stats, nil
}

// GetRecentAnswers 获取最近的答题记录（按用户隔离）
func GetRecentAnswers(limit int, userID uint) ([]model.UserAnswer, error) {
	return repository.GetRecentAnswers(limit, userID)
}

// ClearAllRecords 清除当前用户的所有答题记录
func ClearAllRecords(userID uint) error {
	return repository.ClearAllRecords(userID)
}
