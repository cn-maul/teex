package validator

import (
	"strconv"

	"github.com/gin-gonic/gin"
)

// ParsePagination 从 query string 解析分页参数，带默认值和边界保护。
func ParsePagination(c *gin.Context) (page, size int) {
	page, _ = strconv.Atoi(c.DefaultQuery("page", "1"))
	size, _ = strconv.Atoi(c.DefaultQuery("size", "20"))
	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	return
}
