package service

import (
	"fmt"
	"exam-quiz/internal/model"
	"exam-quiz/internal/repository"
)

// ListQuestions 查询题目列表
func ListQuestions(filter repository.QuestionFilter) ([]model.Question, int64, error) {
	return repository.ListQuestions(filter)
}

// GetQuestion 获取单个题目
func GetQuestion(id uint) (*model.Question, error) {
	return repository.GetQuestion(id)
}

// CreateQuestion 创建题目
func CreateQuestion(question *model.Question) error {
	return repository.CreateQuestion(question)
}

// UpdateQuestion 更新题目
func UpdateQuestion(question *model.Question) error {
	return repository.UpdateQuestion(question)
}

// DeleteQuestion 删除题目
func DeleteQuestion(id uint) error {
	return repository.DeleteQuestion(id)
}

// BatchImportQuestions 批量导入题目
func BatchImportQuestions(questions []model.Question) (int, error) {
	if len(questions) == 0 {
		return 0, nil
	}
	err := repository.BatchCreateQuestions(questions)
	if err != nil {
		return 0, err
	}
	return len(questions), nil
}

// StartQuiz 开始刷题（返回题目列表 + 场次ID）
func StartQuiz(moduleID uint, count int, mode string, difficulty int, tags string) ([]model.Question, uint, error) {
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
		questions, err = repository.GetFilteredWrongQuestions(filter, count)
	case "random":
		questions, err = repository.GetFilteredQuestions(filter, count)
	default:
		questions, err = repository.GetFilteredUnansweredQuestions(filter, count)
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
		return nil, 0, fmt.Errorf("该模块暂无题目")
	}

	// 创建考试场次
	session := &model.ExamSession{
		ModuleID:   moduleID,
		Mode:       mode,
		TotalCount: len(questions),
	}
	if createErr := repository.CreateSession(session); createErr != nil {
		return nil, 0, fmt.Errorf("创建考试场次失败: %w", createErr)
	}

	return questions, session.ID, nil
}
