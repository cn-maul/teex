package repository

import (
	"exam-quiz/internal/database"
	"exam-quiz/internal/model"

	"gorm.io/gorm"
)

// SaveAnswer 保存答题记录
func SaveAnswer(answer *model.UserAnswer) error {
	return database.DB.Create(answer).Error
}

// GetAnswerHistory 获取某题的答题历史
func GetAnswerHistory(questionID uint) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := database.DB.Where("question_id = ?", questionID).
		Order("created_at DESC").
		Find(&answers).Error
	return answers, err
}

// StatsResult 统计结果
type StatsResult struct {
	TotalAnswered  int64   `json:"total_answered"`
	CorrectCount   int64   `json:"correct_count"`
	Accuracy       float64 `json:"accuracy"`
	TotalQuestions  int64   `json:"total_questions"`
	Unanswered     int64   `json:"unanswered"`
}

// GetModuleStats 获取某模块的统计（正确率取每题最后一次答题记录）
func GetModuleStats(moduleID uint) (*StatsResult, error) {
	// 题目总数
	var totalQuestions int64
	err := database.DB.Model(&model.Question{}).
		Where("module_id = ?", moduleID).
		Count(&totalQuestions).Error
	if err != nil {
		return nil, err
	}

	// 已做题数（去重）
	var totalAnswered int64
	err = database.DB.Model(&model.UserAnswer{}).
		Where("question_id IN (SELECT id FROM questions WHERE module_id = ?)", moduleID).
		Distinct("question_id").
		Count(&totalAnswered).Error
	if err != nil {
		return nil, err
	}

	// 正确数：取每题最后一次答题记录中 is_correct=true 的数量
	var correctCount int64
	err = database.DB.Raw(`
		SELECT COUNT(*) FROM (
			SELECT question_id, is_correct
			FROM user_answers
			WHERE question_id IN (SELECT id FROM questions WHERE module_id = ?)
			AND id IN (
				SELECT MAX(id) FROM user_answers
				WHERE question_id IN (SELECT id FROM questions WHERE module_id = ?)
				GROUP BY question_id
			)
		) AS last_answers WHERE is_correct = true
	`, moduleID, moduleID).Scan(&correctCount).Error
	if err != nil {
		return nil, err
	}

	var accuracy float64
	if totalAnswered > 0 {
		accuracy = float64(correctCount) / float64(totalAnswered) * 100
	}

	return &StatsResult{
		TotalAnswered: totalAnswered,
		CorrectCount:  correctCount,
		Accuracy:      accuracy,
		TotalQuestions: totalQuestions,
		Unanswered:    totalQuestions - totalAnswered,
	}, nil
}

// GetOverallStats 获取全局统计（正确率取每题最后一次答题记录）
func GetOverallStats() (*StatsResult, error) {
	var totalQuestions int64
	err := database.DB.Model(&model.Question{}).Count(&totalQuestions).Error
	if err != nil {
		return nil, err
	}

	var totalAnswered int64
	err = database.DB.Model(&model.UserAnswer{}).Distinct("question_id").Count(&totalAnswered).Error
	if err != nil {
		return nil, err
	}

	// 正确数：取每题最后一次答题记录中 is_correct=true 的数量
	var correctCount int64
	err = database.DB.Raw(`
		SELECT COUNT(*) FROM (
			SELECT question_id, is_correct
			FROM user_answers
			WHERE id IN (
				SELECT MAX(id) FROM user_answers GROUP BY question_id
			)
		) AS last_answers WHERE is_correct = true
	`).Scan(&correctCount).Error
	if err != nil {
		return nil, err
	}

	var accuracy float64
	if totalAnswered > 0 {
		accuracy = float64(correctCount) / float64(totalAnswered) * 100
	}

	return &StatsResult{
		TotalAnswered: totalAnswered,
		CorrectCount:  correctCount,
		Accuracy:      accuracy,
		TotalQuestions: totalQuestions,
		Unanswered:    totalQuestions - totalAnswered,
	}, nil
}

// GetRecentAnswers 获取最近的答题记录
func GetRecentAnswers(limit int) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := database.DB.Preload("Question").
		Order("created_at DESC").
		Limit(limit).
		Find(&answers).Error
	return answers, err
}

// ClearModuleRecords 清除某模块的答题记录
func ClearModuleRecords(moduleID uint) error {
	return database.DB.
		Where("question_id IN (SELECT id FROM questions WHERE module_id = ?)", moduleID).
		Delete(&model.UserAnswer{}).Error
}

// ClearAllRecords 清除所有答题记录和考试场次
func ClearAllRecords() error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("1 = 1").Delete(&model.UserAnswer{}).Error; err != nil {
			return err
		}
		return tx.Where("1 = 1").Delete(&model.ExamSession{}).Error
	})
}
