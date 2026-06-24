package service

import (
	"exam-quiz/internal/apperr"
	"exam-quiz/internal/cache"
	"exam-quiz/internal/model"
	"exam-quiz/internal/repository"
	"time"
)

// QuestionFilter wraps the repository filter for handler use.
type QuestionFilter = repository.QuestionFilter

// ListQuestions queries questions with filters and pagination.
func ListQuestions(filter QuestionFilter) ([]model.Question, int64, error) {
	return repository.ListQuestions(filter)
}

// GetQuestion 获取单个题目
func GetQuestion(id uint) (*model.Question, error) {
	return repository.GetQuestion(id)
}

// CreateQuestion 创建题目
func CreateQuestion(question *model.Question) error {
	defer cache.InvalidateAll()
	return repository.CreateQuestion(question)
}

// UpdateQuestion 更新题目
func UpdateQuestion(question *model.Question) error {
	defer cache.InvalidateAll()
	return repository.UpdateQuestion(question)
}

// DeleteQuestion 删除题目
func DeleteQuestion(id uint) error {
	defer cache.InvalidateAll()
	return repository.DeleteQuestion(id)
}

// BatchImportQuestions 批量导入题目
func BatchImportQuestions(questions []model.Question) (int, error) {
	defer cache.InvalidateAll()
	if len(questions) == 0 {
		return 0, nil
	}
	err := repository.BatchCreateQuestions(questions)
	if err != nil {
		return 0, err
	}
	return len(questions), nil
}

// BatchDeleteQuestions 批量删除题目
func BatchDeleteQuestions(ids []uint) (int, error) {
	defer cache.InvalidateAll()
	count, err := repository.BatchDeleteQuestions(ids)
	return count, err
}

// StartQuiz 开始刷题（返回题目列表 + 场次ID）
func StartQuiz(moduleID uint, count int, mode string, difficulty int, tags string, userID uint) ([]model.Question, uint, error) {
	if count <= 0 {
		count = 10
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
		questions, err = repository.GetFilteredWrongQuestions(filter, count, userID)
	case "random":
		questions, err = repository.GetFilteredQuestions(filter, count)
	default:
		questions, err = repository.GetFilteredUnansweredQuestions(filter, count, userID)
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
			extra, err := repository.GetFilteredQuestions(filter, count-len(questions))
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
	if createErr := repository.CreateSession(session); createErr != nil {
		return nil, 0, apperr.Wrapf(500, "创建考试场次失败", createErr)
	}

	return questions, session.ID, nil
}
