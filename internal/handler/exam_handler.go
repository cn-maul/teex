package handler

import (
	"errors"
	"log/slog"

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
		slog.Error("get exam types failed", "error", err)
		response.HandleError(c, err)
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

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	modules, err := service.GetModulesByExamID(id, userID)
	if err != nil {
		slog.Error("get exam modules failed", "error", err)
		response.HandleError(c, err)
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
		response.HandleError(c, err)
		return
	}

	if err := service.CreateExamType(&exam); err != nil {
		slog.Error("create exam type failed", "error", err)
		response.HandleError(c, err)
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
		response.HandleError(c, err)
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
		if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
			slog.Error("update exam type name check failed", "error", err)
			response.HandleError(c, err)
			return
		}
	}

	if err := service.UpdateExamType(&exam); err != nil {
		slog.Error("update exam type failed", "error", err)
		response.HandleError(c, err)
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
		response.HandleError(c, err)
		return
	}

	// Count cascade-deletion impact BEFORE the actual delete.
	modules, questions, answers, countErr := service.CountAffectedByExamType(id)
	if countErr != nil {
		slog.Error("count affected by exam type failed", "error", countErr)
	}

	if err := service.DeleteExamType(id); err != nil {
		slog.Error("delete exam type failed", "error", err)
		response.HandleError(c, err)
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
		response.HandleError(c, err)
		return
	}

	// Check module name uniqueness under the exam type
	if err := service.CheckModuleNameUnique(module.Name, module.ExamTypeID); err != nil {
		response.HandleError(c, err)
		return
	}

	if err := service.CreateModule(&module); err != nil {
		slog.Error("create module failed", "error", err)
		response.HandleError(c, err)
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
		response.HandleError(c, err)
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
			response.HandleError(c, err)
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
			if err != nil && !errors.Is(err, gorm.ErrRecordNotFound) {
				slog.Error("update module name check failed", "error", err)
				response.HandleError(c, err)
				return
			}
		}
	}

	if err := service.UpdateModule(&module); err != nil {
		slog.Error("update module failed", "error", err)
		response.HandleError(c, err)
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
		response.HandleError(c, err)
		return
	}

	// Count cascade-deletion impact
	questions, answers, countErr := service.CountAffectedByModule(id)
	if countErr != nil {
		slog.Error("count affected by module failed", "error", countErr)
	}

	if err := service.DeleteModule(id); err != nil {
		slog.Error("delete module failed", "error", err)
		response.HandleError(c, err)
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

	userID, ok := validator.GetUserID(c)
	if !ok {
		return
	}
	stats, err := service.GetExamStats(id, userID)
	if err != nil {
		slog.Error("get exam stats failed", "error", err)
		response.HandleError(c, err)
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
		slog.Error("export data failed", "error", err)
		response.HandleError(c, err)
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
		slog.Error("import full data failed", "error", err)
		response.HandleError(c, err)
		return
	}
	response.OK(c, result)
}
