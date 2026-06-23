package handler

import (
	"log"
	"strconv"
	"strings"

	"exam-quiz/internal/cache"
	"exam-quiz/internal/response"
	"exam-quiz/internal/service"
	"exam-quiz/internal/validator"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
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

	if req.Count <= 0 {
		req.Count = 10
	}
	if req.Count > 200 {
		req.Count = 200
	}
	if req.Mode == "" {
		req.Mode = "default"
	}
	if req.Difficulty < 0 || req.Difficulty > 5 {
		response.Error(c, 400, "难度范围必须在 0-5 之间")
		return
	}
	// Validate mode
	validModes := map[string]bool{"default": true, "wrong": true, "random": true, "exam": true}
	if !validModes[req.Mode] {
		response.Error(c, 400, "无效的刷题模式")
		return
	}

	// Validate module exists
	if err := service.ValidateModuleExists(req.ModuleID); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "模块不存在")
			return
		}
		response.Error(c, 400, "模块不存在")
		return
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	questions, sessionID, err := service.StartQuiz(req.ModuleID, req.Count, req.Mode, req.Difficulty, req.Tags, userID)
	if err != nil {
		log.Printf("StartQuiz error: %v", err)
		// Business errors like "该模块暂无题目" are safe to show
		errMsg := err.Error()
		if strings.Contains(errMsg, "创建考试场次失败") {
			errMsg = "操作失败，请稍后重试"
		}
		response.Error(c, 400, errMsg)
		return
	}

	response.OK(c, gin.H{
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

	if len(req.UserInput) > 200 {
		response.Error(c, 400, "答案内容过长")
		return
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	result, err := service.SubmitAnswer(req.QuestionID, req.UserInput, req.Duration, req.SessionID, userID)
	if err != nil {
		log.Printf("SubmitAnswer error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
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

	if len(req.Answers) > 500 {
		response.Error(c, 400, "答案数量超出限制")
		return
	}

	var results []service.AnswerResult
	var err error

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	if req.SessionID > 0 {
		results, err = service.SubmitBatchAnswersWithSession(req.SessionID, req.Answers, userID)
	} else {
		results, err = service.SubmitBatchAnswers(req.Answers, userID)
	}

	if err != nil {
		log.Printf("SubmitBatchAnswers error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}

	response.OK(c, results)
}

// GetStats 获取统计数据
func GetStats(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	stats, err := service.GetOverallStats(userID)
	if err != nil {
		log.Printf("GetStats error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
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

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	stats, err := service.GetModuleStats(id, userID)
	if err != nil {
		log.Printf("GetModuleStats error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, stats)
}

// ClearAllRecords 清空当前用户的所有答题记录
func ClearAllRecords(c *gin.Context) {
	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	uid := userID
	if err := service.ClearAllRecords(uid); err != nil {
		log.Printf("ClearAllRecords error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	// Clear the user's stats cache
	cache.InvalidateUserStats(uid)
	response.OKWithMessage(c, nil, "记录已清空")
}

// GetSessions 获取考试场次列表
func GetSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	sessions, total, err := service.GetSessions(page, size, userID)
	if err != nil {
		log.Printf("GetSessions error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
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

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)
	session, err := service.GetSession(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "场次不存在")
			return
		}
		response.Error(c, 500, "操作失败，请稍后重试")
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

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID := userIDRaw.(uint)

	// Check pagination parameters
	pageStr := c.Query("page")
	sizeStr := c.Query("size")

	if pageStr != "" || sizeStr != "" {
		page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
		size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))
		if page < 1 {
			page = 1
		}
		if size < 1 || size > 100 {
			size = 20
		}

		answers, total, err := service.GetSessionAnswersPaginated(id, page, size, userID)
		if err != nil {
			if err == gorm.ErrRecordNotFound {
				response.Error(c, 404, "场次不存在")
				return
			}
			log.Printf("GetSessionAnswers error: %v", err)
			response.Error(c, 500, "操作失败，请稍后重试")
			return
		}
		response.List(c, answers, total)
		return
	}

	answers, err := service.GetSessionAnswers(id, userID)
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "场次不存在")
			return
		}
		log.Printf("GetSessionAnswers error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, answers)
}
