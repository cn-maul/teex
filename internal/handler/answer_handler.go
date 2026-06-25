package handler

import (
	"log/slog"
	"net/http"

	"exam-quiz/internal/response"
	"exam-quiz/internal/service"
	"exam-quiz/internal/validator"

	"github.com/gin-gonic/gin"
)

// StartQuizRequest 开始刷题请求
type StartQuizRequest struct {
	ModuleID   uint   `json:"module_id" binding:"required"`
	Count      int    `json:"count"`
	Mode       string `json:"mode"` // default/wrong/random
	Difficulty int    `json:"difficulty"`
	Tags       string `json:"tags"`
}

// StartQuiz 开始刷题
func StartQuiz(c *gin.Context) {
	var req StartQuizRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	questions, sessionID, err := service.StartQuiz(req.ModuleID, req.Count, req.Mode, req.Difficulty, req.Tags, userID)
	if err != nil {
		slog.Error("start quiz failed", "error", err)
		response.HandleError(c, err)
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":       questions,
		"total":      len(questions),
		"module":     req.ModuleID,
		"mode":       req.Mode,
		"session_id": sessionID,
	})
}

// SubmitAnswerRequest 提交答案请求
type SubmitAnswerRequest struct {
	QuestionID uint   `json:"question_id" binding:"required"`
	UserInput  string `json:"user_input" binding:"required"`
	Duration   int    `json:"duration"`
	SessionID  uint   `json:"session_id"`
}

// SubmitAnswer 提交答案
func SubmitAnswer(c *gin.Context) {
	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	result, err := service.SubmitAnswer(req.QuestionID, req.UserInput, req.Duration, req.SessionID, userID)
	if err != nil {
		slog.Error("submit answer failed", "error", err)
		response.HandleError(c, err)
		return
	}

	response.OK(c, result)
}

// SubmitBatchAnswersRequest 批量提交答案请求
type SubmitBatchAnswersRequest struct {
	Answers   []service.BatchAnswerItem `json:"answers" binding:"required"`
	SessionID uint                       `json:"session_id"`
}

// SubmitBatchAnswers 批量提交答案（考试模式交卷）
func SubmitBatchAnswers(c *gin.Context) {
	var req SubmitBatchAnswersRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}

	results, err := service.SubmitBatchAnswersWithSession(req.SessionID, req.Answers, userID)
	if err != nil {
		slog.Error("submit batch answers failed", "error", err)
		response.HandleError(c, err)
		return
	}

	response.OK(c, results)
}

// GetStats 获取统计数据
func GetStats(c *gin.Context) {
	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	stats, err := service.GetOverallStats(userID)
	if err != nil {
		slog.Error("get stats failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, stats)
}

// GetModuleStats 获取模块统计
func GetModuleStats(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的模块 ID")
		return
	}

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	stats, err := service.GetModuleStats(id, userID)
	if err != nil {
		slog.Error("get module stats failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, stats)
}

// ClearAllRecords 清空当前用户的所有答题记录
func ClearAllRecords(c *gin.Context) {
	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	if err := service.ClearAllRecords(userID); err != nil {
		slog.Error("clear all records failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OKWithMessage(c, nil, "记录已清空")
}

// GetSessions 获取考试场次列表
func GetSessions(c *gin.Context) {
	page, size := validator.ParsePagination(c)

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	sessions, total, err := service.GetSessions(page, size, userID)
	if err != nil {
		slog.Error("get sessions failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.List(c, sessions, total)
}

// GetSession 获取单个考试场次详情
func GetSession(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的场次 ID")
		return
	}

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	session, err := service.GetSession(id, userID)
	if err != nil {
		slog.Error("get session failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, session)
}

// GetSessionAnswers 获取某个场次的答题记录（支持分页）
func GetSessionAnswers(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的场次 ID")
		return
	}

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}

	// Check pagination parameters
	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	if pageStr != "" || sizeStr != "" {
		page, size := validator.ParsePagination(c)

		answers, total, err := service.GetSessionAnswersPaginated(id, page, size, userID)
		if err != nil {
			slog.Error("get session answers failed", "error", err)
			response.HandleError(c, err)
			return
		}
		response.List(c, answers, total)
		return
	}

	answers, err := service.GetSessionAnswers(id, userID)
	if err != nil {
		slog.Error("get session answers failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, answers)
}

// GetDashboardStats 获取仪表盘统计数据
func GetDashboardStats(c *gin.Context) {
	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}

	stats, err := service.GetDashboardStats(userID)
	if err != nil {
		slog.Error("get dashboard stats failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, stats)
}

// GetAdminDashboardStats 获取管理员全局数据看板
func GetAdminDashboardStats(c *gin.Context) {
	stats, err := service.GetAdminDashboardStats()
	if err != nil {
		slog.Error("get admin dashboard stats failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, stats)
}
