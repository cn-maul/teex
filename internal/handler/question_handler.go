package handler

import (
	"fmt"
	"log/slog"

	"exam-quiz/internal/model"
	"exam-quiz/internal/response"
	"exam-quiz/internal/service"
	"exam-quiz/internal/validator"

	"github.com/gin-gonic/gin"
)

// ListQuestions 查询题目列表
func ListQuestions(c *gin.Context) {
	page, size := validator.ParsePagination(c)
	filter := service.QuestionFilter{
		ModuleID:   validator.ParseOptionalUint(c, "module_id"),
		ExamTypeID: validator.ParseOptionalUint(c, "exam_type_id"),
		Type:       c.Query("type"),
		Difficulty: validator.ParseOptionalInt(c, "difficulty", 0),
		Page:       page,
		Size:       size,
	}

	questions, total, err := service.ListQuestions(filter)
	if err != nil {
		slog.Error("list questions failed", "error", err)
		response.HandleError(c, err)
		return
	}

	response.List(c, questions, total)
}

// GetQuestion 获取单个题目
func GetQuestion(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的题目 ID")
		return
	}

	question, err := service.GetQuestion(id)
	if err != nil {
		slog.Error("get question failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, question)
}

// CreateQuestion 创建题目
func CreateQuestion(c *gin.Context) {
	var question model.Question
	if err := c.ShouldBindJSON(&question); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	// Validate module exists
	if err := service.ValidateModuleExists(question.ModuleID); err != nil {
		response.HandleError(c, err)
		return
	}

	if err := validator.ValidateQuestion(&question); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	if err := service.CreateQuestion(&question); err != nil {
		slog.Error("create question failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.Created(c, question)
}

// UpdateQuestion 更新题目
func UpdateQuestion(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的题目 ID")
		return
	}

	// Check existence
	if _, err := service.GetQuestion(id); err != nil {
		slog.Error("update question check failed", "error", err)
		response.HandleError(c, err)
		return
	}

	var question model.Question
	if err := c.ShouldBindJSON(&question); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}
	question.ID = id

	// Validate module_id exists if provided
	if question.ModuleID > 0 {
		if err := service.ValidateModuleExists(question.ModuleID); err != nil {
			response.HandleError(c, err)
			return
		}
	}

	if err := validator.ValidateQuestion(&question); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	if err := service.UpdateQuestion(&question); err != nil {
		slog.Error("update question failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, question)
}

// DeleteQuestion 删除题目
func DeleteQuestion(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的题目 ID")
		return
	}

	// Check existence
	if _, err := service.GetQuestion(id); err != nil {
		slog.Error("delete question check failed", "error", err)
		response.HandleError(c, err)
		return
	}

	if err := service.DeleteQuestion(id); err != nil {
		slog.Error("delete question failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OKWithMessage(c, nil, "删除成功")
}

// ImportQuestions 批量导入题目
func ImportQuestions(c *gin.Context) {
	var questions []model.Question
	if err := c.ShouldBindJSON(&questions); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	result, err := service.ImportQuestions(questions)
	if err != nil {
		slog.Error("import questions failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, gin.H{
		"message":       "导入成功",
		"count":         result.ImportedCount,
		"invalid_count": result.InvalidCount,
	})
}

// BatchDeleteQuestions 批量删除题目
func BatchDeleteQuestions(c *gin.Context) {
	var req struct {
		IDs []uint `json:"ids" binding:"required"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}
	batchLimit := service.GetBatchLimit()
	if len(req.IDs) > batchLimit {
		response.Error(c, 400, fmt.Sprintf("单次删除不能超过 %d 道题目", batchLimit))
		return
	}
	if len(req.IDs) == 0 {
		response.Error(c, 400, "请选择要删除的题目")
		return
	}

	deleted, err := service.BatchDeleteQuestions(req.IDs)
	if err != nil {
		slog.Error("batch delete questions failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, gin.H{"deleted": deleted})
}
