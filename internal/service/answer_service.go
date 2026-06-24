package service

import (
	"log/slog"
	"sort"
	"strings"

	"exam-quiz/internal/apperr"
	"exam-quiz/internal/cache"
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
func SubmitAnswer(questionID uint, userInput string, duration int, sessionID uint, userID uint) (*AnswerResult, error) {
	// 1. 获取题目
	question, err := repository.GetQuestion(questionID)
	if err != nil {
		return nil, apperr.NotFound("题目不存在")
	}

	// 2. 比对答案
	isCorrect := compareAnswers(question.Answer, userInput, question.Type)

	// 3. 保存答题记录
	answer := &model.UserAnswer{
		QuestionID:    questionID,
		ExamSessionID: sessionID,
		UserInput:     userInput,
		IsCorrect:     isCorrect,
		Duration:      duration,
		UserID:        userID,
	}
	if err := repository.SaveAnswer(answer); err != nil {
		return nil, apperr.Wrapf(500, "保存答案失败", err)
	}

	// Auto-finish the session when all questions have been answered.
	// Optimized: count first (lightweight), only fetch full session on last answer.
	if sessionID > 0 {
		// Verify session belongs to the user and is not finished
		session, sessErr := repository.GetSessionByID(sessionID, userID)
		if sessErr != nil {
			return nil, apperr.NotFound("考试场次不存在或无权访问")
		}
		if session.FinishedAt != nil {
			// Session already finished — still return the answer result
			// but skip auto-finish logic
			return &AnswerResult{
				IsCorrect:     isCorrect,
				CorrectAnswer: question.Answer,
				Analysis:      question.Analysis,
				UserInput:     userInput,
			}, nil
		}
		if answerCount, countErr := repository.CountSessionAnswers(sessionID); countErr == nil && session.TotalCount > 0 && answerCount >= int64(session.TotalCount) {
			if answers, aErr := repository.GetSessionAnswersRaw(sessionID); aErr == nil {
				correctCount := 0
				totalDuration := 0
				for _, a := range answers {
					if a.IsCorrect {
						correctCount++
					}
					totalDuration += a.Duration
				}
				if finishErr := repository.FinishSession(sessionID, correctCount, totalDuration, userID); finishErr != nil {
					slog.Warn("auto-finish session failed", "session_id", sessionID, "error", finishErr)
				}
			}
		}
	}

	// 4. 清除统计缓存
	cache.InvalidateModuleStats(question.ModuleID, userID)

	// 5. 返回结果
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
func SubmitBatchAnswers(answers []BatchAnswerItem, sessionID uint, userID uint) ([]AnswerResult, error) {
	if len(answers) == 0 {
		return nil, nil
	}

	// 收集所有 question_id
	ids := make([]uint, len(answers))
	for i, a := range answers {
		ids[i] = a.QuestionID
	}

	// 批量查询题目
	questionMap, err := repository.BatchGetQuestions(ids)
	if err != nil {
		return nil, apperr.Wrapf(500, "查询题目失败", err)
	}

	// 内存中比对答案
	var results []AnswerResult
	var userAnswers []model.UserAnswer

	for _, a := range answers {
		question, ok := questionMap[a.QuestionID]
		if !ok {
			return nil, apperr.NotFound("题目不存在")
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
			UserID:        userID,
		})
	}

	// 批量写入
	if err := repository.BatchCreateAnswers(userAnswers); err != nil {
		return nil, apperr.Wrapf(500, "保存答案失败", err)
	}

	// 清除所有相关模块的统计缓存
	seen := make(map[uint]bool)
	for _, q := range questionMap {
		if !seen[q.ModuleID] {
			seen[q.ModuleID] = true
			cache.InvalidateModuleStats(q.ModuleID, userID)
		}
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
func SubmitBatchAnswersWithSession(sessionID uint, answers []BatchAnswerItem, userID uint) ([]AnswerResult, error) {
	if len(answers) == 0 {
		return nil, nil
	}

	// Verify session belongs to the user
	session, err := repository.GetSessionByID(sessionID, userID)
	if err != nil {
		return nil, apperr.NotFound("考试场次不存在或无权访问")
	}

	// 防止对已结束的场次重复提交
	if session.FinishedAt != nil {
		return nil, apperr.Conflict("该考试场次已结束，无法再次提交")
	}

	// 1. 收集所有 question_id
	ids := make([]uint, len(answers))
	for i, a := range answers {
		ids[i] = a.QuestionID
	}

	// 2. 批量查询题目
	questionMap, err := repository.BatchGetQuestions(ids)
	if err != nil {
		return nil, apperr.Wrapf(500, "查询题目失败", err)
	}

	// 3. 在内存中比对答案并构建批量写入数据
	var results []AnswerResult
	var userAnswers []model.UserAnswer
	correctCount := 0
	totalDuration := 0

	for _, a := range answers {
		question, ok := questionMap[a.QuestionID]
		if !ok {
			return nil, apperr.NotFound("题目不存在")
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
			UserID:        userID,
		})

		if isCorrect {
			correctCount++
		}
		totalDuration += a.Duration
	}

	// 4. 批量写入答题记录
	if err := repository.BatchCreateAnswers(userAnswers); err != nil {
		return nil, apperr.Wrapf(500, "保存答案失败", err)
	}

	// 5. 结束考试场次
	if sessionID > 0 {
		if err := repository.FinishSession(sessionID, correctCount, totalDuration, userID); err != nil {
			slog.Warn("failed to finish session", "session_id", sessionID, "error", err)
		}
	}

	// 6. 清除统计缓存（清除所有相关模块的缓存）
	seen := make(map[uint]bool)
	for _, q := range questionMap {
		if !seen[q.ModuleID] {
			seen[q.ModuleID] = true
			cache.InvalidateModuleStats(q.ModuleID, userID)
		}
	}

	return results, nil
}

// GetSession 获取考试场次详情（校验用户归属）
func GetSession(id uint, userID uint) (*model.ExamSession, error) {
	return repository.GetSessionByID(id, userID)
}

// GetSessions 获取考试场次列表
func GetSessions(page, size int, userID uint) ([]model.ExamSession, int64, error) {
	return repository.GetSessions(page, size, userID)
}

// GetSessionAnswers returns all answers for a session (with user ownership check).
func GetSessionAnswers(sessionID uint, userID uint) ([]model.UserAnswer, error) {
	return repository.GetSessionAnswers(sessionID, userID)
}

// GetSessionAnswersPaginated returns paginated answers for a session (with user ownership check).
func GetSessionAnswersPaginated(sessionID uint, page, size int, userID uint) ([]model.UserAnswer, int64, error) {
	return repository.GetSessionAnswersPaginated(sessionID, page, size, userID)
}

// GetDashboardStats 获取仪表盘统计数据
func GetDashboardStats(userID uint) (*repository.DashboardStats, error) {
	return repository.GetDashboardStats(userID)
}

// GetAdminDashboardStats 获取管理员全局数据看板
func GetAdminDashboardStats() (*repository.AdminDashboardStats, error) {
	return repository.GetAdminDashboardStats()
}
