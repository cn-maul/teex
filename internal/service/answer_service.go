package service

import (
	"fmt"
	"sort"
	"strings"

	"exam-quiz/internal/model"
	"exam-quiz/internal/repository"
)

// AnswerResult 答题结果
type AnswerResult struct {
	IsCorrect     bool   `json:"is_correct"`
	CorrectAnswer string `json:"correct_answer"`
	Analysis      string `json:"analysis"`
	UserInput     string `json:"user_input"`
}

// SubmitAnswer 提交答案
func SubmitAnswer(questionID uint, userInput string, duration int) (*AnswerResult, error) {
	// 1. 获取题目
	question, err := repository.GetQuestion(questionID)
	if err != nil {
		return nil, fmt.Errorf("question not found: %w", err)
	}

	// 2. 比对答案
	isCorrect := compareAnswers(question.Answer, userInput, question.Type)

	// 3. 保存答题记录
	answer := &model.UserAnswer{
		QuestionID: questionID,
		UserInput:  userInput,
		IsCorrect:  isCorrect,
		Duration:   duration,
	}
	if err := repository.SaveAnswer(answer); err != nil {
		return nil, fmt.Errorf("failed to save answer: %w", err)
	}

	// 4. 返回结果
	return &AnswerResult{
		IsCorrect:     isCorrect,
		CorrectAnswer: question.Answer,
		Analysis:      question.Analysis,
		UserInput:     userInput,
	}, nil
}

// compareAnswers 比对用户答案和正确答案
func compareAnswers(correct, userAnswer, questionType string) bool {
	correct = strings.TrimSpace(strings.ToUpper(correct))
	userAnswer = strings.TrimSpace(strings.ToUpper(userAnswer))

	switch questionType {
	case "multi":
		// 多选题：排序后比较（忽略顺序）
		return sortedCompare(correct, userAnswer)
	case "judge":
		// 判断题：直接比较
		return correct == userAnswer
	default:
		// 单选题：直接比较
		return correct == userAnswer
	}
}

// sortedCompare 排序后比较（用于多选题）
func sortedCompare(a, b string) bool {
	aList := strings.Split(a, ",")
	bList := strings.Split(b, ",")
	sort.Strings(aList)
	sort.Strings(bList)
	return strings.Join(aList, ",") == strings.Join(bList, ",")
}

// SubmitBatchAnswers 批量提交答案（考试模式交卷用）
func SubmitBatchAnswers(answers []BatchAnswerItem) ([]AnswerResult, error) {
	var results []AnswerResult
	for _, a := range answers {
		result, err := SubmitAnswer(a.QuestionID, a.UserInput, a.Duration)
		if err != nil {
			return nil, fmt.Errorf("failed to submit answer for question %d: %w", a.QuestionID, err)
		}
		results = append(results, *result)
	}
	return results, nil
}

// BatchAnswerItem 批量提交的单条答案
type BatchAnswerItem struct {
	QuestionID uint   `json:"question_id"`
	UserInput  string `json:"user_input"`
	Duration   int    `json:"duration"`
}

// SubmitBatchAnswersWithSession 批量提交答案并结束考试场次
func SubmitBatchAnswersWithSession(sessionID uint, answers []BatchAnswerItem) ([]AnswerResult, error) {
	var results []AnswerResult
	correctCount := 0
	totalDuration := 0

	for _, a := range answers {
		result, err := SubmitAnswerWithSession(sessionID, a.QuestionID, a.UserInput, a.Duration)
		if err != nil {
			return nil, fmt.Errorf("failed to submit answer for question %d: %w", a.QuestionID, err)
		}
		results = append(results, *result)
		if result.IsCorrect {
			correctCount++
		}
		totalDuration += a.Duration
	}

	// 结束考试场次
	if sessionID > 0 {
		_ = repository.FinishSession(sessionID, correctCount, totalDuration)
	}

	return results, nil
}

// SubmitAnswerWithSession 提交答案并关联场次
func SubmitAnswerWithSession(sessionID uint, questionID uint, userInput string, duration int) (*AnswerResult, error) {
	question, err := repository.GetQuestion(questionID)
	if err != nil {
		return nil, fmt.Errorf("question not found: %w", err)
	}

	isCorrect := compareAnswers(question.Answer, userInput, question.Type)

	answer := &model.UserAnswer{
		QuestionID:    questionID,
		ExamSessionID: sessionID,
		UserInput:     userInput,
		IsCorrect:     isCorrect,
		Duration:      duration,
	}
	if err := repository.SaveAnswer(answer); err != nil {
		return nil, fmt.Errorf("failed to save answer: %w", err)
	}

	return &AnswerResult{
		IsCorrect:     isCorrect,
		CorrectAnswer: question.Answer,
		Analysis:      question.Analysis,
		UserInput:     userInput,
	}, nil
}

// GetSession 获取考试场次详情
func GetSession(id uint) (*model.ExamSession, error) {
	return repository.GetSessionByID(id)
}

// GetSessions 获取考试场次列表
func GetSessions(page, size int) ([]model.ExamSession, int64, error) {
	return repository.GetSessions(page, size)
}

// GetSessionAnswers 获取某个场次的答题记录
func GetSessionAnswers(sessionID uint) ([]model.UserAnswer, error) {
	return repository.GetSessionAnswers(sessionID)
}

// GetModuleStats 获取模块统计
func GetModuleStats(moduleID uint) (*repository.StatsResult, error) {
	return repository.GetModuleStats(moduleID)
}

// GetOverallStats 获取全局统计
func GetOverallStats() (*repository.StatsResult, error) {
	return repository.GetOverallStats()
}

// GetRecentAnswers 获取最近的答题记录
func GetRecentAnswers(limit int) ([]model.UserAnswer, error) {
	return repository.GetRecentAnswers(limit)
}

// ClearAllRecords 清除所有答题记录
func ClearAllRecords() error {
	return repository.ClearAllRecords()
}
