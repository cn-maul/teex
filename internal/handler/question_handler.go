package handler

import (
	"log"

	"exam-quiz/internal/model"
	"exam-quiz/internal/response"
	"exam-quiz/internal/service"
	"exam-quiz/internal/validator"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListQuestions 查询题目列表
func ListQuestions(c *gin.Context) {
	filter := service.QuestionFilter{
		ModuleID:   validator.ParseOptionalUint(c, "module_id"),
		ExamTypeID: validator.ParseOptionalUint(c, "exam_type_id"),
		Type:       c.Query("type"),
		Difficulty: validator.ParseOptionalInt(c, "difficulty", 0),
		Page:       validator.ParseOptionalInt(c, "page", 1),
		Size:       validator.ParseOptionalInt(c, "size", 20),
	}

	if filter.Page < 1 {
		filter.Page = 1
	}
	if filter.Size < 1 || filter.Size > 100 {
		filter.Size = 20
	}

	questions, total, err := service.ListQuestions(filter)
	if err != nil {
		log.Printf("ListQuestions error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
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
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "题目不存在")
			return
		}
		response.Error(c, 500, "操作失败，请稍后重试")
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
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "模块不存在")
			return
		}
		response.Error(c, 400, "模块不存在")
		return
	}

	if err := validator.ValidateQuestion(&question); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	if err := service.CreateQuestion(&question); err != nil {
		log.Printf("CreateQuestion error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
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
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "题目不存在")
			return
		}
		response.Error(c, 500, "操作失败，请稍后重试")
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
			response.Error(c, 400, "指定的模块不存在")
			return
		}
	}

	if err := validator.ValidateQuestion(&question); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	if err := service.UpdateQuestion(&question); err != nil {
		log.Printf("UpdateQuestion error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
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
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "题目不存在")
			return
		}
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}

	if err := service.DeleteQuestion(id); err != nil {
		log.Printf("DeleteQuestion error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
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

	// Limit import batch size
	if len(questions) > 500 {
		response.Error(c, 400, "单次导入不能超过 500 道题目")
		return
	}

	// Validate and filter each question
	var validQuestions []model.Question
	var invalidCount int
	for _, q := range questions {
		if q.ModuleID == 0 {
			invalidCount++
			continue
		}
		if err := validator.ValidateQuestionForImport(&q); err != nil {
			invalidCount++
			continue
		}
		validQuestions = append(validQuestions, q)
	}

	if len(validQuestions) == 0 {
		response.Error(c, 400, "没有有效的题目数据")
		return
	}

	// Validate all referenced ModuleIDs exist
	moduleIDSet := make(map[uint]bool)
	for _, q := range validQuestions {
		moduleIDSet[q.ModuleID] = true
	}
	for moduleID := range moduleIDSet {
		if err := service.ValidateModuleExists(moduleID); err != nil {
			response.Error(c, 400, "模块不存在")
			return
		}
	}

	count, err := service.BatchImportQuestions(validQuestions)
	if err != nil {
		log.Printf("ImportQuestions error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, gin.H{
		"message":       "导入成功",
		"count":         count,
		"invalid_count": invalidCount,
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
	if len(req.IDs) > 500 {
		response.Error(c, 400, "单次删除不能超过 500 道题目")
		return
	}
	if len(req.IDs) == 0 {
		response.Error(c, 400, "请选择要删除的题目")
		return
	}

	deleted, err := service.BatchDeleteQuestions(req.IDs)
	if err != nil {
		log.Printf("BatchDeleteQuestions error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, gin.H{"deleted": deleted})
}
