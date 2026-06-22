package handler

import (
	"log"
	"net/http"
	"strconv"

	"exam-quiz/internal/model"
	"exam-quiz/internal/repository"
	"exam-quiz/internal/service"

	"github.com/gin-gonic/gin"
	"gorm.io/gorm"
)

// ListQuestions 查询题目列表
func ListQuestions(c *gin.Context) {
	filter := repository.QuestionFilter{}

	if moduleID := c.Query("module_id"); moduleID != "" {
		id, _ := strconv.ParseUint(moduleID, 10, 32)
		filter.ModuleID = uint(id)
	}
	if examTypeID := c.Query("exam_type_id"); examTypeID != "" {
		id, _ := strconv.ParseUint(examTypeID, 10, 32)
		filter.ExamTypeID = uint(id)
	}
	if qType := c.Query("type"); qType != "" {
		filter.Type = qType
	}
	if diff := c.Query("difficulty"); diff != "" {
		d, _ := strconv.Atoi(diff)
		filter.Difficulty = d
	}
	if page := c.Query("page"); page != "" {
		p, _ := strconv.Atoi(page)
		if p < 1 {
			p = 1
		}
		filter.Page = p
	}
	if size := c.Query("size"); size != "" {
		s, _ := strconv.Atoi(size)
		if s < 1 || s > 100 {
			s = 20
		}
		filter.Size = s
	}

	questions, total, err := service.ListQuestions(filter)
	if err != nil {
		log.Printf("ListQuestions error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	c.JSON(http.StatusOK, gin.H{
		"data":  questions,
		"total": total,
	})
}

// GetQuestion 获取单个题目
func GetQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的题目 ID"})
		return
	}

	question, err := service.GetQuestion(uint(id))
	if err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": question})
}

// CreateQuestion 创建题目
func CreateQuestion(c *gin.Context) {
	var question model.Question
	if err := c.ShouldBindJSON(&question); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 验证模块是否存在
	if _, err := repository.GetModule(question.ModuleID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "模块不存在"})
			return
		}
		c.JSON(http.StatusBadRequest, gin.H{"error": "模块不存在"})
		return
	}

	// 验证难度范围
	if question.Difficulty < 1 || question.Difficulty > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "难度范围必须在 1-5 之间"})
		return
	}

	// 验证题目类型
	validTypes := map[string]bool{"single": true, "multi": true, "judge": true, "fill": true}
	if !validTypes[question.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的题目类型，支持: single/multi/judge/fill"})
		return
	}

	// 验证必填字段
	if question.Content == "" || question.Answer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "题干和答案不能为空"})
		return
	}

	if err := service.CreateQuestion(&question); err != nil {
		log.Printf("CreateQuestion error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": question})
}

// UpdateQuestion 更新题目
func UpdateQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的题目 ID"})
		return
	}

	// 检查是否存在
	if _, err := service.GetQuestion(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	var question model.Question
	if err := c.ShouldBindJSON(&question); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	question.ID = uint(id)

	// 验证难度范围
	if question.Difficulty < 1 || question.Difficulty > 5 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "难度范围必须在 1-5 之间"})
		return
	}
	// 验证题目类型
	validTypes := map[string]bool{"single": true, "multi": true, "judge": true, "fill": true}
	if !validTypes[question.Type] {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的题目类型，支持: single/multi/judge/fill"})
		return
	}
	// 验证必填字段
	if question.Content == "" || question.Answer == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "题干和答案不能为空"})
		return
	}

	if err := service.UpdateQuestion(&question); err != nil {
		log.Printf("UpdateQuestion error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": question})
}

// DeleteQuestion 删除题目
func DeleteQuestion(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的题目 ID"})
		return
	}

	// 检查是否存在
	if _, err := service.GetQuestion(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "题目不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	if err := service.DeleteQuestion(uint(id)); err != nil {
		log.Printf("DeleteQuestion error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"message": "删除成功"})
}

// ImportQuestions 批量导入题目
func ImportQuestions(c *gin.Context) {
	var questions []model.Question
	if err := c.ShouldBindJSON(&questions); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 限制导入数量
	if len(questions) > 500 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "单次导入不能超过 500 道题目"})
		return
	}

	// 逐条校验并过滤不合法题目
	validTypes := map[string]bool{"single": true, "multi": true, "judge": true, "fill": true}
	var validQuestions []model.Question
	var invalidCount int
	for _, q := range questions {
		// 校验 module_id
		if q.ModuleID == 0 {
			invalidCount++
			continue
		}
		// 校验题目类型
		if !validTypes[q.Type] {
			invalidCount++
			continue
		}
		// 校验难度范围
		if q.Difficulty < 1 || q.Difficulty > 5 {
			invalidCount++
			continue
		}
		// 校验必填字段
		if q.Content == "" || q.Answer == "" {
			invalidCount++
			continue
		}
		validQuestions = append(validQuestions, q)
	}

	if len(validQuestions) == 0 {
		c.JSON(http.StatusBadRequest, gin.H{"error": "没有有效的题目数据"})
		return
	}

	count, err := service.BatchImportQuestions(validQuestions)
	if err != nil {
		log.Printf("ImportQuestions error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":       "导入成功",
		"count":         count,
		"invalid_count": invalidCount,
	})
}
