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
	case "fill":
		// 填空题：去除首尾空格后比较
		return strings.TrimSpace(correct) == strings.TrimSpace(userAnswer)
	default:
		// 单选题：直接比较
		return correct == userAnswer
	}
}

// sortedCompare 排序后比较（用于多选题，逐个去除元素空格）
func sortedCompare(a, b string) bool {
	aList := strings.Split(a, ",")
	bList := strings.Split(b, ",")
	for i := range aList {
		aList[i] = strings.TrimSpace(aList[i])
	}
	for i := range bList {
		bList[i] = strings.TrimSpace(bList[i])
	}
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

// SubmitBatchAnswersWithSession 批量提交答案并结束考试场次（批量查询+批量写入）
func SubmitBatchAnswersWithSession(sessionID uint, answers []BatchAnswerItem) ([]AnswerResult, error) {
	if len(answers) == 0 {
		return nil, nil
	}

	// 1. 收集所有 question_id
	ids := make([]uint, len(answers))
	for i, a := range answers {
		ids[i] = a.QuestionID
	}

	// 2. 批量查询题目
	questionMap, err := repository.BatchGetQuestions(ids)
	if err != nil {
		return nil, fmt.Errorf("failed to batch get questions: %w", err)
	}

	// 3. 在内存中比对答案并构建批量写入数据
	var results []AnswerResult
	var userAnswers []model.UserAnswer
	correctCount := 0
	totalDuration := 0

	for _, a := range answers {
		question, ok := questionMap[a.QuestionID]
		if !ok {
			return nil, fmt.Errorf("question %d not found", a.QuestionID)
		}

		isCorrect := compareAnswers(question.Answer, a.UserInput, question.Type)

		results = append(results, AnswerResult{
			IsCorrect:     isCorrect,
			CorrectAnswer: question.Answer,
			Analysis:      question.Analysis,
			UserInput:     a.UserInput,
		})

		userAnswers = append(userAnswers, model.UserAnswer{
			QuestionID:    a.QuestionID,
			ExamSessionID: sessionID,
			UserInput:     a.UserInput,
			IsCorrect:     isCorrect,
			Duration:      a.Duration,
		})

		if isCorrect {
			correctCount++
		}
		totalDuration += a.Duration
	}

	// 4. 批量写入答题记录
	if err := repository.BatchCreateAnswers(userAnswers); err != nil {
		return nil, fmt.Errorf("failed to batch create answers: %w", err)
	}

	// 5. 结束考试场次
	if sessionID > 0 {
		_ = repository.FinishSession(sessionID, correctCount, totalDuration)
	}

	return results, nil
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


