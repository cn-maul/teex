package service

import (
	"fmt"
	"time"

	"exam-quiz/internal/apperr"
	"exam-quiz/internal/database"
	"exam-quiz/internal/model"
	"exam-quiz/internal/repository"
	"exam-quiz/internal/validator"
)

// QuestionFilter wraps the repository filter for handler use.
type QuestionFilter = repository.QuestionFilter

// ListQuestions queries questions with filters and pagination.
func ListQuestions(filter QuestionFilter) ([]model.Question, int64, error) {
	return repository.ListQuestions(database.DB, filter)
}

// GetQuestion 获取单个题目
func GetQuestion(id uint) (*model.Question, error) {
	return repository.GetQuestion(database.DB, id)
}

// CreateQuestion 创建题目
func CreateQuestion(question *model.Question) error {
	return repository.CreateQuestion(database.DB, question)
}

// UpdateQuestion 更新题目
func UpdateQuestion(question *model.Question) error {
	return repository.UpdateQuestion(database.DB, question)
}

// DeleteQuestion 删除题目
func DeleteQuestion(id uint) error {
	return repository.DeleteQuestion(database.DB, id)
}

// ImportQuestionsResult holds the result of a batch import operation.
type ImportQuestionsResult struct {
	ImportedCount int `json:"imported_count"`
	InvalidCount  int `json:"invalid_count"`
}

// ImportQuestions 批量导入题目（含验证、过滤、moduleID 校验）
func ImportQuestions(questions []model.Question) (*ImportQuestionsResult, error) {
	batchLimit := GetBatchLimit()
	if len(questions) > batchLimit {
		return nil, apperr.BadRequest(fmt.Sprintf("单次导入不能超过 %d 道题目", batchLimit))
	}

	// 验证并过滤每道题
	var validQuestions []model.Question
	var invalidCount int
	for _, q := range questions {
		if q.ModuleID == 0 {
			invalidCount++
			continue
		}
		if err := validator.ValidateQuestionForImport(&q); err != nil {
			invalidCount++
			continue
		}
		validQuestions = append(validQuestions, q)
	}

	if len(validQuestions) == 0 {
		return nil, apperr.BadRequest("没有有效的题目数据")
	}

	// 校验所有引用的 moduleID 存在
	moduleIDSet := make(map[uint]bool)
	for _, q := range validQuestions {
		moduleIDSet[q.ModuleID] = true
	}
	for moduleID := range moduleIDSet {
		if _, err := repository.GetModule(database.DB, moduleID); err != nil {
			return nil, err
		}
	}

	if err := repository.BatchCreateQuestions(database.DB, validQuestions); err != nil {
		return nil, err
	}

	return &ImportQuestionsResult{
		ImportedCount: len(validQuestions),
		InvalidCount:  invalidCount,
	}, nil
}

// BatchImportQuestions 批量导入题目（简单版，无验证）
func BatchImportQuestions(questions []model.Question) (int, error) {
	if len(questions) == 0 {
		return 0, nil
	}
	err := repository.BatchCreateQuestions(database.DB, questions)
	if err != nil {
		return 0, err
	}
	return len(questions), nil
}

// BatchDeleteQuestions 批量删除题目
func BatchDeleteQuestions(ids []uint) (int, error) {
	count, err := repository.BatchDeleteQuestions(database.DB, ids)
	return count, err
}

// StartQuiz 开始刷题（返回题目列表 + 场次ID）
func StartQuiz(moduleID uint, count int, mode string, difficulty int, tags string, userID uint) ([]model.Question, uint, error) {
	// 参数校验和默认值
	if count <= 0 {
		count = 10
	}
	if count > 200 {
		count = 200
	}
	if mode == "" {
		mode = "default"
	}
	validModes := map[string]bool{"default": true, "wrong": true, "random": true, "exam": true}
	if !validModes[mode] {
		return nil, 0, apperr.BadRequest("无效的刷题模式")
	}
	if difficulty < 0 || difficulty > 5 {
		return nil, 0, apperr.BadRequest("难度范围必须在 0-5 之间")
	}

	// 校验模块是否存在
	if err := ValidateModuleExists(moduleID); err != nil {
		return nil, 0, err
	}

	filter := repository.QuizFilter{
		ModuleID:   moduleID,
		Difficulty: difficulty,
		Tags:       tags,
	}

	var questions []model.Question
	var err error

	switch mode {
	case "wrong":
		questions, err = repository.GetFilteredWrongQuestions(database.DB, filter, count, userID)
	case "random":
		questions, err = repository.GetFilteredQuestions(database.DB, filter, count)
	default:
		questions, err = repository.GetFilteredUnansweredQuestions(database.DB, filter, count, userID)
		if err != nil {
			return nil, 0, err
		}
		if len(questions) < count {
			// 收集已选题目ID，补题时排除避免重复
			excludeIDs := make([]uint, len(questions))
			for i, q := range questions {
				excludeIDs[i] = q.ID
			}
			filter.ExcludeIDs = excludeIDs
			extra, err := repository.GetFilteredQuestions(database.DB, filter, count-len(questions))
			if err != nil {
				return nil, 0, err
			}
			questions = append(questions, extra...)
		}
	}
	if err != nil {
		return nil, 0, err
	}

	// 检查是否有可用题目
	if len(questions) == 0 {
		return nil, 0, apperr.BadRequest("该模块暂无题目")
	}

	// 创建考试场次
	session := &model.ExamSession{
		ModuleID:   moduleID,
		Mode:       mode,
		TotalCount: len(questions),
		UserID:     userID,
		StartedAt:  time.Now(),
	}
	if createErr := repository.CreateSession(database.DB, session); createErr != nil {
		return nil, 0, apperr.Wrapf(500, "创建考试场次失败", createErr)
	}

	return questions, session.ID, nil
}
