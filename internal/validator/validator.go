package validator

import (
	"fmt"
	"strconv"

	"exam-quiz/internal/model"

	"github.com/gin-gonic/gin"
)

// ValidQuestionTypes is the set of supported question types.
var ValidQuestionTypes = map[string]bool{
	"single": true,
	"multi":  true,
	"judge":  true,
	"fill":   true,
}

// ValidateQuestion checks that a question has valid content, answer, type, and difficulty.
func ValidateQuestion(q *model.Question) error {
	if q.Content == "" || q.Answer == "" {
		return fmt.Errorf("题干和答案不能为空")
	}
	if !ValidQuestionTypes[q.Type] {
		return fmt.Errorf("无效的题目类型，支持: single/multi/judge/fill")
	}
	if q.Difficulty < 1 || q.Difficulty > 5 {
		return fmt.Errorf("难度范围必须在 1-5 之间")
	}
	return nil
}

// ValidateQuestionForImport validates a question during import, allowing empty answers
// so that questions can be imported first and answers filled in later.
func ValidateQuestionForImport(q *model.Question) error {
	if q.Content == "" {
		return fmt.Errorf("题干不能为空")
	}
	if !ValidQuestionTypes[q.Type] {
		return fmt.Errorf("无效的题目类型")
	}
	if q.Difficulty < 1 || q.Difficulty > 5 {
		return fmt.Errorf("难度范围必须在 1-5 之间")
	}
	return nil
}

// ParseID extracts a uint path parameter from the gin context.
func ParseID(c *gin.Context, name string) (uint, error) {
	s := c.Param(name)
	id, err := strconv.ParseUint(s, 10, 32)
	if err != nil {
		return 0, fmt.Errorf("无效的 %s ID", name)
	}
	return uint(id), nil
}

// ParseOptionalUint extracts an optional uint query parameter, returning 0 if absent.
func ParseOptionalUint(c *gin.Context, name string) uint {
	s := c.Query(name)
	if s == "" {
		return 0
	}
	id, _ := strconv.ParseUint(s, 10, 32)
	return uint(id)
}

// ParseOptionalInt extracts an optional int query parameter, returning defaultVal if absent.
func ParseOptionalInt(c *gin.Context, name string, defaultVal int) int {
	s := c.Query(name)
	if s == "" {
		return defaultVal
	}
	v, _ := strconv.Atoi(s)
	return v
}
