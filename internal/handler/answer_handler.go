package handler

import (
	"log"
	"net/http"
	"strconv"

	"exam-quiz/internal/repository"
	"exam-quiz/internal/service"

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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "难度范围必须在 0-5 之间"})
		return
	}
	// 校验模式值
	validModes := map[string]bool{"default": true, "wrong": true, "random": true, "exam": true}
	if !validModes[req.Mode] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的刷题模式"})
		return
	}

	// 验证模块是否存在
	if _, err := repository.GetModule(req.ModuleID); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模块不存在"})
		return
	}

	questions, sessionID, err := service.StartQuiz(req.ModuleID, req.Count, req.Mode, req.Difficulty, req.Tags)
	if err != nil {
		log.Printf("StartQuiz error: %v", err)
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
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
}

// SubmitAnswer 提交答案
func SubmitAnswer(c *gin.Context) {
	var req SubmitAnswerRequest
	if err := c.ShouldBindJSON(&req); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 校验 user_input 长度
	if len(req.UserInput) > 200 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "答案内容过长"})
		return
	}

	result, err := service.SubmitAnswer(req.QuestionID, req.UserInput, req.Duration)
	if err != nil {
		log.Printf("SubmitAnswer error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": result})
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
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 校验答案数组长度上限
	if len(req.Answers) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "答案数量超出限制"})
		return
	}

	var results []service.AnswerResult
	var err error

	if req.SessionID > 0 {
		results, err = service.SubmitBatchAnswersWithSession(req.SessionID, req.Answers)
	} else {
		results, err = service.SubmitBatchAnswers(req.Answers)
	}

	if err != nil {
		log.Printf("SubmitBatchAnswers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{"data": results})
}

// GetStats 获取统计数据
func GetStats(c *gin.Context) {
	stats, err := service.GetOverallStats()
	if err != nil {
		log.Printf("GetStats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// GetModuleStats 获取模块统计
func GetModuleStats(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模块 ID"})
		return
	}

	stats, err := service.GetModuleStats(uint(id))
	if err != nil {
		log.Printf("GetModuleStats error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": stats})
}

// ClearAllRecords 清空所有答题记录
func ClearAllRecords(c *gin.Context) {
	if err := service.ClearAllRecords(); err != nil {
		log.Printf("ClearAllRecords error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "记录已清空"})
}

// GetSessions 获取考试场次列表
func GetSessions(c *gin.Context) {
	page, _ := strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ := strconv.Atoi(c.DefaultQuery("size", "20"))

	// 兜底默认值
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}

	sessions, total, err := service.GetSessions(page, size)
	if err != nil {
		log.Printf("GetSessions error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": sessions, "total": total})
}

// GetSession 获取单个考试场次详情
func GetSession(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的场次 ID"})
		return
	}

	session, err := service.GetSession(uint(id))
	if err != nil {
		c.JSON(http.StatusNotFound, gin.H{"error": "场次不存在"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": session})
}

// GetSessionAnswers 获取某个场次的答题记录
func GetSessionAnswers(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的场次 ID"})
		return
	}

	answers, err := service.GetSessionAnswers(uint(id))
	if err != nil {
		log.Printf("GetSessionAnswers error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": answers})
}
