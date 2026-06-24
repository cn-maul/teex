package repository

import (
	"exam-quiz/internal/database"
	"exam-quiz/internal/model"
	"time"

	"gorm.io/gorm"
)

// CreateSession 创建考试场次
func CreateSession(session *model.ExamSession) error {
	return database.DB.Create(session).Error
}

// FinishSession 完成考试场次
func FinishSession(sessionID uint, correctCount int, duration int, userID uint) error {
	return database.DB.Model(&model.ExamSession{}).Where("id = ? AND user_id = ?", sessionID, userID).Updates(map[string]interface{}{
		"correct_count": correctCount,
		"duration":      duration,
		"finished_at":   time.Now(),
	}).Error
}

// GetSessionByID 获取单个考试场次（校验用户归属）
func GetSessionByID(id uint, userID uint) (*model.ExamSession, error) {
	var session model.ExamSession
	err := database.DB.Preload("Module").Where("id = ? AND user_id = ?", id, userID).First(&session).Error
	return &session, err
}

// GetSessions 获取考试场次列表（分页）
func GetSessions(page, size int, userID uint) ([]model.ExamSession, int64, error) {
	var sessions []model.ExamSession
	var total int64

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	if err := database.DB.Model(&model.ExamSession{}).Where("user_id = ?", userID).Count(&total).Error; err != nil {
		return nil, 0, err
	}
	err := database.DB.Preload("Module").
		Where("user_id = ?", userID).
		Order("started_at DESC").
		Offset(offset).Limit(size).
		Find(&sessions).Error

	return sessions, total, err
}

// GetSessionAnswers 获取某个场次的所有答题记录（校验用户归属）
func GetSessionAnswers(sessionID uint, userID uint) ([]model.UserAnswer, error) {
	// 先校验场次归属
	var count int64
	if err := database.DB.Model(&model.ExamSession{}).Where("id = ? AND user_id = ?", sessionID, userID).Count(&count).Error; err != nil {
		return nil, err
	}
	if count == 0 {
		return nil, gorm.ErrRecordNotFound
	}

	var answers []model.UserAnswer
	err := database.DB.Preload("Question").
		Where("exam_session_id = ?", sessionID).
		Order("created_at ASC").
		Find(&answers).Error
	return answers, err
}

// CountSessionAnswers returns the number of answers recorded for a session.
func CountSessionAnswers(sessionID uint) (int64, error) {
	var count int64
	err := database.DB.Model(&model.UserAnswer{}).Where("exam_session_id = ?", sessionID).Count(&count).Error
	return count, err
}

// GetSessionAnswersRaw returns all answers for a session without preloading
// the Question relation (lighter than GetSessionAnswers; used for stats).
func GetSessionAnswersRaw(sessionID uint) ([]model.UserAnswer, error) {
	var answers []model.UserAnswer
	err := database.DB.Where("exam_session_id = ?", sessionID).Find(&answers).Error
	return answers, err
}

// GetSessionAnswersPaginated 分页获取某个场次的答题记录（校验用户归属）
func GetSessionAnswersPaginated(sessionID uint, page, size int, userID uint) ([]model.UserAnswer, int64, error) {
	var answers []model.UserAnswer
	var total int64

	if page < 1 {
		page = 1
	}
	if size < 1 || size > 100 {
		size = 20
	}
	offset := (page - 1) * size

	// 先校验场次归属
	var sessionCount int64
	if err := database.DB.Model(&model.ExamSession{}).Where("id = ? AND user_id = ?", sessionID, userID).Count(&sessionCount).Error; err != nil {
		return nil, 0, err
	}
	if sessionCount == 0 {
		return nil, 0, gorm.ErrRecordNotFound
	}

	err := database.DB.Model(&model.UserAnswer{}).Where("exam_session_id = ?", sessionID).Count(&total).Error
	if err != nil {
		return nil, 0, err
	}

	err = database.DB.Preload("Question").
		Where("exam_session_id = ?", sessionID).
		Order("created_at ASC").
		Offset(offset).Limit(size).
		Find(&answers).Error

	return answers, total, err
}
