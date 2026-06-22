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

// GetExamTypes 获取所有考试类型
func GetExamTypes(c *gin.Context) {
	exams, err := service.GetExamTypes()
	if err != nil {
		log.Printf("GetExamTypes error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": exams})
}

// GetExamModules 获取某考试类型下的模块列表
func GetExamModules(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试类型 ID"})
		return
	}

	modules, err := service.GetModulesByExamID(uint(id))
	if err != nil {
		log.Printf("GetExamModules error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": modules})
}

// CreateExamType 创建考试类型
func CreateExamType(c *gin.Context) {
	var exam model.ExamType
	if err := c.ShouldBindJSON(&exam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 校验名称非空
	if exam.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "考试类型名称不能为空"})
		return
	}

	// 检查名称是否已存在
	if existing, _ := repository.GetExamTypeByName(exam.Name); existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该考试类型名称已存在"})
		return
	}

	if err := service.CreateExamType(&exam); err != nil {
		log.Printf("CreateExamType error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": exam})
}

// UpdateExamType 更新考试类型
func UpdateExamType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试类型 ID"})
		return
	}

	// 检查是否存在
	if _, err := repository.GetExamType(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "考试类型不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	var exam model.ExamType
	if err := c.ShouldBindJSON(&exam); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	exam.ID = uint(id)
	if err := service.UpdateExamType(&exam); err != nil {
		log.Printf("UpdateExamType error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": exam})
}

// DeleteExamType 删除考试类型
func DeleteExamType(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的考试类型 ID"})
		return
	}

	// 先检查是否存在
	if _, err := repository.GetExamType(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "考试类型不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	// 计算级联删除的影响数量
	modules, questions, answers, countErr := repository.CountAffectedByExamType(uint(id))
	if countErr != nil {
		log.Printf("CountAffectedByExamType error: %v", countErr)
	}

	if err := service.DeleteExamType(uint(id)); err != nil {
		log.Printf("DeleteExamType error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":             "删除成功",
		"affected_modules":    modules,
		"affected_questions":  questions,
		"affected_answers":    answers,
	})
}

// CreateModule 创建模块
func CreateModule(c *gin.Context) {
	var module model.Module
	if err := c.ShouldBindJSON(&module); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}

	// 校验名称非空
	if module.Name == "" {
		c.JSON(http.StatusBadRequest, gin.H{"error": "模块名称不能为空"})
		return
	}

	// 验证考试类型是否存在
	if _, err := repository.GetExamType(module.ExamTypeID); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "考试类型不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	// 检查同考试类型下是否有同名模块
	if existing, _ := repository.GetModuleByNameAndExamID(module.Name, module.ExamTypeID); existing != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "该考试类型下已存在同名模块"})
		return
	}

	if err := service.CreateModule(&module); err != nil {
		log.Printf("CreateModule error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusCreated, gin.H{"data": module})
}

// UpdateModule 更新模块
func UpdateModule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模块 ID"})
		return
	}

	// 检查是否存在
	if _, err := repository.GetModule(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "模块不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	var module model.Module
	if err := c.ShouldBindJSON(&module); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	module.ID = uint(id)
	if err := service.UpdateModule(&module); err != nil {
		log.Printf("UpdateModule error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": module})
}

// DeleteModule 删除模块
func DeleteModule(c *gin.Context) {
	idStr := c.Param("id")
	id, err := strconv.ParseUint(idStr, 10, 32)
	if err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "无效的模块 ID"})
		return
	}

	// 先检查是否存在
	if _, err := repository.GetModule(uint(id)); err != nil {
		if err == gorm.ErrRecordNotFound {
			c.JSON(http.StatusNotFound, gin.H{"error": "模块不存在"})
			return
		}
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}

	// 计算级联删除的影响数量
	questions, answers, countErr := repository.CountAffectedByModule(uint(id))
	if countErr != nil {
		log.Printf("CountAffectedByModule error: %v", countErr)
	}

	if err := service.DeleteModule(uint(id)); err != nil {
		log.Printf("DeleteModule error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "操作失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{
		"message":            "删除成功",
		"affected_questions": questions,
		"affected_answers":   answers,
	})
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	c.JSON(http.StatusOK, gin.H{"status": "ok"})
}

// ExportData 导出所有数据为JSON
func ExportData(c *gin.Context) {
	data, err := service.ExportAllData()
	if err != nil {
		log.Printf("ExportData error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导出失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// ImportFullData 导入完整数据（考试类型+模块+题目）
func ImportFullData(c *gin.Context) {
	var data service.FullImportData
	if err := c.ShouldBindJSON(&data); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": "请求参数无效"})
		return
	}
	result, err := service.ImportFullData(data)
	if err != nil {
		log.Printf("ImportFullData error: %v", err)
		c.JSON(http.StatusInternalServerError, gin.H{"error": "导入失败，请稍后重试"})
		return
	}
	c.JSON(http.StatusOK, gin.H{"data": result})
}
