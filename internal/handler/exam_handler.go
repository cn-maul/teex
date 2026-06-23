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

// GetExamTypes 获取所有考试类型
func GetExamTypes(c *gin.Context) {
	exams, err := service.GetExamTypes()
	if err != nil {
		log.Printf("GetExamTypes error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, exams)
}

// GetExamModules 获取某考试类型下的模块列表
func GetExamModules(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的考试类型 ID")
		return
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID, ok := userIDRaw.(uint)
	if !ok {
		response.Error(c, 401, "认证信息无效")
		return
	}
	modules, err := service.GetModulesByExamID(id, userID)
	if err != nil {
		log.Printf("GetExamModules error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, modules)
}

// CreateExamType 创建考试类型
func CreateExamType(c *gin.Context) {
	var exam model.ExamType
	if err := c.ShouldBindJSON(&exam); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	if exam.Name == "" {
		response.Error(c, 400, "考试类型名称不能为空")
		return
	}

	if err := service.CheckExamTypeNameUnique(exam.Name); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	if err := service.CreateExamType(&exam); err != nil {
		log.Printf("CreateExamType error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.Created(c, exam)
}

// UpdateExamType 更新考试类型
func UpdateExamType(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的考试类型 ID")
		return
	}

	// Check existence
	if err := service.ValidateExamTypeExists(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "考试类型不存在")
			return
		}
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}

	var exam model.ExamType
	if err := c.ShouldBindJSON(&exam); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}
	exam.ID = id

	// Check name uniqueness (exclude current record)
	if exam.Name != "" {
		existing, err := service.GetExamTypeByName(exam.Name)
		if err == nil && existing.ID != id {
			response.Error(c, 400, "考试类型名称已存在")
			return
		}
		if err != nil && err != gorm.ErrRecordNotFound {
			log.Printf("UpdateExamType name check error: %v", err)
			response.Error(c, 500, "操作失败，请稍后重试")
			return
		}
	}

	if err := service.UpdateExamType(&exam); err != nil {
		log.Printf("UpdateExamType error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, exam)
}

// DeleteExamType 删除考试类型
func DeleteExamType(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的考试类型 ID")
		return
	}

	// Check existence
	if err := service.ValidateExamTypeExists(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "考试类型不存在")
			return
		}
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}

	// Count cascade-deletion impact BEFORE the actual delete.
	// Note: counts are computed outside the transaction used by
	// service.DeleteExamType, so in theory a concurrent insert
	// between the count and the delete could make them slightly
	// stale.  This is acceptable for a stats/audit response and
	// avoids coupling the handler to the repository transaction.
	modules, questions, answers, countErr := service.CountAffectedByExamType(id)
	if countErr != nil {
		log.Printf("CountAffectedByExamType error: %v", countErr)
	}

	if err := service.DeleteExamType(id); err != nil {
		log.Printf("DeleteExamType error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, gin.H{
		"message":            "删除成功",
		"affected_modules":   modules,
		"affected_questions": questions,
		"affected_answers":   answers,
	})
}

// CreateModule 创建模块
func CreateModule(c *gin.Context) {
	var module model.Module
	if err := c.ShouldBindJSON(&module); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}

	if module.Name == "" {
		response.Error(c, 400, "模块名称不能为空")
		return
	}

	// Validate exam type exists
	if err := service.ValidateExamTypeExists(module.ExamTypeID); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "考试类型不存在")
			return
		}
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}

	// Check module name uniqueness under the exam type
	if err := service.CheckModuleNameUnique(module.Name, module.ExamTypeID); err != nil {
		response.Error(c, 400, err.Error())
		return
	}

	if err := service.CreateModule(&module); err != nil {
		log.Printf("CreateModule error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.Created(c, module)
}

// UpdateModule 更新模块
func UpdateModule(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的模块 ID")
		return
	}

	// Check existence
	if err := service.ValidateModuleExists(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "模块不存在")
			return
		}
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}

	var module model.Module
	if err := c.ShouldBindJSON(&module); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}
	module.ID = id

	// Validate exam_type_id exists if provided
	if module.ExamTypeID > 0 {
		if err := service.ValidateExamTypeExists(module.ExamTypeID); err != nil {
			response.Error(c, 400, "指定的考试类型不存在")
			return
		}
	}

	// Check module name uniqueness under the exam type (exclude current record)
	if module.Name != "" {
		examTypeID := module.ExamTypeID
		if examTypeID == 0 {
			// If exam_type_id not changed, use the existing module's exam_type_id
			if existingModule, err := service.GetModule(id); err == nil {
				examTypeID = existingModule.ExamTypeID
			}
		}
		if examTypeID > 0 {
			existing, err := service.GetModuleByNameAndExamID(module.Name, examTypeID)
			if err == nil && existing.ID != id {
				response.Error(c, 400, "该考试类型下已存在同名模块")
				return
			}
			if err != nil && err != gorm.ErrRecordNotFound {
				log.Printf("UpdateModule name check error: %v", err)
				response.Error(c, 500, "操作失败，请稍后重试")
				return
			}
		}
	}

	if err := service.UpdateModule(&module); err != nil {
		log.Printf("UpdateModule error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, module)
}

// DeleteModule 删除模块
func DeleteModule(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的模块 ID")
		return
	}

	// Check existence
	if err := service.ValidateModuleExists(id); err != nil {
		if err == gorm.ErrRecordNotFound {
			response.Error(c, 404, "模块不存在")
			return
		}
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}

	// Count cascade-deletion impact
	questions, answers, countErr := service.CountAffectedByModule(id)
	if countErr != nil {
		log.Printf("CountAffectedByModule error: %v", countErr)
	}

	if err := service.DeleteModule(id); err != nil {
		log.Printf("DeleteModule error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, gin.H{
		"message":            "删除成功",
		"affected_questions": questions,
		"affected_answers":   answers,
	})
}

// GetExamStats 获取某考试类型下各模块的统计信息
func GetExamStats(c *gin.Context) {
	id, err := validator.ParseID(c, "id")
	if err != nil {
		response.Error(c, 400, "无效的考试类型 ID")
		return
	}

	userIDRaw, exists := c.Get("user_id")
	if !exists {
		response.Error(c, 401, "未登录")
		c.Abort()
		return
	}
	userID, ok := userIDRaw.(uint)
	if !ok {
		response.Error(c, 401, "认证信息无效")
		return
	}
	stats, err := service.GetExamStats(id, userID)
	if err != nil {
		log.Printf("GetExamStats error: %v", err)
		response.Error(c, 500, "操作失败，请稍后重试")
		return
	}
	response.OK(c, stats)
}

// HealthCheck 健康检查
func HealthCheck(c *gin.Context) {
	response.OK(c, gin.H{"status": "ok"})
}

// ExportData 导出所有数据为JSON
func ExportData(c *gin.Context) {
	data, err := service.ExportAllData()
	if err != nil {
		log.Printf("ExportData error: %v", err)
		response.Error(c, 500, "导出失败，请稍后重试")
		return
	}
	response.OK(c, data)
}

// ImportFullData 导入完整数据（考试类型+模块+题目）
func ImportFullData(c *gin.Context) {
	var data service.FullImportData
	if err := c.ShouldBindJSON(&data); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}
	result, err := service.ImportFullData(data)
	if err != nil {
		log.Printf("ImportFullData error: %v", err)
		response.Error(c, 500, "导入失败，请稍后重试")
		return
	}
	response.OK(c, result)
}
