package response

import (
	"net/http"

	"github.com/gin-gonic/gin"
)

// OK sends a 200 response with data.
func OK(c *gin.Context, data interface{}) {
	c.JSON(http.StatusOK, gin.H{"data": data})
}

// OKWithMessage sends a 200 response with data and a message.
func OKWithMessage(c *gin.Context, data interface{}, msg string) {
	resp := gin.H{"data": data, "message": msg}
	c.JSON(http.StatusOK, resp)
}

// Created sends a 201 response with data.
func Created(c *gin.Context, data interface{}) {
	c.JSON(http.StatusCreated, gin.H{"data": data})
}

// List sends a 200 response with a paginated data list and total count.
func List(c *gin.Context, data interface{}, total int64) {
	c.JSON(http.StatusOK, gin.H{"data": data, "total": total})
}

// Error sends an error response with the given HTTP status code.
func Error(c *gin.Context, code int, msg string) {
	c.JSON(code, gin.H{"error": msg})
}
