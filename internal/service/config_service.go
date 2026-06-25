package service

import (
	"fmt"
	"strconv"

	"exam-quiz/internal/repository"
)

const (
	batchLimitKey          = "batch_limit"
	defaultBatchLimit      = 500
	minBatchLimit          = 1
	maxBatchLimit          = 10000
	generalRateLimitKey    = "general_rate_limit"
	defaultGeneralRateLimit = 120
	minGeneralRateLimit    = 10
	maxGeneralRateLimit    = 10000
)

// GetBatchLimit 获取批量操作上限（导入/删除/提交）
func GetBatchLimit() int {
	val, err := repository.GetConfig(batchLimitKey)
	if err != nil || val == "" {
		return defaultBatchLimit
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < minBatchLimit || n > maxBatchLimit {
		return defaultBatchLimit
	}
	return n
}

// SetBatchLimit 设置批量操作上限
func SetBatchLimit(limit int) error {
	if limit < minBatchLimit || limit > maxBatchLimit {
		return fmt.Errorf("批量操作上限必须在 %d ~ %d 之间", minBatchLimit, maxBatchLimit)
	}
	return repository.SetConfig(batchLimitKey, strconv.Itoa(limit))
}

// GetGeneralRateLimit 获取每分钟通用请求频率限制
func GetGeneralRateLimit() int {
	val, err := repository.GetConfig(generalRateLimitKey)
	if err != nil || val == "" {
		return defaultGeneralRateLimit
	}
	n, err := strconv.Atoi(val)
	if err != nil || n < minGeneralRateLimit || n > maxGeneralRateLimit {
		return defaultGeneralRateLimit
	}
	return n
}

// SetGeneralRateLimit 设置每分钟通用请求频率限制
func SetGeneralRateLimit(limit int) error {
	if limit < minGeneralRateLimit || limit > maxGeneralRateLimit {
		return fmt.Errorf("请求频率限制必须在 %d ~ %d 之间", minGeneralRateLimit, maxGeneralRateLimit)
	}
	return repository.SetConfig(generalRateLimitKey, strconv.Itoa(limit))
}
