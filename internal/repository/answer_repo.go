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
func GetModuleStats(moduleID uint, userID uint) (*StatsResult, error) {
	// 题目总数
	var totalQuestions int64
	err := database.DB.Model(&model.Question{}).
		Where("module_id = ?", moduleID).
		Count(&totalQuestions).Error
	if err != nil {
		return nil, err
	}

	// 已做题数和正确数：合并为一条查询
	type statsRow struct {
		TotalAnswered int64
		CorrectCount  int64
	}
	var row statsRow
	err = database.DB.Raw(`
		SELECT
			COUNT(*) AS total_answered,
			SUM(CASE WHEN is_correct THEN 1 ELSE 0 END) AS correct_count
		FROM user_answers
		WHERE id IN (
			SELECT MAX(id) FROM user_answers
			WHERE question_id IN (SELECT id FROM questions WHERE module_id = ?) AND user_id = ?
			GROUP BY question_id
		)
	`, moduleID, userID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	totalAnswered := row.TotalAnswered
	correctCount := row.CorrectCount

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
func GetOverallStats(userID uint) (*StatsResult, error) {
	var totalQuestions int64
	err := database.DB.Model(&model.Question{}).Count(&totalQuestions).Error
	if err != nil {
		return nil, err
	}

	// 已做题数和正确数：合并为一条查询
	type statsRow struct {
		TotalAnswered int64
		CorrectCount  int64
	}
	var row statsRow
	err = database.DB.Raw(`
		SELECT
			COUNT(*) AS total_answered,
			SUM(CASE WHEN is_correct THEN 1 ELSE 0 END) AS correct_count
		FROM user_answers
		WHERE id IN (
			SELECT MAX(id) FROM user_answers
			WHERE user_id = ?
			GROUP BY question_id
		)
	`, userID).Scan(&row).Error
	if err != nil {
		return nil, err
	}
	totalAnswered := row.TotalAnswered
	correctCount := row.CorrectCount

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

// GetRecentAnswers 获取最近的答题记录（按用户隔离）
func GetRecentAnswers(limit int, userID uint) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := database.DB.Preload("Question").
		Where("user_id = ?", userID).
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

// ClearAllRecords 清除当前用户的所有答题记录和考试场次
func ClearAllRecords(userID uint) error {
	return database.DB.Transaction(func(tx *gorm.DB) error {
		if err := tx.Where("user_id = ?", userID).Delete(&model.UserAnswer{}).Error; err != nil {
			return err
		}
		return tx.Where("user_id = ?", userID).Delete(&model.ExamSession{}).Error
	})
}

// BatchCreateAnswers 批量创建答题记录
func BatchCreateAnswers(answers []model.UserAnswer) error {
	if len(answers) == 0 {
		return nil
	}
	return database.DB.CreateInBatches(answers, 100).Error
}

// CountAffectedByModule 统计删除某模块级联影响的记录数
func CountAffectedByModule(moduleID uint) (questions int64, answers int64, err error) {
	qErr := database.DB.Model(&model.Question{}).Where("module_id = ?", moduleID).Count(&questions).Error
	if qErr != nil {
		return 0, 0, qErr
	}
	aErr := database.DB.Model(&model.UserAnswer{}).
		Where("question_id IN (SELECT id FROM questions WHERE module_id = ?)", moduleID).Count(&answers).Error
	if aErr != nil {
		return 0, 0, aErr
	}
	return questions, answers, nil
}

// CountAffectedByExamType 统计删除某考试类型级联影响的记录数
func CountAffectedByExamType(examTypeID uint) (modules int64, questions int64, answers int64, err error) {
	mErr := database.DB.Model(&model.Module{}).Where("exam_type_id = ?", examTypeID).Count(&modules).Error
	if mErr != nil {
		return 0, 0, 0, mErr
	}

	var moduleIDs []uint
	if err := database.DB.Model(&model.Module{}).Where("exam_type_id = ?", examTypeID).Pluck("id", &moduleIDs).Error; err != nil {
		return 0, 0, 0, err
	}

	if len(moduleIDs) == 0 {
		return modules, 0, 0, nil
	}

	qErr := database.DB.Model(&model.Question{}).Where("module_id IN ?", moduleIDs).Count(&questions).Error
	if qErr != nil {
		return 0, 0, 0, qErr
	}
	aErr := database.DB.Model(&model.UserAnswer{}).
		Where("question_id IN (SELECT id FROM questions WHERE module_id IN ?)", moduleIDs).Count(&answers).Error
	if aErr != nil {
		return 0, 0, 0, aErr
	}
	return modules, questions, answers, nil
}
