package service

import (
	"fmt"
	"log/slog"
	"sort"
	"strings"

	"exam-quiz/internal/apperr"
	"exam-quiz/internal/database"
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
	if len(userInput) > 200 {
		return nil, apperr.BadRequest("答案内容过长")
	}

	// 1. 获取题目
	question, err := repository.GetQuestion(database.DB, questionID)
	if err != nil {
		return nil, err
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
	if err := repository.SaveAnswer(database.DB, answer); err != nil {
		return nil, apperr.Wrapf(500, "保存答案失败", err)
	}

	// Auto-finish the session when all questions have been answered.
	// Optimized: count first (lightweight), only fetch full session on last answer.
	if sessionID > 0 {
		// Verify session belongs to the user and is not finished
		session, sessErr := repository.GetSessionByID(database.DB, sessionID, userID)
		if sessErr != nil {
			return nil, sessErr
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
		if answerCount, countErr := repository.CountSessionAnswers(database.DB, sessionID); countErr == nil && session.TotalCount > 0 && answerCount >= int64(session.TotalCount) {
			if answers, aErr := repository.GetSessionAnswersRaw(database.DB, sessionID); aErr == nil {
				correctCount := 0
				totalDuration := 0
				for _, a := range answers {
					if a.IsCorrect {
						correctCount++
					}
					totalDuration += a.Duration
				}
				if finishErr := repository.FinishSession(database.DB, sessionID, correctCount, totalDuration, userID); finishErr != nil {
					slog.Warn("auto-finish session failed", "session_id", sessionID, "error", finishErr)
				}
			}
		}
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
		// 填空题：支持多可接受答案（管道分隔，如 "3.14|π|圆周率"）
		for _, ans := range strings.Split(correct, "|") {
			if strings.TrimSpace(ans) == userAnswer {
				return true
			}
		}
		return false
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

	batchLimit := GetBatchLimit()
	if len(answers) > batchLimit {
		return nil, apperr.BadRequest(fmt.Sprintf("单次提交答案数量不能超过 %d", batchLimit))
	}
	if sessionID == 0 {
		return nil, apperr.BadRequest("批量提交必须提供有效的考试场次 ID")
	}

	// Verify session belongs to the user
	session, err := repository.GetSessionByID(database.DB, sessionID, userID)
	if err != nil {
		return nil, err
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
	questionMap, err := repository.BatchGetQuestions(database.DB, ids)
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
	if err := repository.BatchCreateAnswers(database.DB, userAnswers); err != nil {
		return nil, apperr.Wrapf(500, "保存答案失败", err)
	}

	// 5. 结束考试场次
	if sessionID > 0 {
		if err := repository.FinishSession(database.DB, sessionID, correctCount, totalDuration, userID); err != nil {
			slog.Warn("failed to finish session", "session_id", sessionID, "error", err)
		}
	}

	return results, nil
}

// GetSession 获取考试场次详情（校验用户归属）
func GetSession(id uint, userID uint) (*model.ExamSession, error) {
	return repository.GetSessionByID(database.DB, id, userID)
}

// GetSessions 获取考试场次列表
func GetSessions(page, size int, userID uint) ([]model.ExamSession, int64, error) {
	return repository.GetSessions(database.DB, page, size, userID)
}

// GetSessionAnswers returns all answers for a session (with user ownership check).
func GetSessionAnswers(sessionID uint, userID uint) ([]model.UserAnswer, error) {
	return repository.GetSessionAnswers(database.DB, sessionID, userID)
}

// GetSessionAnswersPaginated returns paginated answers for a session (with user ownership check).
func GetSessionAnswersPaginated(sessionID uint, page, size int, userID uint) ([]model.UserAnswer, int64, error) {
	return repository.GetSessionAnswersPaginated(database.DB, sessionID, page, size, userID)
}

// GetDashboardStats 获取仪表盘统计数据
func GetDashboardStats(userID uint) (*repository.DashboardStats, error) {
	return repository.GetDashboardStats(database.DB, userID)
}

// GetAdminDashboardStats 获取管理员全局数据看板
func GetAdminDashboardStats() (*repository.AdminDashboardStats, error) {
	return repository.GetAdminDashboardStats(database.DB, )
}
