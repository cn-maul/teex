package handler

import (
	"log/slog"

	"exam-quiz/internal/response"
	"exam-quiz/internal/service"

	"github.com/gin-gonic/gin"
)

// GetBatchLimit 获取批量操作上限
func GetBatchLimit(c *gin.Context) {
	limit := service.GetBatchLimit()
	response.OK(c, gin.H{"limit": limit})
}

// SetBatchLimit 设置批量操作上限（仅管理员）
func SetBatchLimit(c *gin.Context) {
	var req struct {
		Limit int `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}
	if err := service.SetBatchLimit(req.Limit); err != nil {
		slog.Error("set batch limit failed", "error", err)
		response.Error(c, 400, err.Error())
		return
	}
	response.OKWithMessage(c, gin.H{"limit": req.Limit}, "批量操作上限已更新")
}

// GetGeneralRateLimit 获取通用请求频率限制
func GetGeneralRateLimit(c *gin.Context) {
	limit := service.GetGeneralRateLimit()
	response.OK(c, gin.H{"limit": limit})
}

// SetGeneralRateLimit 设置通用请求频率限制（仅管理员）
func SetGeneralRateLimit(c *gin.Context) {
	var req struct {
		Limit int `json:"limit"`
	}
	if err := c.ShouldBindJSON(&req); err != nil {
		response.Error(c, 400, "请求参数无效")
		return
	}
	if err := service.SetGeneralRateLimit(req.Limit); err != nil {
		slog.Error("set general rate limit failed", "error", err)
		response.Error(c, 400, err.Error())
		return
	}
	response.OKWithMessage(c, gin.H{"limit": req.Limit}, "请求频率限制已更新")
}
